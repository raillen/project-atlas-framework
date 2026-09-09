#!/usr/bin/env python3
import os
import re
import sys

WEAK_KEYS = [
    b"secret", b"password", b"123456", b"admin", b"test", b"dev", b"default"
]

def scan_for_jwt_secrets(filepath):
    # Regex to find common JWT generation/signing patterns
    # Example: jwt.encode(payload, "secret", algorithm="HS256")
    patterns = [
        re.compile(r'jwt\.encode\s*\([^,]+,\s*[\'"]([^\'"]+)[\'"]'),
        re.compile(r'JWT_SECRET\s*=\s*[\'"]([^\'"]+)[\'"]')
    ]

    findings = []
    try:
        with open(filepath, 'r') as f:
            lines = f.readlines()
            for i, line in enumerate(lines):
                for p in patterns:
                    match = p.search(line)
                    if match:
                        secret = match.group(1).encode('utf-8')
                        if secret in WEAK_KEYS or len(secret) < 16:
                            findings.append((i+1, line.strip(), secret.decode('utf-8')))
    except Exception as e:
        print(f"Error reading {filepath}: {e}")

    return findings

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: check_jwt_weak_keys.py <directory_or_file>")
        sys.exit(1)

    target = sys.argv[1]
    files_to_scan = []

    if os.path.isfile(target):
        files_to_scan.append(target)
    elif os.path.isdir(target):
        for root, _, files in os.walk(target):
            for file in files:
                if file.endswith('.py') or file.endswith('.js') or file.endswith('.ts'):
                    files_to_scan.append(os.path.join(root, file))

    total = 0
    for file in files_to_scan:
        issues = scan_for_jwt_secrets(file)
        if issues:
            print(f"Potential weak JWT secrets found in {file}:")
            for line_num, code, secret in issues:
                print(f"  Line {line_num}: Secret='{secret}' (Code: {code})")
                total += 1

    if total > 0:
        print("\nWARNING: Hardcoded or weak JWT secrets detected. Use a secure vault or environment variables.")
        sys.exit(1)
    else:
        print("No weak JWT secrets detected in scan.")
        sys.exit(0)
