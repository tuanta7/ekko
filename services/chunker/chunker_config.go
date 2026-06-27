package chunker

import (
	"time"

	"github.com/tuanta7/ekko/services/adapter/ffmpeg"
)

// Config controls speech detection and the timing of emitted audio chunks.
type Config struct {
	// sampleRate is the number of audio samples processed per second.
	sampleRate int
	// frameDuration is the expected duration of each input audio frame.
	frameDuration time.Duration
	// minSpeech is the minimum detected speech required to emit a chunk.
	minSpeech time.Duration
	// silenceToFinal is the consecutive silence required to finish an utterance.
	silenceToFinal time.Duration
	// speechPad is the silence retained before and after detected speech.
	speechPad time.Duration
	// overlap is the audio retained when an utterance is split at its maximum duration.
	overlap time.Duration
	// partialWindow is the maximum amount of recent audio included in a partial chunk.
	partialWindow time.Duration
	// partialInterval is the amount of new audio required between partial chunks.
	partialInterval time.Duration
	// maxFinalDuration is the maximum utterance length before a forced final chunk.
	maxFinalDuration time.Duration
	// energyThreshold is the minimum RMS amplitude used to classify a frame as speech.
	energyThreshold float64
}

// DefaultConfig contains the standard chunking settings used by NewAudioChunker.
var DefaultConfig = Config{
	sampleRate:       ffmpeg.DefaultSampleRate, // samples per second
	frameDuration:    ffmpeg.DefaultFrameDuration,
	minSpeech:        250 * time.Millisecond,
	silenceToFinal:   700 * time.Millisecond,
	speechPad:        300 * time.Millisecond,
	overlap:          500 * time.Millisecond,
	partialWindow:    5 * time.Second,
	partialInterval:  2 * time.Second,
	maxFinalDuration: 8 * time.Second,
	energyThreshold:  0.01,
}
