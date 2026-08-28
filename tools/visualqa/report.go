package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// checkResult is one verdict for one checklist item.
type checkResult struct {
	Item    string `json:"item"`
	Verdict string `json:"verdict"`
	Reason  string `json:"reason"`
}

// stepResult is the per-step outcome: vision checks and/or a step error.
type stepResult struct {
	Case       string        `json:"case"`
	Step       string        `json:"step"`
	Screenshot string        `json:"screenshot,omitempty"`
	Checks     []checkResult `json:"checks,omitempty"`
	Error      string        `json:"error,omitempty"`
}

// runReport is the machine-readable findings payload; report.md is its
// human-readable rendering.
type runReport struct {
	RunID          string       `json:"run_id"`
	Feature        string       `json:"feature"`
	Target         string       `json:"target"`
	Device         string       `json:"device"`
	Model          string       `json:"model"`
	Date           string       `json:"date"`
	Status         string       `json:"status"`
	FailReason     string       `json:"fail_reason,omitempty"`
	Steps          int          `json:"steps"`
	MaxSteps       int          `json:"max_steps"`
	Screenshots    int          `json:"screenshots"`
	MaxScreenshots int          `json:"max_screenshots"`
	VisionTokens   int          `json:"vision_tokens"`
	StepResults    []stepResult `json:"step_results,omitempty"`
	Diagnostics    []string     `json:"diagnostics,omitempty"`
}

func (r *runReport) counts() (pass, fail, uncertain int) {
	for _, s := range r.StepResults {
		for _, c := range s.Checks {
			switch c.Verdict {
			case "PASS":
				pass++
			case "FAIL":
				fail++
			case "UNCERTAIN":
				uncertain++
			}
		}
	}
	return
}

// write renders report.md and findings.json into dir.
func (r *runReport) write(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "report.md"), []byte(r.markdown()), 0o644); err != nil {
		return err
	}
	fj, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "findings.json"), fj, 0o644)
}

func (r *runReport) markdown() string {
	pass, fail, uncertain := r.counts()
	var b strings.Builder
	fmt.Fprintf(&b, "# Visual QA Report — %s\n\n", r.Feature)
	fmt.Fprintf(&b, "Run: %s | Device: %s | Model: %s\n", r.RunID, r.Device, r.Model)
	fmt.Fprintf(&b, "URL/Flow: %s | Date: %s\n", r.Target, r.Date)
	fmt.Fprintf(&b, "Budget: %d/%d screenshots, %d/%d steps | Vision tokens: %d | Status: %s\n\n",
		r.Screenshots, r.MaxScreenshots, r.Steps, r.MaxSteps, r.VisionTokens, r.Status)
	fmt.Fprintf(&b, "Summary: %d PASS · %d FAIL · %d UNCERTAIN\n", pass, fail, uncertain)
	if r.Status == "FAILED" && r.FailReason != "" {
		fmt.Fprintf(&b, "Reason: %s\n", r.FailReason)
	}

	b.WriteString("\n## Findings\n")
	b.WriteString("| # | Case/Step | Check | Verdict | Reason |\n")
	b.WriteString("|---|---|---|---|---|\n")
	n := 0
	for _, s := range r.StepResults {
		for _, c := range s.Checks {
			n++
			fmt.Fprintf(&b, "| %d | %s / %s | %s | %s | %s |\n", n, s.Case, s.Step, c.Item, c.Verdict, c.Reason)
		}
	}
	if n == 0 {
		b.WriteString("| _no vision findings_ |\n")
	}

	b.WriteString("\n## Screenshots\n")
	anyShot := false
	for _, s := range r.StepResults {
		if s.Screenshot != "" {
			anyShot = true
			fmt.Fprintf(&b, "![%s](%s)\n", s.Step, s.Screenshot)
		}
	}
	if !anyShot {
		b.WriteString("(none)\n")
	}

	b.WriteString("\n## Errors\n")
	anyErr := false
	for _, s := range r.StepResults {
		if s.Error != "" {
			anyErr = true
			fmt.Fprintf(&b, "- %s / %s: %s\n", s.Case, s.Step, s.Error)
		}
	}
	if !anyErr {
		b.WriteString("(none)\n")
	}

	b.WriteString("\n## Diagnostics\n")
	if len(r.Diagnostics) == 0 {
		b.WriteString("(none)\n")
	}
	for _, d := range r.Diagnostics {
		fmt.Fprintf(&b, "- %s\n", d)
	}
	return b.String()
}
