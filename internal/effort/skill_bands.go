package effort

// SkillBand represents an engineer skill level with associated costs
type SkillBand struct {
	Name           string  // e.g., "Principal Engineer"
	AnnualCostLow  float64 // lower bound of annual fully-loaded cost
	AnnualCostHigh float64 // upper bound of annual fully-loaded cost
	Description    string  // brief description
}

// SkillBandCost represents a cost estimate at a specific skill band
type SkillBandCost struct {
	Band     SkillBand
	CostLow  float64 // total cost at lower bound
	CostHigh float64 // total cost at upper bound
	Months   float64 // schedule months
}

// Standard skill bands (2025 fully-loaded costs, US market)
var (
	BandStaff = SkillBand{
		Name:           "Staff Engineer",
		AnnualCostLow:  300000,
		AnnualCostHigh: 350000,
		Description:    "market average IC",
	}

	BandPrincipal = SkillBand{
		Name:           "Principal Engineer",
		AnnualCostLow:  900000,
		AnnualCostHigh: 1100000,
		Description:    "strong IC, system-level thinker",
	}

	BandSeniorPrincipal = SkillBand{
		Name:           "Senior Principal Engineer",
		AnnualCostLow:  1500000,
		AnnualCostHigh: 2000000,
		Description:    "rare, org-shaping capability",
	}
)

// AllBands returns all standard skill bands in order of seniority
func AllBands() []SkillBand {
	return []SkillBand{BandStaff, BandPrincipal, BandSeniorPrincipal}
}

// MonthlyRate returns the monthly cost range for a skill band
func (b SkillBand) MonthlyRate() (low, high float64) {
	return b.AnnualCostLow / 12, b.AnnualCostHigh / 12
}

// CostForMonths calculates total cost for a given number of months
func (b SkillBand) CostForMonths(months float64) (low, high float64) {
	monthlyLow, monthlyHigh := b.MonthlyRate()
	return monthlyLow * months, monthlyHigh * months
}

// CalculateSkillBandCosts computes costs across all skill bands for given effort
func CalculateSkillBandCosts(scheduleMonths float64) []SkillBandCost {
	bands := AllBands()
	result := make([]SkillBandCost, len(bands))

	for i, band := range bands {
		low, high := band.CostForMonths(scheduleMonths)
		result[i] = SkillBandCost{
			Band:     band,
			CostLow:  low,
			CostHigh: high,
			Months:   scheduleMonths,
		}
	}

	return result
}

// EliteOperatorReference represents an observed best-case scenario
type EliteOperatorReference struct {
	Months          float64 // observed schedule
	AISpend         float64 // observed AI tooling cost
	CostRangeLow    float64 // Principal-level estimate
	CostRangeHigh   float64 // SPE-level estimate
	HybridCostLow   float64 // Principal + AI
	HybridCostHigh  float64 // SPE + AI
	VsMarketLow     float64 // reduction factor vs market cost (Principal)
	VsMarketHigh    float64 // reduction factor vs market cost (SPE)
	Description     string  // framing text
}

// CalculateEliteOperatorReference computes the elite operator reference
// based on observed data: 1.5 months, ~$6K AI spend
func CalculateEliteOperatorReference(marketCost float64) EliteOperatorReference {
	const (
		observedMonths  = 1.5
		observedAISpend = 6000
	)

	principalLow, principalHigh := BandPrincipal.CostForMonths(observedMonths)
	speLow, speHigh := BandSeniorPrincipal.CostForMonths(observedMonths)

	// Use midpoints for hybrid calculations
	principalMid := (principalLow + principalHigh) / 2
	speMid := (speLow + speHigh) / 2

	hybridPrincipal := principalMid + observedAISpend
	hybridSPE := speMid + observedAISpend

	return EliteOperatorReference{
		Months:         observedMonths,
		AISpend:        observedAISpend,
		CostRangeLow:   principalMid,
		CostRangeHigh:  speMid,
		HybridCostLow:  hybridPrincipal,
		HybridCostHigh: hybridSPE,
		VsMarketLow:    marketCost / hybridPrincipal,
		VsMarketHigh:   marketCost / hybridSPE,
		Description:    "best-case, skill-compressed scenario",
	}
}
