import os
from PIL import Image
from .base import BaseOCRProvider, OCRResult

class AutoProvider(BaseOCRProvider):
    def __init__(self):
        self.threshold = float(os.environ.get("OCR_LOCAL_CONFIDENCE_THRESHOLD", "65"))
        self.min_length = int(os.environ.get("OCR_MIN_TEXT_LENGTH", "80"))

    def is_available(self) -> bool:
        return True # Handled by the manager logic

    def extract(self, image: Image.Image) -> dict:
        from .manager import manager
        
        tesseract = manager.get_provider("tesseract")
        paddle = manager.get_provider("paddleocr")
        
        if manager.requires_ready_for_analysis:
            paddle_state = manager.get_state("paddleocr").get("state")
            if paddle_state != "ready":
                if not manager.allow_fast_fallback:
                    return {
                        "status": "error",
                        "error_code": "ocr_provider_not_ready",
                        "message": "Required accurate provider (PaddleOCR) is not ready, and fast fallback is disabled.",
                        "selected_provider": "auto"
                    }
        
        if not tesseract and not paddle:
            raise RuntimeError("no_ocr_providers_available")

        results = []
        selected_result = None
        selection_reason = ""

        # Auto logic: if accurate is default/required, use it first or prefer it.
        # But wait, Auto should prefer PaddleOCR if it's the "accurate" provider.
        # Let's change the auto logic to try PaddleOCR first if available and ready.
        paddle_res = None
        if paddle and manager.get_state("paddleocr").get("state") == "ready":
            try:
                paddle_res = paddle.extract(image)
                results.append({
                    "provider": "paddleocr",
                    "mean_confidence": paddle_res.mean_confidence,
                    "text_length": len(paddle_res.text)
                })
            except Exception as e:
                print(f"PaddleOCR failed in auto: {e}")

        # If PaddleOCR is good enough, or it's the only one, return it.
        if paddle_res is not None:
            if paddle_res.mean_confidence >= self.threshold and len(paddle_res.text) >= self.min_length:
                selected_result = paddle_res
                selection_reason = f"PaddleOCR confidence ({paddle_res.mean_confidence}) >= {self.threshold} and text length ok."
                return self._format_response("auto", selected_result.provider, results, selection_reason, selected_result)

        # Try Tesseract if PaddleOCR wasn't good enough or unavailable
        tess_res = None
        if tesseract:
            try:
                tess_res = tesseract.extract(image)
                results.append({
                    "provider": "tesseract",
                    "mean_confidence": tess_res.mean_confidence,
                    "text_length": len(tess_res.text)
                })
            except Exception as e:
                print(f"Tesseract failed in auto: {e}")

        # Compare and select
        if tess_res is not None:
            if paddle_res is None:
                selected_result = tess_res
                selection_reason = "PaddleOCR failed or unavailable, used Tesseract fallback."
                selected_result.warnings.append("Accurate OCR unavailable; fast fallback used. Result requires review.")
            else:
                # Both ran
                if tess_res.mean_confidence > paddle_res.mean_confidence or len(tess_res.text) > len(paddle_res.text):
                    selected_result = tess_res
                    selection_reason = "PaddleOCR confidence below threshold; Tesseract returned higher confidence or more text."
                    selected_result.warnings.append("Accurate OCR unavailable; fast fallback used. Result requires review.")
                else:
                    selected_result = paddle_res
                    selection_reason = "PaddleOCR confidence below threshold, but Tesseract was not better."
        else:
            if paddle_res is not None:
                selected_result = paddle_res
                selection_reason = "PaddleOCR confidence below threshold, and Tesseract failed/unavailable."
            else:
                raise RuntimeError("All OCR providers failed.")

        return self._format_response("auto", selected_result.provider, results, selection_reason, selected_result)

    def _format_response(self, engine, selected, provider_results, reason, result: OCRResult):
        return {
            "status": "ok",
            "ocr_engine": engine,
            "selected_provider": selected,
            "provider_results": provider_results,
            "selection_reason": reason,
            "text": result.text,
            "mean_confidence": result.mean_confidence,
            "warnings": result.warnings
        }
