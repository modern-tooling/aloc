package effort

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

//go:embed profiles/*.json
var profilesFS embed.FS

// ProfileInfo contains metadata about a profile
type ProfileInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// LoadProfile loads a named profile from embedded JSON files and merges with defaults.
// If name is empty, returns the default profile ("faang").
// The profile is merged with DefaultModelConfig(), so only specified values are overridden.
func LoadProfile(name string) (*ModelConfig, error) {
	if name == "" {
		name = DefaultProfileName
	}

	filename := "profiles/" + name + ".json"
	data, err := profilesFS.ReadFile(filename)
	if err != nil {
		available := AvailableProfiles()
		return nil, fmt.Errorf("profile %q not found (available: %s)", name, strings.Join(available, ", "))
	}

	return parseProfileJSON(data)
}

// parseProfileJSON parses profile JSON and merges with defaults
func parseProfileJSON(data []byte) (*ModelConfig, error) {
	cfg := DefaultModelConfig()

	var override struct {
		Name        string                     `json:"name"`
		Description string                     `json:"description"`
		ModelConfig                            // embed all ModelConfig fields
	}

	if err := json.Unmarshal(data, &override); err != nil {
		return nil, fmt.Errorf("invalid profile JSON: %w", err)
	}

	// merge COCOMO models
	for k, v := range override.COCOMOModels {
		existing, ok := cfg.COCOMOModels[k]
		if ok {
			if v.A != 0 {
				existing.A = v.A
			}
			if v.B != 0 {
				existing.B = v.B
			}
			if v.C != 0 {
				existing.C = v.C
			}
			if v.D != 0 {
				existing.D = v.D
			}
			cfg.COCOMOModels[k] = existing
		} else {
			cfg.COCOMOModels[k] = v
		}
	}

	// merge skill bands
	for k, v := range override.SkillBands {
		existing, ok := cfg.SkillBands[k]
		if ok {
			if v.AnnualCostLow != 0 {
				existing.AnnualCostLow = v.AnnualCostLow
			}
			if v.AnnualCostHigh != 0 {
				existing.AnnualCostHigh = v.AnnualCostHigh
			}
			cfg.SkillBands[k] = existing
		} else {
			cfg.SkillBands[k] = v
		}
	}

	// merge other config fields using existing merge functions
	if override.VarianceMultipliers.Optimistic != 0 {
		cfg.VarianceMultipliers.Optimistic = override.VarianceMultipliers.Optimistic
	}
	if override.VarianceMultipliers.Pessimistic != 0 {
		cfg.VarianceMultipliers.Pessimistic = override.VarianceMultipliers.Pessimistic
	}

	mergeAINative(&cfg.AINative, &override.AINative)
	mergeTokenConfig(&cfg.TokenEstimation, &override.TokenEstimation)
	mergeTeamComposition(&cfg.TeamComposition, &override.TeamComposition)
	mergeAILeverageBySkill(&cfg.AILeverageBySkill, &override.AILeverageBySkill)

	if override.DefaultHumanCostMo != 0 {
		cfg.DefaultHumanCostMo = override.DefaultHumanCostMo
	}

	return cfg, nil
}

// AvailableProfiles returns list of embedded profile names
func AvailableProfiles() []string {
	entries, err := profilesFS.ReadDir("profiles")
	if err != nil {
		return []string{}
	}

	var profiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			name := strings.TrimSuffix(entry.Name(), ".json")
			profiles = append(profiles, name)
		}
	}
	return profiles
}

// GetProfileInfo returns metadata about a profile without loading full config
func GetProfileInfo(name string) (*ProfileInfo, error) {
	filename := "profiles/" + name + ".json"
	data, err := profilesFS.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("profile %q not found", name)
	}

	var info ProfileInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("invalid profile JSON: %w", err)
	}
	return &info, nil
}

// ListProfilesWithInfo returns all available profiles with their metadata
func ListProfilesWithInfo() []ProfileInfo {
	entries, err := profilesFS.ReadDir("profiles")
	if err != nil {
		return nil
	}

	var profiles []ProfileInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), path.Ext(entry.Name()))
		info, err := GetProfileInfo(name)
		if err != nil {
			continue
		}
		profiles = append(profiles, *info)
	}
	return profiles
}

// DefaultProfileName is the profile used when none is specified
const DefaultProfileName = "faang"
