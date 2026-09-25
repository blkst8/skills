# Batched product-feedback decision report

## Run configuration

- Assigned run directory: `skills/decision-workspace/iteration-1/batched-feedback/`
- Interpreter: `/home/blkst8/src/github/blkst8/scripts/laya-test/.venv/bin/python`
- Laya package: installed in the environment's site-packages
- `DECISION_LIBRARY=laya`
- `DECISION_MODEL` was explicitly unset for this run, so `Router.predict` omitted the model argument and the installed Router selected `english` automatically (`reason: "English Latin text"`).
- `DECISION_BASE_URL` and `DECISION_API_KEY` were not set; no credential or endpoint was embedded in the script.

## Steps and command

1. Defined one shared portable question schema with exactly one `choice`, one `score`, and one `noul` question.
2. Reused that same schema for all three feedback states in one application-level batch function.
3. The installed `Router.predict` API accepts one state per call, so the script makes three backend calls within the single batch program; this is the public API's state granularity, not three unrelated question formats.
4. Computed area counts, human-request count, average urgency, and urgency ranking in Python after all model answers were captured.
5. Ran:

   ```bash
   cd /home/blkst8/src/github/blkst8/scripts/laya-test
   env -u DECISION_MODEL DECISION_LIBRARY=laya ./.venv/bin/python \
     /home/blkst8/src/github/blkst8/skills/skills/decision-workspace/iteration-1/batched-feedback/outputs/batched_feedback.py
   ```

6. The script wrote the full raw result and Python aggregates to `outputs/observed_report.json`.

## Observed results

The run completed successfully and produced all three records.

| Item | Product area choice | Urgency score | Human-response noul | Router model |
|---|---|---:|---:|---|
| `checkout-001` | `checkout` (p=0.9809, confidence=0.9172) | 1.4975 | 0.0763 | `english` |
| `onboarding-002` | `onboarding` (p=0.9586, confidence=0.8439) | 0.8562 | 0.2421 | `english` |
| `mobile-003` | `mobile` (p=0.7843, confidence=0.4727) | 1.0934 | 0.1331 | `english` |

The urgency score distributions and complete raw answers are in `outputs/observed_report.json`. The observed model answers are preserved even where they differ from the apparent intent of the state; no answer was corrected or invented.

Python-computed aggregates from the captured answers:

- Product-area counts: `checkout=1`, `mobile=1`, `onboarding=1`
- Explicit-human-request count at noul threshold `>= 0.5`: `0`
- Average urgency score: `1.1490333333333334`
- Urgency ranking (high to low): `checkout-001` (1.4975), `mobile-003` (1.0934), `onboarding-002` (0.8562)

The installed package emitted a runtime warning that one checkpoint bucket has an out-of-range temperature and that affected confidence values are uncalibrated. This warning did not prevent execution and is not a test failure.

## Files

- `outputs/batched_feedback.py` — runnable source for the shared-schema batch
- `outputs/observed_report.json` — captured Laya answers, routing metadata, and Python aggregates
- `transcript.md` — this run record
