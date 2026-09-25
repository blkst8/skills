# Skill Benchmark: decision

**Model**: <model-name>
**Date**: 2026-09-25T13:58:12Z
**Evals**: 1, 2, 3 (3 runs each per configuration)

## Summary

| Metric | With-Skill | Without-Skill | Delta |
|--------|------------|---------------|-------|
| Pass Rate | 100% ± 0% | 100% ± 0% | +0.00 |
| Time | 168.0s ± 29.8s | 151.5s ± 125.8s | +16.5s |
| Tokens | 4012 ± 6948 | 0 ± 0 | +4012 |

## Notes

- All six runs passed all five expectations, so this benchmark measures contract completion rather than skill differentiation; the prompts are too explicit for the baseline to fail.
- With-skill and without-skill runs both exercised the installed Laya package and produced observed typed answers, so the comparison does not isolate the value of the skill's prose.
- The with-skill timing mean was 168.0 seconds versus 151.5 seconds for baseline, but the values mix model startup, downloads/cache effects, and agent wall time; they are not a reliable performance comparison.
- The moderation evaluation captured Laya's uncalibrated-confidence warning; the generated code still routed adversarial input to human review, but model confidence should not be treated as calibrated without evaluation.
- The initial concurrent evaluation runs collided in shared output directories. Artifacts were separated into with-skill/run-1 and without-skill/run-1 before aggregation; future runs should provision distinct paths before spawning agents.