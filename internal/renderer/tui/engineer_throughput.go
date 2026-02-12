package tui

import (
	"fmt"
	"strings"

	"github.com/modern-tooling/aloc/internal/git"
	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
)

const (
	barWidth        = 40    // fixed bar width in characters
	maxMultiplier   = 50.0  // hyper benchmark (display cap for AI-augmented engineers)
	baseMultiplier  = 1.0   // baseline benchmark (80 LOC/day)
	tenXMultiplier  = 10.0  // traditional "10x engineer" benchmark
	fillChar        = '█'   // U+2588 full block
	emptyChar       = ' '   // space
	tenXMarkerChar  = '|'   // vertical line for 10x marker
)

// RenderEngineerThroughput renders the engineer throughput analysis table
func RenderEngineerThroughput(metrics *model.EngineerMetrics, theme *renderer.Theme) string {
	if metrics == nil || len(metrics.Engineers) == 0 {
		return ""
	}

	var sb strings.Builder

	// header
	sb.WriteString("\n")
	sb.WriteString(theme.PrimaryBold.Render(fmt.Sprintf("Engineer Throughput (%d months, core+test LOC)", metrics.PeriodMonths)))
	sb.WriteString("\n")
	sb.WriteString(theme.Dim.Render(strings.Repeat("─", 80)))
	sb.WriteString("\n")

	// scale header with benchmark markers
	renderScaleHeader(&sb, theme)

	// render each engineer
	for _, eng := range metrics.Engineers {
		renderEngineerRow(&sb, eng, theme)
	}

	// footer separator
	sb.WriteString(theme.Dim.Render(strings.Repeat("─", 80)))
	sb.WriteString("\n")

	// caveat and benchmark note (always shown)
	sb.WriteString(theme.Dim.Render("⚠ " + metrics.Caveat))
	sb.WriteString("\n")
	sb.WriteString(theme.Dim.Render("| marks the legendary pre-2026 10x engineer (800 LOC/day)"))
	sb.WriteString("\n")

	return sb.String()
}

// renderScaleHeader renders the scale header with 1x, 10x, and 50x markers
func renderScaleHeader(sb *strings.Builder, theme *renderer.Theme) {
	// leftPad positions the scale to align with bar start (after name/loc/ai columns)
	// 2 (indent) + nameWidth + 1 + locWidth + 1 + aiWidth + 2 + 1 (bracket) = aligned with bar
	leftPad := 2 + nameWidth + 1 + locWidth + 1 + aiWidth + 2 + 1

	// calculate marker positions
	baselineRatio := baseMultiplier / maxMultiplier
	baselinePos := max(1, int(float64(barWidth)*baselineRatio))
	tenXRatio := tenXMultiplier / maxMultiplier
	tenXPos := int(float64(barWidth) * tenXRatio)

	// build the scale line with markers
	scaleBar := make([]rune, barWidth)
	for i := range scaleBar {
		scaleBar[i] = ' '
	}

	// place 1x· marker
	if baselinePos < barWidth-2 {
		copy(scaleBar[baselinePos:], []rune("1x"))
	}

	// place 10x marker (the "mythical 10x engineer")
	if tenXPos > 3 && tenXPos < barWidth-4 {
		copy(scaleBar[tenXPos-2:], []rune("10x"))
	}

	// place ·50x marker (right-aligned at end)
	if barWidth >= 4 {
		copy(scaleBar[barWidth-3:], []rune("50x"))
	}

	scaleLine := strings.Repeat(" ", leftPad) + string(scaleBar)
	sb.WriteString(theme.Dim.Render(scaleLine))
	sb.WriteString("\n")
}

// column widths for alignment
const (
	nameWidth = 18 // max name width
	locWidth  = 6  // LOC column width (right-aligned)
	aiWidth   = 4  // AI% column width (right-aligned)
	multWidth = 4  // multiplier column width (right-aligned)
)

// renderEngineerRow renders a single engineer row with bar visualization
func renderEngineerRow(sb *strings.Builder, eng model.EngineerStat, theme *renderer.Theme) {
	// prefer mailmap-resolved name, fall back to email prefix
	name := git.DisplayName(eng.AuthorName, eng.AuthorEmail)
	if len(name) > nameWidth {
		name = name[:nameWidth]
	}
	// pad name BEFORE styling (ANSI codes break width calculation)
	paddedName := fmt.Sprintf("%-*s", nameWidth, name)

	// format LOC (right-aligned)
	locStr := fmt.Sprintf("%*s", locWidth, formatLOCCompact(eng.TotalLOC))

	// format AI percentage (right-aligned)
	aiPct := int(eng.AIPercent * 100)
	aiStr := fmt.Sprintf("%*d%%", aiWidth-1, aiPct)

	// build the bar
	bar := buildMultiplierBar(eng.Multiplier)

	// format multiplier display (right-aligned)
	multStr := fmt.Sprintf("%*s", multWidth, formatMultiplier(eng.Multiplier))

	// render the row with pre-padded fields
	fmt.Fprintf(sb, "  %s %s %s  %s %s\n",
		theme.Primary.Render(paddedName),
		theme.Secondary.Render(locStr),
		theme.Dim.Render(aiStr),
		theme.Primary.Render(bar),
		theme.PrimaryBold.Render(multStr),
	)
}

// formatLOCCompact formats LOC in compact form (e.g., 302.6k, 7.4k, 326)
func formatLOCCompact(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

// buildMultiplierBar creates a horizontal bar visualization with 10x benchmark marker
func buildMultiplierBar(multiplier float64) string {
	// clamp multiplier to max for display
	displayMult := multiplier
	if displayMult > maxMultiplier {
		displayMult = maxMultiplier
	}

	// calculate fill width
	fillWidth := int((displayMult / maxMultiplier) * float64(barWidth))
	if fillWidth < 1 && multiplier >= 1.0 {
		fillWidth = 1 // minimum 1 char for 1x+
	}
	if fillWidth > barWidth {
		fillWidth = barWidth
	}

	// build bar with markers
	var bar strings.Builder
	bar.WriteRune('[')

	// calculate marker positions
	baselineRatio := baseMultiplier / maxMultiplier
	baselinePos := max(1, int(float64(barWidth)*baselineRatio))
	tenXRatio := tenXMultiplier / maxMultiplier
	tenXPos := int(float64(barWidth) * tenXRatio)

	for i := range barWidth {
		if i < fillWidth {
			bar.WriteRune(fillChar)
		} else if i == tenXPos && multiplier < tenXMultiplier {
			// show 10x marker as vertical line for engineers below 10x
			bar.WriteRune(tenXMarkerChar)
		} else if i == baselinePos && fillWidth <= baselinePos {
			bar.WriteRune('·') // baseline marker visible only when not covered
		} else {
			bar.WriteRune(emptyChar)
		}
	}

	bar.WriteRune(']')
	return bar.String()
}

// formatMultiplier formats the multiplier value for display
func formatMultiplier(multiplier float64) string {
	if multiplier >= maxMultiplier {
		return "50x+"
	}
	if multiplier >= 10.0 {
		return fmt.Sprintf("%.0fx", multiplier)
	}
	return fmt.Sprintf("%.1fx", multiplier)
}
