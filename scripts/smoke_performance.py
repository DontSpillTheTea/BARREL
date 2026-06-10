import sys
import urllib.request
import urllib.error
import urllib.parse
import json
import time
import uuid

def main():
    analyze_url = "http://localhost:8080/api/v1/labels/analyze-text"
    
    start = time.time()
    
    payload = {
        "beverage_type": "distilled_spirits",
        "text": "STONE'S THROW SPIRITS\nRye Whiskey\n46% Alc./Vol. (92 Proof)\n750 mL\nGOVERNMENT WARNING: (1) According to the Surgeon General...",
        "expected_fields": {
            "brand_name": "Stone's Throw Spirits",
            "class_type": "Rye Whiskey",
            "alcohol_content": "46% Alc./Vol. (92 Proof)",
            "net_contents": "750 mL",
            "government_warning_present": True
        }
    }
    
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(analyze_url, data=data, headers={"Content-Type": "application/json"})
    
    try:
        with urllib.request.urlopen(req) as response:
            res_data = response.read()
            # Try to parse
            json.loads(res_data)
    except Exception as e:
        print(f"Performance smoke FAILED: {e}")
        sys.exit(1)
        
    elapsed = time.time() - start
    print(f"Analyze text (Fast API) took {elapsed:.2f} seconds.")
    
    if elapsed > 5.0:
        print("Performance FAILED: Text analysis took longer than 5 seconds.")
        sys.exit(1)
    
    print("Performance smoke checks passed.")

if __name__ == "__main__":
    main()
