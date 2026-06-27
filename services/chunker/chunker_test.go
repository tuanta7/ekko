package chunker

import (
	"testing"
	"time"
)

func TestAudioChunkerFinalizesUtteranceAfterSilence(t *testing.T) {
	audioChunker := NewAudioChunker()

	addTestFrames(audioChunker, 3, 0)
	addTestFrames(audioChunker, 10, 0.2)
	chunks := addTestFrames(audioChunker, 7, 0)

	if len(chunks) != 1 {
		t.Fatalf("expected one final chunk, got %d", len(chunks))
	}
	if !chunks[0].Final {
		t.Fatal("expected final chunk")
	}
	if chunks[0].Start != 0 {
		t.Fatalf("expected chunk to include pre-roll from t=0, got %s", chunks[0].Start)
	}
	if chunks[0].End != 1600*time.Millisecond {
		t.Fatalf("expected trailing silence trimmed to speech pad, got end %s", chunks[0].End)
	}
}

func TestAudioChunkerDiscardsShortNoise(t *testing.T) {
	audioChunker := NewAudioChunker()

	addTestFrames(audioChunker, 1, 0.2)
	chunks := addTestFrames(audioChunker, 7, 0)

	if len(chunks) != 0 {
		t.Fatalf("expected no chunks for short noise, got %d", len(chunks))
	}
}

func TestAudioChunkerEmitsPartialAndFinal(t *testing.T) {
	audioChunker := NewAudioChunker()

	chunks := addTestFrames(audioChunker, 20, 0.2)
	if len(chunks) != 1 {
		t.Fatalf("expected one partial chunk, got %d", len(chunks))
	}
	if chunks[0].Final {
		t.Fatal("expected partial chunk")
	}

	chunks = append(chunks, addTestFrames(audioChunker, 7, 0)...)
	if len(chunks) != 2 {
		t.Fatalf("expected partial and final chunk, got %d", len(chunks))
	}
	if !chunks[1].Final {
		t.Fatal("expected final chunk")
	}
}

func TestAudioChunkerPreservesConfiguredOverlapAfterMaximumDuration(t *testing.T) {
	audioChunker := NewAudioChunker()
	audioChunker.Config.partialInterval = time.Hour

	chunks := addTestFrames(audioChunker, 80, 0.2)
	if len(chunks) != 1 || !chunks[0].Final {
		t.Fatalf("expected one maximum-duration final chunk, got %#v", chunks)
	}

	addTestFrames(audioChunker, 3, 0.2)
	chunks = audioChunker.Flush()
	if len(chunks) != 1 {
		t.Fatalf("expected one flushed chunk, got %d", len(chunks))
	}
	if chunks[0].Start != 7500*time.Millisecond {
		t.Fatalf("expected 500ms overlap from t=7.5s, got %s", chunks[0].Start)
	}
	if len(chunks[0].Samples) != samplesForDuration(800*time.Millisecond, audioChunker.Config.sampleRate) {
		t.Fatalf("expected overlap plus new speech, got %d samples", len(chunks[0].Samples))
	}
}

func addTestFrames(audioChunker *AudioChunker, count int, amplitude float32) []AudioChunk {
	frameSamples := samplesForDuration(audioChunker.Config.frameDuration, audioChunker.Config.sampleRate)
	frame := make([]float32, frameSamples)
	for i := range frame {
		frame[i] = amplitude
	}

	var chunks []AudioChunk
	for range count {
		chunks = append(chunks, audioChunker.AddFrame(frame)...)
	}
	return chunks
}
