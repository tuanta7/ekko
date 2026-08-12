package ffmpeg

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"regexp"
	"strings"
)

// Lines look like: [AVFoundation indev @ 0x…] [1] MacBook Pro Microphone
var avDevice = regexp.MustCompile(`\[(\d+)\] (.+)$`)

// ListSources returns the AVFoundation audio device names. macOS has no
// equivalent of Pulse's `.monitor` sinks, so capturing what the machine plays
// (rather than the mic) needs a loopback device such as BlackHole.
func (r *Recorder) ListSources(ctx context.Context) ([]string, error) {
	// The device list goes to stderr and ffmpeg exits non-zero afterwards.
	cmd := exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-f", "avfoundation", "-list_devices", "true", "-i", "")

	var buf bytes.Buffer
	cmd.Stderr = &buf
	_ = cmd.Run()

	var sources []string
	audio := false
	for _, line := range strings.Split(buf.String(), "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasSuffix(line, "audio devices:"):
			audio = true
		case strings.HasSuffix(line, "video devices:"):
			audio = false
		case audio:
			if match := avDevice.FindStringSubmatch(line); match != nil {
				sources = append(sources, match[2])
			}
		}
	}

	if len(sources) == 0 {
		return nil, errors.New("no avfoundation audio devices found")
	}

	return sources, nil
}

// AVFoundation inputs are "<video>:<audio>"; an empty video half records audio
// only, and the audio half accepts a device name as well as an index.
func inputArgs(source string) []string {
	return []string{"-f", "avfoundation", "-i", ":" + source}
}
