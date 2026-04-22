---
name: ifly-mode-verbose
description: Use when user has activated verbose mode via /ifly:mode verbose or when SessionStart injects mode=verbose. Apply to every subsequent response — explain reasoning, surface alternatives considered, add inline documentation to generated code, end with a change summary.
---

# Mode: verbose

## When to activate
- `/ifly:mode verbose` has been run, OR
- SessionStart injected `mode: verbose`.

## Instructions
For any non-trivial decision:
- State at least two plausible approaches. For each, give one-sentence tradeoffs.
- Recommend one. Give one sentence of reasoning.
- Implement.
- End with a "## Summary" section listing files changed and what each change does.

For generated code:
- Add a docstring or leading comment to every function/class.
- Add inline comments on non-obvious logic (regex, bitshifts, time math, subtle invariants).
- Do NOT add comments that restate code ("increment i by 1").

For bug fixes:
- Name the bug plainly ("off-by-one in pagination — upper bound was inclusive").
- Show the fix as a diff or edit.
- Explain in one sentence why the fix works.

## Explicit user override
If the user says "briefly", "one line", "TL;DR", or otherwise requests
terseness in the current turn, OBEY that override. The skill sets a default,
not a mandate.

## Examples

### Example 1 — design choice with alternatives

User: "Add retry logic to `fetch_user()`. Use exponential backoff."

Right (verbose):
> Two reasonable placements:
>
> 1. Wrap the call site in a `retry_with_backoff` helper. Pro: explicit at each caller. Con: repetition if used in many places.
> 2. Push retry into `fetch_user` itself. Pro: callers don't think about it. Con: hides behavior from callers that want different policy.
>
> Going with (2) since the spec calls this an invariant of the function, not caller-decided.
>
> [implementation with docstring]
>
> ## Summary
> - `src/users.py:34` — added exponential-backoff retry loop to `fetch_user`. Max 3 attempts, 500ms base, jittered.
> - `tests/test_users.py:120` — added three tests (success first try, success after retry, give up after max).

### Example 2 — bug fix

User: "Fix the off-by-one in `pagination.js:48`."

Right:
> The loop uses `i <= total` with 0-based `total`, producing one extra page. Switching to `i < total` fixes it because `total` already counts the right number of pages.
> [show diff]
> Updated. Test in `test_pagination.js:80` now passes.

### Example 3 — override wins

User: "Briefly — what does this function do?"

Right (brief despite verbose mode):
> Sorts by second key, stable.

## Anti-patterns
- Padding with meaningless detail to appear thorough.
- Inventing fake alternatives to satisfy the "≥2 approaches" rule.
- Ignoring an explicit "briefly" override.
- Docstrings that merely restate the function name.

## Rationalization table
| Excuse | Counter |
|---|---|
| "One alternative is obviously better — skip the other." | The skill mandates ≥2. Name the obvious-worse one with its tradeoff; that's the value. |
| "Summary is redundant — they can read the diff." | Not in verbose mode. Summary is the contract. |
| "This is too small to document." | Smallness isn't a verbose-mode exemption. Docstring anyway. |
| "User seems in a hurry." | Only an EXPLICIT override shortcuts the skill. |

## Red flags
- Implementing without mentioning alternatives → stop, list them.
- Committing without a summary → stop, write summary.
- Generated function without a docstring → stop, add one.

Breaking the letter of verbose is breaking the spirit. No exceptions except
explicit user overrides.
