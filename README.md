# Ekko

<img align="right" src="./logo.png" alt="Ekko Mascot" width="150"/>

![Status](https://img.shields.io/badge/status-development-orange)
![Language](https://img.shields.io/badge/lang-Go-blue)
![License](https://img.shields.io/badge/license-MIT-green)

A near real-time desktop audio transcription tool with a clean terminal UI. Capture system audio, transcribe on the fly using local Whisper models or Google Gemini API

## Key features

- Local and cloud transcription backends: whisper (local) and gemini (Google API).
- Privacy-first local mode when using Whisper models; no network round trips.
- Clean TUI for live transcription and simple controls.

![Demo](demo.gif)

> Darts scene from Ted Lasso (2x speed)

## Quick Start

```sh
# Build the go binding library
make build

# Run the app
make dev
```
### Prerequisites

Run the script below to install required dependencies

```bash
# Install required dependencies
make install

# or
./install.sh
```

This will install:

- `pulseaudio-utils` - For audio capture
- `ffmpeg` - For audio processing

## Configuration

Environment variables

| Variable         | Description           | Values              |
|------------------|-----------------------|---------------------|
| TRANSCRIBER_MODE | Transcription backend | `gemini`, `whisper` |
| GEMINI_API_KEY   | Google Gemini API key | Your API key        |

### Get a Gemini API Key

- Visit [Google AI Studio](https://aistudio.google.com/app/api-keys)
- Create a new API key
- Add it to your `.env` file

## Todo List

- [x] Add support for local whisper models
- [x] Custom recording chunk duration settings
- [ ] Concurrent audio transcription
- [ ] Real-time word highlighting

