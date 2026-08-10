.PHONY: setup-wails-v3 setup-audio setup-whisper download-model setup dev build
SHELL := /bin/bash
.ONESHELL:

setup: setup-wails3 setup-audio setup-whisper

setup-wails3:
	sudo apt update
	sudo apt install pkg-config gcc libgtk-4-dev libwebkitgtk-6.0-dev
	go install -v github.com/wailsapp/wails/v3/cmd/wails3@latest
	wails3 doctor # Expected: Your system is ready for Wails development!

setup-audio:
	sudo apt update
	sudo apt install pulseaudio-utils ffmpeg cmake
	sudo apt-get install libsdl2-dev

setup-whisper:
	./scripts/setup-whisper.sh

# make download-model MODEL=base.en-q5_1
MODEL ?= tiny.en-q5_1

download-model:
	./scripts/download-model.sh $(MODEL)

dev:
	source ./scripts/setup-env.sh
	EKKO_MODEL=$(MODEL) wails3 dev

build:
	source ./scripts/setup-env.sh
	wails3 build

run:
	source ./scripts/setup-env.sh
	EKKO_MODEL=$(MODEL) ./bin/ekko
