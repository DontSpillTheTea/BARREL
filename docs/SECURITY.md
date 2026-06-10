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

## Azure Target Architecture Posture (Phase 10+)
When deployed to Azure via OpenTofu/Terragrunt:
- **Public HTTPS only**: The API and Web UI are only accessible via HTTPS.
- **Authentication**: Evaluator login/token is provided as a demo tradeoff to protect the API.
- **No public signup**: The authentication system uses a single configured demo account to prevent public signups.
- **Secrets Management**: Azure Vision keys and `BARREL_REVIEW_TOKEN` remain server-side and are injected into the Container App as secrets from OpenTofu or Key Vault.
- **Azure Vision Access**: The backend API uses Azure Vision to perform OCR. The backend sends the image directly; Azure Vision receives uploaded label images.
- **Data Retention**: Review history, including images, expected JSON, OCR extractions, and Reviewer Decisions, are saved to Azure Blob Storage to form the evidence logger.
- **Business Sensitivity**: Raw OCR/image artifacts stored in Blob Storage may contain business-sensitive information.
- **CORS**: A CORS allowlist must restrict cross-origin requests to only known domains (e.g. the frontend app domain).
- **No general LLM calls yet**: We are using Azure Computer Vision OCR. General AI/LLM escalation remains disabled.
- **Cleanup**: The `task azure:infra:destroy` command exists to clean up all deployed Azure resources and avoid rogue consumption.
