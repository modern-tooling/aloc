package git

import (
	"strings"
	"time"

	"github.com/modern-tooling/aloc/internal/model"
)

// Glyphs for sparkline rendering (8 levels)
var glyphs = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// churnToGlyph maps a normalized value (0-1) to a sparkline glyph
// Uses perception-tuned thresholds (non-linear)
func churnToGlyph(v float64) rune {
	switch {
	case v < 0.02:
		return glyphs[0]
	case v < 0.08:
		return glyphs[1]
	case v < 0.18:
		return glyphs[2]
	case v < 0.32:
		return glyphs[3]
	case v < 0.50:
		return glyphs[4]
	case v < 0.70:
		return glyphs[5]
	case v < 0.88:
		return glyphs[6]
	default:
		return glyphs[7]
	}
}

// BuildSparklineString renders buckets as a sparkline string
func BuildSparklineString(buckets []Bucket) string {
	normalized := NormalizeBuckets(buckets)

	var sb strings.Builder
	for _, v := range normalized {
		sb.WriteRune(churnToGlyph(v))
	}

	return sb.String()
}

// ChurnSparkline computes a complete sparkline for a role
// Returns both raw values (for adaptive rendering) and pre-rendered glyphs
// Uses daily resolution for maximum flexibility in adaptive rendering
func ChurnSparkline(events []ChangeEvent, role model.Role, now time.Time, months int, smooth bool) *Sparkline {
	// always use daily resolution for raw values (enables adaptive rendering)
	dailyBuckets := BuildDailyBuckets(now, months)
	AssignChurn(dailyBuckets, events, role)

	// extract raw daily values for adaptive rendering
	values := make([]int, len(dailyBuckets))
	for i, b := range dailyBuckets {
		values[i] = b.Churn
	}

	// for pre-rendered glyphs (JSON output), use weekly resolution
	var displayBuckets []Bucket
	if smooth {
		displayBuckets = BuildBiweeklyBuckets(now, months)
	} else {
		displayBuckets = BuildWeeklyBuckets(now, months)
	}
	AssignChurn(displayBuckets, events, role)

	return &Sparkline{
		Role:    role,
		Buckets: displayBuckets,
		Glyphs:  BuildSparklineString(displayBuckets),
		Values:  values,
	}
}

// BuildChurnSeries creates sparklines for all major roles
func BuildChurnSeries(events []ChangeEvent, now time.Time, months int, smooth bool) map[model.Role]*Sparkline {
	series := make(map[model.Role]*Sparkline)

	// core roles to track
	roles := []model.Role{
		model.RoleCore,
		model.RoleTest,
		model.RoleInfra,
	}

	for _, role := range roles {
		series[role] = ChurnSparkline(events, role, now, months, smooth)
	}

	return series
}

// DownsampleMax reduces values to target count using max pooling
// Max pooling preserves spikes/volatility (critical for insight)
func DownsampleMax(values []int, target int) []int {
	if len(values) <= target {
		return values
	}

	factor := float64(len(values)) / float64(target)
	result := make([]int, target)

	for i := range target {
		start := int(float64(i) * factor)
		end := int(float64(i+1) * factor)
		if end > len(values) {
			end = len(values)
		}

		maxVal := 0
		for j := start; j < end; j++ {
			if values[j] > maxVal {
				maxVal = values[j]
			}
		}
		result[i] = maxVal
	}

	return result
}

// ValuesToGlyphs converts raw churn values to sparkline glyphs
func ValuesToGlyphs(values []int) string {
	if len(values) == 0 {
		return ""
	}

	// find max for normalization
	maxVal := 0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	var sb strings.Builder
	for _, v := range values {
		normalized := 0.0
		if maxVal > 0 {
			normalized = float64(v) / float64(maxVal)
		}
		sb.WriteRune(churnToGlyph(normalized))
	}

	return sb.String()
}

// RenderAdaptiveSparkline renders a sparkline at the target width
func RenderAdaptiveSparkline(values []int, targetWidth int) string {
	if len(values) == 0 {
		return strings.Repeat(string(glyphs[0]), targetWidth)
	}

	// downsample to target width
	downsampled := DownsampleMax(values, targetWidth)

	// convert to glyphs
	return ValuesToGlyphs(downsampled)
}
