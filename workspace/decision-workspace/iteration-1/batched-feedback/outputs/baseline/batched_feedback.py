#!/usr/bin/env python3
"""Run three feedback states through one shared Laya question schema."""
from __future__ import annotations

import json
import os
import platform
import sys
from collections import Counter
from pathlib import Path
from typing import Any

import laya

QUESTIONS: dict[str, dict[str, Any]] = {
    "product_area": {
        "type": "choice",
        "instructions": "Which product area best describes the main request in `feedback.text`?",
        "criteria": {
            "checkout": "Payments, orders, carts, or checkout flow",
            "onboarding": "Account setup, first-run experience, or learning the product",
            "mobile": "Mobile app behavior, mobile notifications, or mobile performance",
            "other": "A product area not covered by the options above",
        },
    },
    "urgency": {
        "type": "score",
        "instructions": "How urgent is the situation described in `feedback.text`?",
        "criteria": [
            "No time pressure; the request can wait for a normal release cycle",
            "A near-term deadline or a blocked workaround requires prompt attention",
            "A critical deadline or an active business blocker requires immediate attention",
        ],
    },
    "human_response_requested": {
        "type": "noul",
        "instructions": "Does `feedback.text` explicitly request a response from a human or a person?",
    },
}

STATES = [
    {"id": "checkout-001", "text": "Checkout rejects valid cards during the Black Friday sale. We cannot take orders today."},
    {"id": "onboarding-002", "text": "The setup wizard is confusing for new teams. A short example would help us continue next month."},
    {"id": "mobile-003", "text": "The iOS app takes 20 seconds to load the daily report. Please have someone call me before our 9 a.m. standup."},
]


def run_batch() -> dict[str, Any]:
    library = os.environ.get("DECISION_LIBRARY", "laya")
    if library != "laya":
        raise RuntimeError(f"DECISION_LIBRARY must be 'laya', got {library!r}")
    router = laya.Router()
    records = []
    for item in STATES:
        # Router.predict is the installed public API. With no model argument,
        # Router automatically selects an appropriate checkpoint per state.
        observed = router.predict({"feedback": {"text": item["text"]}}, QUESTIONS)
        records.append({"id": item["id"], "text": item["text"], "observed": observed})

    area_counts = Counter(r["observed"]["answers"]["product_area"]["choice"] for r in records)
    human_probabilities = {
        r["id"]: float(r["observed"]["answers"]["human_response_requested"]["noul"])
        for r in records
    }
    scores = {r["id"]: float(r["observed"]["answers"]["urgency"]["score"]) for r in records}
    ranking = sorted(scores, key=lambda item_id: (-scores[item_id], item_id))
    return {
        "run": {
            "application_batch": True,
            "state_count": len(records),
            "shared_schema": True,
            "shared_question_ids": list(QUESTIONS),
            "backend_api": "Router.predict (all questions in one forward pass per state)",
            "router_model_selection": "automatic" if not os.environ.get("DECISION_MODEL") else "DECISION_MODEL",
            "decision_library": library,
            "model_env": os.environ.get("DECISION_MODEL"),
            "credentials_or_endpoints_in_source": False,
            "python": platform.python_version(),
            "interpreter": sys.executable,
            "laya_version": getattr(laya, "__version__", "unknown"),
        },
        "questions": QUESTIONS,
        "states": records,
        "python_aggregates": {
            "product_area_counts": dict(sorted(area_counts.items())),
            "urgency_scores": scores,
            "urgency_ranking_high_to_low": ranking,
            "human_response_probabilities": human_probabilities,
            "human_response_count_at_noul_0_5": sum(p >= 0.5 for p in human_probabilities.values()),
            "calculation_note": "All counts, scores, thresholds, and rankings are computed in Python from observed model answers.",
        },
    }


def main() -> int:
    report = run_batch()
    output = Path(__file__).with_name("observed_report.json")
    output.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(report, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
