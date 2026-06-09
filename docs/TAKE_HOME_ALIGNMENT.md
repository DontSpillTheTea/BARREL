# Take-home Alignment

This document explains how BARREL aligns with the take-home requirements.

## Architecture Choices
- **Why Docker Compose helps reviewers**: It provides a repeatable setup across Ubuntu and Windows without complex local dependency management.
- **Why Go + Python is an appropriate technical choice**: Go is excellent for API routing, JSON contracts, and rule evaluation concurrency, while Python excels at image processing and OCR.
- **Why local-first meets network/security constraints**: It honors strict outbound traffic limits and avoids sending sensitive labels to third-party AI endpoints.
- **Why confidence scoring and breadcrumbs show attention to requirements**: Reviewers need to know *why* something flagged, and what rule it violated, rather than a black-box "fail".
