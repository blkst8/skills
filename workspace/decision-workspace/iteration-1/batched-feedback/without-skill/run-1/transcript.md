# Baseline batched-feedback run

## Environment

- Installed interpreter: `/home/blkst8/src/github/blkst8/scripts/laya-test/.venv/bin/python`
- Python: 3.12.3
- Installed Laya: 0.3.5
- Configuration: `DECISION_LIBRARY=laya`
- `DECISION_MODEL` was unset, so `laya.Router()` selected checkpoints automatically.
- No API key, endpoint, or credential is present in the source.

## Command

```bash
cd /home/blkst8/src/github/blkst8/scripts/laya-test
DECISION_LIBRARY=laya .venv/bin/python \
  /home/blkst8/src/github/blkst8/skills/skills/decision-workspace/iteration-1/batched-feedback/outputs/baseline/batched_feedback.py
```

The command completed successfully (exit code 0). The complete captured stdout is in `run_stdout.txt`; stderr, including Laya's temperature warning, is in `run_stderr.txt`.

## What ran

`outputs/baseline/batched_feedback.py` defines one immutable question schema with:

- `product_area`: a `choice` question.
- `urgency`: a `score` question with three distinct situation-based levels.
- `human_response_requested`: a `noul` question for an explicit human response request.

It processes three states in one application-level batch loop. Laya's public `Router.predict` answers all three questions for each state in one forward pass. The same `QUESTIONS` object is used for every state. Python then computes product-area counts, urgency scores/ranking, and the human-response count using the `noul >= 0.5` threshold.

## Observed result

`observed_report.json` contains all three raw observed answers, probabilities, confidence, routing metadata, and Python aggregates. Summary from the observed report:

- Product-area counts: `checkout=1`, `mobile=1`, `onboarding=1`.
- Urgency ranking (high to low): `checkout-001`, `mobile-003`, `onboarding-002`.
- Human-response probabilities: checkout `0.0763`, onboarding `0.2421`, mobile `0.1331`; count at `noul >= 0.5`: `0`.
- Each state routed automatically to the English checkpoint (`routing.model=english`).
