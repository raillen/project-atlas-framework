# Goal Workflow Demo

This example demonstrates Project Atlas's **Goal State Machine** and formal change management workflow.

## Goals Included

### State Machine Progression

| Goal | Priority | State | Description |
|------|----------|-------|-------------|
| P00-G01 | P0 | DRAFT | Newly created, not yet approved |
| P00-G02 | P0 | PLANNED | Approved for planning, awaiting lock |
| P00-G03 | P0 | LOCKED | Design locked, ready for execution |
| P00-G04 | P0 | EXECUTING | Active execution with amendment |
| P00-G05 | P1 | VERIFYING | Verification in progress |
| P00-G06 | P1 | REVIEWING | Final review underway |
| P00-G07 | P2 | DONE | Completed and operational |
| P00-G08 | P1 | BLOCKED | Blocked by dependency |

## State Machine Diagram

```
                ┌─────────────────────────────────────────────────┐
                │                   DRAFT                          │
                │  (Created, not yet approved for planning)       │
                └─┬──────────────────────────────────────────────┬┘
                  │                                              │
    ┌─────────────▼────────────────┐                            │
    │  Can transition to: PLANNED  │                            │
    │  or BLOCKED                  │                            │
    └─────────────┬────────────────┘                            │
                  │                                              │
         ┌────────▼────────┐                                    │
         │    PLANNED      │                                    │
         │  (Approved for  │                                    │
         │   planning)     │                                    │
         └────────┬────────┘                                    │
                  │                                              │
        ┌─────────▼──────────┐                                  │
        │ Can transition to: │                                  │
        │ LOCKED, DRAFT,     │                                  │
        │ or BLOCKED         │                                  │
        └─────────┬──────────┘                                  │
                  │                                              │
           ┌──────▼──────┐                                       │
           │   LOCKED    │                                       │
           │  (Locked    │                                       │
           │   for exec) │                                       │
           └──────┬──────┘                                       │
                  │                                              │
       ┌──────────▼──────────┐                                  │
       │ Can transition to:  │                                  │
       │ EXECUTING, BLOCKED  │                                  │
       └──────────┬──────────┘                                  │
                  │                                              │
             ┌────▼────┐                                         │
             │ EXECUTE │                                         │
             │(In work)│                                         │
             └────┬────┘                                         │
                  │                                              │
        ┌─────────▼────────┐                                    │
        │Can transition to:│                                    │
        │VERIFYING,BLOCKED │                                    │
        └─────────┬────────┘                                    │
                  │                                              │
            ┌─────▼──────┐                                       │
            │  VERIFYING │                                       │
            │(Check work)│                                       │
            └─────┬──────┘                                       │
                  │                                              │
         ┌────────▼────────┐                                    │
         │Can transition:  │                                    │
         │REVIEWING,EXE... │                                    │
         └────────┬────────┘                                    │
                  │                                              │
            ┌─────▼────────┐                                     │
            │  REVIEWING   │                                     │
            │(Final review)│                                     │
            └─────┬────────┘                                     │
                  │                                              │
        ┌─────────▼───────┐                                     │
        │Can transition:  │                                     │
        │DONE, EXECUTING, │                                     │
        │BLOCKED          │                                     │
        └─────────┬───────┘                                     │
                  │                                              │
           ┌──────▼──────┐  (Any state)                         │
           │    DONE     │  (Terminal)                          │
           │(Completed)  │                                       │
           └─────────────┘                                       │
                                                                 │
                    ┌───────────────────────────────────────────┘
                    │
                  ┌─▼────────┐
                  │ BLOCKED  │
                  │ (Waiting)│
                  └─────┬────┘
                        │
          ┌─────────────▼────────────────┐
          │Can transition to: PLANNED,   │
          │LOCKED, EXECUTING             │
          └──────────────────────────────┘
```

## Features Demonstrated

### 1. **Draft Phase** (P00-G01)
- Goal is created in DRAFT state
- Not yet approved or assigned

### 2. **Planning Phase** (P00-G02)
- Goal transitions to PLANNED after approval
- Ready for detailed design or spec work

### 3. **Locking Phase** (P00-G03)
- Goal is LOCKED after design/architecture approval
- Lock contains digest (content signature) and metadata
- Prevents unauthorized modifications
- Only valid transitions: EXECUTING or BLOCKED

### 4. **Amendments** (P00-G04)
- Demonstrates formal change management
- Shows amendment with new revision and digest
- Goal can be amended while EXECUTING
- Amendment tracked in history

### 5. **Execution & Verification** (P00-G05)
- Goal moves through EXECUTING → VERIFYING
- Tests, quality checks, coverage validation

### 6. **Review Phase** (P00-G06)
- REVIEWING state for final approval
- Can go back to EXECUTING if issues found
- Can advance to DONE if approved

### 7. **Completion** (P00-G07)
- Goal reaches DONE state (terminal)
- Operational in production
- No further transitions possible

### 8. **Blocking** (P00-G08)
- Goal can be BLOCKED by dependencies
- Tracked in history with reason
- Can unblock when dependency resolved
- Transitions: PLANNED, LOCKED, EXECUTING

## Using This Example

### View Entire Workflow

```bash
# Navigate to example
cd examples/goal-workflow-demo

# Validate project health
atlas doctor .

# List all goals to see states
atlas goal list

# Examine specific goal
atlas goal list --json | jq '.[] | {id, title, state}'
```

### Compile for Different Targets

```bash
# Generic target (default)
atlas compile --target generic

# GitHub Copilot integration
atlas compile --target codex

# Claude-Code extended mode
atlas compile --target claude-code
```

### Study State Transitions

Each goal file (`.ai/goals/P00-Gxx.goal.json`) contains:

1. **Current state** — `"state": "LOCKED"`
2. **Lock metadata** — `"lock": {...}` (if locked)
3. **Amendments** — `"amendments": [...]` (if amended)
4. **History** — Complete audit trail of transitions

Example from P00-G04:
```json
{
  "id": "P00-G04",
  "state": "EXECUTING",
  "lock": { ... },
  "amendments": [
    {
      "revision": 1,
      "reason": "Add GraphQL endpoint in addition to REST",
      "digest": "sha256-c3d4e5f6g7h8..."
    }
  ],
  "history": [
    { "at": "...", "event": "created", "state": "DRAFT" },
    { "at": "...", "event": "transitioned", "state": "PLANNED" },
    { "at": "...", "event": "transitioned", "state": "LOCKED" },
    { "at": "...", "event": "transitioned", "state": "EXECUTING" },
    { "at": "...", "event": "amended", "revision": 1, "reason": "..." }
  ]
}
```

## State Machine Rules

**Direct Transitions Allowed:**

| From | To | Allowed | Reason |
|------|----|---------| -------|
| DRAFT | PLANNED | ✓ | Approved for planning |
| DRAFT | BLOCKED | ✓ | Issue discovered |
| PLANNED | LOCKED | ✓ | Design approved |
| PLANNED | DRAFT | ✓ | Revert if needed |
| PLANNED | BLOCKED | ✓ | Dependency blocked |
| LOCKED | EXECUTING | ✓ | Begin execution |
| LOCKED | BLOCKED | ✓ | Dependency blocker |
| EXECUTING | VERIFYING | ✓ | Work complete |
| EXECUTING | BLOCKED | ✓ | Blocker discovered |
| VERIFYING | REVIEWING | ✓ | Verification done |
| VERIFYING | EXECUTING | ✓ | Issues found, rework |
| VERIFYING | BLOCKED | ✓ | Blocker discovered |
| REVIEWING | DONE | ✓ | Approved |
| REVIEWING | EXECUTING | ✓ | Rework needed |
| REVIEWING | BLOCKED | ✓ | Blocker discovered |
| BLOCKED | PLANNED | ✓ | Unblock, restart planning |
| BLOCKED | LOCKED | ✓ | Unblock, restart execution |
| BLOCKED | EXECUTING | ✓ | Unblock, resume execution |
| DONE | (none) | ✗ | Terminal state |

## Learning Outcomes

After studying this example, you'll understand:

1. ✓ How goals progress through a formal state machine
2. ✓ When goals can be locked and why
3. ✓ How amendments track formal changes
4. ✓ How goals can be blocked and unblocked
5. ✓ Complete audit trail for compliance/governance
6. ✓ Integration with compilation and context generation

## Next Steps

- Read [Goals 101](../../docs/user-guide/goals-101.md) for detailed guide
- Review [CLI Reference](../../docs/reference/cli.md) for all commands
- Explore [Compiler Architecture](../../docs/development/compiler.md)
