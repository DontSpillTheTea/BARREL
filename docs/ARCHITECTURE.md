# Architecture


**Note**: `task` is the supported command interface. Docker Compose is wrapped by Task. Normal reviewers should use `task`, not Make or raw Docker Compose.


## Topology
BARREL follows a decoupled architecture using Docker Compose:
- **web**: React/Vite frontend
- **api**: Go API responsible for orchestration, rule loading, text extraction validation, and API endpoints
- **ocr-worker**: Python worker responsible for OCR and image-processing

## Responsibilities
- **Go API**: Handles uploads (`POST /api/v1/labels/analyze`), coordinates with the OCR worker, evaluates rules deterministically, performs regex extraction, and produces rule breadcrumbs. It exposes a text-only interface (`POST /api/v1/labels/analyze-text`) to decouple compliance testing from OCR processing.
- **Python OCR worker**: Handles heavy image processing and extracts text locally using Tesseract (`POST /ocr/extract`).
- **Top-level rules catalog**: Rules live in `rules/ttb/` and are loaded by the Go API on startup.

## Runtime
- **Docker Compose**: The preferred demo/reviewer runtime for repeatability. (Verified & Working)
- **go-task**: The primary cross-platform command runner.
- **Local-first**: The system defaults to local-first / no-outbound-AI architecture to meet strict outbound network constraints.

## Endpoints
- Web health dashboard: http://localhost:5173
- API: http://localhost:8080
- OCR worker: internal only

## Current Status & Endpoints

- Local CORS is enabled for the Vite dev dashboard to communicate with the API.
- Web dashboard is at `http://localhost:5173`.
- API is at `http://localhost:8080`.
- OCR worker remains internal-only.
- Single-image analysis UI exists and calls `POST /api/v1/labels/analyze`.
- AI escalation is metadata-only (no actual AI provider is called in the current prototype).
- Batch upload remains a future feature.
