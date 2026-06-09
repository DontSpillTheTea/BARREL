import pytesseract
from PIL import Image
import shutil

def is_tesseract_available() -> bool:
    return shutil.which("tesseract") is not None

def extract_text_and_confidence(image: Image.Image):
    if not is_tesseract_available():
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
            
        return text, mean_conf
    except Exception as e:
        raise RuntimeError(f"ocr_failed: {str(e)}")
