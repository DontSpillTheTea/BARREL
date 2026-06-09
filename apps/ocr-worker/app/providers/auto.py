import os
from PIL import Image
from .base import BaseOCRProvider, OCRResult
from .tesseract import TesseractProvider
from .paddleocr_provider import PaddleOCRProvider

class AutoProvider(BaseOCRProvider):
    def __init__(self):
        self.tesseract = TesseractProvider()
        self.paddle = PaddleOCRProvider()
        self.threshold = float(os.environ.get("OCR_LOCAL_CONFIDENCE_THRESHOLD", "65"))
        self.min_length = int(os.environ.get("OCR_MIN_TEXT_LENGTH", "80"))

    def is_available(self) -> bool:
        return self.tesseract.is_available() or self.paddle.is_available()

    def extract(self, image: Image.Image) -> dict:
        if not self.is_available():
            raise RuntimeError("no_ocr_providers_available")

        results = []
        selected_result = None
        selection_reason = ""

        # Step 1: Run Tesseract
        tess_res = None
        if self.tesseract.is_available():
            try:
                tess_res = self.tesseract.extract(image)
                results.append({
                    "provider": "tesseract",
                    "mean_confidence": tess_res.mean_confidence,
                    "text_length": len(tess_res.text)
                })
            except Exception as e:
                print(f"Tesseract failed in auto: {e}")

        # Step 2: Check if Tesseract is good enough
        if tess_res is not None:
            if tess_res.mean_confidence >= self.threshold and len(tess_res.text) >= self.min_length:
                selected_result = tess_res
                selection_reason = f"Tesseract confidence ({tess_res.mean_confidence}) >= {self.threshold} and text length ok."
                return self._format_response("auto", selected_result.provider, results, selection_reason, selected_result)

        # Step 3: If not, try PaddleOCR
        paddle_res = None
        if self.paddle.is_available():
            try:
                paddle_res = self.paddle.extract(image)
                results.append({
                    "provider": "paddleocr",
                    "mean_confidence": paddle_res.mean_confidence,
                    "text_length": len(paddle_res.text)
                })
            except Exception as e:
                print(f"PaddleOCR failed in auto: {e}")

        # Step 4: Compare and select
        if paddle_res is not None:
            if tess_res is None:
                selected_result = paddle_res
                selection_reason = "Tesseract failed or unavailable, used PaddleOCR."
            else:
                # Compare Tesseract and PaddleOCR
                if paddle_res.mean_confidence > tess_res.mean_confidence or len(paddle_res.text) > len(tess_res.text):
                    selected_result = paddle_res
                    selection_reason = "Tesseract confidence below threshold or text too short; PaddleOCR returned higher confidence or more text."
                else:
                    selected_result = tess_res
                    selection_reason = "Tesseract confidence below threshold, but PaddleOCR was not better."
        else:
            if tess_res is not None:
                selected_result = tess_res
                selection_reason = "Tesseract confidence below threshold, and PaddleOCR failed/unavailable."
            else:
                raise RuntimeError("All OCR providers failed.")

        return self._format_response("auto", selected_result.provider, results, selection_reason, selected_result)

    def _format_response(self, engine, selected, provider_results, reason, result: OCRResult):
        return {
            "ocr_engine": engine,
            "selected_provider": selected,
            "provider_results": provider_results,
            "selection_reason": reason,
            "text": result.text,
            "mean_confidence": result.mean_confidence,
            "warnings": result.warnings
        }
