# Repository Guidelines

## Project Structure & Module Organization

Ekko is a Wails v3 desktop app. Go application entrypoint and service registration live in `main.go`; domain code is under `internal/`, with services in `internal/service`, audio capture in `internal/adapter/ffmpeg`, Whisper integration in `internal/adapter/whisper`, and logging in `internal/logger`. The React/TypeScript UI is in `frontend/src`, static assets and global CSS are in `frontend/public`, and generated Wails bindings are in `frontend/bindings`. Build and packaging assets live in `build/`; setup helpers are in `scripts/`. The `whisper/` directory is vendored upstream code and should be treated as third-party unless explicitly updating Whisper.

## Build, Test, and Development Commands

- `make setup`: installs Wails, system audio dependencies, and builds Whisper support.
- `make dev`: sources `scripts/setup-env.sh` and runs `wails3 dev`.
- `make build`: builds the desktop app through Wails.
- `task dev`: runs Wails dev mode using `build/Config.yml` and the configured Vite port.
- `task build`: builds for the current OS via the platform Taskfile.
- `go test -race ./internal/...`: runs the pure-Go test suite used by CI.
- `go vet ./internal/...`: runs static checks used by CI.
- `cd frontend && npm run build`: type-checks and builds the frontend.

Run `source ./scripts/setup-env.sh` before Go commands that need local Whisper headers or libraries.

## Coding Style & Naming Conventions

Format Go with `gofmt`; keep package names short and lowercase. Put tests beside code as `*_test.go`, and name tests like `TestAudioChunkerEmitsPartialAndFinal`. Use exported names only for cross-package or Wails-bound APIs. Frontend code uses React function components, TypeScript types, and project-local generated bindings from `frontend/bindings`; keep component state and event names explicit.

## Testing Guidelines

Add focused Go unit tests under `internal/...` for service, chunking, and adapter behavior. Prefer deterministic tests that avoid real audio devices, network access, or model downloads. CI currently runs `go vet ./internal/...` and `go test -race ./internal/...`; keep new tests compatible with those commands.

## Commit & Pull Request Guidelines

Recent commits use short, imperative, lowercase summaries such as `update app logo` or `remove unused build`. Keep commits scoped to one change. Pull requests should describe the behavior change, list verification commands, link related issues when available, and include screenshots or screen recordings for UI changes.

## Agent-Specific Instructions

Do not edit generated `frontend/bindings` by hand unless regenerating them with Wails. Avoid unrelated changes in `whisper/` and build artifacts. Preserve existing uncommitted user work while making targeted edits.
