# BARREL Web Dashboard


**Note**: `task` is the supported command interface. Docker Compose is wrapped by Task. Normal reviewers should use `task`, not Make or raw Docker Compose.


This directory contains the React/Vite frontend for BARREL.

## Overview
Currently, it serves as a simple health dashboard and POC summary page.

## Configuration
The frontend communicates with the API via the `VITE_API_BASE_URL` environment variable.

## Development
Run `npm run dev` to start the development server locally.
Alternatively, the Docker Compose stack (managed via `task up`) will run the web service at `http://localhost:5173`.

## Current Status & Endpoints

- Local CORS is enabled for the Vite dev dashboard to communicate with the API.
- Web dashboard is at `http://localhost:5173`.
- API is at `http://localhost:8080`.
- OCR worker remains internal-only.
- Single-image analysis UI uses an async polling mechanism with `POST /api/v1/labels/analyze-async` and `GET /api/v1/jobs/{job_id}` to prevent browser timeouts on slow CPU inference.
- AI escalation is metadata-only (no actual AI provider is called in the current prototype).
- Batch upload remains a future feature.
