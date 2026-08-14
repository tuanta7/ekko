#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"

# See whisper/models/README.md for the full model list
# Default to the lightest one: tiny.en-q5_1
MODEL_NAME="${1:-${EKKO_MODEL:-tiny.en-q5_1}}"
MODEL_DIR="$PROJECT_DIR/assets/ggml"
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
