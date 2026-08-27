package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// flowFile is the authoring contract: one feature per file, one case per
// scenario. The loader is strict — unknown fields or actions fail fast
// before a browser ever launches.
type flowFile struct {
	Feature string     `json:"feature"`
	BaseURL string     `json:"base_url"`
	Cases   []flowCase `json:"cases"`
}

type flowCase struct {
	Name  string     `json:"name"`
	Steps []flowStep `json:"steps"`
}

type flowStep struct {
	Action   string `json:"action"`
	URL      string `json:"url,omitempty"`
	Selector string `json:"selector,omitempty"`
	Text     string `json:"text,omitempty"`
	To       string `json:"to,omitempty"`
	MS       int    `json:"ms,omitempty"`
	Name     string `json:"name,omitempty"`
	Capture  bool   `json:"capture,omitempty"`
}

var stepFieldSets = map[string]map[string]bool{
	"goto":       {"action": true, "url": true},
	"click":      {"action": true, "selector": true, "capture": true},
	"type":       {"action": true, "selector": true, "text": true, "capture": true},
	"scroll":     {"action": true, "to": true, "selector": true, "capture": true},
	"wait":       {"action": true, "ms": true, "selector": true},
	"screenshot": {"action": true, "name": true},
}

func loadFlow(path string) (*flowFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("flow: %w", err)
	}
	var f flowFile
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("flow %s: %w", path, err)
	}
	if err := f.validate(); err != nil {
		return nil, err
	}
	return &f, nil
}

// UnmarshalJSON validates per-action field sets so a typo can never load.
func (s *flowStep) UnmarshalJSON(b []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	var action string
	if err := json.Unmarshal(raw["action"], &action); err != nil {
		return fmt.Errorf("step action: %w", err)
	}
	allowed, ok := stepFieldSets[action]
	if !ok {
		return fmt.Errorf("unknown action %q", action)
	}
	for k := range raw {
		if !allowed[k] {
			return fmt.Errorf("unknown field %q on action %q", k, action)
		}
	}
	type plain flowStep
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	return dec.Decode((*plain)(s))
}

func (f *flowFile) validate() error {
	if strings.TrimSpace(f.Feature) == "" {
		return errors.New("flow: feature must be non-empty")
	}
	u, err := url.Parse(f.BaseURL)
	if err != nil || !u.IsAbs() || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("flow: base_url %q must be an absolute http(s) URL", f.BaseURL)
	}
	if len(f.Cases) == 0 {
		return errors.New("flow: cases must contain at least one case")
	}
	seen := map[string]bool{}
	for _, c := range f.Cases {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			return errors.New("flow: every case needs a non-empty name")
		}
		if seen[name] {
			return fmt.Errorf("flow: duplicate case name %q", name)
		}
		seen[name] = true
		if err := c.validate(); err != nil {
			return fmt.Errorf("flow: case %q: %w", name, err)
		}
	}
	return nil
}

func (c *flowCase) validate() error {
	if len(c.Steps) == 0 {
		return errors.New("steps must not be empty")
	}
	shotNames := map[string]bool{}
	for i, s := range c.Steps {
		switch s.Action {
		case "goto":
			if strings.TrimSpace(s.URL) == "" {
				return fmt.Errorf("step %d: goto requires url", i+1)
			}
		case "click":
			if strings.TrimSpace(s.Selector) == "" {
				return fmt.Errorf("step %d: click requires selector", i+1)
			}
		case "type":
			if strings.TrimSpace(s.Selector) == "" {
				return fmt.Errorf("step %d: type requires selector", i+1)
			}
			if s.Text == "" {
				return fmt.Errorf("step %d: type requires text", i+1)
			}
		case "scroll":
			if s.To == "" && s.Selector == "" {
				return fmt.Errorf("step %d: scroll requires to or selector", i+1)
			}
			if s.To != "" && s.Selector != "" {
				return fmt.Errorf("step %d: scroll takes to OR selector, not both", i+1)
			}
			if s.To != "" && s.To != "top" && s.To != "bottom" {
				return fmt.Errorf("step %d: scroll to must be top or bottom, got %q", i+1, s.To)
			}
		case "wait":
			if (s.MS == 0) == (s.Selector == "") {
				return fmt.Errorf("step %d: wait requires exactly one of ms or selector", i+1)
			}
			if s.MS < 0 {
				return fmt.Errorf("step %d: wait ms must be positive", i+1)
			}
		case "screenshot":
			if strings.TrimSpace(s.Name) == "" {
				return fmt.Errorf("step %d: screenshot requires a name", i+1)
			}
			if shotNames[s.Name] {
				return fmt.Errorf("step %d: duplicate screenshot name %q", i+1, s.Name)
			}
			shotNames[s.Name] = true
		}
	}
	return nil
}

// resolveURL turns a possibly-relative goto URL into an absolute one.
func (f *flowFile) resolveURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("goto url %q: %w", raw, err)
	}
	if u.IsAbs() {
		return raw, nil
	}
	base, err := url.Parse(f.BaseURL)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(u).String(), nil
}

func (f *flowFile) resolveCase(name string) (*flowCase, error) {
	for i := range f.Cases {
		if f.Cases[i].Name == name {
			return &f.Cases[i], nil
		}
	}
	return nil, fmt.Errorf("flow: unknown case %q (available: %s)", name, caseNames(f))
}

func caseNames(f *flowFile) string {
	names := make([]string, 0, len(f.Cases))
	for _, c := range f.Cases {
		names = append(names, c.Name)
	}
	return strings.Join(names, ", ")
}
