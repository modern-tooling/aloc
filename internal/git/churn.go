package git

import "sort"

// CalculateChurnConcentration computes what % of files account for what % of edits
func CalculateChurnConcentration(events []ChangeEvent) ChurnStat {
	if len(events) == 0 {
		return ChurnStat{100, 100}
	}

	// aggregate churn per file
	fileChurn := make(map[string]int)
	totalChurn := 0
	for _, ev := range events {
		churn := ev.Added + ev.Deleted
		fileChurn[ev.Path] += churn
		totalChurn += churn
	}

	if totalChurn == 0 {
		return ChurnStat{100, 100}
	}

	// sort files by churn descending
	type fc struct {
		path  string
		churn int
	}
	files := make([]fc, 0, len(fileChurn))
	for p, c := range fileChurn {
		files = append(files, fc{p, c})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].churn > files[j].churn
	})

	// find smallest N% accounting for ~65% of edits
	target := float64(totalChurn) * 0.65
	cumulative := 0
	for i, f := range files {
		cumulative += f.churn
		if float64(cumulative) >= target {
			return ChurnStat{
				FilePercent: float64(i+1) / float64(len(files)) * 100,
				EditPercent: float64(cumulative) / float64(totalChurn) * 100,
			}
		}
	}

	return ChurnStat{100, 100}
}
