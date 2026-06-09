from abc import ABC, abstractmethod
from typing import Dict, Any, List, Optional
from PIL import Image

class OCRResult:
    def __init__(
        self,
        provider: str,
        text: str,
        mean_confidence: float,
        warnings: List[str] = None,
        metadata: Dict[str, Any] = None
    ):
        self.provider = provider
        self.text = text
        self.mean_confidence = mean_confidence
        self.warnings = warnings or []
        self.metadata = metadata or {}

class BaseOCRProvider(ABC):
    @abstractmethod
    def is_available(self) -> bool:
        pass
        
    @abstractmethod
    def extract(self, image: Image.Image) -> OCRResult:
        pass
