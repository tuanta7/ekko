#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

MODEL="${1:-${MODEL:-ggml-base}}"
MODEL_DIR="${2:-${MODEL_DIR:-ggml}}"
DOWNLOADER="$PROJECT_DIR/whisper/models/download-ggml-model.sh"

usage() {
  cat <<EOF
Usage: $0 [model] [model_dir]

Downloads a Whisper ggml model for Ekko.

Examples:
  $0 base
  $0 ggml-base
  MODEL=medium-q5_0 make download-model

Defaults:
  model:     ggml-base
  model_dir: ggml
EOF
}

case "$MODEL" in
  -h|--help)
    usage
    exit 0
    ;;
esac

if [[ ! -x "$DOWNLOADER" ]]; then
  echo "Missing Whisper downloader: $DOWNLOADER" >&2
  exit 1
fi

if [[ "$MODEL_DIR" != /* ]]; then
  MODEL_DIR="$PROJECT_DIR/$MODEL_DIR"
fi

model_name="$(basename "$MODEL")"
model_name="${model_name%.bin}"
model_name="${model_name#ggml-}"

mkdir -p "$MODEL_DIR"

target="$MODEL_DIR/ggml-$model_name.bin"
if [[ -f "$target" ]]; then
  echo "Model already exists: $target"
  exit 0
fi

"$DOWNLOADER" "$model_name" "$MODEL_DIR"

echo "Model ready: $target"
echo "Use model name: ggml-$model_name"
