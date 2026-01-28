package git

import (
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/modern-tooling/aloc/internal/model"
)

const (
	baselineLOCPerDay = 80   // industry average senior engineer
	workdayFactor     = 0.71 // weekends + PTO adjustment (5/7 * 0.9 holidays/pto)
)

// EngineerOptions controls engineer throughput analysis
type EngineerOptions struct {
	PeriodMonths int // analysis window (default 6)
}

// CalculateEngineerStats computes per-contributor throughput metrics
func CalculateEngineerStats(events []ChangeEvent, opts EngineerOptions) *EngineerAnalysis {
	if len(events) == 0 {
		return nil
	}

	periodMonths := opts.PeriodMonths
	if periodMonths <= 0 {
		periodMonths = 6
	}

	// calculate the analysis window
	now := time.Now()
	windowStart := now.AddDate(0, -periodMonths, 0)

	// aggregate by author email (core+test LOC)
	type authorData struct {
		email        string
		loc          int
		aiCommits    int
		totalCommits int
		firstCommit  time.Time              // earliest commit in window
		commitHashes map[string]bool        // track unique commits
		aiHashes     map[string]bool        // track unique AI commits
	}

	byAuthor := make(map[string]*authorData)

	for _, e := range events {
		if e.AuthorEmail == "" {
			continue
		}

		// count core + test role files (production code and tests)
		if e.Role != model.RoleCore && e.Role != model.RoleTest {
			continue
		}

		// only count events within window
		if e.When.Before(windowStart) {
			continue
		}

		ad, ok := byAuthor[e.AuthorEmail]
		if !ok {
			ad = &authorData{
				email:        e.AuthorEmail,
				firstCommit:  e.When,
				commitHashes: make(map[string]bool),
				aiHashes:     make(map[string]bool),
			}
			byAuthor[e.AuthorEmail] = ad
		}

		// track earliest commit for this author
		if e.When.Before(ad.firstCommit) {
			ad.firstCommit = e.When
		}

		ad.loc += e.Added

		// use commit time as a unique-ish identifier within an author
		// this isn't perfect but avoids needing to pass commit hashes
		commitKey := e.When.Format(time.RFC3339)
		if !ad.commitHashes[commitKey] {
			ad.commitHashes[commitKey] = true
			ad.totalCommits++
			if e.AIAssisted && !ad.aiHashes[commitKey] {
				ad.aiHashes[commitKey] = true
				ad.aiCommits++
			}
		}
	}

	if len(byAuthor) == 0 {
		return nil
	}

	// convert to EngineerStats slice
	var engineers []EngineerStats
	for _, ad := range byAuthor {
		// calculate working days based on author's active period
		// use max(windowStart, firstCommit) as the effective start
		effectiveStart := windowStart
		if ad.firstCommit.After(windowStart) {
			effectiveStart = ad.firstCommit
		}
		activeDays := now.Sub(effectiveStart).Hours() / 24
		workingDays := activeDays * workdayFactor
		if workingDays < 1 {
			workingDays = 1 // minimum 1 working day
		}

		locPerDay := float64(ad.loc) / workingDays
		multiplier := locPerDay / float64(baselineLOCPerDay)
		if multiplier < 1.0 {
			multiplier = 1.0 // floor at 1.0x
		}

		aiPercent := 0.0
		if ad.totalCommits > 0 {
			aiPercent = float64(ad.aiCommits) / float64(ad.totalCommits)
		}

		engineers = append(engineers, EngineerStats{
			AuthorEmail: ad.email,
			TotalLOC:    ad.loc,
			LOCPerDay:   locPerDay,
			Multiplier:  multiplier,
			AIPercent:   aiPercent,
			CommitCount: ad.totalCommits,
		})
	}

	// sort by multiplier descending
	sort.Slice(engineers, func(i, j int) bool {
		return engineers[i].Multiplier > engineers[j].Multiplier
	})

	// calculate median multiplier
	medianMult := calculateMedianMultiplier(engineers)

	return &EngineerAnalysis{
		Engineers:    engineers,
		BaselineLOC:  baselineLOCPerDay,
		PeriodMonths: periodMonths,
		MedianMult:   medianMult,
		Caveat:       "Volume metric only - high LOC may indicate bulk changes, not value delivered",
	}
}

// calculateMedianMultiplier finds the median multiplier
func calculateMedianMultiplier(engineers []EngineerStats) float64 {
	if len(engineers) == 0 {
		return 1.0
	}

	mults := make([]float64, len(engineers))
	for i, e := range engineers {
		mults[i] = e.Multiplier
	}
	slices.Sort(mults)

	mid := len(mults) / 2
	if len(mults)%2 == 0 {
		return (mults[mid-1] + mults[mid]) / 2
	}
	return mults[mid]
}

// EmailPrefix extracts the prefix (username) from an email address
func EmailPrefix(email string) string {
	if email == "" {
		return "unknown"
	}
	at := strings.Index(email, "@")
	if at == -1 {
		return email
	}
	return email[:at]
}
