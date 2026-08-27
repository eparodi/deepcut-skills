package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// envFilePath is the gitignored key file at the repo root. Resolution order
// everywhere: explicit flag > process env > this file > built-in default.
const envFilePath = ".env.visualqa"

const defaultVisionModel = "deepseek-v4-flash-vision-exp"

// loadEnvFile parses simple KEY=VALUE lines. Blank lines and # comments are
// skipped; CRLF is trimmed; later duplicate keys win; a line without '='
// is an error so a mistyped file can never silently drop the key.
func loadEnvFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("env file %s: %w", path, err)
	}
	out := map[string]string{}
	for i, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("env file %s line %d: expected KEY=VALUE, got %q", path, i+1, line)
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out, nil
}

// loadChecklist reads a checklist from a file (verbatim) or a directory (all
// *.md files concatenated in sorted filename order — the "run all groups for
// this device" mode). Non-.md files and subdirectories are skipped.
func loadChecklist(path string) (string, error) {
	st, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("checklist %s: %w", path, err)
	}
	if !st.IsDir() {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("checklist %s: %w", path, err)
		}
		return string(b), nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("checklist dir %s: %w", path, err)
	}
	var parts []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(path, e.Name()))
		if err != nil {
			return "", fmt.Errorf("checklist %s: %w", filepath.Join(path, e.Name()), err)
		}
		parts = append(parts, strings.TrimSpace(string(b)))
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("checklist dir %s: no .md files", path)
	}
	return strings.Join(parts, "\n"), nil
}

// resolveModel applies flag > env > file > default.
func resolveModel(flag, env, file string) string {
	for _, v := range []string{flag, env, file} {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return defaultVisionModel
}

// resolveKey applies env > file. Empty when neither is set.
func resolveKey(env, file string) string {
	for _, v := range []string{env, file} {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// defaultChecklist is the mobile-first QA checklist (overridable via
// --checklist). It doubles as the prompt's stable prefix.
func defaultChecklist() []string {
	return []string{
		"Content fits the viewport width — no horizontal overflow or clipping.",
		"Tap targets are large enough and don't overlap.",
		"Text is readable — no truncation, overlap, or poor contrast.",
		"Images/media load with visible fallbacks — no broken-image icons.",
		"Primary navigation and actions are reachable.",
		"Layout is intentional — no blank regions or misalignment.",
		"Fixed elements (headers/bottom bars) don't cover content.",
		"Nothing in the frame suggests rendering breakage.",
	}
}
