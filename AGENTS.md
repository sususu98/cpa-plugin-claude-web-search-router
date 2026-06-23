# Repository Guidelines

Contributor guide for the **claude-web-search-router** CPA plugin (Claude Code built-in `web_search` routing).

## Project Structure & Module Organization

- `cmd/claude-web-search-router/` — plugin source (ModelRouter + Executor, CGO `c-shared` entry). Tests live beside code as `*_test.go`.
- `testdata/` — optional HTTP/JSON capture fixtures (e.g. `claude_code_web_search_capture.json`); missing files skip related tests.
- `dist/` — local build artifacts (`.dylib` / `.so` / `.dll`); not committed.
- `.github/workflows/build.yml` — CI matrix and tag releases (`v*`).
- `.codex/hooks.json` + `.codex/hooks/` — Codex lifecycle hooks that check CPA example parity and run `go test` / `go vet` after plugin edits.
- `README.md` / `README_CN.md` — user-facing config and behavior; keep config field names in sync with `pluginConfig` in `main.go`.

Upstream reference for ABI and behavior: `$CPA_SOURCE/examples/plugin/claude-web-search-router/go` (default `CPA_SOURCE=/Users/sususu/workspace/CLIProxyAPI`). Host requires **CLIProxyAPI v7.2.31+**. Plugin config key must stay **`claude-web-search-router`**.

## Build, Test, and Development Commands

```bash
go test ./...          # unit tests (cached OK)
go vet ./...           # static analysis
mkdir -p dist
CGO_ENABLED=1 go build -buildmode=c-shared \
  -o dist/claude-web-search-router.dylib \
  ./cmd/claude-web-search-router
```

Hook maintenance (optional):

```bash
.codex/hooks/verify_plugin_contract.sh
.codex/hooks/verify_after_edit.sh full
# Before tagging v* for Plugins Store:
PLUGIN_HOOK_STORE_CHECK=1 sh .codex/hooks/stop.sh
```

## Coding Style & Naming Conventions

- Go **1.26** module; use tabs for Go files per `gofmt`.
- Package name `main` under `cmd/claude-web-search-router`; unexported helpers use camelCase; exported symbols only where required for tests.
- Config YAML tags in `pluginConfig` use snake_case (`antigravity_model`, `require_web_search_only`).
- Prefer `sdk/cliproxy` and `sdk/pluginabi` APIs over copying `internal/` paths from CPA main repo unless intentionally diverging (document in PR).

## Testing Guidelines

- Framework: standard `testing` + table-driven tests in `cmd/claude-web-search-router`.
- Name tests `Test<Behavior>`; routing tests use `mustJSON` helpers and temporary registry clients where needed.
- Run full package tests before PR: `go test ./...` and `go vet ./...`.
- Do not change detection heuristics or fallback order without updating `README_CN.md` and adding/adjusting tests.

## Commit & Pull Request Guidelines

Recent history uses short, scoped prefixes: `ci:`, `fix:`, `docs:` (e.g. `fix: expose plugin metadata version from ldflags`). Keep subjects imperative and under ~72 characters.

PRs should include: what route/ABI/config changed, CPA example diff notes if behavior diverges, and confirmation that `go test ./...` and `go vet ./...` pass. For releases, tags must be `v*` (e.g. `v0.1.0`) to trigger multi-platform artifacts in Actions.

## Agent-Specific Instructions

- Do not edit the CPA main repository unless explicitly requested.
- Before changing routes, `pluginabi` method handling, or config fields, read the CPA example plugin and run `.codex/hooks/verify_plugin_contract.sh`.
- Default fallback chain: `antigravity → codex → xai → tavily`; override with a YAML list on `route` (scalar `fallback` = default chain); detection targets `web_search_20250305` / `web_search_20260209` on `claude` / `anthropic` requests.
