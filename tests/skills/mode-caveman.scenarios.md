# Mode: caveman — pressure scenarios

## Coverage note (2026-04-19)

GREEN dispatches deferred — subagent dispatch quota was exhausted during
this session. Both RED captures and GREEN verifications go to QA.

Caveman is the most distinctive mode: maximally compressed English in
prose, but generated CODE stays idiomatic. The dual-direction constraint
(compress prose, leave code alone) is the most-likely-violated rule, so
the REFACTOR phase should specifically probe a "explain this code" case
where the temptation is to caveman-ize the code's docstrings too.

## Scenario 1: Explain a function (planned)

Prompt: "[paste merge_sort] Explain this function."

Expected violation (baseline): grammatical English paragraphs with
articles, pronouns, filler ("This function recursively sorts the array
by dividing it in half and merging...").

## Scenario 2: Report a change (planned)

Prompt: "You just edited `config.js` to increase the timeout. Tell me
what you did."

Expected violation: "I changed the timeout in config.js from 1000 to
5000." (Articles, pronouns, full sentence.)

## Scenario 3: Answer a factual question (planned)

Prompt: "How does Python's GIL affect CPU-bound threading?"

Expected violation: multi-paragraph explanation with transitional words.

## Pass criteria when GREEN runs

- Scenario 1: e.g. "sorts array. O(n log n). recursive divide. mutates
  nothing. returns new sorted list."
- Scenario 2: e.g. "bumped timeout 1000 → 5000 in config.js:22."
- Scenario 3: e.g. "GIL serializes bytecode. cpu threads run one at a
  time. use multiprocessing for parallel cpu. threads ok for io."

## REFACTOR — critical edge

**Code-vs-prose boundary:** the most subtle violation is applying caveman
rules inside generated code's docstrings or comments. The skill's
Example 5 and red-flags list explicitly call this out. QA pass must run
a scenario like "write a Python function and explain it briefly" and
verify the docstring is normal English while the surrounding prose is
caveman.

Other REFACTOR scenarios (per plan):
- Long factual question → resist padding for politeness.
- Error report → "failed because X" not "the operation failed because of X".
- Apology scenario → no apologies at all; state the fact.

## Why the skill ships before formal GREEN

Skill structure is internally consistent and follows the writing-skills
discipline. The dual-direction rule (prose vs code) is the novel piece,
explicitly documented in Example 5 and Red flags. Pre-merge QA is the
gate where formal RED/GREEN evidence is captured.
