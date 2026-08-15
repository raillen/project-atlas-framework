# Deployment Runbook

Guide for deploying Project Atlas Framework to production environments.

## Prerequisites

- Python 3.10+ (test 3.10, 3.11, 3.12)
- pip ≥ 22.0
- Git 2.30+
- 200MB disk space minimum
- Network access to PyPI

## Installation Methods

### Option 1: From PyPI (Recommended for users)

```bash
pip install project-atlas-framework
atlas --version
```

**Verification:**
```bash
atlas framework-check
```

Expected output:
```
Project Atlas Framework v0.3.0
✓ Schemas: 27 valid
✓ Adapters: 7 ready
✓ Workforce: 118 skills, 26 agents, 14 recipes
✓ No errors detected
```

### Option 2: From Source (Recommended for development)

```bash
git clone https://github.com/your-org/project-atlas-framework.git
cd project-atlas-framework

# Install in development mode
pip install -e .

# Install dev dependencies (optional)
pip install -e '.[dev]'

# Verify installation
atlas framework-check
pytest
```

### Option 3: Docker (Production container)

```dockerfile
FROM python:3.12-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir project-atlas-framework

ENTRYPOINT ["atlas"]
```

Build and run:
```bash
docker build -t atlas:v0.3.0 .
docker run --rm atlas framework-check
```

---

## Configuration

### Project Initialization

Create a new Project Atlas project:

```bash
atlas init my-project --profile profile.json --non-interactive
cd my-project
atlas doctor
```

### Profile Configuration

`project-profile.json`:

```json
{
  "version": 3,
  "protocol": {"version": 3, "compatible": ">=3 <4"},
  "project": {
    "name": "my-project",
    "type": ["web-saas"]
  },
  "stack": {
    "languages": ["python", "typescript"],
    "frameworks": ["fastapi", "react"]
  },
  "features": ["api", "web-ui", "database"],
  "quality": {
    "coverage_target": 80,
    "type_check": true,
    "lint": true
  },
  "ai": {
    "orchestrator": "native",
    "autonomy": "agentic",
    "preferred_models": [
      {"id": "claude/3.5-sonnet", "provider": "anthropic"},
      {"id": "gpt-4o", "provider": "openai"}
    ]
  }
}
```

### Environment Variables

Currently, Atlas uses no mandatory environment variables. Reserved for future:

```bash
ATLAS_CONFIG_DIR      # Project config directory (default: ./.atlas)
ATLAS_LOG_LEVEL       # Logging level (default: INFO)
ATLAS_CACHE_DIR       # Cache directory (default: ~/.cache/atlas)
```

---

## Deployment Checklist

Before deploying to production:

- [ ] Python version tested (3.10, 3.11, 3.12)
- [ ] `pip install -e '.[dev]'` successful
- [ ] `pytest` passes (all tests)
- [ ] `atlas framework-check` passes
- [ ] `atlas doctor <project>` shows HEALTHY
- [ ] Git history clean (`git status`)
- [ ] No uncommitted changes in `.ai/` directory
- [ ] All examples compile: `atlas compile --target generic`
- [ ] CI/CD gates pass (coverage, types, security)

---

## Upgrade Strategy

### Minor/Patch Upgrades (v0.3.0 → v0.3.1)

Safe, automatic:

```bash
pip install --upgrade project-atlas-framework
atlas framework-check  # Verify
atlas doctor <project> # Re-validate
```

### Major Upgrades (v0.2 → v0.3)

Requires migration:

```bash
# Backup existing project
cp -r my-project my-project.backup

# Run migration
atlas migrate my-project

# Verify
atlas doctor my-project

# Review changes in git
git diff

# Commit if satisfied
git add -A && git commit -m "Migrate to v0.3"
```

### Rollback Procedure

If upgrade fails:

```bash
# Option 1: Restore from backup
rm -rf my-project
cp -r my-project.backup my-project

# Option 2: Downgrade package
pip install project-atlas-framework==0.2.0

# Option 3: Git rollback (if changes committed)
git reset --hard <previous-commit>
```

---

## Performance Tuning

### Project Size Scaling

| Metric | Small | Medium | Large |
|--------|-------|--------|-------|
| Goals | 1-10 | 11-50 | 50+ |
| Agents | 5-10 | 10-20 | 20+ |
| Skills | 30-50 | 50-100 | 100+ |
| Commands | <100ms | <500ms | <2s |

### Optimization Tips

1. **Lazy loading:** Large projects load Goals on-demand, not upfront
2. **Caching:** Compiled outputs cached in `.atlas/runtime/`
3. **Parallel compilation:** Can compile multiple targets simultaneously
4. **Incremental validation:** Only changed Goals validated on `atlas doctor`

---

## Troubleshooting

### "Module not found: project_atlas"

```bash
# Reinstall with dev dependencies
pip install -e '.[dev]'

# Or check Python path
python -c "import project_atlas; print(project_atlas.__file__)"
```

### "atlas: command not found"

```bash
# Verify pip location
pip show project-atlas-framework | grep Location

# Reinstall
pip uninstall project-atlas-framework
pip install project-atlas-framework
```

### "Invalid project structure"

```bash
# Run doctor to diagnose
atlas doctor <project>

# Re-initialize if corrupted
atlas init <project> --profile <profile> --non-interactive
```

### Performance degradation

```bash
# Clear cache
rm -rf <project>/.atlas/runtime/

# Re-compile
atlas compile --target generic --path <project>

# Check disk space
df -h
```

---

## Monitoring

### Health Checks

Periodically verify installation:

```bash
# Daily health check
atlas framework-check

# Project-specific validation
atlas doctor <project>

# Re-compile to detect adapter issues
atlas compile --target codex --path <project>
```

### Logs

Atlas currently logs to stdout/stderr. Capture with:

```bash
# Redirect to file
atlas doctor <project> > doctor.log 2>&1

# Monitor in real-time
atlas compile --target generic --path <project> | tee compile.log
```

---

## Production Checklist

For SaaS/production deployments:

- [ ] Separate config from code (`profile.json` in secrets manager)
- [ ] Backup Goals regularly: `cp -r <project>/.ai/goals <backup>/`
- [ ] Archive compiled outputs: `.atlas/runtime/` → S3/storage
- [ ] Monitor CI/CD gates in pipeline
- [ ] Enable audit logging (Goal amendments, state transitions)
- [ ] Document recovery procedures for team
- [ ] Test rollback procedure monthly

---

## See Also

- [Monitoring](./monitoring.md) — Health metrics and alerts
- [Recovery](./recovery.md) — Disaster recovery and rollback
- [Incident Response](./incident-response.md) — Troubleshooting playbooks
