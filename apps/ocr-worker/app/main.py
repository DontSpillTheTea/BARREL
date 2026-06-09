from fastapi import FastAPI, UploadFile, File, Form, HTTPException
from fastapi.responses import JSONResponse
from PIL import Image
import io
from .providers.tesseract import extract_text_and_confidence, is_tesseract_available
from .image.quality import calculate_quality

app = FastAPI(title="BARREL OCR Worker")

ALLOWED_MIMES = {"image/png", "image/jpeg", "image/jpg"}
ALLOWED_EXTS = {".png", ".jpg", ".jpeg"}

@app.get("/health")
def health_check():
    return {"status": "ok", "service": "ocr-worker"}

@app.post("/ocr/extract")
async def ocr_extract(file: UploadFile = File(...)):
    filename = file.filename or "unknown"
    ext = ""
    if "." in filename:
        ext = "." + filename.rsplit(".", 1)[-1].lower()
        
    if file.content_type not in ALLOWED_MIMES or ext not in ALLOWED_EXTS:
        return JSONResponse(status_code=400, content={
            "status": "error",
            "service": "barrel-ocr-worker",
            "error_code": "unsupported_file",
            "message": "Only PNG and JPEG images are supported.",
            "warnings": []
        })
        
    if not is_tesseract_available():
        return JSONResponse(status_code=503, content={
            "status": "error",
            "service": "barrel-ocr-worker",
            "error_code": "ocr_engine_unavailable",
            "message": "Tesseract OCR is not available in this environment.",
            "warnings": ["Install tesseract or run through the Docker Compose OCR worker image."]
        })

    contents = await file.read()
    try:
        image = Image.open(io.BytesIO(contents))
        image.verify()
        image = Image.open(io.BytesIO(contents))
    except Exception:
        return JSONResponse(status_code=400, content={
            "status": "error",
            "service": "barrel-ocr-worker",
            "error_code": "invalid_image",
            "message": "The uploaded file could not be parsed as an image.",
            "warnings": []
        })
        
    try:
        quality_metadata = calculate_quality(image)
        text, mean_conf = extract_text_and_confidence(image)
        
        warnings = []
        if mean_conf == 0.0:
            warnings.append("No confidence values returned from OCR engine.")
            
        return {
            "status": "ok",
            "service": "barrel-ocr-worker",
            "filename": filename,
            "content_type": file.content_type,
            "ocr_engine": "tesseract",
            "text": text,
            "mean_confidence": mean_conf,
            "image_quality": quality_metadata,
            "warnings": warnings
        }
    except RuntimeError as e:
        if str(e).startswith("ocr_failed:"):
            return JSONResponse(status_code=500, content={
                "status": "error",
                "service": "barrel-ocr-worker",
                "error_code": "ocr_failed",
                "message": "The OCR engine failed to process the image.",
                "warnings": []
            })
        raise

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=9090)
