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
