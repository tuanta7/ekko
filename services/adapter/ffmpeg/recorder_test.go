package ffmpeg

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"testing"
)

func TestReadFramesDecodesFloat32LE(t *testing.T) {
	var input bytes.Buffer
	for _, sample := range []float32{0.25, -0.5, 1.0, 0.0} {
		if err := binary.Write(&input, binary.LittleEndian, math.Float32bits(sample)); err != nil {
			t.Fatal(err)
		}
	}

	frames := make(chan Frame, 1)
	if err := readFrames(context.Background(), &input, frames, 4); err != nil {
		t.Fatal(err)
	}

	frame := <-frames
	if len(frame) != 4 {
		t.Fatalf("expected 4 samples, got %d", len(frame))
	}

	expected := []float32{0.25, -0.5, 1.0, 0.0}
	for i, sample := range frame {
		if sample != expected[i] {
			t.Fatalf("sample %d: expected %f, got %f", i, expected[i], sample)
		}
	}
}

func TestReadFramesReturnsWhenBufferFull(t *testing.T) {
	var input bytes.Buffer
	for range 8 {
		if err := binary.Write(&input, binary.LittleEndian, math.Float32bits(float32(0.25))); err != nil {
			t.Fatal(err)
		}
	}

	frames := make(chan Frame)
	err := readFrames(context.Background(), &input, frames, 4)
	if err == nil {
		t.Fatal("expected buffer full error")
	}
}
