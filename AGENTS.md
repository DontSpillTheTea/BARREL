# AGENTS.md

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

BARREL is a local-first OCR and compliance-assist prototype for reviewing alcohol beverage labels with confidence scoring and regulatory breadcrumbs.

This is for a take-home assignment. Keep the project practical, working, and well-documented.

---

## Core rule

Build a **review assistant**, not a black-box compliance oracle.

BARREL must remain:

* local-first by default
* usable without required outbound AI calls
* transparent about confidence and uncertainty
* simple enough for nontechnical compliance agents
* aligned with the take-home requirements
* documented as it evolves

Required statement to preserve:

```text
BARREL is a review assistant, not a final legal determination system.
```

---

## Current architecture direction

BARREL is moving toward a Docker Compose based, polyglot prototype:

```text
React/Vite web UI
→ Go API and rules engine
→ Python OCR worker
→ local OCR engine such as Tesseract
```

Primary services:

```text
apps/web         React/Vite frontend
apps/api         Go API server, upload handling, rules engine, validation, reports
apps/ocr-worker  Python OCR/image-processing worker
rules/ttb        YAML regulatory rule catalog
samples          generated fixtures, expected JSON, prompts, and batch zips
```

Rationale:

* Docker Compose gives a repeatable demo path across Ubuntu and Windows.
* Go is a good fit for API routing, upload handling, JSON contracts, concurrency, and rules evaluation.
* Python is a better fit for OCR and image-processing libraries.
* OCR dependencies are easier to standardize in containers than through native Windows setup.
* The system still runs local-first and does not require outbound AI calls.

Avoid running two competing backend designs long-term. Older Python API scaffold files may exist from bootstrap, but the intended API implementation is Go.

---

## Command runner direction

Use **go-task** as the primary cross-platform task runner.

Primary command interface:

```bash
task dev
```

Compatibility interface:

```bash
make dev
```

Fallback interface:

```bash
docker compose up --build
```

`Taskfile.yml` is the source of truth for automation. The `Makefile` should be a thin compatibility shim that delegates to `task`.

Expected task commands:

```bash
task help
task check-env
task dev
task up
task down
task logs
task build
task test
task test:api
task test:ocr
task test:web
task samples
task clean
task status
```

Windows guidance should prefer Docker Desktop + WSL2. Native development may be supported, but Docker Compose is the preferred demo/reviewer path.

---

## Read first

Before meaningful work, read:

```text
AGENTS.md
README.md
docs/ARCHITECTURE.md
docs/ASSUMPTIONS.md
docs/ROADMAP.md
docs/SECURITY.md
docs/RULES_AND_REGULATORY_BREADCRUMBS.md
docs/TAKE_HOME_ALIGNMENT.md
.env.example
Taskfile.yml
Makefile
docker-compose.yml
```

Treat `docs/` as project memory. Future agents may not have prior chat context.

If one of these files does not exist yet, create or update it when the task requires it.

---

## Documentation rule

Update docs when project state changes.

Examples:

* Architecture changed → update `docs/ARCHITECTURE.md`
* Scope or phase changed → update `docs/ROADMAP.md`
* Security/secrets changed → update `docs/SECURITY.md`
* Rules changed → update `docs/RULES_AND_REGULATORY_BREADCRUMBS.md`
* Assumptions changed → update `docs/ASSUMPTIONS.md`
* Take-home alignment improved → update `docs/TAKE_HOME_ALIGNMENT.md`
* Commands changed → update `README.md`, `Taskfile.yml`, and `Makefile`
* Runtime topology changed → update `docker-compose.yml` and `docs/ARCHITECTURE.md`

Docs are part of the deliverable.

---

## Product direction

Target workflow:

```text
browser upload
→ local OCR
→ image quality scoring
→ deterministic extraction
→ fuzzy/normalized matching
→ advisory rule checks
→ confidence scoring
→ regulatory breadcrumbs
→ batch review UI
→ downloadable report
```

Default config must work with:

```env
AI_PROVIDER=none
OCR_PROVIDER=tesseract
SECRET_PROVIDER=env
```

Cloud AI may be optional later, but must not be required for the core app.

---

## Security rules

Never commit secrets.

Do not commit:

```text
.env
*.pem
*.key
*.p12
*.pfx
credentials.json
secrets.json
```

Use `.env.example` for placeholders.

Azure Key Vault may be documented/scaffolded as an optional future integration because the stakeholder environment is Azure-based, but Azure must not be required for local development.

Uploaded labels should be treated as potentially sensitive.

Security expectations:

* no required outbound AI endpoints
* no arbitrary user-controlled URL fetching
* upload size limits
* image/zip MIME and extension allowlists
* no permanent upload persistence by default
* safe errors, not raw stack traces
* no analytics or tracking
* no unnecessary SaaS dependencies
* containers should eventually run as non-root
* only the web-facing service should be exposed externally in the full Compose topology

---

## Rule/checking rules

Rules should live in the top-level rules catalog:

```text
rules/ttb/
```

Expected rule files:

```text
rules/ttb/health_warning.yaml
rules/ttb/distilled_spirits.yaml
rules/ttb/wine.yaml
rules/ttb/malt_beverages.yaml
```

Legacy/bootstrap rule files may also exist under:

```text
apps/api/app/rules/
```

Prefer the top-level `rules/ttb/` structure going forward.

Rules docs live in:

```text
docs/RULES_AND_REGULATORY_BREADCRUMBS.md
```

Initial rule areas:

* Government Health Warning Statement
* Brand Name
* Alcohol Content
* Net Contents
* Class/Type designation

Use advisory language. Do not claim the rule catalog is complete. Do not fabricate legal authority.

Each check should show a breadcrumb: rule id, citation/source, and explanation.

---

## Confidence rules

Confidence is not compliance.

A label can be:

* likely compliant but low confidence because the image is bad
* clearly mismatched with high confidence
* unreadable and therefore Needs Review

Default statuses:

```text
Pass
Needs Review
Likely Fail
```

Suggested thresholds:

```text
>= 85%    Pass
65-84%    Needs Review
< 65%     Likely Fail
```

Show raw OCR text somewhere inspectable.

---

## UX rules

Prioritize:

1. Drag-and-drop upload
2. Batch summary table
3. Clear status badges
4. Per-label detail view
5. Raw OCR visibility
6. Field-by-field explanations
7. Regulatory breadcrumbs
8. CSV/JSON download

Use plain language. Avoid jargon.

---

## Engineering rules

Prefer a working vertical slice over broad incomplete features.

Good implementation order:

1. Sample fixture matrix
2. Docker Compose scaffold
3. Taskfile command runner
4. Go API health endpoint
5. Python OCR worker health endpoint
6. API-to-worker OCR call
7. Rule catalog loader
8. Warning statement checker
9. ABV/proof/net contents extraction
10. Confidence scoring
11. Upload endpoint
12. Batch/zip endpoint
13. Reports
14. Frontend upload UI
15. Batch results table
16. Detail view

Avoid early overbuild:

* authentication
* database persistence
* COLA integration
* mandatory cloud AI
* full legal rule coverage
* Redis or queues unless needed
* reverse proxy unless needed for deployment
* complex microservice mesh beyond web/api/ocr-worker

---

## Docker and runtime rules

Docker Compose is the preferred demo/reviewer runtime.

Expected services:

```text
web
api
ocr-worker
```

Only expose what is necessary.

Initial development may expose:

```text
web: localhost frontend port
api: localhost API port
ocr-worker: internal only when practical
```

The final demo topology should avoid exposing the OCR worker directly.

Do not require outbound internet at runtime for label analysis.

---

## Git rules

Before changes:

```bash
git status --short
```

After changes:

```bash
git status --short
```

Do not overwrite user work. Read existing files before editing.

Use focused commits. Do not push unless explicitly asked.

---

## Definition of done for any task

A task is done when:

* code/docs are updated
* relevant checks ran, or skipped with a stated reason
* docs reflect any changed project state
* `task` / `make` commands are updated if workflow changed
* git status is reviewed
* summary explains what changed and what remains

---

## Final product story

BARREL helps alcohol label reviewers move faster by extracting label text, checking expected fields, flagging mismatches, scoring confidence, and showing regulatory breadcrumbs.

It is intentionally local-first because the target environment may restrict outbound network access and because production deployment would require careful data handling.

It uses Go for the API and rules workflow, Python for OCR/image processing, Docker Compose for repeatable review/demo setup, and go-task as the cross-platform command runner.

It assists human reviewers. It does not replace them.
