#!/bin/bash

cd ./whisper/bindings/go || { echo "Failed to change directory to ../whisper/bindings/go"; exit 1; }

# Metal is on by default on macOS; 
# CUDA is the Linux/NVIDIA path.
if [[ "$(uname -s)" == "Darwin" ]]; then
  make whisper
else
  GGML_CUDA=1 make whisper
fi
