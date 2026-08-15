#!/usr/bin/env python3
# scripts/triage_issues.py
# Scans recent issues and applies labels based on keyword heuristics.

import subprocess
import json
import sys

def run_gh_command(cmd):
    result = subprocess.run(cmd, capture_output=True, text=True, shell=True)
    if result.returncode != 0:
        print(f"Error running command: {cmd}\n{result.stderr}")
        sys.exit(1)
    return result.stdout

def triage_issues():
    print("Fetching recent open issues...")
    # Fetch issues created in the last 7 days that have no labels
    issues_json = run_gh_command("gh issue list --state open --json number,title,body,labels --limit 50")
    issues = json.loads(issues_json)

    for issue in issues:
        # Skip if already labeled
        if len(issue['labels']) > 0:
            continue
            
        num = issue['number']
        title = issue['title'].lower()
        body = issue['body'].lower()
        content = title + " " + body
        
        labels_to_add = []
        
        if "crash" in content or "exception" in content or "error" in content or "bug" in content:
            labels_to_add.append("bug")
        
        if "feature" in content or "support" in content or "add" in title:
            labels_to_add.append("enhancement")
            
        if "docs" in content or "readme" in content or "typo" in content:
            labels_to_add.append("documentation")
            
        if not labels_to_add:
            labels_to_add.append("needs-triage")
            
        labels_str = ",".join(labels_to_add)
        print(f"Triaging issue #{num} ('{issue['title']}') -> labels: {labels_str}")
        run_gh_command(f"gh issue edit {num} --add-label \"{labels_str}\"")

if __name__ == "__main__":
    triage_issues()
