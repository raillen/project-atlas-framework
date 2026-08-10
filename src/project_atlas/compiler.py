from __future__ import annotations

from pathlib import Path
from typing import Any

from .catalog import by_id
from .io import load_yaml, write_text
from .resources import resource_path

SUPPORTED_TARGETS = {"generic", "chatgpt", "claude", "kimi", "codex", "claude-code", "traycer"}


def _selected(root: Path, kind: str) -> list[str]:
    manifest = load_yaml(root / f".ai/{kind}/manifest.yaml")
    return list(manifest.get(kind, []))


def _render_item(kind: str, item: dict[str, Any]) -> str:
    lines = [f"# {item.get('name', item['id'])}", "", f"ID: `{item['id']}`", ""]
    if item.get("purpose"):
        lines.extend([item["purpose"], ""])
    if item.get("instructions"):
        lines.extend(["## Instructions", "", str(item["instructions"]).strip(), ""])
    if item.get("permissions"):
        lines.extend(["## Permissions", ""])
        for key, value in item["permissions"].items():
            lines.append(f"- `{key}`: `{value}`")
    return "\n".join(lines).rstrip() + "\n"


def compile_target(root: Path, target: str) -> list[Path]:
    if target not in SUPPORTED_TARGETS:
        raise ValueError(f"Unsupported target: {target}")

    agents = by_id("agents")
    skills = by_id("skills")
    selected_agents = _selected(root, "agents")
    selected_skills = _selected(root, "skills")
    adapter_text = resource_path("adapters", f"{target}.md").read_text(encoding="utf-8")
    created: list[Path] = []

    if target == "codex":
        write_text(root / "AGENTS.md", adapter_text)
        created.append(root / "AGENTS.md")
        for item_id in selected_agents:
            path = root / ".codex/agents" / f"{item_id}.md"
            write_text(path, _render_item("agent", agents[item_id]))
            created.append(path)
        for item_id in selected_skills:
            path = root / ".codex/skills" / item_id / "SKILL.md"
            write_text(path, _render_item("skill", skills[item_id]))
            created.append(path)
    elif target == "claude-code":
        write_text(root / "CLAUDE.md", adapter_text)
        created.append(root / "CLAUDE.md")
        for item_id in selected_agents:
            path = root / ".claude/agents" / f"{item_id}.md"
            write_text(path, _render_item("agent", agents[item_id]))
            created.append(path)
        for item_id in selected_skills:
            path = root / ".claude/skills" / item_id / "SKILL.md"
            write_text(path, _render_item("skill", skills[item_id]))
            created.append(path)
    elif target == "traycer":
        path = root / ".traycer/PROJECT_ATLAS.md"
        content = adapter_text + "\n\n## Selected agents\n" + "\n".join(f"- {x}" for x in selected_agents)
        content += "\n\n## Selected skills\n" + "\n".join(f"- {x}" for x in selected_skills)
        write_text(path, content)
        created.append(path)
    else:
        path = root / ".atlas/compiled" / target / "CONTEXT.md"
        sections = [adapter_text, "\n# Active agents\n"]
        sections.extend(_render_item("agent", agents[x]) for x in selected_agents)
        sections.append("\n# Active skills\n")
        sections.extend(_render_item("skill", skills[x]) for x in selected_skills)
        write_text(path, "\n".join(sections))
        created.append(path)
    return created
