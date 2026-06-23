# Ekko

<img align="right" src="build/appicon.png" alt="Ekko Mascot" width="150"/>

![Status](https://img.shields.io/badge/status-development-orange)
![Language](https://img.shields.io/badge/lang-Go-blue)
![License](https://img.shields.io/badge/license-MIT-green)
[![CI](https://github.com/tuanta7/ekko/actions/workflows/ci.yml/badge.svg)](https://github.com/tuanta7/ekko/actions/workflows/ci.yml)

## Setup

Download a Whisper ggml model before starting transcription:

Ekko loads models from `ggml/<model>.bin`; the default app model is `ggml/ggml-base.bin`.

```sh
make download-model
make download-model MODEL=medium-q5_0
```

The model name may be passed with or without the `ggml-` prefix. Downloaded `.bin` files are ignored by Git.

See all available models at [huggingface.co/ggerganov/whisper.cpp](https://huggingface.co/ggerganov/whisper.cpp/tree/main)

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
