# BARREL API


**Note**: `task` is the supported command interface. Docker Compose is wrapped by Task. Normal reviewers should use `task`, not Make or raw Docker Compose.


This directory contains the Go API for BARREL.

* `apps/api` is the Go API target.
* `apps/api/app/` is legacy/bootstrap python scaffold and should not be extended.

## Endpoints

* `GET /health` - API health check
* `POST /api/v1/ocr/extract` - Uploads an image (PNG, JPG, JPEG) to extract raw OCR text and confidence via the configured OCR provider.
* `POST /api/v1/labels/analyze-async`: Primary image analysis endpoint. Reads the file into memory and spins up a background job that either runs the AI-native parser or the debug OCR path. Returns 202 with a job ID.
* `GET /api/v1/jobs/{job_id}`: Poll endpoint for async jobs.
* `POST /api/v1/labels/analyze` - Runs the OCR-only debug path and deterministic analysis in a single pass.

*Note: Rules are loaded from `rules/ttb/`. This is still a review assistant, not a final legal determination system.*

## Local Development
This API is the active BARREL backend. For day-to-day work, run it directly with `task go` and pair it with the frontend via `task web`.

## Current Status & Endpoints

- Local CORS is enabled for the Vite dev dashboard to communicate with the API.
- Web dashboard is at `http://localhost:5173`.
- API is at `http://localhost:8080`.
- Single-image analysis UI exists and calls `POST /api/v1/labels/analyze-async`.
- `ai_native` is the default analysis path.
- Batch upload is implemented and creates one async job per image.
