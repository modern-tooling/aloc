package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCountLOCFromBytes(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		lang     string
		want     int
	}{
		{
			name:     "simple go",
			filename: "test.go",
			content:  "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n",
			lang:     "Go",
			want:     4,
		},
		{
			name:     "with comments",
			filename: "test.go",
			content:  "package main\n\n// this is a comment\nfunc main() {}\n",
			lang:     "Go",
			want:     2,
		},
		{
			name:     "blank lines",
			filename: "test.go",
			content:  "package main\n\n\n\nfunc main() {}\n",
			lang:     "Go",
			want:     2,
		},
		{
			name:     "python with hash comments",
			filename: "test.py",
			content:  "# comment\nimport os\n\ndef main():\n    pass\n",
			lang:     "Python",
			want:     3,
		},
		{
			name:     "shell script with shebang",
			filename: "test.sh",
			content:  "#!/bin/bash\n# comment\necho hello\n",
			lang:     "Shell",
			want:     1, // shebang and comments are excluded
		},
		{
			name:     "shell script multiple commands",
			filename: "test.sh",
			content:  "#!/bin/bash\necho hello\necho world\n",
			lang:     "Shell",
			want:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, tt.filename)
			os.WriteFile(tmpFile, []byte(tt.content), 0644)

			got, err := CountLOC(tmpFile)
			if err != nil {
				t.Fatalf("CountLOC failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("CountLOC = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCountLOC_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.go")
	os.WriteFile(tmpFile, []byte(""), 0644)

	got, err := CountLOC(tmpFile)
	if err != nil {
		t.Fatalf("CountLOC failed: %v", err)
	}
	if got != 0 {
		t.Errorf("CountLOC = %d, want 0", got)
	}
}

func TestCountLOC_NonExistentFile(t *testing.T) {
	_, err := CountLOC("/nonexistent/file.go")
	if err == nil {
		t.Error("CountLOC should fail for nonexistent file")
	}
}

func TestCountLOC_BinaryFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "binary.bin")

	// create a file with NUL byte (binary indicator)
	content := []byte("hello\x00world")
	os.WriteFile(tmpFile, content, 0644)

	got, err := CountLOC(tmpFile)
	if err != nil {
		t.Fatalf("CountLOC failed: %v", err)
	}
	if got != 0 {
		t.Errorf("CountLOC for binary file = %d, want 0", got)
	}
}

func TestCountLOC_BinaryFileNulAtEnd(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "binary2.bin")

	// NUL byte within first 512 bytes should be detected
	content := make([]byte, 600)
	for i := range content {
		content[i] = 'a'
	}
	content[500] = 0 // NUL at position 500

	os.WriteFile(tmpFile, content, 0644)

	got, err := CountLOC(tmpFile)
	if err != nil {
		t.Fatalf("CountLOC failed: %v", err)
	}
	if got != 0 {
		t.Errorf("CountLOC for binary file = %d, want 0", got)
	}
}

func TestCountLOC_TextFileWithNulAfter512(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "text.txt")

	// NUL byte after first 512 bytes should not be detected as binary
	content := make([]byte, 600)
	for i := range content {
		content[i] = 'a'
	}
	content[len(content)-1] = '\n'
	content[550] = 0 // NUL at position 550, beyond check window

	os.WriteFile(tmpFile, content, 0644)

	got, err := CountLOC(tmpFile)
	if err != nil {
		t.Fatalf("CountLOC failed: %v", err)
	}
	// should count as text file with 1 line
	if got != 1 {
		t.Errorf("CountLOC for text file with late NUL = %d, want 1", got)
	}
}
