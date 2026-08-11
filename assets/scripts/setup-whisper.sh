#!/bin/bash

cd ./whisper/bindings/go || { echo "Failed to change directory to ../whisper/bindings/go"; exit 1; }
GGML_CUDA=1 make whisper