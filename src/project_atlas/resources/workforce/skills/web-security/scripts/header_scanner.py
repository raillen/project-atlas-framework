#!/usr/bin/env python3
import urllib.request
import sys

def check_security_headers(url):
    print(f"Scanning headers for: {url}\n")
    try:
        req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
        with urllib.request.urlopen(req) as response:
            headers = response.info()
            
            security_headers = {
                "Strict-Transport-Security": "Missing HSTS! Vulnerable to MITM attacks.",
                "Content-Security-Policy": "Missing CSP! High risk of XSS.",
                "X-Frame-Options": "Missing X-Frame-Options! Vulnerable to Clickjacking.",
                "X-Content-Type-Options": "Missing X-Content-Type-Options! Vulnerable to MIME sniffing.",
                "Referrer-Policy": "Missing Referrer-Policy! May leak sensitive URLs."
            }
            
            score = 100
            print("--- Header Analysis ---")
            for header, msg in security_headers.items():
                val = headers.get(header)
                if val:
                    print(f"[OK] {header}: {val}")
                else:
                    print(f"[FAIL] {msg}")
                    score -= 20
            
            # Check for information leakage
            server = headers.get("Server")
            if server:
                print(f"[-] Info Leak: Server header exposes '{server}'")
                
            x_powered_by = headers.get("X-Powered-By")
            if x_powered_by:
                print(f"[-] Info Leak: X-Powered-By header exposes '{x_powered_by}'")
                
            print(f"\nOverall Security Score: {score}/100")
            if score < 100:
                sys.exit(1)
            
    except Exception as e:
        print(f"Error accessing URL: {e}")
        sys.exit(1)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: header_scanner.py <URL>")
        sys.exit(1)
    
    check_security_headers(sys.argv[1])
