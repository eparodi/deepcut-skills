package main

import (
	"strings"
	"testing"
)

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "strips script style and comments",
			raw:  `<html><head><script>var x=1;</script><style>body{color:red}</style><!-- keep out --></head><body>Hello</body></html>`,
			want: `<html><head></head><body>Hello</body></html>`,
		},
		{
			name: "case-insensitive tags",
			raw:  `<SCRIPT>alert(1)</SCRIPT><STYLE>.x{}</STYLE>ok`,
			want: `ok`,
		},
		{
			name: "multi-line script stripped",
			raw:  "a<script>\nvar x = 1;\nconsole.log(x);\n</script>b",
			want: "ab",
		},
		{
			name: "inline attributes on the open tag",
			raw:  `<script type="text/javascript">x()</script>y`,
			want: `y`,
		},
		{
			name: "nothing to strip",
			raw:  `<body><p>plain</p></body>`,
			want: `<body><p>plain</p></body>`,
		},
		{
			name: "empty input",
			raw:  ``,
			want: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeHTML(tt.raw); got != tt.want {
				t.Errorf("sanitizeHTML(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestPrepareHTMLCaps(t *testing.T) {
	long := strings.Repeat("a", 100)
	got := prepareHTML(long, 50)
	if got != strings.Repeat("a", 50)+htmlTruncatedMarker {
		t.Errorf("prepareHTML capped = %q, want 50 a's + %q", got, htmlTruncatedMarker)
	}

	if got := prepareHTML("short", 50); got != "short" {
		t.Errorf("prepareHTML under cap = %q, want untouched", got)
	}

	if got := prepareHTML(long, 0); got != long {
		t.Errorf("prepareHTML max=0 = %q, want untouched (no cap)", got)
	}

	if got := prepareHTML("", 50); got != "" {
		t.Errorf("prepareHTML empty = %q, want empty", got)
	}
}
