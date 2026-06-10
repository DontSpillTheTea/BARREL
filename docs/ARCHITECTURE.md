# BARREL Architecture

## Product shape

BARREL is an AI-native beverage alcohol label review assistant. It accepts image and zip uploads, parses each label through a hosted image-capable model, applies deterministic checks, stores lightweight review evidence, and exposes a review history/detail viewer for human decision-making.

**Required product statement:** BARREL is a review assistant, not a final legal determination system.

## High-level topology

```text
Browser
→ BARREL web UI
→ BARREL Go API
→ AI-native image parser via hosted API
→ deterministic BARREL checks
→ object/blob storage
→ review history/detail viewer
```

## Analysis flow

### Single image

```text
upload image
→ create async job
→ call ai_native with image + prompt/schema
→ receive structured candidate fields
→ run deterministic checks and expected-vs-observed scoring
→ persist evidence objects
→ show result detail and history entry
```

### Zip batch

```text
upload zip
→ extract image entries
→ create one async job per image
→ process each image through ai_native
→ persist evidence objects per job
→ show queue/history rows
→ allow row click to restore detail for each completed job
```

## Primary services

### Web UI

- React/Vite evaluator interface
- login-gated
- upload form for single image and zip batch submissions
- result detail panel with original image, parsed fields, evidence, confidence, and deterministic checks
- full-width Review History table with selectable rows

### Go API

- receives uploads
- creates async jobs
- submits images to the configured AI provider
- runs deterministic validation and expected-vs-observed matching
- persists image/result/review metadata
- serves history and detail endpoints

### AI parser

- Primary provider ID: `ai_native`
- Hosted API-based image parser using image input plus prompt/schema
- Preferred implementation: Azure OpenAI / Azure AI Foundry hosted API if quota is available
- Temporary fallback: direct OpenAI API using the same provider contract if Azure OpenAI is blocked by subscription quota or region policy

### Optional debug provider

- Provider ID: `azure_vision_ocr`
- Purpose: debug/baseline evidence only
- Not the primary product path

## Deterministic layer

BARREL does not treat the model output as a final compliance decision.

Deterministic logic is responsible for:

- normalization of parsed fields
- expected-vs-observed comparisons
- alcohol content / proof checks
- net contents checks
- government warning validation
- match scoring and advisory status assignment
- confidence and evidence presentation

## Storage model

BARREL prefers object/blob storage over Postgres for review evidence.

Azure implementation:

- Azure Blob Storage for images and review artifacts
- lightweight metadata persisted alongside result artifacts

Recommended storage layout:

```text
jobs/{job_id}/image.<ext>
jobs/{job_id}/ai_raw.json
jobs/{job_id}/result.json
jobs/{job_id}/decision.json
```

If “S3” appears in generic planning language, it means the object-storage pattern. The Azure implementation uses Azure Blob Storage.

## History and detail viewer

Review History is a first-class reviewer feature:

- each processed submission appears row-by-row
- rows show submission and result metadata
- selecting a row restores the original image
- selecting a row restores parsed fields, confidence, evidence, deterministic checks, and review decision state

## Infrastructure direction

- Azure Container Apps for web and API hosting
- Azure Blob Storage for review evidence
- Azure-managed secrets for provider credentials and demo access
- OpenTofu + Terragrunt for infrastructure management

The architecture should stay simple: hosted API-based parsing, deterministic validation, object/blob storage, and a reviewer-focused UI.
