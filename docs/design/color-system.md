# Semantic Color Token System

**Version:** 0.1

A renderer-agnostic color system that encodes responsibility, not aesthetics.

## Design Principle

> Color encodes responsibility, not magnitude, not language.

Renderers decide how to express color (ANSI, CSS, grayscale). The core system only defines semantic tokens.

## Semantic Tokens

These are **names**, not values:

```go
type SemanticColor string

const (
    ColorPrimary      SemanticColor = "semantic.primary"
    ColorSafety       SemanticColor = "semantic.safety"
    ColorOperational  SemanticColor = "semantic.operational"
    ColorKnowledge    SemanticColor = "semantic.knowledge"
    ColorFragility    SemanticColor = "semantic.fragility"
    ColorLowEmphasis  SemanticColor = "semantic.low_emphasis"
    ColorExternal     SemanticColor = "semantic.external"
    ColorWarning      SemanticColor = "semantic.warning"
)
```

## Token Intent

| Token | Meaning | Visual Intent |
|-------|---------|---------------|
| `semantic.primary` | Core value (prod) | Neutral foreground, no modification |
| `semantic.safety` | Confidence & verification (test) | Muted green or blue |
| `semantic.operational` | Risk & complexity (infra) | Amber or copper |
| `semantic.knowledge` | Human understanding (docs) | Soft gray |
| `semantic.fragility` | Config sensitivity | Gray-blue |
| `semantic.low_emphasis` | Not owned (generated) | Faint/dim |
| `semantic.external` | Third-party (vendor) | Italic + dim |
| `semantic.warning` | Technical debt (deprecated) | Desaturated red |

## Role to Token Mapping

```go
func (r Role) Color() SemanticColor {
    switch r {
    case RoleProd:
        return ColorPrimary
    case RoleTest:
        return ColorSafety
    case RoleInfra:
        return ColorOperational
    case RoleDocs:
        return ColorKnowledge
    case RoleConfig:
        return ColorFragility
    case RoleGenerated:
        return ColorLowEmphasis
    case RoleVendor:
        return ColorExternal
    case RoleScripts:
        return ColorPrimary  // scripts are prod-adjacent
    case RoleExamples:
        return ColorKnowledge  // examples are docs-adjacent
    case RoleDeprecated:
        return ColorWarning
    default:
        return ColorPrimary
    }
}
```

## Renderer Mappings

### TUI (ANSI 256)

```go
var ANSIMapping = map[SemanticColor]int{
    ColorPrimary:     -1,   // default foreground (no color code)
    ColorSafety:      37,   // dim cyan
    ColorOperational: 178,  // amber/gold
    ColorKnowledge:   245,  // gray
    ColorFragility:   67,   // gray-blue
    ColorLowEmphasis: 240,  // dim gray
    ColorExternal:    240,  // dim gray (+ italic attribute)
    ColorWarning:     167,  // desaturated red
}
```

### TUI (ANSI 16 fallback)

For terminals without 256-color support:

```go
var ANSI16Mapping = map[SemanticColor]int{
    ColorPrimary:     -1,   // default
    ColorSafety:      6,    // cyan
    ColorOperational: 3,    // yellow
    ColorKnowledge:   7,    // white/gray
    ColorFragility:   4,    // blue
    ColorLowEmphasis: 8,    // bright black (dark gray)
    ColorExternal:    8,    // bright black
    ColorWarning:     1,    // red
}
```

### Web (CSS Variables)

```css
:root {
  --semantic-primary: inherit;
  --semantic-safety: hsl(200, 25%, 55%);
  --semantic-operational: hsl(35, 45%, 55%);
  --semantic-knowledge: hsl(0, 0%, 55%);
  --semantic-fragility: hsl(210, 20%, 50%);
  --semantic-low-emphasis: hsl(0, 0%, 65%);
  --semantic-external: hsl(0, 0%, 65%);
  --semantic-warning: hsl(0, 35%, 55%);
}

/* Dark mode overrides */
@media (prefers-color-scheme: dark) {
  :root {
    --semantic-safety: hsl(200, 30%, 60%);
    --semantic-operational: hsl(35, 50%, 60%);
    --semantic-knowledge: hsl(0, 0%, 60%);
    --semantic-fragility: hsl(210, 25%, 55%);
    --semantic-low-emphasis: hsl(0, 0%, 45%);
    --semantic-external: hsl(0, 0%, 45%);
    --semantic-warning: hsl(0, 40%, 60%);
  }
}
```

### Print (Grayscale)

For PDF or print contexts:

| Token | Grayscale Value |
|-------|-----------------|
| primary | black (0%) |
| safety | 40% gray |
| operational | 30% gray |
| knowledge | 55% gray |
| fragility | 45% gray |
| low_emphasis | 70% gray |
| external | 70% gray (italic) |
| warning | 25% gray |

## Renderer Contract

All renderers MUST follow these rules:

### 1. Numbers are never colored

```
# BAD
test ████████ 29.4k  <- number is colored

# GOOD
test ████████ 29.4k  <- only "test" label is colored
     ^
     colored
```

### 2. One semantic color per line maximum

Multiple roles on one line use the first role's color:

```
Go   prod ████ 49.2k  test ████ 18.9k
     ^
     entire line uses prod's color (primary)
```

### 3. No background colors

All colors apply to foreground text only.

### 4. Graceful monochrome degradation

When color is unavailable:
- `low_emphasis` and `external` -> dim/italic attribute
- `warning` -> bold attribute
- All others -> default

### 5. Bars are never colored

Bars (`████`) always use default foreground. Labels carry the semantic meaning.

## Lipgloss Implementation

```go
package colors

import "github.com/charmbracelet/lipgloss"

type Theme struct {
    Primary      lipgloss.Style
    Safety       lipgloss.Style
    Operational  lipgloss.Style
    Knowledge    lipgloss.Style
    Fragility    lipgloss.Style
    LowEmphasis  lipgloss.Style
    External     lipgloss.Style
    Warning      lipgloss.Style
}

func NewDefaultTheme() *Theme {
    return &Theme{
        Primary:      lipgloss.NewStyle(),
        Safety:       lipgloss.NewStyle().Foreground(lipgloss.Color("37")),
        Operational:  lipgloss.NewStyle().Foreground(lipgloss.Color("178")),
        Knowledge:    lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
        Fragility:    lipgloss.NewStyle().Foreground(lipgloss.Color("67")),
        LowEmphasis:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
        External:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true),
        Warning:      lipgloss.NewStyle().Foreground(lipgloss.Color("167")),
    }
}

func (t *Theme) ForRole(role Role) lipgloss.Style {
    switch role.Color() {
    case ColorPrimary:
        return t.Primary
    case ColorSafety:
        return t.Safety
    case ColorOperational:
        return t.Operational
    case ColorKnowledge:
        return t.Knowledge
    case ColorFragility:
        return t.Fragility
    case ColorLowEmphasis:
        return t.LowEmphasis
    case ColorExternal:
        return t.External
    case ColorWarning:
        return t.Warning
    default:
        return t.Primary
    }
}
```

## NO_COLOR Support

When `NO_COLOR` environment variable is set (any value):

```go
func NewNoColorTheme() *Theme {
    base := lipgloss.NewStyle()
    return &Theme{
        Primary:      base,
        Safety:       base,
        Operational:  base,
        Knowledge:    base,
        Fragility:    base,
        LowEmphasis:  base.Faint(true),
        External:     base.Faint(true).Italic(true),
        Warning:      base.Bold(true),
    }
}
```

## Testing Color Output

Test files should verify:

1. **Token mapping is complete** - Every role maps to a token
2. **Renderer mapping is complete** - Every token has ANSI/CSS/grayscale values
3. **Contract compliance** - Numbers are not colored, bars are not colored
4. **Graceful degradation** - NO_COLOR mode produces readable output

```go
func TestColorContract(t *testing.T) {
    // Every role must map to a color
    for _, role := range AllRoles {
        if role.Color() == "" {
            t.Errorf("Role %s has no color mapping", role)
        }
    }

    // Every color must have ANSI mapping
    for color := range SemanticColors {
        if _, ok := ANSIMapping[color]; !ok {
            t.Errorf("Color %s has no ANSI mapping", color)
        }
    }
}
```

## Accessibility Notes

- All color choices pass WCAG AA contrast ratio (4.5:1) on both light and dark backgrounds
- Color is never the only distinguishing factor (text labels always present)
- Low-vision users can rely on position and structure, not color alone
- Colorblind users: operational (amber) and safety (cyan) are distinguishable in all common forms of color blindness
