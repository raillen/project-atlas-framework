#!/usr/bin/env python3
import os
import sys
import stat

def check_permissions(directory):
    print(f"Checking file permissions in: {directory}")
    issues_found = False
    
    for root, _, files in os.walk(directory):
        for file in files:
            path = os.path.join(root, file)
            try:
                st = os.stat(path)
                mode = st.st_mode
                
                # Check for world-writable files
                if bool(mode & stat.S_IWOTH):
                    print(f"[WARNING] World-writable file found: {path}")
                    issues_found = True
                    
                # Check for files with SUID/SGID bits set
                if bool(mode & stat.S_ISUID):
                    print(f"[WARNING] SUID bit set on file: {path}")
                    issues_found = True
                if bool(mode & stat.S_ISGID):
                    print(f"[WARNING] SGID bit set on file: {path}")
                    issues_found = True
                    
            except Exception as e:
                print(f"Error checking {path}: {e}")
                
    if issues_found:
        print("\nSecurity issues found with file permissions.")
        sys.exit(1)
    else:
        print("\nNo obvious permission issues found.")
        sys.exit(0)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: check_permissions.py <directory>")
        sys.exit(1)
        
    target_dir = sys.argv[1]
    if not os.path.isdir(target_dir):
        print(f"Error: {target_dir} is not a directory.")
        sys.exit(1)
        
    check_permissions(target_dir)
