# Ekko

<img align="right" src="build/appicon.png" alt="Ekko Mascot" width="200"/>

![Status](https://img.shields.io/badge/status-development-orange)
![Language](https://img.shields.io/badge/lang-Go-blue)
![License](https://img.shields.io/badge/license-MIT-green)
[![CI](https://github.com/tuanta7/ekko/actions/workflows/ci.yml/badge.svg)](https://github.com/tuanta7/ekko/actions/workflows/ci.yml)

## Setup

Clone this repository with whisper-cpp submodule

```sh
git clone https://github.com/tuanta7/ekko.git
git submodule update --init --recursive
```

Download Whisper model before starting transcription.

```sh
# tiny.en-q5_1.bin
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

## Packaging (Linux)

Build the binary first, then create a `.deb` in `bin/`:

```sh
make build
wails3 tool package -name ekko -format deb -config ./build/linux/nfpm/nfpm.yaml -out ./bin
```

Swap `-format deb` for `rpm` or `archlinux` for the other distros.

The package ships the binary, the `tiny.en-q5_1` model (to `/usr/share/ekko/ggml/`), the icon and the `.desktop` entry; GTK, WebKit, `ffmpeg` and `pulseaudio-utils` are pulled in as dependencies.

Install and remove:

```sh
sudo apt install ./bin/ekko.deb
sudo apt remove ekko
```

![Ekko](./demo.png)
