import numpy as np
from PIL import Image
from .base import BaseOCRProvider, OCRResult

class PaddleOCRProvider(BaseOCRProvider):
    def __init__(self):
        self._ocr = None
        self._available = None
        
    def is_available(self) -> bool:
        if self._available is not None:
            return self._available
            
        try:
            from paddleocr import PaddleOCR
            import logging
            logging.getLogger("ppocr").setLevel(logging.ERROR)
            
            # Initialize with English, CPU-only to avoid GPU dependency issues locally
            self._ocr = PaddleOCR(use_angle_cls=True, lang="en")
            self._available = True
        except ImportError:
            self._available = False
        except Exception as e:
            print(f"PaddleOCR init failed: {e}")
            self._available = False
            
        return self._available

    def extract(self, image: Image.Image) -> OCRResult:
        if not self.is_available():
            raise RuntimeError("paddleocr_unavailable")
            
        try:
            # Convert PIL image to numpy array (RGB)
            img_np = np.array(image.convert("RGB"))
            
            # Result is a list, where the first element contains lines
            # Each line: [[p1, p2, p3, p4], (text, confidence)]
            result = self._ocr.ocr(img_np)
            
            texts = []
            confidences = []
            
            # The structure returned by paddleocr can be tricky:
            # It's a list of lists or a list of dicts.
            if result and len(result) > 0 and result[0]:
                if isinstance(result[0], dict) and "rec_texts" in result[0]:
                    texts = result[0]["rec_texts"]
                    confidences = result[0].get("rec_scores", [])
                else:
                    for line in result[0]:
                        if isinstance(line, list) and len(line) == 2:
                            box, (text, conf) = line
                            texts.append(text)
                            confidences.append(conf)
            
            full_text = "\n".join(texts)
            
            if confidences:
                # Paddle returns confidence from 0.0 to 1.0, scale to 0-100 to match Tesseract
                mean_conf = (sum(confidences) / len(confidences)) * 100.0
                mean_conf = round(mean_conf, 1)
            else:
                mean_conf = 0.0
                
            return OCRResult(
                provider="paddleocr",
                text=full_text,
                mean_confidence=mean_conf,
                warnings=[],
                metadata={"text_length": len(full_text)}
            )
            
        except Exception as e:
            raise RuntimeError(f"ocr_failed: {str(e)}")
