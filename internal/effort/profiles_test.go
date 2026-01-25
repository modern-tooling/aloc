package effort

import (
	"testing"
)

func TestLoadProfile_FAANG(t *testing.T) {
	cfg, err := LoadProfile("faang")
	if err != nil {
		t.Fatalf("failed to load faang profile: %v", err)
	}

	// verify COCOMO coefficients are calibrated for 80 LOC/day
	organic := cfg.COCOMOModels["organic"]
	if organic.A != 0.43 {
		t.Errorf("organic.A = %v, want 0.43", organic.A)
	}
	if organic.B != 1.05 {
		t.Errorf("organic.B = %v, want 1.05", organic.B)
	}

	semiDetached := cfg.COCOMOModels["semi-detached"]
	if semiDetached.A != 0.53 {
		t.Errorf("semi-detached.A = %v, want 0.53", semiDetached.A)
	}

	embedded := cfg.COCOMOModels["embedded"]
	if embedded.A != 0.64 {
		t.Errorf("embedded.A = %v, want 0.64", embedded.A)
	}
}

func TestLoadProfile_DefaultIsFAANG(t *testing.T) {
	cfg, err := LoadProfile("")
	if err != nil {
		t.Fatalf("failed to load default profile: %v", err)
	}

	// default should be faang profile
	organic := cfg.COCOMOModels["organic"]
	if organic.A != 0.43 {
		t.Errorf("default profile organic.A = %v, want 0.43 (faang)", organic.A)
	}
}

func TestLoadProfile_NotFound(t *testing.T) {
	_, err := LoadProfile("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent profile, got nil")
	}
}

func TestAvailableProfiles(t *testing.T) {
	profiles := AvailableProfiles()
	if len(profiles) == 0 {
		t.Error("expected at least one profile, got none")
	}

	// faang should be in the list
	found := false
	for _, p := range profiles {
		if p == "faang" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("faang not found in available profiles: %v", profiles)
	}
}

func TestGetProfileInfo(t *testing.T) {
	info, err := GetProfileInfo("faang")
	if err != nil {
		t.Fatalf("failed to get faang profile info: %v", err)
	}

	if info.Name != "faang" {
		t.Errorf("info.Name = %q, want \"faang\"", info.Name)
	}
	if info.Description == "" {
		t.Error("expected non-empty description")
	}
}

func TestFAANGProfile_122KLOC(t *testing.T) {
	// set the faang profile as active
	cfg, err := LoadProfile("faang")
	if err != nil {
		t.Fatalf("failed to load faang profile: %v", err)
	}
	SetModelConfig(cfg)
	defer ResetModelConfig()

	// test with 122K LOC (ox-cli example from plan)
	loc := 122000
	estimate := CalculateHumanEffort(loc, COCOMOOptions{CostPerMonth: 15000})

	// with a=0.43, b=1.05 for organic:
	// effort = 0.43 * 122^1.05 ≈ 65-67 PM
	// schedule = 2.5 * 66^0.38 ≈ 13-14 months
	// team = 66 / 13.5 ≈ 5 engineers

	if estimate.EffortPersonMonths < 60 || estimate.EffortPersonMonths > 75 {
		t.Errorf("effort = %.1f PM, want 60-75 PM for 122K LOC", estimate.EffortPersonMonths)
	}

	if estimate.ScheduleMonths < 12 || estimate.ScheduleMonths > 16 {
		t.Errorf("schedule = %.1f months, want 12-16 months for 122K LOC", estimate.ScheduleMonths)
	}

	if estimate.TeamSize < 4 || estimate.TeamSize > 7 {
		t.Errorf("team = %.1f, want 4-7 engineers for 122K LOC", estimate.TeamSize)
	}
}

func TestDefaultProfileName(t *testing.T) {
	if DefaultProfileName != "faang" {
		t.Errorf("DefaultProfileName = %q, want \"faang\"", DefaultProfileName)
	}
}
