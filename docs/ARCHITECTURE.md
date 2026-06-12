# BARREL Architecture

**BARREL is a review assistant, not a final legal determination system.**

## Topology

```text
Browser (React 19 / Vite 8)
  → Azure Container Apps (barrel-web, port 5173)
  → Azure Container Apps (barrel-api, port 8080)
    → Tiered Analysis Pipeline:
       1. Azure Vision OCR (~0.5s, text extraction)
       2. Text Parser via Azure OpenAI gpt-4.1-mini (~2-4s, text-only classification)
       3. Deterministic Validators (<1ms, field-specific regulatory matching)
       4. [If needed] AI Vision via Azure OpenAI gpt-4.1-mini (~10-15s, image input)
    → Azure Blob Storage (review evidence)
  → Review History / Detail Viewer
```

## Analysis Pipeline

### Tiered Provider (default)

1. **OCR Stage**: Azure Vision OCR extracts text with word-level confidence
2. **Text Parse Stage**: gpt-4.1-mini classifies OCR text into structured fields (no image tokens — speed win)
3. **Validation Stage**: Field-specific deterministic comparison (see table below)
4. **Escalation Gate**: Checks for missing required fields, mismatches, low confidence, gov warning issues
5. **AI Vision Stage** (if escalated): gpt-4.1-mini with image input for full vision extraction

Producer/bottler and country of origin are optional and don't trigger escalation.

### Field Comparison Engine

| Field | Strategy | Threshold | CFR |
|-------|----------|-----------|-----|
| Brand Name | Bigram similarity | ≥0.85 Match | § 5.63 / § 4.33 / § 7.23 |
| Class/Type | Bigram similarity | ≥0.85 Match | § 5.141 / § 4.34 / § 7.24 |
| Alcohol Content | Numeric tolerance | ±0.3% spirits, ±1.0%/±1.5% wine | § 5.37 / § 4.36 / § 7.71 |
| Net Contents | Unit normalization | 1% tolerance | § 5.73 / § 4.37 / § 7.27 |
| Government Warning | Verbatim char match | Exact (§ 16.21) | § 16.21 |
| Producer/Bottler | Bigram similarity | ≥0.80 Match | § 5.66 / § 4.40 |
| Country of Origin | Word overlap | ≥50% Match | § 5.75 / § 4.41 |

## Storage

Azure Blob Storage (container: `jobs`, account: `barrelsadev`):
```text
{job_id}/image.png       # Original uploaded label
{job_id}/result.json     # Analysis result with field comparisons
{job_id}/decision.json   # Reviewer approve/reject decision
```

Local development uses filesystem storage at `/tmp/barrel_storage_local`.

## Infrastructure (Azure)

All resources in `barrel-ai-rg` resource group, Sweden Central:
- Container Apps: barrel-api (0.5 CPU, 1GB) + barrel-web (0.25 CPU, 0.5GB)
- Container Registry: barrelacrdev.azurecr.io
- Azure OpenAI: barrel-openai-sweden (gpt-4.1-mini, 50K TPM)
- Azure Vision OCR: barrel-vision-dev (F0)
- Blob Storage: barrelsadev
- Key Vault: barrel-kv-dev
- Log Analytics: barrel-law
- Managed by OpenTofu + Terragrunt (`infra/`)
