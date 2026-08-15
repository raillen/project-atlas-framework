import ast
import sys
import argparse
from pathlib import Path

def get_module_name(filepath, base_dir):
    try:
        rel_path = filepath.relative_to(base_dir)
        return str(rel_path.with_suffix("")).replace("/", ".")
    except ValueError:
        return None

def analyze_imports(src_dir):
    src_path = Path(src_dir)
    deps = {}
    
    for py_file in src_path.rglob("*.py"):
        mod_name = get_module_name(py_file, src_path)
        if not mod_name:
            continue
            
        with open(py_file, "r") as f:
            try:
                tree = ast.parse(f.read())
            except Exception:
                continue
                
        deps[mod_name] = []
        for node in ast.walk(tree):
            if isinstance(node, ast.Import):
                for alias in node.names:
                    deps[mod_name].append(alias.name)
            elif isinstance(node, ast.ImportFrom):
                if node.module:
                    deps[mod_name].append(node.module)
                    
    return deps

def find_cycle(deps, start, current, visited, path):
    visited.add(current)
    path.append(current)
    
    for neighbor in deps.get(current, []):
        # We only care about internal imports
        if not any(neighbor.startswith(m) for m in deps.keys()):
            continue
            
        if neighbor == start:
            return path + [start]
        if neighbor not in visited:
            cycle = find_cycle(deps, start, neighbor, visited.copy(), path.copy())
            if cycle:
                return cycle
    return None

def main():
    parser = argparse.ArgumentParser(description="Detect circular dependencies in Python code.")
    parser.add_argument("src_dir", help="Source directory")
    args = parser.parse_args()
    
    deps = analyze_imports(args.src_dir)
    cycles = []
    
    for mod in deps:
        cycle = find_cycle(deps, mod, mod, set(), [])
        if cycle and cycle not in cycles:
            cycles.append(cycle)
            
    if cycles:
        print("Circular Dependencies Detected:")
        for cycle in cycles:
            print(" -> ".join(cycle))
        sys.exit(1)
    else:
        print("No circular dependencies detected.")

if __name__ == "__main__":
    main()
