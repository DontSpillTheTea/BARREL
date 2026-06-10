import os
import sys
import time
import requests
import json

def main():
    api_url = os.environ.get("BARREL_AZURE_API_URL", "http://localhost:8080")
    password = os.environ.get("BARREL_DEMO_PASSWORD", "fallback-demo-password-123")
    username = os.environ.get("BARREL_DEMO_USERNAME", "evaluator")

    print(f"Testing Review History on {api_url}")
    
    headers = {}
    r = requests.post(f"{api_url}/api/v1/auth/login", json={"username": username, "password": password})
    assert r.status_code == 200, f"Login failed: {r.status_code} {r.text}"
    token = r.json().get("token")
    headers["X-BARREL-REVIEW-TOKEN"] = token

    r = requests.get(f"{api_url}/api/v1/reviews", headers=headers)
    assert r.status_code == 200, f"Get reviews failed: {r.status_code}"
    data = r.json()
    assert "reviews" in data, "Expected 'reviews' array in response"
    reviews = data["reviews"]
    
    if len(reviews) == 0:
        print("No reviews found. Submit a job first.")
        sys.exit(0)
        
    first_review = reviews[0]
    assert "job_id" in first_review, "Missing job_id in summary"
    assert "filename" in first_review, "Missing filename in summary"
    assert "overall_confidence" in first_review, "Missing overall_confidence in summary"
    
    job_id = first_review["job_id"]
    print(f"Fetching detail for {job_id}")
    
    r = requests.get(f"{api_url}/api/v1/reviews/{job_id}", headers=headers)
    assert r.status_code == 200, f"Get review detail failed: {r.status_code}"
    detail = r.json()
    
    assert "summary" in detail, "Missing summary in detail"
    assert "result" in detail, "Missing result in detail"
    assert detail["summary"]["job_id"] == job_id, "Mismatch job_id"
    
    print("Review history endpoints are returning the correct structures.")

if __name__ == "__main__":
    main()
