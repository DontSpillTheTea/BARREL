import requests
import sys
import time

def check_accurate_ocr():
    print("Running accurate OCR smoke test...")
    
    # Wait for ready explicitly just in case
    print("Checking readiness...")
    ready_resp = requests.get("http://localhost:8080/health/ocr-worker-ready")
    if ready_resp.status_code != 200:
        print(f"FAILED: Readiness endpoint returned {ready_resp.status_code}")
        sys.exit(1)
        
    ready_data = ready_resp.json()
    if ready_data.get("status") != "ready":
        print(f"FAILED: Accurate OCR is not ready. Status: {ready_data.get('status')}")
        sys.exit(1)
        
    print("Accurate OCR is ready. Submitting sample image...")
    
    files = {
        'file': ('good_07_brand_case_variation.png', open('samples/generated/good/good_07_brand_case_variation.png', 'rb'), 'image/png')
    }
    data = {
        'beverage_type': 'distilled_spirits',
        'ocr_provider': 'paddleocr'
    }
    
    start_time = time.time()
    try:
        resp = requests.post("http://localhost:8080/api/v1/labels/analyze", files=files, data=data)
    except Exception as e:
        print(f"FAILED: Connection error: {e}")
        sys.exit(1)
        
    elapsed = time.time() - start_time
    
    if resp.status_code != 200:
        print(f"FAILED: Analysis returned {resp.status_code}")
        print(resp.text)
        sys.exit(1)
        
    result = resp.json()
    
    ocr_result = result.get("ocr", {})
    if ocr_result.get("status") == "error":
        print(f"FAILED: OCR returned error: {ocr_result.get('message')}")
        sys.exit(1)
        
    selected_provider = ocr_result.get("selected_provider")
    if selected_provider != "paddleocr":
        print(f"FAILED: Selected provider was {selected_provider}, expected paddleocr")
        sys.exit(1)
        
    text = ocr_result.get("text", "")
    if len(text) < 50:
        print(f"FAILED: Text extracted is too short: {len(text)} characters")
        sys.exit(1)
        
    print(f"Accurate OCR successful! Took {elapsed:.2f} seconds.")
    if elapsed > 5.0:
        print(f"WARNING: Accurate OCR took longer than 5 seconds ({elapsed:.2f}s).")
    
    sys.exit(0)

if __name__ == "__main__":
    check_accurate_ocr()
