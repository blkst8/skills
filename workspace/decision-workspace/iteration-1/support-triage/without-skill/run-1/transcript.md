# Support-ticket triage with Laya — run transcript

Run root: `skills/decision-workspace/iteration-1/support-triage/`
Eval: `skills/decision/evals/evals.json` → eval 1, "support-ticket-triage".

## 1. Environment discovered

Interpreter and package live in the Laya test environment, not on the system
Python (`/usr/bin/python3` has no `laya` module):

```bash
ls /home/blkst8/src/github/blkst8/scripts/laya-test/
# .venv  laya_test.py  persian_score_loadtest.py  loadtest.json  report-*.json

/home/blkst8/src/github/blkst8/scripts/laya-test/.venv/bin/python -c \
  "import laya; print(laya.__version__, laya.__file__)"
# 0.3.5 /home/blkst8/src/github/blkst8/scripts/laya-test/.venv/lib/python3.12/site-packages/laya/__init__.py
```

Relevant API surface read from the installed source before writing any code:

- `laya/router.py:315` — `Router.predict(state, questions, model=None, task=None, lang=None)`
  routes first, then answers **every question in one forward pass**, and returns
  the agent payload plus a `routing` key.
- `laya/router.py:257` — `Router.route` precedence is
  `model` > `task` > detected workflow (opt-in) > `lang` > detected script/language
  > default. With `model=None` and `auto_task_detection=False` (the default),
  selection is automatic and language-driven.
- `laya/agent.py:337-362` — answer shapes: `choice` → `choice`/`probabilities`/
  `confidence`; `score` → `score` as the probability-weighted mean of
  **zero-based** level indices plus `probabilities` keyed by level index and
  `legend`; `noul` → `noul` = P(true) and a `confidence` that is just
  `max(p, 1-p)`, i.e. not an independent confidence signal.
- `laya/common.py:219` — temperature buckets are keyed by type *and* option
  count (`choice:3-5`, `choice:11+`, …).

Checkpoints were already in the local HF cache
(`~/.cache/huggingface/hub/models--convaiinnovations--laya`, ~snapshots
`encoder/`, `multilingual/`, `typed-decisions/`), so the run needs no download.

## 2. Program

`outputs/ticket_triage.py` (also executable). Structure:

| Piece | What it does |
| --- | --- |
| `load_config` | Reads `DECISION_LIBRARY`, `DECISION_LIBRARY_URL`, `DECISION_MODEL`, `DECISION_TIMEOUT`, `DECISION_MAX_RETRIES`, `DECISION_BASE_URL`, `DECISION_API_KEY`. Raises `ConfigError` on an unsupported library, an unparseable timeout, a negative retry count, or a `DECISION_BASE_URL` with no remote transport. Reports only *whether* `DECISION_API_KEY` is set, never its value. |
| `build_questions` | One `choice` (`category`), one `score` (`disruption`), one `noul` (`refund_requested`), all against the same `ticket` state. |
| `build_state` | Sends only `ticket.id`, `ticket.title`, `ticket.text`. |
| `predict_with_deadline` | One request for all three questions, run on a daemon thread bounded by `DECISION_TIMEOUT`. |
| `run_decisions` | Imports `laya` lazily, builds `Router()`, retries only `TimeoutError`/`ConnectionError`/`OSError`, never a rejected request. `model` is passed only when `DECISION_MODEL` is set, so an unset value leaves selection to the Router. |
| `normalised_disruption` | The score is a zero-based expected level, so it is divided by `len(criteria) - 1` **in Python**. |
| `decide` | All policy: floors, branches, queue names, human-review fallback. No model calls. |

Thresholds live in the file as named constants next to `decide`:
`CATEGORY_CONFIDENCE_FLOOR = 0.60`, `DISRUPTION_ESCALATION = 0.75`,
`DISRUPTION_CONFIDENCE_FLOOR = 0.50`, `REFUND_NOUL_THRESHOLD = 0.70`.

Question design follows the skill: the choice has `what` / `not_for` /
`examples` per option plus an `other`; the three score levels are distinct
situations (continues / slowed-with-workaround / cannot-proceed) rather than
degrees; the `noul` has explicit `true` and `false` sides, with "reports a
charge and threatens to leave" placed on the `false` side. Instructions use the
object form and name the state path in backticks.

The `noul` is gated on its probability, not on the returned `confidence` field —
that field is `max(p, 1-p)` and carries no extra information.

## 3. Run

```bash
cd skills/decision-workspace/iteration-1/support-triage/outputs
DECISION_LIBRARY=laya \
DECISION_TIMEOUT=180s \
HF_HUB_OFFLINE=1 \
/home/blkst8/src/github/blkst8/scripts/laya-test/.venv/bin/python \
  ticket_triage.py --ticket-file duplicate_charge_ticket.json --out result.json
```

`HF_HUB_OFFLINE=1` only forces the already-populated local cache; it is not a
model endpoint. `DECISION_MODEL` was deliberately left unset so the Router
selected the checkpoint. Exit code 0. Console output saved verbatim to
`outputs/run.log`; full report JSON to `outputs/result.json`.

### Observed output

```
ticket        : SUP-4411 — Charged twice for invoice #4411
library       : laya (model selection: router-auto)
router model  : english — English Latin text
category      : billing (confidence 0.1397, probs {'billing': 0.5383, 'technical': 0.1795, 'account_access': 0.1564, 'other': 0.1258})
disruption    : score 1.2362 (confidence 0.206, probs {'0': 0.0806, '1': 0.6026, '2': 0.3168})
refund asked  : noul 0.8421 (confidence 0.8421)
route         : human_review -> support.triage
reason        : category confidence 0.14 < floor 0.60; category left to a person
human review  : True
elapsed       : 11.411s in 1 attempt(s)
```

`Router` routing record in `result.json`:
`{"model": "english", "repo": "convaiinnovations/laya", "reason": "English Latin text",
"detection": {"script": "latin", "language": "en", "is_english": true, ...}}`.
Token usage reported by the model: `input_tokens: 766`, `output_tokens: 0`.

The command was run twice in this environment. Both executions returned
identical `answers` and `routing` (the checkpoint is deterministic here); only
the wall-clock elapsed time differs. `outputs/run.log` and `outputs/result.json`
are both from the final execution.

### Observed routing decision

```json
{
  "route": "human_review",
  "queue": "support.triage",
  "reason": "category confidence 0.14 < floor 0.60; category left to a person",
  "automatic_action": false,
  "human_review": true,
  "signals": {
    "category": "billing",
    "category_confidence": 0.1397,
    "category_confidence_floor": 0.6,
    "disruption_score": 1.2362,
    "disruption_normalised": 0.6181,
    "disruption_confidence": 0.206,
    "refund_probability": 0.8421,
    "refund_threshold": 0.7
  }
}
```

## 4. What this run actually shows

- The ticket's **category** answer was the right label (`billing`, top
  probability 0.5383) but the distribution was nearly flat, so the published
  `confidence` was 0.1397. The Python gate caught that and sent the ticket to a
  person — the low-confidence fallback did its job on real model output, not as
  an unreachable branch.
- The `noul` was sharp: P(explicit refund request) = 0.8421, comfortably above
  the 0.70 policy threshold. Had the category gate passed, this ticket would have
  taken `billing.refund_queue`.
- The `score` was a mixture, not a level: expected level 1.2362 over three
  levels with the mass at level 1 (0.6026) and level 2 (0.3168) — the "degraded"
  and "blocked" situations both partly present. Read as a bare level it would
  have been a coin flip; the policy reads the distribution.
- `confidence` for the `noul` (0.8421) is exactly `max(p, 1-p)`, which is why the
  policy thresholds the probability instead.
- Branches not taken by this ticket (`engineering.escalate`,
  `engineering.backlog`, `billing.general`, `support.accounts`, and the `other`
  fallback) are code paths in `decide` that this single-ticket run did not
  reach. The ticket was chosen by the eval, not to cover the branch table.
- The `RuntimeWarning` in `run.log` line 1 is emitted at `Agent` construction
  and concerns the checkpoint's shipped `choice:11+` temperature (0.1006),
  clamped to 0.5. This run's choice has 4 options, so it used the
  `choice:3-5` bucket and that clamped value was not applied to these answers.
  The warning is not caused by the program.

## 5. Files

| File | Contents |
| --- | --- |
| `outputs/ticket_triage.py` | The program. |
| `outputs/duplicate_charge_ticket.json` | Input ticket `SUP-4411`. |
| `outputs/result.json` | Full observed report: config, request (state + questions), raw `response.answers` with probabilities and confidence, `routing`, and `routing_decision`. |
| `outputs/run.log` | Verbatim console output and the library warning. |
| `transcript.md` | This file. |

No credentials, tokens, or inference endpoints appear in the program, the input,
or the saved report. The only URL in the source is the documented default for
`DECISION_LIBRARY_URL` used for provenance in the error message when the `laya`
package is missing; it is overridable by the environment variable.
