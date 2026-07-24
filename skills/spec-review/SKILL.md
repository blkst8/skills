---
name: spec-review
description: >
  Review and validate agent specification (.md) files. Use this skill whenever
  the user provides a .md file containing an agent workflow, prompt template,
  or behavioral specification and asks you to review it, check it, validate it,
  improve it, or make sure it's correct. Also use when the user says things
  like "go over this spec", "review my agent spec", "what does this spec do",
  "check this agent configuration", or "make sure my spec is right". If the
  user gives you a file path that ends in .md and it describes agent behavior,
  instructions, or prompts, consult this skill before proceeding.
compatibility: []
---

# Agent Specification Review

When the user provides a path to an agent specification file (.md), follow this
process to review it thoroughly and, if needed, update it so it accurately
reflects the user's intent.

## Step 1 — Read the spec completely

Use `read` to load the entire file. Do not summarize from a partial read.
If the file is very large, read it in chunks until you have seen every section.

Hold two representations in your mind:
1. **General summary** — what is the overall purpose and scope?
2. **Per-section breakdown** — what does each heading, list, or instruction
   actually require the agent to do?

## Step 2 — Present the two-view review

Show the user:
1. **General view** (2–3 sentences): what the spec does at a high level.
2. **Detailed view** (bullet list): for each major section or instruction
   block, state exactly what behavior it mandates.

Be concrete. Don't say *"It handles user authentication"*; say *"Step 1 tells
the agent to send an OTP email and wait for the user to reply with the code"*.

## Step 3 — Ask validation questions

Construct 3–5 short questions that test whether the documented behavior matches
the user's real intent. Each question should cover a distinct part of the spec.
Examples:
- "In section 4 the spec says the agent should retry 3 times. Do you actually
  want retries, or should it fail immediately?"
- "The spec mentions storing results in `/tmp/output.json`. Is that path
  fine, or should it be configurable?"
- "It looks like the agent is supposed to log at every step. Is that the right
  level of verbosity, or should logging be optional?"

Wait for the user's answers before continuing.

## Step 4 — Update the spec if answers differ

For each answer:
- **If the answer matches the spec**, move on.
- **If the answer differs**, modify the spec file so it reflects the user's
  stated intent.

Use `edit` to make targeted changes (preferred) or `write` to rewrite the file
if the structure needs large changes. Keep the user's original writing style,
phrasing, and formatting conventions unless they explicitly ask for a rewrite.

If you make changes, record them in a running list:
```markdown
### Changes made
- Section 2: changed retry count from 3 to 0
- Section 4: replaced hard-coded path with parameter `{output_path}`
- Section 5: added "logging is enabled by default but can be disabled with `quiet=True`"
```

## Step 5 — Final summary

Report back to the user with:
1. The general view again (brief).
2. Any changes you made (the running list above). If no changes were needed,
   say so explicitly.

## Important rules

- Always confirm with the user before changing the spec. This skill requires an
  interactive question phase — do not silently edit the file.
- Only modify files that the user has explicitly given you or confirmed you may
  change.
- Maintain the spec's structure and formatting as closely as possible when
  editing.
