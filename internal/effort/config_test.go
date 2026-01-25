package effort

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultModelConfig(t *testing.T) {
	cfg := DefaultModelConfig()

	// verify COCOMO models
	if len(cfg.COCOMOModels) != 3 {
		t.Errorf("expected 3 COCOMO models, got %d", len(cfg.COCOMOModels))
	}
	organic, ok := cfg.COCOMOModels["organic"]
	if !ok {
		t.Fatal("missing organic COCOMO model")
	}
	if organic.A != 2.4 || organic.B != 1.05 {
		t.Errorf("unexpected organic coefficients: a=%v, b=%v", organic.A, organic.B)
	}

	// verify variance multipliers
	if cfg.VarianceMultipliers.Optimistic != 0.85 {
		t.Errorf("expected optimistic=0.85, got %v", cfg.VarianceMultipliers.Optimistic)
	}
	if cfg.VarianceMultipliers.Pessimistic != 1.30 {
		t.Errorf("expected pessimistic=1.30, got %v", cfg.VarianceMultipliers.Pessimistic)
	}

	// verify AI native config
	if cfg.AINative.MinimumTeamSize != 2 {
		t.Errorf("expected minimum team size=2, got %v", cfg.AINative.MinimumTeamSize)
	}

	// verify token estimation
	if cfg.TokenEstimation.AvgLOCPerFile != 160 {
		t.Errorf("expected avgLOCPerFile=160, got %d", cfg.TokenEstimation.AvgLOCPerFile)
	}

	// verify skill bands
	if len(cfg.SkillBands) != 6 {
		t.Errorf("expected 6 skill bands, got %d", len(cfg.SkillBands))
	}

	// verify elite reference
	if cfg.EliteReference.ObservedMonths != 1.5 {
		t.Errorf("expected observed months=1.5, got %v", cfg.EliteReference.ObservedMonths)
	}

	// verify default human cost
	if cfg.DefaultHumanCostMo != 15000 {
		t.Errorf("expected default human cost=15000, got %v", cfg.DefaultHumanCostMo)
	}
}

func TestLoadModelConfig_Merge(t *testing.T) {
	// create temp dir
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "model.json")

	// write partial config
	configJSON := `{
		"cocomo_models": {
			"organic": { "a": 3.0 }
		},
		"variance_multipliers": {
			"optimistic": 0.90
		},
		"ai_native": {
			"schedule_factor_low": 0.30
		},
		"skill_bands": {
			"staff": { "annual_cost_low": 350000 }
		},
		"default_human_cost_per_month": 20000
	}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := LoadModelConfig(configPath)
	if err != nil {
		t.Fatalf("LoadModelConfig failed: %v", err)
	}

	// verify partial merge for COCOMO
	organic := cfg.COCOMOModels["organic"]
	if organic.A != 3.0 {
		t.Errorf("expected organic.A=3.0, got %v", organic.A)
	}
	// B should remain default since not overridden
	if organic.B != 1.05 {
		t.Errorf("expected organic.B=1.05 (default), got %v", organic.B)
	}

	// verify variance override
	if cfg.VarianceMultipliers.Optimistic != 0.90 {
		t.Errorf("expected optimistic=0.90, got %v", cfg.VarianceMultipliers.Optimistic)
	}
	// pessimistic should remain default
	if cfg.VarianceMultipliers.Pessimistic != 1.30 {
		t.Errorf("expected pessimistic=1.30 (default), got %v", cfg.VarianceMultipliers.Pessimistic)
	}

	// verify AI native partial override
	if cfg.AINative.ScheduleFactorLow != 0.30 {
		t.Errorf("expected schedule_factor_low=0.30, got %v", cfg.AINative.ScheduleFactorLow)
	}
	// others should remain default
	if cfg.AINative.ScheduleFactorHigh != 0.60 {
		t.Errorf("expected schedule_factor_high=0.60 (default), got %v", cfg.AINative.ScheduleFactorHigh)
	}

	// verify skill band partial override
	staff := cfg.SkillBands["staff"]
	if staff.AnnualCostLow != 350000 {
		t.Errorf("expected staff.AnnualCostLow=350000, got %v", staff.AnnualCostLow)
	}
	// high should remain default
	if staff.AnnualCostHigh != 550000 {
		t.Errorf("expected staff.AnnualCostHigh=550000 (default), got %v", staff.AnnualCostHigh)
	}

	// verify default human cost override
	if cfg.DefaultHumanCostMo != 20000 {
		t.Errorf("expected default human cost=20000, got %v", cfg.DefaultHumanCostMo)
	}
}

func TestLoadModelConfig_FileNotFound(t *testing.T) {
	_, err := LoadModelConfig("/nonexistent/path/model.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadModelConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(configPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := LoadModelConfig(configPath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGetSetModelConfig(t *testing.T) {
	// reset to ensure clean state
	ResetModelConfig()

	// get should return default
	cfg := GetModelConfig()
	if cfg.DefaultHumanCostMo != 15000 {
		t.Errorf("expected default config, got human cost=%v", cfg.DefaultHumanCostMo)
	}

	// set custom config
	custom := DefaultModelConfig()
	custom.DefaultHumanCostMo = 25000
	SetModelConfig(custom)

	// get should return custom
	cfg = GetModelConfig()
	if cfg.DefaultHumanCostMo != 25000 {
		t.Errorf("expected custom config with human cost=25000, got %v", cfg.DefaultHumanCostMo)
	}

	// reset for other tests
	ResetModelConfig()
}

func TestCalculateHumanEffort_WithConfig(t *testing.T) {
	ResetModelConfig()

	// create custom config with different COCOMO coefficients
	custom := DefaultModelConfig()
	custom.COCOMOModels["organic"] = COCOMOCoeffs{A: 3.0, B: 1.10, C: 2.5, D: 0.38}
	SetModelConfig(custom)

	// calculate with custom config
	result := CalculateHumanEffort(10000, COCOMOOptions{
		CostPerMonth: 20000,
		Model:        "organic",
	})

	// verify calculation uses custom coefficients
	// with A=3.0 instead of 2.4, effort should be higher
	if result.EffortPersonMonths < 30 {
		t.Errorf("expected higher effort with A=3.0, got %v", result.EffortPersonMonths)
	}

	ResetModelConfig()
}

func TestCalculateAgenticTeam_WithConfig(t *testing.T) {
	ResetModelConfig()

	// create custom config with different AI-native multipliers
	custom := DefaultModelConfig()
	custom.AINative.ScheduleFactorLow = 0.20 // more aggressive
	custom.AINative.ScheduleFactorHigh = 0.40
	SetModelConfig(custom)

	// calculate
	result := CalculateAgenticTeam(10000, COCOMOOptions{CostPerMonth: 15000})

	// with more aggressive factors, schedule should be shorter
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// verify AI leverage is being applied
	if result.AILeverage.Low <= 1.0 {
		t.Errorf("expected AI leverage > 1.0, got %v", result.AILeverage.Low)
	}

	// verify effective capacity > team size (due to AI leverage)
	if result.EffectiveCapacity.Low <= result.TeamSize.Low {
		t.Errorf("expected effective capacity > team size, got %v <= %v",
			result.EffectiveCapacity.Low, result.TeamSize.Low)
	}

	// schedule should be shorter than conventional × factor due to leverage
	conv := CalculateConventionalTeam(10000, COCOMOOptions{CostPerMonth: 15000})
	baseExpected := conv.ScheduleMo.Low * 0.20
	if result.ScheduleMo.Low >= baseExpected {
		t.Errorf("expected schedule < %v (leverage should compress further), got %v",
			baseExpected, result.ScheduleMo.Low)
	}

	ResetModelConfig()
}

func TestAllBands_ReturnsFive(t *testing.T) {
	ResetModelConfig()

	bands := AllBands()
	if len(bands) != 5 {
		t.Errorf("expected 5 bands (excluding Distinguished), got %d", len(bands))
	}

	expectedOrder := []string{
		"Junior Engineer",
		"Senior Engineer",
		"Staff Engineer",
		"Principal Engineer",
		"Senior Principal Engineer",
	}

	for i, band := range bands {
		if band.Name != expectedOrder[i] {
			t.Errorf("band %d: expected %q, got %q", i, expectedOrder[i], band.Name)
		}
	}

	ResetModelConfig()
}

func TestAllBandsIncludingDistinguished_ReturnsSix(t *testing.T) {
	ResetModelConfig()

	bands := AllBandsIncludingDistinguished()
	if len(bands) != 6 {
		t.Errorf("expected 6 bands, got %d", len(bands))
	}

	// verify last one is Distinguished
	last := bands[len(bands)-1]
	if last.Name != "Distinguished Engineer" {
		t.Errorf("expected last band to be Distinguished, got %q", last.Name)
	}

	ResetModelConfig()
}

func TestSkillBands_UsesConfig(t *testing.T) {
	ResetModelConfig()

	// create custom config
	custom := DefaultModelConfig()
	custom.SkillBands["principal"] = SkillBandConfig{
		AnnualCostLow:  500000,
		AnnualCostHigh: 1200000,
	}
	SetModelConfig(custom)

	band := BandPrincipal()
	if band.AnnualCostLow != 500000 {
		t.Errorf("expected annual cost low=500000, got %v", band.AnnualCostLow)
	}
	if band.AnnualCostHigh != 1200000 {
		t.Errorf("expected annual cost high=1200000, got %v", band.AnnualCostHigh)
	}

	ResetModelConfig()
}

func TestEliteReference_UsesConfig(t *testing.T) {
	ResetModelConfig()

	// create custom config
	custom := DefaultModelConfig()
	custom.EliteReference.ObservedMonths = 2.0
	custom.EliteReference.ObservedAISpend = 8000
	SetModelConfig(custom)

	ref := CalculateEliteOperatorReference(1000000)
	if ref.Months != 2.0 {
		t.Errorf("expected months=2.0, got %v", ref.Months)
	}
	if ref.AISpend != 8000 {
		t.Errorf("expected AI spend=8000, got %v", ref.AISpend)
	}

	ResetModelConfig()
}
