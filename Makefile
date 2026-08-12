.PHONY: setup-wails-v3 setup-audio setup-whisper download-model setup dev build
SHELL := /bin/bash
.ONESHELL:

setup: setup-wails3 setup-audio setup-whisper

ifeq ($(shell uname -s),Darwin)

setup-wails3:
	xcode-select --install || true # WebKit comes from the system, no GTK needed
	go install -v github.com/wailsapp/wails/v3/cmd/wails3@latest
	wails3 doctor # Expected: Your system is ready for Wails development!

# blackhole-2ch is the loopback device that stands in for Pulse's monitor sinks.
setup-audio:
	brew install ffmpeg cmake
	brew install --cask blackhole-2ch

else

setup-wails3:
	sudo apt update
	sudo apt install pkg-config gcc libgtk-4-dev libwebkitgtk-6.0-dev
	go install -v github.com/wailsapp/wails/v3/cmd/wails3@latest
	wails3 doctor # Expected: Your system is ready for Wails development!

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
	source ./assets/scripts/setup-env.sh
	EKKO_MODEL=$(MODEL) wails3 dev

build:
	source ./assets/scripts/setup-env.sh
	wails3 build

run:
	source ./assets/scripts/setup-env.sh
	EKKO_MODEL=$(MODEL) ./bin/ekko
