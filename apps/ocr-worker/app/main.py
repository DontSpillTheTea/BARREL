from fastapi import FastAPI, UploadFile, File, Form, HTTPException
from fastapi.responses import JSONResponse
from PIL import Image
import io
from .providers.tesseract import TesseractProvider
from .providers.paddleocr_provider import PaddleOCRProvider
from .providers.auto import AutoProvider
from .image.quality import calculate_quality
import os

app = FastAPI(title="BARREL OCR Worker")

ALLOWED_MIMES = {"image/png", "image/jpeg", "image/jpg"}
ALLOWED_EXTS = {".png", ".jpg", ".jpeg"}

@app.get("/health")
def health_check():
    return {"status": "ok", "service": "ocr-worker"}

@app.post("/ocr/extract")
async def ocr_extract(
    file: UploadFile = File(...),
    provider: str = Form(None)
):
    if not provider:
        provider = os.environ.get("OCR_PROVIDER", "auto")

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
        
        provider_obj = None
        if provider == "tesseract":
            provider_obj = TesseractProvider()
        elif provider == "paddleocr":
            provider_obj = PaddleOCRProvider()
        else:
            provider_obj = AutoProvider()
            
        if not provider_obj.is_available():
            return JSONResponse(status_code=503, content={
                "status": "error",
                "service": "barrel-ocr-worker",
                "error_code": "ocr_engine_unavailable",
                "message": f"Provider {provider} is not available in this environment.",
                "warnings": []
            })
            
        res = provider_obj.extract(image)
        
        response_dict = {
            "status": "ok",
            "service": "barrel-ocr-worker",
            "filename": filename,
            "content_type": file.content_type,
            "image_quality": quality_metadata
        }
        
        if isinstance(res, dict):
            # Auto provider returns a dict with expanded format
            response_dict.update(res)
        else:
            warnings = res.warnings or []
            if res.mean_confidence == 0.0 and "No confidence values returned from OCR engine." not in warnings:
                warnings.append("No confidence values returned from OCR engine.")
            response_dict.update({
                "ocr_engine": provider,
                "selected_provider": res.provider,
                "provider_results": [{"provider": res.provider, "mean_confidence": res.mean_confidence, "text_length": len(res.text)}],
                "selection_reason": "Single provider requested.",
                "text": res.text,
                "mean_confidence": res.mean_confidence,
                "warnings": warnings
            })
            
        return response_dict
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
