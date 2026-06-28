package chunker

import (
	"math"
	"time"

	"golang.org/x/exp/constraints"
)

// isSpeech reports whether a frame's RMS amplitude meets the speech threshold.
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

// samplesDuration converts a 64-bit sample count to a duration.
func samplesDuration[T constraints.Integer](samples T, sampleRate int) time.Duration {
	sampleCount := float64(samples) / float64(sampleRate)
	return time.Duration(sampleCount * float64(time.Second))
}
