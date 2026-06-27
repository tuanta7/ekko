# Ekko

<img align="right" src="build/appicon.png" alt="Ekko Mascot" width="150"/>

![Status](https://img.shields.io/badge/status-development-orange)
![Language](https://img.shields.io/badge/lang-Go-blue)
![License](https://img.shields.io/badge/license-MIT-green)
[![CI](https://github.com/tuanta7/ekko/actions/workflows/ci.yml/badge.svg)](https://github.com/tuanta7/ekko/actions/workflows/ci.yml)

## Setup

Download the fixed Whisper model before starting transcription. Ekko loads
`ggml/ggml-tiny.en-q5_1.bin` and transcribes in English.

```sh
make download-model
```

Downloaded `.bin` files are ignored by Git.

## Development

Install app and audio dependencies, then build Whisper support:

```sh
make setup
```

Run the app in development mode:

```sh
make dev
```

For a production build:

```sh
make build
```

## Architecture

Ekko is a Wails v3 desktop application with a React frontend and a Go
transcription service. The active audio path is:

```text
React UI -> Wails binding -> TranscribeService -> FFmpeg/PulseAudio
         -> audio chunker -> bounded job queue -> Whisper worker
         -> Wails events -> React UI
```

### Components

- `frontend/src` contains the controls and transcript view. It calls the
  generated `TranscribeService` bindings and listens for transcription events.
- `services.TranscribeService` owns application lifecycle state, the recorder,
  the loaded Whisper model, and the active session. Only one session may run at
  a time.
- `services/adapter/ffmpeg` discovers PulseAudio sources with `pactl` and runs
  FFmpeg to produce mono, 16 kHz, float32 PCM in 100 ms frames.
- `services/chunker` converts PCM frames into rolling partial chunks and stable
  final chunks. Its state machine and timing rules are documented in
  [`services/chunker/README.md`](services/chunker/README.md).
- `services/adapter/whisper` loads the English-only
  `ggml/ggml-tiny.en-q5_1.bin`, serializes model access, and combines Whisper
  segments into transcript text.

### Process Flow

1. Wails registers `TranscribeService` and calls `ServiceStartup`. The service
   stores the application context and event emitter, then initializes the
   recorder and session map.
2. The frontend calls `ListSources`. The recorder runs `pactl list sources
   short` with a three-second timeout and returns the available source names.
3. The frontend calls `Start`. The service selects the first source when none
   is supplied, rejects a second active session, and loads or reuses the fixed
   Whisper model.
4. FFmpeg starts against the PulseAudio source and streams PCM frames through a
   channel sized for 30 seconds of audio. A full recorder buffer is reported as
   an error instead of silently dropping captured audio.
5. A session goroutine passes each frame to the audio chunker. Emitted chunks
   receive monotonically increasing IDs and enter a four-item transcription
   queue. Partial chunks are dropped when this queue is full so stale previews
   do not delay final output; final chunks wait for queue capacity.
6. A single worker transcribes queued chunks with Whisper. Partial results emit
   `transcribe:partial`; final results emit `transcribe:final`. State changes and
   failures emit `transcribe:state` and `transcribe:error` respectively.
7. The React UI replaces the current partial text, appends final transcript
   lines with their audio timestamps, and reflects recording, transcribing,
   error, and stopped states.
8. `Stop` cancels the session context. FFmpeg exits, buffered speech is flushed
   as a final chunk, queued jobs finish, and the service removes the session and
   emits `stopped`. During application shutdown, all active sessions are
   cancelled and drained before the Whisper model is closed.
