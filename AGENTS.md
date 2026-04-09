# Repository Guidelines

## Project Structure & Module Organization

- `cmd/ekko/`: CLI entrypoint (`ekko run`) and flag wiring.
- `internal/`: application internals by domain:
  - `internal/ffmpeg`: audio capture/recording.
  - `internal/handler`: orchestration, scheduling, transcript ordering.
  - `internal/transport/console`: Bubble Tea TUI.
  - `internal/transport/browser`: HTTP/SSE web transport.
  - `internal/whisper`: local Whisper client wrapper.
- `pkg/`: shared utilities (`logger`, `silent`, package-level helpers).
- `scripts/`: setup/build helpers (Whisper env + native build).
- `web/`: static web UI assets.
- `ggml/`: local Whisper model files (for example `ggml-base.bin`).

## Build, Test, and Development Commands

- `make install`: install system dependencies via `scripts/install.sh`.
- `make build`: build native Whisper dependencies.
- `make dev`: run TUI mode (`MODEL=ggml-base` by default).
- `make dev MODEL=ggml-small-q5_1`: run TUI with a different model.
- `make dev-web`: run web mode on `:8080`.
- `go test ./internal/transport/console ./internal/handler ./cmd/ekko`: fast checks that do not require native linker setup.
- `go test ./...`: full suite; may fail without Whisper native libs (`-lwhisper`, `-lggml`).

## Coding Style & Naming Conventions

- Language: Go. Keep code `gofmt`-formatted before submitting.
- Package names are short, lowercase, and domain-based (`handler`, `console`).
- Exported identifiers use `PascalCase`; internal helpers use `camelCase`.
- Keep transport concerns in `internal/transport/*`; avoid leaking UI logic into core handler/ffmpeg layers.

## Testing Guidelines

- Use Go’s `testing` package with `*_test.go` naming.
- Prefer table-driven tests for orchestration/state logic.
- Place tests beside the code under test (for example `internal/ffmpeg/recorder_test.go`).
- For features touching TUI/web behavior, include manual verification steps in the PR description.

## Commit & Pull Request Guidelines

- Follow existing history style: short, imperative summaries (for example `Add ...`, `Refactor ...`, `Fix ...`).
- Keep commits focused by concern (CLI, handler, transport, docs).
- PRs should include:
  - What changed and why.
  - How it was tested (exact commands).
  - Screenshots/GIFs for web or TUI-visible changes.
  - Linked issue/context when applicable.
