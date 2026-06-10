import os
import sys
import time
import requests
import json

def main():
    web_url = os.environ.get("BARREL_AZURE_WEB_URL")
    api_url = os.environ.get("BARREL_AZURE_API_URL")
    username = os.environ.get("BARREL_DEMO_USERNAME", "evaluator")
    password = os.environ.get("BARREL_DEMO_PASSWORD")
    token = os.environ.get("BARREL_REVIEW_TOKEN")

    if not api_url:
        print("SKIP: BARREL_AZURE_API_URL not set")
        sys.exit(0)

    print(f"Checking web URL: {web_url}")
    if web_url:
        r = requests.get(web_url)
        assert r.status_code == 200, f"Web URL failed: {r.status_code}"
    
    print(f"Checking API health: {api_url}/health")
    r = requests.get(f"{api_url}/health")
    assert r.status_code == 200, f"API health failed: {r.status_code}"

    headers = {}
    if password:
        print("Testing login...")
        r = requests.post(f"{api_url}/api/v1/auth/login", json={"username": username, "password": password})
        assert r.status_code == 200, f"Login failed: {r.status_code} {r.text}"
        data = r.json()
        token = data.get("token")

    if token:
        headers["X-BARREL-REVIEW-TOKEN"] = token

    print("Submitting analyze-async job...")
    image_path = "samples/generated/good/good_01_distilled_spirits_clean_front.png"
    with open(image_path, "rb") as f:
        files = {"file": f}
        data = {
            "beverage_type": "distilled_spirits",
            "ocr_provider": "azure_vision",
            "expected_json": json.dumps({"brand_name": "OLD TOM DISTILLERY"})
        }
        r = requests.post(f"{api_url}/api/v1/labels/analyze-async", files=files, data=data, headers=headers)
    
    assert r.status_code == 202, f"Async job submission failed: {r.status_code} {r.text}"
    job_data = r.json()
    job_id = job_data["job_id"]
    poll_url = job_data["poll_url"]
    print(f"Job ID: {job_id}, polling...")

    max_attempts = 30
    result = None
    for i in range(max_attempts):
        time.sleep(2)
        r = requests.get(f"{api_url}{poll_url}", headers=headers)
        assert r.status_code == 200, f"Poll failed: {r.status_code} {r.text}"
        status_data = r.json()
        if status_data["status"] == "succeeded":
            result = status_data["result"]
            break
        elif status_data["status"] == "failed":
            print(f"Job failed: {status_data}")
            sys.exit(1)
        print(f"Waiting... state: {status_data['status']}")
    
    assert result, "Job timed out"
    print("Job succeeded.")

    ocr_provider = result.get("ocr", {}).get("selected_provider")
    assert ocr_provider == "azure_vision", f"Expected azure_vision provider, got: {ocr_provider}"

    ocr_text = result.get("ocr", {}).get("text", "")
    assert len(ocr_text) > 100, f"Expected OCR text length > 100, got: {len(ocr_text)}"

    print("Checking /api/v1/reviews...")
    r = requests.get(f"{api_url}/api/v1/reviews", headers=headers)
    assert r.status_code == 200, f"Reviews list failed: {r.status_code} {r.text}"
    reviews = r.json()
    if isinstance(reviews, dict) and "reviews" in reviews:
        reviews = reviews["reviews"]
    job_ids = [str(rev.get("job_id", rev.get("id"))) for rev in reviews]
    assert job_id in job_ids, f"Job ID {job_id} not found in reviews list: {job_ids}"

    print("Azure smoke test passed successfully.")

if __name__ == "__main__":
    main()
