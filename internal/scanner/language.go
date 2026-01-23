package scanner

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

//go:embed languages.json
var languagesJSON []byte

// LanguageConfig represents a language definition from languages.json
type LanguageConfig struct {
	Name              string     `json:"name"`
	LineComment       []string   `json:"line_comment"`
	MultiLineComments [][]string `json:"multi_line_comments"`
	Extensions        []string   `json:"extensions"`
	Filenames         []string   `json:"filenames"`
	Shebangs          []string   `json:"shebangs"`
	Env               []string   `json:"env"`
	Nested            bool       `json:"nested"`
	Blank             bool       `json:"blank"`
	Literate          bool       `json:"literate"`
	Category          string     `json:"category"`
}

// LanguagesFile represents the top-level JSON structure
type LanguagesFile struct {
	Languages map[string]LanguageConfig `json:"languages"`
}

var (
	// extToLang maps file extensions to language IDs
	extToLang map[string]string
	// filenameToLang maps special filenames to language IDs
	filenameToLang map[string]string
	// shebangToLang maps interpreter names to language IDs
	shebangToLang map[string]string
	// languages holds the full config for each language
	languages map[string]LanguageConfig
)

func init() {
	var lf LanguagesFile
	if err := json.Unmarshal(languagesJSON, &lf); err != nil {
		panic("failed to parse languages.json: " + err.Error())
	}

	extToLang = make(map[string]string)
	filenameToLang = make(map[string]string)
	shebangToLang = make(map[string]string)
	languages = make(map[string]LanguageConfig)

	for id, cfg := range lf.Languages {
		// use display name if set, otherwise use the ID
		displayName := cfg.Name
		if displayName == "" {
			displayName = id
		}
		cfg.Name = displayName
		languages[displayName] = cfg

		// map extensions to language
		for _, ext := range cfg.Extensions {
			extToLang[ext] = displayName
		}

		// map special filenames to language
		for _, fname := range cfg.Filenames {
			filenameToLang[strings.ToLower(fname)] = displayName
		}

		// map shebang interpreters to language
		for _, env := range cfg.Env {
			shebangToLang[env] = displayName
		}

		// parse full shebang patterns for direct matches
		for _, shebang := range cfg.Shebangs {
			// extract interpreter from shebang like "#!/bin/bash" or "#!/usr/bin/env python"
			interpreter := extractInterpreter(shebang)
			if interpreter != "" {
				shebangToLang[interpreter] = displayName
			}
		}
	}
}

// extractInterpreter extracts the interpreter name from a shebang line
func extractInterpreter(shebang string) string {
	shebang = strings.TrimPrefix(shebang, "#!")
	shebang = strings.TrimSpace(shebang)

	// handle /usr/bin/env style
	if strings.Contains(shebang, "env ") {
		parts := strings.Fields(shebang)
		if len(parts) >= 2 {
			return parts[len(parts)-1]
		}
	}

	return filepath.Base(shebang)
}

// DetectLanguage detects the programming language from a file path
func DetectLanguage(path string) string {
	// check special filenames first
	base := filepath.Base(path)
	if lang, ok := filenameToLang[strings.ToLower(base)]; ok {
		return lang
	}

	// check extension
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if lang, ok := extToLang[ext]; ok {
		return lang
	}

	// check shebang for files without extension
	if ext == "" || filepath.Ext(path) == "" {
		if lang := detectFromShebang(path); lang != "" {
			return lang
		}
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

			// handle /usr/bin/env
			if strings.Contains(shebang, "env ") {
				parts := strings.Fields(shebang)
				if len(parts) >= 2 {
					shebang = parts[len(parts)-1]
				}
			}

			// extract interpreter name
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

// GetLanguageConfig returns the configuration for a language
func GetLanguageConfig(lang string) (LanguageConfig, bool) {
	cfg, ok := languages[lang]
	return cfg, ok
}

// GetAllLanguages returns all language configurations
func GetAllLanguages() map[string]LanguageConfig {
	return languages
}
