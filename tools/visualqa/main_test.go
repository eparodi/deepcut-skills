package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunValidationErrors(t *testing.T) {
	badFlow := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(badFlow, []byte(`{"feature":"x"`), 0o600); err != nil {
		t.Fatal(err)
	}
	validFlowPath := writeFlow(t, validFlow)

	// The missing-key cases would proceed to a browser launch if a key is
	// set in this environment — skip them so tests stay hermetic.
	keySet := os.Getenv("DEEPSEEK_API_KEY") != ""
	if _, err := loadEnvFile(envFilePath); err == nil {
		if vals, err := loadEnvFile(envFilePath); err == nil && vals["DEEPSEEK_API_KEY"] != "" {
			keySet = true
		}
	}

	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{"no url or flow", []string{}, exitValidation},
		{"both url and flow", []string{"--url", "http://x", "--flow", "f.json"}, exitValidation},
		{"bad device", []string{"--url", "http://x", "--device", "watch"}, exitValidation},
		{"missing flow file", []string{"--flow", filepath.Join(t.TempDir(), "nope.json")}, exitValidation},
		{"malformed flow json", []string{"--flow", badFlow}, exitValidation},
		{"unknown case", []string{"--flow", validFlowPath, "--case", "nope"}, exitValidation},
		{"unknown flag", []string{"--url", "http://x", "--bogus", "1"}, exitValidation},
		{"bad capture mode", []string{"--url", "http://x", "--capture-mode", "wide"}, exitValidation},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "missing key" && keySet {
				t.Skip("DEEPSEEK_API_KEY set in environment")
			}
			if code := run(tt.args); code != tt.wantCode {
				t.Errorf("run(%v) = %d, want %d", tt.args, code, tt.wantCode)
			}
		})
	}

	// Missing key exits 2 with instructions — but only when no key is
	// available, otherwise the run proceeds toward a browser launch.
	t.Run("missing key", func(t *testing.T) {
		if keySet {
			t.Skip("DEEPSEEK_API_KEY set in environment")
		}
		if code := run([]string{"--url", "http://localhost:3000"}); code != exitValidation {
			t.Errorf("run with no key = %d, want %d", code, exitValidation)
		}
	})
}

// The browser path itself (launch, emulation, screenshot, actions) is
// exercised by the skill's Test Task with a real Chrome — the repo has no
// CI and unit tests must stay offline.

func TestCaptureModeResolution(t *testing.T) {
	full := true
	falseVal := false
	tests := []struct {
		name     string
		step     flowStep
		defMode  string
		defHTML  bool
		wantMode string
		wantHTML bool
	}{
		{"per-step mode wins", flowStep{Mode: "full"}, "viewport", false, "full", false},
		{"run-level mode default", flowStep{}, "full", false, "full", false},
		{"per-step html true wins", flowStep{HTML: &full}, "viewport", false, "viewport", true},
		{"per-step html false overrides run-level", flowStep{HTML: &falseVal}, "viewport", true, "viewport", false},
		{"run-level html default", flowStep{}, "viewport", true, "viewport", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := captureModeFor(tt.step, tt.defMode); got != tt.wantMode {
				t.Errorf("captureModeFor = %q, want %q", got, tt.wantMode)
			}
			if got := htmlFor(tt.step, tt.defHTML); got != tt.wantHTML {
				t.Errorf("htmlFor = %v, want %v", got, tt.wantHTML)
			}
		})
	}
}
