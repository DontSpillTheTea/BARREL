# BARREL OCR Worker


**Note**: `task` is the supported command interface. Docker Compose is wrapped by Task. Normal reviewers should use `task`, not Make or raw Docker Compose.


This directory contains the Python OCR and image processing worker for BARREL.
It owns the OCR/image-processing libraries (PaddleOCR and Tesseract).

## Endpoints

* `GET /health` - Worker health check
* `GET /ready` - OCR provider readiness check
* `POST /ocr/extract` - Extracts text and confidence. Accurate local OCR (PaddleOCR) is the default. Tesseract is explicit fallback only.

## Image Processing
It computes basic image quality stats (blur score and contrast score) entirely in memory using Pillow, and runs PaddleOCR/Tesseract for extraction.

## Environment
When running via Docker Compose (which has been tested and verified as the core runtime), the worker container installs its own system dependencies. PaddleOCR models are warmed up and cached on container startup.
