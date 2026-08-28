package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeResp struct {
	status int
	body   string
}

type fakeOpenAI struct {
	t         *testing.T
	mu        sync.Mutex
	calls     int
	responses []fakeResp
	last      *chatRequest
}

func (f *fakeOpenAI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.URL.Path != "/chat/completions" {
		f.t.Errorf("path = %q, want /chat/completions", r.URL.Path)
	}
	if got := r.Header.Get("Authorization"); got != "Bearer test_key" {
		f.t.Errorf("Authorization = %q, want Bearer test_key", got)
	}
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		f.t.Errorf("decode request: %v", err)
	}
	// the request must carry the image in one of the user content blocks
	user := req.Messages[len(req.Messages)-1]
	if len(user.Content) < 2 {
		f.t.Errorf("user message content blocks = %d, want >= 2 (text + image)", len(user.Content))
	}
	var imgURL string
	for _, part := range user.Content {
		if part.ImageURL != nil {
			imgURL = part.ImageURL.URL
		}
	}
	if !strings.HasPrefix(imgURL, "data:image/png;base64,") {
		f.t.Errorf("no base64 png data URL image block, got %q", imgURL)
	}
	if req.ResponseFormat.Type != "json_object" {
		f.t.Errorf("response_format = %+v, want json_object", req.ResponseFormat)
	}
	if req.MaxTokens < 8192 {
		f.t.Errorf("MaxTokens = %d, want >= 8192 (reasoning_content eats the budget)", req.MaxTokens)
	}

	f.calls++
	idx := f.calls - 1
	if idx >= len(f.responses) {
		idx = len(f.responses) - 1
	}
	f.last = &req
	resp := f.responses[idx]
	w.WriteHeader(resp.status)
	fmt.Fprint(w, resp.body)
}

func (f *fakeOpenAI) lastRequest() *chatRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.last
}

func validBody(checksJSON string) string {
	content, err := json.Marshal(checksJSON) // escape as a JSON string literal
	if err != nil {
		panic(err)
	}
	return `{"choices":[{"message":{"content":` + string(content) + `}}],"usage":{"prompt_tokens":384,"completion_tokens":42}}`
}

func newTestClient(t *testing.T, responses []fakeResp) (*visionClient, *fakeOpenAI, string) {
	t.Helper()
	fake := &fakeOpenAI{t: t, responses: responses}
	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)
	c := &visionClient{
		baseURL: srv.URL,
		apiKey:  "test_key",
		model:   "deepseek-v4-flash-vision-exp",
		retries: 3,
		http:    srv.Client(),
		backoff: func(int) time.Duration { return time.Millisecond },
	}
	return c, fake, srv.URL
}

func TestAnalyzeSuccess(t *testing.T) {
	checks := `{"checks":[{"item":"no horizontal overflow","verdict":"PASS","reason":"fits"},{"item":"tap targets","verdict":"FAIL","reason":"bottom bar covers the CTA"}]}`
	c, fake, _ := newTestClient(t, []fakeResp{{status: 200, body: validBody(checks)}})

	got, usage, err := c.analyze(t.Context(), []byte("fake-png"), "cart-empty", "mobile", "checkout", "checklist", "")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(checks) = %d, want 2", len(got))
	}
	if got[1].Verdict != "FAIL" || !strings.Contains(got[1].Reason, "bottom bar") {
		t.Errorf("check[1] = %+v", got[1])
	}
	if usage.PromptTokens != 384 {
		t.Errorf("PromptTokens = %d, want 384", usage.PromptTokens)
	}
	if fake.calls != 1 {
		t.Errorf("calls = %d, want 1", fake.calls)
	}
}

func TestAnalyzeEmptyContentRepairThenReask(t *testing.T) {
	// JSON-mode quirk: empty content. Repair can't fix emptiness -> one re-ask.
	good := validBody(`{"checks":[{"item":"x","verdict":"PASS","reason":"ok"}]}`)
	c, fake, _ := newTestClient(t, []fakeResp{
		{status: 200, body: validBody(`""`)},
		{status: 200, body: good},
	})
	got, usage, err := c.analyze(t.Context(), []byte("fake-png"), "s", "mobile", "f", "c", "")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(got) != 1 || got[0].Verdict != "PASS" {
		t.Errorf("checks = %+v", got)
	}
	if fake.calls != 2 {
		t.Errorf("calls = %d, want 2 (initial + re-ask)", fake.calls)
	}
	if usage.PromptTokens != 384+384 {
		t.Errorf("PromptTokens = %d, want 768 (two calls)", usage.PromptTokens)
	}
}

func TestAnalyzeFencedJSONRepairsLocally(t *testing.T) {
	// Markdown-fenced JSON is repaired locally — no re-ask needed.
	body := validBody("```json\n{\"checks\":[{\"item\":\"x\",\"verdict\":\"FAIL\",\"reason\":\"overlap\"}]}\n```")
	c, fake, _ := newTestClient(t, []fakeResp{{status: 200, body: body}})
	got, _, err := c.analyze(t.Context(), []byte("fake-png"), "s", "mobile", "f", "c", "")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(got) != 1 || got[0].Verdict != "FAIL" {
		t.Errorf("checks = %+v", got)
	}
	if fake.calls != 1 {
		t.Errorf("calls = %d, want 1 (local repair, no re-ask)", fake.calls)
	}
}

func TestAnalyzeGarbageThenGarbageBecomesUncertain(t *testing.T) {
	c, fake, _ := newTestClient(t, []fakeResp{
		{status: 200, body: validBody(`"not json at all"`)},
		{status: 200, body: validBody(`"still not json"`)},
	})
	got, _, err := c.analyze(t.Context(), []byte("fake-png"), "s", "mobile", "f", "c", "")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(got) != 1 || got[0].Verdict != "UNCERTAIN" {
		t.Errorf("checks = %+v, want a single UNCERTAIN fallback", got)
	}
	if fake.calls != 2 {
		t.Errorf("calls = %d, want 2", fake.calls)
	}
}

func TestAnalyzeInvalidVerdictTriggersReask(t *testing.T) {
	bad := validBody(`{"checks":[{"item":"x","verdict":"SOMETIMES","reason":"r"}]}`)
	good := validBody(`{"checks":[{"item":"x","verdict":"PASS","reason":"ok"}]}`)
	c, fake, _ := newTestClient(t, []fakeResp{{status: 200, body: bad}, {status: 200, body: good}})
	got, _, err := c.analyze(t.Context(), []byte("fake-png"), "s", "mobile", "f", "c", "")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(got) != 1 || got[0].Verdict != "PASS" {
		t.Errorf("checks = %+v", got)
	}
	if fake.calls != 2 {
		t.Errorf("calls = %d, want 2", fake.calls)
	}
}

func TestAnalyzeRetryOn429(t *testing.T) {
	good := validBody(`{"checks":[{"item":"x","verdict":"PASS","reason":"ok"}]}`)
	c, fake, _ := newTestClient(t, []fakeResp{
		{status: 429, body: `{"error":{"message":"rate limited"}}`},
		{status: 200, body: good},
	})
	got, _, err := c.analyze(t.Context(), []byte("fake-png"), "s", "mobile", "f", "c", "")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("checks = %+v", got)
	}
	if fake.calls != 2 {
		t.Errorf("calls = %d, want 2 (1 retry)", fake.calls)
	}
}

func TestAnalyze429ExhaustedFails(t *testing.T) {
	c, fake, _ := newTestClient(t, []fakeResp{
		{status: 429, body: `{"error":{"message":"rate limited"}}`},
		{status: 429, body: `{"error":{"message":"rate limited"}}`},
		{status: 429, body: `{"error":{"message":"rate limited"}}`},
		{status: 429, body: `{"error":{"message":"rate limited"}}`},
	})
	_, _, err := c.analyze(t.Context(), []byte("fake-png"), "s", "mobile", "f", "c", "")
	if err == nil {
		t.Fatal("analyze succeeded, want provider error")
	}
	var pe *providerError
	if !asProviderError(err, &pe) {
		t.Fatalf("err = %v, want *providerError", err)
	}
	if pe.Status != 429 {
		t.Errorf("Status = %d, want 429", pe.Status)
	}
	if fake.calls != 4 {
		t.Errorf("calls = %d, want 4 (initial + 3 retries)", fake.calls)
	}
}

func TestAnalyzeNoRetryOn400(t *testing.T) {
	c, fake, _ := newTestClient(t, []fakeResp{
		{status: 400, body: `{"error":{"message":"bad request"}}`},
	})
	_, _, err := c.analyze(t.Context(), []byte("fake-png"), "s", "mobile", "f", "c", "")
	var pe *providerError
	if !asProviderError(err, &pe) {
		t.Fatalf("err = %v, want *providerError", err)
	}
	if pe.Status != 400 || pe.Retryable {
		t.Errorf("pe = %+v, want Status 400, Retryable false", pe)
	}
	if fake.calls != 1 {
		t.Errorf("calls = %d, want 1 (no retry on 400)", fake.calls)
	}
}

func TestAnalyze402EmptyBalance(t *testing.T) {
	c, fake, _ := newTestClient(t, []fakeResp{
		{status: 402, body: `{"error":{"message":"Insufficient Balance"}}`},
	})
	_, _, err := c.analyze(t.Context(), []byte("fake-png"), "s", "mobile", "f", "c", "")
	var pe *providerError
	if !asProviderError(err, &pe) {
		t.Fatalf("err = %v, want *providerError", err)
	}
	if !pe.EmptyBalance {
		t.Errorf("EmptyBalance = false, want true for 402")
	}
	if fake.calls != 1 {
		t.Errorf("calls = %d, want 1", fake.calls)
	}
}

func asProviderError(err error, target **providerError) bool {
	pe, ok := err.(*providerError)
	if !ok {
		return false
	}
	*target = pe
	return true
}

func TestAnalyzeWithHTML(t *testing.T) {
	checks := validBody(`{"checks":[{"item":"target sizes","verdict":"PASS","reason":"html confirms 44px buttons"}]}`)
	c, fake, _ := newTestClient(t, []fakeResp{{status: 200, body: checks}})
	got, _, err := c.analyze(t.Context(), []byte("fake-png"), "s", "mobile", "f", "c", "<html><body><button>Buy</button></body></html>")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(got) != 1 || got[0].Verdict != "PASS" {
		t.Errorf("checks = %+v", got)
	}
	req := fake.lastRequest()
	if req == nil {
		t.Fatal("no request captured")
	}
	user := req.Messages[len(req.Messages)-1]
	if len(user.Content) != 3 {
		t.Fatalf("user content blocks = %d, want 3 (text + image + html)", len(user.Content))
	}
	if !strings.Contains(user.Content[2].Text, "Page HTML:") || !strings.Contains(user.Content[2].Text, "<button>") {
		t.Errorf("html block = %q, want Page HTML prefix with page markup", user.Content[2].Text)
	}
	system := req.Messages[0].Content[0].Text
	if !strings.Contains(system, "structural evidence") {
		t.Errorf("system prompt lacks the HTML instruction: %q", system)
	}
	if fake.calls != 1 {
		t.Errorf("calls = %d, want 1", fake.calls)
	}
}

func TestAnalyzeWithoutHTMLNoPromptChange(t *testing.T) {
	c, fake, _ := newTestClient(t, []fakeResp{{status: 200, body: validBody(`{"checks":[{"item":"x","verdict":"PASS","reason":"ok"}]}`)}})
	if _, _, err := c.analyze(t.Context(), []byte("fake-png"), "s", "mobile", "f", "c", ""); err != nil {
		t.Fatalf("analyze: %v", err)
	}
	req := fake.lastRequest()
	if req == nil {
		t.Fatal("no request captured")
	}
	system := req.Messages[0].Content[0].Text
	if strings.Contains(system, "structural evidence") {
		t.Errorf("system prompt mentions HTML without html enabled")
	}
	if len(req.Messages[len(req.Messages)-1].Content) != 2 {
		t.Errorf("user content = %d blocks, want 2 without html", len(req.Messages[len(req.Messages)-1].Content))
	}
}

func TestAnalyzeTruncatedBodyRetries(t *testing.T) {
	// A 200 with a truncated JSON body is a transient response — it must
	// retry within the budget, not hard-fail the run.
	good := validBody(`{"checks":[{"item":"x","verdict":"PASS","reason":"ok"}]}`)
	c, fake, _ := newTestClient(t, []fakeResp{
		{status: 200, body: `{"choices":[{"message":{"con`}, // truncated mid-body
		{status: 200, body: good},
	})
	got, _, err := c.analyze(t.Context(), []byte("fake-png"), "s", "mobile", "f", "c", "")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(got) != 1 || got[0].Verdict != "PASS" {
		t.Errorf("checks = %+v", got)
	}
	if fake.calls != 2 {
		t.Errorf("calls = %d, want 2 (retry after truncated body)", fake.calls)
	}
}

func TestAnalyzeTruncatedBodyExhausted(t *testing.T) {
	c, fake, _ := newTestClient(t, []fakeResp{
		{status: 200, body: `{"choices":`}, // truncated every time
		{status: 200, body: `{"choices":`},
		{status: 200, body: `{"choices":`},
		{status: 200, body: `{"choices":`},
	})
	_, _, err := c.analyze(t.Context(), []byte("fake-png"), "s", "mobile", "f", "c", "")
	if err == nil {
		t.Fatal("analyze succeeded, want error after retry budget")
	}
	if fake.calls != 4 {
		t.Errorf("calls = %d, want 4 (initial + 3 retries)", fake.calls)
	}
}

func TestParseChecksClampsToMax(t *testing.T) {
	// A 40-check response is clamped to maxChecksPerStep (32); a 30-check
	// response survives whole (the device "all" checklists are 28 items).
	build := func(n int) string {
		var sb strings.Builder
		sb.WriteString(`{"checks":[`)
		for i := 0; i < n; i++ {
			if i > 0 {
				sb.WriteString(",")
			}
			fmt.Fprintf(&sb, `{"item":"item %d","verdict":"PASS","reason":"r"}`, i)
		}
		sb.WriteString("]}")
		return sb.String()
	}

	got, err := parseChecks(build(40))
	if err != nil {
		t.Fatalf("parseChecks(40): %v", err)
	}
	if len(got) != maxChecksPerStep {
		t.Errorf("len = %d, want clamp to %d", len(got), maxChecksPerStep)
	}

	got, err = parseChecks(build(30))
	if err != nil {
		t.Fatalf("parseChecks(30): %v", err)
	}
	if len(got) != 30 {
		t.Errorf("len = %d, want 30 (no clamp under the cap)", len(got))
	}
}
