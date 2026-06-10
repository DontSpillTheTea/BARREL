import os
import sys
import time
import requests
import json

def main():
    api_url = os.environ.get("BARREL_AZURE_API_URL", "http://localhost:8080")
    password = os.environ.get("BARREL_DEMO_PASSWORD", "fallback-demo-password-123")
    username = os.environ.get("BARREL_DEMO_USERNAME", "evaluator")

    print(f"Testing Batch Upload on {api_url}")
    
    headers = {}
    r = requests.post(f"{api_url}/api/v1/auth/login", json={"username": username, "password": password})
    assert r.status_code == 200, f"Login failed: {r.status_code} {r.text}"
    token = r.json().get("token")
    headers["X-BARREL-REVIEW-TOKEN"] = token

    zip_path = "samples/batches/good_10.zip"
    if not os.path.exists(zip_path):
        print("Batch sample not found. Generating it via python3 scripts/build_sample_batches.py...")
        os.system("python3 scripts/build_sample_batches.py")
        
    assert os.path.exists(zip_path), f"Failed to find or generate {zip_path}"

    print("Uploading batch zip...")
    with open(zip_path, "rb") as f:
        files = {"file": f}
        data = {
            "beverage_type": "distilled_spirits",
            "ocr_provider": "azure_vision",
            "expected_json": json.dumps({"brand_name": "TEST"})
        }
        r = requests.post(f"{api_url}/api/v1/labels/analyze-async", files=files, data=data, headers=headers)
        
    assert r.status_code == 202, f"Batch upload failed: {r.status_code} {r.text}"
    batch_res = r.json()
    assert batch_res.get("batch") is True, "Expected batch response"
    assert "jobs" in batch_res, "Expected jobs in batch response"
    
    jobs = batch_res.get("jobs", [])
    assert len(jobs) >= 1, "Expected at least 1 job in batch"
    print(f"Batch queued {len(jobs)} jobs. Polling first 3...")

    jobs_to_poll = jobs[:3]
    completed_jobs = []

    for job in jobs_to_poll:
        job_id = job["job_id"]
        poll_url = f"/api/v1/jobs/{job_id}"
        
        max_attempts = 30
        job_status = "unknown"
        for i in range(max_attempts):
            time.sleep(2)
            r = requests.get(f"{api_url}{poll_url}", headers=headers)
            assert r.status_code == 200, f"Poll failed for {job_id}: {r.status_code} {r.text}"
            status_data = r.json()
            job_status = status_data["status"]
            if job_status in ["succeeded", "failed"]:
                completed_jobs.append(status_data)
                break
            print(f"Waiting for {job_id}... state: {job_status}")
        
        assert job_status in ["succeeded", "failed"], f"Job {job_id} timed out"

    print("Checking /api/v1/reviews...")
    r = requests.get(f"{api_url}/api/v1/reviews", headers=headers)
    assert r.status_code == 200, f"Reviews list failed: {r.status_code} {r.text}"
    reviews = r.json().get("reviews", [])
    job_ids_in_history = [str(rev.get("job_id", rev.get("id"))) for rev in reviews]

    for job in completed_jobs:
        job_id = job.get("job_id", job.get("id"))
        if not job_id:
            continue
        assert str(job_id) in job_ids_in_history, f"Job {job_id} not found in history"
        
        matched_rev = next((rev for rev in reviews if str(rev.get("job_id", rev.get("id"))) == str(job_id)), None)
        assert matched_rev["filename"], "History missing filename"
        assert matched_rev["ocr_provider"], "History missing provider"
        assert matched_rev["overall_status"], "History missing status"
        
        print(f"Checking detail endpoint for {job_id}...")
        r = requests.get(f"{api_url}/api/v1/reviews/{job_id}", headers=headers)
        assert r.status_code == 200, f"Review detail failed for {job_id}"
        detail = r.json()
        assert detail.get("summary"), "Detail missing summary"
        assert detail.get("result"), "Detail missing result"

    print("Batch test completed successfully.")

if __name__ == "__main__":
    main()
