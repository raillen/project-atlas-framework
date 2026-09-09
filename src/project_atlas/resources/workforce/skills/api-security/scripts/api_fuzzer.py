#!/usr/bin/env python3
import sys
import time
import urllib.error
import urllib.request


# A simple rate limiting tester (fuzzer for API endpoints)
def test_rate_limit(url, requests_per_sec, duration_sec):
    print(f"Starting rate limit test on {url}")
    print(f"Target: {requests_per_sec} req/s for {duration_sec} seconds")

    total_requests = requests_per_sec * duration_sec
    success = 0
    rate_limited = 0
    errors = 0

    interval = 1.0 / requests_per_sec

    for i in range(total_requests):
        start_time = time.time()
        try:
            req = urllib.request.Request(url, headers={'User-Agent': 'API-Fuzzer/1.0'})
            urllib.request.urlopen(req)
            success += 1
        except urllib.error.HTTPError as e:
            if e.code == 429:
                rate_limited += 1
            else:
                errors += 1
                print(f"HTTP Error: {e.code}")
        except Exception as e:
            errors += 1
            print(f"Error: {e}")

        elapsed = time.time() - start_time
        sleep_time = interval - elapsed
        if sleep_time > 0:
            time.sleep(sleep_time)

    print("\n--- Results ---")
    print(f"Total Requests Sent: {total_requests}")
    print(f"Successful (200 OK): {success}")
    print(f"Rate Limited (429 Too Many Requests): {rate_limited}")
    print(f"Other Errors: {errors}")

    if rate_limited == 0:
        print("WARNING: No rate limiting detected! The endpoint might be vulnerable to DoS.")
    else:
        print("OK: Rate limiting appears to be functioning.")

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("Usage: api_fuzzer.py <URL> <requests_per_sec> <duration_sec>")
        sys.exit(1)

    target_url = sys.argv[1]
    rps = int(sys.argv[2])
    duration = int(sys.argv[3])

    test_rate_limit(target_url, rps, duration)
