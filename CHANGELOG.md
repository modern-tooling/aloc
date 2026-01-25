# Changelog

All notable changes to aloc are documented here.

## [v0.4.0] - 2026-01-25

### Added

- **Effort profile system** with `--profile` flag for pre-configured estimation models
- **FAANG profile (default)**: Calibrated to 80 LOC/day based on Smacchia's 14-year productivity study
  - COCOMO coefficients: organic a=0.43 (vs 2.4 in 1981), semi-detached a=0.53, embedded a=0.64
  - Results in ~5-6x more realistic estimates for modern high-performing teams
- **Configurable effort model**: All COCOMO coefficients, skill bands, token params, and AI-native factors now configurable via `--model-config`
- **Blended cost calculation**: Automatic cost estimation based on team composition by size tier
- **AI leverage by skill level**: Principal+ engineers modeled with higher AI tool leverage (up to 10x for Distinguished)
- **Effective capacity display**: TUI now shows AI-leveraged output capacity for agentic teams

### Changed

- Default effort estimates now use FAANG profile (more realistic for modern development)
- `--human-cost 0` now uses blended cost from team composition instead of fixed $15K/month

### Documentation

- Added engineering-cost-methodology.md with full derivation and research references
- Added Smacchia (NDepend) citation for 80 LOC/day productivity benchmark
- Updated ai-cost-model.md with profiles section

## [v0.3.0] - 2026-01-23

### Added

- AI-assisted commit detection in `--git` analysis
- Identifies commits with AI co-author signatures

## [v0.2.0] - 2026-01-23

### Added

- `--git` flag for codebase dynamics analysis
- Churn and stability signals from git history
- Sparkline visualizations for commit activity

## [v0.1.2] - 2026-01-23

### Fixed

- Module path for `go install` compatibility

## [v0.1.1] - 2026-01-23

### Added

- 250+ language support
- `.ignore` file support

## [v0.1.0] - 2026-01-23

### Added

- Initial release
- Semantic LOC counting by responsibility (prod, test, infra, docs, config)
- Language detection and breakdown
- Health ratios visualization
- Effort estimation with COCOMO and AI-native models
