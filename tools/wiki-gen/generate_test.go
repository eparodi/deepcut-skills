package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The pinned checks for the wiki generator. They enforce the spec's
// drift-proofing invariant: any change to a skill (or to catalog.json)
// that isn't reflected in the generated wiki/ fails the suite.

var linkRe = regexp.MustCompile(`\]\(([^)]+)\)`)

// testRepoRoot resolves the repo root from the test's working directory.
func testRepoRoot(t *testing.T) string {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// TestFrontmatterParses pins that every skill in .agents/skills parses:
// frontmatter name equals the directory name, description is non-empty.
func TestFrontmatterParses(t *testing.T) {
	root := testRepoRoot(t)
	skills, err := loadSkills(filepath.Join(root, skillsDirRel))
	if err != nil {
		t.Fatalf("loadSkills: %v", err)
	}
	if len(skills) == 0 {
		t.Fatal("no skills loaded")
	}
	for _, s := range skills {
		if s.Name == "" || s.Description == "" {
			t.Errorf("skill %q has empty name or description", s.Name)
		}
	}
}

// TestCatalogCompleteness pins the bidirectional manifest contract: every
// skill dir must have a catalog entry and every entry must resolve. This is
// the tripwire — adding a skill without cataloging it fails here.
func TestCatalogCompleteness(t *testing.T) {
	root := testRepoRoot(t)
	skills, err := loadSkills(filepath.Join(root, skillsDirRel))
	if err != nil {
		t.Fatal(err)
	}
	cat, err := loadCatalog(filepath.Join(root, catalogRel))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateCatalog(cat, skills); err != nil {
		t.Fatal(err)
	}
}

// TestDeterministic pins byte-identical regeneration for unchanged input.
func TestDeterministic(t *testing.T) {
	root := testRepoRoot(t)
	skillsDir := filepath.Join(root, skillsDirRel)
	catalogPath := filepath.Join(root, catalogRel)

	dir1 := t.TempDir()
	dir2 := t.TempDir()
	if err := generate(skillsDir, catalogPath, dir1); err != nil {
		t.Fatalf("generate 1: %v", err)
	}
	if err := generate(skillsDir, catalogPath, dir2); err != nil {
		t.Fatalf("generate 2: %v", err)
	}
	diffTrees(t, dir1, dir2)
}

// TestWikiUpToDate is the stale check: the committed wiki/ must equal a
// fresh regeneration, byte for byte.
func TestWikiUpToDate(t *testing.T) {
	root := testRepoRoot(t)
	tmp := t.TempDir()
	if err := generate(filepath.Join(root, skillsDirRel), filepath.Join(root, catalogRel), tmp); err != nil {
		t.Fatalf("generate: %v", err)
	}
	diffTrees(t, tmp, filepath.Join(root, wikiDirRel))
}

// TestHomeLinksResolve pins that every relative link on the home page
// resolves to a generated page (no dead links).
func TestHomeLinksResolve(t *testing.T) {
	root := testRepoRoot(t)
	home, err := os.ReadFile(filepath.Join(root, wikiDirRel, "Home.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range linkRe.FindAllStringSubmatch(string(home), -1) {
		target := m[1]
		if strings.HasPrefix(target, "http") || strings.HasPrefix(target, "#") {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, wikiDirRel, target+".md")); err != nil {
			t.Errorf("home link %q does not resolve to a generated page", target)
		}
	}
}

// TestPagesCarryMarker pins that every generated page carries the
// "do not edit by hand" marker, so generated content is never hand-edited.
func TestPagesCarryMarker(t *testing.T) {
	root := testRepoRoot(t)
	err := filepath.WalkDir(filepath.Join(root, wikiDirRel), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(b), "do not edit by hand") {
			t.Errorf("%s: missing generated marker", filepath.Base(path))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// diffTrees reports every difference between two generated trees.
func diffTrees(t *testing.T, a, b string) {
	t.Helper()
	for _, d := range diffDirs(a, b) {
		t.Error(d)
	}
}

// TestSlugify pins the GitHub anchor algorithm (verified against the live
// blob pages 2026-08-26). Anchors drive the Sections links on skill pages.
func TestSlugify(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Phase 0 — Triage: \"Is This Non-Trivial?\"", "phase-0--triage-is-this-non-trivial"},
		{"Process", "process"},
		{"Review Gate", "review-gate"},
		{"Quick Reference", "quick-reference"},
		{"Common AI Hallucinations — Complete Reference", "common-ai-hallucinations--complete-reference"},
		{"DO — Dynamic route with params (Next.js 15+)", "do--dynamic-route-with-params-nextjs-15"},
		{"Rule: Never add `\"use client\"` unless you have to", "rule-never-add-use-client-unless-you-have-to"},
	}
	for _, c := range cases {
		if got := slugify(c.in); got != c.want {
			t.Errorf("slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestSectionsSkipCodeBlocks pins fence-aware extraction: headings inside
// fenced code blocks (template examples) never appear in the outline.
func TestSectionsSkipCodeBlocks(t *testing.T) {
	md := "---\nname: x\ndescription: y\n---\n\n# X\n\n## Real One\n\nSome text.\n\n```markdown\n## Template Heading\n## Another Template\n```\n\n## Real Two\n\n### Sub of real two\n"
	secs := sectionsOf(md)
	if len(secs) != 2 {
		t.Fatalf("got %d sections, want 2: %+v", len(secs), secs)
	}
	if secs[0].Title != "Real One" || secs[1].Title != "Real Two" {
		t.Errorf("unexpected sections: %+v", secs)
	}
	if len(secs[1].Subs) != 1 || secs[1].Subs[0] != "Sub of real two" {
		t.Errorf("unexpected subs for Real Two: %+v", secs[1].Subs)
	}
}

// TestNoTemplateLeak pins that skills embedding template examples (whose
// headings live inside fenced code blocks) produce exactly their real
// section count — regression pin for the leaked-heading bug.
func TestNoTemplateLeak(t *testing.T) {
	root := testRepoRoot(t)
	skills, err := loadSkills(filepath.Join(root, skillsDirRel))
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, s := range skills {
		counts[s.Name] = len(s.Sections)
	}
	// skill-factory, learning-porter, qa, security-engineer, and reviewer
	// embed template examples whose headings live inside fenced code blocks;
	// those must not leak into the outline (regression pin for fence-aware
	// section extraction, 2026-08-26).
	for name, want := range map[string]int{
		"skill-factory":     6,
		"learning-porter":   6,
		"qa":                6,
		"security-engineer": 7,
		"reviewer":          5,
	} {
		if counts[name] != want {
			t.Errorf("%s: %d sections, want %d (template headings leaked?)", name, counts[name], want)
		}
	}
}
