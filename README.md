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

![Ekko](demo.png)