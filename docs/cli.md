# IFLy CLI Reference

The `ifly` binary installs and manages the IFLy Claude Code plugin.

## Install

Download the latest release for your platform from
<https://github.com/ljn7/ifly/releases>, verify the checksum, and run
`ifly install`.

```bash
sha256sum -c ifly-linux-amd64.sha256
chmod +x ifly-linux-amd64 && mv ifly-linux-amd64 ~/.local/bin/ifly
ifly install
```

To compile from a checkout:

```bash
./build-cli.sh      # Linux/macOS/Git Bash
build-cli.bat      # Windows cmd.exe
```

Both scripts sync the plugin embed tree into `cli/plugin/` and write the CLI
binary to `dist/`.

## Commands

### `ifly install`

Interactive TUI installer. Detects OS/arch, locates the Claude config
directory, walks the embedded plugin tree, writes a global `config.yaml`.
If the target plugin directory already exists and contains files, interactive
and `--no-tui` installs ask before overwriting. Use `--overwrite` for CI or
scripted refreshes.

Flags:
- `--overwrite` — replace existing plugin files
- `--no-tui` — non-interactive; uses `--mode`, `--guard`, `--lockdown`
- `--mode <silent|minimal|normal|verbose|caveman>` — default verbosity (default `minimal`)
- `--guard <strict|project|open|off>` — default guard level (default `strict`)
- `--lockdown` — prevent projects/env from loosening guard level

### `ifly init`

Writes a starter `.ifly.yaml` into the current project (or
`$CLAUDE_PROJECT_DIR` if set).

Flags:
- `--force` — overwrite an existing `.ifly.yaml`

### `ifly config`

```
ifly config show                        # print merged config (global + project + env)
ifly config get <dotted.key>
ifly config set <dotted.key> <value>    # writes to global
ifly config set --project <key> <v>     # writes to .ifly.yaml
```

Supported keys: `mode`, `guard.level`, `guard.lockdown`, `telemetry.easter_egg`.

### `ifly paths`

Shows the detected host, IFLy config/state files, project config path, Claude
config candidates, and the `plugins/ifly` install target. The CLI uses these
Claude config roots in priority order:

- Windows: `%APPDATA%\claude`, then `%USERPROFILE%\.claude`
- macOS: `~/Library/Application Support/claude`, then `~/.claude`
- Linux/BSD: `$XDG_CONFIG_HOME/claude`, then `~/.config/claude`, then `~/.claude`

Official Claude Code docs define plugin structure, marketplace installation,
plugin settings in `settings.json`, and slash command locations. The CLI keeps
the installed plugin files under the detected Claude config root's
`plugins/ifly` directory.

### `ifly block`

Manages `guard.blocked_commands`, the literal-substring policy list enforced
by the guard hook at every guard level, including `off`.

```
ifly block list                 # show merged blocked commands
ifly block presets              # show built-in preset names/categories
ifly block categories           # show preset categories
ifly block add "docker rm"      # add one custom global pattern
ifly block preset git-danger    # add a preset to global config
ifly block preset --project git-danger
```

Built-in presets are grouped by category:

- `git-danger` — destructive git history/worktree commands.
- `archive-overwrite` — tar/unzip patterns that overwrite or use absolute roots.
- `shell-wrapper` — nested shell execution wrappers such as `bash -c`, `cmd /c`.
- `filesystem-danger` — broad recursive deletion/permission patterns.

### `ifly mode <silent|minimal|normal|verbose|caveman>`

Writes the active session mode to `state.yaml` without invoking a Claude
slash command. This is the low-token path for changing verbosity outside the
conversation. It preserves any existing session guard override.

Flag:
- `--project` — write `.ifly-state.yaml` in the current project instead of
  the global state file. Hooks prefer project state when present.

### `ifly guard [strict|project|open|off|status]`

Shows or writes the active session guard override in `state.yaml` without
invoking a Claude slash command.

Use `ifly guard status` or `ifly guard` to print the active guard, lockdown
state, and additional directories. Use `ifly guard <level>` to change the
session override. If global lockdown is enabled, the CLI refuses overrides
that would loosen the global guard level.

Flag:
- `--project` — write `.ifly-state.yaml` in the current project instead of
  the global state file.

### `ifly status`

Prints active mode, active guard level, tool toggles, Claude config dir,
project config path, any lockdown violations, and a rotating footer.

### `ifly version [--love]`

Prints `IFLy v<version> — I'm Feeling Lucky`. The hidden `--love` flag
changes the tagline. (If you discover it: keep it to yourself.)

### `ifly update`

Self-updates the `ifly` binary from the latest GitHub release. The plugin
files on disk are *not* touched — run `ifly install --overwrite` to
refresh them after updating the binary.

Flags:
- `--dry-run` — download and verify checksum but do not replace the binary

### `ifly sandbox [-- claude args...]`

**Linux only.** Launches `claude` inside a filesystem namespace:
system dirs are read-only, `/tmp` is a tmpfs, the current project is
read-write. Prefers `bwrap`; falls back to `unshare --mount --pid --fork`
with weaker guarantees.

On macOS and Windows, prints:

```
sandbox requires Linux namespaces; on macOS/Windows use `/ifly:guard strict`
```

## Config precedence

Highest priority first:
1. `IFLY_MODE`, `IFLY_GUARD` env vars
2. Project session state (`<project>/.ifly-state.yaml`, if present — set via `ifly mode --project` or `ifly guard --project`)
3. Global session state (`~/.config/ifly/state.yaml` or OS equivalent — set via `ifly mode`, `ifly guard`, `/ifly:mode`, or `/ifly:guard`)
4. Project config (`$CLAUDE_PROJECT_DIR/.ifly.yaml`)
5. Global config (`~/.config/ifly/config.yaml`)
6. Shipped defaults (`defaults.yaml` embedded at build time)

Scalars replace. `guard.tools.*` merge per key. `additional_dirs`,
`blocked_commands`, `allowed_network`, `sensitive_paths` are additive
unions — a project can add entries but not remove them.

If global config has `guard.lockdown: true`, any lower scope's attempt
to loosen `guard.level` (strict → project → open → off) is dropped and
reported in `ifly status`.

## OS-specific notes

**Windows.** The plugin's hooks are bash scripts. You need Git for Windows
(Git Bash) or WSL installed so Claude Code can spawn `bash` to run them.
`ifly install` prints a warning if no bash is detected. `jq` must also be
on PATH — winget-installed jq lives in
`%LocalAppData%\Microsoft\WinGet\Packages\jqlang.jq_.../` and is not
auto-added.

**macOS.** Hooks work out of the box with system bash + Homebrew jq. No
OS-level sandbox; `level: strict` is the available mitigation.

**Linux.** Full feature surface including `ifly sandbox`. Install `bwrap`
(`apt install bubblewrap`) for the strongest isolation; `unshare` is the
fallback.
