package scanner

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/modern-tooling/aloc/internal/model"
)

// bufferPool reuses 256KB buffers - the documented sweet spot for SSD I/O
var bufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 256*1024)
		return &buf
	},
}

// CountLOC counts lines of code, returning 0 if file is binary.
// This combines binary detection with LOC counting to avoid opening the file twice.
func CountLOC(path string) (int, error) {
	metrics, err := CountLines(path)
	if err != nil {
		return 0, err
	}
	return metrics.Code, nil
}

// CountLines counts all line types (total, blanks, comments, code).
// Returns zero metrics if file is binary.
func CountLines(path string) (model.LineMetrics, error) {
	f, err := os.Open(path)
	if err != nil {
		return model.LineMetrics{}, err
	}
	defer f.Close()

	// Get pooled buffer for binary check
	bufPtr := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufPtr)
	buf := *bufPtr

	// Read first chunk and check for binary
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return model.LineMetrics{}, err
	}

	// Binary check: look for NUL byte in first 512 bytes
	checkLen := min(n, 512)
	for i := 0; i < checkLen; i++ {
		if buf[i] == 0 {
			return model.LineMetrics{}, nil // binary file, no metrics
		}
	}

	// Seek back to start for line counting
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return model.LineMetrics{}, err
	}

	lang := detectLangFromPath(path)
	return countLinesFromReader(f, lang, bufPtr), nil
}

func countLinesFromReader(f *os.File, lang string, bufPtr *[]byte) model.LineMetrics {
	scanner := bufio.NewScanner(f)
	scanner.Buffer(*bufPtr, 256*1024)

	var metrics model.LineMetrics
	inBlockComment := false
	blockStart, blockEnd := getBlockCommentMarkers(lang)
	lineComment := getLineCommentMarker(lang)

	for scanner.Scan() {
		metrics.Total++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Empty line
		if trimmed == "" {
			metrics.Blanks++
			continue
		}

		// Track if this line contributes to code
		isCode := false
		isComment := false

		// Handle block comments
		if blockStart != "" {
			if inBlockComment {
				// Inside block comment
				if idx := strings.Index(trimmed, blockEnd); idx >= 0 {
					inBlockComment = false
					remainder := strings.TrimSpace(trimmed[idx+len(blockEnd):])
					if remainder != "" && !strings.HasPrefix(remainder, lineComment) {
						isCode = true
					} else if remainder != "" {
						isComment = true
					} else {
						isComment = true
					}
				} else {
					isComment = true
				}
			} else if strings.Contains(trimmed, blockStart) {
				// Check for block comment start
				startIdx := strings.Index(trimmed, blockStart)
				beforeComment := strings.TrimSpace(trimmed[:startIdx])
				endIdx := strings.Index(trimmed[startIdx+len(blockStart):], blockEnd)

				if endIdx >= 0 {
					// Block comment starts and ends on same line
					afterComment := strings.TrimSpace(trimmed[startIdx+len(blockStart)+endIdx+len(blockEnd):])
					if beforeComment != "" || afterComment != "" {
						isCode = true
					} else {
						isComment = true
					}
				} else {
					// Block comment starts but doesn't end
					inBlockComment = true
					if beforeComment != "" {
						isCode = true
					} else {
						isComment = true
					}
				}
			} else if lineComment != "" && strings.HasPrefix(trimmed, lineComment) {
				isComment = true
			} else {
				isCode = true
			}
		} else {
			// No block comment support for this language
			if lineComment != "" && strings.HasPrefix(trimmed, lineComment) {
				isComment = true
			} else {
				isCode = true
			}
		}

		if isCode {
			metrics.Code++
		} else if isComment {
			metrics.Comments++
		}
	}

	return metrics
}

// countLOCFromReader is kept for backward compatibility
func countLOCFromReader(f *os.File, lang string, bufPtr *[]byte) int {
	return countLinesFromReader(f, lang, bufPtr).Code
}

func getLineCommentMarker(lang string) string {
	if cfg, ok := GetLanguageConfig(lang); ok {
		if cfg.Blank {
			return "" // language doesn't support comments
		}
		if len(cfg.LineComment) > 0 {
			return cfg.LineComment[0]
		}
	}
	// fallback for unknown languages
	return "//"
}

func getBlockCommentMarkers(lang string) (string, string) {
	if cfg, ok := GetLanguageConfig(lang); ok {
		if cfg.Blank || len(cfg.MultiLineComments) == 0 {
			return "", "" // no block comments
		}
		if len(cfg.MultiLineComments[0]) == 2 {
			return cfg.MultiLineComments[0][0], cfg.MultiLineComments[0][1]
		}
	}
	// fallback for unknown languages
	return "/*", "*/"
}

func detectLangFromPath(path string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	return extToLanguage(ext)
}

// CountLinesWithEmbedded counts lines and extracts embedded code blocks (for Markdown/MDX)
func CountLinesWithEmbedded(path string) (model.LineMetrics, map[string]model.LineMetrics, error) {
	f, err := os.Open(path)
	if err != nil {
		return model.LineMetrics{}, nil, err
	}
	defer f.Close()

	// Get pooled buffer for binary check
	bufPtr := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufPtr)
	buf := *bufPtr

	// Read first chunk and check for binary
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return model.LineMetrics{}, nil, err
	}

	// Binary check: look for NUL byte in first 512 bytes
	checkLen := min(n, 512)
	for i := 0; i < checkLen; i++ {
		if buf[i] == 0 {
			return model.LineMetrics{}, nil, nil // binary file
		}
	}

	// Seek back to start
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return model.LineMetrics{}, nil, err
	}

	lang := detectLangFromPath(path)
	if lang == "Markdown" || lang == "MDX" {
		return countMarkdownWithEmbedded(f, bufPtr)
	}

	// Non-Markdown: use regular counting
	metrics := countLinesFromReader(f, lang, bufPtr)
	return metrics, nil, nil
}

// countMarkdownWithEmbedded parses Markdown and extracts fenced code blocks
func countMarkdownWithEmbedded(f *os.File, bufPtr *[]byte) (model.LineMetrics, map[string]model.LineMetrics, error) {
	scanner := bufio.NewScanner(f)
	scanner.Buffer(*bufPtr, 256*1024)

	var metrics model.LineMetrics
	embedded := make(map[string]model.LineMetrics)

	inCodeBlock := false
	codeBlockLang := ""
	var codeBlockLines []string

	for scanner.Scan() {
		metrics.Total++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check for fenced code block start/end
		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				// Starting a code block
				inCodeBlock = true
				codeBlockLang = strings.TrimPrefix(trimmed, "```")
				codeBlockLang = strings.TrimSpace(codeBlockLang)
				// Normalize language name
				codeBlockLang = normalizeCodeBlockLang(codeBlockLang)
				codeBlockLines = nil
				metrics.Code++ // the ``` line itself is "code" in Markdown
			} else {
				// Ending a code block
				inCodeBlock = false
				metrics.Code++ // the closing ``` line

				// Process accumulated code block
				if codeBlockLang != "" && len(codeBlockLines) > 0 {
					blockMetrics := countCodeBlockLines(codeBlockLines, codeBlockLang)
					existing := embedded[codeBlockLang]
					existing.Total += blockMetrics.Total
					existing.Code += blockMetrics.Code
					existing.Comments += blockMetrics.Comments
					existing.Blanks += blockMetrics.Blanks
					embedded[codeBlockLang] = existing
				}
				codeBlockLang = ""
			}
			continue
		}

		if inCodeBlock {
			// Inside code block - accumulate for later processing
			codeBlockLines = append(codeBlockLines, line)
			metrics.Code++ // code blocks count as code in Markdown
		} else if trimmed == "" {
			metrics.Blanks++
		} else if strings.HasPrefix(trimmed, "<!--") {
			metrics.Comments++
		} else {
			metrics.Code++ // prose is "code" in Markdown
		}
	}

	if len(embedded) == 0 {
		return metrics, nil, nil
	}
	return metrics, embedded, nil
}

// normalizeCodeBlockLang converts code fence language hints to canonical names
func normalizeCodeBlockLang(hint string) string {
	// remove common suffixes/annotations
	hint = strings.Split(hint, " ")[0] // "typescript jsx" -> "typescript"
	hint = strings.Trim(hint, "`")     // remove any stray backticks

	if hint == "" {
		return ""
	}

	hint = strings.ToLower(hint)

	// try extension-to-language map first (handles ts, py, rs, etc.)
	if lang := extToLanguage(hint); lang != "unknown" {
		return lang
	}

	// try matching language name case-insensitively
	for name := range languages {
		if strings.EqualFold(hint, name) {
			return name
		}
	}

	// capitalize first letter for unknown languages
	return strings.ToUpper(hint[:1]) + hint[1:]
}

// countCodeBlockLines counts lines within a code block using language-specific rules
func countCodeBlockLines(lines []string, lang string) model.LineMetrics {
	var metrics model.LineMetrics
	lineComment := getLineCommentMarker(lang)
	blockStart, blockEnd := getBlockCommentMarkers(lang)
	inBlockComment := false

	for _, line := range lines {
		metrics.Total++
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			metrics.Blanks++
			continue
		}

		isCode := false
		isComment := false

		// Handle block comments
		if blockStart != "" {
			if inBlockComment {
				if idx := strings.Index(trimmed, blockEnd); idx >= 0 {
					inBlockComment = false
					remainder := strings.TrimSpace(trimmed[idx+len(blockEnd):])
					if remainder != "" && !strings.HasPrefix(remainder, lineComment) {
						isCode = true
					} else {
						isComment = true
					}
				} else {
					isComment = true
				}
			} else if strings.Contains(trimmed, blockStart) {
				startIdx := strings.Index(trimmed, blockStart)
				beforeComment := strings.TrimSpace(trimmed[:startIdx])
				endIdx := strings.Index(trimmed[startIdx+len(blockStart):], blockEnd)

				if endIdx >= 0 {
					afterComment := strings.TrimSpace(trimmed[startIdx+len(blockStart)+endIdx+len(blockEnd):])
					if beforeComment != "" || afterComment != "" {
						isCode = true
					} else {
						isComment = true
					}
				} else {
					inBlockComment = true
					if beforeComment != "" {
						isCode = true
					} else {
						isComment = true
					}
				}
			} else if lineComment != "" && strings.HasPrefix(trimmed, lineComment) {
				isComment = true
			} else {
				isCode = true
			}
		} else {
			if lineComment != "" && strings.HasPrefix(trimmed, lineComment) {
				isComment = true
			} else {
				isCode = true
			}
		}

		if isCode {
			metrics.Code++
		} else if isComment {
			metrics.Comments++
		}
	}

	return metrics
}
