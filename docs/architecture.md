# aloc Architecture

**Semantic LOC Counter** - A next-generation code metrics tool that classifies code by semantic role, not just language.

## Overview

aloc answers the question humans actually care about: *"What kind of work is this codebase made of?"* rather than *"How many lines of each language?"*

```
Traditional tools:  Go 81,965 LOC
aloc:               Production  62,140
                    Tests       19,825
                    Infra        5,712
```

## Design Principles

1. **Role-first, language-second** - Semantic classification (core/test/infra) matters more than language
2. **Fast inference over deep parsing** - Heuristics from paths/filenames beat slow AST analysis
3. **Confidence + explainability** - Every classification includes confidence score and signals used
4. **Tufte-inspired output** - High data density, low ink, semantic color, no decoration
5. **Renderer-agnostic** - Clean separation between data model and presentation

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                            CLI Layer                                │
│   (argument parsing, orchestration, output format selection)        │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Pipeline Engine                             │
│            scan → infer → aggregate → render                        │
└──────┬─────────────┬──────────────┬──────────────┬──────────────────┘
       │             │              │              │
       ▼             ▼              ▼              ▼
   ┌───────┐    ┌─────────┐   ┌───────────┐   ┌──────────┐
   │Scanner│    │Inference│   │Aggregator │   │Renderers │
   │       │    │ Engine  │   │           │   │          │
   │- walk │    │- rules  │   │- summaries│   │- TUI     │
   │- LOC  │    │- scoring│   │- ratios   │   │- JSON    │
   │- git  │    │- resolve│   │- trends   │   │- HTML    │
   └───┬───┘    └────┬────┘   └─────┬─────┘   └────┬─────┘
       │             │              │              │
       └─────────────┴──────────────┴──────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │  Model Layer  │
                    │ (domain types │
                    │  JSON schema) │
                    └───────────────┘
```

## Core Data Flow

```
Files on disk
      │
      ▼ Scanner
RawFile { path, bytes, loc, language_hint }
      │
      ▼ Inference
FileRecord { path, loc, language, role, confidence, signals }
      │
      ▼ Aggregator
Report { summary, responsibilities, ratios, languages, trends }
      │
      ▼ Renderer
Terminal / JSON / HTML output
```

## Package Structure

```
aloc/
├── cmd/
│   └── aloc/
│       └── main.go           # CLI entrypoint
├── internal/
│   ├── scanner/
│   │   ├── walker.go         # Parallel filesystem traversal
│   │   ├── counter.go        # LOC counting (tokei-style)
│   │   ├── language.go       # Language detection
│   │   └── git.go            # Git metadata (churn, authors)
│   ├── inference/
│   │   ├── engine.go         # Main inference orchestration
│   │   ├── rules.go          # Rule definitions (path, filename, etc.)
│   │   ├── scoring.go        # Weight accumulation + resolution
│   │   └── overrides.go      # User override handling
│   ├── model/
│   │   ├── types.go          # Role, Signal, TestKind enums
│   │   ├── file.go           # RawFile, FileRecord
│   │   ├── report.go         # Report, Summary, Ratios
│   │   └── schema.go         # JSON schema generation
│   ├── aggregator/
│   │   ├── summary.go        # Responsibility totals
│   │   ├── ratios.go         # Key ratio calculations
│   │   ├── languages.go      # Language x role breakdown
│   │   └── trends.go         # Historical trend computation
│   └── renderer/
│       ├── contract.go       # Renderer interface
│       ├── colors.go         # Semantic color tokens
│       ├── tui/
│       │   ├── renderer.go   # Terminal output (lipgloss)
│       │   ├── bars.go       # Bar chart rendering
│       │   └── styles.go     # Theme definitions
│       ├── json/
│       │   └── renderer.go   # JSON output
│       └── html/
│           └── renderer.go   # HTML report generation
├── pkg/
│   └── config/
│       ├── config.go         # Configuration loading
│       └── overrides.go      # Override file parsing
├── testdata/                 # Test fixtures
└── docs/                     # Documentation
```

## Key Components

### 1. Scanner

**Responsibilities:**
- Parallel filesystem traversal (respects .gitignore)
- LOC counting without semantic inference
- Language detection via extension/shebang
- Optional git metadata extraction

**Performance targets:**
- 100k files in < 2 seconds
- Worker pool sized for SSD throughput
- 64-256KB read buffers, reused
- No per-line allocations in hot paths

**Output:** Stream of `RawFile` structs

### 2. Inference Engine

**Responsibilities:**
- Apply heuristic rules to classify files
- Accumulate weighted evidence per role
- Resolve conflicts with confidence penalty
- Record which signals contributed

**Rule priority (highest to lowest):**
1. User overrides (weight: 1.0)
2. Header probes (0.60-0.95)
3. Path fragments (0.55-0.90)
4. Filename patterns (0.60-0.90)
5. Neighborhood consensus (0.40)
6. Extension bias (0.10-0.20)

**Output:** Stream of `FileRecord` structs with role + confidence + signals

### 3. Aggregator

**Responsibilities:**
- Compute totals by role
- Calculate key ratios (test/core, infra/core, etc.)
- Break down by language within role
- Compute trends if historical data exists

**Output:** `Report` struct ready for rendering

### 4. Renderers

**Contract:**
- Receive fully-computed `Report`
- Never perform inference or aggregation
- Apply semantic colors to labels only
- Numbers are never colored
- Graceful ASCII fallback

## Semantic Roles

| Role | Meaning | Detection Priority |
|------|---------|-------------------|
| `core` | Core application logic | Default fallback |
| `test` | Tests (unit, integration, e2e) | Filename + path |
| `infra` | Infrastructure & deployment | Path + filename |
| `docs` | Human-facing documentation | Extension + path |
| `config` | Runtime/build configuration | Filename |
| `generated` | Auto-generated code | Header probe |
| `vendor` | Third-party code | Path (near-certain) |
| `scripts` | One-off tooling/glue | Path |
| `examples` | Samples, demos | Path |
| `deprecated` | Technical debt | Header/comment |

### Test Sub-roles

| Sub-role | Meaning |
|----------|---------|
| `unit` | Fast, isolated tests |
| `integration` | Cross-boundary tests |
| `e2e` | System-level tests |
| `contract` | API/interface guarantees |
| `fixture` | Test data, mocks, golden files |

## Semantic Color System

Colors encode responsibility, not magnitude or language.

| Token | Role | Visual Intent |
|-------|------|---------------|
| `semantic.primary` | core | Neutral foreground |
| `semantic.safety` | test | Muted green/blue |
| `semantic.operational` | infra | Amber/copper |
| `semantic.knowledge` | docs | Soft gray |
| `semantic.fragility` | config | Gray-blue |
| `semantic.low_emphasis` | generated | Faint/dim |
| `semantic.external` | vendor | Italic/dim |
| `semantic.warning` | deprecated | Desaturated red |

**Renderer rule:** Colors apply only to labels, never to numbers.

## Configuration

### User Overrides (`lloc.yaml`)

```yaml
overrides:
  core:
    - internal/core/**
  infra:
    - ops/**
  test:
    - integration/**

exclude:
  - vendor/**
  - node_modules/**

options:
  header_probe: false  # disable content peeking
  neighborhood: true   # enable directory consensus
```

### CLI Flags

```bash
aloc                    # default summary view
aloc--by role           # group by role first
aloc--by language       # group by language first
aloc--risk              # show risk leaderboard
aloc--trend 12m         # include trend sparkline
aloc--format json       # machine-readable output
aloc--format html       # web report
aloc--diff HEAD~30      # narrative diff
```

## Output Examples

### Terminal (default)

```
Codebase Summary
────────────────────────────────────────────────────────────
Files: 493          LOC: 126,433          Languages: 7

Responsibility Breakdown (LOC)
────────────────────────────────────────────────────────────
core        ████████████████████████ 62.1k
test        ████████████             29.4k
infra       ██████                   14.2k
docs        ████                     11.5k
config      ██                        6.1k
generated   █                         2.7k

Key Ratios
────────────────────────────────────────────────────────────
Test / Core           0.47    healthy baseline ≈ 0.5–0.8
Infra / Core          0.23    operational complexity
Docs / Core           0.18    well documented
Generated / Core      0.04    low automation reliance

Language Composition
────────────────────────────────────────────────────────────
Go        core ████████████████ 49.2k   test ████████ 18.9k
MDX       docs ████████████     11.1k
Shell     infra ████████         9.4k
YAML      infra ██████           6.8k

Classification confidence: high (92% auto, 8% heuristic)
```

### JSON Schema

See `docs/design/json-schema.md` for complete schema definition.

## Performance Considerations

### Scanner Optimizations
- Use `github.com/monochromegane/go-gitignore` for fast .gitignore matching
- Buffer pool for file reads
- Parallel traversal with bounded concurrency
- Skip binary files via magic byte detection

### Inference Optimizations
- Static rule tables (no regex compilation per file)
- String prefix/suffix checks over regex where possible
- Neighborhood inference runs as second pass (amortized)
- Header probes read max 10 lines, cached

### Memory Management
- Stream results, don't build giant in-memory models
- Worker pool sized to SSD throughput (~16-32 workers)
- Reuse allocations in hot loops
- GC-friendly struct layouts

## Testing Strategy

| Layer | Test Type | Focus |
|-------|-----------|-------|
| Scanner | Unit | LOC accuracy, language detection |
| Inference | Unit | Rule firing, confidence calculation |
| Aggregator | Unit | Ratio math, totals |
| Renderer | Snapshot | Output format stability |
| Integration | E2E | Real repo scanning, golden files |

## Future Considerations

### v0.2
- Import-based boundary detection
- Churn-weighted confidence
- Language-specific heuristic packs

### v0.3
- PR diff mode (`aloc--diff HEAD~30`)
- CI gate mode (fail on ratio thresholds)
- Trend persistence

### v1.0
- Plugin system for custom rules
- LSP integration
- WASM build for web playgrounds
