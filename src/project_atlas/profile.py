from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Any, Iterable

from jsonschema import Draft202012Validator

from .io import load_data


def _flatten(value: Any) -> list[str]:
    out: list[str] = []
    if isinstance(value, dict):
        for key, child in value.items():
            out.append(str(key).lower())
            out.extend(_flatten(child))
    elif isinstance(value, (list, tuple, set)):
        for child in value:
            out.extend(_flatten(child))
    elif value is not None:
        out.append(str(value).lower())
    return out


@dataclass(frozen=True)
class ProjectProfile:
    raw: dict[str, Any]

    @property
    def project_types(self) -> set[str]:
        value = self.raw.get("project", {}).get("type", [])
        if isinstance(value, str):
            value = [value]
        return {str(item).lower() for item in value}

    @property
    def stack_tokens(self) -> set[str]:
        return set(_flatten(self.raw.get("stack", {})))

    @property
    def features(self) -> set[str]:
        return {str(item).lower() for item in self.raw.get("features", [])}

    @property
    def risks(self) -> set[str]:
        quality = self.raw.get("quality", {})
        explicit = self.raw.get("risk", [])
        values = list(explicit) if isinstance(explicit, list) else [explicit] if explicit else []
        values.extend(key for key, enabled in quality.items() if enabled is True or enabled == "high")
        return {str(item).lower() for item in values}

    @property
    def preferred_models(self) -> list[dict[str, Any]]:
        return list(self.raw.get("ai", {}).get("preferred_models", []))

    @property
    def orchestrator(self) -> str:
        return str(self.raw.get("ai", {}).get("orchestrator", "native"))

    @property
    def autonomy(self) -> str:
        return str(self.raw.get("ai", {}).get("autonomy", "agentic"))


def load_profile(path: Path) -> ProjectProfile:
    return ProjectProfile(load_data(path))


def validate_profile(raw: dict[str, Any], schema: dict[str, Any]) -> list[str]:
    errors = Draft202012Validator(schema).iter_errors(raw)
    return [f"{'.'.join(map(str, e.absolute_path)) or '<root>'}: {e.message}" for e in errors]


def any_match(candidates: Iterable[str], actual: set[str]) -> bool:
    return any(str(candidate).lower() in actual for candidate in candidates)
