package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// browserSession wraps the rod browser + page and collects CDP diagnostics
// (console errors, failed network requests) for the report.
type browserSession struct {
	browser *rod.Browser
	page    *rod.Page
	diag    []string
	stopEvt func()
}

func openBrowser(ctx context.Context, profile deviceProfile) (*browserSession, error) {
	u := launcher.New().Headless(true).Context(ctx).MustLaunch()
	b := rod.New().ControlURL(u).MustConnect()
	page := b.MustPage("about:blank")
	profile.apply(page)

	s := &browserSession{browser: b, page: page}
	wait := page.EachEvent(
		func(e *proto.RuntimeConsoleAPICalled) {
			if e.Type != proto.RuntimeConsoleAPICalledTypeError {
				return
			}
			msg := "(no args)"
			if len(e.Args) > 0 {
				msg = e.Args[0].Value.String()
			}
			s.diag = append(s.diag, "console error: "+msg)
		},
		func(e *proto.NetworkLoadingFailed) {
			s.diag = append(s.diag, fmt.Sprintf("failed request: %s (%s)", e.ErrorText, e.Type))
		},
	)
	s.stopEvt = wait
	go wait()
	return s, nil
}

func (s *browserSession) close() {
	if s.stopEvt != nil {
		s.stopEvt()
	}
	s.browser.MustClose()
}

var unsafeName = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// capture takes a viewport screenshot, writes it to dir, and returns the
// filename plus the PNG bytes (passed straight to the vision client). Names
// are sanitized so a flow can never escape the run dir.
func (s *browserSession) capture(dir, name string) (string, []byte, error) {
	buf := s.page.MustScreenshot()
	file := unsafeName.ReplaceAllString(name, "-") + ".png"
	if err := os.WriteFile(filepath.Join(dir, file), buf, 0o644); err != nil {
		return "", nil, err
	}
	return file, buf, nil
}

// execStep runs one flow action. Failures are returned as errors (rod Must*
// panics are converted), so the caller can fail fast with a clean message.
func (s *browserSession) execStep(step flowStep) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("step %s: %v", step.Action, r)
		}
	}()
	switch step.Action {
	case "goto":
		s.page.MustNavigate(step.URL)
		s.page.MustWaitLoad()
	case "click":
		s.page.MustElement(step.Selector).MustScrollIntoView().MustClick()
	case "type":
		s.page.MustElement(step.Selector).MustScrollIntoView().MustInput(step.Text)
	case "scroll":
		switch step.To {
		case "top":
			s.page.MustEval(`window.scrollTo(0, 0)`)
		case "bottom":
			s.page.MustEval(`window.scrollTo(0, document.body.scrollHeight)`)
		default:
			s.page.MustElement(step.Selector).MustScrollIntoView()
		}
	case "wait":
		if step.MS > 0 {
			time.Sleep(time.Duration(step.MS) * time.Millisecond)
		} else {
			s.page.MustElement(step.Selector)
		}
	}
	return nil
}

// describe renders a step as a short human label for logs/report context.
func describeStep(caseName string, i int, s flowStep) string {
	label := strings.TrimSpace(s.Name)
	if label == "" {
		label = s.Action
		if s.Selector != "" {
			label += " " + s.Selector
		}
	}
	return fmt.Sprintf("%s / step %d (%s)", caseName, i+1, label)
}
