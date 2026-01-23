package git

import "time"

// CalculateStability computes stable core and volatile surface percentages
func CalculateStability(events []ChangeEvent, fileLOC map[string]int, stableMonths int) (stableCore, volatileSurface float64) {
	now := time.Now()
	stableCutoff := now.AddDate(0, -stableMonths, 0)
	volatileCutoff := now.AddDate(0, -6, 0)

	lastModified := make(map[string]time.Time)
	changeCount := make(map[string]int) // changes in last 6 months

	for _, ev := range events {
		if ev.When.After(lastModified[ev.Path]) {
			lastModified[ev.Path] = ev.When
		}
		if ev.When.After(volatileCutoff) {
			changeCount[ev.Path]++
		}
	}

	totalLOC := 0
	stableLOC := 0
	volatileLOC := 0

	for path, loc := range fileLOC {
		totalLOC += loc

		modified, exists := lastModified[path]
		if !exists || modified.Before(stableCutoff) {
			stableLOC += loc
		}

		if changeCount[path] >= 5 {
			volatileLOC += loc
		}
	}

	if totalLOC == 0 {
		return 0, 0
	}

	return float64(stableLOC) / float64(totalLOC),
		float64(volatileLOC) / float64(totalLOC)
}

// CalculateRewritePressure computes delete/total ratio as indicator of rewrites
func CalculateRewritePressure(events []ChangeEvent) float64 {
	totalAdded := 0
	totalDeleted := 0

	for _, ev := range events {
		totalAdded += ev.Added
		totalDeleted += ev.Deleted
	}

	total := totalAdded + totalDeleted
	if total == 0 {
		return 0
	}

	return float64(totalDeleted) / float64(total)
}
