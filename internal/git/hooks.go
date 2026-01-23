package git

// lateVolatility checks for spikes in last 25% of window
func lateVolatility(buckets []Bucket) bool {
	n := len(buckets)
	if n < 6 {
		return false
	}

	max := maxChurn(buckets)
	if max == 0 {
		return false
	}

	spikes := 0
	for i := n - n/4; i < n; i++ {
		if float64(buckets[i].Churn) >= 0.7*float64(max) {
			spikes++
		}
	}

	return spikes >= 2
}

// sustainedHighChurn checks if ≥30% of buckets are high activity
func sustainedHighChurn(buckets []Bucket) bool {
	max := maxChurn(buckets)
	if max == 0 {
		return false
	}

	highCount := 0
	for _, b := range buckets {
		if float64(b.Churn) >= 0.5*float64(max) {
			highCount++
		}
	}

	return float64(highCount)/float64(len(buckets)) >= 0.30
}

// maxChurn finds the maximum churn in a bucket slice
func maxChurn(buckets []Bucket) int {
	max := 0
	for _, b := range buckets {
		if b.Churn > max {
			max = b.Churn
		}
	}
	return max
}
