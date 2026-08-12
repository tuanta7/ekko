# Ekko

![Status](https://img.shields.io/badge/status-development-orange)
![Language](https://img.shields.io/badge/lang-Go-blue)
[![CI](https://github.com/tuanta7/ekko/actions/workflows/ci.yml/badge.svg)](https://github.com/tuanta7/ekko/actions/workflows/ci.yml)

A floating, always-on-top overlay that transcribes what your machine hears locally.

## Requirements

- Linux with PulseAudio/PipeWire (`pactl`) and `ffmpeg`
- GTK4 + WebKitGTK 6.0, Go, Node, and [Wails v3](https://v3.wails.io)
- A C/C++ compiler and `cmake` for whisper.cpp

```txt
pactl (pick source) → ffmpeg (16 kHz mono f32 PCM) → chunker (energy VAD) → whisper.cpp → UI
```

## Quick start

```sh
git clone https://github.com/tuanta7/ekko.git
cd ekko
git submodule update --init --recursive

make setup           # system deps, Wails v3, whisper.cpp
make download-model  # tiny.en-q5_1 into assets/ggml/
make dev             # or: make build && make run
```

## Configuration

| Variable     | Default        | Purpose                                                                                    |
| ------------ | -------------- | ------------------------------------------------------------------------------------------ |
| `EKKO_MODEL` | `tiny.en-q5_1` | Model to load, from `assets/ggml/ggml-<name>.bin`, falling back to `/usr/share/ekko/ggml/` |

Other models work the same way, download one, then run with it:

```sh
make download-model MODEL=base.en-q5_1
make dev MODEL=base.en-q5_1
```

See `whisper/models/README.md` for the full list. Downloaded `.bin` files are
ignored by Git.

## Packaging (Linux)

Build the binary, then create a package in `bin/`:

```sh
make build
wails3 tool package -name ekko -format deb -config ./build/linux/nfpm/nfpm.yaml -out ./bin
```

Swap `-format deb` for `rpm` or `archlinux`.

The package ships the binary, the model, the icon and the `.desktop` entry; GTK, WebKit, `ffmpeg` and `pulseaudio-utils` are pulled in as dependencies.

```sh
sudo apt install ./bin/ekko.deb
sudo apt remove ekko
```

![Ekko](./assets/demo.png)
