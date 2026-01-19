package model

import "time"

// Report is the complete analysis output
type Report struct {
	Meta             Meta             `json:"meta"`
	Summary          Summary          `json:"summary"`
	Responsibilities []Responsibility `json:"responsibilities"`
	Ratios           Ratios           `json:"ratios"`
	Languages        []LanguageComp   `json:"languages"`
	Trend            *Trend           `json:"trend,omitempty"`
	Confidence       ConfidenceInfo   `json:"confidence"`
	Effort           *EffortEstimates `json:"effort,omitempty"`
	Files            []*FileRecord    `json:"files,omitempty"`
}

// Meta contains metadata about the report
type Meta struct {
	SchemaVersion    string    `json:"schema_version"`
	GeneratedAt      time.Time `json:"generated_at"`
	Generator        string    `json:"generator"`
	GeneratorVersion string    `json:"generator_version"`
	Repo             *RepoInfo `json:"repo,omitempty"`
}

// RepoInfo contains repository metadata
type RepoInfo struct {
	Name   string `json:"name,omitempty"`
	Commit string `json:"commit,omitempty"`
	Branch string `json:"branch,omitempty"`
	Root   string `json:"root,omitempty"`
}

// Summary contains high-level statistics
type Summary struct {
	Files     int         `json:"files"`
	LOCTotal  int         `json:"loc_total"`
	Lines     LineMetrics `json:"lines"`
	Languages int         `json:"languages"`
}

// Responsibility contains LOC breakdown by role
type Responsibility struct {
	Role       Role                 `json:"role"`
	LOC        int                  `json:"loc"`
	Files      int                  `json:"files"`
	Confidence float32              `json:"confidence"`
	Breakdown  map[TestKind]float32 `json:"breakdown,omitempty"`
	Notes      []string             `json:"notes,omitempty"`
}

// Ratios contains pre-calculated key ratios
type Ratios struct {
	TestToCore      float32 `json:"test_to_core"`
	InfraToCore     float32 `json:"infra_to_core"`
	DocsToCore      float32 `json:"docs_to_core"`
	GeneratedToCore float32 `json:"generated_to_core"`
	ConfigToCore    float32 `json:"config_to_core"`
}

// LanguageComp contains language composition data
type LanguageComp struct {
	Language         string                  `json:"language"`
	Category         string                  `json:"category,omitempty"`
	LOCTotal         int                     `json:"loc_total"`
	Files            int                     `json:"files"`
	Code             int                     `json:"code"`
	Comments         int                     `json:"comments"`
	Blanks           int                     `json:"blanks"`
	Tests            int                     `json:"tests"`
	Config           int                     `json:"config"`
	Responsibilities map[Role]int            `json:"responsibilities"`
	Embedded         map[string]LineMetrics  `json:"embedded,omitempty"` // for container languages (Markdown, etc.)
}

// Trend contains historical trend data
type Trend struct {
	Window         string    `json:"window"`
	Sparkline      []float32 `json:"sparkline"`
	Direction      string    `json:"direction"`
	Interpretation string    `json:"interpretation"`
}

// ConfidenceInfo contains classification confidence breakdown
type ConfidenceInfo struct {
	AutoClassified float32 `json:"auto_classified"`
	Heuristic      float32 `json:"heuristic"`
	Override       float32 `json:"override"`
}

// EffortEstimates contains human and AI effort estimates
type EffortEstimates struct {
	Human           *HumanEffort            `json:"human,omitempty"`
	AI              *AIEffort               `json:"ai,omitempty"`
	Comparison      *CostComparison         `json:"comparison,omitempty"`
	Conventional    *TeamEstimate           `json:"conventional,omitempty"`    // Market Replacement (Conventional Team)
	Agentic         *TeamEstimate           `json:"agentic,omitempty"`         // AI-Native Team (Agentic/Parallel)
	HybridBreakdown []HybridSavings         `json:"hybrid_breakdown,omitempty"`
	QuickActions    []QuickAction           `json:"quick_actions,omitempty"`
	EliteReference  *EliteOperatorReference `json:"elite_reference,omitempty"`
}

// EstimateRange represents low/high bounds for an estimate
type EstimateRange struct {
	Low  float64 `json:"low"`
	High float64 `json:"high"`
}

// TeamEstimate represents a delivery model's cost/effort estimate with ranges
type TeamEstimate struct {
	Cost         EstimateRange `json:"cost"`           // dollar range
	ScheduleMo   EstimateRange `json:"schedule_months"` // schedule range in months
	TeamSize     EstimateRange `json:"team_size"`       // team size range
	AIToolingMo  EstimateRange `json:"ai_tooling_monthly,omitempty"` // monthly AI tooling cost (for agentic)
	Model        string        `json:"model"`           // description of delivery model
}

// HumanEffort contains COCOMO-style human effort estimates (legacy, single-point)
type HumanEffort struct {
	EstimatedCost      float64 `json:"estimated_cost"`
	EffortPersonMonths float64 `json:"effort_person_months"`
	ScheduleMonths     float64 `json:"schedule_months"`
	TeamSize           float64 `json:"team_size"`
	Model              string  `json:"model"`
}

// AIEffort contains AI token-based cost estimates
type AIEffort struct {
	InputTokens  int64    `json:"input_tokens"`
	OutputTokens int64    `json:"output_tokens"`
	InputCost    float64  `json:"input_cost"`
	OutputCost   float64  `json:"output_cost"`
	TotalCost    float64  `json:"total_cost"`
	Model        string   `json:"model"`
	Assumptions  []string `json:"assumptions"`
}

// CostComparison contains comparative cost analysis between human and AI development
type CostComparison struct {
	AIOnly        float64 `json:"ai_only"`        // Pure AI implementation cost
	HumanOnly     float64 `json:"human_only"`     // COCOMO human-only estimate
	HybridCost    float64 `json:"hybrid_cost"`    // Human + AI assisted cost
	HybridSavings float64 `json:"hybrid_savings"` // Percentage saved (0.30 = 30%)
	Ratio         float64 `json:"ratio"`          // Human/AI ratio (e.g., 6400)
}

// HybridSavings contains per-role breakdown of AI-assisted savings
type HybridSavings struct {
	Role         Role    `json:"role"`
	Reduction    float64 `json:"reduction"`     // 0.30 = 30% reduction
	DollarsSaved float64 `json:"dollars_saved"` // absolute savings
	Description  string  `json:"description"`   // "AI writes stubs → human reviews"
}

// QuickAction represents an actionable recommendation
type QuickAction struct {
	Priority    int     `json:"priority"`
	Description string  `json:"description"`
	Savings     float64 `json:"savings,omitempty"` // dollar amount if applicable
	LOCGap      int     `json:"loc_gap,omitempty"` // LOC needed to reach target
}

// EliteOperatorReference represents an observed best-case scenario
// where a highly skilled engineer + AI achieved exceptional results
type EliteOperatorReference struct {
	HybridCostLow  float64 `json:"hybrid_cost_low"`  // Principal + AI
	HybridCostHigh float64 `json:"hybrid_cost_high"` // SPE + AI
	VsMarketLow    float64 `json:"vs_market_low"`    // reduction factor (Principal)
	VsMarketHigh   float64 `json:"vs_market_high"`   // reduction factor (SPE)
	Description    string  `json:"description"`      // framing text
}
