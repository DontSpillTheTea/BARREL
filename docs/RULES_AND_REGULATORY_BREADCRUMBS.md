# Rules and Regulatory Breadcrumbs


**Note**: `task` is the supported command interface. Docker Compose is wrapped by Task. Normal reviewers should use `task`, not Make or raw Docker Compose.


## Local Rules Catalog
`rules/ttb/` is the canonical location for rule definitions going forward. The Go API loads these YAML files to deterministically evaluate extracted label text against expected regulatory standards.

## Advisory Rule Checks
Rule checks are **advisory prototypes**. The rule catalog is incomplete and does not constitute final legal determinations.

## Breadcrumbs
Citations and regulatory breadcrumbs (e.g., CFR section links and explanations) should appear in the results to assist reviewers in verifying the requirement.

## Deterministic Testing
A text-only endpoint (`POST /api/v1/labels/analyze-text`) exists to task extraction and rule testing fully deterministic, bypassing OCR variability.
