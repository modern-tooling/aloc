package git

import (
	"github.com/modern-tooling/aloc/internal/model"
)

// CalculateOwnershipConcentration computes % of prod LOC owned by single author
func CalculateOwnershipConcentration(events []ChangeEvent, fileLOC map[string]int) float64 {
	// count changes per file per author (already hashed)
	type fileAuthor struct {
		path   string
		author string
	}
	authorChurn := make(map[fileAuthor]int)
	fileTotal := make(map[string]int)

	for _, ev := range events {
		if ev.Role != model.RoleCore {
			continue // only prod code
		}
		key := fileAuthor{ev.Path, ev.Author}
		churn := ev.Added + ev.Deleted
		authorChurn[key] += churn
		fileTotal[ev.Path] += churn
	}

	// for each file, check if single author dominates
	concentratedLOC := 0
	totalProdLOC := 0

	for path, loc := range fileLOC {
		ft := fileTotal[path]
		if ft == 0 {
			continue
		}

		// get all authors for this file
		authorTotals := make(map[string]int)
		for key, churn := range authorChurn {
			if key.path == path {
				authorTotals[key.author] += churn
			}
		}

		// find top author
		maxAuthor := 0
		for _, c := range authorTotals {
			if c > maxAuthor {
				maxAuthor = c
			}
		}

		totalProdLOC += loc
		if float64(maxAuthor)/float64(ft) > 0.50 {
			concentratedLOC += loc
		}
	}

	if totalProdLOC == 0 {
		return 0
	}

	return float64(concentratedLOC) / float64(totalProdLOC)
}
