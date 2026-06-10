import requests
import sys
import time
import json
import os

API_URL = os.environ.get("API_URL", "http://localhost:8080")
SAMPLE_IMAGE = "samples/generated/good/good_07_brand_case_variation.png"

def wait_for_ready():
    print("Checking readiness...")
    for _ in range(60):
        try:
            res = requests.get(f"{API_URL}/health/ocr-worker-ready")
            if res.status_code == 200 and res.json().get("status") == "ready":
                print("Accurate OCR is ready. Submitting sample image...")
                return True
        except requests.RequestException:
            pass
        time.sleep(1)
    print("FAILED: OCR worker never became ready.")
    sys.exit(1)

def main():
    print("Running async OCR smoke test...")
    wait_for_ready()

    # 1. Submit async job
    with open(SAMPLE_IMAGE, "rb") as f:
        files = {"file": f}
        data = {
            "ocr_provider": "paddleocr",
            "beverage_type": "distilled_spirits",
        }
        start_time = time.time()
        res = requests.post(f"{API_URL}/api/v1/labels/analyze-async", files=files, data=data)
        
    if res.status_code != 202:
        print(f"FAILED: Expected 202, got {res.status_code}: {res.text}")
        sys.exit(1)
        
    accepted_latency = time.time() - start_time
    print(f"Job accepted in {accepted_latency:.2f} seconds.")
    if accepted_latency > 2.0:
        print("FAILED: Job creation took longer than 2 seconds.")
        sys.exit(1)

    job_data = res.json()
    job_id = job_data.get("job_id")
    poll_url = job_data.get("poll_url")
    
    if not job_id or not poll_url:
        print("FAILED: Missing job_id or poll_url in 202 response.")
        sys.exit(1)
        
    print(f"Job ID: {job_id}")
    
    # 2. Poll job status
    max_wait = 120
    poll_start = time.time()
    result_data = None
    
    while time.time() - poll_start < max_wait:
        status_res = requests.get(f"{API_URL}{poll_url}")
        if status_res.status_code != 200:
            print(f"FAILED: Poll returned {status_res.status_code}: {status_res.text}")
            sys.exit(1)
            
        job_status = status_res.json()
        status = job_status.get("status")
        
        if status == "succeeded":
            result_data = job_status.get("result")
            break
        elif status == "failed":
            print(f"FAILED: Job failed with error: {job_status.get('error')}")
            sys.exit(1)
            
        time.sleep(2)
        
    if not result_data:
        print("FAILED: Timeout waiting for async job.")
        sys.exit(1)
        
    total_time = time.time() - poll_start
    print(f"Job completed in {total_time:.2f} seconds.")

    # 3. Verify result
    ocr_result = result_data.get("ocr", {})
    provider = ocr_result.get("selected_provider")
    text = ocr_result.get("text", "")
    
    if provider != "paddleocr":
        print(f"FAILED: Selected provider was {provider}, expected paddleocr")
        sys.exit(1)
        
    if len(text) < 100:
        print(f"FAILED: OCR text is too short ({len(text)} chars)")
        sys.exit(1)

    confidence = ocr_result.get("mean_confidence", 0)
    
    print("\nSUCCESS: Async analysis completed correctly.")
    print(f"Selected Provider: {provider}")
    print(f"OCR Confidence: {confidence:.2f}%")
    print(f"Text excerpt: {text[:100]}...")

if __name__ == "__main__":
    main()
