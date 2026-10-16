package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/ysmood/gson"
)

func TestFullPageHeightGuard(t *testing.T) {
	tests := []struct {
		h    int
		want bool
	}{
		{0, true},
		{17679, true}, // proven live: Chrome 151 captured 1125x17679 fine
		{maxFullPageHeight, true},
		{maxFullPageHeight + 1, false},
	}
	for _, tt := range tests {
		if got := fullPageHeightOK(tt.h); got != tt.want {
			t.Errorf("fullPageHeightOK(%d) = %v, want %v", tt.h, got, tt.want)
		}
	}
}

// TestExceptionDiagnostic pins the report-line contract: message from the
// description's first line (or the thrown primitive's value), location from
// the first stack frame, 0-based CDP numbers shown 1-based like the browser
// console and error.stack.
func TestExceptionDiagnostic(t *testing.T) {
	frame := func(url string, line, col int) *proto.RuntimeStackTrace {
		return &proto.RuntimeStackTrace{CallFrames: []*proto.RuntimeCallFrame{
			{FunctionName: "boom", URL: url, LineNumber: line, ColumnNumber: col},
		}}
	}
	errorObj := func(desc string) *proto.RuntimeRemoteObject {
		return &proto.RuntimeRemoteObject{Type: "object", ClassName: "TypeError", Description: desc}
	}

	tests := []struct {
		name string
		d    *proto.RuntimeExceptionDetails
		want string
	}{
		{
			name: "error description plus first stack frame",
			d: &proto.RuntimeExceptionDetails{
				Text:       "Uncaught",
				Exception:  errorObj("TypeError: x is not a function\n    at boom (http://app.test/static/js/x.js:10:5)\n    at http://app.test/:7:1"),
				StackTrace: frame("http://app.test/static/js/x.js", 9, 4),
			},
			want: "uncaught exception: TypeError: x is not a function (http://app.test/static/js/x.js:10:5)",
		},
		{
			name: "thrown primitive carries a value instead of a description",
			d: &proto.RuntimeExceptionDetails{
				Text:       "Uncaught",
				Exception:  &proto.RuntimeRemoteObject{Type: "string", Value: gson.New("boom")},
				StackTrace: frame("http://app.test/", 2, 24),
			},
			want: "uncaught exception: boom (http://app.test/:3:25)",
		},
		{
			name: "no stack falls back to the exception location",
			d: &proto.RuntimeExceptionDetails{
				Text:         "Uncaught",
				URL:          "http://app.test/app.js",
				LineNumber:   41,
				ColumnNumber: 2,
				Exception:    errorObj("Error: gone"),
			},
			want: "uncaught exception: Error: gone (http://app.test/app.js:42:3)",
		},
		{
			name: "no location omits the parentheses",
			d: &proto.RuntimeExceptionDetails{
				Text:      "Uncaught",
				Exception: errorObj("TypeError: anonymous"),
			},
			want: "uncaught exception: TypeError: anonymous",
		},
		{
			name: "description and value absent fall back to the event text",
			d:    &proto.RuntimeExceptionDetails{Text: "Uncaught"},
			want: "uncaught exception: Uncaught",
		},
		{
			name: "nil details",
			d:    nil,
			want: "uncaught exception: (no details)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exceptionDiagnostic(tt.d); got != tt.want {
				t.Errorf("exceptionDiagnostic() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestOpenBrowserCapturesUncaughtException drives a local page that throws
// and asserts the session reports it. It launches a real browser but needs no
// network: when no browser is cached locally it skips instead of letting
// rod's launcher download one (unit tests must stay offline).
func TestOpenBrowserCapturesUncaughtException(t *testing.T) {
	if err := launcher.NewBrowser().Validate(); err != nil {
		t.Skipf("no local browser available: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html><html><body>
<script>
function qaBoom() { throw new TypeError("qa sample failure"); }
qaBoom();
</script>
</body></html>`)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	sess, err := openBrowser(ctx, devicePresets["desktop"])
	if err != nil {
		t.Fatalf("openBrowser: %v", err)
	}
	defer sess.close()

	if err := sess.execStep(flowStep{Action: "goto", URL: srv.URL}); err != nil {
		t.Fatalf("goto: %v", err)
	}

	got := waitForDiagnostic(t, sess, "uncaught exception: ")
	wantPrefix := "uncaught exception: TypeError: qa sample failure ("
	if !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("diagnostic = %q, want prefix %q", got, wantPrefix)
	}
	// The location's file is the page that threw (its URL resolves to the
	// server root); line/col are Chrome's (converted to 1-based by the
	// handler), so only their shape is pinned.
	loc := regexp.MustCompile(`\(` + regexp.QuoteMeta(srv.URL) + `/:\d+:\d+\)$`)
	if !loc.MatchString(got) {
		t.Errorf("diagnostic %q lacks a first-frame (url:line:col) matching %s", got, loc)
	}
}

// waitForDiagnostic polls until a stored diagnostic starts with prefix.
// Exception events arrive asynchronously after load, so a single read right
// after navigation can race them.
func waitForDiagnostic(t *testing.T, sess *browserSession, prefix string) string {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		for _, d := range sess.diagnostics() {
			if strings.HasPrefix(d, prefix) {
				return d
			}
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("no diagnostic with prefix %q within 10s; got %q", prefix, sess.diagnostics())
	return ""
}
