#!/usr/bin/env python3
import os
import re
import sys

# Simplified set of regexes for common secrets
SECRET_PATTERNS = {
    "AWS Access Key": r"(?i)AKIA[0-9A-Z]{16}",
    "Generic Secret / Password": r"(?i)(password|secret|api_key|token)[\s:=]+[\'\"]([^\'\"]{8,})[\'\"]",
    "RSA Private Key": r"-----BEGIN RSA PRIVATE KEY-----",
    "GitHub Token": r"(?i)ghp_[a-zA-Z0-9]{36}"
}

def scan_file_for_secrets(filepath):
    findings = []
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            lines = f.readlines()
            for i, line in enumerate(lines):
                for name, pattern in SECRET_PATTERNS.items():
                    match = re.search(pattern, line)
                    if match:
                        # Redact the secret for the report
                        secret_val = match.group(0) if len(match.groups()) == 0 else match.group(2)
                        redacted = secret_val[:3] + "***" if len(secret_val) > 3 else "***"
                        findings.append((i+1, name, redacted))
    except UnicodeDecodeError:
        pass # Skip binary files
    except Exception as e:
        print(f"Error reading {filepath}: {e}")

    return findings

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: scan_secrets.py <directory>")
        sys.exit(1)

    target_dir = sys.argv[1]
    total_findings = 0

    print(f"Scanning {target_dir} for secrets...")

    for root, _, files in os.walk(target_dir):
        for file in files:
            # Skip .git directory and compiled files
            if '.git' in root or file.endswith(('.pyc', '.class', '.o')):
                continue

            path = os.path.join(root, file)
            issues = scan_file_for_secrets(path)

            if issues:
                print(f"\n[!] Secrets found in {path}:")
                for line_num, secret_type, redacted in issues:
                    print(f"    Line {line_num}: {secret_type} (Found: {redacted})")
                    total_findings += 1

    if total_findings > 0:
        print(f"\n❌ Scan failed. Found {total_findings} potential secrets.")
        sys.exit(1)
    else:
        print("\n✅ Scan passed. No obvious hardcoded secrets found.")
        sys.exit(0)
