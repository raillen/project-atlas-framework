# Writing Agents

Agents in Prumo are defined by Agent Packages. An agent is a specialized persona equipped with a set of skills and permissions.

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
