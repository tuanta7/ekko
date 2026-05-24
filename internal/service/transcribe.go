package service

import (
	"github.com/tuanta7/ekko/internal/adapter/ffmpeg"
	"github.com/tuanta7/ekko/internal/adapter/whisper"
)

type TranscribeService struct {
	recorder *ffmpeg.Recorder
	scriber  *whisper.Scriber
}
