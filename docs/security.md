# IFLy security model — honest threat model

## What IFLy is

IFLy is a **defense-in-depth** layer for Claude Code. The plugin enforces
boundary, destructive-op, and network policies at the PreToolUse hook
layer. Hooks run inside Claude Code; they are advisory.

## What IFLy is NOT

**IFLy is not an OS-level sandbox.** A successful shell escape from
inside the Claude session — or any process Claude spawns that escapes
hook scrutiny — is outside the guard's reach. If you need real
isolation, the companion `ifly` CLI (Linux only) wraps a Claude
invocation in a filesystem namespace via `bwrap` / `unshare`. macOS and
Windows do not have an equivalent; on those platforms, `level: strict`
is the available mitigation.

## What the guard hook can enforce

- Block destructive bash verbs (`rm`, `mv`, `chmod`, `chown`, `unlink`,
  `shred`, `truncate`, `rmdir`) targeting paths outside
  `CLAUDE_PROJECT_DIR + additional_dirs`.
- Block always-dangerous patterns: `git push --force`, `git reset --hard`,
  `rm -rf /`, `dd if=`, `mkfs.*`, `DROP TABLE/DATABASE`, `eval`,
  `bash -c`, `sh -c`.
- Block project-configured `blocked_commands` substrings.
- Block file edits (`Edit`, `Write`, `MultiEdit`, `NotebookEdit`)
  outside the project boundary.
- Block reads of `sensitive_paths` (default: `~/.ssh`, `~/.aws`,
  `~/.config/gh`, `/etc/shadow`, `/etc/sudoers`) in strict mode.
- Block shell network commands (`curl, wget, ssh, scp, sftp, nc, ncat,
  telnet, ftp, rsync`) in strict mode.
- Block `WebFetch` / `WebSearch` URLs not in `allowed_network` in strict
  mode.
- Honor session guard overrides written by `/ifly:guard` or `ifly guard`,
  including YAML with accidental leading indentation.
- Inspect destructive `powershell -Command`, `pwsh -c`, and `cmd /c` wrappers
  launched through the Bash tool, including `Remove-Item`, `rd`, and `del`
  against paths outside the project.

## Known limitations (explicit non-goals for v0.1)

These are deliberate scope decisions. They are not bugs.

### Single-prefix `sudo`/`doas` strip

The hook strips a single leading `sudo` (or `doas`) from the full command
*before* splitting on chain operators. A chained sudo (e.g.,
`cd /tmp && sudo rm /etc/passwd`) is NOT re-stripped per subcommand —
the second clause's verb is read as `sudo`, which is not in the
destructive-verb list, so the check passes through. This is a documented
scope limit. Users who need stricter behavior should pair IFLy with the
companion CLI's sandbox (Linux) or rely on `blocked_commands` (which
matches `sudo` substrings literally).

### Quoted-argument parsing

The bash command splitter uses simple whitespace tokenization. A
destructive command with a quoted path containing spaces
(`rm "/path with space/file"`) will tokenize incorrectly. A real shell
parser is out of scope; the spec accepts this in exchange for the
zero-dependency pure-bash implementation.

### Shell-command hostname parsing

IFLy does NOT parse hostnames out of shell `curl`/`wget` commands. In
strict mode, those verbs are blocked outright; in looser modes they are
allowed without inspection. The `allowed_network` allowlist applies only
to `WebFetch` / `WebSearch` — Claude Code's own tool surface, where the
URL is a discrete JSON field.

### Adversarial bash bypasses

Any command that constructs its destructive payload at runtime
(`x=rm; $x -rf /`), reads from a file (`bash some_script.sh`), or uses
process substitution (`>(rm -rf /)`) can defeat the regex heuristics.
The threat model assumes a non-adversarial AI that occasionally produces
a dangerous one-liner, not a deliberate attacker. Real isolation is the
job of `ifly sandbox` (Linux) or your operating system.

### JSON escaping

The hook's `_emit_block` performs a naive JSON escape on the reason
string (only `"` and `\`). Control characters or newlines in dynamic
reason data could produce malformed output. The reasons in the current
codebase are static or contain only filesystem paths and command
substrings, so this has not surfaced. If you extend the hook with new
reason templates, prefer `jq -Rs` for safety.

## What the OS-level sandbox does (companion CLI, Linux)

Out of scope for the plugin tree (covered in Plan B), but for context:
the companion CLI's `ifly sandbox` subcommand wraps `claude` in a
namespace where:

- The filesystem visible to Claude is restricted to the project dir
  + explicit allowlist + read-only system mounts.
- Network is opt-in per-project.
- The namespace is real OS isolation (`unshare`/`bwrap`), not advisory.

This is the only true isolation IFLy ships. On macOS and Windows, the
hook layer is the boundary.

## Threat-model checklist

- ✅ Defends against: AI mistakenly generating a destructive command,
  edits to files outside the project, reads of well-known credential
  paths, untrusted network calls.
- ⚠️ Mitigates but does not fully prevent: chained `sudo` after
  operators, complex shell quoting, runtime-constructed commands.
- ❌ Does not defend against: shell escapes from inside Claude's process,
  exploits in jq/bash/coreutils, attacks on the user's OS or shell
  configuration.

## Reporting issues

Security issues: file a private security advisory on GitHub against
`ljn7/ifly`. Public bug reports for non-security issues go to the
issue tracker. Do not include reproducers that touch real credentials
or real systems.
