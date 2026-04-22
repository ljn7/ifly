# IFLy plugin reference

This document describes every component of the IFLy Claude Code plugin —
hooks, skills, slash commands, and the configuration schema.

## Components at a glance

| Path | Purpose |
|---|---|
| `.claude-plugin/plugin.json` | Plugin manifest |
| `defaults.yaml` | Built-in defaults read by both the bash hook and the future Go CLI |
| `hooks/guard.sh` | PreToolUse hook — boundary enforcement |
| `hooks/ifly-state` | Slash-command helper for mode/guard/status state reads and writes |
| `hooks/session_start.sh` | SessionStart hook — injects mode context |
| `hooks/lib/*.sh` | Reusable bash modules: YAML parser, path resolver, command splitter, config merger |
| `skills/ifly-mode-<name>/SKILL.md` | Verbosity-mode skill (5 of them) |
| `commands/<name>.md` | Slash command (mode, guard, status) |
| `tests/test_*.sh` | Unit tests for the bash modules and the guard hook |
| `tests/bench_guard.sh` | Latency benchmark (50 ms budget on Linux/macOS CI) |

## The guard hook

`hooks/guard.sh` is invoked by Claude Code as a PreToolUse hook. It reads
the tool-call JSON from stdin, decides allow/block, and writes a decision
JSON to stdout. Exit code semantics:

- `0` + `{"result":"allow"}` — proceed.
- `2` + `{"result":"block","reason":"..."}` — block; reason also written
  to stderr as `ifly: <reason>`.

### Stdin contract

```json
{"tool_name": "Bash", "tool_input": {"command": "rm -rf /"}}
```

For path-based tools (Edit, Write, MultiEdit, NotebookEdit), the relevant
key is `tool_input.file_path` (NotebookEdit also accepts `notebook_path`).
For Read/Glob/Grep, fall-through chain `file_path // path // pattern`.
For WebFetch/WebSearch, `tool_input.url` (WebSearch also accepts `query`).

### Dispatch

| Tool | Handler | Active levels |
|---|---|---|
| `Bash` | `_handle_bash` | strict, project, open, off |
| `Edit`, `Write`, `MultiEdit`, `NotebookEdit` | `_handle_path` | strict, project, open |
| `Read`, `Glob`, `Grep` | `_handle_sensitive_path` | strict only |
| `WebFetch`, `WebSearch` | `_handle_network_allowlist` | strict only |
| anything else | pass-through | n/a |

### `_handle_bash` checks (in order, per subcommand)

The command is first stripped of a single leading `sudo`/`doas` prefix,
then split on `;`, `&&`, `||`, `|` with virtual `cd` tracking. For each
subcommand:

1. **Always-blocked patterns** (skipped if `level: off`):
   `git push --force/-f/--force-with-lease`, `git reset --hard`,
   `rm -rf /`, `dd if=`, `mkfs.*`, `DROP TABLE/DATABASE`,
   `eval`, `bash -c`, `sh -c`.
2. **`blocked_commands` from config** (literal substring, case-sensitive,
   runs at every level including `off`).
3. **Shell network verbs** (strict only): `curl, wget, ssh, scp, sftp, nc,
   ncat, telnet, ftp, rsync` → `_emit_block "shell network in strict: <verb>"`.
4. **Destructive-verb path check** (skipped if `off`): for verbs
   `rm, rmdir, mv, chmod, chown, unlink, shred, truncate`, every
   non-flag arg is resolved (with `truncate -s <size>` correctly skipped)
   and checked against `CLAUDE_PROJECT_DIR + additional_dirs`. Outside →
   block (or warn-and-allow on `level: open`).

### `_handle_path` checks (Edit/Write/MultiEdit/NotebookEdit)

1. Per-tool toggle (`guard.tools.<name>`); default `true`. If `false`, allow.
2. If `level: off`, allow.
3. Resolve `file_path` against `CLAUDE_PROJECT_DIR`. Check membership in
   `project + additional_dirs`. Outside → block on strict/project,
   warn-and-allow on open.

### `_handle_sensitive_path` (Read/Glob/Grep, strict only)

1. If level is not strict, allow.
2. Per-tool toggle; default `false` (opt-in). `defaults.yaml` ships
   `read: true, glob: false, grep: false`.
3. For each prefix in `sensitive_paths`, if the resolved target path
   begins with the resolved prefix → block.

### `_handle_network_allowlist` (WebFetch/WebSearch, strict only)

1. If not strict, allow.
2. Per-tool toggle; default `true`.
3. Extract URL via jq. If not `^https?://` → block "non-HTTP".
4. Extract host via `sed -E 's#^https?://([^/:]+).*#\1#'`.
5. For each entry in `allowed_network`, allow if `host == entry` OR
   `host == *.<entry>` (suffix match). Else block "URL not in allowed_network".

## SessionStart hook

`hooks/session_start.sh` runs once when Claude Code starts a session. It
reads merged config + optional state and prints:

```
# IFLy session context
mode: <mode>
guard: <level>
skill: skills/ifly-mode-<mode>/SKILL.md
Apply the indicated mode skill to every subsequent response in this session.
```

If state has `active_mode: <name>`, that overrides the merged `mode` scalar.
Hooks prefer `<project>/.ifly-state.yaml` when present, otherwise the global
state file (`%APPDATA%/ifly/state.yaml` on Windows, `~/.config/ifly/state.yaml`
on Linux/BSD, or `~/Library/Application Support/ifly/state.yaml` on macOS).

## Verbosity mode skills

| Skill | Behavior |
|---|---|
| `ifly-mode-silent` | Output only the direct answer. No preamble, no summary, no bullets. |
| `ifly-mode-minimal` | One line per action. No headers, no bullets in prose. (Default.) |
| `ifly-mode-normal` | No-op passthrough. Base Claude Code behavior. |
| `ifly-mode-verbose` | ≥2 alternatives with tradeoffs, recommendation, change summary. Docstrings on all generated functions. |
| `ifly-mode-caveman` | Maximally compressed prose (no articles, no pronouns, no filler). Generated CODE stays idiomatic. |

Each skill ships with a `tests/skills/mode-<name>.scenarios.md` documenting
RED/GREEN/REFACTOR pressure scenarios. Skills enforce themselves through
explicit rationalization tables — every excuse the model might generate
has a written counter.

## Slash commands

### `/ifly:mode <silent|minimal|normal|verbose|caveman>`

Runs `hooks/ifly-state mode <variant>` through the Bash tool. The helper
persists the choice to the active state file and preserves any existing guard
override. If `<project>/.ifly-state.yaml` exists, slash commands update it;
otherwise they update global state. The command then applies the matching mode
skill to the rest of the session. Invalid argument prints usage and changes
nothing.

### `/ifly:guard [strict|project|open|off|status]`

Runs `hooks/ifly-state guard <arg>` through the Bash tool. No arg or `status`
prints the active guard level, additional_dirs, and lockdown state. With a
level argument: validates, refuses if global `lockdown: true` would be
loosened, otherwise persists to state.

### `/ifly:status`

Runs `hooks/ifly-state status` through the Bash tool and prints a compact
status block.

All three slash commands set `disable-model-invocation: true`, so Claude
cannot invoke them through the `SlashCommand` tool and their metadata is not
kept in normal model context. Manual `/ifly:*` invocation still consumes
tokens because Claude Code loads the command file and includes command output
in the turn.

## Configuration schema

### `defaults.yaml` (shipped)

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

### Merge precedence

For scalars (e.g., `mode`, `guard.level`, `guard.lockdown`):
**env > project > global > defaults** — last wins.

For lists (e.g., `additional_dirs`, `blocked_commands`, `allowed_network`,
`sensitive_paths`): **unioned additively in source order** — global entries
appear first, then project entries, reindexed from 0.

### Lockdown

If global `guard.lockdown: true` and the project requests a `guard.level`
looser than the global's (rank order: `strict=3 > project=2 > open=1 > off=0`),
the project request is ignored and a warning goes to stderr.

### Per-project `.ifly.yaml`

Drop one in your repo root:

```yaml
guard:
  level: project
  additional_dirs:
    - /tmp/build-artifacts
    - ~/shared-libs
  blocked_commands:
    - "docker rm"
    - "kubectl delete"
  allowed_network:
    - "registry.npmjs.org"
    - "api.github.com"
```

Use the companion CLI to manage blocked commands without hand-editing YAML:

```bash
ifly block presets
ifly block add "docker rm"
ifly block preset git-danger
ifly block preset --project archive-overwrite
```

`blocked_commands` are literal substring matches. They are additive across
global and project config and run at every guard level, including `off`.

### Env overrides

- `IFLY_MODE=<name>` overrides `mode`.
- `IFLY_GUARD=<level>` overrides `guard.level`.
- `IFLY_DEFAULTS=<path>`, `IFLY_GLOBAL=<path>`, `IFLY_PROJECT=<path>`,
  `IFLY_STATE=<path>` — explicit config-source paths (mainly for tests).
  Each respects an explicit empty value (`IFLY_GLOBAL=""`) to opt out of a
  source.

## Testing

| File | What it covers |
|---|---|
| `tests/test_harness_sanity.sh` | The assert harness itself |
| `tests/test_parse_defaults.sh` | YAML-subset parser |
| `tests/test_path_resolve.sh` | Tilde/HOME/.. expansion + boundary check |
| `tests/test_split_command.sh` | Chain splitter + cd tracking |
| `tests/test_config.sh` | Merged-config loader + lockdown |
| `tests/test_guard.sh` | End-to-end guard hook (60+ cases) |
| `tests/test_session_start.sh` | SessionStart context injection |
| `tests/bench_guard.sh` | Latency benchmark (50 ms budget; CI-gated) |
| `tests/skills/mode-*.scenarios.md` | Mode-skill RED/GREEN/REFACTOR docs |

Run any test directly: `bash tests/test_guard.sh`. The bash hook needs `jq`
on PATH; the spec doc and security doc both call this out.

## Platform notes

- **Linux/macOS:** primary target. Bash 4+ required.
- **Git Bash on Windows:** all tests pass; `jq` must be on PATH (winget
  package installs to `~/AppData/Local/Microsoft/WinGet/Packages/jqlang.jq.../`
  by default — add that folder). Process spawn overhead is ~10× Linux, so
  the latency benchmark will exceed the 50 ms budget locally; CI on Linux is
  the authoritative gate.
- **WSL:** treated as Linux.
