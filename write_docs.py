import os

docs_dir = "/home/raillen/Documentos/Projetos/project-atlas-framework/docs"

content_writing_skills = """# Writing Skills

Project Atlas v0.3 introduces Skill Packages v2, a directory-based structure that encapsulates everything an AI agent needs to execute a specific task reliably.

## Skill Package v2 Directory Structure

A skill package is a directory containing a `manifest.json` file, a `SKILL.md` instruction file, and optional supporting directories:

```text
skill-name/
├── manifest.json       # Formal capabilities and requirements
├── SKILL.md            # Natural language instructions for the LLM
├── checks/             # Validation checklists (e.g., pre-flight, post-flight)
├── templates/          # Code/text templates used by the skill
├── references/         # Example implementations or documentation
├── scripts/            # Helper scripts executed by the agent
└── examples/           # Example usage scenarios
```

## Manifest Reference (`manifest.json`)

The manifest defines the formal contract for the skill.

```json
{
  "id": "python-refactor",
  "name": "Python Refactoring Expert",
  "version": "1.0.0",
  "schema_version": "2.0",
  "purpose": "Refactors Python code for readability and performance.",
  "risk_level": "medium",
  "modes": ["autonomous", "interactive"],
  "inputs": [
    {"name": "file_path", "description": "Path to the Python file", "type": "path", "required": true}
  ],
  "outputs": [
    {"name": "diff", "description": "Proposed changes", "type": "string"}
  ],
  "requires": ["python3", "pytest"],
  "conflicts": ["legacy-python-refactor"],
  "capabilities": ["fs.read", "fs.write", "shell.exec"],
  "select": {
    "stack": ["python"],
    "features": ["refactoring"]
  },
  "required_evidence": ["tests_pass", "lint_clean"],
  "stop_conditions": ["Max iterations reached", "Tests fail after 3 attempts"],
  "provenance": {
    "author": "Atlas Team",
    "signature": "SHA256:xyz..."
  }
}
```

## SKILL.md Structure

The `SKILL.md` file must contain the following 10 sections:

1. **Title and Description:** Briefly explain what the skill does.
2. **Purpose:** The exact goal of the skill.
3. **Prerequisites:** What must be true before using this skill.
4. **Inputs:** Expected input variables.
5. **Outputs:** Expected outputs.
6. **Execution Steps:** Step-by-step instructions for the LLM.
7. **Verification:** How to verify the skill worked.
8. **Fallbacks:** What to do if execution fails.
9. **Constraints:** Rules the LLM must not break.
10. **Examples:** Concrete examples of good execution.

## Writing Supporting Files

### Checks (`checks/`)
Create markdown checklists (e.g., `pre-flight.md`) with checkboxes (`- [ ]`) that the agent must complete before or after execution.

### Templates (`templates/`)
Store boilerplate code or text formats the agent should use. Reference them in `SKILL.md`.

## Provenance and Supply Chain
Ensure every skill has an author and a mechanism to verify integrity (e.g., cryptographic signatures) in the `manifest.json`.

## Testing a New Skill
Run `atlas test skill <skill_id>` to validate the schema and run any associated unit tests or conformance checks in FakeRuntime.

## Resolver Selection
The resolver uses the `select` block in the manifest to match skills to the current project context (stack, features, etc.).

## Examples
* **Simple:** A skill that simply formats JSON.
* **Complex:** A skill that sets up a full PostgreSQL database, runs migrations, and verifies connectivity.
"""

content_writing_agents = """# Writing Agents

Agents in Project Atlas are defined by Agent Packages. An agent is a specialized persona equipped with a set of skills and permissions.

## Agent Package Directory Structure

```text
agent-name/
├── manifest.json       # Formal agent definition
└── AGENT.md            # Persona and operational rules
```

## Manifest Reference (`manifest.json`)

```json
{
  "id": "frontend-dev",
  "name": "Frontend Developer",
  "version": "1.0.0",
  "purpose": "Implements React components and styles.",
  "inputs": [{"name": "design_spec", "required": true}],
  "outputs": [{"name": "component_code"}],
  "required_skills": ["react-component-gen", "css-styling"],
  "optional_skills": ["jest-testing"],
  "allowed_capabilities": ["fs.read", "fs.write"],
  "required_evidence": ["visual_regression_passed"],
  "may_delegate": true,
  "max_delegation_depth": 2,
  "handoff_to": ["qa-tester"],
  "risk_level": "low",
  "review_requirement": "none",
  "permissions": ["read_src", "write_src"],
  "stop_conditions": ["Task complete", "Blocked on design"]
}
```

## AGENT.md Structure

- **Purpose:** The agent's mission.
- **Inputs:** What the agent needs to start.
- **Outputs:** What the agent produces.
- **Required Skills:** Tools the agent relies on.
- **Capabilities:** System access level.
- **Procedure:** General operational flow.
- **Must Not:** Anti-patterns and restrictions.
- **Handoff:** When and how to pass work to others.
- **Stop Conditions:** When to halt execution.
- **Escalation:** How to handle unrecoverable errors.

## Agent vs. Skill
Create a new agent when you need a distinct persona with specific permissions or delegation flows. If you just need a new capability for an existing agent, create a skill instead.
"""

content_writing_recipes = """# Writing Recipes

Recipes define complex workflows that coordinate multiple agents and skills across a Directed Acyclic Graph (DAG) of steps.

## Recipe Package Directory Structure

```text
recipe-name/
├── recipe.json         # Formal workflow definition
└── RECIPE.md           # Human-readable explanation
```

## Manifest Reference (`recipe.json`)

```json
{
  "id": "fullstack-feature",
  "version": "1.0.0",
  "purpose": "End-to-end implementation of a feature.",
  "preconditions": ["db_running", "repo_clean"],
  "inputs": ["feature_spec"],
  "steps": [
    {
      "id": "design-db",
      "agent": "backend-dev",
      "action": "schema-design"
    },
    {
      "id": "implement-api",
      "depends_on": ["design-db"],
      "agent": "backend-dev",
      "skills": ["api-gen"]
    }
  ],
  "agents": ["backend-dev", "frontend-dev"],
  "skills": ["schema-design", "api-gen"],
  "gates": [
    {"step": "design-db", "type": "human_approval"}
  ],
  "artifacts": ["schema.sql", "api.ts"],
  "fallbacks": [
    {"step": "implement-api", "action": "retry_with_simplified_spec"}
  ],
  "stop_conditions": ["Feature deployed", "Validation failed 3 times"]
}
```

## Defining Step DAGs
Use the `depends_on` array in each step to define execution order. The runner will execute independent steps in parallel.

## Gate Placement Strategy
Place `human_approval` or `automated_test` gates after critical steps (e.g., database schema changes) before proceeding to dependent steps.

## Fallback Configuration
Define fallback actions for flaky steps to ensure resilience (e.g., switching to a smarter model or simplifying the task).
"""

content_integration_codex = """# Codex Integration

This guide explains how to integrate Project Atlas with Codex agents.

## Compilation Output

Running `atlas compile codex` produces an integration layer tailored for Codex:

```text
.codex/
├── AGENTS.md               # Primary entrypoint for Codex
└── skills/
    └── <skill_id>/         # Fully resolved skill packages (copied over)
```

## How Codex Consumes Compiled Skills

Codex agents are instructed via `AGENTS.md` to look in the `.codex/skills/` directory for available tools and instructions. The compiler ensures all templates, scripts, and checks are bundled correctly.

## Step-by-Step

1. `atlas init`: Initialize the project configuration.
2. `atlas compile codex`: Generate the Codex integration artifacts.
3. Verify output: Check `.codex/AGENTS.md` and ensure skills are present.

## Customization

You can override skills locally by placing modified versions in `src/project_atlas/resources/custom_skills/` before compiling.

## Troubleshooting

- **Missing Skills:** Check your project's `stack` and `features` configuration to ensure the skills match the resolver's select criteria.
- **Malformed AGENTS.md:** Ensure your skill manifests are valid JSON.
"""

content_integration_claude_code = """# Claude Code Integration

This guide explains how to integrate Project Atlas with Claude Code.

## Compilation Output

Running `atlas compile claude-code` produces:

```text
.claude/
├── CLAUDE.md               # Primary entrypoint for Claude
└── skills/
    └── <skill_id>/         # Fully resolved skill packages
```

## How Claude Code Consumes Compiled Skills

Claude is directed by `CLAUDE.md` to utilize the specific directory structure of `.claude/skills/`. The compiler formats the instructions to align with Claude's expected prompt structures.

## Step-by-Step

1. `atlas init`: Initialize Atlas.
2. `atlas compile claude-code`: Build the Claude integration.
3. Verify output: Review `CLAUDE.md`.

## Customization and Overrides

Use project-local overrides in your Atlas config to modify skill selection or agent personas specifically for the Claude compilation target.
"""

content_integration_gemini = """# Gemini Integration

This guide explains how to compile Atlas configurations for Gemini-based workflows.

## Generic Adapter Compilation

Gemini uses the generic adapter. Run `atlas compile generic --target gemini`.

## Configuration

Configure your `atlas.json` to prioritize Gemini-compatible skills (e.g., avoiding skills that rely heavily on specific non-Gemini system prompts).

## AGENTS.md Structure

For Gemini, the generated `AGENTS.md` focuses heavily on clear, step-by-step reasoning and explicit tool-use instructions tailored for Gemini's context window.
"""

content_integration_opencode = """# OpenCode Integration

This guide explains how to use Project Atlas with OpenCode.

## Generic Adapter Compilation

OpenCode utilizes the generic compiler target: `atlas compile generic --target opencode`.

## Configuration and Customization

Ensure your project config specifies OpenCode as the primary workforce engine. OpenCode relies on standard markdown instructions, so minimal specific overrides are usually needed.
"""

content_development_architecture = """# Architecture

Project Atlas is built around a schema-first, compiler-driven architecture.

## High-Level Architecture Diagram

```text
User/Config -> [ CLI ] -> [ Scaffold / Init ]
                              |
                        [ Resolver ] <--- (Registry/Manifests)
                              |
                        [ Compiler ] ---> [ Target Adapters (Codex, Claude, etc.) ]
                              |
                        [ Workforce / Goals ]
```

## Module Responsibilities

- **cli:** Command-line interface entrypoints.
- **compiler:** Translates resolved states into target-specific artifacts (e.g., `.codex/`).
- **resolver:** Determines which skills, agents, and recipes apply based on project context.
- **validator:** Ensures manifests and configs match JSON schemas.
- **scaffolder:** Generates boilerplate for new projects.
- **workforce:** Manages agent definitions and rosters.
- **goals:** Defines high-level objectives.
- **context:** Manages environment and project state.
- **doctor:** Diagnostics and troubleshooting.
- **explain:** Provides traces for resolver decisions.
- **migration:** Upgrades configurations between schema versions.
- **fake_runtime:** A deterministic test environment for skills.

## Data Flow

`profile` -> `resolve` (filter applicable resources) -> `scaffold` (setup project) -> `compile` (generate platform-specific instructions).

## Schema-First Development

All configurations (skills, agents, recipes) are backed by strict JSON schemas. Changes to the domain model require schema updates first.

## Canonical vs Generated Content

Human-maintained (canonical) content lives in `src/project_atlas/resources/` or project config. Generated content (e.g., `.codex/`) is ephemeral and should not be edited manually.

## Protocol Versioning

Manifests use semantic versioning and explicit schema versions (`schema_version: "2.0"`).

## Key Design Decisions

- **Provider-Neutrality:** The core framework does not depend on any specific LLM or orchestrator.
- **Deterministic Resolution:** Given the same inputs, the resolver always produces the same output.
"""

content_development_compiler = """# Compiler

The Compiler translates resolved Atlas state into target-specific agent instructions.

## What the Compiler Does

It takes a resolved set of skills, agents, and recipes, formats them, and writes them to a target directory alongside a primary entrypoint file (like `AGENTS.md` or `CLAUDE.md`).

## Supported Targets

- `codex`
- `claude-code`
- `traycer`
- `generic`

## Skill Compilation

The compiler performs a recursive copy of skill packages (including `checks/`, `templates/`, etc.) into the target directory, ensuring the agent has access to all supporting files locally.

## Entrypoint Generation

It concatenates agent personas and high-level routing rules into `AGENTS.md` or `CLAUDE.md`, providing the LLM with its initial context.

## Adding a New Target

Implement a new class inheriting from `BaseAdapter` in `src/project_atlas/core/adapters/`.

## Determinism

Compilation must be strictly deterministic. Sorting, hashing, and copying operations must produce identical output byte-for-byte across runs.
"""

content_development_resolver = """# Resolver

The Resolver is the heart of Project Atlas, responsible for dynamically selecting the right capabilities for a given context.

## Resolution Algorithm

The resolver applies filters in the following order:
1. Core (always included)
2. Project-Type
3. Stack
4. Features
5. Risk level
6. Bundles
7. Risk-rules
8. Transitive dependencies (resolving `requires` blocks)

## Explainability Traces

Use `atlas explain` to view exactly why a specific skill was included or excluded during resolution.

## Determinism Guarantees

The resolver guarantees deterministic output for a given project configuration and catalog version.

## Testing Resolution

Add tests in `tests/test_resolver.py` using mock catalogs to ensure filtering logic remains sound.

## Select Criteria

Skills, agents, and recipes use the `select` block in their manifests to declare when they apply (e.g., `{"stack": ["python"]}`).
"""

content_development_testing = """# Testing

Project Atlas emphasizes rigorous testing to ensure deterministic behavior and backward compatibility.

## Test Suite Organization

- `test_schemas.py`: Validates schema definitions.
- `test_goals.py`: Goal evaluation logic.
- `test_resolver.py`: Resolution algorithm correctness.
- `test_scaffolder.py`: Project initialization.
- `test_conformance.py`: Validates skills against contracts.
- `test_workforce.py`: Agent/Workforce definition logic.
- `test_cli.py`: CLI command integration tests.
- `test_migration.py`: Schema migration tools.

## Running Tests

Run `pytest` from the repository root. Use `pytest -v` for verbose output.

## Conformance Suite

Ensures that provided skills adhere strictly to the schema and structural rules (e.g., verifying `SKILL.md` sections).

## FakeRuntime

A deterministic execution environment used to test agent interactions and tool calls without invoking real LLMs or touching the live filesystem unnecessarily.

## Golden Fixtures

Tests relying on complex compilation outputs use golden fixtures stored in `tests/fixtures/`. Update these carefully when compiler output intentionally changes.
"""

content_development_contributing = """# Contributing

Thank you for contributing to Project Atlas!

## Development Setup

1. Clone the repository.
2. Run `poetry install` (or equivalent) to install dependencies.
3. Run `pre-commit install` to set up formatting hooks.

## Code Style

Follow PEP 8 for Python code. Use `black` for formatting and `isort` for imports.

## Adding a New Schema

1. Update the schema definition in `src/project_atlas/resources/schemas/`.
2. Generate models if applicable.
3. Add tests in `test_schemas.py`.

## Adding a New Skill, Agent, or Recipe

1. Create the package directory in `src/project_atlas/resources/catalog/`.
2. Write the manifest and markdown file.
3. Ensure conformance tests pass.

## Adding a CLI Command

Add the command to `src/project_atlas/cli/` using the established CLI framework (e.g., Typer or argparse).

## PR Process

- Open a PR with a descriptive title.
- Ensure all CI checks (pytest, linting) pass.
- Request review from core maintainers.

## AGENTS.md Rules

Adhere strictly to the rules defined in the repository's root `AGENTS.md` (e.g., provider-neutrality, Lean Progressive Context).
"""

files_to_write = {
    os.path.join(docs_dir, "authoring/writing-skills.md"): content_writing_skills,
    os.path.join(docs_dir, "authoring/writing-agents.md"): content_writing_agents,
    os.path.join(docs_dir, "authoring/writing-recipes.md"): content_writing_recipes,
    os.path.join(docs_dir, "integration/codex.md"): content_integration_codex,
    os.path.join(docs_dir, "integration/claude-code.md"): content_integration_claude_code,
    os.path.join(docs_dir, "integration/gemini.md"): content_integration_gemini,
    os.path.join(docs_dir, "integration/opencode.md"): content_integration_opencode,
    os.path.join(docs_dir, "development/architecture.md"): content_development_architecture,
    os.path.join(docs_dir, "development/compiler.md"): content_development_compiler,
    os.path.join(docs_dir, "development/resolver.md"): content_development_resolver,
    os.path.join(docs_dir, "development/testing.md"): content_development_testing,
    os.path.join(docs_dir, "development/contributing.md"): content_development_contributing,
}

for path, content in files_to_write.items():
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        f.write(content)

print("Files written successfully.")
