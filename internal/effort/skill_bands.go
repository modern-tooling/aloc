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

// Skill band keys
const (
	BandKeyJunior          = "junior"
	BandKeySenior          = "senior"
	BandKeyStaff           = "staff"
	BandKeyPrincipal       = "principal"
	BandKeySeniorPrincipal = "senior_principal"
	BandKeyDistinguished   = "distinguished"
)

// Skill band descriptions and names
var skillBandMeta = map[string]struct {
	name        string
	description string
}{
	BandKeyJunior:          {"Junior Engineer", "L3/SDE1 (0-2 years)"},
	BandKeySenior:          {"Senior Engineer", "L4-5/SDE2-3 (3-7 years)"},
	BandKeyStaff:           {"Staff Engineer", "L6/SDE4 (7+ years)"},
	BandKeyPrincipal:       {"Principal Engineer", "L7, system-level thinker"},
	BandKeySeniorPrincipal: {"Senior Principal Engineer", "L8, org-shaping capability"},
	BandKeyDistinguished:   {"Distinguished Engineer", "L9+/Fellow, rare"},
}

// getSkillBand returns a SkillBand from config for the given key
func getSkillBand(key string) SkillBand {
	cfg := GetModelConfig()
	bandCfg, ok := cfg.SkillBands[key]
	if !ok {
		// fallback to staff if key not found
		bandCfg = cfg.SkillBands[BandKeyStaff]
	}
	meta := skillBandMeta[key]
	return SkillBand{
		Name:           meta.name,
		AnnualCostLow:  bandCfg.AnnualCostLow,
		AnnualCostHigh: bandCfg.AnnualCostHigh,
		Description:    meta.description,
	}
}

// BandJunior returns the Junior Engineer skill band
func BandJunior() SkillBand {
	return getSkillBand(BandKeyJunior)
}

// BandSenior returns the Senior Engineer skill band
func BandSenior() SkillBand {
	return getSkillBand(BandKeySenior)
}

// BandStaff returns the Staff Engineer skill band
func BandStaff() SkillBand {
	return getSkillBand(BandKeyStaff)
}

// BandPrincipal returns the Principal Engineer skill band
func BandPrincipal() SkillBand {
	return getSkillBand(BandKeyPrincipal)
}

// BandSeniorPrincipal returns the Senior Principal Engineer skill band
func BandSeniorPrincipal() SkillBand {
	return getSkillBand(BandKeySeniorPrincipal)
}

// BandDistinguished returns the Distinguished Engineer skill band
func BandDistinguished() SkillBand {
	return getSkillBand(BandKeyDistinguished)
}

// AllBands returns standard skill bands in order of seniority (ascending).
// Excludes Distinguished/Fellow level as most teams don't have that tier.
func AllBands() []SkillBand {
	return []SkillBand{
		BandJunior(),
		BandSenior(),
		BandStaff(),
		BandPrincipal(),
		BandSeniorPrincipal(),
	}
}

// AllBandsIncludingDistinguished returns all skill bands including Distinguished.
// Use for very large organizations that have Fellow-level engineers.
func AllBandsIncludingDistinguished() []SkillBand {
	return []SkillBand{
		BandJunior(),
		BandSenior(),
		BandStaff(),
		BandPrincipal(),
		BandSeniorPrincipal(),
		BandDistinguished(),
	}
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
	Months         float64 // observed schedule
	AISpend        float64 // observed AI tooling cost
	CostRangeLow   float64 // Principal-level estimate
	CostRangeHigh  float64 // SPE-level estimate
	HybridCostLow  float64 // Principal + AI
	HybridCostHigh float64 // SPE + AI
	VsMarketLow    float64 // reduction factor vs market cost (Principal)
	VsMarketHigh   float64 // reduction factor vs market cost (SPE)
	Description    string  // framing text
}

// CalculateEliteOperatorReference computes the elite operator reference
// based on observed data from config
func CalculateEliteOperatorReference(marketCost float64) EliteOperatorReference {
	cfg := GetModelConfig()
	observedMonths := cfg.EliteReference.ObservedMonths
	observedAISpend := cfg.EliteReference.ObservedAISpend

	principal := BandPrincipal()
	spe := BandSeniorPrincipal()

	principalLow, principalHigh := principal.CostForMonths(observedMonths)
	speLow, speHigh := spe.CostForMonths(observedMonths)

	// use midpoints for hybrid calculations
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
