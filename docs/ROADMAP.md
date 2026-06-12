# Roadmap

## Completed

- Azure Container Apps deployment in East US 2 (Virginia) with public evaluator URL
- Azure Blob Storage + Table Storage for persistent review evidence
- Azure Vision OCR for fast text extraction (~0.5s)
- Tiered analysis pipeline: OCR → text parser → validators → AI vision escalation
- AI-native parser with Azure OpenAI gpt-4.1-mini (vision)
- Seven-field comparison engine with per-field regulatory matching strategies
- PRD-aligned status labels: Match / Mismatch / Missing on Label / Missing in Application Data / Uncertain
- Expected field entry form before submission
- ABV numeric tolerance per beverage type (27 CFR § 5.37, § 4.36, § 7.71)
- Net contents unit-normalized comparison (mL/L/fl oz)
- Strict government warning verbatim validation with hallucination detection
- Government warning prompt hardening (legibility, body confidence, prefix caps, bold detection)
- Bold detection for GOVERNMENT WARNING prefix (27 CFR § 16.22)
- Container size validation per beverage type (27 CFR § 5.71, § 4.72)
- Producer/bottler and country of origin fuzzy validation with CFR citations
- Single-image and ZIP batch upload with async processing
- Review History with Blob + Table Storage evidence rehydration
- Approve/Reject reviewer decision persistence
- CSV export of verification results
- TypeScript frontend with full type annotations
- Per-field similarity scores, AI confidence, and source indicators
- Processing time measurement with timing breakdown
- Image compression pipeline (PNG→JPEG, max 768px)
- Provider path badges (OCR Fast Path / OCR + AI Vision / AI Native)
- Escalation reason display
- Infrastructure documentation for standup from scratch
- Go unit tests, Python smoke tests, Playwright E2E tests

## Known Limitations

- Processing time ~3-5s (OCR fast path) to ~12-18s (escalated) — PRD 5s target met for fast path only
- No COLA system integration
- No production-grade FedRAMP, audit logging, or enterprise RBAC

## Future Work

- Production security hardening
