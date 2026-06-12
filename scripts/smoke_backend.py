#!/usr/bin/env python3
import urllib.request
import json
import sys
import os

def check(url, expected_status=200, is_post=False, payload=None):
    print(f"Checking {url}...")
    req = urllib.request.Request(url, method='POST' if is_post else 'GET')
    if is_post and payload is not None:
        req.add_header('Content-Type', 'application/json')
        data = json.dumps(payload).encode('utf-8')
        req.data = data

    try:
        with urllib.request.urlopen(req) as response:
            if response.status != expected_status:
                print(f"FAIL: {url} returned {response.status}")
                sys.exit(1)
            print(f"OK: {url}")
            return json.loads(response.read().decode('utf-8'))
    except Exception as e:
        print(f"FAIL: {url} threw exception {e}")
        sys.exit(1)

def main():
    # 1. API Health
    api_health_url = "http://localhost:8080/health"
    print(f"Checking {api_health_url}...")
    try:
        req = urllib.request.Request(api_health_url)
        with urllib.request.urlopen(req) as response:
            if response.status != 200:
                print(f"FAILED: {api_health_url} returned {response.status}")
                sys.exit(1)
            print(f"OK: {api_health_url}")
    except Exception as e:
        print(f"FAILED: Could not connect to {api_health_url}: {e}")
        sys.exit(1)

    # 2. Text Analysis
    demo_file = "demo/requests/analyze_text_good.json"
    if os.path.exists(demo_file):
        with open(demo_file, "r") as f:
            payload = json.load(f)
    else:
        payload = {
            "beverage_type": "distilled_spirits",
            "text": "GOVERNMENT WARNING: test",
            "expected_fields": {
                "government_warning_present": True
            }
        }

    res = check("http://localhost:8080/api/v1/labels/analyze-text", is_post=True, payload=payload)
    
    if "overall_status" not in res:
        print("FAIL: overall_status missing")
        sys.exit(1)
    
    if "fields" not in res or not res["fields"]:
        print("FAIL: fields missing or empty")
        sys.exit(1)
        
    for field in res["fields"]:
        if "rule" not in field or not field["rule"].get("id"):
            print(f"FAIL: empty rule breadcrumb for field {field.get('field')}")
            sys.exit(1)

    if "ai_escalation" not in res:
        print("FAIL: ai_escalation missing")
        sys.exit(1)

    print("All smoke checks passed.")
    print("Example Response Summary:")
    print(f"Overall Status: {res['overall_status']}")
    print(f"AI Escalation Eligible: {res['ai_escalation']['eligible']}")

if __name__ == "__main__":
    main()
