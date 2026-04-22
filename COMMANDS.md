# IFLy command reference

This file lists the common CLI and Claude slash-command combinations.

## Core CLI

```powershell
.\dist\ifly.exe version
.\dist\ifly.exe status
.\dist\ifly.exe paths
.\dist\ifly.exe install
.\dist\ifly.exe install --overwrite
.\dist\ifly.exe install --no-tui --mode normal --guard project --lockdown
```

## Mode

Global state:

```powershell
.\dist\ifly.exe mode silent
.\dist\ifly.exe mode minimal
.\dist\ifly.exe mode normal
.\dist\ifly.exe mode verbose
.\dist\ifly.exe mode caveman
```

Project-local state:

```powershell
.\dist\ifly.exe mode --project silent
.\dist\ifly.exe mode --project minimal
.\dist\ifly.exe mode --project normal
.\dist\ifly.exe mode --project verbose
.\dist\ifly.exe mode --project caveman
```

Project-local state writes `.ifly-state.yaml` in the current project. It is
ignored by git by default.

## Guard

Show active guard:

```powershell
.\dist\ifly.exe guard
.\dist\ifly.exe guard status
```

Set global guard state:

```powershell
.\dist\ifly.exe guard strict
.\dist\ifly.exe guard project
.\dist\ifly.exe guard open
.\dist\ifly.exe guard off
```

Set project-local guard state:

```powershell
.\dist\ifly.exe guard --project strict
.\dist\ifly.exe guard --project project
.\dist\ifly.exe guard --project open
.\dist\ifly.exe guard --project off
```

Guard levels:

- `strict` blocks destructive external paths, shell network commands, and enabled sensitive reads.
- `project` blocks destructive paths outside the project and additional dirs.
- `open` warns for external destructive paths.
- `off` disables boundary checks, but `blocked_commands` policy still applies.

## Config

Show merged config:

```powershell
.\dist\ifly.exe config show
```

Read values:

```powershell
.\dist\ifly.exe config get mode
.\dist\ifly.exe config get guard.level
.\dist\ifly.exe config get guard.lockdown
.\dist\ifly.exe config get telemetry.easter_egg
```

Set global config:

```powershell
.\dist\ifly.exe config set mode normal
.\dist\ifly.exe config set guard.level project
.\dist\ifly.exe config set guard.lockdown true
.\dist\ifly.exe config set telemetry.easter_egg false
```

Set project config:

```powershell
.\dist\ifly.exe config set --project mode verbose
.\dist\ifly.exe config set --project guard.level strict
.\dist\ifly.exe config set --project guard.lockdown false
```

Project config writes `.ifly.yaml`, which is intended for team/project policy.

## Project setup

```powershell
.\dist\ifly.exe init
.\dist\ifly.exe init --force
```

Creates or overwrites `.ifly.yaml`.

## Blocked commands

`blocked_commands` are literal substring matches. They apply at every guard
level, including `off`.

View active blocked commands:

```powershell
.\dist\ifly.exe block list
.\dist\ifly.exe block get
```

View built-in presets:

```powershell
.\dist\ifly.exe block presets
.\dist\ifly.exe block categories
```

Add custom global blocked commands:

```powershell
.\dist\ifly.exe block add "docker rm"
.\dist\ifly.exe block add "kubectl delete"
.\dist\ifly.exe block add "git checkout -- ."
.\dist\ifly.exe block add "npm publish"
.\dist\ifly.exe block add "terraform destroy"
```

Add custom project blocked commands:

```powershell
.\dist\ifly.exe block add --project "docker compose down -v"
.\dist\ifly.exe block add --project "pnpm publish"
.\dist\ifly.exe block add --project "git clean -fdx"
```

Add global presets:

```powershell
.\dist\ifly.exe block preset git-danger
.\dist\ifly.exe block preset archive-overwrite
.\dist\ifly.exe block preset shell-wrapper
.\dist\ifly.exe block preset filesystem-danger
```

Add project presets:

```powershell
.\dist\ifly.exe block preset --project git-danger
.\dist\ifly.exe block preset --project archive-overwrite
.\dist\ifly.exe block preset --project shell-wrapper
.\dist\ifly.exe block preset --project filesystem-danger
```

Preset examples:

- `git-danger`: `git reset --hard`, `git clean -fdx`, `git push --force`, `git checkout -- .`
- `archive-overwrite`: `tar --overwrite`, `tar -C /`, `tar --absolute-names`, `unzip -o`
- `shell-wrapper`: `bash -c`, `sh -c`, `powershell -Command`, `pwsh -Command`, `cmd /c`
- `filesystem-danger`: `rm -rf`, `rmdir /s`, `Remove-Item -Recurse`, `chmod -R`, `chown -R`

## Claude slash commands

Inside Claude Code:

```text
/ifly:status
/ifly:mode silent
/ifly:mode minimal
/ifly:mode normal
/ifly:mode verbose
/ifly:mode caveman
/ifly:guard
/ifly:guard status
/ifly:guard strict
/ifly:guard project
/ifly:guard open
/ifly:guard off
```

After changing installed plugin files:

```text
/reload-plugins
```

## Common setups

Strict global default:

```powershell
.\dist\ifly.exe mode minimal
.\dist\ifly.exe guard strict
.\dist\ifly.exe config set guard.lockdown true
```

Loose personal default, strict one project:

```powershell
.\dist\ifly.exe guard open
cd C:\Users\John\my-important-project
.\dist\ifly.exe guard --project strict
```

Normal mode globally, verbose in one repo:

```powershell
.\dist\ifly.exe mode normal
cd C:\Users\John\my-project
.\dist\ifly.exe mode --project verbose
```

Block dangerous git globally:

```powershell
.\dist\ifly.exe block preset git-danger
```

Block publish/deploy only in a project:

```powershell
cd C:\Users\John\work-project
.\dist\ifly.exe block add --project "npm publish"
.\dist\ifly.exe block add --project "terraform apply"
.\dist\ifly.exe block add --project "terraform destroy"
```

Check current state and policy:

```powershell
.\dist\ifly.exe status
.\dist\ifly.exe guard status
.\dist\ifly.exe block list
.\dist\ifly.exe paths
```
