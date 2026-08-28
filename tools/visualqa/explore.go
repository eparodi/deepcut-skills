package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// exploreActionType is the bounded vocabulary the model may propose (US10
// AC2). "type" is gated behind --test-env at validation time.
const (
	exploreClick  = "click"
	exploreScroll = "scroll"
	exploreBack   = "back"
	exploreGoto   = "goto"
	exploreDone   = "done"
	exploreType   = "type"
)

// nextAction is what the model proposes after grading a page.
type nextAction struct {
	Type     string `json:"type"`
	Selector string `json:"selector,omitempty"`
	To       string `json:"to,omitempty"`
	URL      string `json:"url,omitempty"`
	Text     string `json:"text,omitempty"`
}

// targetKey is the identity used by the anti-loop guard: same type + same
// target means the model is stuck.
func (a *nextAction) targetKey() string {
	return a.Type + "|" + a.Selector + "|" + a.To + "|" + a.URL
}

func (a *nextAction) describe() string {
	switch a.Type {
	case exploreClick, exploreType:
		return a.Type + " " + a.Selector
	case exploreScroll:
		if a.To != "" {
			return "scroll " + a.To
		}
		return "scroll " + a.Selector
	case exploreGoto:
		return "goto " + a.URL
	default:
		return a.Type
	}
}

// exploreResponse is the single vision-call payload: checklist verdicts for
// the current page PLUS the next action (US10 design: grade-as-you-go).
type exploreResponse struct {
	Checks     []checkResult `json:"checks"`
	NextAction *nextAction   `json:"next_action"`
}

// parseExploreResponse strictly parses the explorer JSON-mode response. The
// same ladder semantics apply upstream: repair → one re-ask → loop end.
func parseExploreResponse(text string) (*exploreResponse, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("empty vision response content (JSON-mode quirk)")
	}
	var vr exploreResponse
	dec := json.NewDecoder(bytes.NewReader([]byte(text)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&vr); err != nil {
		return nil, err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return nil, errors.New("trailing content after JSON")
	}
	if len(vr.Checks) == 0 {
		return nil, errors.New("explorer response has no checks")
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
	if vr.NextAction == nil {
		return nil, errors.New("explorer response has no next_action")
	}
	if err := validateNextAction(vr.NextAction, true); err != nil {
		return nil, fmt.Errorf("next_action: %w", err)
	}
	return &vr, nil
}

// parseNextAction strictly parses an action-only nudge response (the salvage
// path in analyzeExplore): the checks were accepted, only the action is
// asked for. Same strictness: one value, no trailing content.
func parseNextAction(text string) (*nextAction, error) {
	var vr struct {
		NextAction *nextAction `json:"next_action"`
	}
	dec := json.NewDecoder(bytes.NewReader([]byte(text)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&vr); err != nil {
		return nil, err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return nil, errors.New("trailing content after JSON")
	}
	if vr.NextAction == nil {
		return nil, errors.New("no next_action in response")
	}
	if err := validateNextAction(vr.NextAction, true); err != nil {
		return nil, err
	}
	return vr.NextAction, nil
}

// validateNextAction enforces the vocabulary and per-action required args.
// testEnv gates the "type" (form-filling) action: without --test-env the
// model may only navigate and read (US10 AC2).
func validateNextAction(a *nextAction, testEnv bool) error {
	switch a.Type {
	case exploreClick:
		if strings.TrimSpace(a.Selector) == "" {
			return errors.New("click requires selector")
		}
	case exploreScroll:
		if a.To == "" && a.Selector == "" {
			return errors.New("scroll requires to or selector")
		}
	case exploreBack, exploreDone:
	case exploreGoto:
		if strings.TrimSpace(a.URL) == "" {
			return errors.New("goto requires url")
		}
	case exploreType:
		if !testEnv {
			return errors.New("type action requires --test-env (form-filling mutates state)")
		}
		if strings.TrimSpace(a.Selector) == "" || strings.TrimSpace(a.Text) == "" {
			return errors.New("type requires selector and text")
		}
	default:
		return fmt.Errorf("unknown action %q", a.Type)
	}
	return nil
}

// sameOrigin reports whether u is a relative URL (resolved against the
// origin later) or an absolute URL sharing the origin's scheme+host+port.
// Absolute URLs leaving the origin are rejected (US10 AC3).
func sameOrigin(origin, u string) bool {
	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}
	if !parsed.IsAbs() {
		return true
	}
	o, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return parsed.Scheme == o.Scheme && parsed.Host == o.Host
}

// originOf returns the scheme://host[:port] of a target URL (the cookie
// scope and the same-origin base for exploration).
func originOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// resolveAgainstOrigin resolves a (same-origin relative) action URL against
// the exploration origin, mirroring the flow loader's base_url resolution.
func resolveAgainstOrigin(origin, raw string) string {
	base, err := url.Parse(origin)
	if err != nil {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return base.ResolveReference(u).String()
}

// flowStepFromAction maps a validated exploration action onto the flowStep
// executor (click/scroll/type are already supported; back and goto were
// added for exploration).
func flowStepFromAction(a *nextAction, origin string) flowStep {
	switch a.Type {
	case exploreClick:
		return flowStep{Action: "click", Selector: a.Selector}
	case exploreScroll:
		return flowStep{Action: "scroll", To: a.To, Selector: a.Selector}
	case exploreBack:
		return flowStep{Action: "back"}
	case exploreGoto:
		return flowStep{Action: "goto", URL: resolveAgainstOrigin(origin, a.URL)}
	case exploreType:
		return flowStep{Action: "type", Selector: a.Selector, Text: a.Text}
	}
	return flowStep{Action: a.Type}
}

// antiLoop flags an identical (type, target) as the immediately previous
// step — the model is stuck (US10 AC4).
func antiLoop(prev, cur *nextAction) bool {
	if prev == nil || cur == nil {
		return false
	}
	return prev.targetKey() == cur.targetKey()
}
