# AI Cost Model

aloc estimates development effort using multiple delivery models, reflecting that software cost depends as much on *how* teams work as on how much code exists.

## Usage

```bash
aloc . --effort                    # Include effort estimates
aloc . --effort --ai-model opus    # Use Opus for AI cost calculations
aloc . --effort --human-cost 20000 # $20k/month per engineer
```

## Delivery Models

### Market Replacement (Conventional Team)

Answers: *"What would it cost to replace this codebase using a typical engineering organization?"*

**Assumptions:**
- Conventional team structure with mixed skill distribution
- Sequential work with coordination overhead
- Partial AI usage (autocomplete, chat assistance)
- Productivity variance: 0.85x (optimistic) to 1.30x (pessimistic)

**Based on COCOMO Basic Organic:**

```
Effort (person-months) = 2.4 × (KLOC)^1.05
Schedule (months) = 2.5 × (Effort)^0.38
Team Size = Effort / Schedule
Cost = Effort × cost_per_month
```

Output is presented as a range to reflect uncertainty:

```
Market Replacement Estimate (Conventional Team)
$343K – $525K · 8–10 months · ~3–4 engineers
```

### AI-Native Team (Agentic/Parallel)

Reflects AI-assisted parallel execution under human oversight:

- Work decomposed into tasks executed in parallel
- Multiple AI agents draft code, tests, and documentation
- Humans supervise, review, integrate, and make final decisions
- Coordination overhead reduced but not eliminated

**Multipliers applied to conventional estimates:**
- Schedule: 40–60% of conventional time
- Team size: 33–50% of conventional (minimum 2)
- AI tooling: $2K–$10K/month (or $5K–$20K for large codebases >100K LOC)

```
AI-Native Team Estimate (Agentic/Parallel)
$105K – $215K · 3–6 months · ~2–2 engineers
(AI tooling: $2K–$10K/mo, total $7K–$58K)
```

**This model does not assume replacement of human judgment.** It reflects what's achievable with effective task decomposition and AI-assisted workflows.

## Token Estimation (AI-Only Cost)

For reference, aloc also calculates raw AI token costs:

| Parameter | Default | Description |
|-----------|---------|-------------|
| Output tokens per LOC | 1.0 | ~1 token per line of code |
| Input/Output ratio | 4:1 | Accounts for prompts, context, reasoning |

### Pricing (per million tokens)

| Model | Input | Output | Best For |
|-------|-------|--------|----------|
| Claude Sonnet | $3.00 | $15.00 | Balanced performance/cost |
| Claude Opus | $15.00 | $75.00 | Maximum capability |
| Claude Haiku | $0.25 | $1.25 | Speed and cost efficiency |

*Pricing as of late 2025. Check [Anthropic's pricing](https://www.anthropic.com/pricing) for current rates.*

## Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `--effort` | false | Enable effort estimation |
| `--ai-model` | sonnet | Model for AI cost calculation |
| `--human-cost` | 15000 | Monthly cost per engineer (USD) |

## Assumptions and Limitations

1. **COCOMO is for greenfield development** — Maintenance and refactoring have different economics
2. **Token estimation is approximate** — Actual counts depend on code complexity and formatting
3. **AI-Native estimates assume effective workflows** — Results vary by team capability
4. **Ranges reflect uncertainty** — Point estimates would be false precision
5. **Pricing may change** — Verify current rates for accurate estimates

## Example Output

```
Development Effort Models
────────────────────────────────────────────────────────────────────────────────
Market Replacement Estimate (Conventional Team)
  $343K – $525K · 8–10 months · ~3–4 engineers

AI-Native Team Estimate (Agentic/Parallel)
  $105K – $215K · 3–6 months · ~2–2 engineers
  (AI tooling: $2K–$10K/mo, total $7K–$58K)
```
