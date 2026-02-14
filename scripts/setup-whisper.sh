#!/bin/bash

# Get the absolute path of the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Set up paths relative to project root
WHISPER_CPP_DIR="$PROJECT_DIR/whisper"
WHISPER_BUILD_DIR="$PROJECT_DIR/whisper/build_go"

echo -e "${GREEN}[INFO]${NC} Setting up environment..."
echo "Project root: $PROJECT_DIR"

# Export environment variables
export C_INCLUDE_PATH="$WHISPER_CPP_DIR/include:$WHISPER_CPP_DIR/ggml/include"
export LIBRARY_PATH="$WHISPER_BUILD_DIR/src:$WHISPER_BUILD_DIR/ggml/src"

echo -e "${YELLOW}"[INFO]"${NC}" Environment variables set:
echo "C_INCLUDE_PATH: $C_INCLUDE_PATH"
echo "LIBRARY_PATH: $LIBRARY_PATH"

# If this script is being sourced, don't exit
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then    
  echo -e "${YELLOW}[WARN]${NC} This script should be sourced, not executed directly."    
  echo "       Use: source ./setup-whisper.sh"
fi