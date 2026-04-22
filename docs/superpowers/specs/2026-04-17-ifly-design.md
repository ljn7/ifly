# IFLy — Design Spec

**Status:** approved for implementation
**Date:** 2026-04-17
**Scope:** v0.1.0
**Successor skills:** `superpowers:writing-plans` → `superpowers:subagent-driven-development`

---

## 1. Overview

**IFLy** (*"I'm Feeling Lucky"*) is a Claude Code power-pack shipped as a single repo with two artifacts:

- A **Claude Code plugin** (bash hooks + Markdown skills + slash commands) that enforces safety boundaries and verbosity modes inside Claude Code.
- A companion **Go CLI** (`ifly`) that installs and manages the plugin, exposes config, and — on Linux only — offers real OS-level sandboxing via filesystem namespaces.

The two artifacts version-lockstep (one `VERSION` file at repo root). The CLI embeds the plugin tree via `go:embed` at build time. Users who only want the plugin can install it from the repo without touching the CLI; users who want `ifly install` ergonomics and/or Linux sandbox download the binary.

### 1.1 Branding invariants

- Casing: **`IFLy`** in display text (capital I, F, L, lowercase y). Never `IFLY`, `Ifly`, or `ifly` in prose.
- CLI binary: `ifly` (all lowercase, CLI convention).
- Plugin manifest name: `ifly` (lowercase, Claude Code convention).
- Slash-command prefix: `/ifly:` (lowercase).
- Public expansion: "I'm Feeling Lucky".
- Easter egg: `/ifly:status` and `ifly status` show a footer line. With probability ~10% (sampled via `$RANDOM % 10 == 0`), the footer is `💜 ifly`. If asked what it means, Claude responds `I'm Feeling Lucky, obviously.` and nothing more. The alternative reading is never written down in the repo.

### 1.2 Architectural decisions (decided during brainstorm)

| # | Decision | Rationale |
|---|---|---|
| D1 | **One spec, two implementation plans.** Design both together (shared config, shared branding); implement plugin first, CLI second. | Plugin is shippable without CLI; CLI design benefits from observing real plugin behavior. |
| D2 | **Mode activation: slash command activates skill, SessionStart hook restores across sessions.** State stored in `~/.config/ifly/state.yaml`. | Matches SKILL.md-per-mode layout from the spec; survives session restart. |
| D3 | **CLI embeds plugin tree via `go:embed`.** `ifly update` downloads a newer binary (= newer plugin). | Single static binary, offline install, unified versioning. |
| D4 | **Guard covers Bash, Edit, Write, MultiEdit, NotebookEdit always.** Read/Glob/Grep only in strict mode against a sensitive-paths blocklist. WebFetch/WebSearch gated by `allowed_network` allowlist in strict mode. User toggles per-tool at install. | Closes credential-exfiltration hole without forcing it in all modes. |
| D5 | **Config precedence: env > project > global > built-in defaults.** Scalars replace; lists (`additional_dirs`, `blocked_commands`, `allowed_network`, `sensitive_paths`) merge additively. `guard.level` freely overridable unless global has `lockdown: true`. | Ad-hoc env override, committed team convention, user defaults. |
| D6 | **Strict mode blocks ALL shell network commands** (`curl`, `wget`, `ssh`, `scp`, `sftp`, `nc`, `ncat`, `telnet`, `ftp`, `rsync`). `allowed_network` applies only to WebFetch/WebSearch tools. | Bash-command hostname parsing is adversarial; if Claude needs a URL, use WebFetch. |
| D7 | **License: MIT.** | Standard for Claude Code ecosystem. |

---

## 2. Repository Layout

```
ifly/
├── VERSION                              # single source of truth for version
├── LICENSE                              # MIT
├── README.md
├── CHANGELOG.md
├── defaults.yaml                        # shared defaults (bash hook + Go CLI both read)
├── .ifly.yaml.example                   # example project config
│
├── .claude-plugin/
│   └── plugin.json                      # plugin manifest
├── hooks/
│   ├── guard.sh                         # PreToolUse: boundary + command guard
│   ├── session_start.sh                 # SessionStart: inject active mode
│   └── lib/
│       ├── parse_defaults.sh            # bash YAML-subset reader
│       ├── path_resolve.sh              # symlink/~/.. resolution
│       └── split_command.sh             # chained-command splitter + cd tracking
├── skills/
│   ├── ifly-mode-silent/SKILL.md
│   ├── ifly-mode-minimal/SKILL.md
│   ├── ifly-mode-normal/SKILL.md
│   ├── ifly-mode-verbose/SKILL.md
│   └── ifly-mode-caveman/SKILL.md
├── commands/
│   ├── mode.md
│   ├── guard.md
│   └── status.md
├── tests/
│   ├── test_guard.sh                    # ≥50 test cases
│   ├── bench_guard.sh                   # latency benchmark
│   └── skills/
│       ├── mode-silent.scenarios.md
│       ├── mode-minimal.scenarios.md
│       ├── mode-normal.scenarios.md
│       ├── mode-verbose.scenarios.md
│       └── mode-caveman.scenarios.md
│
├── cli/
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── plugin_fs.go                     # //go:embed directives for plugin tree
│   ├── cmd/
│   │   ├── root.go
│   │   ├── install.go
│   │   ├── init.go
│   │   ├── config.go
│   │   ├── status.go
│   │   ├── update.go
│   │   ├── sandbox.go
│   │   └── version.go
│   └── internal/
│       ├── detect/{os.go, claude.go}
│       ├── paths/paths.go
│       ├── config/{schema.go, load.go, merge.go}
│       ├── state/state.go
│       ├── installer/{install.go, verify.go, uninstall.go}
│       ├── updater/updater.go
│       └── sandbox/sandbox.go           # //go:build linux
│
├── docs/
│   ├── plugin.md                        # plugin reference
│   ├── cli.md                           # CLI reference
│   ├── security.md                      # honest threat model
│   └── superpowers/specs/               # design specs (this file lives here)
│
├── Makefile
└── .github/
    └── workflows/
        ├── ci.yml                       # tests on every PR
        └── release.yml                  # cross-compile + release on tag
```

---

## 3. Plugin Component

### 3.1 PreToolUse hook — `hooks/guard.sh`

**Role:** intercept every tool call, decide allow/block based on merged config.

**stdin contract:** Claude Code writes JSON to stdin, e.g.:

```json
{"tool_name": "Bash", "tool_input": {"command": "rm -rf /tmp/foo"}}
{"tool_name": "Edit", "tool_input": {"file_path": "/etc/passwd", ...}}
{"tool_name": "WebFetch", "tool_input": {"url": "https://evil.example/x"}}
```

**stdout contract:** one JSON object:

- Allow: `{"result": "allow"}` + exit 0
- Block: `{"result": "block", "reason": "<human-readable>"}` + exit 2

**stderr:** one-line human-readable summary for the user (`ifly: blocked rm outside project: /etc/passwd`).

**Latency target:** < 50 ms mean over 100 invocations. No per-invocation subprocess calls into other languages. Config parsed in-process via `parse_defaults.sh` and `yq` if available.

**Dispatch table:**

| tool_name | Handler | Default guard state |
|---|---|---|
| `Bash` | `command_guard` | on |
| `Edit`, `Write`, `MultiEdit`, `NotebookEdit` | `path_guard` | on |
| `Read`, `Glob`, `Grep` | `sensitive_path_guard` | strict mode only |
| `WebFetch`, `WebSearch` | `network_allowlist_guard` | strict mode only |
| any other | pass-through | — |

#### 3.1.1 `command_guard(command)`

1. Strip leading `sudo` (and `doas`).
2. Check against always-blocked patterns (see §3.1.5).
3. Split on `;`, `&&`, `||`, `|` — each subcommand examined independently; `cd <path>` updates a virtual pwd for subsequent subcommands in the chain.
4. For each subcommand:
   - Match against `blocked_commands` from merged config (literal substring, case-sensitive).
   - If verb is in `{rm, rmdir, mv, chmod, chown, unlink, shred, truncate}`: extract path args (skip flags like `-rf`), resolve each to absolute via `path_resolve`, check against allowed roots (project + `additional_dirs`). Outside → block (strict/project) or warn (open).
   - If verb is in `{curl, wget, ssh, scp, sftp, nc, ncat, telnet, ftp, rsync}` AND guard.level == `strict`: block outright.

#### 3.1.2 `path_guard(file_path)`

1. Resolve `file_path` via `path_resolve` (expand `~`/`$HOME`, collapse `..`, walk symlinks with `realpath -m` — `-m` allows nonexistent leaf for `Write` to not-yet-existing files).
2. Prefix-match with trailing-slash guard against allowed roots (project root + `additional_dirs`).
3. Miss → block (strict/project) or warn (open).

#### 3.1.3 `sensitive_path_guard(path)` (strict only)

1. Resolve `path`.
2. For each pattern in `sensitive_paths` (treated as prefix patterns after `~` expansion): if path begins with the pattern, block with reason `reading sensitive path (<pattern>)`.

#### 3.1.4 `network_allowlist_guard(url)` (strict only)

1. Extract host via bash regex `^https?://([^/:]+)`.
2. If host (exact or suffix) matches any entry in `allowed_network`, allow. Otherwise block.
3. Non-HTTP URLs in strict mode: block.

#### 3.1.5 Always-blocked patterns (all levels except `off`)

| Pattern | Reason |
|---|---|
| `git push --force` / `git push -f` / `--force-with-lease` | forced history rewrite |
| `git reset --hard` | irreversible; user must run unguarded |
| `rm -rf /` (with or without flags reordered) | catastrophic |
| `dd if=` | raw device write |
| `mkfs.` (any filesystem) | format |
| `DROP TABLE` / `DROP DATABASE` | SQL data loss |
| `eval ` / `bash -c ` / `sh -c ` (as argv[0]) | shell injection surface |
| `: > ~/.bash_history` / history manipulation | audit trail tampering |

**Exception:** `level: off` skips this list too. That's the point of `off`.

#### 3.1.6 Guard levels (from merged config)

| Level | Behavior |
|---|---|
| `strict` | All guards on. Shell network commands blocked outright. Sensitive-path reads blocked. |
| `project` | Boundary guard on. Shell network commands allowed. Sensitive-path reads not checked. WebFetch/WebSearch not checked. |
| `open` | Out-of-project destructive ops produce stderr warning + exit 0. Always-blocked list still blocks. Sensitive-path reads and WebFetch not checked. |
| `off` | Pass-through. No checks at all (including always-blocked list). |

Set via `~/.config/ifly/config.yaml` (`guard.level`), `.ifly.yaml`, `IFLY_GUARD` env, or `/ifly:guard <level>` (writes to state).

### 3.2 SessionStart hook — `hooks/session_start.sh`

**Role:** on every session, read state.yaml + project .ifly.yaml to determine active mode, inject context so Claude loads the right skill.

**stdout:** context block written into Claude's session prompt:

```
# IFLy session context
Active verbosity mode: silent
Active guard level: strict
Apply skills/ifly-mode-silent/SKILL.md to every response in this session.
```

If no state file and no project config, default to `mode: minimal`, `guard.level: strict`.

### 3.3 Mode skills — `skills/ifly-mode-*/SKILL.md`

Five skills. Each is a **discipline-enforcing** skill (per `superpowers:writing-skills`). Each MUST be written TDD-style: RED-phase pressure-scenario baseline first, then GREEN skill file, then REFACTOR to close rationalization loopholes.

Each skill has the same structure:

```markdown
---
name: ifly-mode-<variant>
description: Use when user has activated <variant> mode via /ifly:mode <variant>
             or when SessionStart injects mode=<variant> from state.
             Apply to every subsequent response in this session.
---

# Mode: <variant>

## When to activate
<specific trigger conditions; not summary of rules>

## Instructions
<concrete, non-negotiable rules in imperative voice>

## Examples
<same input, BEFORE/AFTER output in this mode>

## Anti-patterns
<what breaking the mode looks like + why agents do it>

## Rationalization table
<populated from RED-phase findings>

## Red flags
<self-check list; "if you catch yourself X, stop and reapply">
```

**Per-skill behavior spec:**

| Skill | Behavior |
|---|---|
| `ifly-mode-silent` | No explanations. Output only code, commands, file paths, or direct answers. No preamble, no "here's what I did", no trailing summaries. Edits land silently. |
| `ifly-mode-minimal` | One-line summary per action. No bullet lists. No markdown headers in explanations. Code comments only where non-obvious. |
| `ifly-mode-normal` | No-op passthrough. Exists so `/ifly:mode normal` has a target; instruction body is "behave as the base Claude Code system prompt would." |
| `ifly-mode-verbose` | Explain reasoning. Show alternatives considered. Inline-document generated code. End with change summary. |
| `ifly-mode-caveman` | Maximally compressed English. No articles, no pronouns, no filler. Telegram-style. Example for "explain this function": `"sorts array in-place. O(n log n). mutates input. returns void."` Variable names in generated code stay normal. |

**Budget:** each SKILL.md under 500 words.

### 3.4 Slash commands — `commands/*.md`

#### 3.4.1 `commands/mode.md`

```
Usage: /ifly:mode <silent|minimal|normal|verbose|caveman>
```

Behavior: validate argument. If invalid, print usage and exit without state change. If valid:
1. Write `active_mode: <variant>` and `updated_at: <now>` to `~/.config/ifly/state.yaml` via embedded bash block.
2. Acknowledge: `IFLy mode → <variant>`.
3. Instruct Claude to load the corresponding skill for the rest of the session.

#### 3.4.2 `commands/guard.md`

```
Usage: /ifly:guard                      # show current level + additional dirs
       /ifly:guard <strict|project|open|off>
       /ifly:guard status               # alias for no-arg form
```

With level arg: update `session_guard_override` in state.yaml. If global has `lockdown: true` and the new level loosens from global `guard.level`, reject with stderr warning and refuse the write.

#### 3.4.3 `commands/status.md`

Prints:
- Active mode (from state.yaml, fallback to merged config)
- Active guard level (session override if set, else merged config)
- Guard tool toggles (compact grid)
- `$CLAUDE_PROJECT_DIR`
- Any `.ifly.yaml` overrides detected
- `lockdown` flag from global
- Footer: rotating message; with `$RANDOM % 10 == 0`, show `💜 ifly`. Otherwise one of a small neutral set (`IFLy v<VERSION>`, `feeling lucky`, `guards up`).

### 3.5 Subagent guard inheritance

Subagents spawned via `Task` start independent Claude sessions that invoke our `PreToolUse` hook on their own tool calls. Because the hook reads merged config fresh each invocation, subagents inherit whatever global + project config + env applies. No extra wiring. Documented in `docs/plugin.md`, not coded.

### 3.6 Plugin manifest — `.claude-plugin/plugin.json`

```json
{
  "name": "ifly",
  "version": "0.1.0",
  "description": "Safety boundaries, verbosity control, and workflow optimization for Claude Code",
  "author": {
    "name": "<to fill>",
    "url": "https://github.com/<to fill>"
  }
}
```

Version is overwritten at release time from `VERSION` file.

---

## 4. CLI Component

### 4.1 Dependencies

| Purpose | Package |
|---|---|
| CLI framework | `github.com/spf13/cobra` |
| TUI | `github.com/charmbracelet/bubbletea`, `.../bubbles`, `.../lipgloss` |
| YAML | `gopkg.in/yaml.v3` |
| Self-update | `github.com/minio/selfupdate` |
| GitHub API | `net/http` + `encoding/json` |

CGO disabled. Go 1.22+. Single static binary per target.

### 4.2 Commands

#### `ifly install`

Interactive TUI installer. Flow:

1. Detect OS/arch (`internal/detect/os.go`).
2. Locate Claude config dir (`internal/detect/claude.go`):
   - Linux: `$XDG_CONFIG_HOME/claude` → `~/.config/claude` → `~/.claude`
   - macOS: `~/Library/Application Support/claude` → `~/.claude`
   - Windows: `%APPDATA%\claude` → `%USERPROFILE%\.claude`
3. PATH-lookup for `claude` binary; warn if missing.
4. TUI prompts (defaults in brackets):
   - scope: global vs project-local `[global]`
   - overwrite existing ifly? `[no]`
   - `guard.level` `[strict]`
   - each `guard.tools.*` toggle `[defaults per §4 in global config]`
   - `mode` `[minimal]`
   - `lockdown` `[false]`
5. Walk embedded plugin FS; write files to target directory.
6. Write `~/.config/ifly/config.yaml` (or `%APPDATA%\ifly\config.yaml` on Windows).
7. Verify: `plugin.json` readable, `guard.sh` executable, `defaults.yaml` present.
8. Print next steps (`/ifly:status` in Claude Code).
9. On Windows without Git Bash or WSL detected: loud warning — hooks will not run.

#### `ifly init`

1. Create `$CWD/.ifly.yaml` from template.
2. Ask whether to add `.ifly.yaml` to `.gitignore` `[no — commit it]`.

#### `ifly config`

```
ifly config show              # merged config (global + project + env)
ifly config get <dotted.key>
ifly config set <dotted.key> <value>
```

`set` validates before writing. Writes to global by default; `--project` flag targets `.ifly.yaml`.

#### `ifly status`

Prints:
- Plugin installed? (path, VERSION, checksums match embedded?)
- Hooks present and executable?
- Current merged config summary
- Claude Code detected version
- Same footer easter egg as `/ifly:status`

#### `ifly version [--love]`

Default: `IFLy v0.1.0 — I'm Feeling Lucky`.
With undocumented `--love` (cobra `Hidden: true`): `IFLy v0.1.0 — I Fucking Love You 💜`.

#### `ifly update`

1. Read current `VERSION` (linker-injected).
2. `GET /repos/<owner>/<repo>/releases/latest`.
3. If newer tag: download matching `ifly-<os>-<arch>[.exe]` + `.sha256`.
4. Verify checksum. Atomic replace via `selfupdate`.
5. Print: `Plugin files NOT refreshed. Run 'ifly install --overwrite' to sync.`

We deliberately do not silently re-install plugin on binary update — behavior change stays opt-in.

#### `ifly sandbox` (Linux only; `//go:build linux`)

```
ifly sandbox [-- claude args...]
```

Prefers `bwrap` when present:

```
bwrap \
  --ro-bind /usr /usr --ro-bind /etc /etc \
  --ro-bind /bin /bin --ro-bind /lib /lib --ro-bind /lib64 /lib64 \
  --tmpfs /tmp \
  --bind "$PROJECT_DIR" "$PROJECT_DIR" \
  --dev /dev --proc /proc \
  --unshare-pid --unshare-ipc --unshare-uts --unshare-cgroup \
  --die-with-parent \
  claude "$@"
```

Falls back to `unshare --mount --pid --fork` with minimal bind-mount helper if `bwrap` missing (weaker guarantees; logged to stderr).

On non-Linux platforms, stub:

```
ifly sandbox requires Linux namespaces. On macOS/Windows, use
'/ifly:guard strict' for in-Claude enforcement instead.
```

### 4.3 Build & cross-compile

`Makefile` targets:
- `build` — local binary for host platform
- `test` — run `go test ./cli/...` and `tests/test_guard.sh`
- `release` — five targets, ldflags inject version, produce `.sha256` sidecars
- `install` — copy local binary to `$GOBIN`

---

## 5. Config, State, and Precedence

### 5.1 Files

| Path | Scope | Who writes |
|---|---|---|
| `~/.config/ifly/config.yaml` (or `%APPDATA%\ifly\config.yaml`) | Global | `ifly install`, `ifly config set` |
| `$CLAUDE_PROJECT_DIR/.ifly.yaml` | Project | `ifly init`, hand edit, commit to repo |
| `~/.config/ifly/state.yaml` | Session | `/ifly:mode`, `/ifly:guard` |

### 5.2 Global config schema

```yaml
version: 1

mode: minimal

guard:
  level: strict
  lockdown: false
  tools:
    bash: true
    edit: true
    write: true
    multi_edit: true
    notebook_edit: true
    read: true
    glob: false
    grep: false
    web_fetch: true
    web_search: true
  additional_dirs: []
  blocked_commands: []
  allowed_network: []
  sensitive_paths:
    - "~/.ssh/"
    - "~/.aws/"
    - "~/.config/gh/"
    - "/etc/shadow"
    - "/etc/sudoers"

telemetry:
  easter_egg: true
```

### 5.3 Project config schema

Same shape, every field optional. Only present fields override globals.

### 5.4 State schema

```yaml
version: 1
active_mode: minimal
session_guard_override: null
updated_at: 2026-04-17T10:32:11Z
```

### 5.5 Precedence

Order of increasing priority:

1. Built-in defaults (shipped in `defaults.yaml`)
2. Global config
3. Project config
4. Session state (affects `active_mode` and `session_guard_override` only)
5. Env vars `IFLY_MODE`, `IFLY_GUARD`

### 5.6 Merge rules

| Field type | Rule |
|---|---|
| Scalars (`mode`, `guard.level`, booleans) | Replace; highest scope wins |
| `guard.tools.*` (map of bool) | Per-key override |
| `additional_dirs`, `blocked_commands`, `allowed_network`, `sensitive_paths` | **Additive union**; project can add, cannot remove |
| `guard.level` with global `lockdown: true` | Project override ignored if it loosens; stderr warn |

"Loosens" order: `strict > project > open > off`.

### 5.7 Single source of truth for defaults

`defaults.yaml` at repo root. Bash hook reads at runtime via `yq` if present, else `parse_defaults.sh` (pure-bash YAML-subset reader shipped with plugin). Go CLI embeds it at build time and marshals to struct. Drift prevented by CI check: Go unmarshal matches bash parser output on the same file.

### 5.8 Validation behavior

- `ifly install` and `ifly init` validate at write time.
- `ifly config set` validates before persisting.
- Bash hook validates defensively at read time. Invalid config **fails closed**: `level: strict`, all tools guarded, empty allowlists, loud stderr warning.

---

## 6. Distribution

### 6.1 Release workflow (`.github/workflows/release.yml`)

Triggered on tag push matching `v*.*.*`:

1. Assert `VERSION` file matches tag.
2. `bash tests/test_guard.sh` — must pass.
3. `go test ./cli/...` — must pass.
4. `shellcheck hooks/*.sh tests/*.sh`.
5. `golangci-lint run ./cli/...`.
6. Matrix build for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64. `CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$VERSION"`.
7. Generate `.sha256` per binary.
8. Package plugin standalone: `tar czf ifly-plugin-vX.Y.Z.tar.gz --exclude=cli --exclude=.github .`.
9. Create GitHub Release with all assets.
10. Update marketplace manifest (if wired).

### 6.2 User install paths

| Path | Command | Audience |
|---|---|---|
| Marketplace | Claude Code marketplace UI | Non-technical plugin-only users |
| Manual plugin | `git clone … && claude --plugin-dir ./ifly` | Plugin-only users with CLI comfort |
| CLI installer | Download `ifly` binary → `ifly install` | Users who want sandbox or config management |

---

## 7. Testing Strategy

### 7.1 Plugin — `tests/test_guard.sh`

Pure bash. Each case is `assert_allow <stdin_json>` or `assert_block <stdin_json> "<expected_reason_substring>"`. Exit non-zero on any failure.

**Required ≥50 test cases, organized:**

| Group | Cases |
|---|---|
| Path resolution (`~`, `$HOME`, `..`, symlinks, missing leaf) | ≥10 |
| Boundary enforcement (inside/outside project, `additional_dirs`) | ≥10 |
| Chained-command splitting + `cd` tracking | ≥8 |
| Destructive-op pattern detection (always-blocked list) | ≥6 |
| Network blocking in strict (curl, wget, ssh, nc, scp, sftp) | ≥4 |
| Sensitive-path reads (strict only) | ≥4 |
| Config merge behaviors (global+project, env override, lockdown) | ≥6 |
| Edge cases (path traversal, unicode, long paths, symlinks to symlinks, sudo/doas prefix) | ≥6 |

### 7.2 Latency benchmark — `tests/bench_guard.sh`

100 invocations, measure mean via `time`. Fail if mean > 50 ms. CI-gated.

### 7.3 Skills — `tests/skills/mode-<name>.scenarios.md`

Per writing-skills discipline. Each file documents:
- RED-phase baseline behavior (pressure scenarios that make the unguarded agent violate the mode, verbatim rationalizations captured)
- Compliant-behavior expectations with skill active
- Regression scenarios for every rationalization found

Not automated in CI. Runnable manually via `Agent` tool. Must be complete before the corresponding skill is committed.

### 7.4 CLI — `go test ./cli/...`

- Unit tests per `internal/*` package.
- Config merge table-driven test (≥20 permutations).
- `internal/installer` integration test uses `t.TempDir()` as fake Claude root; asserts file layout end-to-end.
- Sandbox tests `//go:build linux`, skip if `bwrap` absent.
- **No end-to-end test that launches Claude Code.** Plugin's own tests cover hook behavior.

### 7.5 CI matrix

- Plugin tests: ubuntu-latest, macos-latest. We mandate bash 4+ and document the requirement.
- CLI tests: ubuntu, macos, windows.
- Lint: shellcheck, golangci-lint.
- Defaults-drift check: Go unmarshal of `defaults.yaml` must deep-equal bash parser output.

---

## 8. Documentation

- `README.md` — three install paths, 60-second quick start, TUI screenshot.
- `docs/plugin.md` — hooks, skills, slash commands, config reference.
- `docs/cli.md` — CLI command reference.
- `docs/security.md` — honest threat model: hooks are advisory (not OS-level); sandbox is real isolation but Linux-only. Bash-command hostname parsing is explicitly NOT attempted.
- `CHANGELOG.md` — keep-a-changelog.

---

## 9. Non-goals (locked for v0.1)

- Token/cost tracking.
- Superpowers-style workflow orchestration (brainstorming, planning, TDD).
- Custom MCP server integration (v0.2 candidate).
- GUI or web dashboard.
- Session persistence beyond `state.yaml`'s mode.
- Shell-command hostname parsing for `allowed_network` — allowlist applies to WebFetch/WebSearch only.
- Team policy service / remote config sync — `lockdown: true` hints at it but v0.1 is local-only.
- No-op `ifly uninstall` CLI command in v0.1 — covered only via `ifly install --overwrite`.

---

## 10. Risks and Mitigations

| Risk | Mitigation |
|---|---|
| Hook latency > 50 ms breaks UX | CI benchmark; pure bash, no per-command subprocess spawns; cache merged config per invocation |
| Bash hook ↔ Go CLI defaults drift | Single `defaults.yaml`, CI drift check |
| Adversarial bash bypasses parser | Explicit non-goal in `docs/security.md`. Guard is defense-in-depth; real isolation requires `ifly sandbox` (Linux) |
| Users expect sandbox on macOS/Windows | Stub prints clear message pointing to strict mode |
| Windows without Git Bash/WSL | `ifly install` detects + warns; docs make it explicit prerequisite |
| Superpowers overlap confusion | README section "IFLy vs Superpowers" — IFLy is safety + verbosity; superpowers is workflow. Install both if you want both. |

---

## 11. Version lifecycle

- `v0.x.y` pre-1.0. Breaking changes allowed on minor bumps with CHANGELOG entry.
- `v1.0.0` when plugin surface stable, CLI command set frozen, ≥3 external clean-install reports, no open P0 security issues.

---

## 12. Implementation plan outline

Follow-up via `superpowers:writing-plans` produces **two plans**:

**Plan A — Plugin.** Order of work:
1. `defaults.yaml` + `parse_defaults.sh` (shared substrate).
2. `hooks/lib/path_resolve.sh`, `split_command.sh` (building blocks).
3. `hooks/guard.sh` with its test harness — 50+ cases, TDD. Latency bench.
4. `hooks/session_start.sh`.
5. Five mode skills via writing-skills TDD (one at a time).
6. Three slash commands.
7. Plugin manifest + `VERSION` + `LICENSE` + `README.md`.

Plan A exits with a shippable plugin installable via `--plugin-dir`.

**Plan B — CLI.** Built on top of the completed plugin tree. Order:
1. Module bootstrap + `plugin_fs.go` embedding.
2. `internal/config`, `internal/state`, `internal/paths`, `internal/detect`.
3. `cmd/version`, `cmd/status`, `cmd/config` (read-only commands first).
4. `cmd/init`.
5. `cmd/install` (bubbletea TUI).
6. `cmd/update`.
7. `cmd/sandbox` (Linux only).
8. Release workflow + `.github/workflows/{ci,release}.yml`.
9. Documentation pass.

Both plans use `superpowers:subagent-driven-development` for execution (subagents are available). Both happen in a git worktree per `superpowers:using-git-worktrees`. Repo must be initialized as git first — currently not a repo.
