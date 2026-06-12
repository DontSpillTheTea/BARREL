# Security

**BARREL is a review assistant, not a final legal determination system.** This is a prototype; production deployment would require additional security review.

## Secrets
- No secrets committed to the repository
- Local: `.env` file (gitignored)
- Azure: Container Apps secrets injected from IaC, Key Vault for OpenAI key
- Secrets never appear in logs, frontend, or exported reports

## Data Handling
- Uploaded label images may contain commercially sensitive information
- Images are sent to Azure OpenAI and Azure Vision OCR for extraction
- Review evidence (images, results, decisions) stored in Azure Blob Storage (private container)
- Blob container `jobs` is private; no public access
- Local development stores evidence in `/tmp/barrel_storage_local`

## Authentication
- Demo evaluator login (username/password)
- API protected by `X-BARREL-REVIEW-TOKEN` header
- No public signup
- CORS restricted to known frontend origins
- HTTPS enforced on Azure Container Apps

## Upload Validation
- File size limit: 25MB
- Allowed formats: JPEG, PNG, WebP, ZIP
- Extension and MIME type validated server-side

## Prototype Limitations
- Demo auth credentials, not production-grade authentication
- No multi-factor authentication
- No role-based access control
- No audit logging
- No FedRAMP compliance
- No PII compliance program
- No data retention policy enforcement
