#!/usr/bin/env python3
import json
import subprocess
import sys


def check_npm_audit():
    print("Running npm audit...")
    try:
        # Run npm audit returning json
        result = subprocess.run(['npm', 'audit', '--json'], capture_output=True, text=True)
        # npm audit exits with non-zero if vulnerabilities are found

        audit_data = json.loads(result.stdout)
        metadata = audit_data.get('metadata', {})
        vulnerabilities = metadata.get('vulnerabilities', {})

        high = vulnerabilities.get('high', 0)
        critical = vulnerabilities.get('critical', 0)

        print(f"Found vulnerabilities: {vulnerabilities}")

        if high > 0 or critical > 0:
            print("❌ Critical or High vulnerabilities found in dependencies!")
            return False
        else:
            print("✅ No critical/high vulnerabilities found.")
            return True

    except FileNotFoundError:
        print("npm not found. Skipping.")
        return True
    except json.JSONDecodeError:
        print("Failed to parse npm audit output.")
        return False

if __name__ == "__main__":
    # In a real scenario, this would check package.json, requirements.txt, go.mod, etc.
    # using tools like safety, pip-audit, nancy, etc.
    success = check_npm_audit()
    if not success:
        sys.exit(1)
    sys.exit(0)
