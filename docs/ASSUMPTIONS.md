# Assumptions

## Development Environment
- Reviewers may use Windows; Docker Desktop + WSL2 is recommended
- Outbound traffic to Azure OpenAI endpoints must be allowed
- Local development requires Docker and Node.js

## AI Provider
- Azure OpenAI with gpt-4.1-mini (vision) is the primary AI provider
- The model deployment is in Sweden Central due to quota availability
- Processing takes ~12-13 seconds per label (vision inference latency)
- If Azure OpenAI is unavailable, the mock provider returns structured placeholder data

## Test Data
- AI-generated label images are used as test samples (fictional brands)
- Generated samples are NOT represented as approved COLAs
- Test samples include distilled spirits scenarios matching the PRD example

## Regulatory
- Rule checks are advisory prototypes based on 27 CFR Parts 4, 5, 7, and 16
- The rule catalog covers core labeling fields but is not exhaustive
- Government warning validation uses the exact canonical text from 27 CFR § 16.21
- ABV tolerances follow published CFR specifications per beverage type
- All matching logic is transparent and citable (CFR references in every field check)

## Deployment
- Currently local-only via Docker Compose
- Azure Container Apps infrastructure is defined in OpenTofu but not yet deployed
- The `Microsoft.App` resource provider needs registration on the Azure subscription
- When deployed, the app will use Azure Blob Storage instead of local filesystem

## Prototype Scope
- This is a standalone prototype, not a COLA system integration
- No production-grade federal security, PII compliance, or audit logging
- No persistent storage of uploaded labels beyond the current session (local filesystem, cleared on container restart)
- Single-region deployment (no HA/failover)
