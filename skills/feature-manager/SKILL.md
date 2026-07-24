---
name: feature-manager
description: >
  Decompose a feature specification (.md file) into a structured roadmap of goals
  and tasks, written to /tmp/FEATURE-Manager-{project_name}-{timestamp}.md.
  Use this skill whenever the user provides a markdown spec describing a feature,
  product, or system and asks to "break it down", "decompose it", "create goals
  and tasks", "generate a roadmap", "plan the tasks", "split into goals", or
  similar. Also trigger when the user says "feature-manager this", "run
  feature-manager on …", "turn this spec into a task list", or asks Claude to
  plan work from a spec file. The skill ends after presenting the roadmap;
  delegation to subagents happens only after explicit user confirmation.
compatibility: []
---

# Feature Manager

Transform a markdown feature specification into a structured roadmap of goals
and tasks wired to a checklist file the user can pick up and execute.

## When to use

The user has handed you one or more markdown files describing a feature, and
they want it broken down into actionable work. The output is always a single
roadmap file at a fixed path; you do not execute the tasks unless the user
explicitly asks afterwards.

## When NOT to use

- The user wants tasks executed, not planned. That is `goal-prep` / `coordinator`.
- The spec is not markdown, or the user wants a different output format.
- The user wants to review / validate an existing spec — that is `spec-review`.
- The user wants a quick ad-hoc todo list from a chat message — that is plain
  conversational planning, not feature-manager.

---

## Prerequisites (hard stop)

Before doing anything else, verify that the **graphify** skill is installed in
the current environment. The original spec also accepts `codebase-memory-mcp`
as an equivalent; treat that as an acceptable alternative if it is present.

How to check:
- Look for `~/.claude/skills/graphify/SKILL.md` (or the symlink variant).
- Look for `codebase-memory-mcp` in the user's registered MCP servers.

If neither is available, **stop immediately**. Tell the user:

> Feature-manager needs the `graphify` skill (or the `codebase-memory-mcp` MCP)
> to consult the surrounding codebase before decomposing the spec. Please
> install one of them and re-run.

Do not produce a partial roadmap. Do not improvise a substitute.

---

## Inputs

Accept any of:

1. A spec file path: `feature-manager specs/Feature-manager.md`
2. A spec file path with an explicit project name: `feature-manager specs/auth.md --project auth-service`
3. Multiple spec files (decompose them together as one feature):
   `feature-manager specs/auth.md specs/billing.md`

If the user does not provide a file path, ask before running. Do not guess.

---

## Output

A single markdown file at:

```
/tmp/FEATURE-Manager-{project_name}-{timestamp}.md
```

Where:
- `project_name` is kebab-case. If the user did not pass `--project`, infer
  it from the spec filename (strip `.md`, kebab-case it) or the first `# Title`
  in the spec. If neither is meaningful, ask the user.
- `timestamp` is a Unix timestamp in seconds (e.g. `1721846400`). Use the
  `date +%s` shell command — it's unambiguous and avoids locale surprises.

---

## Roadmap file format

The file must follow this exact template. Sections in `[]` are filled in.

```markdown
# [TITLE OF FEATURE]

[One-paragraph description of what the feature does and what these goals
collectively achieve.]

## ROADMAP

### GOALS #1
[1–2 line description of the goal.]

#### TASKS
- [ ] task NO. 1
- [ ] task NO. 2
- [ ] task NO. 3

### GOALS #2
[1–2 line description.]

#### TASKS
- [ ] task NO. 1
- [ ] task NO. 2
- [ ] task NO. 3
```

Rules:
- Goals are numbered `#1`, `#2`, `#3`, ...
- Tasks are numbered `NO. 1`, `NO. 2`, ... under their goal.
- Goal descriptions are 1–2 lines, never longer.
- Every `### GOALS #N` MUST be followed by a `#### TASKS` block.
- Tasks are atomic sentences starting with a verb, no trailing punctuation.
- Use `- [ ]` (unchecked) for every task — never `- [x]`.

---

## Algorithm

Follow these steps in order. The user should see progress, not a long silence.

### Step 1 — Verify prerequisites

Confirm graphify (or codebase-memory-mcp) is present. Hard stop if missing.

### Step 2 — Collect inputs

Determine:
- Spec file path(s)
- Project name (ask if not inferable)

### Step 3 — Create the output file

Create the empty file at the target path with a placeholder header. This gives
something concrete to fill in and lets the user see the location early.

```bash
mkdir -p /tmp
PROJECT_NAME="<kebab-case-name>"
TS=$(date +%s)
OUT="/tmp/FEATURE-Manager-${PROJECT_NAME}-${TS}.md"
touch "$OUT"
echo "$OUT"
```

### Step 4 — Read the spec(s)

Use `read` to load each spec file completely. Hold two representations:
1. **Overall purpose** — what is this feature trying to achieve?
2. **Per-section breakdown** — what does each heading, list, or instruction
   require?

If the spec is large, read in chunks until every section is in context.

### Step 5 — Consult the codebase via graphify

If graphify is available, ask it about the relevant area of the codebase. This
context prevents goals/tasks from being unaware of system constraints. Skip if
the spec is fully self-contained (e.g. greenfield product with no existing
code).

### Step 6 — Decompose into goals

Draft the first pass of goals. For each goal:
- 1–2 line description.
- Covers a coherent slice of the feature (not arbitrary).
- Starts with a noun or verb phrase, not a vague abstraction.

### Step 7 — Split overly-large goals

Walk through every goal. If a goal covers two distinct concerns that could
ship independently, split it. Sign of a goal that should split:
- The description uses "and" between two unrelated actions.
- You can describe two separate acceptance criteria for it.
- A reasonable engineer would estimate it at more than a week of work.

### Step 8 — Merge tiny goals

Walk through every goal. If a goal is too small to stand alone (fewer than
3 tasks, or trivially mergeable with another), merge it with the most
related goal.

### Step 9 — Repeat splitting/merging until stable

If any merges happened, go back to Step 7. Stop when a full pass produces no
changes.

### Step 10 — Generate tasks

For each goal, generate 3–7 tasks. Each task:
- Is a single, atomic action.
- Starts with a verb ("Add", "Write", "Wire", "Test", "Configure").
- Ends without a period.
- Maps to a specific section of the original spec.
- Is independently verifiable (you can say "this is done" when it's done).

### Step 11 — Enforce the 3-task minimum

Walk through every goal. If any goal has fewer than 3 tasks:
- First, try to expand the goal scope (it probably was over-merged).
- Second, split the goal into smaller goals until each has ≥3 tasks.
- Third, merge the goal with a sibling that has related work.

If after all three options a goal still has <3 tasks, **stop and ask the
user**. Do not ship a goal with fewer than 3 tasks.

### Step 12 — Cross-check coverage

Compare the goal/task list against the original spec. For every section,
requirement, or acceptance criterion in the spec, confirm at least one task
covers it. If any spec section is missing:

- Run a second pass (Steps 5–11) focused on the missing section.
- Append new goals/tasks to the file.
- Cap at **3 total passes** to avoid runaway iteration.

After 3 passes, if sections are still missing, stop and tell the user which
sections you couldn't cover and why.

### Step 13 — Write the file

Use `write` to emit the full roadmap file at the target path. Overwrite the
placeholder.

### Step 14 — Show the user

Print:
1. The output file path.
2. A short summary: number of goals, number of tasks.
3. The full roadmap file content (so they can read it inline).

### Step 15 — Prompt before delegating

End the skill with this exact prompt:

> Roadmap is ready at `<path>`. Want me to start delegating the tasks to
> subagents now? (yes / no / edit first)

Do NOT spawn subagents until the user says yes. If they say "edit first",
wait for their changes and re-run only Step 13.

---

## Quality bar

A good feature-manager output has these properties:

- **Complete**: every section of the spec appears as at least one task.
- **Coarse-grained goals**: each goal feels like a unit of work, not a
  sub-task. 3–10 goals is typical for a single feature.
- **Fine-grained tasks**: each task is a single working session. 3–7 tasks
  per goal.
- **No orphan goals**: every goal has 3+ tasks.
- **No orphan tasks**: every task belongs to exactly one goal.
- **Independently meaningful**: each goal could in principle be shipped
  without the others (even if the order matters).
- **Atomicity**: a task is done when one specific thing is true.

## Anti-patterns

- **Goal = task**: "Add user authentication" as a goal with one task. Split.
- **Vague task**: "Set up stuff". Rewrite as a concrete action.
- **Mega-goal**: "Build the entire backend" — split by capability.
- **Task that spans multiple days of work**: split.
- **Task with no clear done condition**: rewrite or remove.
- **Goal with no tasks**: spec drift — remove it or expand it.
- **Roadmap that ignores 80% of the spec**: missing coverage — go back to Step 12.

---

## Reference

For more on the underlying decomposition heuristics and the history of this
skill format, see `references/decomposition-heuristics.md`.
