# Ekko

<img align="right" src="web/logo.png" alt="Ekko Mascot" width="150"/>

![Status](https://img.shields.io/badge/status-development-orange)
![Language](https://img.shields.io/badge/lang-Go-blue)
![License](https://img.shields.io/badge/license-MIT-green)

A near real-time desktop audio transcription tool with a clean terminal and web UI. Capture system audio, transcribe on the fly using local Whisper models or Google Gemini API

## Key features

- Local and cloud transcription backends: whisper (local) and gemini (Google API).
- Privacy-first local mode when using Whisper models; no network round trips.
- Clean UI for live transcription and simple controls.

![Demo](demo.gif)

## Quick Start

![OS](https://img.shields.io/badge/Ubuntu-E95420?style=for-the-badge&logo=Ubuntu&logoColor=white)

```sh
# Install make
sudo apt update
sudo apt install build-essential

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

