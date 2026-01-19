package model

// ModelPricing contains token pricing for an AI model
type ModelPricing struct {
	Name               string  `json:"name"`
	InputCostPerMTok   float64 `json:"input_cost_per_mtok"`   // cost per million input tokens
	OutputCostPerMTok  float64 `json:"output_cost_per_mtok"`  // cost per million output tokens
	ContextWindow      int     `json:"context_window"`
}

// PricingSource contains citation for pricing data
type PricingSource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Claude model pricing constants (as of late 2025)
// Sources:
// - https://www.anthropic.com/claude/sonnet
// - https://blog.promptlayer.com/claude-ai-pricing-choosing-the-right-model/
var (
	ClaudeSonnet = ModelPricing{
		Name:              "Claude Sonnet",
		InputCostPerMTok:  3.00,
		OutputCostPerMTok: 15.00,
		ContextWindow:     200000,
	}

	ClaudeOpus = ModelPricing{
		Name:              "Claude Opus",
		InputCostPerMTok:  15.00,
		OutputCostPerMTok: 75.00,
		ContextWindow:     200000,
	}

	ClaudeHaiku = ModelPricing{
		Name:              "Claude Haiku",
		InputCostPerMTok:  0.25,
		OutputCostPerMTok: 1.25,
		ContextWindow:     200000,
	}
)

// GetModelPricing returns pricing for a model by name
func GetModelPricing(model string) ModelPricing {
	switch model {
	case "opus":
		return ClaudeOpus
	case "haiku":
		return ClaudeHaiku
	case "sonnet":
		fallthrough
	default:
		return ClaudeSonnet
	}
}

// PricingSources contains citations for the pricing data
var PricingSources = []PricingSource{
	{Name: "Anthropic Claude Pricing", URL: "https://www.anthropic.com/claude/sonnet"},
	{Name: "PromptLayer Analysis", URL: "https://blog.promptlayer.com/claude-ai-pricing-choosing-the-right-model/"},
	{Name: "AI Free API", URL: "https://www.aifreeapi.com/en/posts/claude-api-pricing-per-million-tokens"},
}
