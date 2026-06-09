import io
import pytest
from fastapi.testclient import TestClient
from unittest.mock import patch
from PIL import Image

from app.main import app
from app.image.quality import calculate_quality

client = TestClient(app)

def test_unsupported_file_type():
    response = client.post(
        "/ocr/extract",
        files={"file": ("test.txt", b"not an image", "text/plain")}
    )
    assert response.status_code == 400
    assert response.json()["error_code"] == "unsupported_file"

def test_image_quality_scoring_function():
    img = Image.new('RGB', (100, 100), color='gray')
    quality = calculate_quality(img)
    assert quality["width"] == 100
    assert quality["height"] == 100
    assert "contrast_score" in quality
    assert "blur_score" in quality

def test_successful_ocr_mocked():
    with patch("app.main.TesseractProvider") as mock_tess:
        mock_instance = mock_tess.return_value
        mock_instance.is_available.return_value = True
        
        # We need a mock OCRResult
        from app.providers.base import OCRResult
        mock_instance.extract.return_value = OCRResult(
            provider="tesseract",
            text="Mock text",
            mean_confidence=95.0,
            warnings=[],
            metadata={"text_length": 9}
        )
        
        img = Image.new('RGB', (100, 100), color='white')
        img_byte_arr = io.BytesIO()
        img.save(img_byte_arr, format='PNG')
        img_byte_arr.seek(0)
        
        response = client.post(
            "/ocr/extract",
            data={"provider": "tesseract"},
            files={"file": ("test.png", img_byte_arr, "image/png")}
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "ok"
        assert data["text"] == "Mock text"
        assert data["mean_confidence"] == 95.0
        assert data["image_quality"]["width"] == 100

def test_unsupported_provider():
    img = Image.new('RGB', (100, 100), color='white')
    img_byte_arr = io.BytesIO()
    img.save(img_byte_arr, format='PNG')
    img_byte_arr.seek(0)
    
    with patch("app.main.PaddleOCRProvider") as mock_pad:
        mock_instance = mock_pad.return_value
        mock_instance.is_available.return_value = False
        
        response = client.post(
            "/ocr/extract",
            data={"provider": "paddleocr"},
            files={"file": ("test.png", img_byte_arr, "image/png")}
        )
        
        assert response.status_code == 503
        assert response.json()["error_code"] == "ocr_engine_unavailable"

