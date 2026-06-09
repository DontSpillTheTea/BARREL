from paddleocr import PaddleOCR
import numpy as np
from PIL import Image

ocr = PaddleOCR(use_textline_orientation=True, lang='en')
img = Image.open('/samples/generated/good/good_07_brand_case_variation.png')
result = ocr.ocr(np.array(img.convert('RGB')))

print("Type of result:", type(result))
print("Result:", result)
