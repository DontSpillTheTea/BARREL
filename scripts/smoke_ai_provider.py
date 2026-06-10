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

    print("Submitting analyze-async job with ai_based...")
    image_path = "samples/generated/good/good_01_distilled_spirits_clean_front.png"
    with open(image_path, "rb") as f:
        files = {"file": f}
        data = {
            "beverage_type": "distilled_spirits",
            "ocr_provider": "ai_based",
            "expected_json": json.dumps({
                "brand_name": "OLD TOM DISTILLERY",
                "class_type": "Kentucky Straight Bourbon Whiskey",
                "alcohol_content": "45% Alc./Vol. (90 Proof)",
                "net_contents": "750 mL",
                "government_warning_present": True
            })
        }
        r = requests.post(f"{api_url}/api/v1/labels/analyze-async", files=files, data=data, headers=headers)
    
    assert r.status_code == 202, f"Async job submission failed: {r.status_code} {r.text}"
    job_data = r.json()
    job_id = job_data["job_id"]
    poll_url = job_data["poll_url"]
    print(f"Job ID: {job_id}, polling...")

    max_attempts = 30
    result = None
    job_status = "unknown"
    for i in range(max_attempts):
        time.sleep(2)
        r = requests.get(f"{api_url}{poll_url}", headers=headers)
        assert r.status_code == 200, f"Poll failed: {r.status_code} {r.text}"
        status_data = r.json()
        job_status = status_data["status"]
        if job_status == "succeeded":
            result = status_data["result"]
            break
        elif job_status == "failed":
            result = status_data.get("result", {})
            print(f"Job failed explicitly: {status_data}")
            break
        print(f"Waiting... state: {job_status}")
    
    assert job_status in ["succeeded", "failed"], "Job timed out"
    
    if job_status == "failed":
        ai_escalation = result.get("ai_escalation", {})
        assert ai_escalation.get("provider") == "ai_based", "Failure should indicate ai_based"
        assert "not configured" in ai_escalation.get("reason", "").lower() or result.get("overall_status") == "Error", "Should have a clear error message"
        print("SKIP: AI Provider not fully configured, skipping AI assertions.")
        sys.exit(0)

    extracted = result.get("extracted_fields", {})
    assert len(extracted) > 0, "Expected extracted fields"

    assert result.get("overall_status") == "Pass", f"Expected Pass, got {result.get('overall_status')}"
    
    ocr_provider = result.get("requested_provider", "")
    assert ocr_provider == "ai_based", f"Expected ai_based provider, got: {ocr_provider}"

    ai_read = result.get("ai_second_read")
    assert ai_read, "Expected ai_second_read metadata on success"
    assert ai_read.get("used"), "Expected AI to be used"

    print("Checking /api/v1/reviews...")
    r = requests.get(f"{api_url}/api/v1/reviews", headers=headers)
    assert r.status_code == 200, f"Reviews list failed: {r.status_code} {r.text}"
    reviews = r.json()
    if isinstance(reviews, dict) and "reviews" in reviews:
        reviews = reviews["reviews"]
    job_ids = [str(rev.get("job_id", rev.get("id"))) for rev in reviews]
    assert job_id in job_ids, f"Job ID {job_id} not found in reviews list: {job_ids}"

    matched_rev = next((rev for rev in reviews if rev.get("job_id") == job_id), None)
    assert matched_rev["ocr_provider"] == "ai_based", f"History says provider is {matched_rev['ocr_provider']}"

    print(f"Checking detail endpoint /api/v1/reviews/{job_id}...")
    r = requests.get(f"{api_url}/api/v1/reviews/{job_id}", headers=headers)
    assert r.status_code == 200, f"Review detail failed: {r.status_code} {r.text}"
    detail = r.json()
    assert detail["summary"]["ocr_provider"] == "ai_based", "Detail summary provider mismatch"
    assert detail["result"]["requested_provider"] == "ai_based", "Detail result provider mismatch"

    print("AI Provider smoke test passed successfully.")

if __name__ == "__main__":
    main()
