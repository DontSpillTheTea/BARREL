# BARREL: Beverage Alcohol Review & Regulatory Evidence Logger

BARREL is an AI-assisted beverage alcohol label verification prototype for TTB compliance reviewers. It extracts label fields from uploaded images using OCR and AI vision, compares them against expected application data using field-specific regulatory matching rules, and presents results for human review.

**BARREL is a review assistant, not a final legal determination system.**

## Deployed Prototype

| | URL |
|---|---|
| **Web App** | https://barrel-web.mangodune-45281216.swedencentral.azurecontainerapps.io |
| **API** | https://barrel-api.mangodune-45281216.swedencentral.azurecontainerapps.io |
| **Login** | Username: `evaluator` / Password: `fallback-demo-password-123` |

If login fails, clear your browser's localStorage (`BARREL_REVIEW_TOKEN`) and retry.

## How It Works

```text
Upload label image (or ZIP batch)
→ Azure Vision OCR extracts text (~0.5s)
→ Text parser classifies fields into structured candidates (~2-4s)
→ Deterministic validators compare expected vs extracted fields (<1ms)
→ If confidence is high and all fields present → fast result returned
→ If uncertain, missing fields, or mismatch → escalates to AI vision (~10-15s)
→ Result stored in Azure Blob Storage
→ Review History shows all submissions with clickable detail
→ Reviewer approves or rejects
```

## Field Verification

Seven fields are validated with field-specific comparison strategies:

| Field | Strategy | Regulatory Basis |
|-------|----------|-----------------|
| Brand Name | Tolerant bigram similarity (≥85% match) | 27 CFR § 5.63 / § 4.33 / § 7.23 |
| Class/Type | Tolerant bigram similarity | 27 CFR § 5.141 / § 4.34 / § 7.24 |
| Alcohol Content | Numeric tolerance: ±0.3% spirits, ±1.0%/±1.5% wine | 27 CFR § 5.37 / § 4.36 / § 7.71 |
| Net Contents | Unit-normalized comparison (mL/L/fl oz) | 27 CFR § 5.73 / § 4.37 / § 7.27 |
| Government Warning | **Strict** verbatim character match | 27 CFR § 16.21 |
| Producer/Bottler | Fuzzy presence (≥80%) | 27 CFR § 5.66 / § 4.40 |
| Country of Origin | Word overlap (≥50%) | 27 CFR § 5.75 / § 4.41 |

**Status labels**: Match, Mismatch, Missing on Label, Missing in Application Data, Uncertain.

### Government Warning Validation

The government warning check is strict per PRD requirements:
- `GOVERNMENT WARNING:` prefix must be ALL CAPS
- Full statutory body text must match character-by-character against 27 CFR § 16.21 canonical text
- Hallucinated or pseudo-warning text from AI is flagged as Mismatch
- Illegible or partial warning text becomes Uncertain, not Match
- Character-level diff is available in the UI

## Local Development

### Prerequisites

- Docker and Docker Compose
- [Task](https://taskfile.dev/) (`go-task`)
- Azure OpenAI credentials in `.env` (copy from `.env.example`)

### Run locally

```bash
task up:ai        # Start API + Web via Docker Compose
                  # Web: http://localhost:5173
                  # API: http://localhost:8080
```

Or separately:
```bash
task go:ai        # Terminal 1: Go API
task web          # Terminal 2: React frontend
```

### Test commands

| Command | What it tests |
|---------|--------------|
| `cd apps/api && go test ./...` | Go unit tests (ABV tolerance, net contents, similarity, gov warning) |
| `task smoke:local-ai` | End-to-end analysis smoke test |
| `task playwright:local` | Browser E2E tests (upload, history, batch) |
| `task test:web` | Frontend build check |

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | React 19, Vite 8 |
| Backend | Go (net/http, no framework) |
| AI Vision | Azure OpenAI gpt-4.1-mini (Sweden Central) |
| OCR | Azure Vision OCR (Sweden Central) |
| Text Parser | Azure OpenAI gpt-4.1-mini (text-only, no image tokens) |
| Storage | Azure Blob Storage (cloud) / local filesystem (dev) |
| Infrastructure | Azure Container Apps, ACR, Key Vault, OpenTofu + Terragrunt |

## AI/OCR Pipeline

BARREL uses a tiered analysis pipeline:

1. **Azure Vision OCR** extracts text from the label image (~0.5s, ~$0.001/image)
2. **Text Parser** (gpt-4.1-mini, text-only) classifies OCR text into structured field candidates (~2-4s)
3. **Deterministic Validators** compare extracted vs expected values using field-specific strategies
4. **Escalation Gate** checks confidence — if fields are missing, mismatched, or uncertain, escalates to:
5. **AI Vision** (gpt-4.1-mini with image) for full vision-based extraction (~10-15s)

Producer/bottler and country of origin are optional fields and do not trigger escalation.

## Outbound Network Dependencies

| Service | Endpoint | Data Sent |
|---------|----------|-----------|
| Azure OpenAI | `barrel-openai-sweden.openai.azure.com` | Label images (base64) or OCR text for field extraction |
| Azure Vision OCR | `barrel-vision-dev-*.cognitiveservices.azure.com` | Label images for text extraction |
| Azure Blob Storage | `barrelsadev.blob.core.windows.net` | Review evidence (images, results, decisions) |

All services are in Azure Sweden Central. No other outbound domains are required.

## Data Handling

- Uploaded label images and analysis results are stored in Azure Blob Storage (private container)
- Evidence layout: `jobs/{job_id}/image.png`, `jobs/{job_id}/result.json`, `jobs/{job_id}/decision.json`
- Evidence persists across container restarts
- Local development uses filesystem storage (`/tmp/barrel_storage_local`)
- No data is shared with third parties beyond the Azure AI services listed above
- Generated test samples use fictional brands and are not approved COLAs

## Performance

| Path | Typical Time | When Used |
|------|-------------|-----------|
| OCR fast path | ~3-5s | Clear labels with high-confidence OCR |
| OCR + AI escalation | ~12-18s | Mismatches, missing fields, unclear text |
| AI-native only | ~10-15s | Direct AI vision without OCR tier |

The PRD target is ~5 seconds for normal/simple labels. The OCR fast path approaches this; escalated labels take longer. Processing time is displayed in the UI.

## Assumptions

- Reviewers may use Windows; Docker Desktop + WSL2 is recommended for local development
- Azure services require outbound network access
- Generated label images are used for testing (fictional brands, not real COLAs)
- Rule checks are advisory prototypes based on 27 CFR Parts 4, 5, 7, and 16
- Human review is always required; BARREL does not make final legal determinations

## Known Limitations

- Bold detection for `GOVERNMENT WARNING:` prefix is not implemented (documented in PRD as P2)
- Visual bounding boxes on label images are not implemented
- CSV upload for per-image expected data in ZIP batches is not implemented
- No production-grade FedRAMP, audit logging, or enterprise RBAC
- No direct COLA system integration
- AI extraction accuracy depends on image quality; poor images produce Uncertain status
- Processing time varies by label complexity and escalation path

## Infrastructure

See [`infra/README.md`](infra/README.md) for complete Azure deployment instructions, including:
- Prerequisites and tool installation
- Pre-existing resource setup (Azure OpenAI, Vision OCR)
- Step-by-step first deploy
- What gets created (Container Apps, ACR, Storage, Key Vault)
- Updating, tearing down, and troubleshooting

## Trade-offs

| Decision | Rationale |
|----------|-----------|
| Azure OpenAI vision over local models | Better accuracy on complex label layouts; acceptable latency for prototype |
| Tiered OCR→AI pipeline | Balances speed (OCR fast path) with accuracy (AI vision for hard cases) |
| Blob storage over database | Lightweight evidence persistence; sufficient for prototype scale |
| Field-specific matching strategies | PRD requires different strictness: tolerant for brand names, strict for government warning |
| Go backend with no framework | Minimal dependencies, easy deployment, sufficient for prototype scope |
