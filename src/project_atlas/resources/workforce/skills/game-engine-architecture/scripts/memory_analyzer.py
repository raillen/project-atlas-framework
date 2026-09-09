#!/usr/bin/env python3
import re
import sys
from collections import defaultdict


def analyze_memory_log(filepath):
    """
    Analyzes a mock engine memory allocation log to identify potential leaks
    or high fragmentation.
    Log format expected:
    [ALLOC] ptr=0x1234 size=64 tag=Physics
    [FREE]  ptr=0x1234
    """
    allocations = {}
    tag_totals = defaultdict(int)

    try:
        with open(filepath, 'r') as f:
            lines = f.readlines()
    except FileNotFoundError:
        print(f"Log file {filepath} not found.")
        sys.exit(1)

    alloc_regex = re.compile(r'\[ALLOC\]\s+ptr=(0x[0-9a-fA-F]+)\s+size=(\d+)\s+tag=(\w+)')
    free_regex = re.compile(r'\[FREE\]\s+ptr=(0x[0-9a-fA-F]+)')

    for line_no, line in enumerate(lines, 1):
        alloc_match = alloc_regex.search(line)
        if alloc_match:
            ptr = alloc_match.group(1)
            size = int(alloc_match.group(2))
            tag = alloc_match.group(3)
            allocations[ptr] = {'size': size, 'tag': tag, 'line': line_no}
            tag_totals[tag] += size
            continue

        free_match = free_regex.search(line)
        if free_match:
            ptr = free_match.group(1)
            if ptr in allocations:
                tag = allocations[ptr]['tag']
                size = allocations[ptr]['size']
                tag_totals[tag] -= size
                del allocations[ptr]
            else:
                print(f"Warning: Freeing untracked pointer {ptr} at line {line_no}")

    print("=== Memory Leak Analysis ===")
    if allocations:
        print(f"Found {len(allocations)} un-freed allocations!")
        for ptr, info in allocations.items():
            print(f"  Leak: ptr={ptr}, size={info['size']} bytes, tag={info['tag']} (Allocated on line {info['line']})")
    else:
        print("No leaks detected. All tracked allocations were freed.")

    print("\n=== Active Memory by Tag ===")
    for tag, size in tag_totals.items():
        if size > 0:
            print(f"  {tag}: {size} bytes")

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage: memory_analyzer.py <alloc_log_file>")
        sys.exit(1)
    analyze_memory_log(sys.argv[1])
