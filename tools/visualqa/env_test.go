package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeEnvFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env.visualqa")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadEnvFile(t *testing.T) {
	path := writeEnvFile(t, "# comment line\n\nDEEPSEEK_API_KEY=sk-test-abc\nDEEPSEEK_VISION_MODEL=deepseek-v4-pro\r\n  DEEPSEEK_API_KEY=overwritten  \n")
	got, err := loadEnvFile(path)
	if err != nil {
		t.Fatalf("loadEnvFile: %v", err)
	}
	if got["DEEPSEEK_API_KEY"] != "overwritten" {
		t.Errorf("DEEPSEEK_API_KEY = %q, want overwritten (later line wins)", got["DEEPSEEK_API_KEY"])
	}
	if got["DEEPSEEK_VISION_MODEL"] != "deepseek-v4-pro" {
		t.Errorf("DEEPSEEK_VISION_MODEL = %q (CRLF trimmed)", got["DEEPSEEK_VISION_MODEL"])
	}
}

func TestLoadEnvFileMissingIsEmpty(t *testing.T) {
	got, err := loadEnvFile(filepath.Join(t.TempDir(), "nope.env"))
	if err != nil {
		t.Fatalf("loadEnvFile(missing): %v, want nil error + empty map", err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want empty map", got)
	}
}

func TestLoadEnvFileMalformedLineErrors(t *testing.T) {
	path := writeEnvFile(t, "DEEPSEEK_API_KEY=sk-test-abc\nthis line has no equals\n")
	if _, err := loadEnvFile(path); err == nil {
		t.Fatal("loadEnvFile accepted a malformed line, want error")
	}
}

func TestResolveCookiePrecedence(t *testing.T) {
	tests := []struct {
		name            string
		flag, env, file string
		want            string
	}{
		{"flag wins", "session=flag", "session=env", "session=file", "session=flag"},
		{"env beats file", "", "session=env", "session=file", "session=env"},
		{"file beats none", "", "", "session=file", "session=file"},
		{"all empty", "", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveCookie(tt.flag, tt.env, tt.file); got != tt.want {
				t.Errorf("resolveCookie = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseCookie(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantN   string
		wantV   string
		wantErr bool
	}{
		{"valid", "session=eyJhbGciOiJIUzI1NiJ9.abc", "session", "eyJhbGciOiJIUzI1NiJ9.abc", false},
		{"value with equals", "session=a=b=c", "session", "a=b=c", false},
		{"empty value allowed", "session=", "session", "", false},
		{"empty name", "=value", "", "", true},
		{"no equals", "novalue", "", "", true},
		{"empty string", "", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, v, err := parseCookie(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseCookie(%q) succeeded, want error", tt.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseCookie(%q): %v", tt.in, err)
			}
			if n != tt.wantN || v != tt.wantV {
				t.Errorf("parseCookie(%q) = %q=%q, want %q=%q", tt.in, n, v, tt.wantN, tt.wantV)
			}
		})
	}
}

func TestResolveModelPrecedence(t *testing.T) {
	const def = "deepseek-v4-flash-vision-exp"
	tests := []struct {
		name            string
		flag, env, file string
		want            string
	}{
		{"all empty -> default", "", "", "", def},
		{"flag wins over env and file", "--model", "env", "file", "--model"},
		{"env beats file", "", "env", "file", "env"},
		{"file beats default", "", "", "file", "file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveModel(tt.flag, tt.env, tt.file); got != tt.want {
				t.Errorf("resolveModel(%q,%q,%q) = %q, want %q", tt.flag, tt.env, tt.file, got, tt.want)
			}
		})
	}
}

func TestResolveKeyPrecedence(t *testing.T) {
	if got := resolveKey("env", "file"); got != "env" {
		t.Errorf("resolveKey(env,file) = %q, want env", got)
	}
	if got := resolveKey("", "file"); got != "file" {
		t.Errorf("resolveKey(,file) = %q, want file", got)
	}
	if got := resolveKey("", ""); got != "" {
		t.Errorf("resolveKey(,) = %q, want empty", got)
	}
}

func TestDefaultChecklist(t *testing.T) {
	cl := defaultChecklist()
	if len(cl) != 8 {
		t.Fatalf("len(defaultChecklist) = %d, want 8", len(cl))
	}
	if !strings.Contains(cl[0], "viewport") {
		t.Errorf("item 1 = %q, want it to mention the viewport", cl[0])
	}
}

func TestLoadChecklistFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "layout.md")
	if err := os.WriteFile(path, []byte("one\ntwo\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := loadChecklist(path)
	if err != nil {
		t.Fatalf("loadChecklist(file): %v", err)
	}
	if got != "one\ntwo\n" {
		t.Errorf("got %q, want verbatim file content", got)
	}
}

func TestLoadChecklistDirectory(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"b.md":        "b1\nb2\n",
		"a.md":        "a1\na2\n",
		"notes.txt":   "ignored\n", // non-.md must be skipped
		"sub/keep.md": "nested\n",
	}
	for name, content := range files {
		if strings.Contains(name, "/") {
			if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	got, err := loadChecklist(dir)
	if err != nil {
		t.Fatalf("loadChecklist(dir): %v", err)
	}
	// .md files only, sorted by name, joined with \n; non-md and nested skipped.
	if got != "a1\na2\nb1\nb2" {
		t.Errorf("got %q, want sorted .md-only concatenation", got)
	}
}

func TestLoadChecklistErrors(t *testing.T) {
	if _, err := loadChecklist(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Error("loadChecklist(missing) succeeded, want error")
	}
	emptyDir := t.TempDir()
	if _, err := loadChecklist(emptyDir); err == nil {
		t.Error("loadChecklist(dir without .md) succeeded, want error")
	}
}
