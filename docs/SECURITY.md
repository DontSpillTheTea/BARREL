# Security

## General Policies
- **No committed secrets**: Do not commit keys, `.env` files, or production credentials.
- **No required outbound AI**: System operates entirely locally by default.
- **No arbitrary URL fetching**: To prevent SSRF, external URLs are not retrieved based on user input.
- **No permanent upload persistence by default**: Uploads are processed and discarded.
- **Azure Key Vault**: Optional future path for secret management in Azure environments, but not required locally.

## Container Security
- **Container boundaries**: Strict separation between web, api, and ocr-worker.
- **Internal OCR Worker**: The OCR worker is not exposed externally in the final topology.
- **Non-root containers**: Running containers as non-root is a goal.

## Upload Restrictions
- **Upload size limits**: Hard limits on file sizes (e.g. 25MB).
- **MIME/extension allowlists**: Only specific image/zip formats are permitted.
