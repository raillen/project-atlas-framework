# Incident Response Playbook

Troubleshooting guide for common Project Atlas failures and recovery procedures.

## Quick Response Matrix

| Symptom | Likely Cause | Action | Time |
|---------|--------------|--------|------|
| `atlas: command not found` | Installation issue | → Playbook 1 | <5m |
| `Invalid project structure` | Corrupted .ai | → Playbook 2 | <10m |
| `Goal state transition invalid` | Logic error | → Playbook 3 | <15m |
| `Lock digest mismatch` | Unauthorized edit | → Playbook 4 | <5m |
| `Compile failed` | Adapter error | → Playbook 5 | <10m |
| `Doctor shows BROKEN` | Multiple errors | → Playbook 2 | <30m |
| `Performance degradation` | Cache/disk issue | → Playbook 6 | <5m |

---

## Playbook 1: Installation Issues

**Symptoms:**
- `atlas: command not found`
- `ModuleNotFoundError: No module named 'project_atlas'`
- `ImportError` on package startup

### Step 1: Verify Installation

```bash
# Check if installed
pip show project-atlas-framework

# Expected output (v0.3.0 or later):
# Name: project-atlas-framework
# Version: 0.3.0
# Location: /path/to/site-packages
```

**IF NOT FOUND:** → Go to Step 2

### Step 2: Reinstall Package

```bash
# Uninstall completely
pip uninstall -y project-atlas-framework

# Clear cache
rm -rf ~/.cache/pip

# Reinstall from PyPI
pip install project-atlas-framework

# Verify
atlas --version
```

**Expected:** `Project Atlas Framework v0.3.0`

**IF STILL FAILS:** → Go to Step 3

### Step 3: Development Installation

```bash
# Install from source (requires Git)
git clone https://github.com/your-org/project-atlas-framework.git
cd project-atlas-framework

# Install in development mode
pip install -e .

# Verify
atlas --version
atlas framework-check
```

**IF STILL FAILS:** → Check Python version and environment

### Step 4: Check Python Environment

```bash
# Verify Python version (need 3.10+)
python --version

# IF < 3.10, upgrade:
# macOS:
brew install python@3.12

# Ubuntu:
sudo apt-get install python3.12

# Then reinstall Atlas with Python 3.12
python3.12 -m pip install project-atlas-framework
python3.12 -m atlas --version
```

**Resolution Timeline:**
- [ ] 1m: Run Step 1
- [ ] 3m: Run Step 2
- [ ] 5m: Run Step 3
- [ ] 10m: Run Step 4

**Escalation:** If still failing, verify PyPI connectivity and check GitHub issues.

---

## Playbook 2: Corrupted Project

**Symptoms:**
- `atlas doctor` shows multiple ✗ errors
- `.ai/` directory missing or empty
- `atlas.json` invalid

### Step 1: Diagnose

```bash
# Get full diagnostics
atlas doctor <project> > diagnostics.log 2>&1
cat diagnostics.log

# Check structure
tree <project>/.ai/ 2>/dev/null || find <project>/.ai/ -type f

# Validate JSON
python -m json.tool <project>/.ai/manifest.json

# Check Git status
cd <project> && git status
```

**Common errors in diagnostics:**
- `No such file or directory: .ai/` → Missing .ai/ directory
- `Invalid JSON in goals/P*` → Corrupted Goal file
- `Orphaned skill` → Skill referenced but not found

### Step 2: Restore from Backup

```bash
# Option A: Full project restore (fastest if backup exists)
rm -rf <project>
tar -xzf backup-20260815.tar.gz

# Option B: Restore only .ai/ from backup
tar -xzf backup-20260815.tar.gz --strip-components=1 \
  -C <project> project/.ai/

# Verify
atlas doctor <project>
```

**IF BACKUP NOT AVAILABLE:** → Go to Step 3

### Step 3: Restore from Git

```bash
cd <project>

# Check Git history
git log --oneline | head -10

# Find last known good commit
git log --oneline -- .ai/ | head -10

# Reset to that commit
git reset --hard <good-commit>

# Verify
atlas doctor <project>
```

**IF NO GIT HISTORY:** → Go to Step 4

### Step 4: Re-initialize Project

```bash
# Backup corrupted state (for investigation)
cp -r <project> <project>.corrupted

# Create fresh project
atlas init <project> --profile profile.json --non-interactive

# Recompile
atlas compile --target generic --path <project>

# Verify
atlas doctor <project>
```

**WARNING:** This loses all Goals/Agents/Skills. Only use if no backup/Git history.

**Resolution Timeline:**
- [ ] 2m: Run Step 1 (diagnose)
- [ ] 3m: Run Step 2 (restore from backup)
- [ ] 5m: Run Step 3 (restore from Git)
- [ ] 10m: Run Step 4 (re-initialize)

**Escalation:** If all steps fail, data loss may be unavoidable. Escalate to data recovery specialist.

---

## Playbook 3: Goal State Transition Errors

**Symptoms:**
- `Invalid goal transition: DRAFT → EXECUTING`
- `Goal already locked`
- `Cannot transition from BLOCKED`

### Valid State Machine

```
DRAFT → {PLANNED, BLOCKED}
PLANNED → {LOCKED, DRAFT, BLOCKED}
LOCKED → {EXECUTING, BLOCKED}
EXECUTING → {VERIFYING, BLOCKED}
VERIFYING → {REVIEWING, EXECUTING, BLOCKED}
REVIEWING → {DONE, EXECUTING, BLOCKED}
BLOCKED → {PLANNED, LOCKED, EXECUTING}
DONE → {} (terminal)
```

### Step 1: Verify Current State

```bash
# Get Goal's current state
atlas goal list --format json | grep -A5 '"id": "P00-G01"'

# Or directly:
jq .state <project>/.ai/goals/P00-G01.goal.json
```

### Step 2: Find Valid Transition

```bash
# Current state: DRAFT
# Want to go to: EXECUTING
# ERROR: Direct transition not allowed

# Valid path: DRAFT → PLANNED → LOCKED → EXECUTING

# Execute transitions in order:
atlas goal update P00-G01 --state PLANNED
atlas goal update P00-G01 --state LOCKED
atlas goal update P00-G01 --state EXECUTING
```

### Step 3: If Locked

```bash
# Check if Goal is locked
jq .lock <project>/.ai/goals/P00-G01.goal.json

# If locked, only valid transition is → EXECUTING
atlas goal update P00-G01 --state EXECUTING

# After execution completes:
atlas goal update P00-G01 --state VERIFYING
atlas goal update P00-G01 --state REVIEWING
atlas goal update P00-G01 --state DONE
```

### Step 4: If Blocked

```bash
# To unblock, must transition through PLANNED or LOCKED
atlas goal update P00-G01 --state PLANNED

# Resolve the issue
# (e.g., fix missing dependency)

# Then continue normally
atlas goal update P00-G01 --state LOCKED
```

**Resolution Timeline:**
- [ ] 1m: Run Step 1 (check state)
- [ ] 2m: Run Step 2 (find path)
- [ ] 3m: Run Step 3-4 (execute transitions)

---

## Playbook 4: Lock Digest Mismatch

**Symptoms:**
- `Goal P00-G01: Lock digest mismatch`
- Goal was edited after being locked

### Step 1: Investigate

```bash
# Check lock state
jq .lock <project>/.ai/goals/P00-G01.goal.json

# Check history
jq .history[-5:] <project>/.ai/goals/P00-G01.goal.json

# See recent changes
git log -1 -p -- <project>/.ai/goals/P00-G01.goal.json
```

### Step 2: Decide: Accept or Reject Change

**IF CHANGE WAS INTENTIONAL (new amendment):**
```bash
# Accept the new version, update lock
atlas goal update P00-G01 --amend "Updated requirements"
atlas goal update P00-G01 --state LOCKED
```

**IF CHANGE WAS ACCIDENTAL (revert):**
```bash
# Restore to locked version
git checkout HEAD~1 -- <project>/.ai/goals/P00-G01.goal.json

# Re-lock (digest should match now)
atlas doctor <project>
```

### Step 3: Prevent Future Mismatches

```bash
# Use amendments for intentional changes
atlas goal update P00-G01 --amend "Reason for change"

# DO NOT manually edit locked Goal files
# Use CLI: atlas goal update

# After amendments, re-lock
atlas goal update P00-G01 --state LOCKED
```

**Resolution Timeline:**
- [ ] 1m: Run Step 1 (investigate)
- [ ] 2m: Run Step 2 (decide + act)
- [ ] 1m: Run Step 3 (prevent)

---

## Playbook 5: Compile Failures

**Symptoms:**
- `atlas compile --target codex: Failed`
- `.codex/` directory missing or empty
- Adapter workspace corrupted

### Step 1: Identify Failed Target

```bash
# Try each target
for target in generic codex claude-code claude chatgpt kimi traycer; do
  echo "=== Testing $target ==="
  atlas compile --target $target --path <project> && echo "✓ $target" || echo "✗ $target"
done
```

### Step 2: Rebuild Failed Target

```bash
# Clear cache
rm -rf <project>/.atlas/runtime/

# Rebuild specific target
atlas compile --target codex --path <project>

# If still fails:
# Rebuild ALL targets (adapter fixes may apply globally)
atlas compile --target generic --path <project>
atlas compile --target codex --path <project>
```

### Step 3: Verify Output

```bash
# Check if output exists
ls <project>/.codex/ 2>/dev/null | wc -l

# Expected: >0 files

# Validate output format
# (depends on adapter, but should be readable text/JSON)
head -20 <project>/.codex/ENTRYPOINT.md
```

### Step 4: Full Rebuild

```bash
# Clear all adapter outputs
rm -rf <project>/.{codex,claude-code,atlas}

# Recompile all targets
atlas compile --target generic --path <project>

# Run comprehensive check
atlas doctor <project>
```

**IF STILL FAILS:**

```bash
# This indicates adapter bug or incompatible Goal structure
# Escalate to development team with:
# - Project name and goals summary
# - Failed target(s)
# - Error message from compilation

# For immediate workaround, use working targets:
atlas compile --target generic --path <project>  # Usually most robust
```

**Resolution Timeline:**
- [ ] 2m: Run Step 1 (identify)
- [ ] 2m: Run Step 2 (rebuild)
- [ ] 1m: Run Step 3 (verify)
- [ ] 3m: Run Step 4 (full rebuild)

---

## Playbook 6: Performance Degradation

**Symptoms:**
- Commands taking >5s
- Slow compilation
- High CPU/memory usage

### Step 1: Check Disk Space

```bash
# Available space
df -h

# Project size
du -sh <project>

# Cache size
du -sh <project>/.atlas/runtime/
du -sh <project>/.codex/

# Expected:
# Project: 10-100 MB
# Runtime cache: 5-50 MB
# Codex workspace: 50-500 MB
```

**IF DISK FULL:**
```bash
# Clear old builds
rm -rf <project>/.atlas/runtime/
rm -rf <project>/.codex/
rm -rf <project>/.claude*

# Clean Git
git gc  # Optimize Git repository

# If still full, clean backups
rm -f backup-*.tar.gz  # Keep only recent backups
```

### Step 2: Check File Count

```bash
# Number of Goals
ls <project>/.ai/goals/ | wc -l

# Number of agents in workspace
ls <project>/.codex/agents/ 2>/dev/null | wc -l

# If >1000 files, rebuild to clean up
rm -rf <project>/.atlas/runtime/
atlas compile --target generic --path <project>
```

### Step 3: Benchmark Specific Commands

```bash
# Time a command
time atlas doctor <project>

# Profile compilation
time atlas compile --target codex --path <project>

# Expected times:
# doctor: <500ms
# compile generic: <1s
# compile codex: <500ms
```

**IF >3s:** Likely cache corruption or large project

### Step 4: Clean Rebuild

```bash
# Full cache clear
rm -rf <project>/.atlas/

# Recompile all
atlas compile --target generic --path <project>

# Benchmark again
time atlas doctor <project>
```

**Resolution Timeline:**
- [ ] 1m: Run Step 1 (check disk)
- [ ] 1m: Run Step 2 (check files)
- [ ] 2m: Run Step 3 (benchmark)
- [ ] 2m: Run Step 4 (rebuild)

---

## Escalation Path

If incident is not resolved within 30 minutes:

1. **Gather diagnostics:**
   ```bash
   atlas framework-check > diagnostics.txt
   atlas doctor <project> >> diagnostics.txt
   git log --oneline <project>/ | head -20 >> diagnostics.txt
   du -sh <project>/* >> diagnostics.txt
   ```

2. **Document steps taken:**
   - Playbooks executed
   - Commands run and output
   - Time spent

3. **Create GitHub issue with:**
   - Incident summary
   - Reproduction steps
   - Diagnostics output
   - Environment info (OS, Python version, Git version)

4. **Contact development team:**
   - Link to GitHub issue
   - Request ETA for resolution

---

## See Also

- [Deployment](./deployment.md) — Installation troubleshooting
- [Monitoring](./monitoring.md) — Health checks that detect issues early
- [Recovery](./recovery.md) — Data restoration procedures
