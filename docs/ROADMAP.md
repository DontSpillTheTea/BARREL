# Roadmap

## Completed

- Azure Container Apps deployment with public evaluator URL
- Azure Blob Storage for persistent review evidence
- Azure Vision OCR for fast text extraction (~0.5s)
- Tiered analysis pipeline: OCR → text parser → validators → AI vision escalation
- AI-native parser with Azure OpenAI gpt-4.1-mini (vision)
- Seven-field comparison engine with per-field regulatory matching strategies
- PRD-aligned status labels: Match / Mismatch / Missing on Label / Missing in Application Data / Uncertain
- Expected field entry form before submission
- ABV numeric tolerance per beverage type (27 CFR § 5.37, § 4.36, § 7.71)
- Net contents unit-normalized comparison (mL/L/fl oz)
- Strict government warning verbatim validation with hallucination detection
- Government warning prompt hardening (legibility, body confidence, prefix caps)
- Producer/bottler and country of origin fuzzy validation with CFR citations
- Single-image and ZIP batch upload with async processing
- Review History with Azure Blob-backed evidence rehydration
- Approve/Reject reviewer decision persistence
- CSV export of verification results
- Per-field similarity scores, AI confidence, and source indicators
- Processing time measurement with timing breakdown
- Image compression pipeline (PNG→JPEG, max 768px)
- Provider path badges (OCR Fast Path / OCR + AI Vision / AI Native)
- Escalation reason display
- Go unit tests, Python smoke tests, Playwright E2E tests

## Known Limitations

- Processing time ~3-5s (OCR fast path) to ~12-18s (escalated) — PRD 5s target met for fast path only
- Bold detection for `GOVERNMENT WARNING:` prefix not implemented
- No bounding box overlays on label images
- No CSV upload for per-image expected data in ZIP batches
- No COLA system integration
- No production-grade FedRAMP, audit logging, or enterprise RBAC

## Future Work

- Bounding box overlays on extracted fields
- CSV expected-data import for ZIP batches
- Bold text detection for government warning prefix
- COLA integration plan documentation
- Container size validation per beverage type (27 CFR § 5.71, § 4.72)
- Confidence calibration per beverage type
- Production security hardening
