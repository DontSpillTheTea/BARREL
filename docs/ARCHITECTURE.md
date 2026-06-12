# BARREL Architecture

## Product shape

BARREL is an AI-native beverage alcohol label review assistant. It accepts image and zip uploads, parses each label through Azure OpenAI (gpt-4.1-mini with vision), applies field-specific deterministic checks with TTB regulatory rule citations, stores lightweight review evidence, and exposes a review history/detail viewer for human decision-making.

**Required product statement:** BARREL is a review assistant, not a final legal determination system.

## High-level topology

```text
Browser (React 19 / Vite 8, port 5173)
  → BARREL Go API (net/http, port 8080)
    → Azure OpenAI gpt-4.1-mini (Sweden Central, image + structured JSON prompt)
    → Field-specific deterministic comparison engine
    → Local filesystem storage (Docker volume)
  → Review History / Detail viewer
```

## Analysis flow

### Single image

```text
POST /api/v1/labels/analyze-async (image + expected_json + beverage_type)
  → validate file (extension, size)
  → save image to storage
  → create async job (in-memory store)
  → return 202 with job_id

Background goroutine (processJob):
  → compress image (PNG→JPEG, max 768px, quality 50)
  → send to Azure OpenAI with structured extraction prompt
  → receive AINativeExtraction (brand, class, ABV, net contents, producer, country, gov warning with verbatim text)
  → run field-specific comparison engine:
    - Brand/Class: bigram similarity (≥0.85 Match, ≥0.70 Uncertain)
    - ABV: numeric tolerance (±0.3% spirits, ±1.0%/±1.5% wine per 27 CFR)
    - Net Contents: unit-normalized comparison (mL/L/fl oz)
    - Gov Warning: verbatim character-by-character comparison against 27 CFR § 16.21
    - Producer/Country: fuzzy presence matching
  → assign per-field status: Match / Mismatch / Missing on Label / Missing in Application Data / Uncertain
  → persist result with processing time measurement
  → update job status to succeeded

Frontend polls GET /api/v1/jobs/{id} → shows results
```

### Zip batch

```text
POST /api/v1/labels/analyze-async (file.zip)
  → extract images (.png, .jpg, .jpeg)
  → create one async job per image
  → process each in parallel goroutines
  → return 202 with job list
```

## Primary services

### Web UI (apps/web/)
- React 19 + Vite 8, single-file App.jsx
- Layout: top row (upload 1/3 + image viewbox 2/3), full-width field table, review history
- Collapsible expected fields form (brand, class, ABV, net contents, producer, country, beverage type, gov warning)
- Field table with similarity scores, AI confidence, CFR citations
- Expandable government warning character diff
- CSV export of verification results
- Processing time badge
- Approve/Reject decision workflow

### Go API (apps/api/)
- Standard library net/http, no framework
- Async job processing with in-memory store + storage persistence fallback
- AI provider interface: AzureOpenAIProvider (real) + MockProvider (testing)
- Image compression before sending to AI (JPEG quality 50, max 768px)
- Field-specific validators: abv.go, net_contents.go, similarity.go, normalize.go

### AI parser
- Provider: Azure OpenAI (gpt-4.1-mini with vision, Sweden Central)
- Deployment: barrel-ai-native-parser, 50K TPM
- Prompt: structured JSON extraction with strict gov warning transcription instructions
- Image sent as base64 with detail:low hint
- max_tokens: 1000, temperature: 0.1

## Deterministic comparison engine

Each field uses a field-specific comparison strategy:

| Field | Strategy | Threshold | CFR |
|-------|----------|-----------|-----|
| Brand Name | Bigram similarity | ≥0.85 Match | 27 CFR § 5.63 / § 4.33 / § 7.23 |
| Class/Type | Bigram similarity | ≥0.85 Match | 27 CFR § 5.141 / § 4.34 / § 7.24 |
| Alcohol Content | Numeric tolerance | ±0.3% spirits, ±1.0%/±1.5% wine | 27 CFR § 5.37 / § 4.36 / § 7.71 |
| Net Contents | Unit-normalized numeric | 1% tolerance | 27 CFR § 5.73 / § 4.37 / § 7.27 |
| Government Warning | Verbatim character match | Exact | 27 CFR § 16.21 |
| Producer/Bottler | Bigram similarity | ≥0.80 Match | 27 CFR § 5.66 / § 4.40 |
| Country of Origin | Word overlap | ≥50% Match | 27 CFR § 5.75 / § 4.41 |

Low AI confidence (<70%) forces any field to Uncertain regardless of text match.

## Storage model

Currently using local filesystem (Docker volume at /tmp/barrel_storage_local).

Storage layout per job:
```text
{job_id}/image.png
{job_id}/result.json
{job_id}/decision.json
```

Azure Blob Storage is defined in OpenTofu but not yet deployed. When deployed, same layout under `jobs/` container.

## Infrastructure

### Currently deployed
- Azure OpenAI account (Sweden Central) with gpt-4.1-mini deployment

### Defined in OpenTofu but NOT yet deployed
- Azure Container Apps (API + Web)
- Azure Container Registry
- Azure Blob Storage
- Azure Key Vault
- Log Analytics workspace

### Local development
- Docker Compose: api (Go, port 8080) + web (React, port 5173)
- `task up:ai` starts the stack (requires Azure OpenAI creds in .env)
- `task go:ai` + `task web` for separate terminal dev

The architecture stays simple: hosted AI-based parsing, field-specific deterministic validation, local/object storage, and a reviewer-focused UI.
