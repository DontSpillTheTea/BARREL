#!/usr/bin/env python3
import sys
import os
from PIL import Image

# Add apps/ocr-worker to path to import providers
sys.path.append(os.path.join(os.path.dirname(__file__), '..', 'apps', 'ocr-worker'))

from app.providers.tesseract import TesseractProvider
from app.providers.paddleocr_provider import PaddleOCRProvider
from app.providers.auto import AutoProvider

def main():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    image_path = os.path.join(script_dir, '..', 'samples', 'generated', 'good', 'good_07_brand_case_variation.png')
    if not os.path.exists(image_path):
        print(f"Sample image not found: {image_path}")
        sys.exit(1)
        
    print(f"Loading image: {image_path}")
    image = Image.open(image_path)
    
    print("\n--- Tesseract Provider ---")
    tess = TesseractProvider()
    if tess.is_available():
        res = tess.extract(image)
        print(f"Confidence: {res.mean_confidence}")
        print(f"Text length: {len(res.text)}")
        print(f"Excerpt: {res.text[:100].replace(chr(10), ' ')}...")
    else:
        print("Unavailable")
        
    print("\n--- PaddleOCR Provider ---")
    pad = PaddleOCRProvider()
    if pad.is_available():
        res = pad.extract(image)
        print(f"Confidence: {res.mean_confidence}")
        print(f"Text length: {len(res.text)}")
        print(f"Excerpt: {res.text[:100].replace(chr(10), ' ')}...")
    else:
        print("Unavailable")

    print("\n--- Auto Provider ---")
    auto = AutoProvider()
    if auto.is_available():
        res_dict = auto.extract(image)
        print(f"Selected Provider: {res_dict['selected_provider']}")
        print(f"Selection Reason: {res_dict['selection_reason']}")
        print(f"Confidence: {res_dict['mean_confidence']}")
        print(f"Text length: {len(res_dict['text'])}")
        print("Provider Results:")
        for pr in res_dict.get('provider_results', []):
            print(f"  - {pr['provider']}: conf={pr['mean_confidence']}, len={pr['text_length']}")
    else:
        print("Unavailable")

if __name__ == "__main__":
    main()
