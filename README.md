# Ekko

<img align="right" src="build/appicon.png" alt="Ekko Mascot" width="150"/>

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

For a production build:

```sh
make build
```

![Ekko](demo.png)
