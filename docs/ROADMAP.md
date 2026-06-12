# Roadmap

## Completed

### Core verification workflow
- AI-native parser with Azure OpenAI gpt-4.1-mini (vision, Sweden Central, 50K TPM)
- Field-specific comparison engine covering 7 fields with per-field matching strategies
- PRD-aligned status labels: Match / Mismatch / Missing on Label / Missing in Application Data / Uncertain
- Expected field entry form (brand, class/type, ABV, net contents, producer, country, beverage type, gov warning)
- Strict government warning validation: verbatim character-by-character match against 27 CFR § 16.21 canonical text
- Government warning nonsense/AI-text detection (similarity <60%)
- ABV numeric tolerance per beverage type: ±0.3% spirits (§ 5.37), ±1.0%/±1.5% wine (§ 4.36), ±0.3% malt (§ 7.71)
- Net contents unit-normalized comparison (mL/L/fl oz conversion)
- Producer/bottler and country of origin fuzzy validation with CFR citations
- Single-image and ZIP batch upload with async job processing
- Review History table with clickable rows restoring full result detail
- Approve/Reject reviewer decision workflow
- Image compression pipeline (PNG→JPEG, max 768px)

### UI and quality features
- Top-row layout: upload (1/3) + image viewbox (2/3), full-width field table below
- Per-field similarity scores and AI confidence displayed
- Processing time measurement badge (color-coded)
- Expandable government warning character diff view
- Image quality flag warnings from AI response
- CSV export of verification results
- Low AI confidence (<70%) forces Uncertain status

### Testing and validation
- Go unit tests: ABV tolerance, net contents normalization, bigram similarity, analysis, gov warning
- Python smoke test: end-to-end AI-native flow (`task smoke:local-ai`)
- Playwright E2E tests: single image + batch upload (`task playwright:local`)
- Docker Compose local stack verified working

## In Progress

### Azure deployment
- OpenTofu infrastructure defined but not yet applied
- Subscription needs `Microsoft.App` provider registration
- Container Apps, ACR, Blob Storage, Key Vault all pending deployment
- Currently running locally via Docker Compose only

## Known Limitations

- Processing time ~12-13s per label (gpt-4.1-mini vision inference latency)
- PRD 5-second target documented as aspirational; image compression and max_tokens optimizations applied
- No deployed public URL yet (evaluators must run locally)
- Local filesystem storage instead of Azure Blob Storage

## Future Work

- Deploy to Azure Container Apps and provide public URL
- Switch storage from local filesystem to Azure Blob Storage
- Explore text-only LLM pipeline (OCR→text→classify, no image tokens) for sub-5s processing
- Container size validation per beverage type (27 CFR § 5.71, § 4.72)
- Confidence calibration per beverage type
