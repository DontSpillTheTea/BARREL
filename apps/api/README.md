# BARREL API


**Note**: `task` is the supported command interface. Docker Compose is wrapped by Task. Normal reviewers should use `task`, not Make or raw Docker Compose.


This directory contains the Go API for BARREL.

* `apps/api` is the Go API target.
* `apps/api/app/` is legacy/bootstrap python scaffold and should not be extended.

## Endpoints

* `GET /health` - API health check
* `GET /health/ocr-worker` - Proxies health check to the OCR worker
* `POST /api/v1/ocr/extract` - Uploads an image (PNG, JPG, JPEG) to extract raw OCR text, image quality, and OCR confidence via the worker.
* `POST /api/v1/labels/analyze-text` - Pure text endpoint to perform deterministic extraction and compliance logic without needing OCR execution.
* `POST /api/v1/labels/analyze` - Combines OCR extraction and text analysis into a single pass to produce full regulatory breadcrumbs and confidence scoring.

*Note: Rules are loaded from `rules/ttb/`. This is still a review assistant, not a final legal determination system.*

## Docker Compose
This API is designed to run within the Docker Compose stack alongside `ocr-worker`. The stack has been verified to work with Docker Compose.

## Current Status & Endpoints

- Local CORS is enabled for the Vite dev dashboard to communicate with the API.
- Web dashboard is at `http://localhost:5173`.
- API is at `http://localhost:8080`.
- OCR worker remains internal-only.
- Single-image analysis UI exists and calls `POST /api/v1/labels/analyze`.
- AI escalation is metadata-only (no actual AI provider is called in the current prototype).
- Batch upload remains a future feature.
