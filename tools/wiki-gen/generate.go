// Package main implements the deepcut-skills wiki generator: it reads the
// hub's .agents/skills/*/SKILL.md files plus wiki/catalog.json metadata and
// writes the wiki pages under wiki/ (Home.md, _Sidebar.md, <name>.md).
//
// The wiki is generated so the SKILL.md files stay the single source of
// truth — a stale wiki fails the pinned checks (generate_test.go), which is
// how "every skill update updates the wiki" is enforced rather than
// promised.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	skillsDirRel = ".agents/skills"
	catalogRel   = "wiki/catalog.json"
	wikiDirRel   = "wiki"

	repoBlobBase = "https://github.com/eparodi/deepcut-skills/blob/main"
)

const (
	categoryRole    = "role"
	categoryStack   = "stack"
	categoryProcess = "process"
)

// generatedMarker is embedded in every generated page; TestPagesCarryMarker
// pins it.
const generatedMarker = "do not edit by hand"

var categoryDisplay = map[string]string{
	categoryRole:    "Role agents",
	categoryStack:   "Stack skills",
	categoryProcess: "Process skills",
}

type skill struct {
	Name        string
	Description string
	SourcePath  string // repo-relative, e.g. .agents/skills/pm/SKILL.md
	Headings    []string
}

type catalogEntry struct {
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Notes    string   `json:"notes"`
}

type catalog struct {
	Skills map[string]catalogEntry `json:"skills"`
}

// repoRoot returns the repository root by walking up from the working
// directory until a go.mod is found. Works for both the CLI (run from the
// repo root) and tests (run from the package directory).
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found above %s", dir)
		}
		dir = parent
	}
}

// parseFrontmatter extracts name and description from a SKILL.md file's
// YAML frontmatter (--- fences). Values are the text after the first ':',
// trimmed, with one pair of surrounding double quotes stripped. Block
// scalars are rejected loudly instead of being silently mis-parsed.
func parseFrontmatter(content string) (name, description string, err error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return "", "", fmt.Errorf("missing opening frontmatter fence")
	}
	closed := false
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			closed = true
			break
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(strings.TrimSuffix(val, "\r"))
		if (val == "" || strings.HasPrefix(val, ">") || strings.HasPrefix(val, "|")) && (key == "name" || key == "description") {
			return "", "", fmt.Errorf("frontmatter %q: unsupported value %q", key, val)
		}
		if len(val) >= 2 && strings.HasPrefix(val, `"`) && strings.HasSuffix(val, `"`) {
			val = val[1 : len(val)-1]
		}
		switch key {
		case "name":
			name = val
		case "description":
			description = val
		}
	}
	if !closed {
		return "", "", fmt.Errorf("missing closing frontmatter fence")
	}
	if name == "" {
		return "", "", fmt.Errorf("frontmatter: empty name")
	}
	if description == "" {
		return "", "", fmt.Errorf("frontmatter: empty description")
	}
	return name, description, nil
}

// headingsOf returns the body's level-2 section headings (## ...).
func headingsOf(content string) []string {
	var hs []string
	fences := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "---" {
			fences++
			continue
		}
		if fences < 2 {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			hs = append(hs, strings.TrimSpace(strings.TrimPrefix(line, "## ")))
		}
	}
	return hs
}

// loadSkills scans skillsDir for <name>/SKILL.md and parses each one. A
// directory without a SKILL.md is an error: .agents/skills must contain
// only skills.
func loadSkills(skillsDir string) ([]skill, error) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, err
	}
	var skills []skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirName := e.Name()
		p := filepath.Join(skillsDir, dirName, "SKILL.md")
		content, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", p, err)
		}
		name, desc, err := parseFrontmatter(string(content))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		if name != dirName {
			return nil, fmt.Errorf("%s: frontmatter name %q != directory name %q", p, name, dirName)
		}
		skills = append(skills, skill{
			Name:        name,
			Description: desc,
			SourcePath:  filepath.ToSlash(filepath.Join(skillsDirRel, dirName, "SKILL.md")),
			Headings:    headingsOf(string(content)),
		})
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })
	return skills, nil
}

// loadCatalog parses the catalog manifest and validates its categories.
func loadCatalog(path string) (catalog, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return catalog{}, err
	}
	var c catalog
	if err := json.Unmarshal(b, &c); err != nil {
		return catalog{}, fmt.Errorf("parse %s: %w", path, err)
	}
	for name, e := range c.Skills {
		if _, ok := categoryDisplay[e.Category]; !ok {
			return catalog{}, fmt.Errorf("catalog: skill %q has invalid category %q", name, e.Category)
		}
	}
	return c, nil
}

// validateCatalog enforces the bidirectional manifest contract: every skill
// dir must have an entry and every entry must resolve to a real skill.
func validateCatalog(cat catalog, skills []skill) error {
	have := make(map[string]bool, len(skills))
	for _, s := range skills {
		have[s.Name] = true
	}
	for name := range cat.Skills {
		if !have[name] {
			return fmt.Errorf("catalog: entry %q has no matching skill dir", name)
		}
	}
	for _, s := range skills {
		if _, ok := cat.Skills[s.Name]; !ok {
			return fmt.Errorf("catalog: skill %q has no catalog entry (add it to %s)", s.Name, catalogRel)
		}
	}
	return nil
}

func escCell(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

func markerBlock(src string) string {
	return fmt.Sprintf("> **Generated** from %s — %s.", src, generatedMarker)
}

// renderHome renders the wiki index: catalog tables grouped by category.
func renderHome(skills []skill, cat catalog) string {
	var b strings.Builder
	b.WriteString("# Skill Catalog\n\n")
	b.WriteString(markerBlock("`.agents/skills/*/SKILL.md` + `wiki/catalog.json`"))
	b.WriteString("\n> Regenerate with `go run ./tools/wiki-gen`, verify with `go test ./...`.\n\n")
	b.WriteString("The agent skills of the deepcut-skills hub, cataloged from their SKILL.md files.\n\n")

	byCat := map[string][]skill{}
	for _, s := range skills {
		byCat[cat.Skills[s.Name].Category] = append(byCat[cat.Skills[s.Name].Category], s)
	}
	// Stable category order: role, stack, process.
	for _, category := range []string{categoryRole, categoryStack, categoryProcess} {
		group := byCat[category]
		if len(group) == 0 {
			continue
		}
		fmt.Fprintf(&b, "## %s (%d)\n\n", categoryDisplay[category], len(group))
		b.WriteString("| Skill | Description | Page |\n")
		b.WriteString("|---|---|---|\n")
		for _, s := range group {
			fmt.Fprintf(&b, "| `%s` | %s | [%s](%s) |\n", s.Name, escCell(s.Description), s.Name, s.Name)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Regenerating & publishing\n\n")
	b.WriteString("- Regenerate: `go run ./tools/wiki-gen` (then `go test ./...`)\n")
	b.WriteString("- Publish: `./tools/wiki-gen/publish.sh`\n")
	return b.String()
}

// renderSkillPage renders one skill's page.
func renderSkillPage(s skill, e catalogEntry) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", s.Name)
	fmt.Fprintf(&b, "> **Generated** from `%s` — %s.\n", s.SourcePath, generatedMarker)
	fmt.Fprintf(&b, "> Source: [SKILL.md](%s/%s)\n\n", repoBlobBase, s.SourcePath)

	b.WriteString(s.Description + "\n\n")

	meta := "**Category:** " + e.Category
	if len(e.Tags) > 0 {
		meta += " · **Tags:** " + strings.Join(e.Tags, ", ")
	}
	if e.Notes != "" {
		meta += " · **Notes:** " + e.Notes
	}
	b.WriteString(meta + "\n\n")

	if len(s.Headings) > 0 {
		b.WriteString("## Sections\n\n")
		for _, h := range s.Headings {
			b.WriteString("- " + h + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// renderSidebar renders the GitHub-wiki sidebar (category-grouped links).
func renderSidebar(skills []skill, cat catalog) string {
	var b strings.Builder
	b.WriteString("> Generated — " + generatedMarker + "\n\n")
	b.WriteString("- [Home](Home)\n\n")

	byCat := map[string][]skill{}
	for _, s := range skills {
		byCat[cat.Skills[s.Name].Category] = append(byCat[cat.Skills[s.Name].Category], s)
	}
	for _, category := range []string{categoryRole, categoryStack, categoryProcess} {
		group := byCat[category]
		if len(group) == 0 {
			continue
		}
		fmt.Fprintf(&b, "**%s**\n", categoryDisplay[category])
		for _, s := range group {
			fmt.Fprintf(&b, "- [%s](%s)\n", s.Name, s.Name)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// generate writes the full wiki into outDir. Output is deterministic:
// derived only from the skill files, the catalog, and the templates, with
// files written in sorted order.
func generate(skillsDir, catalogPath, outDir string) error {
	skills, err := loadSkills(skillsDir)
	if err != nil {
		return err
	}
	cat, err := loadCatalog(catalogPath)
	if err != nil {
		return err
	}
	if err := validateCatalog(cat, skills); err != nil {
		return err
	}

	files := map[string]string{
		"Home.md":     renderHome(skills, cat),
		"_Sidebar.md": renderSidebar(skills, cat),
	}
	for _, s := range skills {
		files[s.Name+".md"] = renderSkillPage(s, cat.Skills[s.Name])
	}

	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(outDir, n), []byte(files[n]), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", n, err)
		}
	}
	return nil
}
