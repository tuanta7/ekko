# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Ekko is a real-time audio transcription application built with:

- **Backend**: Go with Wails v3 (desktop application framework)
- **Frontend**: React + TypeScript + Vite + TailwindCSS
- **Audio Processing**: FFmpeg (audio recording) + whisper.cpp (transcription)
- **Platform**: Linux (uses PulseAudio for audio capture)

The app captures system audio, processes it through an intelligent chunking system, and transcribes it in real-time using Whisper AI.

## Development Commands

### Initial Setup

```bash
# Check for missing dependencies using Wails doctor
wails3 doctor

# Install all dependencies (Wails v3, audio tools, build Whisper)
make setup

# Download the Whisper model (required before first run)
make download-model
```

The `wails3 doctor` command checks for all required GUI, audio, and build dependencies for Wails development.

The app expects the model at `ggml/ggml-tiny.en-q5_1.bin` (English-only, quantized).

### Running the Application

```bash
# Development mode with hot reload
make dev

# Production build
make build
```

### Testing

```bash
# Run all Go tests
go test ./...

# Run tests for a specific package
go test ./services/chunker -v

# Run tests with coverage
go test -cover ./...
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Build frontend (development)
npm run build:dev

# Build frontend (production)
npm run build
```

## Architecture

### High-Level Data Flow

```
Audio Source (PulseAudio)
    ↓
FFmpeg Recorder (services/adapter/ffmpeg)
    ↓ [100ms PCM frames, 16kHz mono]
AudioChunker (services/chunker)
    ↓ [partial & final audio chunks]
Whisper Scriber (services/adapter/whisper)
    ↓ [transcribed text segments]
TranscribeService (services)
    ↓ [Wails events]
React Frontend (frontend/src)
```

### Backend Architecture

**Main Service: `TranscribeService`** (`services/transcribe.go`)

- Manages transcription sessions
- Coordinates between FFmpeg recorder and Whisper scriber
- Emits events to frontend via Wails event system

**Session Management** (`services/session.go`)

- Each transcription session has a unique ID
- `runSession()` orchestrates the audio → chunks → transcription pipeline
- Uses a job queue with worker goroutines to process chunks

**Audio Chunker** (`services/chunker/`)

- Converts continuous audio stream into transcription-ready chunks
- Uses RMS energy threshold to detect speech vs silence
- Emits **partial chunks** (replaceable, for live updates) every 2 seconds
- Emits **final chunks** after 700ms silence or 8s max duration
- Includes 300ms pre-roll padding and 500ms overlap between forced splits
- See `services/chunker/README.md` for detailed algorithm documentation

**Adapters**:

- `ffmpeg.Recorder`: Captures audio via `pactl` and `ffmpeg`, streams PCM frames
- `whisper.Scriber`: Wraps whisper.cpp Go bindings, thread-safe transcription

**Event System** (`services/event.go`):

- `transcribe:state` - Session state changes (recording, transcribing, stopped)
- `transcribe:partial` - Provisional transcription results (replaceable)
- `transcribe:final` - Completed chunk transcription (permanent)
- `transcribe:error` - Error messages

### Frontend Architecture

**Main Component: `App.tsx`**

- Manages application state and transcription lifecycle
- Integrates with Wails runtime for service calls and events

**Custom Hooks**:

- `useRecorder` - Manages recording state and handles Wails service calls

**State Management** (`lib/state.ts`):

- Tracks transcript segments (partial vs final)
- Handles replacing partial results with newer partials or finals

**Components**:

- `AppHeader` - Window controls and recording UI
- `TranscriptMain` - Displays live transcription results

### Whisper.cpp Integration

The app uses whisper.cpp as a Git submodule. Setup requires:

1. Building whisper.cpp libraries via `scripts/setup-whisper.sh`
2. Setting environment variables in `scripts/setup-env.sh`:
   - `C_INCLUDE_PATH` - Points to whisper headers
   - `LIBRARY_PATH` - Points to compiled libraries
3. The Makefile sources these variables before running Wails commands

**Important**: Always run `make dev` or `make build` (not bare `wails3` commands) to ensure environment is configured.

### Wails v3 Integration

- Services are registered in `main.go` via `application.NewService()`
- Events are registered in `init()` with type information
- Frontend calls Go methods via auto-generated bindings in `@wailsio/runtime`
- Window is frameless, transparent, and always-on-top for overlay use case

## Key Implementation Details

### Audio Chunking Strategy

The chunker maintains an internal state machine:
- **Idle**: Collecting pre-roll (last 300ms of silence)
- **In Speech**: Buffering utterance, emitting partials every 2s
- **Trailing Silence**: Waiting for 700ms silence to finalize

Chunks have timestamps relative to the start of the audio stream, not wall-clock time.

### Transcription Threading

- Whisper model is lazy-loaded on first `Start()` call
- Single `Scriber` instance protected by mutex (whisper.cpp contexts aren't thread-safe)
- Workers process chunks sequentially from job queue
- Final chunks are always processed (blocking), partial chunks are dropped if queue is full

### Session Lifecycle

1. Frontend calls `Start(source)` → creates session, starts FFmpeg stream
2. `runSession()` goroutine manages frame → chunk → transcribe pipeline
3. Frontend calls `Stop(sessionID)` → cancels context, triggers cleanup
4. Session cleanup: flushes pending audio, processes remaining chunks, emits stop event

### Model Path

The Whisper model path is hardcoded in `services/adapter/whisper/scriber.go`:
```go
const modelPath = "ggml/ggml-tiny.en-q5_1.bin"
```

If using a different model, update this constant.

## Testing Notes

- `services/chunker` has comprehensive unit tests covering edge cases
- Use `ffmpeg.Recorder.ListSources()` to enumerate available PulseAudio sources
- For testing without audio hardware, you'd need to mock the FFmpeg adapter

## Common Issues

**Missing dependencies**: Run `wails3 doctor` to identify and install missing packages:
```bash
wails3 doctor
```

**Window transparency not working**:
- Ensure a compositor is running (GNOME/KDE have built-in compositors)
- Install and run picom: `sudo apt install picom && picom -b`
- Check if `WebviewIsTransparent` is set in `main.go` Linux configuration

**Whisper model not found**: Run `make download-model` first

**FFmpeg errors**: Ensure PulseAudio is running and sources are available:
```bash
pactl list sources short
```

**Build failures**: Check that environment is set up:

```bash
source ./scripts/setup-env.sh
echo $C_INCLUDE_PATH
echo $LIBRARY_PATH
```

**Frontend not updating**: Ensure Wails event registration matches between `main.go` and frontend event listeners
