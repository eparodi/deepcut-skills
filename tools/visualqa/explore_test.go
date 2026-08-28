package main

import (
	"strings"
	"testing"
)

func TestParseExploreResponse(t *testing.T) {
	valid := `{"checks":[{"item":"x","verdict":"PASS","reason":"ok"}],"next_action":{"type":"click","selector":"#settings"}}`
	resp, err := parseExploreResponse(valid)
	if err != nil {
		t.Fatalf("parseExploreResponse: %v", err)
	}
	if len(resp.Checks) != 1 || resp.Checks[0].Verdict != "PASS" {
		t.Errorf("checks = %+v", resp.Checks)
	}
	if resp.NextAction == nil || resp.NextAction.Type != "click" || resp.NextAction.Selector != "#settings" {
		t.Errorf("next_action = %+v", resp.NextAction)
	}

	tests := []struct {
		name    string
		in      string
		wantErr string
	}{
		{"missing next_action", `{"checks":[{"item":"x","verdict":"PASS","reason":"ok"}]}`, "next_action"},
		{"missing checks", `{"next_action":{"type":"done"}}`, "checks"},
		{"unknown action type", `{"checks":[{"item":"x","verdict":"PASS","reason":"ok"}],"next_action":{"type":"hover","selector":"#x"}}`, "action"},
		{"click without selector", `{"checks":[{"item":"x","verdict":"PASS","reason":"ok"}],"next_action":{"type":"click"}}`, "selector"},
		{"goto without url", `{"checks":[{"item":"x","verdict":"PASS","reason":"ok"}],"next_action":{"type":"goto"}}`, "url"},
		{"scroll without target", `{"checks":[{"item":"x","verdict":"PASS","reason":"ok"}],"next_action":{"type":"scroll"}}`, "scroll"},
		{"invalid verdict", `{"checks":[{"item":"x","verdict":"MAYBE","reason":"ok"}],"next_action":{"type":"done"}}`, "verdict"},
		{"trailing garbage", valid + ` extra`, "trailing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseExploreResponse(tt.in)
			if err == nil {
				t.Fatal("parseExploreResponse succeeded, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestParseNextAction(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantT   string
		wantErr string
	}{
		{"valid click", `{"next_action":{"type":"click","selector":"#settings"}}`, "click", ""},
		{"valid done", `{"next_action":{"type":"done"}}`, "done", ""},
		{"missing next_action key", `{"action":{"type":"done"}}`, "", "unknown field"},
		{"unknown type", `{"next_action":{"type":"hover","selector":"#x"}}`, "", "action"},
		{"click without selector", `{"next_action":{"type":"click"}}`, "", "selector"},
		{"trailing garbage", `{"next_action":{"type":"done"}} extra`, "", "trailing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := parseNextAction(tt.in)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("parseNextAction succeeded, want error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseNextAction: %v", err)
			}
			if a.Type != tt.wantT {
				t.Errorf("type = %q, want %q", a.Type, tt.wantT)
			}
		})
	}
}

func TestNextActionTypeGate(t *testing.T) {
	base := `{"checks":[{"item":"x","verdict":"PASS","reason":"ok"}],"next_action":{"type":"%s","selector":"#email","text":"hello"}}`
	// type without testEnv is rejected at validation time
	withType := strings.Replace(base, "%s", "type", 1)
	resp, err := parseExploreResponse(withType)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := validateNextAction(resp.NextAction, false); err == nil {
		t.Error("type action accepted without --test-env, want error")
	}
	if err := validateNextAction(resp.NextAction, true); err != nil {
		t.Errorf("type action rejected with --test-env: %v", err)
	}

	// click passes the gate in both modes
	click := strings.Replace(base, `"selector":"#email","text":"hello"`, `"selector":"#email"`, 1)
	click = strings.Replace(click, "%s", "click", 1)
	resp, err = parseExploreResponse(click)
	if err != nil {
		t.Fatalf("parse click: %v", err)
	}
	if err := validateNextAction(resp.NextAction, false); err != nil {
		t.Errorf("click rejected without --test-env: %v", err)
	}
}

func TestSameOriginURL(t *testing.T) {
	origin := "http://127.0.0.1:18080"
	tests := []struct {
		url  string
		want bool
	}{
		{"/settings", true},
		{"settings", true},
		{"http://127.0.0.1:18080/trades", true},
		{"https://evil.com/x", false},
		{"http://127.0.0.1:9999/x", false},
		{"http://other.local/x", false},
	}
	for _, tt := range tests {
		if got := sameOrigin(origin, tt.url); got != tt.want {
			t.Errorf("sameOrigin(%q, %q) = %v, want %v", origin, tt.url, got, tt.want)
		}
	}
}

func TestAntiLoop(t *testing.T) {
	prev := &nextAction{Type: "click", Selector: "#settings"}
	same := &nextAction{Type: "click", Selector: "#settings"}
	diffSel := &nextAction{Type: "click", Selector: "#trades"}
	diffType := &nextAction{Type: "scroll", To: "bottom"}
	if !antiLoop(prev, same) {
		t.Error("identical action not flagged as loop")
	}
	if antiLoop(prev, diffSel) {
		t.Error("different selector flagged as loop")
	}
	if antiLoop(prev, diffType) {
		t.Error("different type flagged as loop")
	}
	if antiLoop(nil, same) {
		t.Error("first action flagged as loop")
	}
}
