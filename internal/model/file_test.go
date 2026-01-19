package model

import (
	"encoding/json"
	"testing"
)

func TestFileRecordJSONRoundTrip(t *testing.T) {
	original := FileRecord{
		Path:       "internal/auth/login.go",
		LOC:        412,
		Language:   "Go",
		Role:       RoleProd,
		Confidence: 0.94,
		Signals:    []Signal{SignalPath, SignalExtension},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded FileRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Path != original.Path {
		t.Errorf("Path = %v, want %v", decoded.Path, original.Path)
	}
	if decoded.LOC != original.LOC {
		t.Errorf("LOC = %v, want %v", decoded.LOC, original.LOC)
	}
	if decoded.Role != original.Role {
		t.Errorf("Role = %v, want %v", decoded.Role, original.Role)
	}
	if decoded.Language != original.Language {
		t.Errorf("Language = %v, want %v", decoded.Language, original.Language)
	}
	if decoded.Confidence != original.Confidence {
		t.Errorf("Confidence = %v, want %v", decoded.Confidence, original.Confidence)
	}
	if len(decoded.Signals) != len(original.Signals) {
		t.Errorf("Signals len = %v, want %v", len(decoded.Signals), len(original.Signals))
	}
}

func TestFileRecordSubRoleOmitEmpty(t *testing.T) {
	record := FileRecord{
		Path:       "main.go",
		LOC:        100,
		Language:   "Go",
		Role:       RoleProd,
		SubRole:    "", // empty
		Confidence: 0.9,
		Signals:    []Signal{SignalExtension},
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// SubRole should be omitted when empty
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal to map failed: %v", err)
	}

	if _, exists := decoded["sub_role"]; exists {
		t.Error("sub_role should be omitted when empty")
	}
}

func TestFileRecordWithSubRole(t *testing.T) {
	record := FileRecord{
		Path:       "main_test.go",
		LOC:        100,
		Language:   "Go",
		Role:       RoleTest,
		SubRole:    TestUnit,
		Confidence: 0.95,
		Signals:    []Signal{SignalFilename},
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal to map failed: %v", err)
	}

	if decoded["sub_role"] != "unit" {
		t.Errorf("sub_role = %v, want unit", decoded["sub_role"])
	}
}

func TestFileRecordAllTestKinds(t *testing.T) {
	testKinds := []TestKind{TestUnit, TestIntegration, TestE2E, TestContract, TestFixture}

	for _, kind := range testKinds {
		record := FileRecord{
			Path:       "test_file.go",
			LOC:        50,
			Language:   "Go",
			Role:       RoleTest,
			SubRole:    kind,
			Confidence: 0.9,
			Signals:    []Signal{SignalPath},
		}

		data, err := json.Marshal(record)
		if err != nil {
			t.Fatalf("json.Marshal failed for %v: %v", kind, err)
		}

		var decoded FileRecord
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("json.Unmarshal failed for %v: %v", kind, err)
		}

		if decoded.SubRole != kind {
			t.Errorf("SubRole = %v, want %v", decoded.SubRole, kind)
		}
	}
}

func TestRawFileFields(t *testing.T) {
	raw := RawFile{
		Path:         "src/main.go",
		Bytes:        1024,
		LOC:          50,
		LanguageHint: "Go",
	}

	if raw.Path != "src/main.go" {
		t.Errorf("Path = %v, want src/main.go", raw.Path)
	}
	if raw.Bytes != 1024 {
		t.Errorf("Bytes = %v, want 1024", raw.Bytes)
	}
	if raw.LOC != 50 {
		t.Errorf("LOC = %v, want 50", raw.LOC)
	}
	if raw.LanguageHint != "Go" {
		t.Errorf("LanguageHint = %v, want Go", raw.LanguageHint)
	}
}

func TestFileRecordJSONFieldNames(t *testing.T) {
	record := FileRecord{
		Path:       "test.go",
		LOC:        100,
		Language:   "Go",
		Role:       RoleProd,
		SubRole:    TestUnit,
		Confidence: 0.95,
		Signals:    []Signal{SignalPath},
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal to map failed: %v", err)
	}

	expectedFields := []string{"path", "loc", "language", "role", "sub_role", "confidence", "signals"}
	for _, field := range expectedFields {
		if _, exists := decoded[field]; !exists {
			t.Errorf("expected JSON field %q not found", field)
		}
	}
}
