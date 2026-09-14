# Ekko

![Status](https://img.shields.io/badge/status-development-orange)
![Language](https://img.shields.io/badge/lang-Go-blue)
[![CI](https://github.com/tuanta7/ekko/actions/workflows/ci.yml/badge.svg)](https://github.com/tuanta7/ekko/actions/workflows/ci.yml)

A floating, always-on-top overlay that transcribes what your machine hears locally.

## Requirements

- Go, Node, and [Wails v3](https://v3.wails.io)
- A C/C++ compiler and `cmake` for whisper.cpp
- FFmpeg plus an audio backend: PulseAudio/PipeWire or AVFoundation

The source list comes from `pactl` on Linux and `ffmpeg -f avfoundation -list_devices` on macOS. whisper.cpp runs on CUDA where available, and on Metal
on Apple silicon.

## Quick start

Clone the repository

```sh
git clone https://github.com/tuanta7/ekko.git
cd ekko
git submodule update --init --recursive
```

Start development server

```sh
make setup           # system deps, Wails v3, whisper.cpp
make download-model  # tiny.en-q5_1 into assets/ggml/
make dev             # or: make build && make run
```

With another model

```sh
make download-model MODEL=base
make dev MODEL=base
```

See `whisper/models/README.md` for the full list. Downloaded `.bin` files are ignored by Git.

![Ekko](./assets/demo.png)
