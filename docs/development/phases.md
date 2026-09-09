# Project Atlas v0.4 Implementation Phases

> **Rule**: Only fully detail current phase + next phase. Future phases remain architectural until dependencies mature.

---

## M0 — Specification Freeze

**Objective**: Establish immutable baseline for migration; prevent scope creep.

**Dependencies**: None (starting state)

**Scope**:
- Record Architecture Decision: Go Core (ADR)
- Inventory all v0.3 observable contracts:
  - CLI commands, subcommands, flags, exit codes
  - JSON output schemas (envelope, data shapes)
  - File system effects (created/modified/deleted paths)
  - Schema validation behavior
  - Golden fixture outputs
- Catalog all Python modules, scripts, tests, resources
- Define Go module path and initial package boundaries
- Create `conformance/` matrix: Python fixture → expected Go output
- Freeze v0.3: no new features in Python except critical bug fixes

**Non-Goals**:
- Writing Go implementation code
- Designing v0.4 features beyond migration needs

**Deliverables**:
- `docs/migration/protocol-inventory.md`
- `docs/migration/conformance-strategy.md`
- `conformance/V03_BASELINE.json`
- `conformance/golden/` populated from Python v0.3
- Go module initialized (`go.mod`)

**Test/Eval Requirements**:
- All 127 Python tests passing (baseline)
- Conformance fixtures defined for: `init`, `resolve`, `validate`, `goal`, `context plan`, `report`, `migrate`, `compile`, `snapshot`, `doctor`, `explain`, `framework-check`

**Acceptance Criteria**:
- [ ] Architecture Freeze ADR committed
- [ ] 100% of v0.3 CLI surface cataloged with golden outputs
- [ ] Conformance matrix covers all critical contracts
- [ ] Go module builds empty `cmd/atlas` + `internal/*` skeleton

**Exit Gate**: M0 complete when `docs/migration/protocol-inventory.md` and `conformance/V03_BASELINE.json` are reviewed and approved.

---

## M1 — Go Foundation

**Objective**: Build minimal Go skeleton that compiles, tests, and runs; establish quality gates.

**Dependencies**: M0 complete

**Scope**:
- Go module + `go.sum`
- `cmd/atlas/main.go` with machine envelope (`protocol_version`, `ok`, `data`, `diagnostics`, `warnings`)
- `internal/protocol`: envelope types, version constants
- `internal/app`: `Service` with `Version()`, `ProjectRoot()` stubs
- `internal/project`: `FindRoot()` via `atlas.json` discovery
- `internal/validation`: JSON Schema validation (Draft 2020-12) — initially delegate to Python or use `gojsonschema`
- `internal/resources`: access to `schemas/` + `src/project_atlas/resources/` via `io/fs.FS` abstraction (prepare for `go:embed`)
- Error model: wrapped errors with `%w`, sentinel errors, no panic for user errors
- Test infrastructure: `go test -race ./...`, table-driven tests
- CI: `gofmt`, `go vet`, `go test -race`, `staticcheck` (recommended)
- Logging/output discipline: stdout for `--json`, stderr for diagnostics, no ANSI in JSON

**Non-Goals**:
- Protocol domain logic (Goals, Plans, etc.)
- Real schema validation (can delegate to Python subprocess initially)
- Full CLI command surface

**Deliverables**:
- `cmd/atlas/main.go` with `version` command (`--json` support)
- `internal/protocol/protocol.go` + tests
- `internal/app/service.go` + tests
- `internal/project/project.go` + tests
- `internal/validation/validation.go` + tests
- `internal/resources/resources.go` + tests
- CI workflow with Go job
- `.gitignore` for Go artifacts

**Test/Eval Requirements**:
- `go test -race ./...` passes
- `go vet ./...` passes
- `gofmt -l .` empty
- `staticcheck ./...` passes (if configured)
- Python tests still 100% passing (no regression)

**Acceptance Criteria**:
- [ ] `go run ./cmd/atlas version` → `0.4.0-dev`
- [ ] `go run ./cmd/atlas --json version` → valid envelope
- [ ] `go test -race ./...` all green
- [ ] CI runs Python + Go quality gates
- [ ] No circular dependencies in `internal/`

**Exit Gate**: M1 complete when all Go quality gates pass in CI and `atlas version --json` produces valid envelope.

---

## M2 — Protocol Parity

**Objective**: Port all v0.3 protocol domain to Go with 100% conformance on critical contracts.

**Dependencies**: M1 complete

**Scope**:
### Protocol Domain (`internal/protocol/`)
- `goals/`: Goal v2 struct, lock/verify, amendment, state transitions, hashing
- `plans/`: Plan, Task DAG, cycle detection
- `tasks/`: Task struct, execution state
- `events/`: Event protocol, structured logging
- `evidence/`: Evidence struct, validation
- `gates/`: Gate, waiver, release readiness

### Project Service (`internal/project/`)
- `atlas.json` load/parse with schema validation
- Profile loading, capabilities, model policy
- Project state snapshot

### Resolution Service (`internal/resolver/`)
- Workforce resolution (agents, skills, recipes)
- Risk assessment
- Model policy evaluation
- Context selection

### Validator (`internal/validator/`)
- Full JSON Schema validation (Draft 2020-12)
- Schema registry with cross-ref resolution
- Project validation against schemas

### CLI Commands (Parity)
- `atlas init`, `resolve`, `validate`
- `atlas goal new|state|amend|list`
- `atlas context plan`
- `atlas report add|summary`
- `atlas migrate`
- `atlas doctor`, `explain`, `framework-check`

### Compiler (`internal/compiler/`)
- Target adapter interface
- Codex, Claude Code, Generic generators

**Non-Goals**:
- Documentation System v2 (contracts, profiles, UI pack)
- Living Plan / Interview Engine
- Adoption Engine
- Control Plane (Run, Budget, Context Compiler, etc.)
- Experience Layer
- OpenCode native harness

**Deliverables**:
- Full protocol domain packages with tests
- All 12 CLI commands ported with `--json` support
- Compiler targets: Codex, Claude Code, Generic
- Conformance suite executing Go vs Python on identical fixtures

**Test/Eval Requirements**:
- **Conformance 100% on critical contracts**: identical exit codes, JSON envelopes, filesystem effects
- Unit tests for all domain logic
- Integration tests with temp repos
- Golden fixtures: `conformance/golden/<cmd>.golden.json` matched exactly
- `go test -race ./...`, `go vet`, `gofmt`, `staticcheck` all pass

**Acceptance Criteria**:
- [ ] `atlas init --non-interactive --profile X` creates valid project
- [ ] `atlas resolve` matches Python workforce/skills/recipes exactly
- [ ] `atlas validate` catches same schema violations
- [ ] `atlas goal` lifecycle identical (lock, amendment, transitions)
- [ ] `atlas context plan` produces same budget/strategy
- [ ] `atlas doctor` finds same issues
- [ ] `atlas compile --target codex|claude-code|generic` generates byte-identical or semantically equivalent outputs
- [ ] `atlas explain` outputs match for workforce, agent, skill, recipe, context, model, execution
- [ ] Conformance diff: zero failures on critical contracts

**Exit Gate**: M2 complete when `go test -run Conformance` (or equivalent) shows 100% parity on all cataloged critical contracts.

---

## M3 — Compiler Parity

**Objective**: Full compiler target parity including ancillary targets; conformance gate.

**Dependencies**: M2 complete

**Scope**:
- All v0.3 compiler targets ported and tested
- Ancillary targets (ChatGPT, Kimi, Traycer, etc.) if still supported
- Template system with `go:embed` for adapter templates
- Generated artifact ownership markers
- Conformance validation on compiler outputs

**Non-Goals**: New compiler targets, OpenCode native (M10)

**Deliverables**: All compiler targets with golden fixture tests

**Test/Eval Requirements**: 100% conformance on all compiler outputs

**Acceptance Criteria**: [ ] All v0.3 targets pass conformance; no regressions

**Exit Gate**: M3 complete when compiler conformance = 100% and M2 gate still holds.

---

## M4 — Distribution v1

**Objective**: Usable install/uninstall experience; single binary releases.

**Dependencies**: M3 complete

**Scope**:
- GitHub Releases with signed binaries (Linux amd64/arm64, macOS amd64/arm64, Windows amd64)
- Install script (`curl | sh`) with checksum verification
- Homebrew tap
- Windows packaging baseline (Scoop/WinGet)
- `atlas setup` wizard (detect harnesses, select integrations, validate PATH, run doctor)
- `atlas install connector <id>`, `atlas uninstall [--connectors|--purge-cache|--purge-global-config]`
- Installation manifest (versions, paths, connectors, managed config fragments)
- Cleanup manifests per connector
- Portable mode (`atlas --home ./path`)
- Idempotent install/uninstall

**Non-Goals**: Auto-update from non-GitHub sources, complex multi-user server

**Deliverables**: Release artifacts, install script, Homebrew formula, `atlas setup/install/uninstall`

**Test/Eval Requirements**: Install/uninstall on clean VMs for all platforms; idempotency verified

**Acceptance Criteria**:
- [ ] `curl -fsSL install.sh | sh` installs working `atlas`
- [ ] `atlas uninstall` removes binary + connectors + cache (not project data)
- [ ] `atlas setup` detects harnesses and configures project-local adapters
- [ ] Homebrew install works
- [ ] Windows package installs

**Exit Gate**: M4 complete when release pipeline produces signed binaries and install/uninstall verified on all target platforms.

---

## M5 — Documentation System v2

**Dependencies**: M2, M4

**Scope**:
- Documentation Contracts (knowledge requirements → required docs)
- Documentation Profiles (capability → doc pack)
- Coverage/Readiness engine (missing/partial/ready/not-applicable)
- UI Documentation Pack (semantic wireframes, tokens, components, states, accessibility)
- Docs Delta (incremental updates on decisions)
- Contradiction Framework (detection, reporting, resolution proposals)

**Non-Goals**: Living Plan interview (M6), Adoption (M7)

**Exit Gate**: Documentation contracts enforceable; profiles selectable; delta tracked; contradictions detected.

---

## M6 — Living Plan

**Dependencies**: M5

**Scope**:
- Interview protocol (conversational, incremental)
- Decision extraction with authority model
- Open questions tracking
- Confidence scoring
- Readiness gates (block implementation until critical decisions confirmed)
- Incremental documentation updates from decisions

**Exit Gate**: New project can be planned from zero to implementation-ready via interview.

---

## M7 — Adoption Engine

**Dependencies**: M5, M6

**Scope**:
- Repository scanner (file types, frameworks, configs, docs)
- Semantic documentation mapping (non-Atlas layouts → Atlas concepts)
- Capability detection
- Confidence ledger (scored mapping)
- Adoption report
- Migration proposals (non-destructive, reversible)

**Exit Gate**: Existing non-Atlas repo can be audited and migrated incrementally.

---

## M8 — History + Traceability

**Dependencies**: M2, M6

**Scope**:
- Implementation Journal (synthesis, not chain-of-thought)
- Experiment/Rejection/Debt registers
- Typed traceability graph (req↔dec↔code↔test↔doc↔evidence)
- `atlas trace <ref>` CLI

**Exit Gate**: Any code change traceable to decision/goal; journal queryable.

---

## M9 — Experience Layer

**Dependencies**: M5, M6, M7, M8

**Scope**:
- Session events (structured, not transcript)
- Summaries (per session, per Goal)
- Handoff Protocol (state transfer between sessions/agents)
- Experience Provider Contract
- Experience Proposals (validated before promotion)
- Retention policies

**Exit Gate**: Agent can hand off to another agent/session with full structured context.

---

## M10 — OpenCode Native Harness

**Dependencies**: M3, M4, M9

**Scope**:
- OpenCode native compiler (TypeScript plugin + agents + skills + commands + hooks)
- Atlas primary agent for OpenCode
- Subagents, skills, commands mapped to OpenCode primitives
- Tool guards (pre-tool validation)
- Session hooks (start/end/tool events)
- Connector Contract tests for OpenCode

**Exit Gate**: `atlas connector install opencode` produces fully functional native integration.

---

## M11 — Connector SDK

**Dependencies**: M10

**Scope**:
- Capability negotiation protocol
- Cleanup manifests
- Test kit (fixtures, contract tests)
- Gemini CLI connector
- Claude Code / Codex harness elevation

**Exit Gate**: New connector can be built against SDK and passes contract tests.

---

## M12 — Team/Advanced Runtime

**Dependencies**: M11 (only after real usage validates need)

**Scope**: Shared runtime, leases, concurrency coordination, optional server — **deferred until proven necessary**.

---

## Phase Detail Rule

Only **M0** and **M1** are fully detailed above. M2–M4 have sufficient detail for planning. M5–M12 remain architectural — they will be detailed when their dependencies are mature and implementation begins.