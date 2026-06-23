package ffmpeg

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

type Frame []float32

func readFrames(ctx context.Context, reader io.Reader, frames chan<- Frame, frameSamples int) error {
	buf := make([]byte, frameSamples*4)
	samples := make([]float32, frameSamples)
	var index int64

	for {
		if _, err := io.ReadFull(reader, buf); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return nil
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}

		for i := range samples {
			bits := binary.LittleEndian.Uint32(buf[i*4 : i*4+4])
			samples[i] = math.Float32frombits(bits)
		}

		frame := append([]float32(nil), samples...)

		select {
		case frames <- frame:
			index++
		case <-ctx.Done():
			return ctx.Err()
		default:
			return errors.New("audio recorder buffer is full")
		}
	}
}
