#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

MODEL_NAME="tiny.en-q5_1"
MODEL_DIR="$PROJECT_DIR/ggml"
DOWNLOADER="$PROJECT_DIR/whisper/models/download-ggml-model.sh"

if [[ ! -x "$DOWNLOADER" ]]; then
  echo "Missing Whisper downloader: $DOWNLOADER" >&2
  exit 1
fi

mkdir -p "$MODEL_DIR"

target="$MODEL_DIR/ggml-$MODEL_NAME.bin"
if [[ -f "$target" ]]; then
  echo "Model already exists: $target"
  exit 0
fi

"$DOWNLOADER" "$MODEL_NAME" "$MODEL_DIR"

echo "Model ready: $target"
