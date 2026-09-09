import argparse
import ast
import sys
from pathlib import Path


def get_imports(filepath):
    with open(filepath, "r") as f:
        tree = ast.parse(f.read(), filename=str(filepath))

    imports = []
    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            for alias in node.names:
                imports.append(alias.name)
        elif isinstance(node, ast.ImportFrom):
            if node.module:
                imports.append(node.module)
    return imports

def main():
    parser = argparse.ArgumentParser(description="Detect architectural layer violations.")
    parser.add_argument("src_dir", help="Source directory to analyze")
    args = parser.parse_args()

    src_path = Path(args.src_dir)
    domain_files = list(src_path.rglob("domain/**/*.py"))

    violations = []
    # Domain layer should not import from infrastructure or application
    disallowed_in_domain = ["infrastructure", "application", "interfaces"]

    for df in domain_files:
        imports = get_imports(df)
        for imp in imports:
            for disallowed in disallowed_in_domain:
                if disallowed in imp:
                    violations.append(f"Violation in {df}: Domain layer imports {imp}")

    if violations:
        print("Architectural Violations Found:")
        for v in violations:
            print(f" - {v}")
        sys.exit(1)
    else:
        print("No architectural violations detected.")

if __name__ == "__main__":
    main()
