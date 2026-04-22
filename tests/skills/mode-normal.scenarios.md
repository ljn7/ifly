# Mode: normal — pressure scenarios

## Scope

`ifly-mode-normal` is a no-op passthrough. There is no behavior to enforce
beyond "do not apply other mode rules". The pressure-test for normal mode
is therefore an *equivalence* test: a fresh subagent given the SKILL.md
prefix should produce output indistinguishable from a fresh subagent given
no skill at all.

## Equivalence test (deferred to QA)

GREEN dispatch deferred — subagent dispatch hit API rate limit during
this session. Run during QA.

### Procedure

1. Dispatch a fresh subagent with prompt P, no skill prefix. Capture reply A.
2. Dispatch a fresh subagent with the SKILL.md text prefixed, then prompt P. Capture reply B.
3. Compare A and B for length, structure (headers/bullets/preamble), and
   tone. They should be substantially equivalent — same general shape, no
   suppression or expansion attributable to the skill.

### Suggested prompt P

> Briefly explain how `git stash` works.

### Pass criteria

- Reply B is NOT terser than reply A by more than ~30%.
- Reply B is NOT longer than reply A by more than ~30%.
- Reply B does NOT remove headers/bullets that reply A has.
- Reply B does NOT add headers/bullets that reply A lacks.
- Neither reply contains text like "normal mode active" — that would be a
  bug per the skill's anti-patterns.

## Why no RED capture

There is no violation to capture. The skill defines absence of behavior.
The test is structural, not behavioral.

## REFACTOR

If the equivalence test fails, the skill text itself is the bug — not
some rationalization to plug. Revisit Instructions and Anti-patterns.
