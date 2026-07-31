---
name: spec-review
description: >
  Review and validate agent specification (.md) files by walking the user
  through one validation question at a time. Use this skill whenever the user
  provides a .md file containing an agent workflow, prompt template, or
  behavioral specification and asks you to review it, check it, validate it,
  improve it, or make sure it's correct. Also use when the user says things
  like "go over this spec", "review my agent spec", "what does this spec do",
  "check this agent configuration", or "make sure my spec is right". If the
  user gives you a file path that ends in .md and it describes agent behavior,
  instructions, or prompts, consult this skill before proceeding.
compatibility: []
---

# Agent Specification Review

When the user provides a path to an agent specification file (.md), walk them
through the spec one validation question at a time. The user answers each
question with one of:

- **yes** — the spec is right as-is for this point. Move on.
- **no** — the spec is wrong on this point, but the user does not want to
  say how. Flag it as unresolved and move on.
- **tell you correct way** — the user pastes the correct text / instruction.
  Apply it to the spec.

This skill is interactive. It does not silently edit files.

## Step 1 — Read the spec completely

Use `read` to load the entire file. Do not summarize from a partial read.
If the file is very large, read it in chunks until you have seen every section.

Hold two representations in your mind:
1. **General summary** — what is the overall purpose and scope?
2. **Per-section breakdown** — what does each heading, list, or instruction
   actually require the agent to do?

If the file does not exist or is not an agent specification, say so and stop.
Do not invent or assume content.

## Step 2 — Present the general view

Show the user 2–3 sentences describing what the spec does at a high level.
This orients them before the questions begin. Keep it brief — the questions
are the main event.

Do not present a detailed per-section breakdown here. Save that detail for
the questions in Step 3, which surface it one piece at a time.

## Step 3 — Ask validation questions, one at a time

Construct 3–10 short questions that test whether the documented behavior
matches the user's real intent. Each question covers one distinct part of
the spec.

Example questions:
- "In section 4 the spec says the agent should retry 3 times. Do you actually
  want retries, or should it fail immediately?"
- "The spec mentions storing results in `/tmp/output.json`. Is that path
  fine, or should it be configurable?"
- "It looks like the agent is supposed to log at every step. Is that the right
  level of verbosity, or should logging be optional?"

### Format

For each question:
1. Quote the relevant spec line or section verbatim so the user can see
   exactly what is being asked about.
2. State the question in one sentence.
3. End with the three-option prompt:

   ```
   y / n / tell you correct way
   ```

### Flow

Ask one question. **Wait for the user's answer before asking the next.** Do
not batch questions or proceed to the next one speculatively.

For each answer:
- **yes** → record `confirmed`, move to the next question.
- **no** → record `unresolved: <one-sentence description of the gap>`,
  move to the next question. Do not invent a fix.
- **tell you correct way** → the user provides the corrected text. Record
  `override: <user's correction>` and move to the next question.

Keep the recorded list in your head; you will use it in Step 4.

### If the user interrupts the loop

If the user answers something that does not match any of the three options
(e.g. "what do you think?", "skip this", "go back to question 2"), handle it
without breaking the loop:
- "skip this" → treat as **no** and move on.
- "go back to question N" → re-ask question N, ignoring the previous answer.
- Free-form clarification → answer briefly, then re-ask the same question.

## Step 4 — Apply changes after the last question

After the final question is answered, process the recorded list:

- **`confirmed` entries**: no change.
- **`unresolved` entries**: leave the spec line as-is, but call it out in
  the final summary as a flagged gap. Do not invent a fix.
- **`override` entries**: apply the user's correction to the spec.

Apply overrides with `edit` (preferred — targeted change) or `write`
(only if the structure needs large changes). When editing:
- Match the user's original writing style and formatting conventions.
- Make the smallest change that captures the user's intent.
- Do not introduce changes the user did not request.

If you make any edits, record them in a running list:

```markdown
### Changes made
- Section 2: replaced retry count from 3 to 0 (user override)
- Section 4: replaced hard-coded path with `{output_path}` (user override)
```

## Step 5 — Final summary

Report back to the user with:
1. The general view (one sentence recap).
2. The number of questions asked and how they were resolved
   (confirmed / unresolved / override).
3. The list of changes made, if any. If no changes were needed, say so
   explicitly.
4. The list of unresolved gaps, with a one-line note on each.

Keep the summary short. The user has already seen the questions and answers.

## Important rules

- One question at a time. Wait for the answer before asking the next.
- Always confirm with the user before changing the spec. This skill requires
  the interactive question phase — do not silently edit the file.
- Only modify files that the user has explicitly given you or confirmed you
  may change.
- Maintain the spec's structure and formatting as closely as possible when
  editing.
- If the user answers "no", do not guess what they meant. Flag it as
  unresolved and move on.
- If the user gives "tell you correct way" but the correction is vague
  ("just make it better"), ask one short follow-up to pin down what they
  mean before applying. Do not apply ambiguous corrections.