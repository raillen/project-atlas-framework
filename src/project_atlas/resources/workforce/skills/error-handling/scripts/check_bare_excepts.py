import argparse
import ast
import sys


class BareExceptChecker(ast.NodeVisitor):
    def __init__(self):
        self.issues = []

    def visit_ExceptHandler(self, node):
        if node.type is None:
            self.issues.append((node.lineno, "Bare 'except:' found. Use 'except Exception:' instead."))
        elif isinstance(node.type, ast.Name) and node.type.id == "BaseException":
            self.issues.append((node.lineno, "Catching 'BaseException' is discouraged. Catch specific exceptions or 'Exception' instead."))
        self.generic_visit(node)

def main():
    parser = argparse.ArgumentParser(description="Check for bare except blocks in Python files.")
    parser.add_argument("file", help="Python file to analyze")
    args = parser.parse_args()

    try:
        with open(args.file, "r") as f:
            tree = ast.parse(f.read(), filename=args.file)
    except Exception as e:
        print(f"Error reading file {args.file}: {e}")
        sys.exit(1)

    checker = BareExceptChecker()
    checker.visit(tree)

    if checker.issues:
        print(f"Error Handling Issues in {args.file}:")
        for lineno, msg in checker.issues:
            print(f"  Line {lineno}: {msg}")
        sys.exit(1)
    else:
        print(f"No bare excepts found in {args.file}.")

if __name__ == "__main__":
    main()
