import os
import requests
import time
import sys

API_URL = os.environ.get("BARREL_AZURE_API_URL", "http://localhost:8080")
REVIEW_TOKEN = os.environ.get("BARREL_REVIEW_TOKEN", "fallback-review-token-123")
HEADERS = {"X-BARREL-REVIEW-TOKEN": REVIEW_TOKEN}
IMAGE_PATH = "samples/generated/good/good_10_dense_but_readable_label.png"

def main():
    print(f"Testing API: {API_URL}")
    print(f"Uploading image: {IMAGE_PATH}")

    if not os.path.exists(IMAGE_PATH):
        print("Image not found! Please ensure it exists.")
        sys.exit(1)

    with open(IMAGE_PATH, 'rb') as f:
        res = requests.post(f"{API_URL}/api/v1/labels/analyze-async",
                            headers=HEADERS,
                            data={"ocr_provider": "azure_vision"},
                            files={"file": f})
    
    if res.status_code != 202:
        print(f"Failed to submit async job: {res.text}")
        sys.exit(1)
        
    job_id = res.json()["job_id"]
    print(f"Job ID: {job_id}, polling...")

    status = "processing"
    result = None
    for _ in range(30):
        time.sleep(2)
        res = requests.get(f"{API_URL}/api/v1/jobs/{job_id}", headers=HEADERS)
        if res.status_code == 200:
            job = res.json()
            status = job.get("status")
            if status in ("success", "failed"):
                result = job.get("result")
                break
    
    if status != "success":
        print(f"Job failed or timed out: {status}")
        sys.exit(1)
        
    ai_read = result.get("ai_second_read", {})
    eligible = ai_read.get("eligible", False)
    print(f"Azure Vision OCR completed.")
    print(f"AI Second Read Eligible: {eligible}")
    if eligible:
        print(f"Reason: {ai_read.get('reason')}")
        
    if not eligible:
        print("Test failed: good_10_dense_but_readable_label.png should be eligible for second read.")
        sys.exit(1)

    print("Attempting manual AI Second Read trigger...")
    res = requests.post(f"{API_URL}/api/v1/jobs/{job_id}/second-read", headers=HEADERS)
    if res.status_code == 400 and "not configured" in res.text:
        print("Azure OpenAI not configured. Manual second read unavailable.")
        print("Test passed in disabled mode.")
        sys.exit(0)
    elif res.status_code == 200:
        print("AI Second Read completed successfully!")
        data = res.json()
        print("Candidates:", data.get("ai_second_read", {}).get("candidates"))
        print("Findings:", data.get("ai_second_read", {}).get("findings"))
        print("Test passed in enabled mode.")
        sys.exit(0)
    else:
        print(f"Failed to run second read: {res.status_code} {res.text}")
        sys.exit(1)

if __name__ == "__main__":
    main()
