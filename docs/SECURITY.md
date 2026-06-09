# Security


**Note**: `task` is the supported command interface. Docker Compose is wrapped by Task. Normal reviewers should use `task`, not Make or raw Docker Compose.


## General Policies
- **No committed secrets**: Do not commit keys, `.env` files, or production credentials.
- **No required outbound AI**: System operates entirely locally by default.
- **No arbitrary URL fetching**: To prevent SSRF, external URLs are not retrieved based on user input.
- **No permanent upload persistence by default**: Uploads are processed and discarded.
- **Azure Key Vault**: Optional future path for secret management in Azure environments, but not required locally.

## Container Security
- **Container boundaries**: Strict separation between web, api, and ocr-worker. (Verified working via Docker Compose)
- **Internal OCR Worker**: The OCR worker is not exposed externally in the final topology.
- **Non-root containers**: Running containers as non-root is a goal.

## Upload Restrictions
- **Upload size limits**: Hard limits on file sizes (e.g. 25MB).
- **MIME/extension allowlists**: Only specific image/zip formats are permitted.

## AI Escalation
AI escalation is metadata-only for now. No external AI calls are made. Floci-AZ remains future/optional.

## Current Status & Endpoints

- Local CORS is enabled for the Vite dev dashboard to communicate with the API.
- Web dashboard is at `http://localhost:5173`.
- API is at `http://localhost:8080`.
- OCR worker remains internal-only.
- Single-image analysis UI exists and calls `POST /api/v1/labels/analyze`.
- AI escalation is metadata-only (no actual AI provider is called in the current prototype).
- Batch upload remains a future feature.
