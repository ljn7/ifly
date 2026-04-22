# IFLy test coverage

This document summarizes what the current test suite covers and what still
needs manual or environment-specific verification.

## Verified locally

The Go CLI test suite passes:

```bash
cd cli
go test ./...
```

The Bash hook/library suite passes under Git Bash:

```bash
bash tests/test_guard.sh
bash tests/test_path_resolve.sh
bash tests/test_session_start.sh
bash tests/test_parse_defaults.sh
bash tests/test_split_command.sh
bash tests/test_config.sh
```

Latest manual run:

- `tests/test_guard.sh`: 83 passed, 0 failed
- `tests/test_path_resolve.sh`: 20 passed, 0 failed
- `tests/test_session_start.sh`: 6 passed, 0 failed
- `tests/test_parse_defaults.sh`: 8 passed, 0 failed
- `tests/test_split_command.sh`: 8 passed, 0 failed
- `tests/test_config.sh`: 12 passed, 0 failed

## Go CLI coverage

The Go tests cover:

- CLI package buildability.
- `install` writing plugin files and global config.
- install verification for required plugin files.
- installer overwrite refusal and overwrite success.
- executable hook file mode handling.
- config loading, merging, env overrides, and lockdown.
- state read/write/merge behavior.
- Claude config path detection for supported OS families.
- `mode` global state and `--project` state.
- `guard` global state and `--project` state.
- `status`, `version`, `sandbox`, and updater helper behavior.

## Bash guard coverage

The guard tests cover:

- always-blocked destructive patterns:
  - `git push --force`
  - `git reset --hard`
  - `rm -rf /`
  - `dd if=`
  - `mkfs`
  - `DROP TABLE`
  - `eval`
  - `bash -c`
- destructive path checks:
  - `rm`
  - `rmdir`
  - `mv`
  - `chmod`
  - `chown`
  - `shred`
  - `truncate`
- inside-project allow behavior.
- outside-project block behavior.
- `additional_dirs` allow behavior.
- path traversal blocking.
- unset `CLAUDE_PROJECT_DIR` fallback to current directory.
- Windows/MSYS/drive path normalization.
- PowerShell destructive command blocking:
  - `Remove-Item`
  - `Move-Item`
  - `del`
  - `rd`
  - `rm`
  - script blocks
- nested shell wrappers:
  - Bash to PowerShell
  - Bash to `pwsh`
  - Bash to `cmd`
- guard levels:
  - `strict`
  - `project`
  - `open`
  - `off`
- shell network commands in strict mode:
  - `curl`
  - `wget`
  - `ssh`
  - `nc`
- `WebFetch` allowlist behavior.
- suffix domain allowlist behavior.
- non-HTTP URL blocking.
- sensitive read blocking:
  - `~/.ssh`
  - `~/.aws`
  - `/etc/shadow`
- per-tool toggles.
- `blocked_commands` policy.
- lockdown behavior.
- session guard override from state.
- indented YAML state parsing.

## Bash library coverage

Path resolver tests cover:

- `~`
- `$HOME`
- `${HOME}`
- `..` cleanup
- absolute paths
- relative paths
- Windows drive paths
- MSYS paths
- inside/outside boundary checks

Command splitter tests cover:

- single commands
- `;`
- `&&`
- `||`
- `|`
- virtual `cd` tracking
- relative `cd`
- `cd ..`

Config parser tests cover:

- flat scalars
- nested maps
- string lists
- quoted strings
- comments
- blank lines
- nested lists
- merge precedence
- env overrides
- lockdown
- empty lists

## Manually verified

- Windows batch build via `build-cli.bat`.
- CLI command sweep in a temporary config/project environment.
- installed plugin command files are synced into the Claude plugin directory.
- slash command files use Claude's `!` Bash expansion syntax.

## Not fully automated yet

These still need live or environment-specific checks:

- Real Claude Code slash command execution end-to-end.
- GitHub Actions release creation on an actual tag push.
- Linux `ifly sandbox` behavior with namespaces.
- macOS live install path behavior.
- Project `.ifly-state.yaml` priority inside Bash hook tests. CLI behavior is
  tested, but a dedicated hook test should still be added.

## Recommended pre-release checklist

```bash
cd cli
go test ./...
cd ..

bash tests/test_guard.sh
bash tests/test_path_resolve.sh
bash tests/test_session_start.sh
bash tests/test_parse_defaults.sh
bash tests/test_split_command.sh
bash tests/test_config.sh

./build-cli.sh
./dist/ifly version
./dist/ifly status
```

On Windows:

```powershell
.\build-cli.bat
.\dist\ifly.exe version
.\dist\ifly.exe status
```
