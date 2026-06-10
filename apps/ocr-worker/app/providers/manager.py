import os
import time
from typing import Dict, Any, Optional
from PIL import Image
from .base import BaseOCRProvider
from .tesseract import TesseractProvider
from .paddleocr_provider import PaddleOCRProvider
from .auto import AutoProvider

class ProviderManager:
    def __init__(self):
        self.providers: Dict[str, BaseOCRProvider] = {}
        self.provider_states: Dict[str, Dict[str, Any]] = {}
        self.default_provider = os.environ.get("OCR_PROVIDER", "paddleocr")
        self.deep_ocr_enabled = os.environ.get("OCR_DEEP_OCR_ENABLED", "false").lower() == "true"
        self.warmup_on_startup = os.environ.get("OCR_WARMUP_ON_STARTUP", "false").lower() == "true"
        self.requires_ready_for_analysis = os.environ.get("OCR_REQUIRE_READY_FOR_ANALYSIS", "true").lower() == "true"
        self.allow_fast_fallback = os.environ.get("OCR_ALLOW_FAST_FALLBACK", "false").lower() == "true"
        self.is_warming = False

    def get_provider(self, name: str) -> Optional[BaseOCRProvider]:
        return self.providers.get(name)

    def get_state(self, name: str) -> Dict[str, Any]:
        return self.provider_states.get(name, {
            "provider": name,
            "state": "unavailable",
            "message": "Not initialized.",
            "last_error": ""
        })

    def initialize(self):
        self.is_warming = True
        
        # Fast provider
        tess = TesseractProvider()
        tess_state = {"provider": "tesseract", "state": "unavailable", "message": "", "last_error": ""}
        if tess.is_available():
            self.providers["tesseract"] = tess
            tess_state["state"] = "ready"
            tess_state["message"] = "Tesseract available."
        else:
            tess_state["message"] = "Tesseract not installed."
        self.provider_states["tesseract"] = tess_state
        
        # Auto provider wraps others
        auto = AutoProvider()
        self.providers["auto"] = auto
        self.provider_states["auto"] = {"provider": "auto", "state": "ready", "message": "Auto logic ready.", "last_error": ""}

        # Deep provider
        pad_state = {"provider": "paddleocr", "state": "disabled", "message": "Deep OCR disabled.", "last_error": ""}
        if self.deep_ocr_enabled or self.warmup_on_startup:
            pad_state["state"] = "initializing"
            pad_state["message"] = "Initializing PaddleOCR..."
            self.provider_states["paddleocr"] = pad_state
            
            import threading
            def _warmup_paddle():
                try:
                    start_time = time.time()
                    pad = PaddleOCRProvider()
                    if pad.is_available():
                        # Run a tiny generated warmup image
                        warmup_img = Image.new('RGB', (100, 100), color='white')
                        pad.extract(warmup_img)
                        
                        self.providers["paddleocr"] = pad
                        warmup_ms = int((time.time() - start_time) * 1000)
                        self.provider_states["paddleocr"]["state"] = "ready"
                        self.provider_states["paddleocr"]["message"] = "PaddleOCR initialized and warmup succeeded."
                        self.provider_states["paddleocr"]["warmup_ms"] = warmup_ms
                        self.provider_states["paddleocr"]["last_error"] = ""
                    else:
                        self.provider_states["paddleocr"]["state"] = "failed"
                        self.provider_states["paddleocr"]["message"] = "PaddleOCR failed to report availability."
                except Exception as e:
                    self.provider_states["paddleocr"]["state"] = "failed"
                    self.provider_states["paddleocr"]["message"] = "PaddleOCR warmup failed."
                    self.provider_states["paddleocr"]["last_error"] = str(e)
                finally:
                    self.is_warming = False

            threading.Thread(target=_warmup_paddle, daemon=True).start()
        else:
            self.provider_states["paddleocr"] = pad_state
            self.is_warming = False

# Global instance
manager = ProviderManager()
