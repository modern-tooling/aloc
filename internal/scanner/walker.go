package scanner

import (
	"context"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
)

// knownSourceExtensions contains extensions that are recognized as source code
// in quick mode. Files without these extensions are skipped to avoid processing
// binaries and generated files.
var knownSourceExtensions = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true, ".jsx": true,
	".java": true, ".kt": true, ".scala": true, ".rs": true, ".c": true, ".cpp": true,
	".h": true, ".hpp": true, ".cs": true, ".rb": true, ".php": true, ".swift": true,
	".m": true, ".mm": true, ".sql": true, ".sh": true, ".bash": true, ".zsh": true,
	".yaml": true, ".yml": true, ".json": true, ".xml": true, ".html": true, ".css": true,
	".scss": true, ".sass": true, ".less": true, ".vue": true, ".svelte": true,
	".md": true, ".mdx": true, ".rst": true, ".txt": true,
	".tf": true, ".hcl": true, ".proto": true, ".graphql": true,
	".lua": true, ".r": true, ".R": true, ".pl": true, ".pm": true,
	".ex": true, ".exs": true, ".erl": true, ".hs": true, ".clj": true,
	".toml": true, ".ini": true, ".cfg": true, ".conf": true,
}

func isKnownSourceExtension(ext string) bool {
	return knownSourceExtensions[ext]
}

type Walker struct {
	root       string
	numWorkers int
	exclude    []string
	deepMode   bool
	gitignore  *GitIgnore
}

type WalkOptions struct {
	NumWorkers int
	Exclude    []string
	DeepMode   bool
}

func NewWalker(root string, opts WalkOptions) (*Walker, error) {
	if opts.NumWorkers <= 0 {
		// adaptive worker count: min(32, 4*GOMAXPROCS) per impl-ideas.txt
		opts.NumWorkers = min(32, 4*runtime.GOMAXPROCS(0))
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	// Resolve symlinks so WalkDir can traverse the actual directory
	absRoot, err = filepath.EvalSymlinks(absRoot)
	if err != nil {
		return nil, err
	}

	gitignore, _ := LoadGitIgnore(absRoot) // ignore errors, gitignore is optional

	return &Walker{
		root:       absRoot,
		numWorkers: opts.NumWorkers,
		exclude:    opts.Exclude,
		deepMode:   opts.DeepMode,
		gitignore:  gitignore,
	}, nil
}

func (w *Walker) Walk(ctx context.Context) (<-chan string, <-chan error) {
	// large buffer prevents walker from stalling on slow consumers
	paths := make(chan string, 8192)
	errs := make(chan error, 256)

	go func() {
		defer close(paths)
		defer close(errs)

		err := filepath.WalkDir(w.root, func(path string, d fs.DirEntry, err error) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if err != nil {
				errs <- err
				return nil
			}

			// Check gitignore
			if w.gitignore != nil && w.gitignore.Match(path, d.IsDir()) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// skip excluded patterns
			relPath, _ := filepath.Rel(w.root, path)
			for _, pattern := range w.exclude {
				if matched, _ := filepath.Match(pattern, relPath); matched {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
				if strings.Contains(relPath, strings.TrimSuffix(strings.TrimPrefix(pattern, "**/"), "/**")) {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}

			// skip directories
			if d.IsDir() {
				name := d.Name()
				// skip common cache, build, and dependency directories
				switch name {
				case ".git", "vendor", "node_modules",
					// package manager caches
					".pnpm-store", ".yarn", ".npm",
					// build/cache directories
					".terraform", ".terragrunt-cache",
					".nx", ".turbo", ".next", ".nuxt", ".cache",
					".venv", "venv", "__pycache__", ".pytest_cache",
					".gradle", ".m2",
					// IDE directories
					".idea", ".vscode",
					// OS directories
					".DS_Store",
					// git hooks
					".husky",
					// other caches
					"dist", "build", "target", "out",
					".angular", ".svelte-kit",
					// generated/temp directories
					"generated", "tmp":
					return filepath.SkipDir
				}
				return nil
			}

			// in quick mode, only process files with known source extensions
			// (skip extensionless files which are usually binaries or generated)
			if !w.deepMode {
				ext := strings.ToLower(filepath.Ext(path))
				if ext == "" || !isKnownSourceExtension(ext) {
					return nil
				}
			}

			// binary check moved to CountLOC for single file open
			paths <- path
			return nil
		})

		if err != nil && err != context.Canceled {
			errs <- err
		}
	}()

	return paths, errs
}
