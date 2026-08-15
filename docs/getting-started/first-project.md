# First Project Walkthrough

## 1. Initialize a Project
```bash
mkdir my-app && cd my-app
git init
atlas init --non-interactive --profile /path/to/project-profile.json
```

## 2. Check Diagnostics
```bash
atlas doctor
```

## 3. Create and Lock a Goal
```bash
atlas goal new P01-G01 "Core Application Engine" --phase P01 --objective "Build resilient core application logic."
atlas goal state P01-G01 PLANNED
atlas goal state P01-G01 LOCKED
```

## 4. Compile Target Adapter
```bash
atlas compile --target codex
```
