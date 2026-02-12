package aggregator

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/modern-tooling/aloc/internal/git"
	"github.com/modern-tooling/aloc/internal/model"
)

const Version = "0.2.0"

type Options struct {
	IncludeFiles     bool
	RepoInfo         *model.RepoInfo
	IncludeEffort    bool
	EffortOpts       EffortOptions
	GitAnalysis      bool
	GitOpts          git.Options
	EngineerAnalysis bool
	EngineerOpts     git.EngineerOptions
}

func Compute(records []*model.FileRecord, opts Options) *model.Report {
	responsibilities := ComputeResponsibilities(records)

	report := &model.Report{
		Meta: model.Meta{
			SchemaVersion:    "1.0",
			GeneratedAt:      time.Now().UTC(),
			Generator:        "aloc",
			GeneratorVersion: Version,
			Repo:             opts.RepoInfo,
		},
		Summary:          ComputeSummary(records),
		Responsibilities: responsibilities,
		Ratios:           ComputeRatios(responsibilities),
		Languages:        ComputeLanguageBreakdown(records),
		Confidence:       computeConfidenceInfo(records),
	}

	if opts.IncludeEffort {
		report.Effort = ComputeEffortWithResponsibilities(
			report.Summary.LOCTotal,
			report.Summary.Lines,
			responsibilities,
			report.Ratios,
			opts.EffortOpts,
		)
	}

	if opts.IncludeFiles {
		report.Files = records
	}

	// git analysis (optional)
	if opts.GitAnalysis && opts.RepoInfo != nil && opts.RepoInfo.Root != "" {
		gitMetrics, err := git.Analyze(opts.RepoInfo.Root, records, opts.GitOpts)
		if err != nil {
			log.Printf("git analysis: %v", err)
		} else if gitMetrics != nil {
			report.Git = convertGitMetrics(gitMetrics)

			// apply git adjustments to effort if both present
			if report.Effort != nil && gitMetrics.NetAdjustment != 0 {
				applyGitAdjustments(report.Effort, gitMetrics.NetAdjustment)
			}
		}
	} else if opts.RepoInfo != nil && opts.RepoInfo.Root != "" {
		// detect git repo for hint (lightweight)
		hint, err := git.DetectRepo(opts.RepoInfo.Root)
		if err == nil && hint != nil && hint.HasGit {
			report.GitHint = convertGitHint(hint)
		}
	}

	// engineer throughput analysis (optional, separate from git analysis)
	if opts.EngineerAnalysis && opts.RepoInfo != nil && opts.RepoInfo.Root != "" {
		engineerAnalysis, err := computeEngineerMetrics(opts.RepoInfo.Root, records, opts.EngineerOpts)
		if err != nil {
			log.Printf("engineer analysis: %v", err)
		} else if engineerAnalysis != nil {
			report.Engineer = engineerAnalysis
		}
	}

	return report
}

// computeEngineerMetrics runs engineer throughput analysis
func computeEngineerMetrics(root string, records []*model.FileRecord, opts git.EngineerOptions) (*model.EngineerMetrics, error) {
	// parse git history with author emails preserved
	events, err := git.ParseHistory(git.ParseOptions{
		SinceMonths:     opts.PeriodMonths,
		Root:            root,
		PreserveAuthors: true,
	})
	if err != nil {
		return nil, fmt.Errorf("parse git history: %w", err)
	}

	if len(events) == 0 {
		return nil, nil
	}

	// map roles to events
	git.MapRoles(events, records)

	// calculate engineer stats
	analysis := git.CalculateEngineerStats(events, opts)
	if analysis == nil {
		return nil, nil
	}

	return convertEngineerAnalysis(analysis), nil
}

// convertEngineerAnalysis converts internal engineer analysis to model format
func convertEngineerAnalysis(a *git.EngineerAnalysis) *model.EngineerMetrics {
	engineers := make([]model.EngineerStat, len(a.Engineers))
	for i, e := range a.Engineers {
		engineers[i] = model.EngineerStat{
			AuthorEmail: e.AuthorEmail,
			AuthorName:  e.AuthorName,
			TotalLOC:    e.TotalLOC,
			LOCPerDay:   e.LOCPerDay,
			Multiplier:  e.Multiplier,
			AIPercent:   e.AIPercent,
			CommitCount: e.CommitCount,
		}
	}

	return &model.EngineerMetrics{
		Engineers:    engineers,
		BaselineLOC:  a.BaselineLOC,
		PeriodMonths: a.PeriodMonths,
		MedianMult:   a.MedianMult,
		Caveat:       a.Caveat,
	}
}

// convertGitMetrics converts internal git metrics to model format
func convertGitMetrics(g *git.GitMetrics) *model.GitMetrics {
	m := &model.GitMetrics{
		ChurnConcentration: model.GitChurnStat{
			FilePercent: g.ChurnConcentration.FilePercent,
			EditPercent: g.ChurnConcentration.EditPercent,
		},
		StableCore:             g.StableCore,
		VolatileSurface:        g.VolatileSurface,
		RewritePressure:        g.RewritePressure,
		OwnershipConcentration: g.OwnershipConcentration,
		ParallelismSignal:      g.ParallelismSignal,
		AITimeline:             g.AITimeline,
		HasAnyAI:               g.HasAnyAI,
		NetAdjustment:          g.NetAdjustment,
		WindowMonths:           g.WindowMonths,
		BucketCount:            g.BucketCount,
		CommitCount:            g.CommitCount,
		AnalysisNote:           g.AnalysisNote,
	}

	// convert sparklines (include raw values for adaptive rendering)
	if len(g.ChurnSeries) > 0 {
		m.ChurnSeries = make(map[model.Role]model.GitSparkline)
		for role, sparkline := range g.ChurnSeries {
			m.ChurnSeries[role] = model.GitSparkline{
				Glyphs: sparkline.Glyphs,
				Values: sparkline.Values,
			}
		}
	}

	// convert adjustments
	for _, adj := range g.Adjustments {
		m.Adjustments = append(m.Adjustments, model.GitEffortAdjustment{
			Reason:     adj.Reason,
			Adjustment: adj.Adjustment,
		})
	}

	return m
}

// convertGitHint converts internal git hint to model format
func convertGitHint(h *git.RepoHint) *model.GitHint {
	hint := &model.GitHint{
		HasGit:   h.HasGit,
		IsActive: h.IsActive,
	}

	if h.RepoAge > 0 {
		hint.RepoAge = humanizeDuration(h.RepoAge)
	}

	if !h.LastCommit.IsZero() {
		hint.LastCommit = humanizeTime(h.LastCommit)
	}

	return hint
}

// humanizeDuration formats a duration in human-readable form
func humanizeDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days >= 365 {
		years := float64(days) / 365.25
		return fmt.Sprintf("%.1f years", years)
	}
	if days >= 30 {
		months := float64(days) / 30.44
		return fmt.Sprintf("%.1f months", months)
	}
	if days > 0 {
		return fmt.Sprintf("%d days", days)
	}
	hours := int(d.Hours())
	if hours > 0 {
		return fmt.Sprintf("%d hours", hours)
	}
	return "moments"
}

// humanizeTime formats a time relative to now
func humanizeTime(t time.Time) string {
	d := time.Since(t)
	days := int(d.Hours() / 24)

	if days == 0 {
		hours := int(d.Hours())
		if hours == 0 {
			return "just now"
		}
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	if days == 1 {
		return "yesterday"
	}
	if days < 7 {
		return fmt.Sprintf("%d days ago", days)
	}
	if days < 30 {
		weeks := days / 7
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
	if days < 365 {
		months := days / 30
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
	years := float64(days) / 365.25
	return fmt.Sprintf("%.1f years ago", years)
}

// applyGitAdjustments applies git-based effort adjustments
func applyGitAdjustments(effort *model.EffortEstimates, netAdjustment float64) {
	factor := 1.0 + netAdjustment

	if effort.Human != nil {
		effort.Human.EstimatedCost *= factor
	}

	if effort.Comparison != nil {
		effort.Comparison.HumanOnly *= factor
		effort.Comparison.HybridCost *= factor
	}

	if effort.Conventional != nil {
		effort.Conventional.Cost.Low *= factor
		effort.Conventional.Cost.High *= factor
	}

	if effort.Agentic != nil {
		effort.Agentic.Cost.Low *= factor
		effort.Agentic.Cost.High *= factor
	}
}

func computeConfidenceInfo(records []*model.FileRecord) model.ConfidenceInfo {
	var totalLOC int
	var highConfLOC int
	var overrideLOC int

	for _, r := range records {
		totalLOC += r.LOC
		if r.Confidence >= 0.80 {
			highConfLOC += r.LOC
		}
		if slices.Contains(r.Signals, model.SignalOverride) {
			overrideLOC += r.LOC
		}
	}

	if totalLOC == 0 {
		return model.ConfidenceInfo{}
	}

	return model.ConfidenceInfo{
		AutoClassified: float32(highConfLOC) / float32(totalLOC),
		Heuristic:      float32(totalLOC-highConfLOC-overrideLOC) / float32(totalLOC),
		Override:       float32(overrideLOC) / float32(totalLOC),
	}
}
