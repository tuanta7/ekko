package session

import (
	"math"
	"time"

	"github.com/tuanta7/ekko/services/adapter/ffmpeg"
)

const (
	defaultEnergyThreshold = 0.01
)

type ChunkerConfig struct {
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

var DefaultChunkerConfig = ChunkerConfig{
	sampleRate:       ffmpeg.DefaultSampleRate,
	frameDuration:    ffmpeg.DefaultFrameDuration,
	minSpeech:        250 * time.Millisecond,
	silenceToFinal:   700 * time.Millisecond,
	speechPad:        300 * time.Millisecond,
	overlap:          500 * time.Millisecond,
	partialWindow:    5 * time.Second,
	partialInterval:  2 * time.Second,
	maxFinalDuration: 8 * time.Second,
	energyThreshold:  defaultEnergyThreshold,
}

type AudioChunk struct {
	Samples []float32
	Start   time.Duration
	End     time.Duration
	Final   bool
}

type AudioChunker struct {
	Config ChunkerConfig

	sampleCursor int64

	inSpeech       bool
	speechStart    int64
	speechSamples  []float32
	preRollSamples []float32

	silenceSamples      int
	activeSpeechSamples int
	lastPartialAt       int
}

func NewAudioChunker() *AudioChunker {
	return &AudioChunker{
		Config: DefaultChunkerConfig,
	}
}

func (c *AudioChunker) AddFrame(samples []float32) []AudioChunk {
	if len(samples) == 0 {
		return nil
	}

	frameStart := c.sampleCursor
	c.sampleCursor += int64(len(samples))

	if isSpeech(samples, c.Config.energyThreshold) {
		return c.addSpeechFrame(samples, frameStart)
	}

	return c.addSilenceFrame(samples)
}

func (c *AudioChunker) Flush() []AudioChunk {
	if !c.inSpeech {
		return nil
	}

	chunk, ok := c.finalChunk(0)
	c.resetAfterFinal(nil)
	if !ok {
		return nil
	}
	return []AudioChunk{chunk}
}

func (c *AudioChunker) addSpeechFrame(samples []float32, frameStart int64) []AudioChunk {
	if !c.inSpeech {
		c.inSpeech = true
		c.speechStart = frameStart - int64(len(c.preRollSamples))
		c.speechSamples = append([]float32(nil), c.preRollSamples...)
		c.silenceSamples = 0
		c.activeSpeechSamples = 0
		c.lastPartialAt = 0
	}

	c.speechSamples = append(c.speechSamples, samples...)
	c.activeSpeechSamples += len(samples)
	c.silenceSamples = 0

	var chunks []AudioChunk
	if c.shouldEmitPartial() {
		chunks = append(chunks, c.partialChunk())
		c.lastPartialAt = len(c.speechSamples)
	}

	if durationForSamples(len(c.speechSamples), c.Config.sampleRate) >= c.Config.maxFinalDuration {
		chunk, ok := c.finalChunk(0)
		tail := tailSamples(chunk.Samples, samplesForDuration(c.Config.overlap, c.Config.sampleRate))
		c.resetAfterFinal(tail)
		if ok {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

func (c *AudioChunker) addSilenceFrame(samples []float32) []AudioChunk {
	if !c.inSpeech {
		c.appendPreRoll(samples)
		return nil
	}

	c.speechSamples = append(c.speechSamples, samples...)
	c.silenceSamples += len(samples)

	if durationForSamples(c.silenceSamples, c.Config.sampleRate) < c.Config.silenceToFinal {
		return nil
	}

	trailingSilenceToDrop := c.silenceSamples - samplesForDuration(c.Config.speechPad, c.Config.sampleRate)
	if trailingSilenceToDrop < 0 {
		trailingSilenceToDrop = 0
	}

	chunk, ok := c.finalChunk(trailingSilenceToDrop)
	tail := tailSamples(c.speechSamples, samplesForDuration(c.Config.speechPad, c.Config.sampleRate))
	c.resetAfterFinal(tail)
	if !ok {
		return nil
	}

	return []AudioChunk{chunk}
}

func (c *AudioChunker) shouldEmitPartial() bool {
	if len(c.speechSamples) == 0 {
		return false
	}
	if durationForSamples(c.activeSpeechSamples, c.Config.sampleRate) < c.Config.minSpeech {
		return false
	}
	return durationForSamples(len(c.speechSamples)-c.lastPartialAt, c.Config.sampleRate) >= c.Config.partialInterval
}

func (c *AudioChunker) partialChunk() AudioChunk {
	windowSamples := samplesForDuration(c.Config.partialWindow, c.Config.sampleRate)
	startOffset := 0
	if len(c.speechSamples) > windowSamples {
		startOffset = len(c.speechSamples) - windowSamples
	}

	samples := append([]float32(nil), c.speechSamples[startOffset:]...)
	startSample := c.speechStart + int64(startOffset)
	endSample := startSample + int64(len(samples))

	return AudioChunk{
		Samples: samples,
		Start:   durationForSamples64(startSample, c.Config.sampleRate),
		End:     durationForSamples64(endSample, c.Config.sampleRate),
		Final:   false,
	}
}

func (c *AudioChunker) finalChunk(dropTailSamples int) (AudioChunk, bool) {
	end := len(c.speechSamples) - dropTailSamples
	if end < 0 {
		end = 0
	}

	if durationForSamples(c.activeSpeechSamples, c.Config.sampleRate) < c.Config.minSpeech {
		return AudioChunk{}, false
	}

	samples := append([]float32(nil), c.speechSamples[:end]...)
	endSample := c.speechStart + int64(len(samples))

	return AudioChunk{
		Samples: samples,
		Start:   durationForSamples64(c.speechStart, c.Config.sampleRate),
		End:     durationForSamples64(endSample, c.Config.sampleRate),
		Final:   true,
	}, true
}

func (c *AudioChunker) resetAfterFinal(preRoll []float32) {
	c.inSpeech = false
	c.speechStart = 0
	c.speechSamples = nil
	c.silenceSamples = 0
	c.activeSpeechSamples = 0
	c.lastPartialAt = 0
	c.preRollSamples = append([]float32(nil), preRoll...)
	c.capPreRoll()
}

func (c *AudioChunker) appendPreRoll(samples []float32) {
	c.preRollSamples = append(c.preRollSamples, samples...)
	c.capPreRoll()
}

func (c *AudioChunker) capPreRoll() {
	maxSamples := samplesForDuration(c.Config.speechPad, c.Config.sampleRate)
	if len(c.preRollSamples) > maxSamples {
		c.preRollSamples = append([]float32(nil), c.preRollSamples[len(c.preRollSamples)-maxSamples:]...)
	}
}

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

func samplesForDuration(duration time.Duration, sampleRate int) int {
	return int(duration.Seconds() * float64(sampleRate))
}

func durationForSamples(samples int, sampleRate int) time.Duration {
	return durationForSamples64(int64(samples), sampleRate)
}

func durationForSamples64(samples int64, sampleRate int) time.Duration {
	return time.Duration(float64(samples) / float64(sampleRate) * float64(time.Second))
}

func tailSamples(samples []float32, count int) []float32 {
	if count <= 0 || len(samples) == 0 {
		return nil
	}
	if len(samples) <= count {
		return append([]float32(nil), samples...)
	}
	return append([]float32(nil), samples[len(samples)-count:]...)
}
