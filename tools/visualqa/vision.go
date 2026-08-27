package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// The vision client follows the ai-engineer platform facts (verified against
// api-docs.deepseek.com, 2026-08-26): images only in user messages, base64
// data URL, JSON mode, retry 429/500/503 only, never 4xx, repair → one
// re-ask → UNCERTAIN on malformed JSON-mode content.

type chatRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	ResponseFormat responseFormat `json:"response_format"`
	MaxTokens      int            `json:"max_tokens"`
}

type chatMessage struct {
	Role    string        `json:"role"`
	Content []contentPart `json:"content"`
}

type contentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL string `json:"url"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func (r *chatResponse) content() string {
	if len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Message.Content
}

type providerError struct {
	Status       int
	Message      string
	Retryable    bool
	EmptyBalance bool
}

func (e *providerError) Error() string {
	return fmt.Sprintf("provider error %d: %s", e.Status, e.Message)
}

type visionClient struct {
	baseURL string
	apiKey  string
	model   string
	retries int
	timeout time.Duration
	http    *http.Client
	backoff func(attempt int) time.Duration
}

type visionUsage struct {
	PromptTokens     int
	CompletionTokens int
}

// defaultBackoff: exponential base with full jitter, capped at 5s.
func defaultBackoff(attempt int) time.Duration {
	base := time.Duration(250*math.Pow(2, float64(attempt))) * time.Millisecond
	if base > 5*time.Second {
		base = 5 * time.Second
	}
	if base <= 0 {
		base = time.Millisecond
	}
	return time.Duration(rand.Int63n(int64(base)))
}

const systemPromptTemplate = `You are a visual QA reviewer. A screenshot of a web page will be provided.
Evaluate the screenshot against each checklist item and respond with ONLY a JSON object of this exact shape:
{"checks": [{"item": "<checklist item>", "verdict": "PASS|FAIL|UNCERTAIN", "reason": "<short reason>"}]}
Every checklist item must appear exactly once. Verify each item against what is visible in the screenshot: PASS only when the frame shows the criterion clearly holding, FAIL only when it clearly does not. When the screenshot cannot show the required evidence, mark the item UNCERTAIN — never PASS on missing evidence.
Checklist:
%s`

const reaskSuffix = "\nYour previous response was invalid. Return ONLY the JSON object, no markdown, in this shape: {\"checks\":[{\"item\":\"...\",\"verdict\":\"PASS|FAIL|UNCERTAIN\",\"reason\":\"...\"}]}."

// maxChecksPerStep caps the parsed response so a runaway model can't flood
// the report; the device "all" checklists (28 items) fit comfortably.
const maxChecksPerStep = 32

// analyze sends one screenshot to the vision model and returns the parsed
// checks. Malformed responses go through the ladder: local repair → one
// re-ask → a single UNCERTAIN fallback (never a hard crash).
func (c *visionClient) analyze(ctx context.Context, png []byte, step, device, feature, checklist string) ([]checkResult, visionUsage, error) {
	var usage visionUsage
	system := fmt.Sprintf(systemPromptTemplate, checklist)
	userText := fmt.Sprintf("Screenshot from a %s viewport, step %q of %q.", device, step, feature)

	resp, err := c.chatOnce(ctx, png, userText, system)
	if err != nil {
		return nil, usage, err
	}
	addUsage(&usage, resp)
	if checks, err := parseChecks(resp.content()); err == nil {
		return checks, usage, nil
	}
	// local repair (no extra API call)
	if repaired := repairJSON(resp.content()); repaired != resp.content() {
		if checks, err := parseChecks(repaired); err == nil {
			return checks, usage, nil
		}
	}
	// one re-ask
	resp2, err := c.chatOnce(ctx, png, userText+reaskSuffix, system)
	if err != nil {
		return nil, usage, err
	}
	addUsage(&usage, resp2)
	if checks, err := parseChecks(resp2.content()); err == nil {
		return checks, usage, nil
	}
	return []checkResult{{
		Item:    "vision-parse",
		Verdict: "UNCERTAIN",
		Reason:  "unparseable vision response after local repair and one re-ask",
	}}, usage, nil
}

func addUsage(u *visionUsage, r *chatResponse) {
	if r.Usage == nil {
		return
	}
	u.PromptTokens += r.Usage.PromptTokens
	u.CompletionTokens += r.Usage.CompletionTokens
}

// chatOnce performs the HTTP call with retry/backoff on retryable errors.
func (c *visionClient) chatOnce(ctx context.Context, png []byte, userText, system string) (*chatResponse, error) {
	body := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: []contentPart{{Type: "text", Text: system}}},
			{Role: "user", Content: []contentPart{
				{Type: "text", Text: userText},
				{Type: "image_url", ImageURL: &imageURL{
					URL: "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
				}},
			}},
		},
		ResponseFormat: responseFormat{Type: "json_object"},
		MaxTokens:      4096,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encode vision request: %w", err)
	}

	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("vision request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.http.Do(req)
		if err != nil {
			if attempt < c.retries {
				sleepCtx(ctx, c.backoff(attempt))
				continue
			}
			return nil, fmt.Errorf("vision request: %w", err)
		}
		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil && attempt < c.retries {
			sleepCtx(ctx, c.backoff(attempt))
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var cr chatResponse
			if err := json.Unmarshal(respBody, &cr); err != nil {
				return nil, fmt.Errorf("vision response: %w", err)
			}
			return &cr, nil
		}

		pe := &providerError{Status: resp.StatusCode}
		var ebody struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(respBody, &ebody)
		pe.Message = ebody.Error.Message
		if pe.Message == "" {
			pe.Message = http.StatusText(resp.StatusCode)
		}
		switch resp.StatusCode {
		case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusServiceUnavailable:
			pe.Retryable = true
		case http.StatusPaymentRequired:
			pe.EmptyBalance = true
		}

		if pe.Retryable && attempt < c.retries {
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, err := strconv.Atoi(ra); err == nil && secs > 0 && secs <= 10 {
					sleepCtx(ctx, time.Duration(secs)*time.Second)
					continue
				}
			}
			sleepCtx(ctx, c.backoff(attempt))
			continue
		}
		return nil, pe
	}
}

func sleepCtx(ctx context.Context, d time.Duration) {
	select {
	case <-time.After(d):
	case <-ctx.Done():
	}
}

// parseChecks strictly parses the vision JSON-mode response.
func parseChecks(text string) ([]checkResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("empty vision response content (JSON-mode quirk)")
	}
	var vr struct {
		Checks []checkResult `json:"checks"`
	}
	dec := json.NewDecoder(strings.NewReader(text))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&vr); err != nil {
		return nil, err
	}
	if len(vr.Checks) == 0 {
		return nil, errors.New("vision response has no checks")
	}
	for i := range vr.Checks {
		c := &vr.Checks[i]
		switch c.Verdict {
		case "PASS", "FAIL", "UNCERTAIN":
		default:
			return nil, fmt.Errorf("invalid verdict %q", c.Verdict)
		}
		c.Item = clampString(c.Item, 120)
		c.Reason = clampString(c.Reason, 200)
	}
	if len(vr.Checks) > maxChecksPerStep {
		vr.Checks = vr.Checks[:maxChecksPerStep]
	}
	return vr.Checks, nil
}

// repairJSON strips markdown fences and extracts the first balanced JSON
// object. Best-effort: the re-ask is the fallback when this can't help.
func repairJSON(s string) string {
	s = strings.ReplaceAll(s, "```json", "")
	s = strings.ReplaceAll(s, "```", "")
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end <= start {
		return s
	}
	return s[start : end+1]
}

func clampString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
