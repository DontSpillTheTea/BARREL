# BARREL OCR Worker

This directory contains the Python OCR and image processing worker for BARREL.
It owns the OCR/image-processing libraries (e.g., Tesseract).

## Endpoints

* `GET /health` - Worker health check
* `POST /ocr/extract` - Extracts text and confidence using local Tesseract.

## Image Processing
It computes basic image quality stats (blur score and contrast score) entirely in memory using Pillow, and runs pytesseract for extraction.

## Environment
When running via Docker Compose, the worker container installs its own system Tesseract. This avoids requiring Tesseract on the host machine.
