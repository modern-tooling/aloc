package git

import "github.com/modern-tooling/aloc/internal/model"

// CalculateEffortAdjustments determines effort multipliers based on git signals
func CalculateEffortAdjustments(
	churnStat ChurnStat,
	stableCore, volatileSurface, rewritePressure, ownershipConc float64,
	churnSeries map[model.Role]*Sparkline,
) ([]EffortAdjustment, float64) {
	var adjustments []EffortAdjustment

	// high churn concentration (hotspots)
	if churnStat.FilePercent < 15 && churnStat.EditPercent > 60 {
		adjustments = append(adjustments, EffortAdjustment{
			Reason:     "High churn concentration",
			Adjustment: 0.10,
		})
	}

	// sustained prod churn
	if prodSparkline, ok := churnSeries[model.RoleCore]; ok {
		if sustainedHighChurn(prodSparkline.Buckets) {
			adjustments = append(adjustments, EffortAdjustment{
				Reason:     "Sustained prod churn",
				Adjustment: 0.12,
			})
		}
	}

	// late infra volatility
	if infraSparkline, ok := churnSeries[model.RoleInfra]; ok {
		if lateVolatility(infraSparkline.Buckets) {
			adjustments = append(adjustments, EffortAdjustment{
				Reason:     "Late infra volatility",
				Adjustment: 0.08,
			})
		}
	}

	// ownership risk
	if ownershipConc > 0.30 {
		adjustments = append(adjustments, EffortAdjustment{
			Reason:     "Ownership concentration",
			Adjustment: 0.10,
		})
	}

	// rewrite pressure
	if rewritePressure > 0.45 {
		adjustments = append(adjustments, EffortAdjustment{
			Reason:     "Rewrite-heavy segments",
			Adjustment: 0.06,
		})
	}

	// stable foundation (negative adjustment - reduces effort)
	if stableCore > 0.50 && volatileSurface < 0.10 {
		adjustments = append(adjustments, EffortAdjustment{
			Reason:     "Stable foundation",
			Adjustment: -0.05,
		})
	}

	// calculate net
	net := 0.0
	for _, a := range adjustments {
		net += a.Adjustment
	}

	return adjustments, net
}
