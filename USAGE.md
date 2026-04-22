# IFLy usage guide

This guide shows the normal day-to-day workflow after IFLy is installed.

## First check

Confirm the CLI works:

```bash
ifly version
ifly paths
ifly status
```

On Windows, if you are running from the build directory:

```powershell
.\dist\ifly.exe version
.\dist\ifly.exe paths
.\dist\ifly.exe status
```

In Claude Code:

```text
/reload-plugins
/ifly:status
```

## Choose a default mode

Modes control how verbose Claude should be.

```bash
ifly mode silent
ifly mode minimal
ifly mode normal
ifly mode verbose
ifly mode caveman
```

Recommended default:

```bash
ifly mode minimal
```

Inside Claude Code:

```text
/ifly:mode minimal
```

The CLI path is preferred when you want to avoid spending slash-command tokens.

## Choose a default guard

Guard levels control how strict the safety hook is.

```bash
ifly guard strict
ifly guard project
ifly guard open
ifly guard off
```

Recommended default:

```bash
ifly guard strict
```

Meaning:

- `strict`: strongest hook policy; blocks shell network commands and sensitive reads.
- `project`: blocks destructive file operations outside the project.
- `open`: warns instead of blocking external destructive paths.
- `off`: disables boundary checks, but `blocked_commands` still apply.

Check current guard:

```bash
ifly guard
ifly guard status
```

Inside Claude Code:

```text
/ifly:guard strict
/ifly:guard status
```

## Use project-local state

Use project-local state when one repo needs a different mode or guard from your
global default.

```bash
cd /path/to/project
ifly mode --project verbose
ifly guard --project project
```

This writes:

```text
.ifly-state.yaml
```

That file is ignored by git by default and is intended to be personal.

## Use project policy

Use project config when you want team/project policy that can be committed.

```bash
cd /path/to/project
ifly init
```

This creates:

```text
.ifly.yaml
```

Example:

```yaml
version: 1
mode: normal
guard:
  level: project
  additional_dirs:
    - ./dist
    - ./build
  blocked_commands:
    - "terraform destroy"
    - "npm publish"
  allowed_network:
    - api.github.com
```

## Add blocked commands

Blocked commands are literal substring matches and apply at every guard level.

View active policy:

```bash
ifly block list
```

Add custom commands globally:

```bash
ifly block add "docker rm"
ifly block add "kubectl delete"
ifly block add "terraform destroy"
```

Add custom commands for one project:

```bash
ifly block add --project "npm publish"
ifly block add --project "docker compose down -v"
```

Use presets:

```bash
ifly block presets
ifly block preset git-danger
ifly block preset filesystem-danger
```

Project preset:

```bash
ifly block preset --project git-danger
```

## Common workflows

Strict global setup:

```bash
ifly mode minimal
ifly guard strict
ifly block preset git-danger
ifly block preset filesystem-danger
```

Verbose mode for one project:

```bash
cd /path/to/project
ifly mode --project verbose
```

Strict guard for one important project:

```bash
cd /path/to/project
ifly guard --project strict
```

Allow build output outside the repo:

```yaml
guard:
  additional_dirs:
    - /tmp/build-output
```

Then check:

```bash
ifly status
```

## Updating plugin files

After rebuilding or updating the CLI:

```bash
ifly install --overwrite
```

In Claude Code:

```text
/reload-plugins
/ifly:status
```

## Troubleshooting

Show paths:

```bash
ifly paths
```

Show merged config:

```bash
ifly config show
```

Show active guard:

```bash
ifly guard status
```

Show blocked commands:

```bash
ifly block list
```

If Claude shows old mode/guard values, check state files:

Windows:

```powershell
Get-Content "$env:APPDATA\ifly\state.yaml"
Get-Content .\.ifly-state.yaml
```

Linux/macOS:

```bash
cat ~/.config/ifly/state.yaml
cat .ifly-state.yaml
```

If slash commands do not appear:

```text
/reload-plugins
```
