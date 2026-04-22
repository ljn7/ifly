# IFLy

**I'm Feeling Lucky** — safety boundaries and verbosity control for Claude Code.

IFLy gives you hard guardrails on destructive shell commands, file edits, and
network calls; a five-mode verbosity dial (silent → minimal → normal → verbose
→ caveman); and an easter egg you can probably guess.

## Status

`v0.1.0` in development. Plugin tree is complete; the companion CLI ships
separately.

## Install

Three paths, in order of convenience:

### 1. Marketplace (planned)

```bash
claude plugin install ljn7/ifly
```

(Pending publication to the official Claude plugin marketplace.)

### 2. Local plugin directory

```bash
git clone https://github.com/ljn7/ifly ~/.claude/plugins/ifly
claude --plugin-dir ~/.claude/plugins/ifly
```

### 3. Companion CLI (planned)

The `ifly` Go CLI handles installation, config validation, and (on Linux) a
real OS-level sandbox. Installer detects your Claude config root and wires up
the plugin and hooks. Tracked in the Plan B implementation plan.

To compile the CLI from this checkout:

```bash
./build-cli.sh      # Linux/macOS/Git Bash
build-cli.bat      # Windows cmd.exe
```

## 60-second quick start

After installing the plugin in a project, in any Claude Code session:

```
/ifly:mode minimal       # one-line summaries, no headers
/ifly:guard strict       # block destructive ops outside the project
/ifly:status             # see what's active
```

For the same state changes without spending Claude slash-command tokens, run
the companion CLI in your terminal:

```bash
ifly mode minimal
ifly guard strict
ifly mode --project verbose
ifly guard --project project
ifly status
ifly paths
ifly block presets
```

Now the guard hook will block things like `rm -rf /etc`, `git push --force`,
or a shell `curl` (in strict mode), while letting normal in-project work
through. Try `/ifly:mode silent` for terse output, `/ifly:mode caveman` for
the experience nobody asked for, or `/ifly:mode verbose` when you want
explicit alternatives shown.

## What you get

- **Boundary enforcement.** PreToolUse hook (`hooks/guard.sh`) blocks
  destructive bash verbs, edit/write outside the project, sensitive reads
  (`~/.ssh`, `/etc/shadow`), and shell network commands when in `strict`.
- **Always-blocked patterns.** `git push --force`, `git reset --hard`,
  `rm -rf /`, `dd if=`, `mkfs.*`, `DROP TABLE`, `eval`, `bash -c` — blocked
  at every level except `off`.
- **Per-project policy.** Drop a `.ifly.yaml` in your repo to add directories
  to the allowlist or extend the blocked-command list.
- **Lockdown.** Set `lockdown: true` in your global config and projects can
  only tighten, not loosen, the guard level.
- **Verbosity skills.** Five mode skills change Claude's output shape. Each
  has a documented rationalization table so it survives pressure.
- **Slash commands.** `/ifly:mode`, `/ifly:guard`, `/ifly:status`.
- **CLI policy editing.** `ifly block add`, `ifly block preset`, and
  `ifly paths` manage blocked-command policy and show the Claude plugin target
  without spending slash-command tokens.
- **Project-local state.** `ifly mode --project ...` and
  `ifly guard --project ...` write `.ifly-state.yaml`, ignored by git by
  default, and hooks prefer it when present.

Slash commands are intentionally thin wrappers around `hooks/ifly-state`.
They are marked `disable-model-invocation: true` so Claude cannot call them
autonomously, but manual slash-command use still consumes tokens because
Claude Code expands command files into the conversation.

## Disclaimer

IFLy is a safety layer, not a guarantee. Claude Code hooks are advisory checks
inside Claude's tool flow; they are not an operating-system sandbox and cannot
protect against every shell feature, parser edge case, external program, or
manual command you run yourself. The guard can reduce common destructive
mistakes, especially around project boundaries, but you should still review
commands before approving them and keep backups for important work.

On Linux, `ifly sandbox` offers stronger isolation when available. On Windows
and macOS, IFLy relies on hook enforcement and configured policy. Do not use
IFLy as the only control for high-risk environments, secrets, production
systems, or irreversible data operations.

## Configuration

Three layers, last wins for scalars; lists union additively:

1. `defaults.yaml` (shipped with the plugin)
2. `~/.config/ifly/config.yaml` (user-global, optional)
3. `<project>/.ifly.yaml` (per-project, optional)
4. Env overrides: `IFLY_MODE`, `IFLY_GUARD`

See `docs/plugin.md` for the full schema and `.ifly.yaml.example` for a
copy-pasteable starter.

## Documentation

- [`INSTALL.md`](INSTALL.md) — setup for Windows, Linux, and macOS.
- [`USAGE.md`](USAGE.md) — day-to-day usage after installation.
- [`COMMANDS.md`](COMMANDS.md) — CLI and slash-command combinations with
  examples.
- [`COVERAGE.md`](COVERAGE.md) — what the current automated and manual tests
  cover.
- [`docs/plugin.md`](docs/plugin.md) — every config knob, every hook, every
  skill, every slash command.
- [`docs/security.md`](docs/security.md) — honest threat model. What the hook
  can and can't enforce.
- [`CHANGELOG.md`](CHANGELOG.md) — keep-a-changelog format.

## IFLy vs Superpowers

[Superpowers](https://github.com/obra/superpowers) is a workflow framework:
brainstorming, planning, subagent-driven TDD, code review. IFLy is a safety
and verbosity layer. They do different things and play well together —
install both if you want both. IFLy will not duplicate Superpowers'
workflow skills.

## License

MIT. See [`LICENSE`](LICENSE).

## Name

`IFLy` — capital I, capital F, capital L, lowercase y. The CLI binary, the
plugin manifest name, and the slash-command prefix are all lowercase `ifly`
per CLI convention. Expansion: *I'm Feeling Lucky*.
