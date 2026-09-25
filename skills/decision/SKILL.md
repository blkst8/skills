---
name: decision
description: >
  Build, test, and improve programs that use typed decision models such as
  Laya, Jev, or any compatible choice/score/noul backend. Use this skill whenever
  the user asks to classify, score, gate, route, extract, triage, moderate,
  or make other bounded judgments from text or structured state, especially
  when they mention Laya, decision models, probabilities, confidence thresholds,
  choice/score/noul questions, or a decision-model API.
compatibility: []
---

# Writing and improving decision-model programs

Use this skill to build programs with typed decision models. The default model is [Laya](https://github.com/NandhaKishorM/laya), an open-source non-autoregressive decision engine. The same guidance applies to another decision model that accepts the same request and response schema.

A decision model reads one `state`, answers every question in the request independently, and returns a probability distribution over the answer space defined by the question. It does not reason in steps, generate an essay, or own application control flow. Code owns retrieval, arithmetic, branching, weights, thresholds, and side effects. The model owns fast, bounded judgments.

A good decision question is one a knowledgeable person can answer in a second from the supplied state. “Does this message convey urgency?” is suitable. “Analyze this message and decide what to do” is not. Split the latter into small questions and compose the answers in code.

The default local Laya adapter uses the `laya` package. Configure the library source, optional inference endpoint, credentials, and model through environment variables rather than hard-coding them.

## Configure the decision backend

The following variables form the portable configuration surface. A deployment may use a local library, a remote HTTP service, or another model with the same schema.

| Variable | Required | Meaning |
| --- | --- | --- |
| `DECISION_LIBRARY` | No | Decision library or adapter to load. Default: `laya`. |
| `DECISION_LIBRARY_URL` | No | Installation or library source. Default: `https://github.com/NandhaKishorM/laya`. |
| `DECISION_API_KEY` | For authenticated services | API key or bearer token. Keep it secret; never put it in prompts, source, or logs. |
| `DECISION_BASE_URL` | For remote services | Base URL of the inference API. Omit it when the library runs locally. |
| `DECISION_MODEL` | No | Model or checkpoint name. For the Laya `Router`, values can include `english`, `multilingual`, and `typed-decisions`; the corresponding checkpoint repositories are `laya`, `laya-multilingual`, and `laya-typed-decisions`. A router may select a checkpoint when the model is omitted. |
| `DECISION_TIMEOUT` | No | Request timeout, for example `30s`. Use a finite timeout in production. |
| `DECISION_MAX_RETRIES` | No | Maximum retry count for transient transport failures. Do not retry invalid requests or ambiguous decisions automatically. |

Example:

```bash
export DECISION_LIBRARY=laya
export DECISION_LIBRARY_URL=https://github.com/NandhaKishorM/laya
export DECISION_MODEL=typed-decisions
# Required only when DECISION_BASE_URL points to an authenticated service:
export DECISION_BASE_URL=https://decision.example.internal
export DECISION_API_KEY='...'
```

Treat environment variables as deployment configuration, not as model instructions. Validate them at startup. Fail with a clear message when a required variable is missing; do not silently fall back to a different model, endpoint, or credential.

The Laya library can run locally without an API key or a base URL:

```bash
python -m pip install laya
```

For a source checkout, use the configured `DECISION_LIBRARY_URL` as the installation source. Do not import a model by path from untrusted input, and do not send private state to a remote endpoint without an explicit policy decision.

## The portable request schema

All adapters expose the same logical request:

```json
{
  "state": {
    "ticket": {
      "title": "Checkout fails after payment",
      "text": "The payment succeeded, but the checkout page keeps loading."
    }
  },
  "questions": {
    "category": {
      "type": "choice",
      "instructions": "Which category best describes `ticket.title` and `ticket.text`?",
      "criteria": {
        "billing": "Payments, charges, invoices, and refunds",
        "technical": "Broken features, errors, and outages",
        "other": "Anything not covered by another option"
      }
    },
    "severity": {
      "type": "score",
      "instructions": "How disruptive is the issue in `ticket.text`?",
      "criteria": [
        "No meaningful disruption to the user",
        "A degraded feature with a workaround",
        "A blocking issue with no workaround"
      ]
    },
    "refund_requested": {
      "type": "noul",
      "instructions": "Does `ticket.text` explicitly request a refund or credit?"
    }
  }
}
```

The `state` is structured data or text. Keep question IDs in the application; IDs identify answers locally and need not be sent to the model. Every question in a request is evaluated independently and may be evaluated in parallel. A question never sees another question’s answer in the same request.

The response has this logical shape:

```json
{
  "answers": {
    "category": {
      "choice": "technical",
      "probabilities": {"billing": 0.03, "technical": 0.91, "other": 0.06},
      "confidence": 0.82
    },
    "severity": {
      "score": 1.0,
      "probabilities": [0.08, 0.76, 0.16],
      "confidence": 0.76,
      "legend": ["No meaningful disruption", "Degraded with workaround", "Blocking without workaround"]
    },
    "refund_requested": {
      "noul": 0.03
    }
  },
  "routing": {
    "model": "laya"
  }
}
```

Transport details vary. Laya’s Python API returns a dictionary from `Router.predict`; a remote adapter should preserve the logical `state`, `questions`, and `answers` fields even if its wire format differs. Keep schema translation in the adapter, not in business logic.

## Workflow

1. List the decisions the application must make. Express each as a branch, threshold, ranking, or route.
2. Write one atomic question per judgment. Split any question that weighs two properties.
3. Choose the primitive whose answer the application can act on directly.
4. Build the smallest state that answers every question. Compute deterministic work in code.
5. Put questions that share the same state into one request, including questions needed only by some branches.
6. Combine answers in code with branches, weights, and confidence gates.
7. Test against labeled examples. Inspect probabilities on misses, then revise one or two questions at a time.

Do not turn the model into a hidden policy engine. If a decision requires sequential reasoning, fetch data, or an action policy, make that dependency explicit in code.

## Choose the primitive

| Primitive | Use it when | Returned signal | Code acts on it with |
| --- | --- | --- | --- |
| `choice` | The answer is one of a known, unordered set | `choice`, `probabilities`, optional `confidence` | A branch per option |
| `score` | The answer is an ordered spectrum that can be described as distinct situations | `score`, `probabilities`, optional `confidence`, optional `legend` | A threshold, rank, or weight |
| `noul` | The answer is a crisp yes/no condition and its probability is the signal | `noul` from 0 to 1 | An `if` on a threshold |

- Add an `other` or `none of the above` option when a choice may not cover every input.
- `noul: 0.5` means the model is uncertain, not “medium.” Use `score` for a degree.
- A `noul` needs a crisp condition. “Is this candidate strong in Python?” is vague. “Does the resume state that the candidate used Python at work?” is crisp.
- Use `choice` or several `noul` questions when there is no meaningful in-between state.
- Do not ask a decision model to count, sum, compare dates, or perform arbitrary arithmetic. Do that in code.

## Write the instructions

- State the exact condition. The model answers the words and boundaries provided.
- Ask one property per question. Hidden second judgments reduce clarity and confidence.
- Name the state field being judged with a backticked path, such as `` `ticket.messages[0].text` ``.
- Use direct language. Avoid double negatives, nested properties, and questions requiring several inference hops.
- Keep numeric level labels out of the instructions. Describe the situations in `criteria` instead of saying “rate from 0 to 2.”
- Put the complete question in `instructions`. The question ID is an application key, not a semantic instruction.
- Keep policy in code. “A shared address cannot override a name conflict” is an application rule, not a model judgment.
- When a wrong answer reveals a missing condition, add that condition rather than relying on an unstated implication.

Instructions may be a string, an object, or an adapter-supported array. Prefer an object when the question has labeled parts or supporting data. Pass schemas, taxonomies, and rows as structured JSON; do not flatten them into a string template.

```json
"instructions": {
  "question": "Does `message` ask the recipient to disclose a sensitive credential?",
  "inspect": "message",
  "focus": "Look for a request to send the credential itself, not a request to change or reset it."
}
```

Common instruction fields include `question`, `focus`, `inspect`, `note`, `compare` for state paths, and `field` for a shared record containing `name`, `type`, `unit`, and `description`. If an adapter uses a different field set, map it in the adapter while preserving the meaning.

## Write the criteria

Criteria extend the instruction. Both must describe the same property and direction. A `noul` whose `true` side describes “no” is especially error-prone.

**Choice.** Map every option to a description. Make close options contrastive:

```json
"billing": {
  "what": "Charges, invoices, refunds, or subscriptions",
  "not_for": "Order tracking or account access",
  "examples": ["I was charged twice", "Where is my refund?"]
}
```

**Score.** List levels from low to high. Use only as many levels as you can describe distinctly:

- Describe situations, not degrees such as “moderately severe.”
- Make each level stand alone. The model evaluates each level independently; “worse than the previous level” is not useful.
- Keep one dimension per score. Do not combine punctuality, intelligence, and experience in one question.
- Give a rare extreme its own level when code must handle it differently.
- An adapter may support an object level, for example `{"summary": "One clearly stated change", "signals": ["A single fix or feature"]}`.

**Noul.** Criteria may be omitted for an obvious question. For a subtle boundary, provide `true` and `false` sides, each with a description and examples. Put the neighboring case in the side it belongs to.

Examples should be short, concrete instances such as “I was charged twice,” not descriptions such as “a message about a billing problem.” Do not put all domain knowledge into examples; general rules belong in instructions and criteria.

## Build the state

- Send only fields needed by the questions. Unrelated context reduces clarity and makes failures hard to diagnose.
- Retrieve and filter in code first. If code cannot filter, ask a relevance `noul` per passage and retain passages that pass a threshold.
- Keep the state structured so instructions can point to stable paths.
- Convert opaque numeric encodings to words or named buckets before sending them. Prefer “red” over `#ff0000`, and a computed named bucket over a raw magnitude when the question is categorical.
- Compute date order, duration, windows, counts, and sums in code and send the result only when the model must judge it.
- Respect the selected model’s context limit. Limits are model-specific; do not copy limits from another backend. Measure token usage with the actual tokenizer or adapter.
- Treat state text as able to influence the answer. Decision models are not a security boundary. State the criteria precisely, test prompt-injection and self-describing content, and gate high-impact actions.

## Compose the answers in code

**Speculative fan-out.** Put every question that might be needed for the current state into one request, including questions used only by one branch. The application ignores answers it does not need. This usually reduces latency compared with sequential calls, but do not add irrelevant questions merely to fill a batch.

**Second requests.** Make a second request only when its state or question cannot be constructed without the first answer, such as a classification that determines which document to fetch. Questions in one request cannot depend on sibling answers.

**Confidence-gated routing.** An answer says what; confidence or probability says how safely to act. Set a floor below which no automatic action runs, then use a higher threshold for expensive or dangerous actions. Start with conservative thresholds and tune them on labeled data. Standard paths are: act, ask for confirmation or flag, and hand off to a person.

Do not assume every model returns a `confidence` field. For a `noul`, use a calibrated probability threshold and retain the full distribution when the adapter provides one. A distance from 0.5 is not a universal confidence score; calibrate it for the model and domain.

**Composite scoring.** Split a complex judgment into one `score` per dimension. Normalize a score by `len(criteria) - 1` only when the model’s score uses zero-based ordered levels. Confirm the adapter’s score convention first. Combine normalized dimensions with explicit weights in code; change policy in code, not by rewriting questions.

**Intent routing.** Classify with a `choice` and place a complexity `score` beside it when needed. Route each intent to deterministic code, a specialist model, or a person. Send low-confidence classifications to a safe fallback.

**Taxonomy walk.** Ask one `choice` per tree level and walk the tree in code. Give each option a concise criteria value describing what belongs under it. Trim large subtrees to direct children and representative leaves. Follow multiple branches when probabilities are close and a later decision can resolve them.

**Counting.** Ask one `noul` per item in a single request, then sum the items passing your code threshold. Do not ask the model for a count directly.

**Dates.** Extract date parts with a `choice` over enumerated values, including `not stated`. Assemble and compare dates in code.

**Extraction.** Generate candidates with a regex, database query, or generative model. Ask the decision model to select among candidates with a `choice`, or verify one candidate with a `noul`. Validate the final value against the source schema.

## Laya example

Laya is the default implementation. The adapter reads `DECISION_MODEL`; when it is unset, Laya’s `Router` can select a checkpoint from the request and return the selected model in `routing.model`.

```python
import os

from laya import Router


def build_router() -> Router:
    library = os.getenv("DECISION_LIBRARY", "laya")
    if library != "laya":
        raise ValueError(f"Unsupported DECISION_LIBRARY: {library}")

    return Router()


def classify_ticket(ticket_id: str, ticket_text: str) -> None:
    router = build_router()
    questions = {
        "category": {
            "type": "choice",
            "instructions": "Which category best describes the main request in `ticket`?",
            "criteria": {
                "bug_report": "Something is broken or producing errors",
                "billing": "Charges, invoices, refunds, or subscriptions",
                "other": "Anything else",
            },
        },
        "severity": {
            "type": "score",
            "instructions": "How disruptive is the issue reported in `ticket`?",
            "criteria": [
                "Cosmetic; no impact to functionality",
                "Broken or degraded feature; a workaround exists",
                "Blocking issue; no workaround exists",
            ],
        },
        "refund_requested": {
            "type": "noul",
            "instructions": "Does `ticket` explicitly ask for a refund or credit?",
        },
    }

    model = os.getenv("DECISION_MODEL")
    kwargs = {"model": model} if model else {}
    result = router.predict({"ticket": ticket_text}, questions, **kwargs)
    answers = result["answers"]

    category = answers["category"]
    severity = answers["severity"]
    refund_requested = answers["refund_requested"]

    if category.get("confidence", 0.0) < 0.6:
        route_to_human(ticket_id, reason="low category confidence")
    elif category["choice"] == "bug_report":
        normalized = severity["score"] / (len(questions["severity"]["criteria"]) - 1)
        if normalized > 0.75 and severity.get("confidence", 0.0) > 0.5:
            escalate(ticket_id)
        else:
            add_to_backlog(ticket_id)
    elif category["choice"] == "billing" and refund_requested["noul"] > 0.7:
        route_to_billing(ticket_id, refund_likely=True)
```

The example intentionally keeps routing policy in Python. For a remote adapter, use `DECISION_BASE_URL` and `DECISION_API_KEY` in the transport layer, preserve the same logical schema, and never expose credentials to the model state.

## Read the answers

- `score` may be a probability-weighted mean of ordered levels. The same score can represent certainty at one level or a mixture across levels; read `probabilities` with it.
- Use a score for thresholds, ranking, or rounding to a declared level. Do not treat an ordinal scale as a precisely calibrated physical quantity.
- `confidence`, when present, describes how peaked the answer distribution is; it does not guarantee correctness.
- A `noul` may not have a separate `confidence` field. Use its probability and a domain-calibrated threshold.
- Every answer remains within the declared answer space, so application code does not parse generated prose.
- Preserve the response distribution for analysis and calibration. Do not retain only the selected option when misses need investigation.

## Improve a program

Find the failing question before changing anything. Collect labeled examples, run them, and compare each answer and probability with the label.

| Symptom | Likely cause | Fix |
| --- | --- | --- |
| Wrong answers with high confidence | The condition was interpreted literally or was underspecified | State the exact condition and put the boundary case in criteria. |
| Low confidence on a choice | Options overlap or no option fits | Add `what`, `not_for`, and examples; add `other` when needed. |
| Low confidence on a score | Levels overlap, the question combines dimensions, or state is insufficient | Rewrite levels as distinct situations, split the question, and add missing state. |
| Scores cluster in the middle | Levels are vague degrees or raw numbers | Describe one concrete situation per level. |
| Extreme cases look alike | The extreme has no distinct level | Add a level only if code must treat it differently. |
| A `noul` hovers near 0.5 | The condition or boundary is vague | Define it and add true/false examples. |
| Accuracy falls as inputs grow | State contains irrelevant detail | Filter in code and send only necessary fields. |
| Errors in counts, sums, dates, or numeric nearness | The model is being asked to compute | Move computation to code and ask only for extraction or per-item judgments. |
| Errors in nested or negated questions | The question requires too much indirection | Ask a direct question, name the path, and split it if needed. |
| The answer follows text inside state | Untrusted content influenced the judgment | Tighten criteria, test adversarial inputs, and gate the action. |
| Rewording trades one error for another | One question weighs several properties | Split it into atomic questions and combine in code. |
| Each answer is right but the final decision is wrong | Policy is wrong | Change code weights or thresholds. |
| The program is slow or costly | Independent questions use sequential calls | Merge compatible questions into one request; retain only true dependencies. |

Revision rules:

- Change one or two questions per revision. Probabilities can shift unpredictably; leave questions that discriminate well untouched.
- Evaluate revisions on labeled data. Higher confidence alone is not evidence of better accuracy.
- Keep the answer space stable once code depends on it. Adding or removing an option or level changes the meaning of earlier answers.
- Put general rules in instructions and criteria; put specific names and values in examples.
- Recalibrate thresholds after changing models, prompts, criteria, or state construction. Do not transfer confidence thresholds blindly between decision models.
- Compare models on the same labeled set, state construction, question schema, and evaluation metric before selecting `DECISION_MODEL`.

## Checklist

- [ ] The skill is used for typed judgment, classification, scoring, yes/no decisions, and decision routing.
- [ ] `DECISION_LIBRARY`, `DECISION_LIBRARY_URL`, `DECISION_BASE_URL`, `DECISION_API_KEY`, and `DECISION_MODEL` have clear deployment roles.
- [ ] The default library is Laya, with `https://github.com/NandhaKishorM/laya` as its source URL.
- [ ] Each question asks one property and a person could answer it in a second.
- [ ] The primitive matches how code uses the answer.
- [ ] Instructions state the exact condition and name state paths in backticks.
- [ ] Criteria agree with instructions and point in the same direction.
- [ ] Score levels describe distinct situations and follow the selected model’s level convention.
- [ ] Choices that may not cover every input have an `other` or equivalent option.
- [ ] Code performs counting, arithmetic, date comparison, retrieval, and side effects.
- [ ] State contains only what questions need and respects the selected model’s context limit.
- [ ] Compatible questions travel in one request; dependent questions use a later request.
- [ ] Every automatic action has a calibrated threshold and a safe fallback.
- [ ] Weights and policy live in code, not in the model’s instructions.
- [ ] API keys and environment values are kept out of prompts, source, logs, and model state.
- [ ] Labeled examples evaluate every model or prompt revision.
