#!/usr/bin/env python3
"""Support-ticket triage with the Laya decision model.

The model answers three bounded questions about one ticket. All policy --
thresholds, branches, queue names, and the human-review fallback -- lives in
this Python file. Nothing here decides anything the model did not answer.

Configuration comes from the environment only (the portable decision surface):
  DECISION_LIBRARY        adapter to load; this program implements "laya" only
  DECISION_LIBRARY_URL    installation/library source (recorded, never fetched here)
  DECISION_MODEL          checkpoint name; unset => Laya Router auto-selects
  DECISION_TIMEOUT        finite request deadline, e.g. "30s"
  DECISION_MAX_RETRIES    retries for transient transport failures only
  DECISION_BASE_URL       remote inference base URL (unsupported by this adapter)
  DECISION_API_KEY        bearer token for an authenticated remote service

No credential, endpoint, or model name is hard-coded. Secrets are never echoed:
the report records only whether a variable is set.

Usage:
  python ticket_triage.py --ticket-file duplicate_charge_ticket.json \
      --out result.json
"""
from __future__ import annotations

import argparse
import json
import os
import re
import sys
import threading
import time
from typing import Any, Dict, List, Optional

# The only adapter implemented here. Its source URL is the upstream default of
# the DECISION_LIBRARY_URL variable, used only for provenance in the report.
SUPPORTED_LIBRARY = "laya"
LIBRARY_URL_DEFAULT = "https://github.com/NandhaKishorM/laya"

# --------------------------------------------------------------------------
# Policy. These numbers are application policy, not model configuration.
# They are deliberately conservative and belong to a human who can recalibrate
# them against labelled tickets.
# --------------------------------------------------------------------------
CATEGORY_CONFIDENCE_FLOOR = 0.60   # below this, a person decides the category
DISRUPTION_ESCALATION = 0.75       # normalised score above this is a page
DISRUPTION_CONFIDENCE_FLOOR = 0.50  # ...and only if the score is peaked
REFUND_NOUL_THRESHOLD = 0.70       # P(explicit refund request) above this

# Transient = the request never got a verdict. A rejected request or an
# ambiguous decision is not retried, per the decision skill's guidance.
_TRANSIENT = (TimeoutError, ConnectionError, OSError)


class ConfigError(Exception):
    """Deployment configuration is missing, unusable, or unsupported."""


# --------------------------------------------------------------------------
# configuration
# --------------------------------------------------------------------------
def parse_duration(raw: str) -> float:
    """Parse '30s' / '1m' / '500ms' / '30' into seconds. Must be finite."""
    m = re.fullmatch(r"\s*(\d+(?:\.\d+)?)\s*(ms|s|m|h)?\s*", raw)
    if not m:
        raise ConfigError(f"DECISION_TIMEOUT is not a duration: {raw!r}")
    value, unit = float(m.group(1)), (m.group(2) or "s")
    seconds = value / 1000.0 if unit == "ms" else value * {"s": 1, "m": 60, "h": 3600}[unit]
    if seconds <= 0:
        raise ConfigError(f"DECISION_TIMEOUT must be positive, got {raw!r}")
    return seconds


def load_config() -> Dict[str, Any]:
    """Read the portable configuration surface. Never returns secret values."""
    library = os.getenv("DECISION_LIBRARY", SUPPORTED_LIBRARY).strip().lower()
    if library != SUPPORTED_LIBRARY:
        raise ConfigError(
            f"DECISION_LIBRARY={library!r} is not implemented by this program; "
            f"the only supported adapter is {SUPPORTED_LIBRARY!r}. "
            "Refusing to silently fall back to a different library."
        )

    raw_retries = os.getenv("DECISION_MAX_RETRIES", "0").strip()
    if not raw_retries.isdigit():
        raise ConfigError(f"DECISION_MAX_RETRIES must be a non-negative integer, got {raw_retries!r}")
    max_retries = int(raw_retries)

    timeout = parse_duration(os.getenv("DECISION_TIMEOUT", "30s"))

    base_url = os.getenv("DECISION_BASE_URL", "").strip()
    if base_url:
        raise ConfigError(
            "DECISION_BASE_URL is set but this adapter runs the library locally and "
            "has no remote transport. Implement the remote transport (sending the same "
            "state/questions schema with DECISION_API_KEY) or unset DECISION_BASE_URL; "
            "do not let a remote deployment silently answer from a local checkpoint."
        )

    return {
        "library": library,
        "library_url": os.getenv("DECISION_LIBRARY_URL", LIBRARY_URL_DEFAULT),
        "model": os.getenv("DECISION_MODEL", "").strip() or None,  # None => Router decides
        "model_selection": "router-auto" if not os.getenv("DECISION_MODEL", "").strip() else "explicit",
        "timeout_seconds": timeout,
        "max_retries": max_retries,
        "api_key_present": bool(os.getenv("DECISION_API_KEY", "").strip()),
        "remote_endpoint_configured": bool(base_url),
    }


# --------------------------------------------------------------------------
# the decision request
# --------------------------------------------------------------------------
def build_questions() -> Dict[str, Any]:
    """One atomic question per judgement, each answerable in about a second."""
    return {
        "category": {
            "type": "choice",
            "instructions": {
                "question": "Which single category best describes the main request in `ticket.text`?",
                "inspect": "ticket.text",
                "focus": "Classify the request the customer is making, not the channel or their tone.",
            },
            "criteria": {
                "billing": {
                    "what": "Charges, invoices, refunds, subscriptions, or plan changes",
                    "not_for": "Product errors, sign-in problems, or requests for a feature",
                    "examples": ["We were charged twice this month", "Please refund invoice #4411"],
                },
                "technical": {
                    "what": "Something is broken, erroring, or performing badly",
                    "not_for": "Money already charged, or questions about how to use a feature",
                    "examples": ["The export job fails every night", "The dashboard returns a 500"],
                },
                "account_access": {
                    "what": "Signing in, passwords, permissions, or seats and members",
                    "not_for": "Anything about an amount charged",
                    "examples": ["I cannot log in after the reset", "Remove a teammate from our org"],
                },
                "other": {
                    "what": "Any request that is not charges, product faults, or access",
                    "not_for": "Do not use this merely because the request is unclear",
                    "examples": ["Who is your account manager?", "Can you move us to the EU region?"],
                },
            },
        },
        "disruption": {
            "type": "score",
            "instructions": {
                "question": "How much does the issue described in `ticket.text` disrupt the customer's use of the product?",
                "inspect": "ticket.text",
                "focus": "Judge the consequence for the customer, not how strongly they word it.",
            },
            "criteria": [
                {
                    "summary": "Work continues; the customer is only asking a question or a courtesy",
                    "signals": ["No failure is reported", "No deadline or consequence is stated"],
                },
                {
                    "summary": "A task is slowed or repeated, but a workaround lets the customer proceed",
                    "signals": ["A feature is degraded", "The customer describes a manual alternative"],
                },
                {
                    "summary": "The customer cannot complete the task at all, and says so",
                    "signals": ["Work is blocked", "The customer names a deadline, loss, or cancellation"],
                },
            ],
        },
        "refund_requested": {
            "type": "noul",
            "instructions": {
                "question": "Does `ticket.text` explicitly ask for money back?",
                "inspect": "ticket.text",
                "focus": "Only an explicit request to refund, credit, or return a charge counts.",
            },
            "criteria": {
                "true": {
                    "description": "The customer asks for a refund, credit, or reversal of a charge",
                    "examples": ["Please refund the duplicate charge", "Credit our account for March"],
                },
                "false": {
                    "description": "The customer reports a charge, questions it, or threatens to leave, but never asks for money back",
                    "examples": ["We were billed twice", "Why is this invoice higher than last month?"],
                },
            },
        },
    }


def build_state(ticket: Dict[str, str]) -> Dict[str, Any]:
    """Smallest state that answers all three questions. Nothing extra is sent."""
    return {
        "ticket": {
            "id": ticket["id"],
            "title": ticket["title"],
            "text": ticket["text"],
        }
    }


# --------------------------------------------------------------------------
# model call
# --------------------------------------------------------------------------
def predict_with_deadline(router: Any, state: Any, questions: Any, model: Optional[str], timeout: float) -> Dict[str, Any]:
    """One request, all questions, bounded by a finite deadline.

    The worker runs in a daemon thread so an overrunning forward pass cannot
    keep the process alive; the deadline bounds how long this program waits,
    not how long the underlying compute may already be running.
    """
    box: Dict[str, Any] = {}

    def work() -> None:
        kwargs = {"model": model} if model else {}
        try:
            box["result"] = router.predict(state, questions, **kwargs)
        except BaseException as exc:  # re-raised on the caller's thread
            box["error"] = exc

    worker = threading.Thread(target=work, name="laya-predict", daemon=True)
    worker.start()
    worker.join(timeout)
    if worker.is_alive():
        raise TimeoutError(f"decision request exceeded DECISION_TIMEOUT ({timeout}s)")
    if "error" in box:
        raise box["error"]
    return box["result"]


def run_decisions(config: Dict[str, Any], state: Any, questions: Any) -> tuple:
    """Import the configured library, then predict, retrying only transient failures."""
    if config["library"] != SUPPORTED_LIBRARY:
        raise ConfigError(f"no adapter for library {config['library']!r}")
    try:
        from laya import Router
    except ImportError as exc:  # pragma: no cover - environment dependent
        raise ConfigError(
            "the 'laya' package is not importable in this interpreter. "
            f"Install it from {config['library_url']!r} into the environment you run this with."
        ) from exc

    router = Router()
    started = time.time()
    attempts = 0
    last_error: Optional[BaseException] = None
    for attempt in range(config["max_retries"] + 1):
        attempts = attempt + 1
        try:
            result = predict_with_deadline(router, state, questions, config["model"], config["timeout_seconds"])
            return result, attempts, round(time.time() - started, 3), None
        except _TRANSIENT as exc:
            last_error = exc  # transport-level only; bad requests are not retried
    raise RuntimeError(f"decision request failed after {attempts} attempt(s): {last_error}")


# --------------------------------------------------------------------------
# policy: pure Python, no model calls
# --------------------------------------------------------------------------
def normalised_disruption(answer: Dict[str, Any], levels: int) -> float:
    """Map the model's zero-based expected level onto 0..1 in code.

    Laya returns a probability-weighted mean of zero-based level indices, so the
    divisor is len(criteria) - 1. This arithmetic belongs here, not in a question.
    """
    return float(answer["score"]) / (levels - 1)


def decide(answers: Dict[str, Any], questions: Dict[str, Any]) -> Dict[str, Any]:
    """Turn the three answers into one route. Confidence gates the action."""
    category = answers["category"]
    disruption = answers["disruption"]
    refund = answers["refund_requested"]

    category_confidence = float(category.get("confidence", 0.0))
    disruption_confidence = float(disruption.get("confidence", 0.0))
    refund_probability = float(refund["noul"])
    disruption_norm = normalised_disruption(disruption, len(questions["disruption"]["criteria"]))

    decision: Dict[str, Any] = {
        "route": None,
        "queue": None,
        "reason": None,
        "automatic_action": False,
        "human_review": False,
        "signals": {
            "category": category["choice"],
            "category_confidence": category_confidence,
            "category_confidence_floor": CATEGORY_CONFIDENCE_FLOOR,
            "disruption_score": disruption["score"],
            "disruption_normalised": round(disruption_norm, 4),
            "disruption_confidence": disruption_confidence,
            "refund_probability": refund_probability,
            "refund_threshold": REFUND_NOUL_THRESHOLD,
        },
    }

    # Gate 1: an unpicked category is never automated.
    if category_confidence < CATEGORY_CONFIDENCE_FLOOR:
        decision.update(
            route="human_review",
            queue="support.triage",
            reason=(
                f"category confidence {category_confidence:.2f} < floor "
                f"{CATEGORY_CONFIDENCE_FLOOR:.2f}; category left to a person"
            ),
            human_review=True,
        )
        return decision

    # Gate 2: the disruption score only pages a human when it is both high and peaked.
    blocking = disruption_norm >= DISRUPTION_ESCALATION and disruption_confidence >= DISRUPTION_CONFIDENCE_FLOOR

    if category["choice"] == "billing":
        if refund_probability >= REFUND_NOUL_THRESHOLD:
            decision.update(
                route="billing.refund_queue",
                queue="billing.refunds",
                reason=(
                    f"billing ticket with explicit refund request (P={refund_probability:.2f} "
                    f">= {REFUND_NOUL_THRESHOLD:.2f})"
                    + ("; customer is blocked" if blocking else "")
                ),
                automatic_action=True,
                refund_request_likely=True,
            )
        else:
            decision.update(
                route="billing.general",
                queue="billing.inquiries",
                reason=(
                    f"billing ticket without an explicit refund request "
                    f"(P={refund_probability:.2f} < {REFUND_NOUL_THRESHOLD:.2f})"
                ),
                automatic_action=True,
                refund_request_likely=False,
            )
    elif category["choice"] == "technical":
        if blocking:
            decision.update(
                route="engineering.escalate",
                queue="engineering.oncall",
                reason=(
                    f"blocking technical issue: disruption {disruption_norm:.2f} >= "
                    f"{DISRUPTION_ESCALATION:.2f} at confidence {disruption_confidence:.2f}"
                ),
                automatic_action=True,
            )
        else:
            decision.update(
                route="engineering.backlog",
                queue="engineering.backlog",
                reason=(
                    f"technical issue without a confirmed block: disruption {disruption_norm:.2f} "
                    f"< {DISRUPTION_ESCALATION:.2f} or confidence {disruption_confidence:.2f} "
                    f"< {DISRUPTION_CONFIDENCE_FLOOR:.2f}"
                ),
                automatic_action=True,
            )
    elif category["choice"] == "account_access":
        decision.update(
            route="support.accounts",
            queue="support.accounts",
            reason=(
                "account access request "
                f"(P(access)={category['probabilities']['account_access']:.2f})"
            ),
            automatic_action=True,
        )
    else:
        decision.update(
            route="human_review",
            queue="support.triage",
            reason=f"category {category['choice']!r} has no automated route; a person decides",
            human_review=True,
        )
    return decision


# --------------------------------------------------------------------------
# entry point
# --------------------------------------------------------------------------
def triage(ticket: Dict[str, str]) -> Dict[str, Any]:
    config = load_config()
    questions = build_questions()
    state = build_state(ticket)

    result, attempts, elapsed, error = run_decisions(config, state, questions)
    answers = result.get("answers", {})

    missing = [qid for qid in questions if qid not in answers]
    if missing:
        raise RuntimeError(f"model did not answer: {', '.join(missing)}")

    report = {
        "ticket": ticket,
        "config": config,
        "request": {"state": state, "questions": questions},
        "response": result,
        "routing_decision": decide(answers, questions),
        "call": {"attempts": attempts, "elapsed_seconds": elapsed, "error": error},
    }
    return report


def load_ticket(path: Optional[str], inline: Optional[str]) -> Dict[str, str]:
    if path:
        with open(path, "r", encoding="utf-8") as handle:
            ticket = json.load(handle)
    else:
        ticket = {"id": "inline", "title": "Inline ticket", "text": inline or ""}
    for field in ("title", "text"):
        if not str(ticket.get(field, "")).strip():
            raise ConfigError(f"ticket field {field!r} is required and must be non-empty")
    ticket.setdefault("id", "unnamed")
    return {"id": str(ticket["id"]), "title": str(ticket["title"]), "text": str(ticket["text"])}


def main(argv: Optional[List[str]] = None) -> int:
    parser = argparse.ArgumentParser(description="Triage one support ticket with the Laya decision model.")
    parser.add_argument("--ticket-file", help="JSON file with id/title/text")
    parser.add_argument("--text", help="Ticket body, used when --ticket-file is absent")
    parser.add_argument("--out", help="Write the full report JSON here (default: stdout)")
    parser.add_argument("--quiet", action="store_true", help="Suppress the human-readable summary")
    args = parser.parse_args(argv)

    try:
        ticket = load_ticket(args.ticket_file, args.text)
        report = triage(ticket)
    except (ConfigError, RuntimeError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 2

    payload = json.dumps(report, indent=2, ensure_ascii=False)
    if args.out:
        with open(args.out, "w", encoding="utf-8") as handle:
            handle.write(payload + "\n")

    if not args.quiet:
        answers = report["response"]["answers"]
        routing = report["response"].get("routing", {})
        decision = report["routing_decision"]
        print(f"ticket        : {report['ticket']['id']} — {report['ticket']['title']}")
        print(f"library       : {report['config']['library']} (model selection: {report['config']['model_selection']})")
        print(f"router model  : {routing.get('model')} — {routing.get('reason')}")
        print(f"category      : {answers['category']['choice']} "
              f"(confidence {answers['category']['confidence']}, "
              f"probs {answers['category']['probabilities']})")
        print(f"disruption    : score {answers['disruption']['score']} "
              f"(confidence {answers['disruption']['confidence']}, "
              f"probs {answers['disruption']['probabilities']})")
        print(f"refund asked  : noul {answers['refund_requested']['noul']} "
              f"(confidence {answers['refund_requested'].get('confidence')})")
        print(f"route         : {decision['route']} -> {decision['queue']}")
        print(f"reason        : {decision['reason']}")
        print(f"human review  : {decision['human_review']}")
        print(f"elapsed       : {report['call']['elapsed_seconds']}s in {report['call']['attempts']} attempt(s)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
