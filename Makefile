.PHONY: install build dev start
SHELL := /bin/bash
.ONESHELL:

MODEL ?= ggml-base

install:
	./scripts/install.sh

build:
	./scripts/build-whisper.sh

dev: install build
	source ./scripts/setup-whisper.sh
	go build -o ekko ./cmd/ekko
	./ekko run --model $(MODEL)

dev-web: install build
	source ./scripts/setup-whisper.sh
	go build -o ekko ./cmd/ekko
	./ekko run --web --addr :8080 --model $(MODEL)

start:
	source ./scripts/setup-whisper.sh
	./ekko run
