<div align="center">

# 🧰 blkst8/skills

### A personal arsenal of agent skills — small, opinionated, load-bearing.

[![Skills](https://img.shields.io/badge/skills-4-blueviolet)](#-the-skills)
[![Platform](https://img.shields.io/badge/agents-Claude%20%7C%20Codex%20%7C%20Cursor-success)](#-installation)
[![License](https://img.shields.io/badge/license-MIT-lightgrey)](#-license)
[![Format](https://img.shields.io/badge/format-SKILL.md-orange)](#-anatomy)

*Each skill is a single, composable `SKILL.md` you can drop into any agent harness that supports progressive-disclosure instructions.*

</div>

---

## ✨ What is this?

This repository is my **personal collection of agent skills** — self-contained `SKILL.md` files that teach AI agents a specific workflow, convention, or behavior. They are:

- **Bounded.** One skill = one job. No kitchen-sink prompts.
- **Composable.** Skills stack. A `feature-manager` roadmap can be reviewed by `spec-review`; `self-healing` runs underneath every task.
- **Harness-agnostic.** Any agent that loads `SKILL.md` can use them.
- **Installable everywhere.** A single shell script publishes them globally and to every git repo on your machine.

> Think of this repo as a toolbox. Pull only the `SKILL.md` you need; ignore the rest.

---

## 🎯 The Skills

| Skill | What it does | Always-on? |
|:------|:-------------|:----------:|
| [**`trump`**](skills/trump/SKILL.md) | A binary emphasis layer — CRITICAL concepts in ALL CAPS, with a live PRIORITY REGISTER that survives every step of reasoning. | ✅ |
| [**`self-healing`**](skills/self-healing/SKILL.md) | Six-phase improvement cycle (Observe → Reflect → Extract → Score → Generate → Register) that turns repeated patterns into new skills automatically. | ✅ |
| [**`spec-review`**](skills/spec-review/SKILL.md) | Interactive two-view review of any agent-spec `.md` file, followed by targeted questions and edits to match the user's real intent. | — |
| [**`feature-manager`**](skills/feature-manager/SKILL.md) | Decomposes a markdown feature spec into a structured roadmap of goals + tasks, written to a versioned `/tmp/` checklist. | — |

### The four skills, in one breath

- **`trump`** makes sure the agent never loses sight of what matters.
- **`self-healing`** makes sure the agent gets better every time it runs.
- **`spec-review`** makes sure the agent does what *you* actually meant.
- **`feature-manager`** makes sure big ideas become small, shippable tasks.

---

## 🏗️ Anatomy of a Skill

Every skill in this repo follows the same minimal shape:

```
skills/<name>/
├── SKILL.md            # the only required file
├── evals/              # (optional) trigger & behavior evals
└── references/         # (optional) deeper theory, tables, history
```

`SKILL.md` opens with a YAML frontmatter block that drives discovery:

```yaml
---
name: skill-name
description: >
  Plain-language trigger conditions. When should this skill fire?
  Which phrases, file types, and user intents should activate it?
---
```

A good skill is **specific about when to fire** and **rigorous about how to behave**. The four in this repo model that.

---

## ⚡ Installation

A single script installs every skill globally and seeds it into every git repo under `~/src/`.

```bash
git clone https://github.com/blkst8/skills.git ~/src/github/blkst8/skills
cd ~/src/github/blkst8/skills
./install-default-agents.sh
```

What it does:

1. Copies every `skills/*/` directory into `~/.agents/skills/`.
2. Walks `~/src/` and copies `~/.agents/skills/` into each repo's `.agents/`.
3. Installs a `post-checkout` git hook so freshly-cloned repos auto-populate `.agents/`.

> **Tip:** Override `TARGET_DIR` to seed a different root, or just symlink `~/.agents/skills/` to `skills/` for live updates.

### Manual / per-skill install

Want only one skill? Copy a single directory:

```bash
cp -r skills/feature-manager ~/.agents/skills/
```

Or symlink it from a checkout:

```bash
ln -s "$(pwd)/skills/trump" ~/.agents/skills/trump
```

---

## 🧩 Companion Skills (not in this repo)

These are skills I use daily that live elsewhere. See [`other-skills.md`](other-skills.md) for install snippets.

- [**`ast-grep`**](https://github.com/ast-grep/ast-grep) — structural code search for refactors and codemods.
- [**`caveman`**](https://github.com/JuliusBrussee/caveman) — token-saving compressed communication mode.
- [**`goalbuddy`**](https://github.com/tolibear/goalbuddy) — structured goal intake for long-running Claude Code work.

---

## 📚 Specs & Theory

Long-form thinking that backs the skills lives in [`specs/`](specs/):

| File | Topic |
|:-----|:------|
| [AI Spec – Self-Improving Agent](specs/AI%20Spec%20%E2%80%93%20Self-Improving%20Agent%20(Theory%20&%20Mechanism).md) | Full theory of the self-healing six-phase cycle. |
| [CAPS Emphasis Agent Skill — Spec](specs/CAPS%20Emphasis%20Agent%20Skill%20%E2%80%94%20Spec.md) | Design notes for the binary-emphasis pattern. |
| [Feature-manager](specs/Feature-manager.md) | The decomposition heuristics behind `feature-manager`. |
| [GO_CODEBASE_STYLE](specs/GO_CODEBASE_STYLE.md) / [GUIDE](specs/GO_CODEBASE_STYLE_GUIDE.md) | Go conventions for any agent touching Go codebases. |
| [standard-contractions-and-informal](specs/standard-contractions-and-informal.md) | Voice rules for terse, human-feeling output. |

---

## 🛠️ Authoring Your Own

The fastest path:

1. Copy any `skills/<name>/` directory.
2. Rewrite the frontmatter `description` to be **specific** about trigger conditions.
3. Write the body in three parts: *When to use / When NOT to use / Algorithm*.
4. Add at least one eval under `evals/` to lock the trigger phrase.
5. Open a PR.

The skills in this repo are the same shape you'd ship.

---

## 🤝 Contributing

Issues and PRs welcome — but read [`spec-review`](skills/spec-review/SKILL.md) first. It will save you a round trip.

---

## 📄 License

[MIT](LICENSE) — do what you want, just keep the attribution.

---

<div align="center">

<sub>Built and maintained by <a href="https://github.com/blkst8">@blkst8</a>. If a skill here saved you a debugging session, star the repo ⭐.</sub>

</div>
