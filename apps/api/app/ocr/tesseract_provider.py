from .provider import OCRProvider

class TesseractProvider(OCRProvider):
    def extract_text(self, image_path: str) -> str:
        # Placeholder for Tesseract OCR integration
        return "Extracted text placeholder"
