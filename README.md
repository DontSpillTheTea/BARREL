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

## Getting Started

BARREL uses `go-task` as the primary cross-platform command runner. Task wraps Docker Compose internally.
Normal reviewers should run task commands, not task or docker compose directly.

### Requirements
- Docker Desktop or Docker Engine
- Docker Compose plugin
- go-task

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
- Single-image analysis UI exists and calls `POST /api/v1/labels/analyze`.
- AI escalation is metadata-only (no actual AI provider is called in the current prototype).
- Batch upload remains a future feature.
