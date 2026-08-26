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
