# Take-home Alignment

How BARREL addresses the take-home assignment requirements.

## PRD Requirement Alignment

| PRD Requirement | BARREL Implementation |
|----------------|----------------------|
| FR-1: Label upload | Single image + ZIP batch drag-and-drop upload |
| FR-2: Expected data entry | Collapsible form with all 7 fields + beverage type before submission |
| FR-3: Field extraction | Azure OpenAI gpt-4.1-mini with vision + structured JSON extraction |
| FR-4: Brand name (tolerant) | Bigram similarity matching, case/punctuation insensitive (≥0.85 Match) |
| FR-5: Class/type | Bigram similarity matching (≥0.85 Match) |
| FR-6: Alcohol content | Numeric parsing with beverage-type-specific tolerance per 27 CFR |
| FR-7: Net contents | Unit-normalized comparison (mL/L/fl oz conversion) |
| FR-8: Government warning (strict) | Verbatim character-by-character match against 27 CFR § 16.21 canonical text |
| FR-9: Match status labels | Match / Mismatch / Missing on Label / Missing in Application Data / Uncertain |
| FR-10: Overall summary | Status badge + confidence percentage + processing time |
| FR-11: Human review | Original image + side-by-side expected vs extracted + CFR citations |
| FR-12: Batch upload | ZIP processing with per-image async jobs |
| FR-13: Error handling | Clear error messages, 90-second timeout, graceful AI failure |
| FR-14: Test labels | Sample images in `samples/generated/`, Playwright + smoke tests |

## Matching Strictness (PRD Section 11)

| Field | PRD Requirement | Implementation |
|-------|----------------|----------------|
| Brand name | Tolerant (case, punctuation) | Bigram similarity, "STONE'S THROW" = "Stone's Throw" |
| Net contents | Tolerant (formatting) | Unit normalization, "750 mL" = "750ml" |
| Alcohol content | Tolerant (numeric) | ±0.3% spirits, ±1.0%/±1.5% wine per CFR |
| Government warning | Strict (exact wording, ALL CAPS) | Verbatim character match, nonsense detection, expandable diff |

## Additional TTB Coverage Beyond PRD Minimum

| Field | CFR | Notes |
|-------|-----|-------|
| Producer/Bottler | § 5.66, § 4.40 | Fuzzy presence matching |
| Country of Origin | § 5.75, § 4.41 | Word overlap matching for imports |
| Malt ABV | § 7.71 | Numeric tolerance (not in PRD, but required by TTB) |

## Performance (NFR-1)

- PRD target: ~5 seconds per label
- Actual: ~12-13 seconds (gpt-4.1-mini vision inference)
- Optimizations applied: image compression (1MB→120KB), detail:low, max_tokens 1000, 50K TPM
- Documented as known limitation; model inference latency is the bottleneck
- Processing time visible in UI via color-coded badge

## Architecture Justification

- **Go API**: good for HTTP, async orchestration, deterministic validation, structured JSON
- **AI-native vision**: sends images directly to model for field extraction (labels have complex layouts that benefit from vision understanding)
- **Deterministic comparison engine**: does not blindly trust model output; applies field-specific matching rules with regulatory citations
- **Local-first development**: Docker Compose runs the full stack locally; Azure deployment is additive
- **Object/blob storage**: sufficient for review evidence without premature database engineering

## What BARREL Is Not

- Not a final legal determination engine
- Not a COLA system integration
- Not a production-grade federal system (no FedRAMP, no PII compliance, no audit logging)
