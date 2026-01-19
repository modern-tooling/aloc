# JSON Schema Specification

**Version:** 1.0

The JSON output is the source of truth. Renderers are pure views that consume this schema.

## Design Principles

- No renderer assumptions
- Ratios are first-class (not recomputed by renderers)
- Confidence is explicit (Tufte honesty)
- Trend data is pre-digested (no charts computed at render time)
- Semantic roles are a closed vocabulary

## Top-Level Schema

```json
{
  "meta": {
    "schema_version": "1.0",
    "generated_at": "2026-01-15T23:10:00Z",
    "generator": "lloc",
    "generator_version": "0.1.0",
    "repo": {
      "name": "example-repo",
      "commit": "abc123def456",
      "branch": "main",
      "root": "/path/to/repo"
    }
  },

  "summary": {
    "files": 493,
    "loc_total": 126433,
    "languages": 7
  },

  "responsibilities": [
    {
      "role": "core",
      "loc": 62140,
      "files": 214,
      "confidence": 0.94
    },
    {
      "role": "test",
      "loc": 29420,
      "files": 161,
      "confidence": 0.91,
      "breakdown": {
        "unit": 0.81,
        "integration": 0.11,
        "e2e": 0.08
      }
    },
    {
      "role": "infra",
      "loc": 14210,
      "files": 62,
      "confidence": 0.88,
      "notes": ["42% CI/CD definitions"]
    },
    {
      "role": "docs",
      "loc": 11530,
      "files": 38,
      "confidence": 0.92
    },
    {
      "role": "config",
      "loc": 6120,
      "files": 15,
      "confidence": 0.85
    },
    {
      "role": "generated",
      "loc": 2713,
      "files": 3,
      "confidence": 0.97
    }
  ],

  "ratios": {
    "test_to_core": 0.47,
    "infra_to_core": 0.23,
    "docs_to_core": 0.18,
    "generated_to_core": 0.04,
    "config_to_core": 0.10
  },

  "languages": [
    {
      "language": "Go",
      "loc_total": 68130,
      "files": 312,
      "responsibilities": {
        "core": 49210,
        "test": 18920
      }
    },
    {
      "language": "MDX",
      "loc_total": 11120,
      "files": 34,
      "responsibilities": {
        "docs": 11120
      }
    },
    {
      "language": "Shell",
      "loc_total": 9400,
      "files": 28,
      "responsibilities": {
        "infra": 8200,
        "scripts": 1200
      }
    },
    {
      "language": "YAML",
      "loc_total": 6800,
      "files": 45,
      "responsibilities": {
        "infra": 4800,
        "config": 2000
      }
    }
  ],

  "trend": {
    "test_to_core": {
      "window": "12m",
      "sparkline": [0.32, 0.35, 0.38, 0.41, 0.44, 0.46, 0.47],
      "direction": "up",
      "interpretation": "upward since Q2"
    }
  },

  "confidence": {
    "auto_classified": 0.92,
    "heuristic": 0.08,
    "override": 0.00
  },

  "files": [
    {
      "path": "internal/auth/login.go",
      "loc": 412,
      "language": "Go",
      "role": "core",
      "confidence": 0.94,
      "signals": ["path", "extension"]
    },
    {
      "path": "internal/auth/login_test.go",
      "loc": 688,
      "language": "Go",
      "role": "test",
      "sub_role": "unit",
      "confidence": 0.98,
      "signals": ["filename", "path"]
    }
  ]
}
```

## Field Definitions

### meta

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `schema_version` | string | yes | Schema version (semver) |
| `generated_at` | string | yes | ISO 8601 timestamp |
| `generator` | string | yes | Tool name ("lloc") |
| `generator_version` | string | yes | Tool version |
| `repo.name` | string | no | Repository name |
| `repo.commit` | string | no | Current commit SHA |
| `repo.branch` | string | no | Current branch |
| `repo.root` | string | no | Absolute path to repo root |

### summary

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `files` | integer | yes | Total files analyzed |
| `loc_total` | integer | yes | Total lines of code |
| `languages` | integer | yes | Count of distinct languages |

### responsibilities[]

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `role` | string | yes | Semantic role (see vocabulary) |
| `loc` | integer | yes | Lines of code in this role |
| `files` | integer | yes | Files in this role |
| `confidence` | float | yes | Aggregate confidence (0.0-1.0) |
| `breakdown` | object | no | Sub-role percentages (test only) |
| `notes` | string[] | no | Human-readable annotations |

### ratios

All ratios are pre-calculated as `role_loc / core_loc`.

| Field | Type | Description |
|-------|------|-------------|
| `test_to_core` | float | Test LOC / Core LOC |
| `infra_to_core` | float | Infra LOC / Core LOC |
| `docs_to_core` | float | Docs LOC / Core LOC |
| `generated_to_core` | float | Generated LOC / Core LOC |
| `config_to_core` | float | Config LOC / Core LOC |

### languages[]

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `language` | string | yes | Language name |
| `loc_total` | integer | yes | Total LOC for this language |
| `files` | integer | yes | File count |
| `responsibilities` | object | yes | LOC breakdown by role |

### trend

Optional. Present only if historical data exists.

| Field | Type | Description |
|-------|------|-------------|
| `window` | string | Time window ("12m", "6m", "30d") |
| `sparkline` | float[] | Data points for visualization |
| `direction` | string | "up", "down", "stable" |
| `interpretation` | string | Human-readable trend description |

### confidence

| Field | Type | Description |
|-------|------|-------------|
| `auto_classified` | float | Percentage classified by rules |
| `heuristic` | float | Percentage requiring heuristic fallback |
| `override` | float | Percentage from user overrides |

### files[]

Optional detailed file list. May be omitted in summary mode.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `path` | string | yes | Relative path from repo root |
| `loc` | integer | yes | Lines of code |
| `language` | string | yes | Detected language |
| `role` | string | yes | Assigned role |
| `sub_role` | string | no | Test sub-role (unit/integration/e2e/contract/fixture) |
| `confidence` | float | yes | Classification confidence |
| `signals` | string[] | yes | Signals that contributed to classification |

## Semantic Role Vocabulary

Closed enumeration. Renderers should treat these as known values.

**Primary roles:**
- `core` - Core application logic
- `test` - Tests (unit, integration, e2e)
- `infra` - Infrastructure & deployment
- `docs` - Human-facing documentation
- `config` - Runtime/build configuration
- `generated` - Auto-generated code
- `vendor` - Third-party code
- `scripts` - One-off tooling/glue
- `examples` - Samples, demos
- `deprecated` - Technical debt

**Test sub-roles:**
- `unit` - Fast, isolated
- `integration` - Cross-boundary
- `e2e` - System-level
- `contract` - API/interface guarantees
- `fixture` - Data, mocks, golden files

## Signal Vocabulary

Closed enumeration. Used in explainability output.

- `path` - Directory path matched a rule
- `filename` - Filename pattern matched
- `extension` - Extension bias applied
- `neighborhood` - Sibling files influenced classification
- `header` - Header probe detected marker
- `override` - User override applied

## Renderer Contract

Renderers consuming this schema must follow these rules:

1. **Numbers are never colored** - Semantic color applies only to labels
2. **One semantic color per line maximum**
3. **No legends required** - Layout should be self-explanatory
4. **Color must degrade gracefully** - Monochrome must remain readable
5. **Ratios are displayed as-is** - No re-computation
6. **Confidence is always visible** - Never hidden from users

## Diff Schema Extension

When `--diff` flag is used, add:

```json
{
  "diff": {
    "base_commit": "abc123",
    "head_commit": "def456",
    "responsibilities_delta": {
      "core": { "loc": +1200, "pct": +1.9 },
      "test": { "loc": +800, "pct": +2.7 },
      "infra": { "loc": +2100, "pct": +17.3 }
    },
    "ratios_delta": {
      "test_to_core": +0.02,
      "infra_to_core": +0.03
    },
    "highlights": [
      "infra grew 17% (CI/CD additions)",
      "test coverage improved slightly"
    ]
  }
}
```
