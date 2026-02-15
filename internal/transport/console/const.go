package console

import "time"

type screen int

const (
	screenMenu screen = iota
	screenRecording
)

type option = int

const (
	optionStartSession option = iota
	optionAudioSource
	optionChunkDuration
	optionExit
)

const (
	defaultChunkDuration = time.Second
)

var (
	menuOptions = []string{"Start Session", "Audio Source", "Chunk Duration", "Exit"}
)
