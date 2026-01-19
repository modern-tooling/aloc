# Implementation Guide

Step-by-step implementation plan for parallel development by multiple agents.

## Prerequisites

- Go 1.22+
- Understanding of the architecture docs

## Phase Overview

```
Phase 1: Foundation (Model + Scanner)
    |
    v
Phase 2: Inference Engine
    |
    v
Phase 3: Aggregation
    |
    v
Phase 4: TUI Renderer
    |
    v
Phase 5: CLI + Integration
    |
    v
Phase 6: Additional Renderers + Polish
```

## Phase 1: Foundation

### 1.1 Project Scaffolding

Create the directory structure:

```
aloc/
├── cmd/aloc/main.go
├── internal/
│   ├── scanner/
│   ├── inference/
│   ├── model/
│   ├── aggregator/
│   └── renderer/
├── pkg/config/
├── testdata/
├── go.mod
└── go.sum
```

Initialize module:
```bash
go mod init github.com/[org]/aloc
```

### 1.2 Model Layer (`internal/model/`)

**File: `types.go`**

Define core enums:

```go
type Role string

const (
    RoleProd       Role = "prod"
    RoleTest       Role = "test"
    RoleInfra      Role = "infra"
    RoleDocs       Role = "docs"
    RoleConfig     Role = "config"
    RoleGenerated  Role = "generated"
    RoleVendor     Role = "vendor"
    RoleScripts    Role = "scripts"
    RoleExamples   Role = "examples"
    RoleDeprecated Role = "deprecated"
)

type TestKind string

const (
    TestUnit        TestKind = "unit"
    TestIntegration TestKind = "integration"
    TestE2E         TestKind = "e2e"
    TestContract    TestKind = "contract"
    TestFixture     TestKind = "fixture"
)

type Signal string

const (
    SignalPath         Signal = "path"
    SignalFilename     Signal = "filename"
    SignalExtension    Signal = "extension"
    SignalNeighborhood Signal = "neighborhood"
    SignalHeader       Signal = "header"
    SignalOverride     Signal = "override"
)
```

**File: `file.go`**

```go
type RawFile struct {
    Path         string
    Bytes        int64
    LOC          int
    LanguageHint string
}

type FileRecord struct {
    Path       string    `json:"path"`
    LOC        int       `json:"loc"`
    Language   string    `json:"language"`
    Role       Role      `json:"role"`
    SubRole    TestKind  `json:"sub_role,omitempty"`
    Confidence float32   `json:"confidence"`
    Signals    []Signal  `json:"signals"`
}
```

**File: `report.go`**

See JSON schema doc for complete struct definitions.

### 1.3 Scanner Layer (`internal/scanner/`)

**Dependencies:**
```bash
go get github.com/go-git/go-git/v5
go get github.com/monochromegane/go-gitignore
```

**File: `walker.go`**

```go
type Walker struct {
    root       string
    ignorer    gitignore.Matcher
    numWorkers int
}

func NewWalker(root string, numWorkers int) (*Walker, error)
func (w *Walker) Walk(ctx context.Context) (<-chan RawFile, <-chan error)
```

Key implementation notes:
- Use `filepath.WalkDir` for traversal
- Check `.gitignore` at each directory level
- Skip binary files (check magic bytes)
- Worker pool receives file paths, emits `RawFile`

**File: `counter.go`**

```go
func CountLOC(path string) (int, error)
```

Implementation:
- Read file into buffer (reuse buffers via sync.Pool)
- Count non-blank, non-comment lines
- Use language-specific comment detection

**File: `language.go`**

```go
func DetectLanguage(path string, header []byte) string
```

Implementation:
- Primary: extension mapping
- Secondary: shebang detection
- Fallback: "unknown"

## Phase 2: Inference Engine

### 2.1 Rule Tables (`internal/inference/rules.go`)

Define static tables:

```go
type PathRule struct {
    Fragment string
    Role     Role
    Weight   float32
}

var PathRules = []PathRule{
    {"/test/", RoleTest, 0.60},
    {"/tests/", RoleTest, 0.60},
    // ... see heuristic-rules.md for complete list
}

type FilenameRule struct {
    Suffix string  // or Prefix, Contains
    Role   Role
    Weight float32
}

var FilenameRules = []FilenameRule{
    {"_test.", RoleTest, 0.75},
    {".spec.", RoleTest, 0.70},
    // ...
}
```

### 2.2 Scoring Engine (`internal/inference/scoring.go`)

```go
type RoleScore struct {
    Weights map[Role]float32
    Signals map[Role][]Signal
}

func NewRoleScore() *RoleScore
func (s *RoleScore) Add(role Role, weight float32, signal Signal)
func (s *RoleScore) Resolve() (Role, float32, []Signal)
```

### 2.3 Inference Orchestration (`internal/inference/engine.go`)

```go
type Engine struct {
    overrides *Overrides
    enableHeaderProbe bool
}

func NewEngine(opts ...Option) *Engine
func (e *Engine) Infer(file *RawFile, ctx *Context) *FileRecord
```

Context contains neighborhood info for second-pass inference.

## Phase 3: Aggregation

### 3.1 Summary Computation (`internal/aggregator/summary.go`)

```go
func ComputeSummary(records []*FileRecord) *Summary
func ComputeResponsibilities(records []*FileRecord) []Responsibility
```

### 3.2 Ratio Calculation (`internal/aggregator/ratios.go`)

```go
func ComputeRatios(responsibilities []Responsibility) *Ratios
```

### 3.3 Language Breakdown (`internal/aggregator/languages.go`)

```go
func ComputeLanguageBreakdown(records []*FileRecord) []LanguageComposition
```

## Phase 4: TUI Renderer

### 4.1 Dependencies

```bash
go get github.com/charmbracelet/lipgloss
```

### 4.2 Color Tokens (`internal/renderer/colors.go`)

See `color-system.md` for implementation.

### 4.3 TUI Renderer (`internal/renderer/tui/renderer.go`)

```go
type TUIRenderer struct {
    theme  *Theme
    width  int
    writer io.Writer
}

func NewTUIRenderer(opts ...Option) *TUIRenderer
func (r *TUIRenderer) Render(report *Report) error
```

Sections to implement:
1. Header block
2. Responsibility ledger
3. Key ratios
4. Language composition
5. Confidence footer

### 4.4 Bar Rendering (`internal/renderer/tui/bars.go`)

```go
func RenderBar(value, maxValue int, maxWidth int) string
```

## Phase 5: CLI Integration

### 5.1 Dependencies

```bash
go get github.com/spf13/cobra
go get github.com/spf13/viper
```

### 5.2 CLI Structure (`cmd/aloc/main.go`)

```go
var rootCmd = &cobra.Command{
    Use:   "aloc [path]",
    Short: "Semantic LOC counter",
    RunE:  run,
}

func init() {
    rootCmd.Flags().String("format", "tui", "Output format (tui, json, html)")
    rootCmd.Flags().String("by", "role", "Primary grouping (role, language)")
    rootCmd.Flags().Bool("risk", false, "Show risk leaderboard")
    rootCmd.Flags().String("config", "", "Config file path")
}
```

### 5.3 Pipeline Orchestration

```go
func run(cmd *cobra.Command, args []string) error {
    // 1. Parse config
    cfg, err := config.Load(configPath)

    // 2. Create scanner
    walker := scanner.NewWalker(root, runtime.NumCPU())

    // 3. Scan files
    rawFiles, errs := walker.Walk(ctx)

    // 4. Infer roles (parallel)
    engine := inference.NewEngine(cfg.Overrides...)
    records := inferAll(engine, rawFiles)

    // 5. Aggregate
    report := aggregator.Compute(records)

    // 6. Render
    renderer := selectRenderer(format)
    return renderer.Render(report)
}
```

## Phase 6: Polish & Additional Features

### 6.1 JSON Renderer

```go
type JSONRenderer struct {
    pretty bool
    writer io.Writer
}

func (r *JSONRenderer) Render(report *Report) error {
    enc := json.NewEncoder(r.writer)
    if r.pretty {
        enc.SetIndent("", "  ")
    }
    return enc.Encode(report)
}
```

### 6.2 Config File Support (`pkg/config/`)

```go
type Config struct {
    Overrides map[Role][]string
    Exclude   []string
    Options   struct {
        HeaderProbe  bool
        Neighborhood bool
    }
}

func Load(path string) (*Config, error)
```

### 6.3 Git Integration (`internal/scanner/git.go`)

For future churn/author analysis:

```go
func GetFileChurn(repo *git.Repository, path string, days int) (int, error)
func GetFileAuthors(repo *git.Repository, path string, days int) (int, error)
```

## Testing Strategy

### Unit Tests

| Package | Focus |
|---------|-------|
| `model` | JSON serialization round-trip |
| `scanner/counter` | LOC accuracy against known files |
| `scanner/language` | Extension + shebang detection |
| `inference/rules` | Rule firing conditions |
| `inference/scoring` | Weight accumulation, confidence calculation |
| `aggregator` | Math correctness |
| `renderer/tui` | Output format stability (snapshot tests) |

### Integration Tests

Use `testdata/` fixtures:

```
testdata/
├── simple-go/          # Basic Go project
├── monorepo/           # Multi-language project
├── generated-heavy/    # Generated code detection
└── expected/           # Golden file outputs
```

### Benchmarks

```go
func BenchmarkScanLargeRepo(b *testing.B)
func BenchmarkInference(b *testing.B)
func BenchmarkFullPipeline(b *testing.B)
```

Target: 100k files in < 2 seconds on SSD.

## Parallelization Strategy

For multiple agent implementation:

### Independent Work Streams

1. **Model + Types** - No dependencies, can start immediately
2. **Scanner** - Depends only on model
3. **Inference Rules** - Depends only on model
4. **Inference Engine** - Depends on rules + model
5. **Aggregator** - Depends on model
6. **TUI Renderer** - Depends on model + colors
7. **JSON Renderer** - Depends on model
8. **CLI** - Integrates everything

### Suggested Agent Assignment

| Agent | Packages | Dependencies |
|-------|----------|--------------|
| A | model, scanner | None |
| B | inference (rules, scoring, engine) | model |
| C | aggregator | model |
| D | renderer (colors, tui, json) | model |
| E | CLI, config, integration | All |

Agents A-D can work in parallel. Agent E integrates after others complete.

## Quality Gates

Before merge:

- [ ] All tests pass
- [ ] No race conditions (`go test -race`)
- [ ] Benchmarks don't regress
- [ ] `golangci-lint` passes
- [ ] Coverage > 80%
- [ ] Output matches wireframe spec
