# Security

## Product posture

BARREL handles potentially sensitive label images and reviewer decisions. The system is advisory only.

**Required product statement:** BARREL is a review assistant, not a final legal determination system.

## Core policies

- No committed secrets.
- Review token or evaluator login must gate API access.
- Secrets stay in environment variables or Azure-managed secret stores.
- Blob containers must not be public.
- Do not log raw secrets, API keys, or sensitive auth material.
- Do not log full raw label contents indiscriminately in application logs.
- No arbitrary URL fetching from user-provided inputs.

## AI provider handling

- Images and parsed metadata are sent to the configured AI provider when `ai_native` is used.
- No model provider call should happen without an explicit configured provider.
- Preferred hosted provider is Azure OpenAI / Azure AI Foundry if quota is available.
- Direct OpenAI API may be used as a temporary fallback if Azure OpenAI is blocked by subscription quota or region policy.
- `azure_vision_ocr` may be used for debug/baseline evidence but is not the primary parser.

## Data handling

Preferred review evidence storage is object/blob storage.

Azure implementation:

- Azure Blob Storage for stored evidence
- non-public containers
- paths such as:
  - `jobs/{job_id}/image.<ext>`
  - `jobs/{job_id}/ai_raw.json`
  - `jobs/{job_id}/result.json`
  - `jobs/{job_id}/decision.json`

Stored artifacts may contain business-sensitive label images and parsed metadata. Treat them accordingly.

## Access control

- No public signup.
- HTTPS-only inbound access in Azure.
- Demo evaluator credentials are a take-home tradeoff and must remain server-controlled.
- CORS should only allow known frontend origins.

## Upload restrictions

- enforce upload size limits
- restrict MIME types and extensions to supported image and zip formats
- reject unsupported or malformed uploads early

## Operational expectations

- Use local validation first and Azure smoke second.
- Do not claim deployment success without smoke tests.
- Do not expose provider secrets in scripts, docs, screenshots, or commits.
