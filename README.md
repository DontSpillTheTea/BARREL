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

BARREL uses `go-task` as the primary cross-platform command runner.

Preferred demo path:
```bash
task dev
```

Compatibility fallback:
```bash
make dev
```

Raw fallback:
```bash
docker compose up --build
```
