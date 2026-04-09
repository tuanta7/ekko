# Ekko

<img align="right" src="web/logo.png" alt="Ekko Mascot" width="150"/>

![Status](https://img.shields.io/badge/status-development-orange)
![Language](https://img.shields.io/badge/lang-Go-blue)
![License](https://img.shields.io/badge/license-MIT-green)
[![CI](https://github.com/tuanta7/ekko/actions/workflows/ci.yml/badge.svg)](https://github.com/tuanta7/ekko/actions/workflows/ci.yml)

A near real-time desktop audio transcription tool with a clean terminal and web UI. Capture system audio, transcribe on the fly using local Whisper models or Google Gemini API.

## Key features

- Local and cloud transcription backends: Whisper (local) and Gemini (Google API).
- Privacy-first local mode when using Whisper models; no network round trips.
- Clean TUI for live transcription and simple controls.
- Optional web UI served over HTTP.

![Demo](demo.gif)

## Quick Start

![OS](https://img.shields.io/badge/Ubuntu-E95420?style=for-the-badge&logo=Ubuntu&logoColor=white)

### 1. Install system dependencies

```sh
sudo apt update && sudo apt install build-essential
make install   # installs pulseaudio-utils and ffmpeg
```

### 2. Build the whisper.cpp Go binding

```sh
make build
```

### 3. Download a model

Download a GGML Whisper model from [huggingface.co/ggerganov/whisper.cpp](https://huggingface.co/ggerganov/whisper.cpp/tree/main) and place it in `ggml/`.
See [`ggml/README.md`](ggml/README.md) for details.

```sh
# example — download the smallest English model (~150 MB)
wget -P ggml/ https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin
```

### 4. Run

```sh
# TUI mode (default)
make dev                       # uses MODEL=ggml-base by default
make dev MODEL=ggml-small-q5_1 # override model

# Web mode (serves on :8080)
make dev-web
```

Or run the binary directly after building:

```sh
source ./scripts/setup-whisper.sh
go build -o ekko ./cmd/ekko

# TUI mode (default)
./ekko run

# Web mode
./ekko run --web --addr :8080

# Different model
./ekko run --model ggml-small-q5_1
```

## CLI reference

```
USAGE:
   ekko run [OPTIONS]

OPTIONS:
   --web, -w               Enable web/server mode (default: false)
   --addr value, -a value  Address to listen on in web mode (default: ":8080")
   --model value, -m value Whisper GGML model name to use (default: "ggml-base")
   --help, -h              show help
```

## Prerequisites

| Dependency | Purpose |
|---|---|
| `pulseaudio-utils` | Audio capture (`parecord`) |
| `ffmpeg` | Audio processing / conversion |
| Go ≥ 1.22 | Build toolchain |
| cmake, gcc | Build whisper.cpp native library |

Run `make install` to install the runtime dependencies automatically.

## Troubleshooting

- **`ffmpeg` not found** — run `make install` or `sudo apt install ffmpeg`.
- **No audio captured** — ensure PulseAudio (or PipeWire with the PulseAudio compatibility layer) is running: `pulseaudio --check`.
- **Whisper model not found** — confirm the `.bin` file is in `ggml/` and the name passed to `--model` matches (without the `.bin` extension).
- **`C_INCLUDE_PATH` / `LIBRARY_PATH` errors** — run `source ./scripts/setup-whisper.sh` before building/running outside of Make.

## License

MIT — see [LICENSE](LICENSE).
