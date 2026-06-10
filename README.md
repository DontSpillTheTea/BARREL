# BARREL: Beverage Alcohol Review & Regulatory Evidence Logger

A local-first OCR and compliance-assist prototype for reviewing alcohol beverage labels with confidence scoring and regulatory breadcrumbs.

## What this is
BARREL is a review assistant, not a final legal determination system. The prototype is designed to run without required outbound AI calls.

Current implementation status is scaffold/POC, not full app.

## Architecture Highlights
- Go API + Python OCR worker + React/Vite web.
- Docker Compose is used for repeatability across Ubuntu and Windows.
- Windows recommendation is Docker Desktop + WSL2.
- Local-first architecture; Azure Key Vault is an optional future integration, not required locally.
- Two local OCR paths: a Fast path (Tesseract, default) and an optional Deep path (PaddleOCR, enabled via config).

## Getting Started

BARREL uses `go-task` as the primary cross-platform command runner. Task wraps Docker Compose internally.
Normal reviewers should run task commands, not task or docker compose directly.

### Requirements
- Docker Desktop or Docker Engine
- Docker Compose plugin
- go-task
- Azure CLI
- OpenTofu
- Terragrunt

### Quick Start

**Note**: `task` is the supported command interface. Docker Compose is wrapped by Task. Normal reviewers should use `task`, not Make or raw Docker Compose.


```bash
task check-env
task up
task smoke
task down
```

### Verification / Troubleshooting

The Docker Compose setup has been tested and verified as the core demo path, wrapped by Task.
If you need to verify services manually:

To verify health:
```bash
curl -sS http://localhost:8080/health
curl -sS http://localhost:8080/health/ocr-worker
```

To test deterministic text analysis:
```bash
curl -sS -X POST http://localhost:8080/api/v1/labels/analyze-text \
  -H "Content-Type: application/json" \
  -d '{"beverage_type":"distilled_spirits","text":"GOVERNMENT WARNING: test","expected_fields":{"government_warning_present":true}}'
```

To test image analysis (assuming the stack is up):
```bash
curl -sS -X POST http://localhost:8080/api/v1/labels/analyze \
  -F "file=@samples/generated/good/good_01_distilled_spirits_clean_front.png" \
  -F "beverage_type=distilled_spirits" \
  -F 'expected_json={"brand_name":"OLD TOM DISTILLERY","class_type":"Kentucky Straight Bourbon Whiskey","alcohol_content":"45% Alc./Vol. (90 Proof)","net_contents":"750 mL","government_warning_present":true}'
```
*Note: OCR has limitations (e.g., misreading small fonts), which is why BARREL provides confidence scores and flags uncertain results for human review.*

## Current Status & Endpoints

- Local CORS is enabled for the Vite dev dashboard to communicate with the API.
- Web dashboard is at `http://localhost:5173`.
- API is at `http://localhost:8080`.
- OCR worker remains internal-only.
- `/health` means the process is alive. `/ready` means the OCR provider is ready for requests.
- AI escalation is metadata-only (no actual AI provider is called in the current prototype).
- If accurate OCR is slow, that is surfaced honestly.
- Batch upload remains a future feature.

## Key Technical Rules

* **Local-First Context:** Do not send labels or text to cloud AI providers unless explicitly allowed in the active feature phase.
* **Accuracy-First OCR:** PaddleOCR is the default deep OCR provider. It is warmed and cached by the OCR worker. Tesseract is used only as an explicit fallback.
* **Async Analysis:** Browser analysis uses the `POST /api/v1/labels/analyze-async` job endpoint. Initial upload response is fast, and the frontend polls for completion. This prevents browser timeouts on slow CPU hardware.
* **No Magic Updates:** Changes to project goals or architecture must be updated in `docs/` or `AGENTS.md`.

## Azure Deployment Costs

When deployed to Azure, the BARREL architecture attempts to keep costs low using serverless, consumption-based, and free-tier resources where applicable:
* **Frontend**: Azure Container Apps (Consumption) or Azure Static Web Apps (Free tier) where possible.
* **API Engine**: Azure Container Apps (Consumption plan, with free grant if available).
* **Azure AI Vision OCR**: Pay-per-use, but free-tier (F0) is configurable for demo purposes. See Azure pricing for updates.
* **State / Storage**: Azure Blob Storage incurs minor costs for data retention.

Pricing changes over time. Check the [Azure Pricing Calculator](https://azure.microsoft.com/en-us/pricing/calculator/) for the latest metrics, and configure manual Azure budget alerts in your subscription portal to prevent overages.

## Scripts & Validation

* `task smoke` - Validates full component topology and API/OCR communication.
* `task smoke:fast-api` - Test text-only API performance.
* `task smoke:async-analysis` - Validates the full async image OCR flow using the accurate local provider.
* `task azure:smoke` - Verifies the deployed Azure endpoints and checks `azure_vision` OCR extraction.
