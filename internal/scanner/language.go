package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

var extToLang = map[string]string{
	"go":     "Go",
	"ts":     "TypeScript",
	"tsx":    "TypeScript",
	"js":     "JavaScript",
	"jsx":    "JavaScript",
	"mjs":    "JavaScript",
	"cjs":    "JavaScript",
	"py":     "Python",
	"pyw":    "Python",
	"rs":     "Rust",
	"java":   "Java",
	"kt":     "Kotlin",
	"kts":    "Kotlin",
	"swift":  "Swift",
	"rb":     "Ruby",
	"c":      "C",
	"h":      "C",
	"cpp":    "C++",
	"cc":     "C++",
	"cxx":    "C++",
	"hpp":    "C++",
	"hxx":    "C++",
	"cs":     "C#",
	"sh":     "Shell",
	"bash":   "Shell",
	"zsh":    "Shell",
	"fish":   "Shell",
	"yaml":   "YAML",
	"yml":    "YAML",
	"json":   "JSON",
	"toml":   "TOML",
	"md":     "Markdown",
	"mdx":    "MDX",
	"sql":    "SQL",
	"tf":     "Terraform",
	"tfvars": "Terraform",
	"hcl":    "HCL",
	"proto":  "Protocol Buffers",
	"html":   "HTML",
	"htm":    "HTML",
	"css":    "CSS",
	"scss":   "SCSS",
	"sass":   "SASS",
	"less":   "LESS",
	"xml":    "XML",
	"vue":    "Vue",
	"svelte": "Svelte",
	"php":    "PHP",
	"pl":     "Perl",
	"pm":     "Perl",
	"r":      "R",
	"lua":    "Lua",
	"ex":     "Elixir",
	"exs":    "Elixir",
	"erl":    "Erlang",
	"hrl":    "Erlang",
	"hs":     "Haskell",
	"scala":  "Scala",
	"clj":    "Clojure",
	"cljs":   "Clojure",
	"dart":   "Dart",
	"groovy": "Groovy",
	"gradle": "Groovy",
	"txt":    "Plain Text",
	"text":   "Plain Text",
}

var shebangToLang = map[string]string{
	"python":  "Python",
	"python3": "Python",
	"node":    "JavaScript",
	"bash":    "Shell",
	"sh":      "Shell",
	"zsh":     "Shell",
	"ruby":    "Ruby",
	"perl":    "Perl",
	"php":     "PHP",
}

func DetectLanguage(path string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if lang, ok := extToLang[ext]; ok {
		return lang
	}

	// Check shebang for files without extension
	if ext == "" || filepath.Ext(path) == "" {
		if lang := detectFromShebang(path); lang != "" {
			return lang
		}
	}

	// Special filenames
	base := filepath.Base(path)
	switch strings.ToLower(base) {
	case "dockerfile":
		return "Dockerfile"
	case "makefile", "gnumakefile":
		return "Makefile"
	case "cakefile":
		return "CoffeeScript"
	case "gemfile", "rakefile":
		return "Ruby"
	case "justfile":
		return "Just"
	case "taskfile.yml", "taskfile.yaml":
		return "YAML"
	}

	return "unknown"
}

func detectFromShebang(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#!") {
			shebang := strings.TrimPrefix(line, "#!")
			shebang = strings.TrimSpace(shebang)

			// Handle /usr/bin/env
			if strings.Contains(shebang, "env ") {
				parts := strings.Fields(shebang)
				if len(parts) >= 2 {
					shebang = parts[len(parts)-1]
				}
			}

			// Extract interpreter name
			shebang = filepath.Base(shebang)
			if lang, ok := shebangToLang[shebang]; ok {
				return lang
			}
		}
	}
	return ""
}

func extToLanguage(ext string) string {
	if lang, ok := extToLang[ext]; ok {
		return lang
	}
	return "unknown"
}
