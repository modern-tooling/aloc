package effort

import (
	"fmt"

	"github.com/modern-tooling/aloc/internal/model"
)

// TokenEstimate contains estimated token counts for AI processing
type TokenEstimate struct {
	InputTokens  int64    `json:"input_tokens"`
	OutputTokens int64    `json:"output_tokens"`
	TotalTokens  int64    `json:"total_tokens"`
	Assumptions  []string `json:"assumptions"`
}

// TokenEstimationOptions controls token estimation behavior
type TokenEstimationOptions struct {
	// CharsPerToken is characters per token (~4 for Claude)
	CharsPerToken float64
	// Lines contains code and comment line counts
	Lines model.LineMetrics
}

// DefaultTokenOptions returns default token estimation parameters
func DefaultTokenOptions() TokenEstimationOptions {
	return TokenEstimationOptions{
		CharsPerToken: 4.0,
	}
}

// EstimateTokens estimates token consumption for generating/analyzing code
// Uses character-based estimation (honest, not LOC-based myths)
func EstimateTokens(loc int, opts TokenEstimationOptions) TokenEstimate {
	if opts.CharsPerToken <= 0 {
		opts.CharsPerToken = 4.0
	}

	// Estimate characters: ~25 chars per code line, ~40 chars per comment line
	codeChars := int64(opts.Lines.Code * 25)
	commentChars := int64(opts.Lines.Comments * 40)
	totalChars := codeChars + commentChars

	// Input tokens from character count
	inputTokens := totalChars / int64(opts.CharsPerToken)

	// Output tokens are summary-dominated, not LOC-proportional
	// Typical analysis: 500-2000 tokens regardless of codebase size
	// For estimation, use small fraction of input
	outputTokens := inputTokens / 4
	if outputTokens < 500 {
		outputTokens = 500
	}
	if outputTokens > 5000 {
		outputTokens = 5000
	}

	totalTokens := inputTokens + outputTokens

	assumptions := []string{
		"~4 characters per token (Claude tokenizer)",
		"Input: code + comments ingested",
		"Output: summary-dominated (not LOC-proportional)",
	}

	return TokenEstimate{
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
		Assumptions:  assumptions,
	}
}

// EstimateTokensFromLines estimates tokens using line metrics
func EstimateTokensFromLines(lines model.LineMetrics) TokenEstimate {
	opts := DefaultTokenOptions()
	opts.Lines = lines
	return EstimateTokens(lines.Code, opts)
}

// EstimateOutputTokens returns estimated output tokens for given LOC
func EstimateOutputTokens(loc int) int64 {
	// Output is summary-dominated, not LOC-proportional
	tokens := int64(loc / 10)
	if tokens < 500 {
		return 500
	}
	if tokens > 5000 {
		return 5000
	}
	return tokens
}

// EstimateInputTokens returns estimated input tokens given output tokens and ratio
func EstimateInputTokens(outputTokens int64, ratio float64) int64 {
	if ratio <= 0 {
		ratio = 4.0
	}
	return int64(float64(outputTokens) * ratio)
}

// getTokenParams returns token estimation parameters from config
func getTokenParams() (avgLOCPerFile, iterationsPerFile, contextPerCall, outputPerCall int) {
	cfg := GetModelConfig()
	return cfg.TokenEstimation.AvgLOCPerFile,
		cfg.TokenEstimation.IterationsPerFile,
		cfg.TokenEstimation.ContextPerCall,
		cfg.TokenEstimation.OutputPerCall
}

// ImplementationTokenEstimate contains token estimates for full implementation
type ImplementationTokenEstimate struct {
	Files        int      `json:"files"`
	APICalls     int      `json:"api_calls"`
	InputTokens  int64    `json:"input_tokens"`
	OutputTokens int64    `json:"output_tokens"`
	TotalTokens  int64    `json:"total_tokens"`
	Assumptions  []string `json:"assumptions"`
}

// EstimateImplementationTokens estimates tokens for AI to implement code
// This is more realistic than analysis-only estimation:
// - Multiple iterations per file (implement → test → fix cycles)
// - Context loading per API call
// - Code generation output (not just analysis)
func EstimateImplementationTokens(loc int) ImplementationTokenEstimate {
	avgLOCPerFile, iterationsPerFile, contextPerCall, outputPerCall := getTokenParams()

	// Calculate number of files
	files := loc / avgLOCPerFile
	if files < 1 {
		files = 1
	}

	// Total API calls = files × iterations
	apiCalls := files * iterationsPerFile

	// Token estimation
	inputTokens := int64(apiCalls) * int64(contextPerCall)
	outputTokens := int64(apiCalls) * int64(outputPerCall)
	totalTokens := inputTokens + outputTokens

	assumptions := []string{
		fmt.Sprintf("~%d LOC per file average", avgLOCPerFile),
		fmt.Sprintf("~%d iterations per file (implement + test + fix cycles)", iterationsPerFile),
		fmt.Sprintf("~%dK input tokens per call (context)", contextPerCall/1000),
		fmt.Sprintf("~%dK output tokens per call (code generation)", outputPerCall/1000),
	}

	return ImplementationTokenEstimate{
		Files:        files,
		APICalls:     apiCalls,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
		Assumptions:  assumptions,
	}
}
