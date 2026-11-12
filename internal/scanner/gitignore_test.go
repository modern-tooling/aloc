package scanner

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func walkedPaths(t *testing.T, root string) []string {
	t.Helper()
	w, err := NewWalker(root, WalkOptions{})
	if err != nil {
		t.Fatalf("NewWalker: %v", err)
	}
	paths, errs := w.Walk(context.Background())
	var got []string
	for p := range paths {
		rel, err := filepath.Rel(w.root, p)
		if err != nil {
			t.Fatalf("Rel(%q): %v", p, err)
		}
		got = append(got, filepath.ToSlash(rel))
	}
	for err := range errs {
		t.Fatalf("walk error: %v", err)
	}
	sort.Strings(got)
	return got
}

func TestWalkerHonorsNestedGitignore(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		".gitignore": "*.txt\n",
		// nested rules apply only beneath sub/, as in git
		"sub/.gitignore": "skills/*-team/\n*.gen.go\n!keep.txt\n",
		"sub/.ignore":    "scratch/\n",

		"main.go":                       "package main\n",
		"notes.txt":                     "ignored by root\n",
		"sub/x.go":                      "package sub\n",
		"sub/x.gen.go":                  "package sub\n",
		"sub/keep.txt":                  "re-included by nested negation\n",
		"sub/drop.txt":                  "ignored by root\n",
		"sub/skills/catch-up-team/a.md": "ignored by nested anchored dir rule\n",
		"sub/skills/core/a.md":          "kept\n",
		"sub/scratch/s.go":              "package scratch\n",
		// sub/.gitignore must not reach a sibling directory
		"other/y.gen.go": "package other\n",
	})

	got := walkedPaths(t, root)
	want := []string{
		"main.go",
		"other/y.gen.go",
		"sub/keep.txt",
		"sub/skills/core/a.md",
		"sub/x.go",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("walk:\n got  %v\n want %v", got, want)
	}
}

func TestNestedPatternDoesNotApplyToItsOwnDirectory(t *testing.T) {
	root := t.TempDir()
	// "sub" inside sub/.gitignore means sub/sub, not sub itself
	writeTree(t, root, map[string]string{
		"sub/.gitignore": "sub\n",
		"sub/a.go":       "package sub\n",
		"sub/sub/b.go":   "package sub\n",
	})

	got := walkedPaths(t, root)
	want := []string{"sub/a.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("walk:\n got  %v\n want %v", got, want)
	}
}
