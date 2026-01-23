package git

// CalculateParallelismSignal determines if multiple authors work concurrently
func CalculateParallelismSignal(events []ChangeEvent) string {
	// group commits by week and count unique authors
	weekAuthors := make(map[string]map[string]bool)

	for _, ev := range events {
		week := ev.When.Format("2006-W02")
		if weekAuthors[week] == nil {
			weekAuthors[week] = make(map[string]bool)
		}
		weekAuthors[week][ev.Author] = true
	}

	if len(weekAuthors) == 0 {
		return "low"
	}

	// count weeks with multiple authors
	multiAuthorWeeks := 0
	for _, authors := range weekAuthors {
		if len(authors) > 1 {
			multiAuthorWeeks++
		}
	}

	ratio := float64(multiAuthorWeeks) / float64(len(weekAuthors))

	switch {
	case ratio < 0.20:
		return "low"
	case ratio < 0.50:
		return "moderate"
	default:
		return "high"
	}
}
