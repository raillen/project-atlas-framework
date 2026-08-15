import ast
import sys
import argparse

class ComplexityVisitor(ast.NodeVisitor):
    def __init__(self):
        self.complexity = {}

    def visit_FunctionDef(self, node):
        score = 1 # Base score
        for child in ast.walk(node):
            if isinstance(child, (ast.If, ast.While, ast.For, ast.ExceptHandler, ast.With)):
                score += 1
            elif isinstance(child, ast.BoolOp) and isinstance(child.op, (ast.And, ast.Or)):
                score += len(child.values) - 1
        self.complexity[node.name] = score
        self.generic_visit(node)

def main():
    parser = argparse.ArgumentParser(description="Analyze cyclomatic complexity of Python files.")
    parser.add_argument("file", help="Python file to analyze")
    parser.add_argument("--threshold", type=int, default=10, help="Complexity threshold")
    args = parser.parse_args()

    try:
        with open(args.file, "r") as f:
            tree = ast.parse(f.read(), filename=args.file)
    except Exception as e:
        print(f"Error reading file {args.file}: {e}")
        sys.exit(1)

    visitor = ComplexityVisitor()
    visitor.visit(tree)

    failed = False
    for func, score in visitor.complexity.items():
        if score > args.threshold:
            print(f"WARNING: Function '{func}' has a complexity of {score} (threshold: {args.threshold})")
            failed = True
        else:
            print(f"OK: Function '{func}' complexity is {score}")
            
    if failed:
        sys.exit(1)

if __name__ == "__main__":
    main()
