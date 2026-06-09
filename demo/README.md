# Demo Requests


**Note**: `task` is the supported command interface. Docker Compose is wrapped by Task. Normal reviewers should use `task`, not Make or raw Docker Compose.


This folder contains JSON payloads used for testing the `analyze-text` API endpoint.

## Note

BARREL uses `task` as the primary command runner. The primary flow to test the backend is:

```bash
task smoke
```

However, if you want to test these payloads manually for debugging:

```bash
curl -sS -X POST http://localhost:8080/api/v1/labels/analyze-text \
  -H "Content-Type: application/json" \
  --data @demo/requests/analyze_text_good.json
```
