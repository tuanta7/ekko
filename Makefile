.PHONY: env deps whisper
SHELL := /bin/bash
.ONESHELL:

install:
	./scripts/install.sh

build: install
	./scripts/build-whisper.sh

dev:
	source ./scripts/setup-whisper.sh
	go build -o ekko ./cmd/ekko
	./ekko run --server :8080 --model ggml-medium-q5_0

start:
	source ./scripts/setup-whisper.sh
	go build -o ekko ./cmd/ekko
	./ekko run