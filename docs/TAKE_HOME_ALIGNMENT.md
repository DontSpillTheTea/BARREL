# Take-home Alignment

This document explains how BARREL aligns with the take-home requirements in its current AI-native form.

**Note:** `task` is the supported daily command interface. `make setup` is for bootstrap convenience.

## Why this implementation fits the assignment

- It is a practical deployed review tool, not a speculative platform build.
- It uses AI-native image parsing where layout understanding matters.
- It keeps deterministic checks and an evidence trail instead of pretending the model is the compliance authority.
- It centers the human review workflow: inspect, compare, decide, and log.
- It uses lightweight object/blob storage instead of overbuilding around a database too early.

## Product approach

BARREL accepts one label image or a zip file of images, parses each image through a hosted image-capable model, runs deterministic validation, stores lightweight evidence objects, and presents the result in a reviewer-facing history/detail interface.

This is aligned with the core product statement:

```text
BARREL is a review assistant, not a final legal determination system.
```

## Architecture choices

- **Go API**: a good fit for HTTP handling, async job orchestration, structured JSON contracts, and deterministic validation.
- **Hosted AI parser**: appropriate for image-native field extraction without standing up managed compute.
- **Azure-first deployment**: consistent with the take-home environment and the current product direction.
- **Object/blob review storage**: sufficient for review evidence without prematurely introducing Postgres.

## Verification approach

- local verification before Azure smoke
- API and smoke tests before any success claim
- AI provider endpoint proof before BARREL deployment claims
- reviewer-visible history/detail behavior treated as part of the product, not an optional extra

## Evidence and reviewability

BARREL keeps the system reviewable by:

- surfacing parsed fields
- surfacing confidence and evidence
- applying deterministic checks
- retaining lightweight evidence artifacts
- letting reviewers reopen prior submissions from Review History

## What this is not

- not a final legal determination engine
- not a GPU or managed-ML infrastructure exercise
- not a database-heavy compliance platform unless later scale needs prove that necessary
