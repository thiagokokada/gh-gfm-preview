# Repository Guidelines

## Project Structure & Module Organization
`main.go` is the entrypoint and delegates CLI setup to `cmd/cli.go`. Core runtime logic lives under `internal/`: `internal/server` serves and renders Markdown, `internal/app` handles application flow, `internal/browser` manages browser launching, `internal/watcher` handles file watching, and `internal/assert` holds test helpers. Static frontend assets and generated vendor files live in `internal/server/static/` and `internal/server/static/generated/`. Sample Markdown fixtures live in `testdata/`.

## Build, Test, and Development Commands
Use standard Go commands from the repository root:

- `go build` builds the local binary.
- `go run . README.md` starts the preview server against a file.
- `go generate ./...` regenerates checked-in frontend assets via `internal/server/_tools/generate-assets.go`.
- `GOCACHE=/tmp/go-build CGO_ENABLED=0 go test ./...` runs the full test suite in restricted environments.
- `golangci-lint run` runs the configured Go linters; `nix develop` provides a shell with tooling preinstalled.
- `./jslint.sh internal/server/static/script.js` checks browser JavaScript with the pinned JSLint version.

## Coding Style & Naming Conventions
Follow idiomatic Go and keep files `gofmt`-clean. This repository also enables `gofumpt`, `goimports`, `gci`, and a large `golangci-lint` ruleset, so prefer small functions, explicit error handling, and standard import grouping. Use lowercase package names, `CamelCase` for exported identifiers, and descriptive flag names such as `--directory-listing-show-extensions`. Keep generated assets in `internal/server/static/generated/`; edit the generator or source inputs instead of hand-editing generated files.

## Testing Guidelines
Place tests next to the code as `*_test.go`; existing coverage is package-local and table-driven where useful. Reuse fixtures from `testdata/` for Markdown rendering and directory-listing cases. Server tests open local listeners, so if sandboxed execution blocks sockets, rerun them in a less restricted shell before concluding a regression.

## Commit & Pull Request Guidelines
Recent history uses short, imperative subjects such as `Improve tests`, `Fix JSLint issues in map renderer`, and occasional conventional prefixes like `fix:` or `chore:`. Keep commit titles concise, present tense, and focused on one change. Pull requests should explain user-visible behavior, mention affected flags or rendering modes, link the issue when applicable, and include screenshots or terminal examples for UI/rendering changes.
