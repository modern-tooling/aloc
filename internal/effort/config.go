package effort

import (
	"encoding/json"
	"os"
	"sync"
)

// COCOMOCoeffs represents coefficients for a COCOMO model type
type COCOMOCoeffs struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
	C float64 `json:"c"`
	D float64 `json:"d"`
}

// VarianceConfig controls optimistic/pessimistic multipliers
type VarianceConfig struct {
	Optimistic  float64 `json:"optimistic"`
	Pessimistic float64 `json:"pessimistic"`
}

// AINativeConfig controls AI-native team estimation parameters
type AINativeConfig struct {
	ScheduleFactorLow       float64 `json:"schedule_factor_low"`
	ScheduleFactorHigh      float64 `json:"schedule_factor_high"`
	TeamSizeFactorLow       float64 `json:"team_size_factor_low"`
	TeamSizeFactorHigh      float64 `json:"team_size_factor_high"`
	MinimumTeamSize         float64 `json:"minimum_team_size"`
	ToolingMonthlyLow       float64 `json:"tooling_monthly_low"`
	ToolingMonthlyHigh      float64 `json:"tooling_monthly_high"`
	LargeCodebaseThresholdK float64 `json:"large_codebase_threshold_kloc"`
	ToolingMonthlyLowLarge  float64 `json:"tooling_monthly_low_large"`
	ToolingMonthlyHighLarge float64 `json:"tooling_monthly_high_large"`
}

// TokenConfig controls token estimation parameters
type TokenConfig struct {
	AvgLOCPerFile     int     `json:"avg_loc_per_file"`
	IterationsPerFile int     `json:"iterations_per_file"`
	ContextPerCall    int     `json:"context_per_call"`
	OutputPerCall     int     `json:"output_per_call"`
	CharsPerToken     float64 `json:"chars_per_token"`
}

// SkillBandConfig represents cost ranges for a skill band
type SkillBandConfig struct {
	AnnualCostLow  float64 `json:"annual_cost_low"`
	AnnualCostHigh float64 `json:"annual_cost_high"`
}

// EliteRefConfig controls elite operator reference parameters
type EliteRefConfig struct {
	ObservedMonths  float64 `json:"observed_months"`
	ObservedAISpend float64 `json:"observed_ai_spend"`
}

// TeamCompositionConfig defines the mix of engineering levels
// Ratios should sum to 1.0
type TeamCompositionConfig struct {
	Junior          float64 `json:"junior"`
	Senior          float64 `json:"senior"`
	Staff           float64 `json:"staff"`
	Principal       float64 `json:"principal"`
	SeniorPrincipal float64 `json:"senior_principal"`
	Distinguished   float64 `json:"distinguished"`
}

// TeamCompositionBySize maps team size thresholds to composition
// Use this for more granular control based on organization size
type TeamCompositionBySize struct {
	Small      TeamCompositionConfig `json:"small"`      // 5-20 engineers
	Medium     TeamCompositionConfig `json:"medium"`     // 50-100 engineers
	Large      TeamCompositionConfig `json:"large"`      // 200-500 engineers
	Enterprise TeamCompositionConfig `json:"enterprise"` // 1000+ engineers
	Mega       TeamCompositionConfig `json:"mega"`       // 2000+ engineers
}

// AILeverageBySkill defines productivity multipliers for AI-assisted work by skill level.
// Principal+ engineers get more leverage from AI because they can:
// - Decompose work effectively into agent-scoped tasks
// - Recognize incorrect AI output quickly
// - Orchestrate multiple specialized agents
// - Design architecture while agents implement
type AILeverageBySkill struct {
	Junior          float64 `json:"junior"`
	Senior          float64 `json:"senior"`
	Staff           float64 `json:"staff"`
	Principal       float64 `json:"principal"`
	SeniorPrincipal float64 `json:"senior_principal"`
	Distinguished   float64 `json:"distinguished"`
}

// ModelConfig contains all configurable effort model parameters
type ModelConfig struct {
	COCOMOModels          map[string]COCOMOCoeffs    `json:"cocomo_models"`
	VarianceMultipliers   VarianceConfig             `json:"variance_multipliers"`
	AINative              AINativeConfig             `json:"ai_native"`
	TokenEstimation       TokenConfig                `json:"token_estimation"`
	SkillBands            map[string]SkillBandConfig `json:"skill_bands"`
	EliteReference        EliteRefConfig             `json:"elite_reference"`
	TeamComposition       TeamCompositionConfig      `json:"team_composition"`
	TeamCompositionBySize *TeamCompositionBySize     `json:"team_composition_by_size,omitempty"`
	AILeverageBySkill     AILeverageBySkill          `json:"ai_leverage_by_skill"`
	DefaultHumanCostMo    float64                    `json:"default_human_cost_per_month"`
}

// DefaultModelConfig returns the default configuration with all hardcoded values
func DefaultModelConfig() *ModelConfig {
	return &ModelConfig{
		COCOMOModels: map[string]COCOMOCoeffs{
			"organic":       {A: 2.4, B: 1.05, C: 2.5, D: 0.38},
			"semi-detached": {A: 3.0, B: 1.12, C: 2.5, D: 0.35},
			"embedded":      {A: 3.6, B: 1.20, C: 2.5, D: 0.32},
		},
		VarianceMultipliers: VarianceConfig{
			Optimistic:  0.85,
			Pessimistic: 1.30,
		},
		AINative: AINativeConfig{
			ScheduleFactorLow:       0.40,
			ScheduleFactorHigh:      0.60,
			TeamSizeFactorLow:       0.33,
			TeamSizeFactorHigh:      0.50,
			MinimumTeamSize:         2,
			ToolingMonthlyLow:       2000,
			ToolingMonthlyHigh:      10000,
			LargeCodebaseThresholdK: 100,
			ToolingMonthlyLowLarge:  5000,
			ToolingMonthlyHighLarge: 20000,
		},
		TokenEstimation: TokenConfig{
			AvgLOCPerFile:     160,
			IterationsPerFile: 10,
			ContextPerCall:    20000,
			OutputPerCall:     2000,
			CharsPerToken:     4.0,
		},
		SkillBands: map[string]SkillBandConfig{
			"junior":           {AnnualCostLow: 130000, AnnualCostHigh: 200000},
			"senior":           {AnnualCostLow: 180000, AnnualCostHigh: 350000},
			"staff":            {AnnualCostLow: 280000, AnnualCostHigh: 550000},
			"principal":        {AnnualCostLow: 450000, AnnualCostHigh: 1100000},
			"senior_principal": {AnnualCostLow: 1000000, AnnualCostHigh: 2200000},
			"distinguished":    {AnnualCostLow: 2500000, AnnualCostHigh: 5000000},
		},
		EliteReference: EliteRefConfig{
			ObservedMonths:  1.5,
			ObservedAISpend: 6000,
		},
		// Default composition for medium-sized teams (50-200 engineers)
		TeamComposition: TeamCompositionConfig{
			Junior:          0.20,
			Senior:          0.45,
			Staff:           0.22,
			Principal:       0.10,
			SeniorPrincipal: 0.03,
			Distinguished:   0.00, // most teams don't have this level
		},
		// AI leverage multipliers by skill level
		// Principal+ engineers get more leverage because they can orchestrate agents effectively
		// These are conservative estimates; actual leverage may be higher for skilled practitioners
		// Research basis: GitHub Copilot (55% faster), MIT/McKinsey studies (40-300% gains)
		// Orchestration leverage for senior engineers is largely unmeasured but observed to be substantial
		AILeverageBySkill: AILeverageBySkill{
			Junior:          1.5,  // basic autocomplete gains, struggles to validate output
			Senior:          2.5,  // effective with copilot-style tools, can review AI code
			Staff:           4.0,  // can orchestrate agent workflows, designs for AI implementation
			Principal:       6.0,  // effective multi-agent orchestration, rapid validation
			SeniorPrincipal: 8.0,  // designs systems for agent swarms, maximizes parallelization
			Distinguished:   10.0, // organization-scale AI leverage, defines patterns others follow
		},
		DefaultHumanCostMo: 15000,
	}
}

// LoadModelConfig loads configuration from a JSON file and merges with defaults
func LoadModelConfig(path string) (*ModelConfig, error) {
	cfg := DefaultModelConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// unmarshal into a map first to handle partial overrides
	var override ModelConfig
	if err := json.Unmarshal(data, &override); err != nil {
		return nil, err
	}

	// merge COCOMO models
	for k, v := range override.COCOMOModels {
		existing, ok := cfg.COCOMOModels[k]
		if ok {
			// partial override: only replace non-zero values
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

	// merge variance multipliers
	if override.VarianceMultipliers.Optimistic != 0 {
		cfg.VarianceMultipliers.Optimistic = override.VarianceMultipliers.Optimistic
	}
	if override.VarianceMultipliers.Pessimistic != 0 {
		cfg.VarianceMultipliers.Pessimistic = override.VarianceMultipliers.Pessimistic
	}

	// merge AI native config
	mergeAINative(&cfg.AINative, &override.AINative)

	// merge token estimation
	mergeTokenConfig(&cfg.TokenEstimation, &override.TokenEstimation)

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

	// merge elite reference
	if override.EliteReference.ObservedMonths != 0 {
		cfg.EliteReference.ObservedMonths = override.EliteReference.ObservedMonths
	}
	if override.EliteReference.ObservedAISpend != 0 {
		cfg.EliteReference.ObservedAISpend = override.EliteReference.ObservedAISpend
	}

	// merge team composition
	mergeTeamComposition(&cfg.TeamComposition, &override.TeamComposition)

	// merge team composition by size (if provided)
	if override.TeamCompositionBySize != nil {
		if cfg.TeamCompositionBySize == nil {
			cfg.TeamCompositionBySize = DefaultTeamCompositionBySize()
		}
		mergeTeamCompositionBySize(cfg.TeamCompositionBySize, override.TeamCompositionBySize)
	}

	// merge AI leverage by skill
	mergeAILeverageBySkill(&cfg.AILeverageBySkill, &override.AILeverageBySkill)

	// merge default human cost
	if override.DefaultHumanCostMo != 0 {
		cfg.DefaultHumanCostMo = override.DefaultHumanCostMo
	}

	return cfg, nil
}

func mergeAINative(dst, src *AINativeConfig) {
	if src.ScheduleFactorLow != 0 {
		dst.ScheduleFactorLow = src.ScheduleFactorLow
	}
	if src.ScheduleFactorHigh != 0 {
		dst.ScheduleFactorHigh = src.ScheduleFactorHigh
	}
	if src.TeamSizeFactorLow != 0 {
		dst.TeamSizeFactorLow = src.TeamSizeFactorLow
	}
	if src.TeamSizeFactorHigh != 0 {
		dst.TeamSizeFactorHigh = src.TeamSizeFactorHigh
	}
	if src.MinimumTeamSize != 0 {
		dst.MinimumTeamSize = src.MinimumTeamSize
	}
	if src.ToolingMonthlyLow != 0 {
		dst.ToolingMonthlyLow = src.ToolingMonthlyLow
	}
	if src.ToolingMonthlyHigh != 0 {
		dst.ToolingMonthlyHigh = src.ToolingMonthlyHigh
	}
	if src.LargeCodebaseThresholdK != 0 {
		dst.LargeCodebaseThresholdK = src.LargeCodebaseThresholdK
	}
	if src.ToolingMonthlyLowLarge != 0 {
		dst.ToolingMonthlyLowLarge = src.ToolingMonthlyLowLarge
	}
	if src.ToolingMonthlyHighLarge != 0 {
		dst.ToolingMonthlyHighLarge = src.ToolingMonthlyHighLarge
	}
}

func mergeTokenConfig(dst, src *TokenConfig) {
	if src.AvgLOCPerFile != 0 {
		dst.AvgLOCPerFile = src.AvgLOCPerFile
	}
	if src.IterationsPerFile != 0 {
		dst.IterationsPerFile = src.IterationsPerFile
	}
	if src.ContextPerCall != 0 {
		dst.ContextPerCall = src.ContextPerCall
	}
	if src.OutputPerCall != 0 {
		dst.OutputPerCall = src.OutputPerCall
	}
	if src.CharsPerToken != 0 {
		dst.CharsPerToken = src.CharsPerToken
	}
}

func mergeTeamComposition(dst, src *TeamCompositionConfig) {
	if src.Junior != 0 {
		dst.Junior = src.Junior
	}
	if src.Senior != 0 {
		dst.Senior = src.Senior
	}
	if src.Staff != 0 {
		dst.Staff = src.Staff
	}
	if src.Principal != 0 {
		dst.Principal = src.Principal
	}
	if src.SeniorPrincipal != 0 {
		dst.SeniorPrincipal = src.SeniorPrincipal
	}
	if src.Distinguished != 0 {
		dst.Distinguished = src.Distinguished
	}
}

func mergeTeamCompositionBySize(dst, src *TeamCompositionBySize) {
	// merge each size tier if it has any non-zero values
	if hasNonZeroComposition(&src.Small) {
		mergeTeamComposition(&dst.Small, &src.Small)
	}
	if hasNonZeroComposition(&src.Medium) {
		mergeTeamComposition(&dst.Medium, &src.Medium)
	}
	if hasNonZeroComposition(&src.Large) {
		mergeTeamComposition(&dst.Large, &src.Large)
	}
	if hasNonZeroComposition(&src.Enterprise) {
		mergeTeamComposition(&dst.Enterprise, &src.Enterprise)
	}
	if hasNonZeroComposition(&src.Mega) {
		mergeTeamComposition(&dst.Mega, &src.Mega)
	}
}

func hasNonZeroComposition(c *TeamCompositionConfig) bool {
	return c.Junior != 0 || c.Senior != 0 || c.Staff != 0 ||
		c.Principal != 0 || c.SeniorPrincipal != 0 || c.Distinguished != 0
}

func mergeAILeverageBySkill(dst, src *AILeverageBySkill) {
	if src.Junior != 0 {
		dst.Junior = src.Junior
	}
	if src.Senior != 0 {
		dst.Senior = src.Senior
	}
	if src.Staff != 0 {
		dst.Staff = src.Staff
	}
	if src.Principal != 0 {
		dst.Principal = src.Principal
	}
	if src.SeniorPrincipal != 0 {
		dst.SeniorPrincipal = src.SeniorPrincipal
	}
	if src.Distinguished != 0 {
		dst.Distinguished = src.Distinguished
	}
}

// GetTeamCompositionForSize returns the appropriate team composition for a given team size.
// If TeamCompositionBySize is configured, uses size-based lookup; otherwise uses default TeamComposition.
func GetTeamCompositionForSize(teamSize float64) TeamCompositionConfig {
	cfg := GetModelConfig()

	// If no size-based config, return default composition
	if cfg.TeamCompositionBySize == nil {
		return cfg.TeamComposition
	}

	bySize := cfg.TeamCompositionBySize
	switch {
	case teamSize >= 2000:
		return bySize.Mega
	case teamSize >= 1000:
		return bySize.Enterprise
	case teamSize >= 200:
		return bySize.Large
	case teamSize >= 50:
		return bySize.Medium
	default:
		return bySize.Small
	}
}

// BlendedMonthlyCost calculates the weighted average monthly cost based on team composition.
// Returns (low, high) cost range.
func BlendedMonthlyCost(composition TeamCompositionConfig) (low, high float64) {
	cfg := GetModelConfig()

	// helper to get monthly cost for a band
	bandCost := func(key string) (float64, float64) {
		band, ok := cfg.SkillBands[key]
		if !ok {
			return 0, 0
		}
		return band.AnnualCostLow / 12, band.AnnualCostHigh / 12
	}

	// weight each band by its ratio
	weights := map[string]float64{
		BandKeyJunior:          composition.Junior,
		BandKeySenior:          composition.Senior,
		BandKeyStaff:           composition.Staff,
		BandKeyPrincipal:       composition.Principal,
		BandKeySeniorPrincipal: composition.SeniorPrincipal,
		BandKeyDistinguished:   composition.Distinguished,
	}

	for key, ratio := range weights {
		if ratio > 0 {
			bandLow, bandHigh := bandCost(key)
			low += ratio * bandLow
			high += ratio * bandHigh
		}
	}

	return low, high
}

// GetAILeverage returns the AI leverage multiplier for a skill band key
func GetAILeverage(bandKey string) float64 {
	cfg := GetModelConfig()
	switch bandKey {
	case BandKeyJunior:
		return cfg.AILeverageBySkill.Junior
	case BandKeySenior:
		return cfg.AILeverageBySkill.Senior
	case BandKeyStaff:
		return cfg.AILeverageBySkill.Staff
	case BandKeyPrincipal:
		return cfg.AILeverageBySkill.Principal
	case BandKeySeniorPrincipal:
		return cfg.AILeverageBySkill.SeniorPrincipal
	case BandKeyDistinguished:
		return cfg.AILeverageBySkill.Distinguished
	default:
		return 1.0
	}
}

// EffectiveTeamCapacity calculates the AI-leveraged effective capacity of a team.
// Returns the equivalent number of "conventional engineer units" the team can produce.
//
// Example: A team of 2 principals (leverage 3.5×) has effective capacity of 7.0,
// meaning they can produce output equivalent to 7 conventional engineers.
func EffectiveTeamCapacity(headcount float64, composition TeamCompositionConfig) float64 {
	cfg := GetModelConfig()
	leverage := cfg.AILeverageBySkill

	// weighted sum: headcount × ratio × leverage for each band
	effective := headcount * (composition.Junior*leverage.Junior +
		composition.Senior*leverage.Senior +
		composition.Staff*leverage.Staff +
		composition.Principal*leverage.Principal +
		composition.SeniorPrincipal*leverage.SeniorPrincipal +
		composition.Distinguished*leverage.Distinguished)

	return effective
}

// BlendedAILeverage calculates the weighted average AI leverage for a team composition.
// Useful for understanding the overall AI effectiveness of a team mix.
func BlendedAILeverage(composition TeamCompositionConfig) float64 {
	cfg := GetModelConfig()
	leverage := cfg.AILeverageBySkill

	return composition.Junior*leverage.Junior +
		composition.Senior*leverage.Senior +
		composition.Staff*leverage.Staff +
		composition.Principal*leverage.Principal +
		composition.SeniorPrincipal*leverage.SeniorPrincipal +
		composition.Distinguished*leverage.Distinguished
}

// DefaultTeamCompositionBySize returns the default size-based composition config
func DefaultTeamCompositionBySize() *TeamCompositionBySize {
	return &TeamCompositionBySize{
		// Small teams (5-20): mostly senior, some staff, few juniors
		Small: TeamCompositionConfig{
			Junior: 0.10, Senior: 0.60, Staff: 0.25, Principal: 0.05,
		},
		// Medium teams (50-100): pyramid emerges
		Medium: TeamCompositionConfig{
			Junior: 0.15, Senior: 0.50, Staff: 0.25, Principal: 0.08, SeniorPrincipal: 0.02,
		},
		// Large teams (200-500): full IC ladder
		Large: TeamCompositionConfig{
			Junior: 0.20, Senior: 0.45, Staff: 0.22, Principal: 0.10, SeniorPrincipal: 0.03,
		},
		// Enterprise (1000+): more juniors, SPE common
		Enterprise: TeamCompositionConfig{
			Junior: 0.25, Senior: 0.40, Staff: 0.20, Principal: 0.10, SeniorPrincipal: 0.04, Distinguished: 0.01,
		},
		// Mega (2000+): full ladder with Fellows
		Mega: TeamCompositionConfig{
			Junior: 0.25, Senior: 0.40, Staff: 0.18, Principal: 0.10, SeniorPrincipal: 0.05, Distinguished: 0.02,
		},
	}
}

// package-level active config with thread-safe access
var (
	activeConfig     *ModelConfig
	activeConfigOnce sync.Once
	activeConfigMu   sync.RWMutex
)

// SetModelConfig sets the active model configuration
func SetModelConfig(cfg *ModelConfig) {
	activeConfigMu.Lock()
	defer activeConfigMu.Unlock()
	activeConfig = cfg
}

// GetModelConfig returns the active model configuration
// Returns DefaultModelConfig() if none has been set
func GetModelConfig() *ModelConfig {
	activeConfigMu.RLock()
	cfg := activeConfig
	activeConfigMu.RUnlock()

	if cfg != nil {
		return cfg
	}

	activeConfigOnce.Do(func() {
		activeConfigMu.Lock()
		if activeConfig == nil {
			activeConfig = DefaultModelConfig()
		}
		activeConfigMu.Unlock()
	})

	activeConfigMu.RLock()
	defer activeConfigMu.RUnlock()
	return activeConfig
}

// ResetModelConfig resets to default configuration (primarily for testing)
func ResetModelConfig() {
	activeConfigMu.Lock()
	defer activeConfigMu.Unlock()
	activeConfig = nil
	activeConfigOnce = sync.Once{}
}
