package model

// LineMetrics contains detailed line counts
type LineMetrics struct {
	Total    int `json:"total"`    // raw line count
	Blanks   int `json:"blanks"`   // empty lines
	Comments int `json:"comments"` // comment-only lines
	Code     int `json:"code"`     // code lines (LOC)
}

// RawFile is the scanner output before semantic inference
type RawFile struct {
	Path         string
	Bytes        int64
	LOC          int                    // code lines (for backward compat)
	Lines        LineMetrics            // detailed line metrics
	LanguageHint string
	Embedded     map[string]LineMetrics // code blocks embedded in this file (e.g., Markdown)
}

// FileRecord is a file with semantic classification
type FileRecord struct {
	Path       string                 `json:"path"`
	LOC        int                    `json:"loc"`
	Lines      LineMetrics            `json:"lines,omitempty"`
	Language   string                 `json:"language"`
	Role       Role                   `json:"role"`
	SubRole    TestKind               `json:"sub_role,omitempty"`
	Confidence float32                `json:"confidence"`
	Signals    []Signal               `json:"signals"`
	Embedded   map[string]LineMetrics `json:"embedded,omitempty"` // embedded code blocks by language
}
