# First Project Walkthrough

## 1. Initialize a Project
```bash
mkdir my-app && cd my-app
git init
prumo init --non-interactive --profile /path/to/project-profile.json
```

## 2. Check Diagnostics
```bash
prumo doctor
```

## 3. Create and Lock a Goal
```bash
prumo goal new P01-G01 "Core Application Engine" --phase P01 --objective "Build resilient core application logic."
prumo goal state P01-G01 PLANNED
prumo goal state P01-G01 LOCKED
```

## 4. Compile Target Adapter
```bash
prumo compile --target codex
```
