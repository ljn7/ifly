# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.1-plugin] - 2026-04-21

### Fixed
- **PowerShell bypass** (security). Guard hook only watched `Bash`; Claude Code
  on Windows could pivot to `PowerShell(Remove-Item ...)` and slip past the
  boundary check. Added `PowerShell` to `hooks.json` matcher and a new
  `_handle_powershell` that mirrors `_handle_bash`: always-blocked patterns
  (`Format-Volume`, `Stop-Computer`/`Restart-Computer`, `Invoke-Expression`/`iex`,
  `Remove-Item` on drive root), and a destructive-cmdlet boundary check covering
  `Remove-Item`, `Move-Item`, `Rename-Item`, `Clear-Content`, `Clear-Item`,
  `Set-Acl` and their aliases (`rm`, `del`, `rd`, `rmdir`, `erase`, `ri`, `mv`,
  `move`, `mi`, `ren`, `rni`, `clc`, `cli`). `-Path`/`-Destination` quoted values
  are unquoted before the path check. 12 new guard tests.
- **Windows drive-letter paths resolved as relative** (correctness bug that
  false-positive-blocked legitimate Edit/Write of in-project files on Windows).
  `ifly_resolve_path` now normalizes `C:/foo`, `C:\foo`, `c:\`, and drive-
  relative `C:foo` to MSYS `/c/foo` form before the absolute/relative check.
  6 new path_resolve tests.

### Added
- 12 guard tests for PowerShell coverage (block outside, allow inside, alias
  coverage for `rm`/`del`, always-blocked for format/halt/iex, off-level bypass).
- 6 path_resolve tests for Windows drive-letter edge cases.

## [0.1.0-cli] - 2026-04-21

### Added
- `cli/` — Go companion CLI (`github.com/ljn7/ifly/cli`) with cobra root,
  embedded plugin tree via `go:embed`, and 10 internal packages:
  `paths`, `detect`, `config`, `state`, `egg`, `installer`, `updater`,
  `sandbox` (Linux build-tag), plus the `cmd/` subcommand files.
- Subcommands:
  - `ifly install` — bubbletea TUI installer + `--no-tui` headless mode
    with `--mode`, `--guard`, `--lockdown`, `--overwrite` flags. Walks the
    embedded plugin tree into the detected Claude config dir, writes a
    global `config.yaml`, runs post-install verification.
  - `ifly init` — writes a starter `.ifly.yaml` template, `--force` to
    overwrite.
  - `ifly config show|get|set` — merged-config read + scalar writes to
    global or `--project` scope with validation (mode, guard.level,
    lockdown, telemetry.easter_egg).
  - `ifly status` — merged config summary with rotating easter-egg footer
    (`💜 ifly` ~10% of invocations; respects `telemetry.easter_egg: false`).
  - `ifly version [--love]` — version banner; hidden `--love` flag reveals
    the tagline.
  - `ifly update [--dry-run]` — self-update from the latest GitHub release
    with SHA256 sidecar verification.
  - `ifly sandbox` — Linux-only filesystem-namespace wrapper for `claude`
    (bwrap preferred, unshare fallback). Stub on macOS/Windows points at
    `level: strict`.
- Config-merge semantics in Go matching the bash hook: scalars replace,
  lists union-dedupe, `guard.tools.*` merge per-key with `*bool` so "unset"
  stays distinct from "explicitly false", lockdown blocks loosening.
- Unit tests — 10 packages, all pass. Filesystem tests use `t.TempDir()`.
  XDG-sensitive tests skip on Windows (APPDATA path only). Linux sandbox
  tests gated by `//go:build linux`.
- Defaults drift test — shells out to `hooks/lib/parse_defaults.sh` and
  compares scalar values against the Go unmarshal. Fails CI on drift.
- `Makefile` — `embed` (portable cp-based, rsync-free), `build` with
  ldflags version injection, `test` (bash + Go), `lint` (shellcheck +
  golangci-lint), `release` (cross-compile 5 targets + sha256 sidecars +
  plugin tarball), `install`, `clean`.
- `.github/workflows/ci.yml` — matrix (ubuntu, macos) for plugin tests +
  shellcheck, (ubuntu, macos, windows) for Go tests + lint, separate
  defaults-drift job.
- `.github/workflows/release.yml` — on `v*.*.*` tags, asserts VERSION
  matches, runs full test suite, builds release artifacts, creates GitHub
  Release with cross-platform binaries and plugin tarball.
- `docs/cli.md` — CLI command reference with config-precedence rules and
  per-OS notes (Windows bash/jq requirement, Linux sandbox setup).

### Deviations from the Plan B text (all documented in commits)
- `rsync` unavailable on Windows Git Bash — Makefile embed target uses
  `cp -r`. Shipped CI also uses `cp -r` for the embed step.
- `hooks/guard.sh`/`session_start.sh` were renamed to extensionless
  `guard`/`session-start` during Plan A's smoke-test fix; Makefile,
  CI workflow, and installer verify step all reference the new names
  (installer `findGuardHook` accepts either for compatibility).
- `parse_defaults.sh` is a stdin→stdout filter, not a sourceable library
  with an `ifly_defaults_get` helper the plan assumed. Drift test was
  rewritten to invoke the filter once and parse its key=value output.
- `sandbox` subcommand uses `DisableFlagParsing=true`, which sends
  `--help` into args; added explicit `--help/-h` short-circuit so cobra's
  usage screen still shows.
- `state.Save` preserves a caller-set `UpdatedAt` and only stamps `now`
  when zero (plan's "always stamp" was wall-clock-flaky in the round-trip
  test).

### Known limitations
- Integration tests for full `ifly install` flow skip on Windows because
  the Go `paths.GlobalConfigDir()` keys off `%APPDATA%` there, and
  overriding it per-test is fragile. CI on Linux/macOS runs the full
  matrix.
- `make build` / `make test` commands rely on `make` being installed. On
  Windows without `make`, use `go build` / `go test ./...` directly.

## [0.1.0-plugin] - 2026-04-19

### Added
- Plugin manifest and shipped defaults (`.claude-plugin/plugin.json`, `defaults.yaml`).
- PreToolUse guard hook (`hooks/guard.sh`) with four handlers:
  - `_handle_bash` — always-blocked patterns (`git push --force`, `rm -rf /`,
    `dd if=`, `mkfs.*`, `DROP TABLE`, `eval`, `bash -c`, ...), config-driven
    `blocked_commands`, shell-network blocking in strict mode, and per-subcommand
    destructive-verb path checks (rm/rmdir/mv/chmod/chown/unlink/shred/truncate)
    with virtual `cd` tracking across chain operators.
  - `_handle_path` — Edit/Write/MultiEdit/NotebookEdit boundary enforcement
    against `CLAUDE_PROJECT_DIR + additional_dirs`, with per-tool toggles and
    open-mode warn-and-allow.
  - `_handle_sensitive_path` — strict-only Read/Glob/Grep gate against
    `sensitive_paths` prefixes (default: ~/.ssh, ~/.aws, ~/.config/gh,
    /etc/shadow, /etc/sudoers).
  - `_handle_network_allowlist` — strict-only WebFetch/WebSearch URL allowlist
    with exact + suffix host matching; non-HTTP URLs blocked outright.
- SessionStart hook (`hooks/session_start.sh`) injecting active mode + guard
  level + skill pointer; honors `~/.config/ifly/state.yaml` `active_mode`
  override.
- Reusable bash modules in `hooks/lib/`:
  - `parse_defaults.sh` — YAML-subset reader (scalars, nested maps, lists).
  - `path_resolve.sh` — tilde/HOME/.. expansion + boundary check.
  - `split_command.sh` — chain splitter with `cd` tracking.
  - `config.sh` — merged-config loader (defaults > global > project, env override,
    list union, lockdown enforcement with strict=3/project=2/open=1/off=0 ranks).
- Five verbosity-mode skills:
  - `ifly-mode-silent` — direct answer only; no preamble or summary.
  - `ifly-mode-minimal` — one-line summaries; no headers, no bullets in prose.
    (Default.)
  - `ifly-mode-normal` — no-op passthrough.
  - `ifly-mode-verbose` — alternatives + tradeoffs + change summaries; respects
    explicit "briefly" override.
  - `ifly-mode-caveman` — maximally compressed prose; idiomatic generated code.
- Three slash commands: `/ifly:mode`, `/ifly:guard`, `/ifly:status`. Status
  command includes a rotating easter-egg footer (`💜 ifly`, ~10%).
- Test suite — 7 files, 111 assertions, all passing:
  - Test harness (`tests/lib/assert.sh`).
  - Unit tests for each `hooks/lib/` module.
  - End-to-end guard tests (`tests/test_guard.sh`, 60 functions covering all
    spec §7.1 groups: path resolution, boundary, chain splitting, destructive
    patterns, network, sensitive reads, config merge, lockdown, edge cases).
  - SessionStart hook tests.
- Latency benchmark (`tests/bench_guard.sh`) — 50 ms budget, CI-gated on
  Linux/macOS.
- Documentation: README quickstart, `docs/plugin.md` reference,
  `docs/security.md` honest threat model, `docs/smoke-test.md` v0.1.0
  notes.
- `.ifly.yaml.example` per-project config starter.

### Known limitations (deliberate scope)
- Single-prefix `sudo`/`doas` strip — chained sudo after `&&`/`;` is not
  re-stripped per subcommand (documented in `docs/security.md`).
- Whitespace-tokenized argument parsing — quoted paths with spaces tokenize
  incorrectly.
- No hostname extraction from shell `curl`/`wget` — strict mode blocks the
  verbs outright.
- Naive JSON escaping in `_emit_block` (only `"` and `\`).
