# Mode: verbose — pressure scenarios

## Coverage note (2026-04-19)

GREEN dispatches deferred — subagent dispatch quota was exhausted during
this session. RED captures (which only require dispatching to a fresh
unguided subagent) and GREEN verifications (which require dispatching with
the SKILL.md text prefixed) are both deferred to QA.

Verbose mode is the OPPOSITE direction from silent/minimal. The RED
violation we hunt is *terseness when expansion was called for* — the
default Claude tendency to pick one approach silently and ship it without
showing alternatives.

## Scenario 1: Design choice (planned, not run)

Prompt: "Add retry logic to `fetch_user()`. Use exponential backoff."

Expected violation (baseline): subagent picks one approach, implements,
commits. No alternatives discussed. No change summary.

## Scenario 2: Small bug fix (planned, not run)

Prompt: "Fix the off-by-one in `pagination.js:48`."

Expected violation: one-line fix, no reasoning shown.

## Scenario 3: Generated code (planned, not run)

Prompt: "Write a Python function that memoizes by the first argument."

Expected violation: code without docstring or inline comments on the
non-obvious `args[0]` keying.

## Rationalizations to anticipate

Predicted from prior experience (these drive the skill table even before
formal RED):
- "The user didn't ask for alternatives." → Verbose mode requires them.
- "Comments clutter short code." → Not in verbose mode.
- "A change summary is redundant if they can read the diff." → Verbose mode contract.

## Pass criteria when GREEN runs (per scenario)

- Scenario 1: ≥2 alternatives with one-sentence tradeoffs, recommendation
  with one-sentence reason, implementation, "## Summary" section listing
  files changed.
- Scenario 2: name the bug plainly, show fix, one-sentence why.
- Scenario 3: docstring + at least one inline comment on the `args[0]`
  keying decision.

## REFACTOR — explicit override

Critical edge case: verbose mode must NOT override an explicit user
"briefly" / "one line" / "TL;DR" request in a single turn. The skill's
"Explicit user override" section addresses this. QA pass should run a
verbose-active scenario where the user's prompt contains "briefly" and
confirm the reply is brief.

## Why the skill ships before formal GREEN

Skill structure mirrors silent/minimal patterns (which captured RED
evidence and, in silent's case, passed GREEN earlier this session). The
rationalization table and red-flags list are derived from the writing-skills
discipline applied to known verbose-mode failure modes. Pre-merge QA pass
is the gate, not pre-commit.
