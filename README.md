# BARREL: Beverage Alcohol Review & Regulatory Evidence Logger

BARREL is an AI-native beverage alcohol label review assistant for evaluating label images against expected fields and advisory regulatory checks.

**Required product statement:** BARREL is a review assistant, not a final legal determination system.

## What BARREL does

- Accepts individual image uploads and `.zip` batch uploads.
- Uses the `ai_native` parser as the default analysis path.
- Sends each image to a hosted image-capable model with a prompt/schema to extract review fields.
- Runs deterministic BARREL checks and expected-vs-observed matching on top of the parsed output.
- Persists lightweight review evidence in object/blob-style storage.
- Shows a row-by-row Review History table where each submission can be reopened.
- Lets reviewers inspect the original image, parsed fields, confidence, evidence, deterministic checks, and review decision.

## Core workflow

```text
Evaluator logs into the web app
→ drags/drops one label image or a zip file
→ BARREL creates one async job per image
→ ai_native parses the image into structured candidate fields
→ BARREL runs deterministic checks and expected-vs-observed scoring
→ image/result/review evidence is stored in object/blob storage
→ Review History shows prior submissions row-by-row
→ clicking a row restores the original image and full result detail
→ reviewer approves, rejects, or marks needs-more-info
```

## Extracted fields

The AI-native parser targets these fields from the label image:

- brand name
- class/type
- alcohol content / proof
- net contents
- producer/bottler
- government warning text/presence
- country of origin if visible
- disclosures / legibility flags if visible

## Provider model

- Primary parser: `ai_native`
- Preferred backing provider: Azure OpenAI / Azure AI Foundry hosted API with image input
- Temporary fallback if Azure quota or subscription policy blocks deployment: direct OpenAI API using the same provider interface
- Optional debug/baseline provider: `azure_vision_ocr`

`azure_vision_ocr` is debug evidence only. It is not the normal product path and should not be documented as the primary parser.

## Storage model

BARREL prefers lightweight object/blob storage over a database for review evidence.

Azure implementation:

- Azure Blob Storage
- `jobs/{job_id}/image.<ext>`
- `jobs/{job_id}/ai_raw.json`
- `jobs/{job_id}/result.json`
- `jobs/{job_id}/decision.json`

If a document mentions “S3”, read it as the generic object-storage equivalent. The Azure implementation uses Azure Blob Storage.

## Architecture

```text
Browser
→ BARREL web UI
→ BARREL Go API
→ AI-native image parser via hosted API
→ deterministic BARREL checks
→ lightweight object/blob storage
→ review history/detail viewer
```

## Command interface

- `make setup` is for initial bootstrap and tool installation convenience.
- `task` is the daily command interface.

## Setup and verification

```text
make setup
az login
task azure:login-check
task dev
task smoke:local-ai
task azure:deploy
task azure:smoke:ai-native
```

Use local validation first. Do not claim success without smoke coverage.

## Evaluator access

- Demo username: `evaluator`
- Demo password: `fallback-demo-password-123` unless overridden by environment or Azure secrets

No public signup is expected for the take-home environment.
