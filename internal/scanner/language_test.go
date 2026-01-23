package scanner

import "testing"

func TestDetectLanguage_Extensions(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"main.go", "Go"},
		{"app.ts", "TypeScript"},
		{"app.tsx", "TSX"},
		{"script.js", "JavaScript"},
		{"script.jsx", "JSX"},
		{"main.py", "Python"},
		{"lib.rs", "Rust"},
		{"App.java", "Java"},
		{"app.kt", "Kotlin"},
		{"app.swift", "Swift"},
		{"script.rb", "Ruby"},
		{"main.c", "C"},
		{"main.cpp", "C++"},
		{"main.cc", "C++"},
		{"App.cs", "C#"},
		{"script.sh", "Shell"},
		{"config.yaml", "YAML"},
		{"config.yml", "YAML"},
		{"data.json", "JSON"},
		{"config.toml", "TOML"},
		{"README.md", "Markdown"},
		{"query.sql", "SQL"},
		{"main.tf", "HCL"},
		{"service.proto", "Protocol Buffers"},
		{"index.html", "HTML"},
		{"style.css", "CSS"},
		{"style.scss", "Sass"},
		{"unknown.xyz", "unknown"},
	}

	for _, tt := range tests {
		got := DetectLanguage(tt.path)
		if got != tt.want {
			t.Errorf("DetectLanguage(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestDetectLanguage_SpecialFiles(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"Dockerfile", "Dockerfile"},
		{"Makefile", "Makefile"},
		{"GNUmakefile", "Makefile"},
		{"Rakefile", "Rakefile"},
		{"justfile", "Just"},
	}

	for _, tt := range tests {
		got := DetectLanguage(tt.path)
		if got != tt.want {
			t.Errorf("DetectLanguage(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestExtToLanguage(t *testing.T) {
	if extToLanguage("go") != "Go" {
		t.Error("extToLanguage(go) should return Go")
	}
	if extToLanguage("xyz") != "unknown" {
		t.Error("extToLanguage(xyz) should return unknown")
	}
}
