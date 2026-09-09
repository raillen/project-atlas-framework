import argparse
import ast
import sys


class LockDetector(ast.NodeVisitor):
    def __init__(self):
        self.nested_locks = []

    def visit_With(self, node):
        # Check if the context manager looks like a lock (e.g., has 'lock' in the name)
        is_lock = False
        for item in node.items:
            if isinstance(item.context_expr, ast.Name) and 'lock' in item.context_expr.id.lower():
                is_lock = True
            elif isinstance(item.context_expr, ast.Attribute) and 'lock' in item.context_expr.attr.lower():
                is_lock = True

        if is_lock:
            # Look for nested locks
            for child in ast.walk(node):
                if child != node and isinstance(child, ast.With):
                    for sub_item in child.items:
                        name = ""
                        if isinstance(sub_item.context_expr, ast.Name):
                            name = sub_item.context_expr.id
                        elif isinstance(sub_item.context_expr, ast.Attribute):
                            name = sub_item.context_expr.attr

                        if 'lock' in name.lower():
                            self.nested_locks.append((node.lineno, child.lineno))

        self.generic_visit(node)

def main():
    parser = argparse.ArgumentParser(description="Detect potential deadlocks from nested lock acquisitions.")
    parser.add_argument("file", help="Python file to analyze")
    args = parser.parse_args()

    try:
        with open(args.file, "r") as f:
            tree = ast.parse(f.read(), filename=args.file)
    except Exception as e:
        print(f"Error reading file {args.file}: {e}")
        sys.exit(1)

    detector = LockDetector()
    detector.visit(tree)

    if detector.nested_locks:
        print(f"Potential Deadlock Warnings in {args.file}:")
        for l1, l2 in detector.nested_locks:
            print(f"  Outer lock at line {l1} contains inner lock at line {l2}. Ensure consistent lock acquisition order.")
        sys.exit(1)
    else:
        print(f"No nested locks detected in {args.file}.")

if __name__ == "__main__":
    main()
