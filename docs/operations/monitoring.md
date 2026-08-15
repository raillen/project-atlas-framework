# Monitoring Guide

Observability and health monitoring for Project Atlas deployments.

## Health Checks

### Framework Health

```bash
# Quick framework validation
atlas framework-check

# Output:
# Project Atlas Framework v0.3.0
# ✓ Schemas: 27 valid
# ✓ Adapters: 7 ready
# ✓ Workforce: 118 skills, 26 agents, 14 recipes
# ✓ No errors detected
```

**Frequency:** On startup, daily in production

**Success criteria:** All 4 lines ✓

### Project Health

```bash
# Comprehensive project diagnostics
atlas doctor <project>

# With JSON output for parsing
atlas doctor <project> --json
```

**Output checklist:**
- ✓ atlas.json: valid
- ✓ Goals: N valid, 0 broken
- ✓ Agents: N resolved
- ✓ Skills: N resolved
- ✓ Recipes: N valid
- ✓ Policies: all consistent
- ✓ Overall: HEALTHY

**Frequency:** After each Goal state change, before deployment

---

## Key Metrics

### CLI Performance

```bash
# Measure command execution time
time atlas compile --target generic --path <project>

# Expected:
# real: <2s for small projects
# user: <1s
```

| Command | Expected Time | Alert Threshold |
|---------|---------------|-----------------|
| `atlas doctor` | <500ms | >2s |
| `atlas compile generic` | <1s | >3s |
| `atlas compile codex` | <500ms | >2s |
| `atlas goal new` | <100ms | >500ms |
| `atlas goal list` | <100ms | >500ms |

### Disk Usage

```bash
# Monitor project size
du -sh <project>

# Breakdown
du -sh <project>/.ai/*
du -sh <project>/.atlas/*
du -sh <project>/.codex/*
du -sh <project>/.claude/*
```

**Typical sizes:**
- `.ai/goals/`: 50-200KB (N goals)
- `.ai/agents/`: 10-50KB (manifest)
- `.ai/skills/`: 10-50KB (manifest)
- `.atlas/runtime/`: 50-500KB (compiled outputs)
- `.codex/` or `.claude/`: 100-1MB (full workspace trees)

**Alert if:**
- Project > 100MB (investigate orphaned files)
- Runtime cache > 50MB (rebuild with `rm -rf .atlas/runtime/`)

### Coverage Metrics

```bash
# After running tests with coverage
pytest --cov=<src> --cov-report=term-missing

# Check coverage trend
pytest --cov=<src> --cov-report=html
open htmlcov/index.html
```

**Goals:**
- CLI coverage: >75%
- Core modules: >80%
- Schemas: >90% (stable)

---

## Health Monitoring Checklist

### Daily (Automated)

- [ ] `atlas framework-check` passes
- [ ] No errors in logs (stdout/stderr)
- [ ] Project size within limits (<100MB)
- [ ] Disk space > 1GB free

### Weekly (Manual)

- [ ] `atlas doctor <project>` = HEALTHY
- [ ] Example projects compile successfully
- [ ] All adapters generate valid output
- [ ] Git history clean (no orphaned files)

### Monthly (Full Audit)

- [ ] Run full test suite: `pytest`
- [ ] Run CI/CD checks locally:
  ```bash
  pytest --cov
  mypy src/project_atlas
  ruff check src tests
  pip-audit
  ```
- [ ] Rebuild all compiler targets:
  ```bash
  for target in generic codex claude-code claude chatgpt kimi traycer; do
    atlas compile --target $target && echo "✓ $target"
  done
  ```
- [ ] Review Goal state transitions (amendments, locks)
- [ ] Validate all examples migrate and compile

---

## Logging & Debugging

### Verbose Output

```bash
# Most commands support output capture
atlas doctor <project> > diagnostics.log 2>&1

# For debugging, redirect stderr
atlas compile --target generic 2> compile_errors.log
```

### Audit Trail

Goals automatically record state transitions:

```bash
# View Goal history
cat <project>/.ai/goals/P00-G01.goal.json | grep -A5 '"history"'

# Expected format:
# "history": [
#   {"at": "2026-08-15T...", "event": "created", "state": "DRAFT"},
#   {"at": "2026-08-15T...", "event": "transitioned", "state": "PLANNED", "reason": "..."},
#   {"at": "2026-08-15T...", "event": "amended", "revision": 2, "digest": "..."}
# ]
```

### Common Issues to Monitor

1. **Lock digest mismatches** → Goal was edited after locking
   ```bash
   atlas verify-locks <project>  # (future feature)
   ```

2. **Orphaned skills/agents** → Manifest lists non-existent IDs
   ```bash
   atlas doctor <project>  # Detects and reports
   ```

3. **Missing adapter files** → Compilation failed
   ```bash
   ls -la <project>/.codex/  # Should be non-empty
   ```

4. **Stale cache** → Old compiled outputs
   ```bash
   rm -rf <project>/.atlas/runtime/
   atlas compile --target generic
   ```

---

## Alerting Strategy

### Automatic Alerts

Set up monitoring with your observability platform:

```yaml
alerts:
  - name: AtlasFrameworkUnhealthy
    condition: atlas framework-check exit != 0
    action: notify_ops

  - name: ProjectUnhealthy
    condition: atlas doctor <project> doesn't contain "HEALTHY"
    action: notify_team

  - name: CompileFailure
    condition: atlas compile exit != 0
    action: notify_dev

  - name: DiskSpaceWarning
    condition: <project> size > 50MB
    action: notify_ops

  - name: TestCoverageDrop
    condition: coverage < 80%
    action: block_merge
```

### Manual Monitoring

Daily checklist:

```bash
#!/bin/bash
# daily_check.sh

echo "=== Framework Health ==="
atlas framework-check || exit 1

echo "=== Project Health ==="
atlas doctor my-project || exit 1

echo "=== Performance ==="
time atlas compile --target generic --path my-project || exit 1

echo "=== All checks passed ==="
exit 0
```

Run with cron:

```bash
# Add to crontab
0 9 * * * /path/to/daily_check.sh >> /var/log/atlas.log 2>&1
```

---

## Performance Dashboards

### Metrics to Track

```
atlas.framework.health        [0=unhealthy, 1=healthy]
atlas.project.health          [0=unhealthy, 1=healthy]
atlas.compile.duration_ms     [latency per target]
atlas.project.size_bytes      [growth over time]
atlas.tests.pass_rate         [% of tests passing]
atlas.coverage_percent        [code coverage %]
```

### Example Grafana Dashboard

```json
{
  "panels": [
    {
      "title": "Framework Health",
      "targets": [{"expr": "atlas_framework_health"}]
    },
    {
      "title": "Compile Time (ms)",
      "targets": [{"expr": "atlas_compile_duration_ms"}]
    },
    {
      "title": "Project Size (MB)",
      "targets": [{"expr": "atlas_project_size_bytes / 1e6"}]
    },
    {
      "title": "Test Coverage %",
      "targets": [{"expr": "atlas_coverage_percent"}]
    }
  ]
}
```

---

## Troubleshooting via Monitoring

### Symptom: Slow compilation

**Check:**
```bash
du -sh <project>/.atlas/runtime/
ls -la <project>/.codex/ | wc -l  # Number of agent/skill files
```

**Action:** Clear cache, rebuild
```bash
rm -rf <project>/.atlas/runtime/
atlas compile --target generic
```

### Symptom: Inconsistent state

**Check:**
```bash
atlas doctor <project>
cat <project>/.ai/goals/*.goal.json | grep "lock" | wc -l
```

**Action:** Re-validate Goals
```bash
atlas doctor <project> --json > state.json
# Review state.json for anomalies
```

### Symptom: Disk space growing

**Check:**
```bash
du -sh <project>/.atlas/*
git status  # Check for untracked files
```

**Action:** Clean temporary files
```bash
rm -rf <project>/.atlas/snapshots/*  # Old snapshots
rm -rf <project>/.atlas/runtime/*     # Rebuild if needed
git clean -fd <project>               # Remove untracked
```

---

## See Also

- [Deployment](./deployment.md) — Installation and configuration
- [Recovery](./recovery.md) — Disaster recovery
- [Incident Response](./incident-response.md) — Troubleshooting playbooks
