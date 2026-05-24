package ffmpeg

type Recorder struct {
}

func NewRecorder() *Recorder {
	return &Recorder{}
}

func (r *Recorder) ListSources() []string {
	return []string{}
}
