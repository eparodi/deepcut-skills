package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	check := flag.Bool("check", false, "regenerate into a temp dir and fail (exit 1) if wiki/ is stale")
	flag.Parse()

	root, err := repoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "wiki-gen:", err)
		os.Exit(1)
	}
	skillsDir := filepath.Join(root, skillsDirRel)
	catalogPath := filepath.Join(root, catalogRel)
	wikiDir := filepath.Join(root, wikiDirRel)

	if *check {
		tmp, err := os.MkdirTemp("", "wiki-gen-check-")
		if err != nil {
			fmt.Fprintln(os.Stderr, "wiki-gen:", err)
			os.Exit(1)
		}
		defer os.RemoveAll(tmp)
		if err := generate(skillsDir, catalogPath, tmp); err != nil {
			fmt.Fprintln(os.Stderr, "wiki-gen:", err)
			os.Exit(1)
		}
		stale := diffDirs(tmp, wikiDir)
		if len(stale) > 0 {
			for _, s := range stale {
				fmt.Fprintln(os.Stderr, "stale:", s)
			}
			os.Exit(1)
		}
		fmt.Println("wiki is up to date")
		return
	}

	if err := generate(skillsDir, catalogPath, wikiDir); err != nil {
		fmt.Fprintln(os.Stderr, "wiki-gen:", err)
		os.Exit(1)
	}
	fmt.Println("wiki regenerated")
}

// readAll walks dir and returns every generated .md file's content keyed by
// its slash-separated relative path. Non-markdown files (e.g. the
// catalog.json input) are not generated output and are ignored.
func readAll(dir string) map[string][]byte {
	files := map[string][]byte{}
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		files[filepath.ToSlash(rel)] = b
		return nil
	})
	return files
}

// diffDirs returns a description of every difference between generated dir a
// and reference dir b: files missing from b, files present only in b, and
// files whose content differs.
func diffDirs(a, b string) []string {
	filesA := readAll(a)
	filesB := readAll(b)
	var diffs []string
	for name, content := range filesA {
		other, ok := filesB[name]
		if !ok {
			diffs = append(diffs, fmt.Sprintf("%s: missing in %s", name, b))
			continue
		}
		if string(content) != string(other) {
			diffs = append(diffs, fmt.Sprintf("%s: differs from %s", name, b))
		}
	}
	for name := range filesB {
		if _, ok := filesA[name]; !ok {
			diffs = append(diffs, fmt.Sprintf("%s: extra in %s (not generated)", name, b))
		}
	}
	return diffs
}
