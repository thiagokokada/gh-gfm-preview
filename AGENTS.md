# Repository Guidelines

## Project Structure & Module Organization
`main.go` is the entrypoint and delegates CLI setup to `cmd/cli.go`. Core
runtime logic lives under `internal/`: `internal/server` serves and renders
Markdown, `internal/app` holds reusable Markdown and filesystem helpers,
`internal/browser` manages browser launching, `internal/watcher` handles file
watching, and `internal/assert` holds test helpers. Static frontend assets live
in `internal/server/static/`, generated vendor files in
`internal/server/static/generated/`, and sample Markdown fixtures in
`testdata/`.

## Build, Test, and Development Commands
Use standard Go commands from the repository root:

- `go build` builds the local binary.
- `go run . README.md` starts the preview server against a file.
- `go generate ./...` regenerates checked-in frontend assets via
  `internal/server/_tools/generate-assets.go`.
- `GOCACHE=/tmp/go-build CGO_ENABLED=0 go test ./...` runs the full test suite
  in restricted environments.
- `golangci-lint run` runs the configured Go linters; `nix develop` provides a
  shell with tooling preinstalled.
- `./jslint.sh internal/server/static/script.js` checks browser JavaScript with
  the pinned JSLint version.

## Coding Style & Naming Conventions
Follow idiomatic Go and keep files `gofmt`-clean. This repository also enables
`gofumpt`, `goimports`, `gci`, and a large `golangci-lint` ruleset, so prefer
small functions, explicit error handling, and standard import grouping. Use
lowercase package names, `CamelCase` for exported identifiers, and descriptive
flag names such as `--directory-listing-show-extensions`. Keep generated assets
in `internal/server/static/generated/`; edit the generator or source inputs
instead of hand-editing generated files. In server code, treat URL paths as
slash-based paths: use `path` for URLs, breadcrumbs, and file-tree links, and
reserve `filepath` for host filesystem paths.

## Filesystem Boundaries
Directory mode is rooted with `os.Root` in `internal/server`; do not
reintroduce plain-path reads there. Shared directory helpers in `internal/app`
use `fs.FS` (`FindReadmeFS`, `ListDirectoryContentsFS`) so rooted and
non-rooted callers can reuse them safely. If you need a host path, keep the
conversion local and only for operations that truly require it, such as watcher
registration.

## Testing Guidelines
Place tests next to the code as `*_test.go`; existing coverage is package-local
and table-driven where useful. Reuse fixtures from `testdata/` for Markdown
rendering and directory-listing cases. Prefer handler-level tests with
`httptest.NewRecorder` and `httptest.NewRequest` for `/__/md` and
directory-mode regressions; they cover path normalization, README lookup, and
escape rejection without requiring loopback sockets. Tests that use
`httptest.NewServer` need a less restricted environment.

## Commit & Pull Request Guidelines
Recent history uses short, imperative subjects such as `Improve tests`, `Fix
JSLint issues in map renderer`, and occasional conventional prefixes like
`fix:` or `chore:`. Keep commit titles concise, present tense, and focused on
one change. Pull requests should explain user-visible behavior, mention
affected flags or rendering modes, link the issue when applicable, and include
screenshots or terminal examples for UI/rendering changes.
