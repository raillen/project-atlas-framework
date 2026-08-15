# Recovery Procedures

Disaster recovery and data restoration for Project Atlas projects.

## Backup Strategy

### What to Backup

**Critical (restore from these):**
- `.ai/` — All Goals, Agents, Skills (canonical source)
- `profile.json` — Project configuration
- `atlas.json` — Project manifest and version info

**Derived (can rebuild):**
- `.atlas/runtime/` — Compiled outputs (regenerate with `atlas compile`)
- `.codex/`, `.claude-code/`, etc — Adapter workspaces (regenerate)
- `.git/` — Use `git clone` or `git restore` instead

### Backup Frequency

```bash
# Daily backup (incremental)
tar --exclude='.atlas/runtime' --exclude='.codex' -czf \
  backup-$(date +%Y%m%d).tar.gz <project>/

# Weekly full backup (to S3 or equivalent)
aws s3 sync <project>/.ai/ s3://backups/project/goals/
aws s3 sync <project>/ s3://backups/project/latest/ \
  --exclude ".atlas/runtime/*" --exclude ".codex/*"

# Retention: keep daily for 30 days, weekly for 1 year
```

### Backup Verification

```bash
# Test backup integrity
tar -tzf backup-20260815.tar.gz | head -20

# Verify restored structure
tar -xzf backup-20260815.tar.gz -O project/.ai/manifest.json | jq .
```

---

## Recovery Scenarios

### Scenario 1: Corrupted Single Goal

**Symptom:**
```
atlas doctor <project>
# ✗ Goal P00-G01: Invalid JSON
# ✗ Overall: BROKEN
```

**Recovery:**

Option A: Restore from backup
```bash
# Find most recent backup before corruption
ls -lt backup-*.tar.gz | head -5

# Extract just that Goal file
tar -xzf backup-20260815.tar.gz \
  project/.ai/goals/P00-G01.goal.json

# Verify
atlas doctor <project>
```

Option B: Git history (if committed)
```bash
# Check git log
git log --oneline -- <project>/.ai/goals/P00-G01.goal.json | head -5

# Restore to known good state
git checkout <good-commit> -- <project>/.ai/goals/P00-G01.goal.json

# Verify
atlas doctor <project>
```

Option C: Manual repair
```bash
# View corrupted file
cat <project>/.ai/goals/P00-G01.goal.json

# If JSON is malformed, restore minimal Goal structure:
cat > <project>/.ai/goals/P00-G01.goal.json <<EOF
{
  "id": "P00-G01",
  "title": "Goal Title",
  "state": "DRAFT",
  "history": [{"at": "2026-08-15T...", "event": "recovered"}]
}
EOF

# Verify
atlas doctor <project>
```

### Scenario 2: Entire Project Corrupted

**Symptom:**
```
atlas doctor <project>
# No .ai directory or completely broken structure
```

**Recovery:**

Option A: Full restore from backup (fastest)
```bash
# Stop all running processes
pkill atlas

# Remove corrupted project
rm -rf <project>

# Restore from backup
tar -xzf backup-20260815.tar.gz

# Verify
atlas doctor <project>

# Recompile outputs
atlas compile --target generic --path <project>
```

Option B: Git reset (if changes committed)
```bash
cd <project>

# Check status
git status

# If no recent changes, reset to last known good commit
git log --oneline | head -10
git reset --hard <good-commit>

# Rebuild derived outputs
atlas compile --target generic
```

Option C: Re-initialize and restore Goals only
```bash
# Create fresh project
atlas init <project> --profile profile.json --non-interactive

# Restore only .ai/ from backup
tar -xzf backup-20260815.tar.gz --strip-components=1 \
  -C <project> project/.ai/

# Verify and recompile
atlas doctor <project>
atlas compile --target generic --path <project>
```

### Scenario 3: Lost Adapter Files

**Symptom:**
```bash
ls <project>/.codex/
# (empty or missing directory)

atlas compile --target codex
# ✗ Failed to write output
```

**Recovery:**

Simple: Just recompile

```bash
# Remove stale cache
rm -rf <project>/.atlas/runtime/

# Recompile
atlas compile --target codex --path <project>

# Verify
ls <project>/.codex/agents/ | wc -l  # Should have files
atlas doctor <project>
```

### Scenario 4: Lock Digest Mismatch

**Symptom:**
```
atlas doctor <project>
# ✗ Goal P00-G01: Lock digest mismatch (locked but content changed)
```

**Cause:** Goal was edited after being locked (data integrity issue)

**Recovery:**

Option A: Audit and accept change
```bash
# Review what changed
git diff <project>/.ai/goals/P00-G01.goal.json

# If intentional, update lock:
# (future feature: atlas update-lock <goal>)
# For now, manually update lock field in .goal.json

# Or revert to locked version
git checkout HEAD -- <project>/.ai/goals/P00-G01.goal.json
```

Option B: Restore from backup (if unintentional change)
```bash
tar -xzf backup-20260815.tar.gz \
  project/.ai/goals/P00-G01.goal.json

atlas doctor <project>
```

---

## Git-Native Recovery

Project Atlas is Git-first. Use Git for all recovery:

### View Goal History

```bash
# See all changes to a Goal
git log -p -- <project>/.ai/goals/P00-G01.goal.json | head -50

# See who changed it
git log --oneline --follow -- <project>/.ai/goals/P00-G01.goal.json

# See exactly what changed
git show <commit>:<project>/.ai/goals/P00-G01.goal.json
```

### Recover Deleted Goal

```bash
# Find when it was deleted
git log -p -- <project>/.ai/goals/P00-G01.goal.json | grep "delete\|remove"

# Restore from commit before deletion
git checkout <commit-before-delete> -- <project>/.ai/goals/P00-G01.goal.json

# Verify and re-add
atlas doctor <project>
git add <project>/.ai/goals/P00-G01.goal.json
git commit -m "Restore Goal P00-G01"
```

### Audit Trail

```bash
# See all Goal amendments (state transitions + edits)
git log --all --grep="Goal\|amendment" -- <project>/.ai/goals/

# Filter by date
git log --since="2 weeks ago" --until="1 week ago" -- <project>/.ai/goals/

# Generate audit report
git log --format="%h %ai %s" -- <project>/.ai/ > audit.txt
```

---

## Rollback Procedures

### Rollback a Goal Amendment

```bash
# Last amendment to Goal P00-G01
git show HEAD:<project>/.ai/goals/P00-G01.goal.json | grep -A5 "history"

# If last change was unintended, revert
git revert --no-commit <commit-hash>

# Or reset to previous state
git reset --soft HEAD~1
git restore -- <project>/.ai/goals/P00-G01.goal.json

# Verify
atlas doctor <project>
```

### Rollback Project to Earlier Version

```bash
# See all project changes
git log --oneline <project>/

# Rollback to specific commit
git reset --hard <commit-hash>

# Or create a new branch at that point
git checkout <commit-hash> -- <project>/

# Verify health after rollback
atlas doctor <project>
```

### Rollback Adapter Workspace (e.g., .codex)

Adapter workspaces are derived—just delete and rebuild:

```bash
# Remove outdated workspace
rm -rf <project>/.codex

# Recompile
atlas compile --target codex --path <project>

# Verify
atlas doctor <project>
```

---

## Corruption Detection

### Automated Checks

```bash
# Run framework check
atlas framework-check

# Run project doctor (detects schema violations, orphaned files)
atlas doctor <project>

# Verify all Goals are valid JSON
for f in <project>/.ai/goals/*.goal.json; do
  python -m json.tool "$f" > /dev/null || echo "✗ $f"
done

# Check for orphaned files
git status <project>/ | grep "deleted\|modified"
```

### Manual Inspection

```bash
# Check .ai directory structure
tree <project>/.ai/

# Expected:
# ├── manifest.json (project metadata)
# ├── goals/
# │   ├── P00-G01.goal.json
# │   └── ...
# ├── agents/
# │   └── manifest.json
# └── skills/
#     └── manifest.json

# Validate manifest
jq . <project>/.ai/manifest.json

# Count Goals
ls <project>/.ai/goals/*.goal.json | wc -l

# Check for .gitignore violations
grep "\.ai" <project>/.gitignore  # Should NOT ignore .ai/
```

---

## Disaster Recovery Drill

Test your recovery procedure monthly:

```bash
#!/bin/bash
# test_recovery.sh

set -e

PROJECT="my-project"
BACKUP="backup-latest.tar.gz"

echo "=== Starting Recovery Drill ==="

# 1. Create restore directory
mkdir -p /tmp/recovery_test
cd /tmp/recovery_test

# 2. Restore from backup
tar -xzf ../../$BACKUP

# 3. Verify structure
test -d "$PROJECT/.ai/goals/" || exit 1
test -f "$PROJECT/.ai/manifest.json" || exit 1

# 4. Run health check
atlas doctor "$PROJECT" || exit 1

# 5. Recompile
atlas compile --target generic --path "$PROJECT" || exit 1

# 6. Verify output
test -d "$PROJECT/.atlas/runtime/" || exit 1

echo "=== Recovery Drill PASSED ==="
exit 0
```

Run drill:

```bash
chmod +x test_recovery.sh
./test_recovery.sh && echo "✓ Recovery procedure validated"
```

---

## Data Retention & Compliance

### Retention Policy

```
Daily backups:   Keep 30 days
Weekly backups:  Keep 1 year
Audit logs:      Keep 7 years (if applicable)
```

### Deletion Procedure

```bash
# Delete backups older than 30 days
find . -name "backup-*.tar.gz" -mtime +30 -delete

# Verify Git history is preserved (not deleted)
git log --all | head -5

# Securely delete sensitive backups
shred -vfz -n 3 backup-sensitive-20260815.tar.gz
```

---

## See Also

- [Deployment](./deployment.md) — Installation and upgrades
- [Monitoring](./monitoring.md) — Health checks and alerting
- [Incident Response](./incident-response.md) — Failure playbooks
