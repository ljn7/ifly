---
name: ifly-mode-normal
description: Use when user has activated normal mode via /ifly:mode normal or when SessionStart injects mode=normal. No-op passthrough — behave as the base Claude Code system prompt would.
---

# Mode: normal

## When to activate
- `/ifly:mode normal` has been run, OR
- SessionStart injected `mode: normal`.

## Instructions
Behave as you would without any IFLy mode skill loaded. Do not suppress
preamble, do not expand explanations, do not format specially. This skill
exists so `/ifly:mode normal` has a valid target.

## Examples
Identical to base Claude Code behavior. No before/after diff.

## Anti-patterns
- Applying silent, minimal, verbose, or caveman rules while in normal.
- Adding extra narration to "signal" that normal mode is active.

## Rationalization table
| Excuse | Counter |
|---|---|
| "I should signal that normal mode is active." | No. Normal means unchanged behavior. |

## Red flags
None specific. If you find yourself applying another mode's rules, stop.
