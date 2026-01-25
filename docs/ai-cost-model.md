# AI Cost Model

aloc estimates development effort using multiple delivery models, reflecting that software cost depends as much on *how* teams work as on how much code exists.

## Usage

```bash
aloc . --effort                     # Include effort estimates (uses faang profile)
aloc . --effort --profile faang     # Explicit profile selection
aloc . --effort --ai-model opus     # Use Opus for AI cost calculations
aloc . --effort --human-cost 20000  # $20k/month per engineer
aloc . --effort --model-config x.json  # Custom config file (overrides profile)
```

## Delivery Models

### Market Replacement (Conventional Team)

Answers: *"What would it cost to replace this codebase using a typical engineering organization?"*

**Assumptions:**
- Conventional team structure with mixed skill distribution
- Sequential work with coordination overhead
- Partial AI usage (autocomplete, chat assistance)
- Productivity variance: 0.85x (optimistic) to 1.30x (pessimistic)

**Based on COCOMO Basic Organic (calibrated for modern productivity):**

```
Effort (person-months) = a × (KLOC)^1.05   # a=0.43 for faang profile
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

## Profiles

aloc ships with pre-calibrated effort profiles. The default profile (`faang`) is calibrated for modern, high-performing engineering teams.

### FAANG Profile (Default)

Calibrated to **80 LOC/day** (1,760 LOC/month) based on Smacchia's 14-year productivity study. This reflects sustained net-additive LOC after deletions and rework, while maintaining high code quality.

| Model | 1981 COCOMO | FAANG Profile | Productivity |
|-------|-------------|---------------|--------------|
| organic | 2.4 | 0.43 | ~80 LOC/day |
| semi-detached | 3.0 | 0.53 | ~80 LOC/day |
| embedded | 3.6 | 0.64 | ~80 LOC/day |

**Example: 122K LOC codebase**

| Metric | 1981 COCOMO | FAANG Profile |
|--------|-------------|---------------|
| Effort | 367 PM | ~66 PM |
| Schedule | ~25 months | ~12-14 months |
| Team | ~15 engineers | ~5-6 engineers |

See [Engineering Cost Methodology](./engineering-cost-methodology.md) for detailed derivation and research references.

## Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `--effort` | true | Enable effort estimation |
| `--profile` | faang | Effort estimation profile |
| `--ai-model` | sonnet | Model for AI cost calculation |
| `--human-cost` | 0 | Monthly cost per engineer (0 = use blended cost) |
| `--model-config` | - | Path to JSON file with model parameter overrides |

### Custom Model Configuration

Override any effort model parameter via JSON file:

```bash
aloc . --effort --model-config ./model.json
```

All parameters are optional; unspecified values use defaults. Example:

```json
{
  "cocomo_models": {
    "organic": { "a": 2.4, "b": 1.05, "c": 2.5, "d": 0.38 }
  },
  "variance_multipliers": {
    "optimistic": 0.85,
    "pessimistic": 1.30
  },
  "ai_native": {
    "schedule_factor_low": 0.40,
    "schedule_factor_high": 0.60,
    "team_size_factor_low": 0.33,
    "team_size_factor_high": 0.50,
    "minimum_team_size": 2,
    "tooling_monthly_low": 2000,
    "tooling_monthly_high": 10000,
    "large_codebase_threshold_kloc": 100,
    "tooling_monthly_low_large": 5000,
    "tooling_monthly_high_large": 20000
  },
  "token_estimation": {
    "avg_loc_per_file": 160,
    "iterations_per_file": 10,
    "context_per_call": 20000,
    "output_per_call": 2000,
    "chars_per_token": 4.0
  },
  "skill_bands": {
    "junior": { "annual_cost_low": 130000, "annual_cost_high": 200000 },
    "senior": { "annual_cost_low": 180000, "annual_cost_high": 350000 },
    "staff": { "annual_cost_low": 280000, "annual_cost_high": 550000 },
    "principal": { "annual_cost_low": 450000, "annual_cost_high": 1100000 },
    "senior_principal": { "annual_cost_low": 1000000, "annual_cost_high": 2200000 },
    "distinguished": { "annual_cost_low": 2500000, "annual_cost_high": 5000000 }
  },
  "team_composition": {
    "junior": 0.20,
    "senior": 0.45,
    "staff": 0.22,
    "principal": 0.10,
    "senior_principal": 0.03,
    "distinguished": 0.00
  },
  "team_composition_by_size": {
    "small": { "junior": 0.10, "senior": 0.60, "staff": 0.25, "principal": 0.05 },
    "medium": { "junior": 0.15, "senior": 0.50, "staff": 0.25, "principal": 0.08, "senior_principal": 0.02 },
    "large": { "junior": 0.20, "senior": 0.45, "staff": 0.22, "principal": 0.10, "senior_principal": 0.03 },
    "enterprise": { "junior": 0.25, "senior": 0.40, "staff": 0.20, "principal": 0.10, "senior_principal": 0.04, "distinguished": 0.01 },
    "mega": { "junior": 0.25, "senior": 0.40, "staff": 0.18, "principal": 0.10, "senior_principal": 0.05, "distinguished": 0.02 }
  },
  "ai_leverage_by_skill": {
    "junior": 1.5,
    "senior": 2.5,
    "staff": 4.0,
    "principal": 6.0,
    "senior_principal": 8.0,
    "distinguished": 10.0
  },
  "elite_reference": {
    "observed_months": 1.5,
    "observed_ai_spend": 6000
  },
  "default_human_cost_per_month": 15000
}
```

**Key parameters:**

| Category | Parameter | Default | Description |
|----------|-----------|---------|-------------|
| COCOMO | `a`, `b` | 2.4, 1.05 | Effort coefficients (organic) |
| COCOMO | `c`, `d` | 2.5, 0.38 | Schedule coefficients |
| Variance | `optimistic` | 0.85 | Low-end effort multiplier |
| Variance | `pessimistic` | 1.30 | High-end effort multiplier |
| AI-Native | `schedule_factor_*` | 0.40-0.60 | Schedule compression vs conventional |
| AI-Native | `team_size_factor_*` | 0.33-0.50 | Team size reduction vs conventional |
| AI-Native | `tooling_monthly_*` | $2K-$10K | Monthly AI tooling costs |
| Tokens | `iterations_per_file` | 10 | API calls per file (implement/test/fix) |
| Skill Bands | `annual_cost_*` | varies | Fully-loaded annual cost range (USD) |
| Team Comp | `team_composition` | see below | Default mix of engineering levels |
| Team Comp | `team_composition_by_size` | see below | Size-based composition (small/medium/large/enterprise/mega) |

### Team Composition

Configure the mix of engineering levels to calculate blended costs:

| Size Tier | Engineers | Junior | Senior | Staff | Principal | SPE | Distinguished |
|-----------|-----------|--------|--------|-------|-----------|-----|---------------|
| Small | 5-20 | 10% | 60% | 25% | 5% | — | — |
| Medium | 50-100 | 15% | 50% | 25% | 8% | 2% | — |
| Large | 200-500 | 20% | 45% | 22% | 10% | 3% | — |
| Enterprise | 1000+ | 25% | 40% | 20% | 10% | 4% | 1% |
| Mega | 2000+ | 25% | 40% | 18% | 10% | 5% | 2% |

### AI Leverage by Skill Level

Principal+ engineers get significantly more leverage from AI because they can effectively orchestrate agents:

| Band | Leverage | Rationale |
|------|----------|-----------|
| Junior | 1.5× | Autocomplete gains, but struggles to validate AI output |
| Senior | 2.5× | Effective with copilot-style tools, can review AI code |
| Staff | 4.0× | Can orchestrate agent workflows, designs for AI implementation |
| Principal | 6.0× | Effective multi-agent orchestration, rapid validation |
| Senior Principal | 8.0× | Designs systems for agent swarms, maximizes parallelization |
| Distinguished | 10.0× | Organization-scale AI leverage, defines patterns others follow |

**Research basis:**
- GitHub Copilot study (2022): 55% faster task completion
- MIT study (2023): 40% productivity boost with ChatGPT
- McKinsey (2023): Up to 3× gains for some software tasks
- Orchestration leverage for senior engineers is largely unmeasured but observed to be substantial

**Example:** A team of 2 principals (6× leverage each) has effective capacity equivalent to 12 conventional engineers.

For detailed methodology, research references, and customization guidance, see [Engineering Cost Methodology](./engineering-cost-methodology.md).

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
