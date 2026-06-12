# Take-home Alignment

How BARREL addresses the official PRD requirements (`/.agents/prd_official.md`).

## P0: Must Have

| # | Requirement | Status |
|---|------------|--------|
| 1 | Single-label upload | Complete |
| 2 | Manual entry of expected application fields | Complete (collapsible form, 7 fields + beverage type) |
| 3 | Extraction of label text/field candidates | Complete (Azure Vision OCR + AI vision) |
| 4 | Verification for brand, class/type, ABV, net contents, gov warning | Complete (7 fields with CFR citations) |
| 5 | Field-level match/mismatch/missing/uncertain statuses | Complete |
| 6 | Side-by-side display of expected and extracted values | Complete |
| 7 | Strict government warning check | Complete (verbatim char match, hallucination detection, legibility awareness) |
| 8 | Tolerant brand-name comparison | Complete (bigram similarity, "STONE'S THROW" = "Stone's Throw") |
| 9 | Clear error handling | Complete |
| 10 | README with setup, run, assumptions, limitations | Complete |
| 11 | Deployed accessible prototype | Complete (Azure Container Apps URL) |

## P1: Should Have

| # | Requirement | Status |
|---|------------|--------|
| 1 | Batch upload and batch summary | Complete (ZIP) |
| 2 | Producer/bottler address | Complete (fuzzy matching, § 5.66 / § 4.40) |
| 3 | Country of origin | Complete (word overlap, § 5.75 / § 4.41) |
| 4 | Beverage type selection | Complete (dropdown: spirits/wine/malt) |
| 5 | Confidence indicators | Complete (similarity scores, AI confidence, timing) |
| 6 | Better handling of imperfect images | Complete (quality flags, legibility detection, Uncertain status) |
| 7 | Performance measurement | Complete (per-stage timing breakdown in UI) |
| 8 | Documented outbound network dependencies | Complete (README lists all Azure services) |

## P2: Nice to Have

| # | Requirement | Status |
|---|------------|--------|
| 1 | Bounding boxes on label text | Not implemented |
| 2 | CSV upload for batch expected data | Not implemented |
| 3 | Exportable verification report | Complete (CSV export) |
| 4 | Image preprocessing | Complete (compression, quality flags) |
| 5 | Bold detection for GOVERNMENT WARNING | Not implemented (documented limitation) |
| 6 | Future COLA integration plan | Documented as future work |

## Matching Strictness (PRD Section 11)

| Field | PRD Requirement | Implementation |
|-------|----------------|----------------|
| Brand name | Tolerant | Bigram similarity ≥0.85, case/punctuation insensitive |
| Net contents | Tolerant | Unit normalization (750 mL = 750ml) |
| Alcohol content | Tolerant (numeric) | ±0.3% spirits, ±1.0%/±1.5% wine per 27 CFR |
| Government warning | Strict | Verbatim character match, prefix must be ALL CAPS, body hallucination detection |

## Architecture Justification

- **Go API**: HTTP handling, async job orchestration, deterministic validation
- **Tiered OCR→AI pipeline**: Fast OCR for clear labels (~3-5s), AI vision only when needed
- **Azure Blob Storage**: Lightweight evidence persistence without premature database engineering
- **Field-specific validators**: Each field has its own matching strategy per regulatory requirements
- **Advisory-only posture**: All results are for human review, not automated approval
