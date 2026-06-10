import io
import pytest
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock
from PIL import Image

from app.main import app
from app.image.quality import calculate_quality

# Trigger lifespan for tests
with TestClient(app) as client:
    pass

def test_unsupported_file_type():
    with TestClient(app) as client:
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
    with patch("app.main.manager") as mock_manager:
        from app.providers.base import OCRResult
        
        mock_provider = MagicMock()
        mock_provider.extract.return_value = OCRResult(
            provider="paddleocr",
            text="Mock text paddle",
            mean_confidence=98.0,
            warnings=[],
            metadata={"text_length": 16}
        )
        
        mock_manager.get_provider.return_value = mock_provider
        mock_manager.default_provider = "paddleocr"
        mock_manager.get_state.return_value = {"state": "ready"}
        
        img = Image.new('RGB', (100, 100), color='white')
        img_byte_arr = io.BytesIO()
        img.save(img_byte_arr, format='PNG')
        img_byte_arr.seek(0)
        
        with TestClient(app) as client:
            response = client.post(
                "/ocr/extract",
                data={"provider": "paddleocr"},
                files={"file": ("test.png", img_byte_arr, "image/png")}
            )
            
            assert response.status_code == 200
            data = response.json()
            assert data["status"] == "ok"
            assert data["text"] == "Mock text paddle"
            assert data["mean_confidence"] == 98.0
            assert data["image_quality"]["width"] == 100

def test_unsupported_provider():
    img = Image.new('RGB', (100, 100), color='white')
    img_byte_arr = io.BytesIO()
    img.save(img_byte_arr, format='PNG')
    img_byte_arr.seek(0)
    
    with patch("app.main.manager") as mock_manager:
        mock_manager.get_provider.return_value = None
        mock_manager.get_state.return_value = {"state": "ready"} # to bypass the not ready check
        mock_manager.default_provider = "tesseract"
        
        with TestClient(app) as client:
            response = client.post(
                "/ocr/extract",
                data={"provider": "fake_provider"},
                files={"file": ("test.png", img_byte_arr, "image/png")}
            )
            
            assert response.status_code == 400
            assert response.json()["error_code"] == "unsupported_provider"

def test_paddleocr_not_ready():
    img = Image.new('RGB', (100, 100), color='white')
    img_byte_arr = io.BytesIO()
    img.save(img_byte_arr, format='PNG')
    img_byte_arr.seek(0)
    
    with patch("app.main.manager") as mock_manager:
        mock_manager.get_state.return_value = {"state": "initializing"}
        mock_manager.default_provider = "paddleocr"
        
        with TestClient(app) as client:
            response = client.post(
                "/ocr/extract",
                data={"provider": "paddleocr"},
                files={"file": ("test.png", img_byte_arr, "image/png")}
            )
            
            assert response.status_code == 503
            assert response.json()["error_code"] == "ocr_provider_not_ready"
