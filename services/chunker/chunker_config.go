package chunker

import (
	"time"

	"github.com/tuanta7/ekko/services/adapter/ffmpeg"
)

type Config struct {
	sampleRate       int
	frameDuration    time.Duration
	minSpeech        time.Duration
	silenceToFinal   time.Duration
	speechPad        time.Duration
	overlap          time.Duration
	partialWindow    time.Duration
	partialInterval  time.Duration
	maxFinalDuration time.Duration
	energyThreshold  float64
}

var DefaultConfig = Config{
	sampleRate:       ffmpeg.DefaultSampleRate,
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
