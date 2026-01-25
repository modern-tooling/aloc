package tui

import (
	"fmt"
	"strings"

	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

// RenderDevelopmentCost renders the effort comparison section with delivery model estimates
// If git metrics are present, shows base estimate, adjustment, and adjusted estimate
func RenderDevelopmentCost(effort *model.EffortEstimates, git *model.GitMetrics, theme *renderer.Theme) string {
	if effort == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(theme.PrimaryBold.Render("Development Effort Models") + "\n")
	b.WriteString(theme.Dim.Render(strings.Repeat("─", 80)) + "\n")

	// Market Replacement Estimate (Conventional Team)
	if effort.Conventional != nil {
		conv := effort.Conventional
		b.WriteString(theme.Secondary.Render("Market Replacement Estimate (Conventional Team)") + "\n")

		if git != nil && git.NetAdjustment != 0 {
			// show base, adjustment, and adjusted estimate
			renderEstimateWithAdjustment(&b, conv, git, theme)
		} else {
			// no git adjustment - single line
			fmt.Fprintf(&b, "  %s – %s · %.0f–%.0f months · ~%.0f–%.0f engineers\n",
				formatCurrencyCompact(conv.Cost.Low),
				formatCurrencyCompact(conv.Cost.High),
				conv.ScheduleMo.Low, conv.ScheduleMo.High,
				conv.TeamSize.Low, conv.TeamSize.High)
		}
		b.WriteString("\n")
	}

	// AI-Native Team Estimate (Agentic/Parallel)
	if effort.Agentic != nil {
		ag := effort.Agentic
		b.WriteString(theme.Secondary.Render("AI-Native Team Estimate (Agentic/Parallel)") + "\n")

		if git != nil && git.NetAdjustment != 0 {
			// show base, adjustment, and adjusted estimate
			renderEstimateWithAdjustment(&b, ag, git, theme)
		} else {
			// no git adjustment - single line
			fmt.Fprintf(&b, "  %s – %s · %.0f–%.0f months · ~%.0f–%.0f engineers\n",
				formatCurrencyCompact(ag.Cost.Low),
				formatCurrencyCompact(ag.Cost.High),
				ag.ScheduleMo.Low, ag.ScheduleMo.High,
				ag.TeamSize.Low, ag.TeamSize.High)
		}

		// AI tooling: monthly rate and total cost on separate line
		if ag.AIToolingMo.High > 0 {
			aiTotalLow := ag.AIToolingMo.Low * ag.ScheduleMo.Low
			aiTotalHigh := ag.AIToolingMo.High * ag.ScheduleMo.High
			fmt.Fprintf(&b, "  %s\n",
				theme.Dim.Render(fmt.Sprintf("(AI tooling: %s–%s/mo, total %s–%s)",
					formatCurrencyCompact(ag.AIToolingMo.Low),
					formatCurrencyCompact(ag.AIToolingMo.High),
					formatCurrencyCompact(aiTotalLow),
					formatCurrencyCompact(aiTotalHigh))))
		}

		// AI leverage: show effective capacity if leverage is significant
		if ag.AILeverage.High > 1.0 && ag.EffectiveCapacity.High > 0 {
			fmt.Fprintf(&b, "  %s\n",
				theme.Dim.Render(fmt.Sprintf("(AI leverage: %.1f–%.1f× · effective capacity: %.0f–%.0f engineers)",
					ag.AILeverage.Low, ag.AILeverage.High,
					ag.EffectiveCapacity.Low, ag.EffectiveCapacity.High)))
		}
	}

	// Disclaimer
	b.WriteString("\n")
	b.WriteString(theme.Dim.Render("* Rough estimates only, +/- depending on the effectiveness and experience of engineers") + "\n")

	return b.String()
}

// renderEstimateWithAdjustment renders base estimate, git adjustment, and adjusted estimate
func renderEstimateWithAdjustment(b *strings.Builder, estimate *model.TeamEstimate, git *model.GitMetrics, theme *renderer.Theme) {
	factor := 1.0 + git.NetAdjustment

	// compute base (unadjusted) values by dividing adjusted by factor
	baseCostLow := estimate.Cost.Low / factor
	baseCostHigh := estimate.Cost.High / factor
	baseSchedLow := estimate.ScheduleMo.Low / factor
	baseSchedHigh := estimate.ScheduleMo.High / factor

	// base estimate
	fmt.Fprintf(b, "  %-22s %s – %s · %.0f–%.0f months\n",
		theme.Dim.Render("Base estimate"),
		formatCurrencyCompact(baseCostLow),
		formatCurrencyCompact(baseCostHigh),
		baseSchedLow, baseSchedHigh)

	// git adjustment with reason(s)
	adjustmentStyle := theme.Warning
	if git.NetAdjustment < 0 {
		adjustmentStyle = theme.Success
	}
	reasons := formatAdjustmentReasons(git)
	fmt.Fprintf(b, "  %-22s %s\n",
		theme.Dim.Render("Git adjustment"),
		adjustmentStyle.Render(fmt.Sprintf("%+.0f%% (%s)", git.NetAdjustment*100, reasons)))

	// adjusted estimate (current values)
	fmt.Fprintf(b, "  %-22s %s – %s · %.0f–%.0f months · ~%.0f–%.0f engineers\n",
		theme.Secondary.Render("Adjusted estimate"),
		formatCurrencyCompact(estimate.Cost.Low),
		formatCurrencyCompact(estimate.Cost.High),
		estimate.ScheduleMo.Low, estimate.ScheduleMo.High,
		estimate.TeamSize.Low, estimate.TeamSize.High)
}

// formatAdjustmentReasons creates a compact string of adjustment reasons
func formatAdjustmentReasons(git *model.GitMetrics) string {
	if len(git.Adjustments) == 0 {
		return "git signals"
	}

	// for single adjustment, use its reason
	if len(git.Adjustments) == 1 {
		return compactReason(git.Adjustments[0].Reason)
	}

	// for multiple, list the top contributors
	var parts []string
	for _, adj := range git.Adjustments {
		parts = append(parts, compactReason(adj.Reason))
	}

	if len(parts) > 2 {
		return fmt.Sprintf("%s +%d more", strings.Join(parts[:2], ", "), len(parts)-2)
	}
	return strings.Join(parts, ", ")
}

// compactReason converts verbose reason to compact form
func compactReason(reason string) string {
	// map verbose reasons to compact forms
	compactMap := map[string]string{
		"High churn concentration":  "churn hotspots",
		"Sustained prod churn":      "sustained churn",
		"Late infra volatility":     "late infra changes",
		"Ownership concentration":   "ownership risk",
		"Rewrite-heavy segments":    "rewrite pressure",
		"Stable foundation":         "stable foundation",
	}

	if compact, ok := compactMap[reason]; ok {
		return compact
	}
	return strings.ToLower(reason)
}
