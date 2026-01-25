# Engineering Cost Methodology

This document explains how aloc calculates engineering costs, the research behind the models, and how to customize parameters for your organization.

## Overview

Engineering cost estimation combines three models:

1. **COCOMO** - Effort and schedule estimation based on code size
2. **Team Composition** - Weighted costs based on skill level distribution
3. **AI-Native Adjustments** - Efficiency gains from AI-assisted development

## COCOMO Model

### Background

The Constructive Cost Model (COCOMO) was developed by Barry Boehm at USC in 1981, based on analysis of 63 software projects. It remains one of the most validated and widely-used software cost estimation models.

**Key Reference:**
> Boehm, B. W. (1981). *Software Engineering Economics*. Prentice-Hall.

### Formula

```
Effort (person-months) = a × KLOC^b
Schedule (months)      = c × Effort^d
Team Size              = Effort / Schedule
Cost                   = Effort × cost_per_month
```

### Coefficients by Project Type

| Model | a | b | c | d | Use Case |
|-------|---|---|---|---|----------|
| Organic | 2.4 | 1.05 | 2.5 | 0.38 | Small teams, familiar tech, flexible requirements |
| Semi-detached | 3.0 | 1.12 | 2.5 | 0.35 | Medium projects, mixed experience |
| Embedded | 3.6 | 1.20 | 2.5 | 0.32 | Tight constraints, hardware integration |

**Default:** Organic (most applicable to modern software development)

### FAANG Profile: Calibrated Coefficients

The original COCOMO coefficients (1981) reflect 1970s-era development productivity. Modern development with frameworks, OSS libraries, package managers, better IDEs, and CI/CD achieves significantly higher throughput.

**Calibration Target: 80 LOC/day (1,760 LOC/month)**

This target derives from Smacchia's long-term productivity data [11], representing sustained net-additive LOC after deletions and rework, while maintaining high code quality and test coverage.

| Model | 1981 Coefficient | FAANG Coefficient | Implied LOC/month |
|-------|------------------|-------------------|-------------------|
| Organic | 2.4 | **0.43** | ~1,760 |
| Semi-detached | 3.0 | **0.53** | ~1,760 |
| Embedded | 3.6 | **0.64** | ~1,760 |

**Derivation:**

The COCOMO effort formula `Effort = a × KLOC^b` implies productivity inversely proportional to 'a':

```
Original COCOMO (a=2.4): ~312 LOC/month
Target productivity:      1,760 LOC/month (80 LOC/day × 22 days)

a_new = 312 × 2.4 / 1,760 = 0.43
```

**Example: 122K LOC codebase**

| Metric | 1981 COCOMO | FAANG Profile |
|--------|-------------|---------------|
| Effort | 367 PM | ~66 PM |
| Schedule | ~25 months | ~13-14 months |
| Team | ~15 engineers | ~5 engineers |

This implies FAANG-tier teams with modern tooling are **~5-6× more productive** than the 1981 baseline. This is consistent with:
- 40+ years of tooling improvements (IDEs, debuggers, version control)
- Framework and library ecosystem maturity
- CI/CD reducing integration overhead
- Code generation and autocomplete tools

### Non-Linear Scaling

The model captures two critical insights about software development:

1. **Effort scales superlinearly with size** (b > 1.0)
   - 10× more code requires ~11× more effort, not 10×
   - Additional effort accounts for integration complexity, communication overhead, and defect rates

2. **Schedule scales sublinearly with effort** (d < 1.0)
   - Doubling effort only increases schedule by ~30%
   - Reflects Brooks's Law: adding people to a late project makes it later
   - You cannot infinitely compress schedule by adding engineers

### Example Calculations

| KLOC | Effort (PM) | Schedule (mo) | Team Size |
|------|-------------|---------------|-----------|
| 10 | 27 | 8.0 | 3.4 |
| 50 | 147 | 11.0 | 13.4 |
| 100 | 302 | 12.7 | 23.8 |
| 500 | 1,652 | 18.1 | 91.3 |

### Variance Multipliers

Real-world projects vary significantly from the baseline. The model applies multipliers to generate ranges:

| Multiplier | Value | Reflects |
|------------|-------|----------|
| Optimistic | 0.85× | Experienced team, clear requirements, good tooling, low technical debt |
| Pessimistic | 1.30× | Junior team, unclear scope, legacy constraints, high rework rate |

**Research basis:** These values align with COCOMO II's effort multipliers and observed variance in empirical studies.

> Boehm, B. W., et al. (2000). *Software Cost Estimation with COCOMO II*. Prentice-Hall.

## Team Composition Model

### Research on IC Ladder Distribution

Engineering organizations follow predictable patterns in skill level distribution as they scale. Research sources include:

1. **Radford/Aon Compensation Surveys** - Industry benchmarks on level distribution across thousands of companies
2. **Levels.fyi** - Crowdsourced compensation data with level distribution by company size
3. **Will Larson, "Staff Engineer" (2021)** - Documents patterns like "1 Staff per 8-10 engineers"
4. **Charity Majors, "The Engineer/Manager Pendulum"** - Discusses level compression in smaller organizations

### Default Composition by Organization Size

| Team Size | Junior | Senior | Staff | Principal | SPE | Distinguished |
|-----------|--------|--------|-------|-----------|-----|---------------|
| 5-20 (Small) | 10% | 60% | 25% | 5% | — | — |
| 50-100 (Medium) | 15% | 50% | 25% | 8% | 2% | — |
| 200-500 (Large) | 20% | 45% | 22% | 10% | 3% | — |
| 1000+ (Enterprise) | 25% | 40% | 20% | 10% | 4% | 1% |
| 2000+ (Mega) | 25% | 40% | 18% | 10% | 5% | 2% |

### Rule of Thumb Ratios

These ratios emerge from industry data:

| Level | Ratio | Notes |
|-------|-------|-------|
| Staff | 1 per 5-10 seniors | Technical leadership for teams |
| Principal | 1 per 20-40 ICs | Cross-team technical direction |
| Senior Principal | 1 per 100-200 ICs | Org-level architecture |
| Distinguished/Fellow | 1 per 500-2000 ICs | Company-wide technical vision |

**Note:** These ratios vary significantly by company culture. Google, Meta, and Netflix have very different IC ladder philosophies.

### Skill Band Compensation (2025 US Fully-Loaded Costs)

"Fully-loaded" includes salary, benefits, equity, payroll taxes, recruiting, facilities, equipment, and management overhead.

| Band | Level Equiv. | Annual Low | Annual High | Monthly Low | Monthly High |
|------|--------------|------------|-------------|-------------|--------------|
| Junior | L3/SDE1 (0-2 yrs) | $130,000 | $200,000 | $10,833 | $16,667 |
| Senior | L4-5/SDE2-3 (3-7 yrs) | $180,000 | $350,000 | $15,000 | $29,167 |
| Staff | L6/SDE4 (7+ yrs) | $280,000 | $550,000 | $23,333 | $45,833 |
| Principal | L7 | $450,000 | $1,100,000 | $37,500 | $91,667 |
| Senior Principal | L8 | $1,000,000 | $2,200,000 | $83,333 | $183,333 |
| Distinguished | L9+/Fellow | $2,500,000 | $5,000,000 | $208,333 | $416,667 |

**Sources:**
- Levels.fyi compensation data (2024-2025)
- Radford Global Technology Survey
- Company-specific data from public filings and verified submissions

**Note:** Ranges overlap intentionally to reflect variance by location (SF vs Austin), company stage (FAANG vs Series B), and domain (ML vs web).

### Blended Cost Calculation

Rather than using a single "cost per engineer," we calculate weighted average based on team composition:

```
blended_monthly_cost = Σ(ratio_i × band_monthly_cost_i)
```

**Example for a 100-person engineering org (Medium composition):**

| Band | Ratio | Monthly Cost (mid) | Weighted |
|------|-------|-------------------|----------|
| Junior | 15% | $13,750 | $2,063 |
| Senior | 50% | $22,083 | $11,042 |
| Staff | 25% | $34,583 | $8,646 |
| Principal | 8% | $64,583 | $5,167 |
| SPE | 2% | $133,333 | $2,667 |
| **Total** | | | **$29,585** |

This is significantly higher than the default $15,000/month because it accounts for the full cost distribution including senior ICs.

## AI-Native Team Model

### Adjustments Applied

The AI-native model modifies conventional estimates based on observed gains from AI-assisted development:

| Parameter | Conventional | AI-Native | Rationale |
|-----------|-------------|-----------|-----------|
| Schedule | 100% | 40-60% | Parallel task execution via AI agents |
| Team Size | 100% | 33-50% | Reduced drafting time, automated testing |
| Minimum Team | 1 | 2 | Human review remains essential |
| AI Tooling | $0 | $2K-$20K/mo | API costs, tooling subscriptions |

### What Changes vs. What Doesn't

**Reduced by AI assistance:**
- Initial code drafting time
- Test generation and maintenance
- Documentation writing
- Boilerplate and repetitive tasks
- Context switching overhead

**Not significantly changed:**
- Architecture decisions
- Code review and integration
- Debugging complex issues
- Requirements gathering
- Stakeholder communication

### Tooling Cost Assumptions

| Codebase Size | Monthly AI Tooling |
|---------------|-------------------|
| < 100K LOC | $2,000 - $10,000 |
| > 100K LOC | $5,000 - $20,000 |

Based on observed API costs for teams using Claude, GPT-4, and similar models intensively.

### AI Leverage by Skill Level

Not all engineers get equal leverage from AI assistance. Principal+ engineers are significantly more effective because they can:

1. **Decompose problems** into well-scoped agent tasks
2. **Recognize incorrect output** quickly based on experience
3. **Orchestrate multiple agents** with specialized roles
4. **Design architecture** while delegating implementation to agents
5. **Review efficiently** because they know what to look for

| Skill Level | AI Leverage | Rationale |
|-------------|-------------|-----------|
| Junior | 1.5× | Autocomplete gains; struggles to validate AI output; may accept hallucinated solutions |
| Senior | 2.5× | Effective with copilot-style tools; can review AI code and spot errors |
| Staff | 4.0× | Can orchestrate agent workflows; designs for AI implementation |
| Principal | 6.0× | Effective multi-agent orchestration; rapid validation; designs systems for agents |
| Senior Principal | 8.0× | Maximizes parallelization; agent swarm orchestration; deep understanding of what to delegate |
| Distinguished | 10.0× | Organization-scale AI leverage; defines patterns for company-wide AI adoption |

**Effective Capacity Calculation:**

```
effective_capacity = Σ(headcount_i × ratio_i × leverage_i)
```

**Example:** A 10-person team with medium composition:
- 1.5 juniors × 1.2 = 1.8
- 5 seniors × 1.8 = 9.0
- 2.5 staff × 2.5 = 6.25
- 0.8 principals × 3.5 = 2.8
- 0.2 SPE × 4.5 = 0.9

**Total effective capacity: 20.75** (equivalent to ~21 conventional engineers)

This explains why small teams of senior engineers with AI can outproduce larger conventional teams.

## Configuration

### JSON Config Example

```json
{
  "cocomo_models": {
    "organic": { "a": 2.4, "b": 1.05, "c": 2.5, "d": 0.38 }
  },
  "variance_multipliers": {
    "optimistic": 0.85,
    "pessimistic": 1.30
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
  "ai_native": {
    "schedule_factor_low": 0.40,
    "schedule_factor_high": 0.60,
    "team_size_factor_low": 0.33,
    "team_size_factor_high": 0.50,
    "minimum_team_size": 2
  },
  "ai_leverage_by_skill": {
    "junior": 1.5,
    "senior": 2.5,
    "staff": 4.0,
    "principal": 6.0,
    "senior_principal": 8.0,
    "distinguished": 10.0
  }
}
```

### Customization Tips

1. **Adjust skill bands** for your market (SF vs remote, startup vs FAANG)
2. **Modify team composition** to match your organization's IC ladder
3. **Tune AI-native factors** based on your team's AI adoption maturity
4. **Use size-based composition** for multi-org or portfolio analysis

## Assumptions and Limitations

1. **COCOMO is for greenfield development** - Maintenance, refactoring, and migrations have different economics
2. **Fully-loaded costs vary** - Our defaults assume US tech market; adjust for your region
3. **Team composition is organization-specific** - Verify ratios match your reality
4. **AI-native estimates assume effective workflows** - Teams new to AI assistance may not achieve these gains
5. **Ranges reflect uncertainty** - Point estimates would be false precision

## References

### Primary Sources

1. Boehm, B. W. (1981). *Software Engineering Economics*. Prentice-Hall.
2. Boehm, B. W., et al. (2000). *Software Cost Estimation with COCOMO II*. Prentice-Hall.
3. Brooks, F. P. (1975). *The Mythical Man-Month*. Addison-Wesley.

### Industry Data

4. Radford Global Technology Survey (2024). Aon.
5. Levels.fyi Engineering Compensation Data (2024-2025).
6. Larson, W. (2021). *Staff Engineer: Leadership Beyond the Management Track*. Self-published.
7. Majors, C. (2017). "The Engineer/Manager Pendulum." honeycomb.io blog.

### Software Estimation Research

8. McConnell, S. (2006). *Software Estimation: Demystifying the Black Art*. Microsoft Press.
9. Jones, C. (2008). *Applied Software Measurement*. 3rd ed. McGraw-Hill.
10. Jones, C. (2017). *Software Methodologies: A Quantitative Guide*. Auerbach Publications.

### Developer Productivity Data

11. Smacchia, P. (2020). ["Mythical man month: 10 lines per developer day."](https://blog.ndepend.com/mythical-man-month-10-lines-per-developer-day/) NDepend Blog.
    - Reports **80 LOC/day** average over 14 years of full-time development
    - Methodology: Logical LOC (PDB sequence points), long-term averaging
    - Context: High code quality and testing standards maintained
    - Key insight: LOC becomes accurate estimation tool when calibrated to specific development context

### AI-Assisted Development

12. GitHub (2022). ["Research: Quantifying GitHub Copilot's impact on developer productivity."](https://github.blog/news-insights/research/research-quantifying-github-copilots-impact-on-developer-productivity-and-happiness/) GitHub Blog.
13. Peng, S., et al. (2023). ["The Impact of AI on Developer Productivity: Evidence from GitHub Copilot."](https://arxiv.org/abs/2302.06590) arXiv:2302.06590.
    - Controlled experiment: treatment group completed HTTP server task **55.8% faster**
    - 95 professional programmers recruited via Upwork (May-June 2022)
