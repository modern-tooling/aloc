package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// GitIgnore holds parsed gitignore patterns
type GitIgnore struct {
	patterns []gitignorePattern
	root     string
}

type gitignorePattern struct {
	pattern  string
	negated  bool
	dirOnly  bool
	anchored bool
}

// LoadGitIgnore loads and parses ignore files (.gitignore and .ignore)
// Patterns from .ignore take precedence over .gitignore (loaded second)
func LoadGitIgnore(root string) (*GitIgnore, error) {
	gi := &GitIgnore{root: root}

	// load in order: .gitignore, then .ignore (later files take precedence)
	for _, filename := range []string{".gitignore", ".ignore"} {
		patterns, err := loadIgnoreFile(filepath.Join(root, filename))
		if err != nil {
			return nil, err
		}
		gi.patterns = append(gi.patterns, patterns...)
	}

	return gi, nil
}

// loadIgnoreFile parses a single ignore file and returns its patterns
func loadIgnoreFile(path string) ([]gitignorePattern, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var patterns []gitignorePattern
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, parsePattern(line))
	}
	return patterns, scanner.Err()
}

func parsePattern(line string) gitignorePattern {
	p := gitignorePattern{pattern: line}

	// Check for negation
	if strings.HasPrefix(line, "!") {
		p.negated = true
		p.pattern = line[1:]
	}

	// Check for directory-only
	if strings.HasSuffix(p.pattern, "/") {
		p.dirOnly = true
		p.pattern = strings.TrimSuffix(p.pattern, "/")
	}

	// Check for anchored patterns
	if strings.HasPrefix(p.pattern, "/") {
		p.anchored = true
		p.pattern = p.pattern[1:]
	} else if strings.Contains(p.pattern, "/") {
		p.anchored = true
	}

	return p
}

// Match checks if a path should be ignored
func (gi *GitIgnore) Match(path string, isDir bool) bool {
	if len(gi.patterns) == 0 {
		return false
	}

	relPath, err := filepath.Rel(gi.root, path)
	if err != nil {
		return false
	}
	relPath = filepath.ToSlash(relPath)

	ignored := false
	for _, p := range gi.patterns {
		if p.dirOnly && !isDir {
			continue
		}

		matched := matchPattern(p.pattern, relPath, p.anchored)
		if matched {
			ignored = !p.negated
		}
	}
	return ignored
}

func matchPattern(pattern, path string, anchored bool) bool {
	// Simple glob matching
	if anchored {
		matched, _ := filepath.Match(pattern, path)
		if matched {
			return true
		}
		// Also try matching just the first part for directory patterns
		parts := strings.Split(path, "/")
		if len(parts) > 0 {
			matched, _ = filepath.Match(pattern, parts[0])
			if matched {
				return true
			}
		}
	} else {
		// Match against any path component
		parts := strings.Split(path, "/")
		for _, part := range parts {
			matched, _ := filepath.Match(pattern, part)
			if matched {
				return true
			}
		}
	}
	return false
}
