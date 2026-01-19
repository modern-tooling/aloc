package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestReportJSONRoundTrip(t *testing.T) {
	original := Report{
		Meta: Meta{
			SchemaVersion:    "1.0",
			GeneratedAt:      time.Now().UTC().Truncate(time.Second),
			Generator:        "aloc",
			GeneratorVersion: "0.1.0",
		},
		Summary: Summary{
			Files:     100,
			LOCTotal:  50000,
			Languages: 5,
		},
		Responsibilities: []Responsibility{
			{Role: RoleProd, LOC: 30000, Files: 60, Confidence: 0.95},
			{Role: RoleTest, LOC: 20000, Files: 40, Confidence: 0.92},
		},
		Ratios: Ratios{
			TestToProd:  0.67,
			InfraToProd: 0.10,
		},
		Languages: []LanguageComp{
			{Language: "Go", LOCTotal: 40000, Files: 80},
		},
		Confidence: ConfidenceInfo{
			AutoClassified: 0.92,
			Heuristic:      0.08,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Report
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Summary.Files != original.Summary.Files {
		t.Errorf("Summary.Files = %v, want %v", decoded.Summary.Files, original.Summary.Files)
	}
	if decoded.Summary.LOCTotal != original.Summary.LOCTotal {
		t.Errorf("Summary.LOCTotal = %v, want %v", decoded.Summary.LOCTotal, original.Summary.LOCTotal)
	}
	if len(decoded.Responsibilities) != len(original.Responsibilities) {
		t.Errorf("Responsibilities len = %v, want %v", len(decoded.Responsibilities), len(original.Responsibilities))
	}
	if decoded.Meta.SchemaVersion != original.Meta.SchemaVersion {
		t.Errorf("Meta.SchemaVersion = %v, want %v", decoded.Meta.SchemaVersion, original.Meta.SchemaVersion)
	}
}

func TestReportOptionalFieldsOmitted(t *testing.T) {
	report := Report{
		Meta: Meta{
			SchemaVersion: "1.0",
			GeneratedAt:   time.Now().UTC(),
		},
		Summary:          Summary{Files: 10},
		Responsibilities: []Responsibility{},
		Languages:        []LanguageComp{},
		// Trend and Files are nil
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal to map failed: %v", err)
	}

	if _, exists := decoded["trend"]; exists {
		t.Error("trend should be omitted when nil")
	}
	if _, exists := decoded["files"]; exists {
		t.Error("files should be omitted when nil")
	}
}

func TestReportWithTrend(t *testing.T) {
	report := Report{
		Meta: Meta{
			SchemaVersion: "1.0",
			GeneratedAt:   time.Now().UTC(),
		},
		Summary:          Summary{Files: 10},
		Responsibilities: []Responsibility{},
		Languages:        []LanguageComp{},
		Trend: &Trend{
			Window:         "30d",
			Sparkline:      []float32{0.1, 0.2, 0.3},
			Direction:      "up",
			Interpretation: "Test coverage increasing",
		},
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal to map failed: %v", err)
	}

	if _, exists := decoded["trend"]; !exists {
		t.Error("trend should be present when set")
	}
}

func TestReportWithFiles(t *testing.T) {
	report := Report{
		Meta: Meta{
			SchemaVersion: "1.0",
			GeneratedAt:   time.Now().UTC(),
		},
		Summary:          Summary{Files: 1},
		Responsibilities: []Responsibility{},
		Languages:        []LanguageComp{},
		Files: []*FileRecord{
			{Path: "main.go", LOC: 100, Role: RoleProd, Confidence: 0.9, Signals: []Signal{SignalPath}},
		},
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Report
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(decoded.Files) != 1 {
		t.Errorf("Files len = %v, want 1", len(decoded.Files))
	}
}

func TestMetaWithRepoInfo(t *testing.T) {
	meta := Meta{
		SchemaVersion:    "1.0",
		GeneratedAt:      time.Now().UTC(),
		Generator:        "aloc",
		GeneratorVersion: "0.1.0",
		Repo: &RepoInfo{
			Name:   "my-project",
			Commit: "abc123",
			Branch: "main",
			Root:   "/path/to/repo",
		},
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Meta
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Repo == nil {
		t.Fatal("Repo should not be nil")
	}
	if decoded.Repo.Name != "my-project" {
		t.Errorf("Repo.Name = %v, want my-project", decoded.Repo.Name)
	}
}

func TestMetaWithoutRepoInfo(t *testing.T) {
	meta := Meta{
		SchemaVersion: "1.0",
		GeneratedAt:   time.Now().UTC(),
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal to map failed: %v", err)
	}

	if _, exists := decoded["repo"]; exists {
		t.Error("repo should be omitted when nil")
	}
}

func TestResponsibilityWithBreakdown(t *testing.T) {
	resp := Responsibility{
		Role:       RoleTest,
		LOC:        10000,
		Files:      50,
		Confidence: 0.9,
		Breakdown: map[TestKind]float32{
			TestUnit:        0.6,
			TestIntegration: 0.3,
			TestE2E:         0.1,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Responsibility
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(decoded.Breakdown) != 3 {
		t.Errorf("Breakdown len = %v, want 3", len(decoded.Breakdown))
	}
	if decoded.Breakdown[TestUnit] != 0.6 {
		t.Errorf("Breakdown[unit] = %v, want 0.6", decoded.Breakdown[TestUnit])
	}
}

func TestResponsibilityWithNotes(t *testing.T) {
	resp := Responsibility{
		Role:       RoleProd,
		LOC:        5000,
		Files:      25,
		Confidence: 0.85,
		Notes:      []string{"Core business logic", "High complexity"},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Responsibility
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(decoded.Notes) != 2 {
		t.Errorf("Notes len = %v, want 2", len(decoded.Notes))
	}
}

func TestLanguageCompWithResponsibilities(t *testing.T) {
	lang := LanguageComp{
		Language: "Go",
		LOCTotal: 30000,
		Files:    60,
		Responsibilities: map[Role]int{
			RoleProd: 20000,
			RoleTest: 10000,
		},
	}

	data, err := json.Marshal(lang)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded LanguageComp
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Responsibilities[RoleProd] != 20000 {
		t.Errorf("Responsibilities[prod] = %v, want 20000", decoded.Responsibilities[RoleProd])
	}
}

func TestRatiosAllFields(t *testing.T) {
	ratios := Ratios{
		TestToProd:      0.67,
		InfraToProd:     0.10,
		DocsToProd:      0.05,
		GeneratedToProd: 0.15,
		ConfigToProd:    0.03,
	}

	data, err := json.Marshal(ratios)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Ratios
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.TestToProd != 0.67 {
		t.Errorf("TestToProd = %v, want 0.67", decoded.TestToProd)
	}
	if decoded.InfraToProd != 0.10 {
		t.Errorf("InfraToProd = %v, want 0.10", decoded.InfraToProd)
	}
}

func TestConfidenceInfoAllFields(t *testing.T) {
	conf := ConfidenceInfo{
		AutoClassified: 0.85,
		Heuristic:      0.10,
		Override:       0.05,
	}

	data, err := json.Marshal(conf)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded ConfidenceInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.AutoClassified != 0.85 {
		t.Errorf("AutoClassified = %v, want 0.85", decoded.AutoClassified)
	}
	if decoded.Heuristic != 0.10 {
		t.Errorf("Heuristic = %v, want 0.10", decoded.Heuristic)
	}
	if decoded.Override != 0.05 {
		t.Errorf("Override = %v, want 0.05", decoded.Override)
	}
}

func TestTrendJSONRoundTrip(t *testing.T) {
	original := Trend{
		Window:         "30d",
		Sparkline:      []float32{0.5, 0.55, 0.6, 0.58, 0.62},
		Direction:      "up",
		Interpretation: "Test coverage trending upward",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Trend
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Window != original.Window {
		t.Errorf("Window = %v, want %v", decoded.Window, original.Window)
	}
	if decoded.Direction != original.Direction {
		t.Errorf("Direction = %v, want %v", decoded.Direction, original.Direction)
	}
	if len(decoded.Sparkline) != len(original.Sparkline) {
		t.Errorf("Sparkline len = %v, want %v", len(decoded.Sparkline), len(original.Sparkline))
	}
}
