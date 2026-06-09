# Architecture

## Topology
BARREL follows a decoupled architecture using Docker Compose:
- **web**: React/Vite frontend
- **api**: Go API responsible for orchestration, rule loading, text extraction validation, and API endpoints
- **ocr-worker**: Python worker responsible for OCR and image-processing

## Responsibilities
- **Go API**: Handles uploads (`POST /api/v1/ocr/extract`), coordinates with the OCR worker, evaluates rules deterministically, and serves reports.
- **Python OCR worker**: Handles heavy image processing and extracts text locally using Tesseract (`POST /ocr/extract`).
- **Top-level rules catalog**: Rules live in `rules/ttb/` and are loaded by the Go API.

## Runtime
- **Docker Compose**: The preferred demo/reviewer runtime for repeatability.
- **go-task**: The primary cross-platform command runner.
- **Local-first**: The system defaults to local-first / no-outbound-AI architecture to meet strict outbound network constraints.
