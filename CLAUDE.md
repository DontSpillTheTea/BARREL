# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is BARREL

BARREL (Beverage Alcohol Review & Regulatory Evidence Logger) is an AI-native label review assistant. It accepts beverage label images, extracts fields via Azure OpenAI (GPT-4o), runs deterministic regulatory checks against TTB rules, and persists review evidence for human decision-making.

**Required product statement:** BARREL is a review assistant, not a final legal determination system.

## Commands

`task` (Taskfile.yml) is the daily command interface. `make setup` is for initial bootstrap only.

| Action | Command |
|--------|---------|
| Run API locally | `task go:ai` |
| Run frontend locally | `task web` |
| Docker stack (requires Azure OpenAI creds in .env) | `task up:ai` |
| Stop stack | `task down` |
| Go unit tests | `task test:api` |
| Single Go package test | `cd apps/api && go test -v ./internal/analysis` |
| Frontend build check | `task test:web` |
| Frontend lint | `cd apps/web && npm run lint` |
| Playwright E2E | `task playwright:local` |
| AI smoke test | `task smoke:local-ai` |
| Azure deploy | `task azure:build && task azure:deploy` |
| Azure smoke | `task azure:smoke:ai-native` |
| Check env setup | `task check-env` |

## Architecture

```
Browser (React 19 / Vite 8, port 5173)
  -> Go API (net/http, port 8080)
       -> Azure OpenAI GPT-4o (ai_native parser, image input)
       -> Deterministic checks (analysis package + TTB rule catalog)
       -> Object storage (local filesystem or Azure Blob)
```

The OCR worker (`apps/ocr-worker/`, Python/Flask, port 9090) is an optional debug baseline; the primary path is `ai_native` which sends images directly to Azure OpenAI.

### Request flow

1. Frontend POSTs image (or .zip batch) to `/api/v1/labels/analyze-async`
2. API creates an async job per image, returns 202 with job IDs
3. Background goroutine (`processJob` in `main.go`) runs the pipeline:
   - Saves image to storage
   - If `ai_native`: calls `aiProvider.SecondRead()` with image bytes, then `analysis.AnalyzeAI()` for deterministic scoring
   - If OCR provider: calls OCR worker, then `analysis.AnalyzeText()` for text-based checks
4. Frontend polls `GET /api/v1/jobs/{id}` until status is `succeeded`
5. Results appear in Review History table; reviewer can approve/reject/needs-more-info

### Two analysis paths

- **`analysis.AnalyzeText()`** (`analysis.go`): OCR text-based. Uses regex validators and fuzzy matching against expected fields. Computes AI escalation eligibility.
- **`analysis.AnalyzeAI()`** (`ai_native.go`): AI-extracted fields. Compares AI candidates against expected fields with case-insensitive matching. Both paths produce `[]FieldCheckResult` with status Pass/Needs Review/Likely Fail.

## Key packages (apps/api/internal/)

- **ai/**: `Provider` interface with `SecondRead(ctx, SecondReadInput) -> *AISecondRead`. Implementations: `azureopenai.go` (real), `mock.go` (fallback when no credentials).
- **analysis/**: Deterministic field validation. `AnalyzeText` for OCR path, `AnalyzeAI` for AI-native path.
- **models/**: All shared types. Core: `LabelAnalysisResult`, `ExpectedLabelFields`, `FieldCheckResult`, `AINativeExtraction`, `AISecondRead`.
- **storage/**: `Provider` interface for persisting images, results, and decisions. `local.go` (filesystem) and `azureblob.go` (Azure Blob).
- **jobs/**: In-memory async job store with status tracking. Falls back to storage for persistence across restarts.
- **rules/**: Loads TTB regulatory YAML from `rules/ttb/`. Indexed by ID with CFR citation breadcrumbs.
- **security/**: Auth (demo username/password + review token), CORS, upload validation.
- **validators/**: Field extraction helpers (ABV regex, net contents, government warning, fuzzy matching).

## Configuration

All config via environment variables (see `.env.example`). Key settings:

- `AI_NATIVE_ENABLED` / `AZURE_OPENAI_ENDPOINT` / `AZURE_OPENAI_API_KEY` / `AZURE_OPENAI_DEPLOYMENT`: AI provider config. Without valid credentials, falls back to mock provider.
- `STORAGE_PROVIDER`: `local` (default, writes to `BARREL_STORAGE_DIR`) or `azure`/`azure_blob`.
- `BARREL_DEMO_USERNAME` / `BARREL_DEMO_PASSWORD` / `BARREL_REVIEW_TOKEN`: Auth credentials.
- `RULESET_PATH`: Path to TTB rule YAML files (default: `../../rules/ttb` relative to API binary).

## API routes

All `/api/v1/*` routes require auth via `X-BARREL-REVIEW-TOKEN` header or login session.

- `POST /api/v1/labels/analyze-async` - Primary endpoint. Accepts image or .zip, returns job IDs.
- `GET /api/v1/jobs/{id}` - Poll job status.
- `GET /api/v1/reviews` - List all review summaries.
- `GET /api/v1/reviews/{id}` - Review detail with full result.
- `GET /api/v1/reviews/{id}/image` - Original uploaded image.
- `POST /api/v1/reviews/{id}/decision` - Save reviewer decision.
- `POST /api/v1/auth/login` - Get session token.
- `GET /health` - Health check (no auth).

## Storage layout

```
jobs/{job_id}/image.<ext>     # Original uploaded image
jobs/{job_id}/ai_raw.json     # Raw AI response
jobs/{job_id}/result.json     # LabelAnalysisResult
jobs/{job_id}/decision.json   # Reviewer decision
```

## Infrastructure

Azure deployment via OpenTofu + Terragrunt (`infra/`). Resources: Container Apps (API + Web), Blob Storage, Azure OpenAI with GPT-4o deployment, Key Vault, Container Registry.

## Frontend

Single-file React app (`apps/web/src/App.jsx`). No router or state library. Handles login, image upload (drag-drop), job polling, review history table, detail panel with original image, and reviewer decision UI. E2E tests in `apps/web/tests/e2e/`.
