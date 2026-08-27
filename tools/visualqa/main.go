package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	exitOK         = 0
	exitValidation = 2
	exitProvider   = 3
	exitBudget     = 4
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// run parses flags, validates everything it can (flow, device, key), then
// executes the browser session. Validation errors exit 2 before a browser
// ever launches.
func run(args []string) int {
	fs := flag.NewFlagSet("visualqa", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	urlFlag := fs.String("url", "", "URL to verify (one-shot mode; mutually exclusive with --flow)")
	flowFlag := fs.String("flow", "", "path to a feature flow JSON (flow mode)")
	caseFlag := fs.String("case", "", "run only this named case from the flow")
	deviceFlag := fs.String("device", "mobile", "device profile: mobile | tablet | desktop")
	checklistFlag := fs.String("checklist", "", "custom QA checklist (text or @file path)")
	outFlag := fs.String("out", "artifacts/visualqa", "output root directory")
	maxStepsFlag := fs.Int("max-steps", 15, "hard cap on executed steps per run")
	maxShotsFlag := fs.Int("max-screenshots", 12, "hard cap on screenshots per run")
	timeoutFlag := fs.Duration("timeout", 5*time.Minute, "per-run timeout")
	retriesFlag := fs.Int("retries", 3, "vision-call retry budget for 429/500/503")
	modelFlag := fs.String("model", "", "vision model (default deepseek-v4-flash-vision-exp)")
	if err := fs.Parse(args); err != nil {
		return exitValidation
	}

	if (*urlFlag == "") == (*flowFlag == "") {
		fmt.Fprintln(os.Stderr, "usage: exactly one of --url or --flow is required")
		return exitValidation
	}
	profile, ok := deviceProfileFor(*deviceFlag)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown device %q (mobile | tablet | desktop)\n", *deviceFlag)
		return exitValidation
	}

	// Load the flow (or synthesize the one-shot case) before any network work.
	var flow *flowFile
	feature := "one-shot"
	target := *urlFlag
	var execCases []execCase
	if *flowFlag != "" {
		f, err := loadFlow(*flowFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitValidation
		}
		flow = f
		feature = f.Feature
		target = fmt.Sprintf("%s (base %s)", *flowFlag, f.BaseURL)
		if *caseFlag != "" {
			c, err := f.resolveCase(*caseFlag)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return exitValidation
			}
			execCases = []execCase{{name: c.Name, steps: c.Steps}}
		} else {
			for i := range f.Cases {
				execCases = append(execCases, execCase{name: f.Cases[i].Name, steps: f.Cases[i].Steps})
			}
		}
	} else {
		execCases = []execCase{{
			name: "one-shot",
			steps: []flowStep{
				{Action: "goto", URL: *urlFlag},
				{Action: "screenshot", Name: "page"},
			},
		}}
	}

	// Key + model resolution: flag > env > .env.visualqa > default.
	fileVals, err := loadEnvFile(envFilePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitValidation
	}
	key := resolveKey(os.Getenv("DEEPSEEK_API_KEY"), fileVals["DEEPSEEK_API_KEY"])
	if key == "" {
		fmt.Fprintln(os.Stderr, "DEEPSEEK_API_KEY not found: set it in .env.visualqa (repo root) or the DEEPSEEK_API_KEY environment variable")
		return exitValidation
	}
	model := resolveModel(*modelFlag, os.Getenv("DEEPSEEK_VISION_MODEL"), fileVals["DEEPSEEK_VISION_MODEL"])

	checklist := *checklistFlag
	if strings.HasPrefix(checklist, "@") {
		b, err := os.ReadFile(strings.TrimPrefix(checklist, "@"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "checklist file: %v\n", err)
			return exitValidation
		}
		checklist = string(b)
	}
	if checklist == "" {
		checklist = strings.Join(defaultChecklist(), "\n")
	}

	vision := &visionClient{
		baseURL: "https://api.deepseek.com",
		apiKey:  key,
		model:   model,
		retries: *retriesFlag,
		timeout: 30 * time.Second,
		http:    &http.Client{Timeout: 30 * time.Second},
		backoff: defaultBackoff,
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeoutFlag)
	defer cancel()

	return execute(ctx, executeParams{
		url:            *urlFlag,
		flow:           flow,
		feature:        feature,
		target:         target,
		profile:        profile,
		execCases:      execCases,
		checklist:      checklist,
		outRoot:        *outFlag,
		maxSteps:       *maxStepsFlag,
		maxScreenshots: *maxShotsFlag,
		vision:         vision,
	})
}

type execCase struct {
	name  string
	steps []flowStep
}

type executeParams struct {
	url            string
	flow           *flowFile
	feature        string
	target         string
	profile        deviceProfile
	execCases      []execCase
	checklist      string
	outRoot        string
	maxSteps       int
	maxScreenshots int
	vision         *visionClient
}

func execute(ctx context.Context, p executeParams) int {
	runID := "run-" + time.Now().UTC().Format("20060102T150405Z")
	dir := p.outRoot + "/" + runID

	report := &runReport{
		RunID:          runID,
		Feature:        p.feature,
		Target:         p.target,
		Device:         p.profile.Name,
		Model:          p.vision.model,
		Date:           time.Now().UTC().Format(time.RFC3339),
		Status:         "COMPLETED",
		MaxSteps:       p.maxSteps,
		MaxScreenshots: p.maxScreenshots,
	}
	defer func() {
		if err := report.write(dir); err != nil {
			fmt.Fprintf(os.Stderr, "write report: %v\n", err)
		}
	}()

	sess, err := openBrowser(ctx, p.profile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "browser: %v\n", err)
		report.Status = "FAILED"
		report.FailReason = "browser launch failed: " + err.Error()
		return exitProvider
	}
	defer sess.close()
	report.Diagnostics = append(report.Diagnostics, sess.diag...)

	stepsRun := 0
	shots := 0
	for _, ec := range p.execCases {
		for i, step := range ec.steps {
			// budget checks first — a run may never exceed its caps.
			stepsRun++
			if stepsRun > p.maxSteps {
				report.Status = "FAILED"
				report.FailReason = fmt.Sprintf("budget exceeded: %d steps (> %d)", stepsRun, p.maxSteps)
				return exitBudget
			}
			if err := ctx.Err(); err != nil {
				report.Status = "FAILED"
				report.FailReason = "timeout: " + err.Error()
				return exitBudget
			}

			label := describeStep(ec.name, i, step)
			fmt.Fprintf(os.Stderr, "[%d/%d] %s\n", stepsRun, p.maxSteps, label)

			res := stepResult{Case: ec.name, Step: label}
			if step.Action == "goto" && p.flow != nil {
				step.URL, err = p.flow.resolveURL(step.URL)
				if err != nil {
					res.Error = err.Error()
					report.Status = "FAILED"
					report.FailReason = fmt.Sprintf("step %d in case %q: %s", i+1, ec.name, err.Error())
					report.StepResults = append(report.StepResults, res)
					report.Steps = stepsRun
					return exitProvider
				}
			}
			if err := sess.execStep(step); err != nil {
				res.Error = err.Error()
				report.Status = "FAILED"
				report.FailReason = fmt.Sprintf("case %q step %d: %s", ec.name, i+1, err.Error())
				report.StepResults = append(report.StepResults, res)
				report.Steps = stepsRun
				return exitProvider
			}

			capture := step.Action == "screenshot" || step.Capture
			if capture {
				shots++
				if shots > p.maxScreenshots {
					report.Status = "FAILED"
					report.FailReason = fmt.Sprintf("budget exceeded: %d screenshots (> %d)", shots, p.maxScreenshots)
					report.Steps = stepsRun
					return exitBudget
				}
				name := step.Name
				if name == "" {
					name = step.Action
				}
				file, png, err := sess.capture(dir, name)
				if err != nil {
					res.Error = "capture: " + err.Error()
					report.Status = "FAILED"
					report.FailReason = res.Error
					report.StepResults = append(report.StepResults, res)
					report.Steps = stepsRun
					return exitProvider
				}
				res.Screenshot = file

				checks, usage, err := p.vision.analyze(ctx, png, res.Step, p.profile.Name, p.feature, p.checklist)
				report.VisionTokens += usage.PromptTokens
				if err != nil {
					report.Status = "FAILED"
					report.FailReason = "vision: " + err.Error()
					report.StepResults = append(report.StepResults, res)
					report.Steps = stepsRun
					return exitProvider
				}
				res.Checks = checks
			}
			report.StepResults = append(report.StepResults, res)
			report.Steps = stepsRun
		}
	}
	report.Screenshots = shots
	report.Diagnostics = append(report.Diagnostics, sess.diag...)

	fmt.Fprintf(os.Stderr, "report: %s/report.md\n", dir)
	return exitOK
}
