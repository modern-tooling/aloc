package effort

import (
	"math"

	"github.com/modern-tooling/aloc/internal/model"
)

// COCOMOOptions contains parameters for COCOMO estimation
type COCOMOOptions struct {
	// CostPerMonth is the average monthly cost per engineer (default: 15000)
	CostPerMonth float64
	// Model type: "organic", "semi-detached", "embedded" (default: organic)
	Model string
}

// DefaultCOCOMOOptions returns default COCOMO parameters
func DefaultCOCOMOOptions() COCOMOOptions {
	return COCOMOOptions{
		CostPerMonth: 15000,
		Model:        "organic",
	}
}

// HumanEffortEstimate contains COCOMO calculation results
type HumanEffortEstimate struct {
	EffortPersonMonths float64 `json:"effort_person_months"`
	ScheduleMonths     float64 `json:"schedule_months"`
	TeamSize           float64 `json:"team_size"`
	EstimatedCost      float64 `json:"estimated_cost"`
	Model              string  `json:"model"`
	KLOC               float64 `json:"kloc"`
}

// getCocomoCoeffs returns COCOMO coefficients for the specified model type
func getCocomoCoeffs(modelType string) (a, b, c, d float64) {
	cfg := GetModelConfig()
	coeffs, ok := cfg.COCOMOModels[modelType]
	if !ok {
		coeffs = cfg.COCOMOModels["organic"]
	}
	return coeffs.A, coeffs.B, coeffs.C, coeffs.D
}

// CalculateHumanEffort estimates development effort using COCOMO Basic model
// COCOMO Basic Organic:
//
//	Effort (PM) = a * (KLOC)^b
//	Time (months) = c * (Effort)^d
//	Team = Effort / Time
//	Cost = Effort * cost_per_month
func CalculateHumanEffort(loc int, opts COCOMOOptions) HumanEffortEstimate {
	if opts.Model == "" {
		opts.Model = "organic"
	}

	cfg := GetModelConfig()
	// validate model exists in config
	if _, ok := cfg.COCOMOModels[opts.Model]; !ok {
		opts.Model = "organic"
	}

	a, b, c, d := getCocomoCoeffs(opts.Model)

	kloc := float64(loc) / 1000.0
	if kloc < 0.001 {
		return HumanEffortEstimate{Model: "COCOMO Basic " + opts.Model, KLOC: kloc}
	}

	// Effort in person-months
	effortPM := a * math.Pow(kloc, b)

	// Schedule in months
	schedule := c * math.Pow(effortPM, d)

	// Team size
	team := effortPM / schedule

	// Determine cost: use provided value or blended cost based on team composition
	var costPerMonth float64
	if opts.CostPerMonth > 0 {
		costPerMonth = opts.CostPerMonth
	} else {
		// use blended cost based on team size
		composition := GetTeamCompositionForSize(team)
		low, high := BlendedMonthlyCost(composition)
		costPerMonth = (low + high) / 2 // use midpoint for single-value estimate
	}

	// Cost
	cost := effortPM * costPerMonth

	return HumanEffortEstimate{
		EffortPersonMonths: effortPM,
		ScheduleMonths:     schedule,
		TeamSize:           team,
		EstimatedCost:      cost,
		Model:              "COCOMO Basic " + opts.Model,
		KLOC:               kloc,
	}
}

// CalculateConventionalTeam estimates Market Replacement cost for a conventional team.
// Returns a range (low-high) based on productivity and coordination variance.
// Uses blended costs based on team composition when opts.CostPerMonth is not set.
func CalculateConventionalTeam(loc int, opts COCOMOOptions) *model.TeamEstimate {
	cfg := GetModelConfig()

	kloc := float64(loc) / 1000.0
	if kloc < 0.001 {
		return nil
	}

	// use organic COCOMO as baseline
	a, b, c, d := getCocomoCoeffs("organic")

	// Calculate team sizes first (needed for blended cost lookup)
	effortLow := a * math.Pow(kloc, b) * cfg.VarianceMultipliers.Optimistic
	scheduleLow := c * math.Pow(effortLow, d)
	teamLow := effortLow / scheduleLow

	effortHigh := a * math.Pow(kloc, b) * cfg.VarianceMultipliers.Pessimistic
	scheduleHigh := c * math.Pow(effortHigh, d)
	teamHigh := effortHigh / scheduleHigh

	// Determine cost per month: use provided value or blended cost based on team composition
	var costPerMonthLow, costPerMonthHigh float64
	if opts.CostPerMonth > 0 {
		// explicit cost provided via --human-cost flag
		costPerMonthLow = opts.CostPerMonth
		costPerMonthHigh = opts.CostPerMonth
	} else {
		// use blended cost based on team size and composition
		compositionLow := GetTeamCompositionForSize(teamLow)
		compositionHigh := GetTeamCompositionForSize(teamHigh)
		costPerMonthLow, _ = BlendedMonthlyCost(compositionLow)
		_, costPerMonthHigh = BlendedMonthlyCost(compositionHigh)
	}

	costLow := effortLow * costPerMonthLow
	costHigh := effortHigh * costPerMonthHigh

	return &model.TeamEstimate{
		Cost:       model.EstimateRange{Low: costLow, High: costHigh},
		ScheduleMo: model.EstimateRange{Low: scheduleLow, High: scheduleHigh},
		TeamSize:   model.EstimateRange{Low: teamLow, High: teamHigh},
		Model:      "Conventional Team (COCOMO-based)",
	}
}

// CalculateAgenticTeam estimates effort for an AI-native, agentic delivery model.
// Assumes parallel AI agent execution with human supervision/integration.
//
// Conservative assumptions (defensible to senior engineers):
//   - Humans still do final design & merge
//   - AI output requires review and correction
//   - Task decomposition is imperfect
//   - Not all work parallelizes
//   - Coordination overhead remains, just reduced
//
// What changes: parallelism (via agents), idle time, drafting cost, iteration cycles
// What does NOT: architecture decisions, final correctness, integration, human review
//
// AI leverage by skill is factored in: principal+ engineers get more from AI tools
// because they can orchestrate agents effectively and validate output quickly.
func CalculateAgenticTeam(loc int, opts COCOMOOptions) *model.TeamEstimate {
	cfg := GetModelConfig()

	kloc := float64(loc) / 1000.0
	if kloc < 0.001 {
		return nil
	}

	conv := CalculateConventionalTeam(loc, opts)
	if conv == nil {
		return nil
	}

	ai := cfg.AINative

	// Base team size reduction from AI assistance
	baseTeamLow := conv.TeamSize.Low * ai.TeamSizeFactorLow
	baseTeamHigh := conv.TeamSize.High * ai.TeamSizeFactorHigh
	if baseTeamLow < ai.MinimumTeamSize {
		baseTeamLow = ai.MinimumTeamSize
	}

	// Get team composition for the AI-native team size
	// Smaller AI-native teams tend to be more senior-heavy
	compositionLow := GetTeamCompositionForSize(baseTeamLow)
	compositionHigh := GetTeamCompositionForSize(baseTeamHigh)

	// Calculate AI leverage multiplier based on team composition
	// Senior-heavy teams get more leverage from AI tools
	leverageLow := BlendedAILeverage(compositionLow)
	leverageHigh := BlendedAILeverage(compositionHigh)

	// Effective team output = actual headcount × AI leverage
	// This means a team of 3 principals (leverage 3.5×) outputs like 10.5 conventional engineers
	effectiveCapacityLow := baseTeamLow * leverageLow
	effectiveCapacityHigh := baseTeamHigh * leverageHigh

	// Schedule compression is enhanced by AI leverage
	// Higher leverage = faster completion because each person is more effective
	// Base factor adjusted by sqrt of leverage (diminishing returns on parallelization)
	scheduleLow := conv.ScheduleMo.Low * ai.ScheduleFactorLow / math.Sqrt(leverageLow)
	scheduleHigh := conv.ScheduleMo.High * ai.ScheduleFactorHigh / math.Sqrt(leverageHigh)

	// Determine cost per month: use provided value or blended cost
	var costPerMonthLow, costPerMonthHigh float64
	if opts.CostPerMonth > 0 {
		costPerMonthLow = opts.CostPerMonth
		costPerMonthHigh = opts.CostPerMonth
	} else {
		costPerMonthLow, _ = BlendedMonthlyCost(compositionLow)
		_, costPerMonthHigh = BlendedMonthlyCost(compositionHigh)
	}

	// Cost = actual headcount × schedule × blended monthly cost + AI tooling
	humanCostLow := baseTeamLow * scheduleLow * costPerMonthLow
	humanCostHigh := baseTeamHigh * scheduleHigh * costPerMonthHigh

	// AI tooling: configurable, higher for large codebases
	aiToolingLow := ai.ToolingMonthlyLow
	aiToolingHigh := ai.ToolingMonthlyHigh
	if kloc > ai.LargeCodebaseThresholdK {
		aiToolingLow = ai.ToolingMonthlyLowLarge
		aiToolingHigh = ai.ToolingMonthlyHighLarge
	}

	// Total cost = human labor + AI tooling over schedule duration
	costLow := humanCostLow + (aiToolingLow * scheduleLow)
	costHigh := humanCostHigh + (aiToolingHigh * scheduleHigh)

	return &model.TeamEstimate{
		Cost:              model.EstimateRange{Low: costLow, High: costHigh},
		ScheduleMo:        model.EstimateRange{Low: scheduleLow, High: scheduleHigh},
		TeamSize:          model.EstimateRange{Low: baseTeamLow, High: baseTeamHigh},
		EffectiveCapacity: model.EstimateRange{Low: effectiveCapacityLow, High: effectiveCapacityHigh},
		AIToolingMo:       model.EstimateRange{Low: aiToolingLow, High: aiToolingHigh},
		AILeverage:        model.EstimateRange{Low: leverageLow, High: leverageHigh},
		Model:             "AI-Native Team (Agentic/Parallel)",
	}
}
