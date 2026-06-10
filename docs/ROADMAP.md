# Roadmap

**Status:** Phases 1–3 complete.



**Note**: `task` is the supported command interface. Docker Compose is wrapped by Task. Normal reviewers should use `task`, not Make or raw Docker Compose.


## Phase 0: Scaffold
1. sample fixture matrix (Completed)
2. Taskfile + Docker Compose scaffold (Completed & Verified)
3. Go API health endpoint (Completed)
4. Python OCR worker health endpoint (Completed)

## Phase 1: Core OCR Pipeline
5. API-to-worker OCR call (Completed & Verified)
6. rule loading (Completed & Verified)
7. deterministic extraction (Completed & Verified)

## Phase 2: Analysis & UI (Completed)
8. single-image analysis (Completed - Async)
9. frontend upload UI (Completed)
10. Reviewer History and Storage (Completed)

## Phase 3: Azure Cloud Deployment (In Progress)
11. Azure Provider integration (Vision API) (Completed)
12. Azure Infrastructure (OpenTofu + Terragrunt / Container Apps) (Scaffolded)
13. Security tokens (Completed)
14. Batch/zip analysis (Future)

## Current Status & Endpoints

- Web dashboard is at `http://localhost:5173`.
- API is at `http://localhost:8080`.
- OCR worker remains internal-only.
- Single-image analysis UI exists and calls `POST /api/v1/labels/analyze-async`.
- Reviewer workspace with history and decision states is integrated.
- AI escalation is metadata-only (no actual AI provider is called in the current prototype).
- Batch upload remains a future feature.
