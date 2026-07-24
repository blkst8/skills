# Decomposition Heuristics

Reference guide for feature-manager when going from a spec to goals and tasks.

## Goal vs task — the test

A **goal** answers: "What is a human-sized piece of work that, when done, delivers
something a user could see or a stakeholder could sign off on?"

A **task** answers: "What is the smallest action a single engineer (or single
subagent) can complete in one focused session?"

If the answer to "is this goal done?" is a yes/no demo, it's a goal. If the
answer is "we'd have to think about what else", it isn't.

## When to split a goal

A goal should split when ANY of these is true:

1. The goal description contains two independent actions joined by "and".
   - "Add login and reset password" → split → "Add login" + "Reset password".
2. You can write two distinct acceptance criteria for it.
   - Acceptance 1: "User can sign in."
   - Acceptance 2: "User can reset password."
3. It would take more than ~1 week of focused work.
4. Different roles would own different parts (frontend vs backend, infra vs UX).
5. It could ship independently of the rest of the feature.

## When to merge goals

A goal should merge into a sibling when ANY of these is true:

1. It has fewer than 3 tasks and won't grow (truly trivial).
2. It is fully dependent on another goal and brings no independent value.
3. Splitting it produced a goal whose description is just a sub-step of the
   other ("Add the button" + "Make the button clickable").
4. The two goals would always be tested, reviewed, and shipped together.

## When to split a task

A task should split when:

1. You can't write "this is done" in one sentence.
2. It would touch more than ~3 files in unrelated areas.
3. Different people would naturally own different parts.
4. It combines "implement" and "test" — split into two.
5. It uses "and" between two distinct actions.

## When to merge tasks

A task should merge into a sibling when:

1. It has no independent acceptance criterion.
2. Doing the other task inevitably completes this one.
3. Splitting produced a "set up" / "use" pair — they belong together.

## The "3 task minimum" rule

The spec says every goal must have ≥3 tasks. This is enforced because goals
with 1–2 tasks are usually:

- **Tasks masquerading as goals** — they should be merged into a parent.
- **Over-decomposed** — the work was too aggressively split.

Try to expand first. If the goal genuinely doesn't grow, merge it.

## Coverage check

When reviewing the spec, every section/requirement/bullet of significance
should map to at least one task. To do this:

1. List every numbered/headed section in the spec.
2. List every requirement, "must", "shall", or acceptance criterion.
3. For each one, find a task that claims it. If none does, generate one.

A 90% coverage is failure. The remaining 10% covers optional or implied
behavior.
