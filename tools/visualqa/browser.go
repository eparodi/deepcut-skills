package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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
	mu      sync.Mutex
	diag    []string
}

func openBrowser(ctx context.Context, profile deviceProfile) (*browserSession, error) {
	u := launcher.New().Headless(true).Context(ctx).MustLaunch()
	b := rod.New().ControlURL(u).MustConnect()
	page := b.MustPage("about:blank")
	profile.apply(page)

	s := &browserSession{browser: b, page: page}
	// NOTE: the EachEvent wait func must be invoked exactly ONCE (via go).
	// It stops itself when the CDP connection closes, so close() never
	// calls it again — a second call blocks forever in the event loop.
	wait := page.EachEvent(
		func(e *proto.RuntimeConsoleAPICalled) {
			if e.Type != proto.RuntimeConsoleAPICalledTypeError {
				return
			}
			msg := "(no args)"
			if len(e.Args) > 0 {
				msg = e.Args[0].Value.String()
			}
			s.addDiag("console error: " + msg)
		},
		func(e *proto.NetworkLoadingFailed) {
			s.addDiag(fmt.Sprintf("failed request: %s (%s)", e.ErrorText, e.Type))
		},
	)
	go wait()
	return s, nil
}

func (s *browserSession) addDiag(line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.diag = append(s.diag, line)
}

func (s *browserSession) diagnostics() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.diag...)
}

func (s *browserSession) close() {
	// Bound the close: a Chrome that won't shut down (e.g. sandboxed envs)
	// must not block the report write or process exit. Closing the CDP
	// connection also ends the EachEvent loop goroutine. rod's leakless
	// layer kills the browser when the parent dies, so a lingering process
	// is cleaned up regardless.
	done := make(chan struct{})
	go func() {
		s.browser.MustClose()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		fmt.Fprintln(os.Stderr, "warn: browser close timed out")
	}
}

var unsafeName = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// maxFullPageHeight guards the Chrome capture itself (device px). Beyond it
// the capture is silently clipped, so taller pages fall back to a viewport
// capture with a diagnostic. 17679 device px was captured cleanly by Chrome
// 151 (2026-08-27); 32768 leaves comfortable headroom.
const maxFullPageHeight = 32768

func fullPageHeightOK(h int) bool { return h <= maxFullPageHeight }

// capture takes a screenshot (viewport or full-page), writes it to dir, and
// returns the filename plus the PNG bytes (passed straight to the vision
// client). Names are sanitized so a flow can never escape the run dir. The
// guard checks DEVICE pixels (CSS height × DPR), because that is what the
// capture and the vision API limits operate on. rod panics are recovered so
// a browser hiccup produces a FAILED run, never a process crash.
func (s *browserSession) capture(dir, name string, full bool, dpr float64) (file string, png []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("capture %q: %v", name, r)
		}
	}()
	if full {
		cssH := int(s.page.MustEval(`function() { return Math.max(document.documentElement.scrollHeight, document.body.scrollHeight) }`).Num())
		if devH := int(float64(cssH) * dpr); !fullPageHeightOK(devH) {
			s.addDiag(fmt.Sprintf("full-page capture skipped: page %d device px tall exceeds the %d px capture cap — captured viewport instead", devH, maxFullPageHeight))
			full = false
		}
	}
	var buf []byte
	if full {
		buf = s.page.MustScreenshotFullPage()
		// The vision API rejects images with a side > 8192px (400
		// "unsupported image", 2026-08-27). Downscale to fit so the whole
		// page still ships; the model sees ~800x800 of detail regardless,
		// and the HTML channel carries the structural precision.
		if scaled, err := fitMaxDimension(buf, maxVisionImageDimension); err == nil {
			buf = scaled
		} else {
			s.addDiag("full-page downscale failed: " + err.Error())
		}
	} else {
		buf = s.page.MustScreenshot()
	}
	file = unsafeName.ReplaceAllString(name, "-") + ".png"
	if err := os.WriteFile(filepath.Join(dir, file), buf, 0o644); err != nil {
		return "", nil, err
	}
	return file, buf, nil
}

// pageHTML fetches the serialized DOM for the structural-evidence channel
// (US8). rod panics on eval failure; recovered into a clean error.
func (s *browserSession) pageHTML() (html string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("fetch html: %v", r)
		}
	}()
	return s.page.MustEval(`function() { return document.documentElement.outerHTML }`).Str(), nil
}

// injectCookie sets a session cookie scoped to the origin BEFORE the first
// navigation (US9). The value lives only in the CDP session — never logged.
func (s *browserSession) injectCookie(name, value, origin string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("inject cookie: %v", r)
		}
	}()
	s.page.MustSetCookies(&proto.NetworkCookieParam{Name: name, Value: value, URL: origin})
	return nil
}

// currentPath returns the page's location.pathname — used to track visited
// pages so the explorer can be nudged toward unvisited sections.
func (s *browserSession) currentPath() (path string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("current path: %v", r)
		}
	}()
	return s.page.MustEval(`function() { return location.pathname }`).Str(), nil
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
		// Bound the whole click chain: rod has no default timeout, so a
		// missing or covered selector hangs the run forever (2026-09-05:
		// an htmx nav selector that didn't exist hung a QA run past its
		// run timeout). The timeout panic is recovered into a step
		// error by the caller — the run fails fast with the selector.
		p := s.page.Timeout(15 * time.Second)
		p.MustElement(step.Selector).MustScrollIntoView().MustClick()
	case "type":
		s.page.MustElement(step.Selector).MustScrollIntoView().MustInput(step.Text)
	case "scroll":
		switch step.To {
		case "top":
			s.page.MustEval(`function() { window.scrollTo(0, 0) }`)
		case "bottom":
			s.page.MustEval(`function() { window.scrollTo(0, document.body.scrollHeight) }`)
		default:
			s.page.MustElement(step.Selector).MustScrollIntoView()
		}
	case "back":
		s.page.MustNavigateBack()
		s.page.MustWaitLoad()
	case "wait":
		if step.MS > 0 {
			time.Sleep(time.Duration(step.MS) * time.Millisecond)
		} else {
			s.page.MustElement(step.Selector)
		}
	}
	return nil
}

// describeStep renders a step as a short human label for logs/report
// context. The case name is carried separately (stepResult.Case), so it is
// NOT repeated here.
func describeStep(i int, s flowStep) string {
	label := strings.TrimSpace(s.Name)
	if label == "" {
		label = s.Action
		if s.Selector != "" {
			label += " " + s.Selector
		}
	}
	return fmt.Sprintf("step %d (%s)", i+1, label)
}
