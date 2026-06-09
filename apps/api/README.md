# BARREL API

This directory contains the Go API for BARREL.

* `apps/api` is the Go API target.
* `apps/api/app/` is legacy/bootstrap python scaffold and should not be extended.

## Endpoints

* `GET /health` - API health check
* `GET /health/ocr-worker` - Proxies health check to the OCR worker
* `POST /api/v1/ocr/extract` - Uploads an image (PNG, JPG, JPEG) to extract raw OCR text, image quality, and OCR confidence via the worker.

*Note: This is an intermediate OCR step and does not yet perform deterministic compliance rule validation.*
