# CLI Reference

Complete command reference for the `atlas` command-line tool.

## Global Options

```bash
atlas [--version] [--help]
```

| Option | Description |
|--------|-------------|
| `--version` | Show Atlas version |
| `--help` | Show help for main command |

---

## init

Initialize a new Project Atlas project.

```bash
atlas init <path> [--profile <profile_path>] [--non-interactive]
```

| Option | Description | Example |
|--------|-------------|---------|
| `<path>` | Project directory (required) | `./my-project` |
| `--profile` | Path to profile JSON | `./project-profile.json` |
| `--non-interactive` | Skip prompts (requires `--profile`) | `--non-interactive` |

**Examples:**

```bash
# Interactive mode (prompts for name, type, models)
atlas init my-project

# Non-interactive with profile
atlas init my-project --profile examples/brasa/project-profile.json --non-interactive
```

**Output:**
```
Initialized Project Atlas v0.3 in /path/to/my-project
Agents: 19 | Skills: 80 | Recipes: 7
```

---

## validate

Validate a project's `atlas.json` and structure.

```bash
atlas validate [<path>]
```

| Option | Description |
|--------|-------------|
| `<path>` | Project path (default: current) |

**Examples:**

```bash
atlas validate
atlas validate ./my-project
```

---

## goal

Manage Goals (objectives with acceptance criteria and state machine).

### goal new

Create a new Goal.

```bash
atlas goal new <id> <title> --phase <phase> [--path <path>]
```

| Option | Description |
|--------|-------------|
| `<id>` | Goal ID (e.g., `P00-G01`) |
| `<title>` | Goal title |
| `--phase` | Phase identifier (e.g., `P00`, `P01`) |
| `--path` | Project path (default: current) |

**Example:**

```bash
atlas goal new P00-G01 "Setup CI/CD pipeline" --phase P00
```

**Output:**
```
Created Goal P00-G01 in /path/.ai/goals/P00-G01.goal.json
```

### goal list

List all Goals in the project.

```bash
atlas goal list [--path <path>]
```

**Example:**

```bash
atlas goal list
```

**Output:**
```
Phase P00 — Foundation
  P00-G01  Setup CI/CD pipeline                    [DRAFT]
  P00-G02  Core authentication                     [PLANNED]

Phase P01 — Features
  P01-G01  User profiles                           [LOCKED]
```

### goal state

Transition a Goal to a new state.

```bash
atlas goal state <id> <new_state> [--reason <reason>] [--path <path>]
```

| Option | Description |
|--------|-------------|
| `<id>` | Goal ID |
| `<new_state>` | Target state (PLANNED, LOCKED, EXECUTING, VERIFYING, REVIEWING, DONE) |
| `--reason` | Transition reason (optional) |
| `--path` | Project path (default: current) |

**Valid transitions:**

```
DRAFT → PLANNED or BLOCKED
PLANNED → LOCKED, DRAFT, or BLOCKED
LOCKED → EXECUTING or BLOCKED
EXECUTING → VERIFYING or BLOCKED
VERIFYING → REVIEWING, EXECUTING, or BLOCKED
REVIEWING → DONE, EXECUTING, or BLOCKED
BLOCKED → PLANNED, LOCKED, or EXECUTING
DONE → (no further transitions)
```

**Example:**

```bash
atlas goal state P00-G01 PLANNED --reason "Details finalized"
atlas goal state P00-G01 LOCKED
```

**Output:**
```
Goal P00-G01 transitioned to PLANNED
Goal P00-G01 transitioned to LOCKED
```

### goal amend

Amend a locked Goal with formal changes.

```bash
atlas goal amend <id> --file <amendment_file> [--reason <reason>] [--approved-by <person>] [--path <path>]
```

| Option | Description |
|--------|-------------|
| `<id>` | Goal ID |
| `--file` | Amendment JSON file |
| `--reason` | Reason for amendment |
| `--approved-by` | Approver email/name |
| `--path` | Project path (default: current) |

**Amendment file format:**

```json
{
  "id": "amendment-1",
  "goal_id": "P00-G01",
  "changes": {
    "add_acceptance": ["New criterion"],
    "remove_acceptance": ["Old criterion"],
    "add_constraint": ["New constraint"]
  },
  "reason": "Stakeholder request",
  "approved_by": "pm@example.com"
}
```

**Example:**

```bash
atlas goal amend P00-G01 --file amendment.json --reason "Scope expanded"
```

---

## context

Manage project context and context planning.

### context plan

Plan what context to include for a task.

```bash
atlas context plan <description> [--path <path>] [--json]
```

| Option | Description |
|--------|-------------|
| `<description>` | Task description |
| `--path` | Project path (default: current) |
| `--json` | Output as JSON |

**Example:**

```bash
atlas context plan "add database migration"
atlas context plan "add database migration" --json
```

**Output (text):**
```
Context Plan for "add database migration"
├─ LPC Budget: medium
├─ Max depth: 1
├─ Suggested items:
│  ├─ active Goal
│  ├─ ENTRYPOINT.md
│  ├─ docs/ATLAS.md
│  └─ relevant schema
```

---

## compile

Compile project to target adapter formats.

```bash
atlas compile --target <target> [--path <path>]
```

| Option | Description |
|--------|-------------|
| `--target` | Target adapter (required) |
| `--path` | Project path (default: current) |

**Supported targets:**

| Target | Use Case |
|--------|----------|
| `generic` | Provider-agnostic runtime (default) |
| `codex` | GitHub Copilot Chat |
| `claude-code` | Claude Code tool |
| `claude` | Raw Claude API |
| `chatgpt` | OpenAI ChatGPT |
| `kimi` | Moonshot Kimi LLM |
| `traycer` | Anthropic trace runtime |

**Examples:**

```bash
atlas compile --target generic
atlas compile --target codex --path ./my-project
atlas compile --target claude-code
```

**Output:**
```
Compiled to [.codex/ | .claude/ | .atlas/runtime/compiled/{target}/]
Created X files
```

---

## report

Manage task reports and project intelligence.

### report add

Add a task report to project intelligence.

```bash
atlas report add <report_file> [--path <path>]
```

| Option | Description |
|--------|-------------|
| `<report_file>` | Path to report JSON |
| `--path` | Project path (default: current) |

**Report file format:**

```json
{
  "id": "TASK-1",
  "status": "success",
  "type": "test",
  "tokens": {
    "input": 1000,
    "output": 200,
    "cached": 0
  },
  "cost": {
    "direct": {
      "amount": 0.12,
      "currency": "USD",
      "provenance": "observed"
    }
  }
}
```

**Example:**

```bash
atlas report add report.json
```

### report summary

Show project intelligence summary.

```bash
atlas report summary [--path <path>]
```

**Example:**

```bash
atlas report summary
```

**Output:**
```
Project Intelligence Summary
├─ Total tasks: 42
├─ Success rate: 95%
├─ Avg tokens/task: 1250
├─ Total spend: $12.45 (observed)
└─ Estimated total: $13.20
```

---

## migrate

Migrate a v0.1/v0.2 project to v0.3.

```bash
atlas migrate [<path>]
```

| Option | Description |
|--------|-------------|
| `<path>` | Project path (default: current) |

**Example:**

```bash
atlas migrate ./legacy-project
```

**Output:**
```
Migrating /path from v0.2 to v0.3...
✓ Converted atlas.json
✓ Migrated goals
✓ Migrated manifest
Migration complete
```

---

## snapshot

Create a point-in-time snapshot of project state.

```bash
atlas snapshot [<path>]
```

| Option | Description |
|--------|-------------|
| `<path>` | Project path (default: current) |

**Example:**

```bash
atlas snapshot
```

**Output:**
```
Snapshot created: .atlas/snapshots/2026-08-15T03-46-31.json
State: P00 Foundation [DRAFT], P01 Features [PLANNED]
```

---

## doctor

Run diagnostics on project health.

```bash
atlas doctor [<path>] [--json]
```

| Option | Description |
|--------|-------------|
| `<path>` | Project path (default: current) |
| `--json` | Output as JSON |

**Example:**

```bash
atlas doctor
atlas doctor ./my-project --json
```

**Output (text):**
```
Project Atlas Doctor
├─ Atlas.json: ✓ valid
├─ Goals: ✓ 5 valid, 0 broken
├─ Agents: ✓ 19 resolved
├─ Skills: ✓ 80 resolved
├─ Recipes: ✓ 7 valid
├─ Policies: ✓ all consistent
└─ Overall: ✓ HEALTHY
```

---

## resolve

Resolve workforce (agents, skills, recipes) for a project profile.

```bash
atlas resolve <profile_path> [--json]
```

| Option | Description |
|--------|-------------|
| `<profile_path>` | Path to project profile JSON |
| `--json` | Output as JSON |

**Example:**

```bash
atlas resolve examples/brasa/project-profile.json
atlas resolve examples/brasa/project-profile.json --json
```

**Output (text):**
```
Agents:
  - implementer
  - reviewer
  - debugger
Skills:
  - clean-code
  - test-design
  - ...
Recipes:
  - feature-standard
  - bug-fix-workflow
```

---

## explain

Explain workforce entities and project intelligence.

### explain workforce

Explain workforce selection for a project.

```bash
atlas explain workforce [--path <path>]
```

**Example:**

```bash
atlas explain workforce
```

### explain agent

Explain a specific agent.

```bash
atlas explain agent <agent_id>
```

**Example:**

```bash
atlas explain agent implementer
```

### explain skill

Explain a specific skill.

```bash
atlas explain skill <skill_id>
```

**Example:**

```bash
atlas explain skill clean-code
```

### explain recipe

Explain a specific recipe.

```bash
atlas explain recipe <recipe_id>
```

**Example:**

```bash
atlas explain recipe feature-standard
```

### explain model

Explain model routing for a specific agent.

```bash
atlas explain model <agent_id> [--path <path>]
```

**Example:**

```bash
atlas explain model implementer
```

### explain context

Explain context retrieval for a goal.

```bash
atlas explain context <goal_id> [--path <path>]
```

**Example:**

```bash
atlas explain context P00-G01
```

---

## framework-check

Validate that the framework installation is healthy.

```bash
atlas framework-check
```

**Example:**

```bash
atlas framework-check
```

**Output:**
```
Project Atlas Framework v0.3.0
✓ Schemas: 27 valid
✓ Adapters: 7 ready
✓ Workforce: 118 skills, 26 agents, 14 recipes
✓ No errors detected
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error (invalid project, validation failed, etc.) |
| `2` | Argument parsing error (bad command syntax) |

---

## Environment Variables

None currently, but reserved for future use (e.g., `ATLAS_CONFIG_DIR`, `ATLAS_LOG_LEVEL`).

---

## See Also

- [Goals 101](../user-guide/goals-101.md) — Getting started with Goals
- [Compiler](./compiler.md) — Compilation to adapters
- [Project Layout](../reference/project-layout.md) — Directory structure
