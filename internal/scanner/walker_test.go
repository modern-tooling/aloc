package scanner

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// walkRelPaths runs a quick or deep walk over root and returns the
// slash-separated paths it emits, relative to root and sorted.
func walkRelPaths(t *testing.T, root string, deep bool) []string {
	t.Helper()
	w, err := NewWalker(root, WalkOptions{DeepMode: deep})
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

func writeFiles(t *testing.T, root string, files map[string]string) {
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

func TestQuickModeScansEveryLanguageExtension(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		// ES and TypeScript module extensions: whole Node codebases use these
		"server.mjs": "export const a = 1\n",
		"legacy.cjs": "module.exports = {}\n",
		"types.mts":  "export type A = string\n",
		"shim.cts":   "export = {}\n",
		// languages.json knows these, the old hand-written allowlist did not
		"engine.cc": "int main() {}\n",
		"app.dart":  "void main() {}\n",
		"build.zig": "pub fn main() void {}\n",
		// kept from the old allowlist although languages.json has no entry
		"nginx.conf": "worker_processes 1;\n",
		// exported or machine-written: skipped in quick mode
		"icon.svg":     "<svg></svg>\n",
		"report.ipynb": "{}\n",
		"fw.hex":       ":00000001FF\n",
		// unknown extension: skipped
		"data.bin": "\x00\x01",
	})

	got := walkRelPaths(t, root, false)
	want := []string{
		"app.dart", "build.zig", "engine.cc", "legacy.cjs", "nginx.conf",
		"server.mjs", "shim.cts", "types.mts",
	}
	if len(got) != len(want) {
		t.Fatalf("quick walk:\n got  %v\n want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("quick walk:\n got  %v\n want %v", got, want)
		}
	}
}

func TestDeepModeStillScansQuickModeSkips(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"icon.svg":     "<svg></svg>\n",
		"report.ipynb": "{}\n",
	})

	got := walkRelPaths(t, root, true)
	if len(got) != 2 {
		t.Fatalf("deep walk should include svg and ipynb, got %v", got)
	}
}
