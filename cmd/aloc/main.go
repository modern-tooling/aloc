package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/modern-tooling/aloc/internal/aggregator"
	"github.com/modern-tooling/aloc/internal/effort"
	"github.com/modern-tooling/aloc/internal/git"
	"github.com/modern-tooling/aloc/internal/inference"
	"github.com/modern-tooling/aloc/internal/model"
	"github.com/modern-tooling/aloc/internal/renderer"
	jsonrenderer "github.com/modern-tooling/aloc/internal/renderer/json"
	"github.com/modern-tooling/aloc/internal/renderer/tui"
	"github.com/modern-tooling/aloc/internal/scanner"
	"github.com/modern-tooling/aloc/pkg/config"
	"github.com/spf13/cobra"
)

// Build-time variables (set via ldflags)
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

var (
	formatFlag         string
	noColorFlag        bool
	filesFlag          bool
	prettyFlag         bool
	headerProbeFlag    bool
	configFlag         string
	effortFlag         bool
	noEffortFlag       bool
	aiModelFlag        string
	humanCostFlag      float64
	deepFlag           bool
	versionFlag        bool
	noEmbeddedFlag     bool
	gitFlag            bool
	gitMonthsFlag      int
	gitSmoothFlag      bool
	modelConfigFlag    string
	profileFlag        string
	engineerFlag       bool
	engineerMonthsFlag int
)

var rootCmd = &cobra.Command{
	Use:   "aloc [path]",
	Short: "Semantic LOC counter - understand your codebase by role",
	Long: `aloc analyzes codebases and classifies files by semantic role
(prod, test, infra, docs, etc.) rather than just language.

Quick mode (default): Scans only files with known source extensions.
Deep mode (--deep): Also analyzes extensionless files and probes headers.

Output includes responsibility breakdown, key ratios, and
language composition with Tufte-inspired visualization.`,
	Args: cobra.MaximumNArgs(1),
	RunE: run,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("aloc %s\n", version)
		fmt.Printf("  commit:  %s\n", commit)
		fmt.Printf("  built:   %s\n", buildTime)
		fmt.Printf("  go:      %s\n", runtime.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print version and exit")
	rootCmd.Flags().StringVarP(&formatFlag, "format", "f", "tui", "Output format (tui, json)")
	rootCmd.Flags().BoolVar(&noColorFlag, "no-color", false, "Disable colors")
	rootCmd.Flags().BoolVar(&filesFlag, "files", false, "Include file-level details in output")
	rootCmd.Flags().BoolVar(&prettyFlag, "pretty", false, "Pretty-print JSON output")
	rootCmd.Flags().BoolVar(&headerProbeFlag, "header-probe", false, "Enable header content probing")
	rootCmd.Flags().BoolVar(&deepFlag, "deep", false, "Enable expensive analysis (header probing, extensionless files)")
	rootCmd.Flags().StringVarP(&configFlag, "config", "c", "", "Config file path")
	rootCmd.Flags().BoolVar(&effortFlag, "effort", true, "Include effort estimates (human and AI cost)")
	rootCmd.Flags().BoolVar(&noEffortFlag, "no-effort", false, "Disable effort estimates")
	rootCmd.Flags().StringVar(&aiModelFlag, "ai-model", "sonnet", "AI model for cost estimation (sonnet, opus, haiku)")
	rootCmd.Flags().Float64Var(&humanCostFlag, "human-cost", 0, "Monthly cost per engineer (0 = use blended cost from team composition)")
	rootCmd.Flags().BoolVar(&noEmbeddedFlag, "no-embedded", false, "Hide embedded code blocks in Markdown")
	rootCmd.Flags().BoolVar(&gitFlag, "git", false, "Enable git history analysis for churn and stability signals")
	rootCmd.Flags().IntVar(&gitMonthsFlag, "git-months", 6, "Months of history for sparklines")
	rootCmd.Flags().BoolVar(&gitSmoothFlag, "git-smooth", false, "Use bi-weekly buckets instead of weekly for smoother sparklines")
	rootCmd.Flags().StringVar(&modelConfigFlag, "model-config", "", "Path to JSON file with effort model configuration overrides")
	rootCmd.Flags().StringVar(&profileFlag, "profile", "faang", "Effort estimation profile (faang)")
	rootCmd.Flags().BoolVar(&engineerFlag, "engineer", false, "Show engineer throughput analysis (replaces standard output)")
	rootCmd.Flags().IntVar(&engineerMonthsFlag, "engineer-months", 6, "Months of history for engineer analysis")
}

func main() {
	// Handle -v/--version flag before cobra
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("aloc %s (%s)\n", version, commit)
		return
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load model config early (before any effort calculations)
	// Priority: --model-config file > --profile > default profile (faang)
	if modelConfigFlag != "" {
		modelCfg, err := effort.LoadModelConfig(modelConfigFlag)
		if err != nil {
			return fmt.Errorf("model config error: %w", err)
		}
		effort.SetModelConfig(modelCfg)
	} else {
		// load profile (defaults to "faang" if not specified)
		modelCfg, err := effort.LoadProfile(profileFlag)
		if err != nil {
			return fmt.Errorf("profile error: %w", err)
		}
		effort.SetModelConfig(modelCfg)
	}

	// Determine root path
	root := "."
	if len(args) > 0 {
		root = args[0]
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Load config
	var cfg *config.Config
	if configFlag != "" {
		cfg, err = config.Load(configFlag)
	} else {
		cfg, err = config.LoadFromDir(absRoot)
	}
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	// Create scanner
	s, err := scanner.NewScanner(absRoot, scanner.Options{
		NumWorkers: runtime.NumCPU() * 2,
		Exclude:    cfg.Exclude,
		DeepMode:   deepFlag,
	})
	if err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	// Scan files
	rawFiles, errs := s.Scan(ctx)

	// Collect files
	var files []*model.RawFile
	for f := range rawFiles {
		files = append(files, f)
	}

	// Log errors (non-fatal)
	for err := range errs {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found in %s", absRoot)
	}

	// Create inference engine
	engine := inference.NewEngine(inference.Options{
		HeaderProbe:  deepFlag || headerProbeFlag || cfg.Options.HeaderProbe,
		Neighborhood: cfg.Options.Neighborhood,
		Overrides:    cfg.Overrides,
	})

	// Infer roles
	records := engine.InferBatch(files)

	// Determine if effort should be included (default true, unless --no-effort)
	includeEffort := effortFlag && !noEffortFlag

	// Auto-enable git when engineer mode is set
	enableGit := gitFlag || engineerFlag

	// Aggregate
	report := aggregator.Compute(records, aggregator.Options{
		IncludeFiles:  filesFlag,
		IncludeEffort: includeEffort,
		EffortOpts: aggregator.EffortOptions{
			IncludeHuman:      includeEffort,
			IncludeAI:         includeEffort,
			AIModel:           aiModelFlag,
			HumanCostPerMonth: humanCostFlag,
		},
		RepoInfo: &model.RepoInfo{
			Name: filepath.Base(absRoot),
			Root: absRoot,
		},
		GitAnalysis: enableGit,
		GitOpts: git.Options{
			SparklineMonths: gitMonthsFlag,
			StabilityMonths: 18,
			Smooth:          gitSmoothFlag,
		},
		EngineerAnalysis: engineerFlag,
		EngineerOpts: git.EngineerOptions{
			PeriodMonths: engineerMonthsFlag,
		},
	})

	// Select renderer
	opts := renderer.Options{
		Writer:     os.Stdout,
		NoColor:    noColorFlag || renderer.ShouldDisableColor(),
		Pretty:     prettyFlag,
		NoEmbedded: noEmbeddedFlag,
	}

	// Engineer mode uses separate render path (replaces standard output)
	if engineerFlag {
		return renderEngineerMode(report, opts, formatFlag)
	}

	var r renderer.Renderer
	switch formatFlag {
	case "json":
		r = jsonrenderer.NewJSONRenderer(opts)
	default:
		r = tui.NewTUIRenderer(opts)
	}

	return r.Render(report)
}

// renderEngineerMode renders only the engineer throughput analysis
func renderEngineerMode(report *model.Report, opts renderer.Options, format string) error {
	if report.Engineer == nil {
		return fmt.Errorf("no engineer data available (requires git history with identifiable authors)")
	}

	switch format {
	case "json":
		r := jsonrenderer.NewJSONRenderer(opts)
		// create a minimal report with just engineer data
		engineerReport := &model.Report{
			Meta:     report.Meta,
			Engineer: report.Engineer,
		}
		return r.Render(engineerReport)
	default:
		// TUI mode - render only the engineer table
		var theme *renderer.Theme
		if opts.NoColor {
			theme = renderer.NewNoColorTheme()
		} else {
			theme = renderer.NewDefaultTheme()
		}
		output := tui.RenderEngineerThroughput(report.Engineer, theme)
		_, err := opts.Writer.Write([]byte(output))
		return err
	}
}
