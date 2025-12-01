.PHONY: env-example install-deps setup-whisper
SHELL := /bin/bash

env-example:
	awk -F'=' 'BEGIN {OFS="="} \
    	/^[[:space:]]*#/ {print; next} \
    	/^[[:space:]]*$$/ {print ""; next} \
    	NF>=1 {gsub(/^[[:space:]]+|[[:space:]]+$$/, "", $$1); print $$1"="}' .env > .env.example
	echo ".env.example generated successfully."

install-deps:
	./scripts/install.sh

setup-whisper:
	source ./scripts/setup-whisper.sh