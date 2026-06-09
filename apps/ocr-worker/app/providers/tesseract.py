import pytesseract
from PIL import Image
import shutil
from .base import BaseOCRProvider, OCRResult

class TesseractProvider(BaseOCRProvider):
    def is_available(self) -> bool:
        return shutil.which("tesseract") is not None
        
    def extract(self, image: Image.Image) -> OCRResult:
        if not self.is_available():
            raise RuntimeError("tesseract_unavailable")
            
        try:
            text = pytesseract.image_to_string(image)
            
            data = pytesseract.image_to_data(image, output_type=pytesseract.Output.DICT)
            confidences = [int(c) for c in data['conf'] if c != '-1' and str(c).strip() != '']
            
            if confidences:
                mean_conf = sum(confidences) / len(confidences)
                mean_conf = round(mean_conf, 1)
            else:
                mean_conf = 0.0
                
            return OCRResult(
                provider="tesseract",
                text=text,
                mean_confidence=mean_conf,
                warnings=[] if mean_conf > 0 else ["No confidence values returned from OCR engine."],
                metadata={"text_length": len(text)}
            )
        except Exception as e:
            raise RuntimeError(f"ocr_failed: {str(e)}")

# Keep these for backward compatibility during transition if needed
def is_tesseract_available() -> bool:
    return TesseractProvider().is_available()

def extract_text_and_confidence(image: Image.Image):
    res = TesseractProvider().extract(image)
    return res.text, res.mean_confidence
