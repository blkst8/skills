#!/usr/bin/env python3
"""Run one application-level feedback batch through the installed Laya Router.

The backend accepts one state per predict call.  This program makes the
application-level batch explicit: it runs all three states in one process with
the same immutable question schema, then performs every aggregate in Python.
"""
from __future__ import annotations

import json
import os
import platform
import sys
from collections import Counter
from pathlib import Path
from typing import Any, Dict, List

import laya


# One shared schema is used for every item.  The score criteria describe
# situations rather than asking the model to assign a numeric label.
QUESTIONS: Dict[str, Dict[str, Any]] = {
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

STATES: List[Dict[str, str]] = [
    {
        "id": "checkout-001",
        "text": "Checkout rejects valid cards during the Black Friday sale. We cannot take orders today.",
    },
    {
        "id": "onboarding-002",
        "text": "The setup wizard is confusing for new teams. A short example would help us continue next month.",
    },
    {
        "id": "mobile-003",
        "text": "The iOS app takes 20 seconds to load the daily report. Please have someone call me before our 9 a.m. standup.",
    },
]


def run_shared_schema_batch(router: laya.Router) -> List[Dict[str, Any]]:
    """Process one application batch with a shared question schema.

    Laya's public Router API evaluates a single state and all questions in that
    state in one forward pass.  The loop is therefore the portable application
    batch adapter; it does not create a different schema per item.
    """
    records: List[Dict[str, Any]] = []
    for state in STATES:
        kwargs: Dict[str, Any] = {}
        # Leave model unset so Laya Router performs automatic selection.  A
        # deployment may still opt into a checkpoint through DECISION_MODEL.
        configured_model = os.environ.get("DECISION_MODEL")
        if configured_model:
            kwargs["model"] = configured_model
        result = router.predict({"feedback": {"text": state["text"]}}, QUESTIONS, **kwargs)
        records.append({"id": state["id"], "text": state["text"], "raw_result": result})
    return records


def compute_python_aggregates(records: List[Dict[str, Any]]) -> Dict[str, Any]:
    """Compute counts, average, and ranking deterministically in Python."""
    answers = [record["raw_result"]["answers"] for record in records]
    area_counts = Counter(answer["product_area"]["choice"] for answer in answers)
    urgency_ranking = sorted(
        (
            {
                "id": record["id"],
                "score": record["raw_result"]["answers"]["urgency"]["score"],
                "legend": record["raw_result"]["answers"]["urgency"]["legend"],
            }
            for record in records
        ),
        key=lambda item: (-item["score"], item["id"]),
    )
    scores = [float(answer["urgency"]["score"]) for answer in answers]
    human_probabilities = [float(answer["human_response_requested"]["noul"]) for answer in answers]
    return {
        "area_counts": dict(sorted(area_counts.items())),
        "human_response_count_at_0_5": sum(probability >= 0.5 for probability in human_probabilities),
        "human_response_item_ids_at_0_5": [
            record["id"]
            for record in records
            if float(record["raw_result"]["answers"]["human_response_requested"]["noul"]) >= 0.5
        ],
        "urgency_average_score": sum(scores) / len(scores),
        "urgency_ranking_high_to_low": urgency_ranking,
    }


def main() -> int:
    library = os.environ.get("DECISION_LIBRARY", "laya")
    if library != "laya":
        raise RuntimeError(f"DECISION_LIBRARY must be 'laya', got {library!r}")

    router = laya.Router()
    records = run_shared_schema_batch(router)
    report = {
        "run": {
            "application_batch": True,
            "state_count": len(records),
            "shared_question_ids": list(QUESTIONS),
            "shared_schema": True,
            "backend_api": "Router.predict",
            "router_model_selection": "automatic" if not os.environ.get("DECISION_MODEL") else "DECISION_MODEL",
            "decision_library": library,
            "model_env": os.environ.get("DECISION_MODEL"),
            "base_url_configured": bool(os.environ.get("DECISION_BASE_URL")),
            "api_key_configured": bool(os.environ.get("DECISION_API_KEY")),
            "python": platform.python_version(),
            "interpreter": sys.executable,
        },
        "questions": QUESTIONS,
        "states": records,
        "python_aggregates": compute_python_aggregates(records),
    }
    output_path = Path(__file__).with_name("observed_report.json")
    output_path.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(report, indent=2, sort_keys=True))
    print(f"\nSaved observed report to {output_path}")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"ERROR: {type(exc).__name__}: {exc}", file=sys.stderr)
        raise
