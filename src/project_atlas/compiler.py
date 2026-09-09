from __future__ import annotations

from pathlib import Path
from typing import Any

from .catalog import by_id
from .io import existing_data_path, load_data, write_text
from .resources import resource_path

SUPPORTED_TARGETS = {"generic", "chatgpt", "claude", "kimi", "codex", "claude-code", "traycer"}


def _selected(root: Path, kind: str) -> list[str]:
    path = existing_data_path(
        root / f".ai/{kind}/manifest.json",
        root / f".ai/{kind}/manifest.yaml",
    )
    manifest = load_data(path)
    return list(manifest.get(kind, []))


def _render_item(kind: str, item: dict[str, Any]) -> str:
    """Compact platform representation; canonical catalog remains JSON."""
    lines = [f"# {item.get('name', item['id'])}", f"ID: `{item['id']}`"]
    if item.get("purpose"):
        lines.append(str(item["purpose"]).strip())
    if item.get("instructions"):
        lines.extend(["", "## Instructions", str(item["instructions"]).strip()])
    if item.get("permissions"):
        allowed = ", ".join(f"{key}={value}" for key, value in item["permissions"].items())
        lines.extend(["", f"Permissions: {allowed}"])
    return "\n".join(lines).rstrip() + "\n"


def _compile_skill_package(target_dir: Path, skill_id: str, fallback_item: dict[str, Any]) -> list[Path]:
    workforce_skill = resource_path("workforce", "skills", skill_id)
    created: list[Path] = []
    if workforce_skill.exists() and workforce_skill.is_dir():
        for file in sorted(workforce_skill.rglob("*")):
            if file.is_file():
                rel = file.relative_to(workforce_skill)
                dest = target_dir / rel
                dest.parent.mkdir(parents=True, exist_ok=True)
                dest.write_bytes(file.read_bytes())
                created.append(dest)
    else:
        dest = target_dir / "SKILL.md"
        write_text(dest, _render_item("skill", fallback_item))
        created.append(dest)
    return created


def _runtime_entrypoint(root: Path, target: str, adapter_text: str, agents: list[str], skills: list[str]) -> Path:
    path = root / ".atlas/runtime/compiled" / target / "ENTRYPOINT.md"
    content = (
        adapter_text.rstrip()
        + "\n\n## Lean Progressive Context\n"
        + "Start at `ENTRYPOINT.md`, `atlas.json`, the active Goal and `docs/ATLAS.md`. "
        + "Do not preload the repository. Expand only relevant document sections/symbols/tests.\n\n"
        + "Selected agents: " + ", ".join(agents)
        + "\nSelected skills: " + ", ".join(skills)
        + "\n"
    )
    write_text(path, content)
    return path


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
            agent_pkg = resource_path("workforce", "agents", item_id, "AGENT.md")
            if agent_pkg.exists():
                write_text(path, agent_pkg.read_text(encoding="utf-8"))
            elif item_id in agents:
                write_text(path, _render_item("agent", agents[item_id]))
            created.append(path)
        for item_id in selected_skills:
            skill_target_dir = root / ".codex/skills" / item_id
            created.extend(_compile_skill_package(skill_target_dir, item_id, skills.get(item_id, {"id": item_id})))
    elif target == "claude-code":
        write_text(root / "CLAUDE.md", adapter_text)
        created.append(root / "CLAUDE.md")
        for item_id in selected_agents:
            path = root / ".claude/agents" / f"{item_id}.md"
            agent_pkg = resource_path("workforce", "agents", item_id, "AGENT.md")
            if agent_pkg.exists():
                write_text(path, agent_pkg.read_text(encoding="utf-8"))
            elif item_id in agents:
                write_text(path, _render_item("agent", agents[item_id]))
            created.append(path)
        for item_id in selected_skills:
            skill_target_dir = root / ".claude/skills" / item_id
            created.extend(_compile_skill_package(skill_target_dir, item_id, skills.get(item_id, {"id": item_id})))
    elif target == "traycer":
        path = root / ".traycer/PROJECT_ATLAS.md"
        content = (
            adapter_text.rstrip()
            + "\n\n## LPC/PCA\n"
            + "Use `ENTRYPOINT.md` + active Goal + progressive context. Generated context is not canonical.\n"
            + "\n## Selected agents\n" + "\n".join(f"- {x}" for x in selected_agents)
            + "\n\n## Selected skills\n" + "\n".join(f"- {x}" for x in selected_skills)
        )
        write_text(path, content)
        created.append(path)
    else:
        created.append(_runtime_entrypoint(root, target, adapter_text, selected_agents, selected_skills))
    return created
