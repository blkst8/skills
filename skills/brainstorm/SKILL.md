---
name: brainstorm
description: >
  Evaluate a developer's engineering problem and recommend the best approach
  with a score. Use this skill whenever a developer describes an issue they
  need to solve and optionally proposes an initial idea — the typical case
  is something like "we have a problem with X, should we do Y?" or "I need
  to solve X, my idea is Y". The skill weighs approaches against four
  measurements (complexity, dependency, maintainability, extendability) from
  two perspectives (fast MVP and quality) and recommends one with a rate
  out of 5. Use this skill even when the user does not explicitly ask for
  "brainstorm" but is asking for help choosing between engineering approaches.
  Do NOT use this skill when the user just wants code written, when the
  issue has no real tradeoff, or when the user wants raw ideation without
  scoring.
compatibility: []
---

# Brainstorm — approach evaluation

Given a developer who has an issue and optionally an initial idea, generate
candidate approaches, score them, and recommend one with a rate. The skill
does not write code. It returns a structured recommendation the developer
acts on.

## Inputs

Read the user's prompt and extract:

- **Issue** — the problem to solve (1-3 sentences).
- **Initial idea** — their proposed solution (optional).

If the issue is too vague to evaluate (e.g. "improve performance" with no
context), ask one clarifying question before proceeding. Do not ask if you
can make reasonable assumptions.

## Output

Every run returns, in plain text (terminal-friendly, no nested lists
deeper than 2):

1. The recommended approach with an overall rate.
2. 1-2 runner-ups with their rates.
3. Discarded approaches, summarized by top reason.
4. A short rationale (one sentence per measurement).

Use this template:

```
Issue: <user's issue>
Your idea: <user's initial idea, or "none">

Recommended: <approach description>
Rate: <X.X / 5>
Path: MVP | Quality

Why:
- Complexity <score>/5 — <one sentence>
- Dependency <score>/5 — <one sentence>
- Maintainability <score>/5 — <one sentence>
- Extendability <score>/5 — <one sentence>

Runner-up: <approach description> — Rate <X.X / 5> (<one-line reason it lost>)
Discarded: <count> approaches. Top reason: <one common reason>.
```

## Measurements

Score every approach 1-5 on four axes. Higher is better aligned with the
goal. Use these rubrics verbatim — do not invent your own:

### Complexity

How much cognitive overhead does this approach impose on future readers
and modifiers?

- 1 — magic. Deeply clever. Requires domain knowledge to follow.
- 2 — non-obvious, but documented.
- 3 — standard pattern. Readable with effort.
- 4 — straightforward. Follows project conventions.
- 5 — boring. Obvious. Everyone has seen this before.

### Dependency

How much new dependency burden does this approach add?

- 1 — adds 3+ new external dependencies.
- 2 — adds 1-2 new dependencies.
- 3 — adds 1 well-known dependency already common in the ecosystem.
- 4 — uses dependencies the project already has.
- 5 — zero new dependencies. Stdlib or built-in only.

### Maintainability

How easy is this approach to keep working over months/years?

- 1 — fragile. Will break under expected changes.
- 2 — needs careful attention to keep working.
- 3 — stable under normal use.
- 4 — self-documenting or well-tested.
- 5 — boring, tested, hard to break.

### Extendability

How easily can this approach grow to cover adjacent future needs?

- 1 — dead end. Rewrite required for any change.
- 2 — narrow. Supports only the current case.
- 3 — supports the current case and obvious adjacent cases.
- 4 — flexible. Clear extension path.
- 5 — composable. Slots into larger systems.

## Perspectives

Each approach is evaluated from two perspectives and combined at the end.

### MVP Path

The fastest path to a working solution.

- Allowed: existing project deps, stdlib, simple patterns.
- Forbidden: new external dependencies, large refactors, speculative
  abstractions.
- Weight heavily: Complexity, Dependency.
- Allowed to be lossy on: Maintainability, Extendability.

### Quality Path

The best long-term solution.

- Allowed: new dependencies when justified, refactors, abstractions.
- Forbidden: gratuitous complexity. Magic without documentation.
- Weight equally: all four measurements.
- Hard rule: every measurement must score >= 3, otherwise the node is
  pruned (it fails the quality bar).

## Algorithm

Two parallel evaluation streams, one per perspective. Each streams
candidate variations of the user's initial idea (or seed approaches if
no idea was given).

### Constants

- `BEAM_WIDTH = 3` — max candidate variations per parent idea.
- `MAX_DEPTH = 3` — root counts as depth 0.
- `MAX_NODES = 12` per stream.
- `CONVERGENCE_THRESHOLD = 0.5` — weighted-score gap to consider the two
  streams close enough to merge.

### Steps

1. **Seed both streams.**
   - If the user provided an initial idea, each stream's root is that idea
     interpreted under its perspective's constraints.
   - If no initial idea, generate 2-3 candidate roots per stream from the
     issue alone.

2. **Branch in parallel.** At each depth level, for each surviving parent,
   generate up to `BEAM_WIDTH` child candidates. Each child must be a
   meaningful variation of the parent, not a trivial tweak.

3. **Score every new candidate** on all 4 measurements. Quality-path
   candidates scoring < 3 on any measurement get their subtree pruned.

4. **Stopping condition.** Stop when ANY of:
   - `MAX_DEPTH` reached.
   - `MAX_NODES` reached per stream.
   - Both streams converge on the same approach (description similarity
     >= 80%).
   - Top-scored leaves from each stream have a weighted-score gap
     < `CONVERGENCE_THRESHOLD`.

5. **Select connectors.** Pick the highest-weighted-score leaf from each
   stream. These are the two connector candidates — one MVP, one Quality.

6. **Apply merge rule.** See below.

### Merge rule

Two cases:

**Case A — MVP and Quality describe the same approach, or one is a strict
subset of the other.**
Recommend the MVP version. It's cheaper and simpler. Note that both
perspectives converged.

**Case B — they diverge.**
Recommend the connector with the higher weighted score, subject to:
- Quality connector must score >= 3 on every measurement (already
  enforced by pruning).
- If Quality scores higher, prefer it (it cleared the bar).
- If MVP scores higher or equal, prefer MVP (it's cheaper and the Quality
  bar wasn't worth paying for).

### Weighted score

```
weighted = complexity * 0.35
         + dependency * 0.25
         + maintainability * 0.20
         + extendability * 0.20
```

Complexity is weighted highest because cleverness is the most common
long-term regret.

### Final rate

The recommended approach's rate = average of its 4 raw measurement scores,
rounded to 1 decimal.

```
rate = (complexity + dependency + maintainability + extendability) / 4
```

The weighted score is used only for ranking candidates within and across
the two streams. The final displayed rate is the raw average — that is
what the user sees.

## Examples

### Example 1 — MVP wins

Issue: "Need to validate email addresses in user signup form."
Initial idea: "Use a regex."

```
Issue: Validate emails in signup form
Your idea: Use a regex

Recommended: Send a confirmation email; do basic regex shape check client-side.
Rate: 4.5 / 5
Path: MVP

Why:
- Complexity 5/5 — well-understood pattern, every dev has seen it.
- Dependency 5/5 — zero new deps. Uses the built-in mailer.
- Maintainability 4/5 — clear logic, easy to debug.
- Extendability 4/5 — can add DKIM/SPF checks later without rewrite.

Runner-up: regex + `email-validator` package — Rate 3.8 / 5 (Quality path, but extra dep is overkill at current scope).
Discarded: 4 approaches. Top reason: overengineered for the problem.
```

### Example 2 — Quality wins

Issue: "Migrate the auth system from session cookies to OAuth2 with 3 identity providers."
Initial idea: "Add OAuth2 with `requests-oauthlib`."

```
Issue: Migrate auth to OAuth2 with 3 IdPs
Your idea: Add OAuth2 with requests-oauthlib

Recommended: Use Authlib + structured provider adapters; isolate per-IdP config.
Rate: 4.0 / 5
Path: Quality

Why:
- Complexity 4/5 — abstractions explicit, each IdP is a separate module.
- Dependency 4/5 — one well-maintained lib (Authlib), replaces three smaller ones.
- Maintainability 5/5 — provider config isolated, easy to update one without breaking others.
- Extendability 5/5 — adding a 4th IdP is a config + adapter file, no core change.

Runner-up: requests-oauthlib directly, one provider at a time — Rate 2.5 / 5 (MVP, but maintainability and extendability tank the moment a 2nd and 3rd provider are added).
Discarded: 3 approaches. Top reason: hidden coupling between IdPs.
```

## How many candidates to actually generate

You do not need to literally enumerate `MAX_NODES = 12` candidates. In
practice:

- If the issue and initial idea are concrete, generate 2-3 MVP variations
  and 2-3 Quality variations. Pick the best from each stream. Done.
- If the issue is open-ended, run the full branching algorithm to
  surface options the user might not have considered.
- Always stop early when the merge rule has a clear answer (Case A or
  an obvious Case B). Do not pad with weak candidates.

## What not to do

- Do not write or modify code. This skill only recommends.
- Do not invent scores. If you cannot justify a score, score it lower and
  say why.
- Do not pad the runner-ups. If only one candidate is worth showing,
  show one.
- Do not include the full tree in the output. Only the recommendation,
  the runner-up, and the discarded count.
- Do not contradict the rubrics. If you write "uses existing deps" the
  dependency score must be >= 4.