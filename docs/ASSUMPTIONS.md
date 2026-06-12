# Assumptions

## Environment
- Reviewers may use Windows; Docker Desktop + WSL2 recommended for local development
- Outbound network access to Azure services (OpenAI, Vision OCR, Blob Storage) is required
- Local development requires Docker, Task, and Azure OpenAI credentials in `.env`

## AI/OCR Extraction
- Azure OpenAI gpt-4.1-mini provides vision-capable field extraction
- Azure Vision OCR provides fast text extraction for the tiered pipeline
- OCR/AI extraction accuracy depends on image quality; poor images produce Uncertain status
- The tiered pipeline escalates to AI vision when OCR results are uncertain or mismatched
- Producer/bottler and country of origin are optional fields and don't trigger escalation

## Regulatory
- Rule checks are advisory prototypes based on 27 CFR Parts 4, 5, 7, and 16
- Government warning validation uses the exact canonical text from 27 CFR § 16.21
- ABV tolerances follow published CFR specifications per beverage type
- The rule catalog covers core labeling fields but is not exhaustive
- Human review is always required; BARREL does not make final legal determinations

## Test Data
- AI-generated label images are used as test samples (fictional brands)
- Generated samples are NOT represented as approved COLAs

## Performance
- OCR fast path targets ~3-5s for clear labels
- Escalated path (OCR + AI vision) takes ~12-18s
- Performance varies by image quality, label complexity, and Azure service latency

## Deployment
- Azure Container Apps in Sweden Central (matches Azure OpenAI region)
- Azure Blob Storage for durable review evidence
- Evidence persists across container restarts
- Infrastructure managed by OpenTofu + Terragrunt
