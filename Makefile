.PHONY: setup-wails-v3 setup-audio setup-whisper download-model setup dev build
SHELL := /bin/bash

# macOS ships make 3.81, which ignores .ONESHELL — every recipe line is its own
# shell, so anything sourcing setup-env.sh must chain onto the same line.
ENV := source ./assets/scripts/setup-env.sh &&

setup: setup-wails3 setup-audio setup-whisper

ifeq ($(shell uname -s),Darwin)

setup-wails3:
	xcode-select --install || true # WebKit comes from the system, no GTK needed
	go install -v github.com/wailsapp/wails/v3/cmd/wails3@latest
	wails3 doctor

# blackhole-2ch is the loopback device that stands in for Pulse's monitor sinks.
setup-audio:
	brew install ffmpeg cmake
	brew install --cask blackhole-2ch

else

setup-wails3:
	sudo apt update
	sudo apt install pkg-config gcc libgtk-4-dev libwebkitgtk-6.0-dev
	go install -v github.com/wailsapp/wails/v3/cmd/wails3@latest
	wails3 doctor

setup-audio:
	sudo apt update
	sudo apt install pulseaudio-utils ffmpeg cmake
	sudo apt-get install libsdl2-dev

endif

setup-whisper:
	./assets/scripts/setup-whisper.sh

# make download-model MODEL=base.en-q5_1
MODEL ?= tiny.en-q5_1

download-model:
	./assets/scripts/download-model.sh $(MODEL)

dev:
	$(ENV) EKKO_MODEL=$(MODEL) wails3 dev

build:
	$(ENV) wails3 build

run:
	$(ENV) EKKO_MODEL=$(MODEL) ./bin/ekko
