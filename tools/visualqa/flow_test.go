package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validFlow = `{
  "feature": "checkout",
  "base_url": "http://localhost:3000",
  "cases": [
    {
      "name": "happy path",
      "steps": [
        { "action": "goto", "url": "/cart" },
        { "action": "screenshot", "name": "cart-empty" },
        { "action": "click", "selector": "#add-to-cart", "capture": true },
        { "action": "wait", "ms": 800 },
        { "action": "screenshot", "name": "cart-filled" },
        { "action": "scroll", "to": "bottom", "capture": true }
      ]
    },
    {
      "name": "empty cart guard",
      "steps": [
        { "action": "goto", "url": "/checkout" },
        { "action": "screenshot", "name": "guard" }
      ]
    }
  ]
}`

func writeFlow(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "flow.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFlowHappyPath(t *testing.T) {
	path := writeFlow(t, validFlow)
	f, err := loadFlow(path)
	if err != nil {
		t.Fatalf("loadFlow: %v", err)
	}
	if f.Feature != "checkout" {
		t.Errorf("Feature = %q, want checkout", f.Feature)
	}
	if f.BaseURL != "http://localhost:3000" {
		t.Errorf("BaseURL = %q", f.BaseURL)
	}
	if len(f.Cases) != 2 {
		t.Fatalf("len(Cases) = %d, want 2", len(f.Cases))
	}
	if f.Cases[0].Name != "happy path" || len(f.Cases[0].Steps) != 6 {
		t.Errorf("case 0 = %+v, want 6 steps", f.Cases[0])
	}
	if f.Cases[1].Steps[1].Action != "screenshot" || f.Cases[1].Steps[1].Name != "guard" {
		t.Errorf("case 1 step 1 = %+v", f.Cases[1].Steps[1])
	}
}

func TestLoadFlowErrors(t *testing.T) {
	replace := func(old, new string) string { return strings.Replace(validFlow, old, new, 1) }

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "malformed JSON",
			content: `{"feature": "x",`,
			wantErr: "flow",
		},
		{
			name:    "unknown top-level field",
			content: replace(`"cases": [`, `"evil": 1, "cases": [`),
			wantErr: "unknown field",
		},
		{
			name:    "missing feature",
			content: replace(`"feature": "checkout",`, ``),
			wantErr: "feature",
		},
		{
			name:    "base_url not absolute",
			content: replace(`"base_url": "http://localhost:3000"`, `"base_url": "/relative"`),
			wantErr: "base_url",
		},
		{
			name:    "no cases",
			content: replace(`"cases": [`, `"cases": []`),
			wantErr: "cases",
		},
		{
			name:    "missing case name",
			content: replace(`"name": "happy path",`, ``),
			wantErr: "name",
		},
		{
			name:    "duplicate case names",
			content: replace(`"name": "empty cart guard",`, `"name": "happy path",`),
			wantErr: "duplicate",
		},
		{
			name:    "empty steps",
			content: replace(`"steps": [`, `"steps": []`),
			wantErr: "steps",
		},
		{
			name:    "unknown action",
			content: replace(`"action": "screenshot", "name": "cart-empty"`, `"action": "hover", "selector": "#x"`),
			wantErr: "hover",
		},
		{
			name:    "unknown step field",
			content: replace(`"action": "click", "selector": "#add-to-cart", "capture": true`, `"action": "click", "selector": "#add-to-cart", "evil": 1`),
			wantErr: "unknown field",
		},
		{
			name:    "capture not allowed on screenshot",
			content: replace(`"action": "screenshot", "name": "cart-empty"`, `"action": "screenshot", "name": "cart-empty", "capture": true`),
			wantErr: "unknown field",
		},
		{
			name:    "goto without url",
			content: replace(`{ "action": "goto", "url": "/cart" },`, `{ "action": "goto" },`),
			wantErr: "url",
		},
		{
			name:    "click without selector",
			content: replace(`"action": "click", "selector": "#add-to-cart", "capture": true`, `"action": "click", "capture": true`),
			wantErr: "selector",
		},
		{
			name:    "type without text",
			content: `{"feature":"f","base_url":"http://localhost:3000","cases":[{"name":"c","steps":[{"action":"goto","url":"/"},{"action":"type","selector":"#email"}]}]}`,
			wantErr: "text",
		},
		{
			name:    "wait without ms or selector",
			content: `{"feature":"f","base_url":"http://localhost:3000","cases":[{"name":"c","steps":[{"action":"wait"}]}]}`,
			wantErr: "wait",
		},
		{
			name:    "wait with both ms and selector",
			content: `{"feature":"f","base_url":"http://localhost:3000","cases":[{"name":"c","steps":[{"action":"wait","ms":100,"selector":"#x"}]}]}`,
			wantErr: "wait",
		},
		{
			name:    "scroll with invalid target",
			content: `{"feature":"f","base_url":"http://localhost:3000","cases":[{"name":"c","steps":[{"action":"scroll","to":"sideways"}]}]}`,
			wantErr: "scroll",
		},
		{
			name:    "scroll with both to and selector",
			content: `{"feature":"f","base_url":"http://localhost:3000","cases":[{"name":"c","steps":[{"action":"scroll","to":"top","selector":"#x"}]}]}`,
			wantErr: "scroll",
		},
		{
			name:    "screenshot without name",
			content: `{"feature":"f","base_url":"http://localhost:3000","cases":[{"name":"c","steps":[{"action":"screenshot"}]}]}`,
			wantErr: "name",
		},
		{
			name:    "duplicate screenshot names in a case",
			content: `{"feature":"f","base_url":"http://localhost:3000","cases":[{"name":"c","steps":[{"action":"screenshot","name":"dup"},{"action":"screenshot","name":"dup"}]}]}`,
			wantErr: "duplicate",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFlow(t, tt.content)
			_, err := loadFlow(path)
			if err == nil {
				t.Fatal("loadFlow succeeded, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestResolveURL(t *testing.T) {
	f := &flowFile{Feature: "f", BaseURL: "http://localhost:3000"}
	tests := []struct {
		raw  string
		want string
	}{
		{"/cart", "http://localhost:3000/cart"},
		{"/cart?x=1", "http://localhost:3000/cart?x=1"},
		{"https://example.com/page", "https://example.com/page"},
	}
	for _, tt := range tests {
		got, err := f.resolveURL(tt.raw)
		if err != nil {
			t.Errorf("resolveURL(%q): %v", tt.raw, err)
			continue
		}
		if got != tt.want {
			t.Errorf("resolveURL(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

func TestLoadFlowCaptureModes(t *testing.T) {
	content := `{"feature":"f","base_url":"http://localhost:3000","cases":[{"name":"c","steps":[
		{"action":"goto","url":"/"},
		{"action":"screenshot","name":"full","mode":"full","html":true},
		{"action":"screenshot","name":"vp","mode":"viewport"},
		{"action":"click","selector":"#x","capture":true,"mode":"full"},
		{"action":"scroll","to":"bottom","capture":true,"html":false}
	]}]}`
	f, err := loadFlow(writeFlow(t, content))
	if err != nil {
		t.Fatalf("loadFlow: %v", err)
	}
	steps := f.Cases[0].Steps
	if steps[1].Mode != "full" || steps[1].HTML == nil || !*steps[1].HTML {
		t.Errorf("screenshot full step = mode %q html %v, want mode full html true", steps[1].Mode, steps[1].HTML)
	}
	if steps[2].Mode != "viewport" || steps[2].HTML != nil {
		t.Errorf("screenshot vp step = mode %q html %v, want mode viewport html unset", steps[2].Mode, steps[2].HTML)
	}
	if steps[3].Mode != "full" {
		t.Errorf("click step mode = %q, want full", steps[3].Mode)
	}
	if steps[4].HTML == nil || *steps[4].HTML {
		t.Errorf("scroll step html = %v, want explicit false", steps[4].HTML)
	}
}

func TestLoadFlowCaptureModeErrors(t *testing.T) {
	step := func(s string) string {
		return `{"feature":"f","base_url":"http://localhost:3000","cases":[{"name":"c","steps":[` + s + `]}]}`
	}
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "invalid mode value",
			content: step(`{"action":"screenshot","name":"x","mode":"wide"}`),
			wantErr: "mode",
		},
		{
			name:    "html non-boolean",
			content: step(`{"action":"screenshot","name":"x","html":"yes"}`),
			wantErr: "html",
		},
		{
			name:    "mode on goto",
			content: step(`{"action":"goto","url":"/","mode":"full"}`),
			wantErr: "unknown field",
		},
		{
			name:    "html on wait",
			content: step(`{"action":"wait","ms":100,"html":true}`),
			wantErr: "unknown field",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadFlow(writeFlow(t, tt.content))
			if err == nil {
				t.Fatal("loadFlow succeeded, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestResolveCase(t *testing.T) {
	path := writeFlow(t, validFlow)
	f, err := loadFlow(path)
	if err != nil {
		t.Fatal(err)
	}
	c, err := f.resolveCase("empty cart guard")
	if err != nil {
		t.Fatalf("resolveCase: %v", err)
	}
	if c.Name != "empty cart guard" {
		t.Errorf("got case %q", c.Name)
	}
	if _, err := f.resolveCase("nope"); err == nil {
		t.Fatal("resolveCase(nope) succeeded, want error")
	}
}
