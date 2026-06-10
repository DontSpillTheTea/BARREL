# BARREL Agent Guide

Project: **BARREL: Beverage Alcohol Review & Regulatory Evidence Logger**

Repo:
```text
git@github.com:DontSpillTheTea/barrel.git
```

Local path:
```text
/home/ama/github/barrel
```

BARREL is an AI-native beverage alcohol label review assistant for label image review, deterministic checking, and evidence logging.

This is for a take-home assignment. Keep the project practical, working, and well-documented.

## Core rule

Build a review assistant, not a final legal determination system.

Required statement to preserve:
```text
BARREL is a review assistant, not a final legal determination system.
```

## Current direction

BARREL is full-in on Azure-hosted, API-based AI-native parsing.

```text
React/Vite web UI
→ Go API
→ hosted image-capable AI parser
→ deterministic BARREL checks
→ object/blob review evidence storage
→ review history/detail viewer
```

Current product direction:

- `ai_native` is the primary parser.
- Azure OpenAI / Azure AI Foundry hosted API is preferred when quota is available.
- Direct OpenAI API is an acceptable temporary fallback if Azure OpenAI is blocked by subscription quota or region policy.
- `azure_vision_ocr` can exist as optional debug/baseline evidence only.
- Local-first OCR worker paths are not the primary goal and should not be presented as the normal evaluator workflow.
- Object/blob storage is preferred over Postgres unless requirements later prove otherwise.

## User workflow

1. Evaluator logs in with predefined credentials.
2. Uploads one image or a `.zip` batch.
3. BARREL creates one async job per image.
4. BARREL submits the image to `ai_native`.
5. BARREL runs deterministic validation and expected-vs-observed scoring.
6. BARREL stores image/result/review evidence in object/blob storage.
7. Review History shows each submission row-by-row.
8. Clicking a history row restores the original image and full parsed detail.
9. Human reviewer approves, rejects, or marks needs-more-info.

Human review is the fallback. OCR fallback is not the primary product story.

## Command runner direction

- `make setup` is the bootstrap entry point for installing local tools.
- `task` is the supported day-to-day command interface.

Expected flow:

1. `make setup`
2. `az login`
3. `task azure:login-check`
4. `task dev`
5. `task test:api`
6. `task smoke:local-ai`
7. `task azure:deploy`
8. `task azure:smoke:ai-native`

Use local tests first and Azure smoke second.

## Engineering rules

- Do not push unless explicitly asked.
- Do not deploy before a working AI model endpoint is proven.
- Do not run heavy Playwright tests against Azure unless explicitly requested.
- Do not randomly region-hop Azure OpenAI. Prove which subscription, region, model, and deployment actually have usable quota first.
- Do not claim success without smoke tests.
- Do not silently fall back from `ai_native` to `azure_vision_ocr`.
- Do not describe Ollama, Hugging Face endpoints, GPU VMs, or Azure ML managed endpoints as the preferred implementation.
- Keep docs updated whenever architecture or workflow changes.

## Storage model

Preferred review evidence layout:

- `jobs/{job_id}/image.<ext>`
- `jobs/{job_id}/ai_raw.json`
- `jobs/{job_id}/result.json`
- `jobs/{job_id}/decision.json`

Azure implementation uses Azure Blob Storage.

## Security expectations

- No public signup.
- HTTPS-only inbound access in Azure.
- Evaluator login or review token gates API access.
- Secrets stay in environment variables or Azure-managed secret stores.
