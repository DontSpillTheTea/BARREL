# Rules and Regulatory Breadcrumbs

## Rule Catalog

`rules/ttb/` contains YAML rule definitions loaded by the Go API at startup. Each rule includes a CFR citation, source URL, and explanation displayed in the verification results.

### Coverage by Beverage Type

**Distilled Spirits (27 CFR Part 5)**
| Rule | CFR | Check |
|------|-----|-------|
| Brand Name | § 5.63 | Presence, bigram similarity ≥0.85 |
| Class/Type | § 5.141 | Presence, bigram similarity ≥0.85 |
| Alcohol Content | § 5.65, tolerance per § 5.37 | Numeric ±0.3% ABV |
| Net Contents | § 5.73 | Unit-normalized metric comparison |
| Producer/Bottler | § 5.66 | Fuzzy presence ≥0.80 |
| Country of Origin | § 5.75 | Word overlap ≥50% (imports) |

**Wine (27 CFR Part 4)**
| Rule | CFR | Check |
|------|-----|-------|
| Brand Name | § 4.33 | Presence, bigram similarity ≥0.85 |
| Class/Type | § 4.34 | Presence, bigram similarity ≥0.85 |
| Alcohol Content | § 4.36 | Numeric ±1.0% (>14% ABV) or ±1.5% (≤14% ABV) |
| Net Contents | § 4.37 | Unit-normalized comparison |
| Producer/Bottler | § 4.40 | Fuzzy presence ≥0.80 |
| Country of Origin | § 4.41 | Word overlap ≥50% (imports) |

**Malt Beverages (27 CFR Part 7)**
| Rule | CFR | Check |
|------|-----|-------|
| Brand Name | § 7.23 | Presence, bigram similarity ≥0.85 |
| Class/Type | § 7.24 | Presence, bigram similarity ≥0.85 |
| Alcohol Content | § 7.71 | Numeric ±0.3% ABV |
| Net Contents | § 7.27 | Unit-normalized comparison |

**Health Warning (27 CFR Part 16, all beverages)**
| Rule | CFR | Check |
|------|-----|-------|
| Government Warning | § 16.21 | Verbatim character-by-character match against canonical statutory text |

### Government Warning Canonical Text (27 CFR § 16.21)

```
GOVERNMENT WARNING: (1) According to the Surgeon General, women should not drink
alcoholic beverages during pregnancy because of the risk of birth defects.
(2) Consumption of alcoholic beverages impairs your ability to drive a car or
operate machinery, and may cause health problems.
```

Validation enforces:
- `GOVERNMENT WARNING:` prefix must be ALL CAPS
- Full warning text must match the canonical text character-by-character
- Similarity <60% flagged as suspected AI-generated/corrupted text
- Any deviation produces Mismatch status with expandable character diff in the UI

## Field Comparison Strategies

| Strategy | Fields | How it works |
|----------|--------|--------------|
| Bigram similarity | Brand, Class/Type, Producer | Dice coefficient on character bigrams; case/punctuation tolerant |
| Numeric tolerance | ABV | Parse numeric value, apply beverage-type-specific tolerance from CFR |
| Unit normalization | Net Contents | Convert mL/L/fl oz to common unit, compare with 1% tolerance |
| Verbatim match | Government Warning | Character-by-character comparison against canonical text |
| Word overlap | Country of Origin | Percentage of expected words found in extracted text |

## Advisory Status

Rule checks are **advisory prototypes**. The rule catalog does not constitute final legal determinations. Citations and regulatory breadcrumbs (CFR section links and explanations) appear in verification results to assist reviewers.

## Source URLs

All rules link to the Electronic Code of Federal Regulations (eCFR) at `https://www.ecfr.gov/current/title-27/`.
