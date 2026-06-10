from fastapi import FastAPI, UploadFile, File, Form, HTTPException
from fastapi.responses import JSONResponse
from PIL import Image
import io
import os
import asyncio
from concurrent.futures import ThreadPoolExecutor

from .providers.manager import manager
from .image.quality import calculate_quality

# Define timeout for providers
OCR_PROVIDER_TIMEOUT = int(os.environ.get("OCR_PROVIDER_TIMEOUT_SECONDS", "4"))
OCR_DEEP_TIMEOUT = int(os.environ.get("OCR_DEEP_PROVIDER_TIMEOUT_SECONDS", "8"))

executor = ThreadPoolExecutor(max_workers=4)

async def lifespan(app: FastAPI):
    global executor
    if executor._shutdown:
        executor = ThreadPoolExecutor(max_workers=4)
    # Startup: Initialize providers based on config
    manager.initialize()
    yield
    # Shutdown
    executor.shutdown(wait=False)

app = FastAPI(title="BARREL OCR Worker", lifespan=lifespan)

ALLOWED_MIMES = {"image/png", "image/jpeg", "image/jpg"}
ALLOWED_EXTS = {".png", ".jpg", ".jpeg"}

@app.get("/health")
def health_check():
    return {"status": "ok", "service": "ocr-worker"}

@app.get("/ready")
def readiness_check():
    providers_list = [
        manager.get_state("tesseract"),
        manager.get_state("paddleocr")
    ]
    
    status = "ready"
    if manager.requires_ready_for_analysis:
        paddle_state = manager.get_state("paddleocr").get("state")
        if manager.is_warming or paddle_state == "initializing":
            status = "warming"
        elif paddle_state == "ready":
            status = "ready"
        else:
            if manager.allow_fast_fallback:
                status = "degraded"
            else:
                status = "not_ready"
                
    return {
        "status": status,
        "service": "ocr-worker",
        "default_provider": manager.default_provider,
        "requires_ready_for_analysis": manager.requires_ready_for_analysis,
        "providers": providers_list
    }

async def run_with_timeout(func, *args, timeout_sec: int):
    loop = asyncio.get_running_loop()
    return await asyncio.wait_for(
        loop.run_in_executor(executor, func, *args),
        timeout=timeout_sec
    )

@app.post("/ocr/extract")
async def ocr_extract(
    file: UploadFile = File(...),
    provider: str = Form(None)
):
    if not provider:
        provider = manager.default_provider

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
        
        # Check provider availability
        if provider not in ["tesseract", "paddleocr", "auto"]:
            return JSONResponse(status_code=400, content={
                "status": "error",
                "service": "barrel-ocr-worker",
                "error_code": "unsupported_provider",
                "message": f"Provider {provider} is not supported.",
                "warnings": []
            })
            
        if provider == "paddleocr":
            state = manager.get_state("paddleocr")
            if state.get("state") != "ready":
                return JSONResponse(status_code=503, content={
                    "status": "error",
                    "error_code": "ocr_provider_not_ready",
                    "message": "PaddleOCR is not ready.",
                    "selected_provider": "paddleocr"
                })

        provider_obj = manager.get_provider(provider)
        if not provider_obj:
            return JSONResponse(status_code=503, content={
                "status": "error",
                "service": "barrel-ocr-worker",
                "error_code": "ocr_engine_unavailable",
                "message": f"Provider {provider} is not available in this environment.",
                "warnings": []
            })
            
        timeout = OCR_DEEP_TIMEOUT if provider in ["paddleocr", "auto"] else OCR_PROVIDER_TIMEOUT
        
        try:
            res = await run_with_timeout(provider_obj.extract, image, timeout_sec=timeout)
        except asyncio.TimeoutError:
            return JSONResponse(status_code=504, content={
                "status": "error",
                "service": "barrel-ocr-worker",
                "error_code": "ocr_timeout",
                "message": f"Provider {provider} timed out after {timeout} seconds.",
                "warnings": []
            })
        
        response_dict = {
            "status": "ok",
            "service": "barrel-ocr-worker",
            "filename": filename,
            "content_type": file.content_type,
            "image_quality": quality_metadata
        }
        
        if isinstance(res, dict):
            # Auto provider returns a dict with expanded format
            if res.get("status") == "error":
                return JSONResponse(status_code=503, content=res)
            response_dict.update(res)
        else:
            warnings = res.warnings or []
            if provider == "tesseract":
                warnings.append("Fast fallback OCR may miss or corrupt label text. Results require review.")
                
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
    uvicorn.run("app.main:app", host="0.0.0.0", port=9090, reload=True)
