#!/usr/bin/env python3
"""Triage one support ticket with Laya's portable decision schema."""
from __future__ import annotations

import argparse
import json
import os
from typing import Any


QUESTIONS: dict[str, dict[str, Any]] = {
    "category": {
        "type": "choice",
        "instructions": "Which category best describes `ticket.title` and `ticket.text`?",
        "criteria": {
            "billing": "Duplicate charges, payments, invoices, subscriptions, or refunds",
            "technical": "Broken features, errors, outages, or access problems",
            "other": "Anything not covered by billing or technical support",
        },
    },
    "disruption": {
        "type": "score",
        "instructions": "How disruptive is the issue described in `ticket.text`?",
        "criteria": [
            "No meaningful disruption; the user can continue normally",
            "A degraded feature; the user has a workaround",
            "A blocking issue; the user has no workaround",
        ],
    },
    "refund_requested": {
        "type": "noul",
        "instructions": "Does `ticket.text` explicitly request a refund or credit?",
    },
}


def build_router():
    """Build the configured local Laya adapter without endpoint credentials."""
    library = os.getenv("DECISION_LIBRARY", "laya")
    if library != "laya":
        raise ValueError(f"Unsupported DECISION_LIBRARY: {library}")
    from laya import Router

    return Router()


def route(answers: dict[str, Any]) -> dict[str, Any]:
    """Apply deterministic support policy and a conservative confidence gate."""
    category = answers["category"]
    disruption = answers["disruption"]
    refund_requested = answers["refund_requested"]
    category_confidence = float(category.get("confidence", 0.0))
    refund_probability = float(refund_requested.get("noul", 0.0))

    if category_confidence < 0.60:
        route_name = "human_review"
        reason = "category confidence is below 0.60"
    elif category.get("choice") == "billing" and refund_probability >= 0.70:
        route_name = "billing_refund"
        reason = "billing category and explicit refund probability is at least 0.70"
    else:
        route_name = "standard_support_queue"
        reason = "confidence clears the gate and the billing refund policy does not match"

    return {
        "route": route_name,
        "reason": reason,
        "thresholds": {
            "minimum_category_confidence": 0.60,
            "minimum_refund_probability": 0.70,
        },
        "signals": {
            "category": category.get("choice"),
            "category_confidence": category_confidence,
            "disruption_score": disruption.get("score"),
            "refund_probability": refund_probability,
        },
    }


def triage_ticket(ticket: dict[str, str]) -> dict[str, Any]:
    """Run one three-question decision request and apply Python routing."""
    if not isinstance(ticket, dict):
        raise TypeError("ticket must be a JSON object with title and text")
    for field in ("title", "text"):
        if not isinstance(ticket.get(field), str) or not ticket[field].strip():
            raise ValueError(f"ticket.{field} must be a non-empty string")

    router = build_router()
    state = {"ticket": ticket}
    # Omit model when DECISION_MODEL is unset so Laya Router selects automatically.
    model = os.getenv("DECISION_MODEL")
    kwargs = {"model": model} if model else {}
    raw_result = router.predict(state, QUESTIONS, **kwargs)
    answers = raw_result["answers"]
    return {
        "ticket": ticket,
        "decision_request": {"state": state, "questions": QUESTIONS},
        "raw_result": raw_result,
        "answers": answers,
        "routing_decision": route(answers),
        "configuration": {
            "DECISION_LIBRARY": os.getenv("DECISION_LIBRARY", "laya"),
            "DECISION_MODEL": model or "automatic Laya Router selection",
        },
    }


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--ticket", required=True, help="Ticket JSON object")
    args = parser.parse_args()
    result = triage_ticket(json.loads(args.ticket))
    print(json.dumps(result, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
