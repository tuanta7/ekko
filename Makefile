.PHONY: setup-wails-v3 setup-audio dev
SHELL := /bin/bash
.ONESHELL:

MODEL ?= ggml-base

setup-wails-v3:
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

dev: setup-audio
	source ./scripts/setup-env.sh
	wails3 dev
