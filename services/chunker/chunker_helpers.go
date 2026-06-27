package chunker

import "math"

// Returns true if the loudness is greater than or equal to the threshold
func isSpeech(samples []float32, threshold float64) bool {
	if len(samples) == 0 {
		return false
	}

	var sum float64
	for _, sample := range samples {
		sum += float64(sample * sample)
	}

	rms := math.Sqrt(sum / float64(len(samples)))
	return rms >= threshold
}
