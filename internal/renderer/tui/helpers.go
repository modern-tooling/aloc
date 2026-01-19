package tui

import "fmt"

// formatNumber formats an integer with comma separators
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d,%03d,%03d", n/1000000, (n/1000)%1000, n%1000)
}

// formatLOCShort formats LOC in a fixed-width format
func formatLOCShort(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%5d", n)
	}
	return fmt.Sprintf("%5.1fk", float64(n)/1000)
}

// formatMagnitude formats large numbers with K/M suffix
func formatMagnitude(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.2fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.2fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

// truncate truncates a string to maxLen, adding ellipsis if needed
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
