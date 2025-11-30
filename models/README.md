# Models

Download whisper ggml models from [huggingface.co/ggerganov/whisper.cpp](https://huggingface.co/ggerganov/whisper.cpp/tree/main) and put them here.

## Building the Whisper Library

The Go bindings depend on the compiled C/C++ library. A build script has already been provided in the Go bindings directory.

```shell
# Set up the environment
source setup-whisper.sh 

[INFO] Setting up environment...
Project root: /home/tuanta/Projects/transcript
[INFO] Environment variables set:
C_INCLUDE_PATH: /home/tuanta/Projects/transcript/whisper/include:/home/tuanta/Projects/transcript/whisper/ggml/include
LIBRARY_PATH: /home/tuanta/Projects/transcript/whisper/build_go/src:/home/tuanta/Projects/transcript/whisper/build_go/ggml/src
```

Install make and cmake, then build the library.

```shell
# Install dependencies
sudo apt install cmake

# Build the libwhisper.a library
cd whisper/bindings/go
make whisper
```