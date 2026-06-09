#!/usr/bin/env python3
import urllib.request
import sys
import os

def check_web():
    url = os.environ.get("BARREL_WEB_BASE_URL", "http://localhost:5173")
    print(f"Checking {url}...")
    try:
        req = urllib.request.Request(url, method='GET')
        with urllib.request.urlopen(req) as response:
            if response.status != 200:
                print(f"FAIL: {url} returned {response.status}")
                sys.exit(1)
            
            content = response.read().decode('utf-8')
            if 'BARREL' in content or 'root' in content or '<div id="root"></div>' in content:
                print(f"OK: {url}")
            else:
                print(f"FAIL: {url} returned successful status but unexpected content")
                sys.exit(1)
    except Exception as e:
        print(f"FAIL: {url} threw exception {e}")
        sys.exit(1)

def main():
    check_web()
    print("Web smoke checks passed.")

if __name__ == "__main__":
    main()
