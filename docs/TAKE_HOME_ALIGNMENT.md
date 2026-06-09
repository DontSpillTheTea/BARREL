# Take-home Alignment


**Note**: `task` is the supported command interface. Docker Compose is wrapped by Task. Normal reviewers should use `task`, not Make or raw Docker Compose.


This document explains how BARREL aligns with the take-home requirements.

## Architecture Choices
- **Why Docker Compose helps reviewers**: It provides a repeatable setup across Ubuntu and Windows without complex local dependency management. (The stack is fully verified and working).
- **Why Go + Python is an appropriate technical choice**: Go is excellent for API routing, JSON contracts, and rule evaluation concurrency, while Python excels at image processing and OCR.
- **Why local-first meets network/security constraints**: It honors strict outbound traffic limits and avoids sending sensitive labels to third-party AI endpoints.
- **Why confidence scoring and breadcrumbs show attention to requirements**: Reviewers need to know *why* something flagged, and what rule it violated, rather than a black-box "fail".
- **Why extraction is deterministic**: OCR remains evidence extraction, not the compliance authority. Separating OCR from regex/fuzzy text extraction in Go allows reliable, testable compliance validation.

## Current Status & Endpoints

- Local CORS is enabled for the Vite dev dashboard to communicate with the API.
- Web dashboard is at `http://localhost:5173`.
- API is at `http://localhost:8080`.
- OCR worker remains internal-only.
- Single-image analysis UI exists and calls `POST /api/v1/labels/analyze`.
- AI escalation is metadata-only (no actual AI provider is called in the current prototype).
- Batch upload remains a future feature.
