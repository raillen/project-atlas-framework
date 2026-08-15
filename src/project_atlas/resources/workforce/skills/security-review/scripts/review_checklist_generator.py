#!/usr/bin/env python3
import json
import os
import sys

CHECKLIST = [
    {"id": "SR-01", "category": "Input Validation", "desc": "Are all inputs validated against a strict allowlist?"},
    {"id": "SR-02", "category": "Input Validation", "desc": "Is output encoding applied to prevent XSS?"},
    {"id": "SR-03", "category": "Authentication", "desc": "Are sensitive routes protected by authentication middleware?"},
    {"id": "SR-04", "category": "Authorization", "desc": "Are authorization checks performed before modifying resources?"},
    {"id": "SR-05", "category": "Data Protection", "desc": "Are secrets/credentials hardcoded? (Should be False)"},
    {"id": "SR-06", "category": "Data Protection", "desc": "Is sensitive data encrypted in transit and at rest?"},
    {"id": "SR-07", "category": "Error Handling", "desc": "Do error messages leak stack traces or sensitive internal paths?"},
]

def generate_markdown_checklist(pr_number):
    markdown = f"## Security Review Checklist for PR #{pr_number}\n\n"
    markdown += "Please review the following items and check them off if applicable and secure.\n\n"
    
    current_category = ""
    for item in CHECKLIST:
        if item["category"] != current_category:
            current_category = item["category"]
            markdown += f"### {current_category}\n"
        markdown += f"- [ ] **{item['id']}**: {item['desc']}\n"
    
    return markdown

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: review_checklist_generator.py <PR_NUMBER>")
        sys.exit(1)
        
    pr_num = sys.argv[1]
    content = generate_markdown_checklist(pr_num)
    print(content)
