#!/usr/bin/env python3
import ast
import os
import sys


class SubprocessVisitor(ast.NodeVisitor):
    def __init__(self):
        self.vulnerabilities = []

    def visit_Call(self, node):
        if isinstance(node.func, ast.Attribute) and node.func.attr in ['Popen', 'call', 'check_call', 'check_output', 'run']:
            if isinstance(node.func.value, ast.Name) and node.func.value.id == 'subprocess':
                # Check for shell=True
                shell_true = False
                for keyword in node.keywords:
                    if keyword.arg == 'shell' and isinstance(keyword.value, ast.Constant) and keyword.value.value is True:
                        shell_true = True
                        break

                if shell_true:
                    self.vulnerabilities.append({
                        'line': node.lineno,
                        'code': ast.unparse(node),
                        'type': 'Command Injection Risk (shell=True)'
                    })
        self.generic_visit(node)

def scan_file(filepath):
    try:
        with open(filepath, 'r') as f:
            code = f.read()
        tree = ast.parse(code)
        visitor = SubprocessVisitor()
        visitor.visit(tree)
        return visitor.vulnerabilities
    except Exception as e:
        print(f"Error parsing {filepath}: {e}", file=sys.stderr)
        return []

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python check_subprocess_usage.py <directory_or_file>")
        sys.exit(1)

    target = sys.argv[1]
    files_to_scan = []

    if os.path.isfile(target):
        files_to_scan.append(target)
    elif os.path.isdir(target):
        for root, _, files in os.walk(target):
            for file in files:
                if file.endswith('.py'):
                    files_to_scan.append(os.path.join(root, file))

    total_vulns = 0
    for file in files_to_scan:
        vulns = scan_file(file)
        if vulns:
            print(f"Risky subprocess usage found in {file}:")
            for v in vulns:
                print(f"  Line {v['line']}: {v['type']}")
                print(f"    Code: {v['code']}")
            total_vulns += len(vulns)

    if total_vulns > 0:
        sys.exit(1)
    else:
        print("No obvious shell=True vulnerabilities found.")
        sys.exit(0)
