# Support-ticket triage evaluation run (with decision skill)

## Scope

Implemented and ran the eval-1 support-ticket triage example for a duplicate-charge ticket. The program uses the decision skill's portable `state`/`questions`/`answers` shape, imports the installed local `laya` package, and leaves model selection to `laya.Router` when `DECISION_MODEL` is unset.

## Files

- `outputs/support_triage.py` — runnable Python program.
- `outputs/observed_result.json` — actual JSON emitted by the program, including the request, raw Laya result, typed answers, probabilities/confidence, Router selection, and Python routing decision.
- `outputs/run.stderr` — stderr captured from the evaluation run.

## Command

From `skills/decision-workspace/iteration-1/support-triage`:

```bash
DECISION_LIBRARY=laya /home/blkst8/src/github/blkst8/scripts/laya-test/.venv/bin/python outputs/support_triage.py --ticket '{"title":"Duplicate charge on invoice #4411","text":"We were charged twice for March. Please refund the duplicate today."}' > outputs/observed_result.json 2> outputs/run.stderr
```

No `DECISION_MODEL`, API key, base URL, or other endpoint was supplied. The program read `DECISION_LIBRARY=laya` and omitted the model argument, causing automatic Router selection.

## Observed result

The process exited successfully (exit code 0) in 40.91 seconds. Laya selected the `english` Router model (`raw_result.routing.model`), backed by `laya-rl-agent`, with reason `English Latin text`.

For the single request, the observed answers were:

- Category: `billing`; confidence `0.8264`; probabilities `{billing: 0.9613, other: 0.0204, technical: 0.0183}`.
- Disruption: score `1.0414`; confidence `0.2155`; probabilities `{0: 0.453, 1: 0.0526, 2: 0.4944}`.
- Explicit refund request: `noul: 0.8003`; confidence `0.8003`.

The Python policy uses a `0.60` minimum category confidence gate and sends the ticket to `billing_refund` when the category is `billing` and refund probability is at least `0.70`; otherwise it uses the standard queue. The observed route was `billing_refund` because category confidence was `0.8264` and refund probability was `0.8003`.

The captured stderr contains an unauthenticated Hugging Face download warning and a Laya checkpoint-temperature calibration warning. No credentials or inference endpoints are embedded in the source.
