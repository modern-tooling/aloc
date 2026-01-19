# Terminal Wireframe Specification

**Version:** 0.1

A Tufte-inspired terminal output spec: structure, visual cues, semantic color meaning, and why each element earns its ink.

## Global Design Rules

### Typography & Layout
- Monospaced terminal font assumed
- Fixed-width columns
- Unicode box drawing allowed (light weight only)
- Left-aligned text; numbers right-aligned
- Whitespace is primary structure (no heavy borders)
- Target width: 80-100 columns

### Color Rules
- Color applies only to labels, never to numeric values
- One semantic color per line maximum
- No background colors
- No legends required
- Must degrade gracefully to monochrome

### Icon Rules
- Optional, ASCII-safe only
- At most one symbol per line
- Only when meaning is non-obvious

## Section Specifications

### 1. Header Block (Context & Scale)

```
Codebase Summary
────────────────────────────────────────────────────────────
Files: 493          LOC: 126,433          Languages: 7
```

**Rules:**
- Exactly 3 metrics
- Fixed order: Files -> LOC -> Languages
- No color
- Section never changes height
- Numbers formatted with commas at thousands

### 2. Responsibility Ledger (Primary Block)

```
Responsibility Breakdown (LOC)
────────────────────────────────────────────────────────────
core        ████████████████████████ 62.1k
test        ████████████             29.4k
infra       ██████                   14.2k
docs        ████                     11.5k
config      ██                        6.1k
generated   █                         2.7k
```

**Structure:**
- Left column: semantic role label (12 chars fixed width)
- Middle: bar (relative scale within section, max 24 chars)
- Right: absolute value with 'k' suffix for thousands

**Bar Rules:**
- Bars are monochrome (█ character)
- Bars encode relative magnitude only
- Bar length = (role_loc / max_role_loc) * 24
- Bars do not encode semantics (labels do)

**Semantic Color:**
Apply color to role label only:

| Role | Color Intent |
|------|--------------|
| core | default foreground |
| test | muted green/blue (ANSI 37/36) |
| infra | amber/copper (ANSI 178) |
| docs | soft gray (ANSI 245) |
| config | gray-blue (ANSI 67) |
| generated | faint gray (ANSI 240) |
| vendor | dim italic (ANSI 240) |
| deprecated | desaturated red (ANSI 167) |

### 3. Key Ratios Block

```
Key Ratios
────────────────────────────────────────────────────────────
Test / Core           0.47    healthy baseline = 0.5-0.8
Infra / Core          0.23    operational complexity
Docs / Core           0.18    well documented
Generated / Core      0.04    low automation reliance
```

**Rules:**
- Ratios sorted by importance, not magnitude
- Values printed with two decimal places
- Commentary is plain text, no color
- No bars, no charts
- This section exists to replace mental math

**Order of ratios:**
1. Test / Core (most important)
2. Infra / Core
3. Docs / Core
4. Generated / Core
5. Config / Core (only if notable)

### 4. Language Composition (Small Multiples)

```
Language Composition
────────────────────────────────────────────────────────────
Go        core ████████████████ 49.2k   test ████████ 18.9k
MDX       docs ████████████     11.1k
Shell     infra ████████         9.4k
YAML      infra ██████           6.8k
```

**Rules:**
- One line per language
- Only show dominant responsibilities (>=10% of language total)
- Never show zero or trivial categories
- Semantic color applies to responsibility label only
- Language name in default foreground
- Skip languages with <1% total LOC

**Layout:**
```
<lang:10> <role1:8> <bar:16> <val:6>   [<role2:8> <bar:8> <val:6>]
```

Second responsibility is optional, shown only if significant.

### 5. Marginal Notes (Optional, Sparse)

```
infra ██████ 14.2k
        ↑ 42% CI/CD definitions
```

**Rules:**
- Indented under the relevant line
- One note max per item
- No color
- No repetition of numbers
- Use ↑ or → as pointer characters

**When to show notes:**
- Test breakdown (unit/integration/e2e split)
- Infra breakdown (CI/CD vs terraform vs docker)
- Notable concentration in one language

### 6. Temporal Cue (Optional Sparkline)

```
Test / Core Trend (12 months)
▁▂▃▄▆▆▇▇▆▆▇▇   upward since Q2
```

**Rules:**
- Only rendered if historical data exists
- Single sparkline per report
- No axes, no labels
- One textual interpretation allowed (after sparkline)
- Use block elements: ▁▂▃▄▅▆▇█

**Sparkline characters mapping:**
| Range | Character |
|-------|-----------|
| 0-12% | ▁ |
| 12-25% | ▂ |
| 25-37% | ▃ |
| 37-50% | ▄ |
| 50-62% | ▅ |
| 62-75% | ▆ |
| 75-87% | ▇ |
| 87-100% | █ |

### 7. Directory/File Drilldown (Scoped View)

```
auth/
────────────────────────────────────────────────────────────
login.go              412   core
login_test.go         688   test
jwt.go                931   core   ~ hot path
mock_identity.go      201   test:fixture
```

**Rules:**
- Sorted by semantic role, then size descending
- Size always right-aligned (6 chars)
- Optional symbol only when noteworthy
- Semantic color only on role label
- Directory header shows path relative to repo root

**Symbols (ASCII-safe):**
| Symbol | Meaning |
|--------|---------|
| `~` | Hot path (high churn) |
| `!` | Warning (low confidence) |
| `*` | Recently changed |

### 8. Confidence Disclosure (Footer)

```
Classification confidence: high (92% auto, 8% heuristic)
```

**Rules:**
- Plain text
- No color
- Always last section
- Show only if confidence varies

**Confidence thresholds:**
| Overall | Label |
|---------|-------|
| >= 90% | high |
| 70-89% | medium |
| < 70% | low |

## Forbidden Elements

These must NEVER appear in the output:

- Pie charts
- Stacked bars
- Legends
- Background colors
- Emojis
- Percentages without absolute values
- Color-only meaning (always pair with text)
- Repeated totals across sections
- More than one sparkline

## ASCII Fallback Mode

When `NO_COLOR` is set or terminal doesn't support Unicode:

```
Codebase Summary
----------------------------------------------------------------
Files: 493          LOC: 126,433          Languages: 7

Responsibility Breakdown (LOC)
----------------------------------------------------------------
core        ######################## 62.1k
test        ############             29.4k
infra       ######                   14.2k
docs        ####                     11.5k
config      ##                        6.1k
generated   #                         2.7k
```

**Fallback rules:**
- Use `#` instead of `█`
- Use `-` instead of `─`
- All color tokens become default foreground
- Sparklines use `.o*` characters instead of blocks

## Compact Mode

For CI logs or narrow terminals (60 cols):

```
lloc summary: 493 files, 126k LOC
  core:62.1k test:29.4k infra:14.2k docs:11.5k
  test/core:0.47 infra/core:0.23
  confidence:92%
```

**Compact rules:**
- Single-line summary
- No bars
- No marginal notes
- No sparklines
- Minimal whitespace
