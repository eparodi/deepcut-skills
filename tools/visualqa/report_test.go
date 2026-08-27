package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleReport() *runReport {
	return &runReport{
		RunID:          "run-20260826T101500Z",
		Feature:        "checkout",
		Target:         "flows/checkout.json (base http://localhost:3000)",
		Device:         "mobile",
		Model:          "deepseek-v4-flash-vision-exp",
		Date:           "2026-08-26T10:15:00Z",
		Status:         "COMPLETED",
		Steps:          2,
		MaxSteps:       15,
		Screenshots:    2,
		MaxScreenshots: 12,
		VisionTokens:   768,
		StepResults: []stepResult{
			{
				Case:       "happy path",
				Step:       "cart-empty",
				Screenshot: "s1.png",
				Checks: []checkResult{
					{Item: "no horizontal overflow", Verdict: "PASS", Reason: "fits"},
					{Item: "tap targets", Verdict: "FAIL", Reason: "bottom bar covers the CTA"},
				},
			},
			{
				Case:       "happy path",
				Step:       "cart-filled",
				Screenshot: "s2.png",
				Checks: []checkResult{
					{Item: "images load", Verdict: "UNCERTAIN", Reason: "not enough detail"},
				},
			},
		},
		Diagnostics: []string{"console error: Uncaught TypeError at checkout.js:42"},
	}
}

func TestReportWrite(t *testing.T) {
	dir := t.TempDir()
	r := sampleReport()
	if err := r.write(dir); err != nil {
		t.Fatalf("write: %v", err)
	}

	md, err := os.ReadFile(filepath.Join(dir, "report.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(md)

	wantContains := []string{
		"# Visual QA Report — checkout",
		"Run: run-20260826T101500Z",
		"Device: mobile",
		"Model: deepseek-v4-flash-vision-exp",
		"Status: COMPLETED",
		"Budget: 2/12 screenshots, 2/15 steps",
		"Vision tokens: 768",
		"Summary: 1 PASS · 1 FAIL · 1 UNCERTAIN",
		"| happy path / cart-empty | tap targets | FAIL | bottom bar covers the CTA |",
		"![cart-empty](s1.png)",
		"## Diagnostics",
		"console error: Uncaught TypeError at checkout.js:42",
	}
	for _, w := range wantContains {
		if !strings.Contains(s, w) {
			t.Errorf("report.md missing %q", w)
		}
	}

	// findings.json
	fj, err := os.ReadFile(filepath.Join(dir, "findings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var parsed runReport
	if err := json.Unmarshal(fj, &parsed); err != nil {
		t.Fatalf("findings.json is not valid JSON: %v", err)
	}
	if parsed.RunID != r.RunID || parsed.Device != "mobile" || len(parsed.StepResults) != 2 {
		t.Errorf("parsed = %+v", parsed)
	}
	if parsed.StepResults[0].Checks[1].Verdict != "FAIL" {
		t.Errorf("findings.json check verdict = %q", parsed.StepResults[0].Checks[1].Verdict)
	}
	if parsed.Diagnostics[0] != r.Diagnostics[0] {
		t.Errorf("findings.json diagnostics = %v", parsed.Diagnostics)
	}
}

func TestReportWriteFailedStatus(t *testing.T) {
	dir := t.TempDir()
	r := sampleReport()
	r.Status = "FAILED"
	r.FailReason = "step 3 (click #add-to-cart): element not found"
	if err := r.write(dir); err != nil {
		t.Fatal(err)
	}
	md, err := os.ReadFile(filepath.Join(dir, "report.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(md)
	if !strings.Contains(s, "Status: FAILED") {
		t.Errorf("missing FAILED status")
	}
	if !strings.Contains(s, r.FailReason) {
		t.Errorf("missing fail reason %q", r.FailReason)
	}
}

func TestReportWriteNoDiagnostics(t *testing.T) {
	dir := t.TempDir()
	r := sampleReport()
	r.Diagnostics = nil
	if err := r.write(dir); err != nil {
		t.Fatal(err)
	}
	md, err := os.ReadFile(filepath.Join(dir, "report.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(md), "(none)") {
		t.Errorf("expected a (none) marker for empty diagnostics")
	}
}
