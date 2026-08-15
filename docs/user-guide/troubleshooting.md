# Troubleshooting Guide & FAQ

## Common Errors

### E001: Module Not Found

**Message:** `ModuleNotFoundError: No module named 'project_atlas'`

**Causes:**
- Atlas not installed
- Wrong Python environment
- Virtual environment not activated

**Solutions:**

```bash
# Install or reinstall
pip install project-atlas-framework

# Activate virtual environment (if using one)
source venv/bin/activate

# Verify installation
python -c "import project_atlas; print(project_atlas.__version__)"
```

---

### E002: Command Not Found

**Message:** `atlas: command not found`

**Causes:**
- Atlas not installed in PATH
- Shell cache not refreshed

**Solutions:**

```bash
# Reinstall
pip uninstall -y project-atlas-framework
pip install project-atlas-framework

# Refresh shell
exec bash  # or: source ~/.bashrc

# Verify
atlas --version
```

---

### E003: Invalid Project Structure

**Message:** `✗ Invalid project structure`

**Causes:**
- Missing `.ai/` directory
- Corrupted `.ai/manifest.json`
- Missing required Goal files

**Solutions:**

```bash
# Diagnose
atlas doctor <project>

# Option 1: Restore from backup
tar -xzf backup.tar.gz

# Option 2: Reset Git
git reset --hard HEAD

# Option 3: Re-initialize
atlas init <project> --profile profile.json --non-interactive
```

See [Recovery Procedures](./operations/recovery.md) for more.

---

### E004: Goal State Transition Invalid

**Message:** `Invalid goal transition: DRAFT → EXECUTING`

**Causes:**
- Attempted direct transition not allowed by state machine
- Goal in BLOCKED state

**Solutions:**

```bash
# Check current state
jq .state <project>/.ai/goals/P00-G01.goal.json

# Follow valid path:
# DRAFT → PLANNED → LOCKED → EXECUTING → VERIFYING → REVIEWING → DONE

atlas goal update P00-G01 --state PLANNED
atlas goal update P00-G01 --state LOCKED
atlas goal update P00-G01 --state EXECUTING
```

See [Goal State Machine](./user-guide/goals-101.md#state-machine) for diagram.

---

### E005: Lock Digest Mismatch

**Message:** `Goal P00-G01: Lock digest mismatch (locked but content changed)`

**Causes:**
- Goal manually edited after being locked
- Amendment not applied through CLI

**Solutions:**

```bash
# Option 1: Accept the change (new amendment)
atlas goal update P00-G01 --amend "Updated requirements"
atlas goal update P00-G01 --state LOCKED

# Option 2: Revert to locked version
git checkout HEAD~1 -- <project>/.ai/goals/P00-G01.goal.json

# Option 3: Use Git to see what changed
git diff HEAD -- <project>/.ai/goals/P00-G01.goal.json
```

**Prevention:** Always use CLI for changes: `atlas goal update`, not manual edits.

---

### E006: Compile Failed

**Message:** `atlas compile --target codex: Failed to write output`

**Causes:**
- Adapter crash or incompatible Goal structure
- Disk space full
- Corrupted cache

**Solutions:**

```bash
# Clear cache
rm -rf <project>/.atlas/runtime/

# Rebuild
atlas compile --target generic --path <project>

# Check disk space
df -h

# Try alternative target if specific target fails
atlas compile --target claude-code --path <project>  # May work
```

---

### E007: Doctor Shows BROKEN

**Message:** `✗ Overall: BROKEN`

**Causes:**
- Multiple errors in project structure
- Orphaned files or references

**Solutions:**

```bash
# Get full diagnostic output
atlas doctor <project> > diagnostics.log

# Review each error in diagnostics.log

# For each error:
# - E002 (Corrupted Goal): Restore from backup or Git
# - E003 (Orphaned Skill): Remove from manifest or add missing skill
# - E004 (Invalid JSON): Repair or restore

# After fixes, re-run
atlas doctor <project>
```

---

### E008: Permission Denied

**Message:** `PermissionError: [Errno 13] Permission denied: '<project>/.ai/manifest.json'`

**Causes:**
- File ownership issue
- Read-only file system
- Insufficient permissions

**Solutions:**

```bash
# Fix file permissions
chmod 644 <project>/.ai/manifest.json
chmod 755 <project>/.ai/goals/

# Or fix entire project
chmod -R u+w <project>

# Check file ownership
ls -la <project>/.ai/manifest.json

# If wrong owner, change it
chown -R $(whoami) <project>

# Check if filesystem is read-only
touch <project>/.test && rm <project>/.test
```

---

### E009: Cache Corruption

**Message:** Slow commands, unexpected behavior after compilation

**Causes:**
- Stale runtime cache
- Incomplete compilation
- Disk space issues

**Solutions:**

```bash
# Clear all caches
rm -rf <project>/.atlas/runtime/
rm -rf <project>/.codex/
rm -rf <project>/.claude*

# Recompile
atlas compile --target generic --path <project>

# Verify
atlas doctor <project>
```

---

### E010: Git Integration Issues

**Message:** `fatal: not a git repository` or `git command failed`

**Causes:**
- Project not in Git repository
- Git not installed
- Corrupted .git directory

**Solutions:**

```bash
# Initialize Git (if needed)
cd <project>
git init
git config user.email "you@example.com"
git config user.name "Your Name"
git add .
git commit -m "Initial commit"

# Repair corrupted repository
git fsck --full

# Verify Git works
git log --oneline | head -5
```

---

## FAQ

### Q: How do I start a new project?

**A:**
```bash
atlas init my-project --profile profile.json --non-interactive
cd my-project
atlas doctor my-project  # Should show HEALTHY
```

See [Goals 101](./user-guide/goals-101.md) for step-by-step guide.

---

### Q: How do I change a Goal's state?

**A:**
```bash
# Check current state
atlas goal list

# Transition to next valid state
atlas goal update <goal-id> --state <new-state>

# Valid transitions follow state machine (see E004)
```

---

### Q: Can I edit Goals with a text editor?

**A:**
Not recommended. Goal files are locked after LOCKED state, and manual edits will cause digest mismatches.

**Always use CLI:**
```bash
atlas goal update <goal-id> --title "New Title"
atlas goal update <goal-id> --amend "Change reason"
```

---

### Q: What happens if I accidentally delete a Goal file?

**A:**

Option 1: Restore from Git
```bash
git checkout HEAD -- <project>/.ai/goals/P00-G01.goal.json
```

Option 2: Restore from backup
```bash
tar -xzf backup.tar.gz project/.ai/goals/P00-G01.goal.json
```

Option 3: Re-create manually (if very small project)
```bash
# Recreate in CLI
atlas goal new --title "..." --priority P0
```

See [Recovery Procedures](./operations/recovery.md).

---

### Q: How do I back up my project?

**A:**
```bash
# Quick backup
tar --exclude='.atlas/runtime' --exclude='.codex' -czf backup.tar.gz <project>/

# To cloud storage (e.g., S3)
aws s3 sync <project>/.ai/ s3://my-bucket/backups/

# Automated daily backup (add to crontab)
0 9 * * * tar -czf ~/backups/backup-$(date +\%Y\%m\%d).tar.gz <project>/
```

See [Backup Strategy](./operations/recovery.md#backup-strategy).

---

### Q: How do I migrate from v0.2 to v0.3?

**A:**
```bash
# Automatic migration
atlas migrate <project>

# Verify
atlas doctor <project>

# Commit changes
git add <project>/atlas.json
git commit -m "Migrate to v0.3"
```

---

### Q: Which compile target should I use?

**A:**

| Target | Use Case | Output |
|--------|----------|--------|
| `generic` | Default, most compatible | `.atlas/runtime/compiled/generic/` |
| `codex` | GitHub Copilot integration | `.codex/agents/`, `.codex/skills/` |
| `claude-code` | Claude Code extended mode | `.claude-code/agents/`, `.claude-code/skills/` |
| `claude` | Claude standard | `.atlas/runtime/compiled/claude/` |
| `chatgpt` | ChatGPT | `.atlas/runtime/compiled/chatgpt/` |
| `kimi` | Kimi | `.atlas/runtime/compiled/kimi/` |
| `traycer` | Traycer | `.traycer/PROJECT_ATLAS.md` |

**Default: `generic`** (most portable)

```bash
atlas compile --target generic --path <project>
```

---

### Q: How do I check if my project is healthy?

**A:**
```bash
atlas doctor <project>

# Expected output:
# ✓ atlas.json: valid
# ✓ Goals: 5 valid, 0 broken
# ✓ Agents: 3 resolved
# ✓ Skills: 12 resolved
# ✓ Recipes: 2 valid
# ✓ Policies: all consistent
# ✓ Overall: HEALTHY
```

---

### Q: What do I do if doctor shows BROKEN?

**A:**
```bash
# Get detailed diagnostics
atlas doctor <project> > diagnostics.log
cat diagnostics.log

# Find error (e.g., "✗ Goal P00-G01: Invalid JSON")

# Restore from backup or Git
tar -xzf backup.tar.gz
# or
git reset --hard HEAD

# Re-run doctor
atlas doctor <project>
```

See [Playbook 2: Corrupted Project](./operations/incident-response.md#playbook-2-corrupted-project).

---

### Q: How do I compile for multiple targets?

**A:**
```bash
# Compile all targets
for target in generic codex claude-code claude chatgpt kimi traycer; do
  atlas compile --target $target --path <project>
done

# Or use a script
atlas compile --target generic --path <project>
atlas compile --target codex --path <project>
atlas compile --target claude-code --path <project>
# ... etc
```

---

### Q: Can I share my project with others?

**A:**
```bash
# Push to Git (recommended)
git add <project>/
git commit -m "Share project"
git push origin main

# Others can clone:
git clone <repo>
cd my-project
atlas doctor .  # Validate
```

**Best practice:** Use `.gitignore` to exclude compiled outputs
```bash
# In <project>/.gitignore
/.atlas/runtime/
/.codex/
/.claude-code/
/.claude/
/.chatgpt/
/.kimi/
/.traycer/
```

---

## Performance Tips

### Q: Compilation is slow. How do I speed it up?

**A:**

1. **Clear cache:**
   ```bash
   rm -rf <project>/.atlas/runtime/
   ```

2. **Check disk space:**
   ```bash
   df -h
   # Need at least 1GB free
   ```

3. **Reduce Goal count for testing:**
   - Create a minimal test project with 1-2 Goals
   - Verify compilation works
   - Gradually increase

4. **Compile one target at a time:**
   ```bash
   atlas compile --target generic --path <project>
   # Wait for completion before next target
   ```

---

### Q: Doctor is slow. How do I speed it up?

**A:**

1. **First run is slowest** (builds schema cache)
   ```bash
   atlas doctor <project>  # Slow first time
   atlas doctor <project>  # Faster next time
   ```

2. **Large projects (100+ Goals):**
   - Doctor validates all Goals
   - Normal behavior, not a bug
   - Can take 5-30s depending on project size

3. **Check system resources:**
   ```bash
   free -h       # Memory
   df -h         # Disk space
   top -b -n1    # CPU usage
   ```

---

## Getting Help

### Where to find information

- **[Goals 101](./user-guide/goals-101.md)** — Getting started with Goals
- **[CLI Reference](./reference/cli.md)** — All commands and options
- **[Compiler Architecture](./development/compiler.md)** — How compilation works
- **[Operations Runbooks](./operations/)** — Deployment, monitoring, recovery
- **[GitHub Issues](https://github.com/your-org/project-atlas-framework/issues)** — Report bugs

### Reporting Bugs

When reporting an issue, include:

1. **Atlas version:**
   ```bash
   atlas --version
   ```

2. **Python version:**
   ```bash
   python --version
   ```

3. **Full error message:**
   ```bash
   atlas doctor <project> 2>&1 | tee error.log
   ```

4. **Reproduction steps**

5. **Expected vs. actual behavior**

---

## See Also

- [Incident Response](./operations/incident-response.md) — Playbooks for common failures
- [Recovery Procedures](./operations/recovery.md) — Data restoration
- [Monitoring](./operations/monitoring.md) — Health checks
- [CLI Reference](./reference/cli.md) — All commands
