# Heuristic Scoring Rules

**Version:** 0.1

A probabilistic rules engine for semantic file classification. Fast, explainable, composable, tunable.

## Core Model

Each file accumulates weighted evidence for multiple roles. At the end, normalize, select, compute confidence.

```go
type RoleScore struct {
    Weights map[Role]float32
    Signals map[Role][]Signal
}

func (s *RoleScore) Add(role Role, weight float32, signal Signal) {
    s.Weights[role] += weight
    s.Signals[role] = append(s.Signals[role], signal)
}
```

**Critical rule:** Weights add, never overwrite.

## Scoring Philosophy

1. **Start with no assumptions** - No global priors by default
2. **Prefer path > filename > extension > content** - Signal reliability order
3. **Multiple weak signals beat one strong one** - Agreement matters
4. **Confidence comes from agreement, not magnitude** - Corroboration rewarded

## Rule Priority Order

When rules conflict:

1. User overrides (always win)
2. Vendor detection (near-certain)
3. Generated detection (header-based)
4. Test detection (filename/path strong)
5. Infra detection (filename/path)
6. Everything else

## Rule Classes

### 1. PathIntent (Strongest Signal)

Applies if path contains directory segment match (case-insensitive).

| Path Fragment | Role | Weight |
|--------------|------|--------|
| `/test/`, `/tests/`, `/__tests__/` | test | +0.60 |
| `/spec/`, `/specs/` | test | +0.55 |
| `/unit/` | test:unit | +0.60 |
| `/integration/` | test:integration | +0.65 |
| `/e2e/` | test:e2e | +0.70 |
| `/infra/`, `/terraform/`, `/pulumi/`, `/helm/` | infra | +0.65 |
| `/.github/workflows/` | infra | +0.70 |
| `/.gitlab-ci/` | infra | +0.70 |
| `/ci/`, `/cd/` | infra | +0.60 |
| `/deploy/`, `/deployment/` | infra | +0.65 |
| `/docs/`, `/doc/`, `/documentation/` | docs | +0.65 |
| `/site/`, `/website/` | docs | +0.60 |
| `/config/`, `/configs/`, `/configuration/` | config | +0.55 |
| `/settings/` | config | +0.50 |
| `/scripts/`, `/tools/`, `/bin/` | scripts | +0.55 |
| `/hack/` | scripts | +0.50 |
| `/examples/`, `/samples/`, `/demo/`, `/demos/` | examples | +0.55 |
| `/vendor/`, `/third_party/`, `/external/` | vendor | +0.90 |
| `/node_modules/` | vendor | +0.95 |
| `/dist/`, `/build/`, `/out/` | generated | +0.70 |
| `/gen/`, `/generated/`, `/pb/` | generated | +0.80 |

**Notes:**
- Path match applies once per file
- Vendor paths are near-certain (0.90+)
- CI paths outrank generic infra paths

### 2. FilenamePattern (Strong, Precise)

Matches filename patterns (case-insensitive).

| Pattern | Role | Weight |
|---------|------|--------|
| `*_test.*` | test | +0.75 |
| `*_test` (no ext) | test | +0.70 |
| `*.spec.*` | test | +0.70 |
| `*.test.*` | test | +0.70 |
| `*_spec.*` | test | +0.70 |
| `*.e2e.*` | test:e2e | +0.80 |
| `*_e2e.*` | test:e2e | +0.80 |
| `*.integration.*` | test:integration | +0.80 |
| `*_integration.*` | test:integration | +0.80 |
| `*_fixture.*` | test:fixture | +0.60 |
| `*_mock.*` | test:fixture | +0.55 |
| `*_stub.*` | test:fixture | +0.55 |
| `*_fake.*` | test:fixture | +0.55 |
| `testdata/*` | test:fixture | +0.65 |
| `Dockerfile*` | infra | +0.85 |
| `docker-compose*` | infra | +0.80 |
| `Makefile` | infra | +0.65 |
| `Taskfile*` | infra | +0.65 |
| `Justfile` | infra | +0.65 |
| `*.tf` | infra | +0.90 |
| `*.tfvars` | infra | +0.90 |
| `helmfile*` | infra | +0.85 |
| `*.hcl` | infra | +0.80 |
| `Vagrantfile` | infra | +0.80 |
| `.env*` | config | +0.85 |
| `*.env` | config | +0.80 |
| `config.*` | config | +0.60 |
| `settings.*` | config | +0.60 |
| `*.config.*` | config | +0.55 |
| `*.conf` | config | +0.55 |
| `README*` | docs | +0.80 |
| `CHANGELOG*` | docs | +0.75 |
| `CONTRIBUTING*` | docs | +0.75 |
| `LICENSE*` | docs | +0.70 |
| `AUTHORS*` | docs | +0.70 |

**Note:** Filename patterns beat path if conflict exists.

### 3. ExtensionBias (Weak Signal)

Used only to nudge, never decide alone.

| Extension | Role | Weight |
|-----------|------|--------|
| `.md`, `.mdx` | docs | +0.20 |
| `.rst` | docs | +0.20 |
| `.adoc` | docs | +0.20 |
| `.txt` | docs | +0.10 |
| `.yaml`, `.yml` | config | +0.15 |
| `.toml` | config | +0.15 |
| `.json` | config | +0.10 |
| `.ini` | config | +0.15 |
| `.lock` | generated | +0.90 |
| `.sum` | generated | +0.85 |
| `.pb.go` | generated | +0.90 |
| `.pb.ts` | generated | +0.90 |
| `.gen.go` | generated | +0.85 |
| `.generated.ts` | generated | +0.85 |
| `.d.ts` | generated | +0.40 |
| `.sql` | (context-dependent) | +0.00 |
| `.proto` | docs | +0.40 |

**Note:** Extension bias is ignored if path/filename already decisive (weight > 0.50).

### 4. DirectoryConsensus (Neighborhood Inference)

Run as second pass after initial classification.

**Rule:** If >= 70% of sibling files resolve to the same role with confidence >= 0.7:
- Apply that role with +0.40
- Record signal `Neighborhood`

**Examples:**
- 8/10 files in `/internal/billing/` classified as `prod` -> remaining 2 get +0.40 prod
- 5/6 files in `/test/fixtures/` classified as `test:fixture` -> remaining 1 gets +0.40 test:fixture

**Implementation notes:**
- Skip directories with < 3 files
- Skip vendor directories (already classified)
- Only apply to files with confidence < 0.60

### 5. HeaderProbe (Optional, Capped)

**Strict limits:**
- Read max first 10 lines
- Regex only, no AST
- Must be explicitly enabled

| Pattern | Role | Weight |
|---------|------|--------|
| `Code generated by` | generated | +0.95 |
| `DO NOT EDIT` | generated | +0.90 |
| `@generated` | generated | +0.90 |
| `AUTO-GENERATED` | generated | +0.85 |
| `This file was automatically generated` | generated | +0.90 |
| `terraform {` | infra | +0.80 |
| `provider "` | infra | +0.75 |
| `describe(` (JS/TS) | test | +0.60 |
| `test(` (JS/TS) | test | +0.60 |
| `func Test` (Go) | test | +0.70 |
| `@Test` (Java) | test | +0.65 |
| `#[test]` (Rust) | test | +0.70 |
| `def test_` (Python) | test | +0.65 |
| `package main` in `/cmd/` | prod | +0.60 |
| `// Deprecated:` | deprecated | +0.70 |
| `// DEPRECATED` | deprecated | +0.65 |

**Header probe can override weak conflicts but not vendor detection.**

### 6. UserOverride (Always Wins)

From `lloc.yaml`:

```yaml
overrides:
  prod:
    - internal/core/**
    - pkg/domain/**
  infra:
    - ops/**
    - platform/**
  test:
    - integration/**
```

Override behavior:
- Weight: 1.0
- Signal: `Override`
- Confidence: 1.0
- Supersedes all other rules

## Conflict Resolution

After all rules have fired:

```go
func (s *RoleScore) Resolve() (Role, float32, []Signal) {
    // Sort by weight descending
    ranked := sortByWeight(s.Weights)

    topRole, topWeight := ranked[0]
    confidence := topWeight

    // Ambiguity penalty
    if len(ranked) > 1 {
        secondWeight := ranked[1].Weight
        if topWeight - secondWeight < 0.15 {
            confidence *= 0.8  // 20% penalty
        }
    }

    // Agreement bonus
    signals := s.Signals[topRole]
    agreementFactor := min(1.0, float32(len(signals)) * 0.25)
    confidence = min(confidence * agreementFactor, 1.0)

    return topRole, confidence, signals
}
```

**Agreement factor:**
- 1 signal -> 0.25x
- 2 signals -> 0.50x
- 3 signals -> 0.75x
- 4+ signals -> 1.0x

This rewards multiple independent confirmations.

## Tie-Break Order

If equal weight after resolution:

1. Override
2. Vendor
3. Generated
4. Test
5. Infra
6. Prod
7. Docs
8. Config
9. Scripts
10. Examples

This avoids misclassifying third-party or generated code as prod.

## Default Classification

If no rules fire and weight is 0 for all roles:
- Role: `prod`
- Confidence: 0.30
- Signals: []

## Tuning Guidance

### Increasing Precision
- Raise weight thresholds for ambiguity penalty
- Reduce extension bias weights
- Disable neighborhood inference

### Increasing Recall
- Lower path/filename weights
- Enable header probes
- Lower neighborhood consensus threshold to 60%

### Language-Specific Adjustments

**Go:**
- `_test.go` is extremely reliable (+0.85)
- `/internal/` suggests `prod` (+0.20)
- `/cmd/` suggests `prod` (+0.30)

**TypeScript/JavaScript:**
- `.spec.ts` and `.test.ts` reliable (+0.75)
- `/src/__tests__/` reliable (+0.70)
- `.d.ts` may be generated or handwritten (low weight)

**Python:**
- `test_*.py` prefix convention (+0.75)
- `*_test.py` suffix convention (+0.70)
- `/tests/` directory reliable (+0.65)

## Explainability Output

Every classification must be explainable:

```json
{
  "path": "internal/auth/login_test.go",
  "role": "test",
  "sub_role": "unit",
  "confidence": 0.92,
  "signals": ["filename", "path"],
  "weights": {
    "test": 1.35,
    "prod": 0.20
  }
}
```

This is non-negotiable for trust and debugging.

## Implementation Checklist

- [ ] Rule tables are static (no regex compilation per file)
- [ ] Path matching uses fast string contains/prefix checks
- [ ] Filename patterns use suffix checks, not regex
- [ ] Neighborhood inference runs as second pass only
- [ ] Header probes are opt-in and cached
- [ ] All classifications include confidence + signals
- [ ] Override parsing happens once at startup
