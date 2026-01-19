package effort

import (
	"testing"
)

func TestCalculateHumanEffort(t *testing.T) {
	tests := []struct {
		name    string
		loc     int
		opts    COCOMOOptions
		wantMin float64 // minimum expected cost
		wantMax float64 // maximum expected cost
	}{
		{
			name:    "small project 1k LOC",
			loc:     1000,
			opts:    DefaultCOCOMOOptions(),
			wantMin: 30000,
			wantMax: 50000,
		},
		{
			name:    "medium project 10k LOC",
			loc:     10000,
			opts:    DefaultCOCOMOOptions(),
			wantMin: 350000,
			wantMax: 500000,
		},
		{
			name:    "large project 100k LOC",
			loc:     100000,
			opts:    DefaultCOCOMOOptions(),
			wantMin: 4000000,
			wantMax: 6000000,
		},
		{
			name:    "zero LOC",
			loc:     0,
			opts:    DefaultCOCOMOOptions(),
			wantMin: 0,
			wantMax: 0,
		},
		{
			name: "custom cost per month",
			loc:  10000,
			opts: COCOMOOptions{
				CostPerMonth: 20000,
				Model:        "organic",
			},
			wantMin: 450000,
			wantMax: 700000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateHumanEffort(tt.loc, tt.opts)

			if result.EstimatedCost < tt.wantMin || result.EstimatedCost > tt.wantMax {
				t.Errorf("EstimatedCost = %v, want between %v and %v",
					result.EstimatedCost, tt.wantMin, tt.wantMax)
			}

			if tt.loc > 0 {
				if result.ScheduleMonths <= 0 {
					t.Error("ScheduleMonths should be positive for non-zero LOC")
				}
				if result.TeamSize <= 0 {
					t.Error("TeamSize should be positive for non-zero LOC")
				}
			}
		})
	}
}

func TestCOCOMOModels(t *testing.T) {
	loc := 50000

	organic := CalculateHumanEffort(loc, COCOMOOptions{Model: "organic"})
	semiDetached := CalculateHumanEffort(loc, COCOMOOptions{Model: "semi-detached"})
	embedded := CalculateHumanEffort(loc, COCOMOOptions{Model: "embedded"})

	// Embedded should require more effort than semi-detached, which should require more than organic
	if organic.EffortPersonMonths >= semiDetached.EffortPersonMonths {
		t.Error("Semi-detached should require more effort than organic")
	}
	if semiDetached.EffortPersonMonths >= embedded.EffortPersonMonths {
		t.Error("Embedded should require more effort than semi-detached")
	}
}

func TestDefaultCOCOMOOptions(t *testing.T) {
	opts := DefaultCOCOMOOptions()

	if opts.CostPerMonth != 15000 {
		t.Errorf("CostPerMonth = %v, want 15000", opts.CostPerMonth)
	}
	if opts.Model != "organic" {
		t.Errorf("Model = %v, want organic", opts.Model)
	}
}

func TestCalculateHumanEffortDefaultsZeroOpts(t *testing.T) {
	// passing zero-value options should use defaults
	result := CalculateHumanEffort(10000, COCOMOOptions{})

	if result.EstimatedCost <= 0 {
		t.Error("EstimatedCost should be positive with default options")
	}
	if result.Model != "COCOMO Basic organic" {
		t.Errorf("Model = %v, want 'COCOMO Basic organic'", result.Model)
	}
}

func TestCalculateHumanEffortInvalidModel(t *testing.T) {
	result := CalculateHumanEffort(10000, COCOMOOptions{Model: "invalid-model"})

	// should fall back to organic
	if result.Model != "COCOMO Basic organic" {
		t.Errorf("Model = %v, want 'COCOMO Basic organic' for invalid model", result.Model)
	}
}

func TestCalculateHumanEffortKLOC(t *testing.T) {
	result := CalculateHumanEffort(5000, DefaultCOCOMOOptions())

	expectedKLOC := 5.0
	if result.KLOC != expectedKLOC {
		t.Errorf("KLOC = %v, want %v", result.KLOC, expectedKLOC)
	}
}
