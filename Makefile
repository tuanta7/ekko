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
	./ekko run --web

start:
	source ./scripts/setup-whisper.sh
	go build -o ekko ./cmd/ekko
	./ekko run

env:
	awk -F'=' 'BEGIN {OFS="="} \
    	/^[[:space:]]*#/ {print; next} \
    	/^[[:space:]]*$$/ {print ""; next} \
    	NF>=1 {gsub(/^[[:space:]]+|[[:space:]]+$$/, "", $$1); print $$1"="}' .env > .env.example
	echo ".env.example generated successfully."