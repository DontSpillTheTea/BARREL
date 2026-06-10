# Architecture


**Note**: `task` is the supported command interface. Docker Compose is wrapped by Task. Normal reviewers should use `task`, not Make or raw Docker Compose.


## Topology
BARREL follows a decoupled architecture using Docker Compose:
- **web**: React/Vite frontend
- **api**: Go API responsible for orchestration, rule loading, text extraction validation, and API endpoints
- **ocr-worker**: Python worker responsible for OCR and image-processing

## Target Architecture

1. **Frontend (React/Vite)**
   - Browser-based review interface
   - Submits images via multipart form POST (`/api/v1/labels/analyze-async`)
   - Polls async job endpoints (`/api/v1/jobs/{job_id}`) so UI doesn't timeout on slow CPU inference

2. **Backend API (Go)**
   - Upload handler, validation, routing
   - Async Job manager (in-memory, no DB required for POC)
   - Rule execution engine (reads local `.yaml` config)
   - Communicates to OCR worker over local HTTP

3. **OCR Worker (Python)**
   - `FastAPI` service running `paddleocr` or `tesseract`
   - Image pre-processing (downscale, grayscale)
   - Primary data extraction point
   - Pre-warms deep learning models in the background on startup

4. **External/AI (Future)**
   - Only used as fallback when OCR confidence is low
   - Floci-AZ / Azure OpenAI (Strictly metadata-only right now)

## Responsibilities
- **Go API**: Handles uploads (`POST /api/v1/labels/analyze`), coordinates with the OCR worker, evaluates rules deterministically, performs regex extraction, and produces rule breadcrumbs. It exposes a text-only interface (`POST /api/v1/labels/analyze-text`) to decouple compliance testing from OCR processing.
### Python OCR Worker (`apps/ocr-worker`)
A FastAPI service that handles the raw image processing.
- **Accuracy-First**: It uses PaddleOCR as the default, deep OCR provider. This is initialized and warmed up on container startup in a background thread, ensuring readiness requests return immediately.
- **Fallback**: Tesseract is kept as a fast fallback or diagnostic provider. Fast fallback OCR is not intended as the default evidence path.
- Returns plain extracted text, bounding boxes, mean confidence, and provider metadata.
- Pre-processes images using PIL and OpenCV.
- Evaluates raw image quality before extraction.
- **Top-level rules catalog**: Rules live in `rules/ttb/` and are loaded by the Go API on startup.

## Runtime
- **Docker Compose**: The preferred demo/reviewer runtime for repeatability. (Verified & Working)
- **go-task**: The primary cross-platform command runner.
- **Local-first**: The system defaults to local-first / no-outbound-AI architecture to meet strict outbound network constraints.

## Endpoints
- Web health dashboard: http://localhost:5173
- API: http://localhost:8080
- OCR worker: internal only

## Storage & Review
- **Storage Abstraction**: The Go API includes a `storage` package that can write review history locally (`data/reviews`) or to Azure Blob Storage.
- **Review History**: Stores the original uploaded label, expected JSON, extracted OCR data, confidence scoring, and human reviewer decisions.
- **Security**: The API uses a token-based middleware (`BARREL_REVIEW_TOKEN`) to secure endpoints.

## Azure Target Architecture (Phase 10+)
To improve speed and quality while remaining cost-effective, BARREL has pivoted to support Azure deployment:
- **Infrastructure**: Terraform/Terragrunt handles provisioning (`infra/terraform/dev`).
- **Compute**: Azure Container Apps run the Go API and React Web frontend.
- **OCR**: Azure Computer Vision is integrated as the primary, high-speed, high-quality OCR provider via the Go API `ocr/providers` package.
- **Local Fallback**: PaddleOCR remains as the accurate offline/local fallback.
- **State**: Review histories and images are stored in Azure Blob Storage.

## Current Status & Endpoints

- Local CORS is enabled for the Vite dev dashboard to communicate with the API.
- Web dashboard is at `http://localhost:5173`.
- API is at `http://localhost:8080`.
- OCR worker remains internal-only.
- Single-image analysis UI exists and calls `POST /api/v1/labels/analyze-async` to avoid browser timeouts.
- Reviewer workspace UI is available, including history and decision controls.
- Azure deployment tasks (`azure:infra:apply`, `azure:deploy`) are integrated via go-task.
