package chunker

import (
	"time"
)

// AudioChunk contains a timestamped section of audio ready for transcription.
type AudioChunk struct {
	// Samples contain normalized PCM audio samples.
	Samples []float32
	// Start is the chunk's offset from the beginning of the audio stream.
	Start time.Duration
	// End is the chunk's end offset from the beginning of the audio stream.
	End time.Duration
	// Final reports whether the chunk completes an utterance.
	Final bool
}

// AudioChunker groups incoming audio frames into partial and final speech chunks.
type AudioChunker struct {
	// Config defines the speech detection and chunk emission behavior.
	Config Config

	// sampleCursor is the total number of input samples processed so far.
	sampleCursor int64
	// inSpeech reports whether the chunker is currently collecting an utterance.
	inSpeech bool
	// speechStart is the absolute sample offset where the current buffer begins.
	speechStart int64
	// speechSamples contains the current utterance, including retained padding.
	speechSamples []float32
	// preRollSamples contains recent idle audio to prepend when speech begins.
	preRollSamples []float32
	// silenceSamples is the current run of non-speech samples in the utterance.
	silenceSamples int
	// activeSpeechSamples is the total number of samples classified as speech.
	activeSpeechSamples int
	// lastPartialAt is the buffer length when the previous partial chunk was emitted.
	lastPartialAt int
}

func NewAudioChunker() *AudioChunker {
	return &AudioChunker{
		Config: DefaultConfig,
	}
}

// AddFrame classifies and stores one audio frame, returning any chunks it completes.
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

// Flush finalizes and returns any pending utterance that contains enough speech.
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

// addSpeechFrame adds a detected speech frame and emits chunks when their limits are reached.
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

	if samplesDuration(len(c.speechSamples), c.Config.sampleRate) >= c.Config.maxFinalDuration {
		chunk, ok := c.finalChunk(0)
		tail := tailSamples(chunk.Samples, samplesForDuration(c.Config.overlap, c.Config.sampleRate))
		c.resetAfterFinal(tail)
		if ok {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

// addSilenceFrame retains padding and finalizes an utterance after sustained silence.
func (c *AudioChunker) addSilenceFrame(samples []float32) []AudioChunk {
	if !c.inSpeech {
		c.appendPreRoll(samples)
		return nil
	}

	c.speechSamples = append(c.speechSamples, samples...)
	c.silenceSamples += len(samples)

	if samplesDuration(c.silenceSamples, c.Config.sampleRate) < c.Config.silenceToFinal {
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

// shouldEmitPartial reports whether enough speech and new audio exist for a partial chunk.
func (c *AudioChunker) shouldEmitPartial() bool {
	if len(c.speechSamples) == 0 {
		return false
	}

	if samplesDuration(c.activeSpeechSamples, c.Config.sampleRate) < c.Config.minSpeech {
		return false
	}

	return samplesDuration(len(c.speechSamples)-c.lastPartialAt, c.Config.sampleRate) >= c.Config.partialInterval
}

// partialChunk copies the most recent configured window into a non-final chunk.
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
		Start:   samplesDuration(startSample, c.Config.sampleRate),
		End:     samplesDuration(endSample, c.Config.sampleRate),
		Final:   false,
	}
}

// finalChunk builds a final chunk after removing the requested number of trailing samples.
func (c *AudioChunker) finalChunk(dropTailSamples int) (AudioChunk, bool) {
	end := len(c.speechSamples) - dropTailSamples
	if end < 0 {
		end = 0
	}

	if samplesDuration(c.activeSpeechSamples, c.Config.sampleRate) < c.Config.minSpeech {
		return AudioChunk{}, false
	}

	samples := append([]float32(nil), c.speechSamples[:end]...)
	endSample := c.speechStart + int64(len(samples))

	return AudioChunk{
		Samples: samples,
		Start:   samplesDuration(c.speechStart, c.Config.sampleRate),
		End:     samplesDuration(endSample, c.Config.sampleRate),
		Final:   true,
	}, true
}

// resetAfterFinal clears the utterance state and preserves the supplied audio as pre-roll.
func (c *AudioChunker) resetAfterFinal(preRoll []float32) {
	c.inSpeech = false
	c.speechStart = 0
	c.speechSamples = nil
	c.silenceSamples = 0
	c.activeSpeechSamples = 0
	c.lastPartialAt = 0
	c.preRollSamples = append([]float32(nil), preRoll...)
}

// appendPreRoll adds idle audio to the rolling pre-speech buffer.
func (c *AudioChunker) appendPreRoll(samples []float32) {
	c.preRollSamples = append(c.preRollSamples, samples...)
	c.capPreRoll()
}

// capPreRoll limits pre-roll audio to the configured speech padding duration.
func (c *AudioChunker) capPreRoll() {
	maxSamples := samplesForDuration(c.Config.speechPad, c.Config.sampleRate)
	if len(c.preRollSamples) > maxSamples {
		c.preRollSamples = append([]float32(nil), c.preRollSamples[len(c.preRollSamples)-maxSamples:]...)
	}
}

// samplesForDuration converts a duration to a sample count at the given sample rate.
func samplesForDuration(duration time.Duration, sampleRate int) int {
	return int(duration.Seconds() * float64(sampleRate))
}

// tailSamples returns a copy of at most count samples from the end of an audio buffer.
func tailSamples(samples []float32, count int) []float32 {
	if count <= 0 || len(samples) == 0 {
		return nil
	}
	if len(samples) <= count {
		return append([]float32(nil), samples...)
	}
	return append([]float32(nil), samples[len(samples)-count:]...)
}
