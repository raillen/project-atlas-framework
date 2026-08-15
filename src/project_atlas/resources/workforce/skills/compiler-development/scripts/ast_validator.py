import ast
import sys

def validate_ast(filepath: str):
    """
    Parses a Python file and validates that its AST doesn't use forbidden nodes
    (e.g., exec or eval) to enforce a strict subset of Python for a custom compiler.
    """
    try:
        with open(filepath, 'r') as f:
            source = f.read()
    except FileNotFoundError:
        print(f"Error: File not found: {filepath}")
        sys.exit(1)

    try:
        tree = ast.parse(source)
    except SyntaxError as e:
        print(f"Syntax error in {filepath}: {e}")
        sys.exit(1)

    forbidden_nodes = (ast.Exec, ast.Eval) if hasattr(ast, 'Exec') else ()
    
    violations = []
    for node in ast.walk(tree):
        if isinstance(node, ast.Call):
            if isinstance(node.func, ast.Name) and node.func.id in ['eval', 'exec']:
                violations.append((node.lineno, f"Forbidden built-in used: {node.func.id}"))
        elif forbidden_nodes and isinstance(node, forbidden_nodes):
            violations.append((node.lineno, f"Forbidden statement type used"))

    if violations:
        print(f"Validation failed for {filepath}:")
        for lineno, msg in violations:
            print(f"  Line {lineno}: {msg}")
        sys.exit(1)
    else:
        print(f"Validation passed for {filepath}.")
        sys.exit(0)

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage: python ast_validator.py <source_file>")
        sys.exit(1)
    validate_ast(sys.argv[1])
