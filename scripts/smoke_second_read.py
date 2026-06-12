import os
import requests
import time
import sys

API_URL = os.environ.get("BARREL_AZURE_API_URL", "http://localhost:8080")
REVIEW_TOKEN = os.environ.get("BARREL_REVIEW_TOKEN", "fallback-review-token-123")
HEADERS = {"X-BARREL-REVIEW-TOKEN": REVIEW_TOKEN}
IMAGE_PATH = "samples/generated/good/good_10_dense_but_readable_label.png"

def main():
    print("Legacy manual second-read workflow has been retired.")
    print("Use scripts/smoke_local_ai.py or scripts/smoke_ai_provider.py to validate the AI-native parser path.")
    sys.exit(0)

if __name__ == "__main__":
    main()
