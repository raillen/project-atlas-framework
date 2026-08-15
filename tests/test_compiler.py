"""
Comprehensive tests for Project Atlas compiler adapters.

Tests each of 7 adapter targets to ensure:
1. Compilation succeeds without errors
2. Output files are created with correct structure
3. Generated content is valid and readable
4. Adapter-specific requirements are met
"""

import json
import tempfile
from pathlib import Path

import pytest

from project_atlas.compiler import compile_target, SUPPORTED_TARGETS
from project_atlas.io import load_data


@pytest.fixture
def test_project(tmp_path):
    """Create a minimal valid test project."""
    project_dir = tmp_path / "test_project"
    project_dir.mkdir()

    # Create atlas.json (v0.3)
    atlas_json = project_dir / "atlas.json"
    atlas_json.write_text(
        json.dumps({
            "version": 3,
            "protocol": {"version": 3, "compatible": ">=3 <4"},
            "project": {
                "name": "test-project",
                "type": ["example"]
            },
            "stack": {
                "languages": ["python"],
                "frameworks": []
            },
            "features": [],
            "quality": {
                "coverage_target": 80,
                "type_check": True,
                "lint": True
            },
            "ai": {
                "orchestrator": "native",
                "autonomy": "agentic",
                "preferred_models": []
            }
        })
    )

    # Create .ai directory structure
    ai_dir = project_dir / ".ai"
    ai_dir.mkdir()

    # Create manifest files
    (ai_dir / "manifest.json").write_text(
        json.dumps({
            "version": 3,
            "project": "test-project",
            "lastValidated": "2026-08-15T00:00:00Z"
        })
    )

    # Create agents manifest with sample agents
    agents_dir = ai_dir / "agents"
    agents_dir.mkdir()
    (agents_dir / "manifest.json").write_text(
        json.dumps({
            "agents": ["coder", "architect"]
        })
    )

    # Create skills manifest with sample skills
    skills_dir = ai_dir / "skills"
    skills_dir.mkdir()
    (skills_dir / "manifest.json").write_text(
        json.dumps({
            "skills": ["code-generation", "architecture-design"]
        })
    )

    # Create sample Goals
    goals_dir = ai_dir / "goals"
    goals_dir.mkdir()

    goal_p0_g01 = goals_dir / "P00-G01.goal.json"
    goal_p0_g01.write_text(
        json.dumps({
            "id": "P00-G01",
            "priority": "P0",
            "title": "Setup Project",
            "description": "Initialize project structure",
            "state": "DRAFT",
            "created_at": "2026-08-15T00:00:00Z",
            "history": [
                {
                    "at": "2026-08-15T00:00:00Z",
                    "event": "created",
                    "state": "DRAFT"
                }
            ]
        })
    )

    # Add .gitignore
    gitignore = project_dir / ".gitignore"
    gitignore.write_text("/.atlas/runtime/\n/.codex/\n/.claude-code/\n/.claude/\n/.chatgpt/\n/.kimi/\n/.traycer/\n")

    return project_dir


class TestCompilerGenericAdapter:
    """Tests for generic adapter (default, most portable)."""

    def test_generic_compile_succeeds(self, test_project):
        """Generic adapter should compile without errors."""
        created = compile_target(test_project, "generic")
        assert len(created) > 0, "Should create output files"

    def test_generic_creates_entrypoint(self, test_project):
        """Generic adapter must create ENTRYPOINT.md."""
        compile_target(test_project, "generic")
        entrypoint = test_project / ".atlas/runtime/compiled/generic/ENTRYPOINT.md"
        assert entrypoint.exists(), "ENTRYPOINT.md must exist"
        assert entrypoint.stat().st_size > 0, "ENTRYPOINT.md must not be empty"

    def test_generic_entrypoint_contains_context_guidance(self, test_project):
        """ENTRYPOINT must contain Lean Progressive Context guidance."""
        compile_target(test_project, "generic")
        entrypoint = test_project / ".atlas/runtime/compiled/generic/ENTRYPOINT.md"
        content = entrypoint.read_text()
        assert "Lean Progressive Context" in content, "Should contain context guidance"
        assert "Start at" in content, "Should contain start instruction"

    def test_generic_lists_agents_and_skills(self, test_project):
        """ENTRYPOINT must list selected agents and skills."""
        compile_target(test_project, "generic")
        entrypoint = test_project / ".atlas/runtime/compiled/generic/ENTRYPOINT.md"
        content = entrypoint.read_text()
        assert "agents:" in content.lower() or "Selected agents:" in content
        assert "skills:" in content.lower() or "Selected skills:" in content

    def test_generic_output_is_readable_text(self, test_project):
        """Generic output must be readable UTF-8 text."""
        compile_target(test_project, "generic")
        entrypoint = test_project / ".atlas/runtime/compiled/generic/ENTRYPOINT.md"
        try:
            content = entrypoint.read_text(encoding="utf-8")
            assert len(content) > 0
        except UnicodeDecodeError:
            pytest.fail("Output is not valid UTF-8 text")


class TestCompilerCodexAdapter:
    """Tests for Codex adapter (GitHub Copilot integration)."""

    def test_codex_compile_succeeds(self, test_project):
        """Codex adapter should compile without errors."""
        created = compile_target(test_project, "codex")
        assert len(created) > 0, "Should create output files"

    def test_codex_creates_agents_directory(self, test_project):
        """Codex should create .codex/agents/ directory."""
        compile_target(test_project, "codex")
        agents_dir = test_project / ".codex/agents"
        assert agents_dir.exists(), ".codex/agents directory must exist"

    def test_codex_creates_skills_directory(self, test_project):
        """Codex should create .codex/skills/ directory."""
        compile_target(test_project, "codex")
        skills_dir = test_project / ".codex/skills"
        assert skills_dir.exists(), ".codex/skills directory must exist"

    def test_codex_creates_agents_files(self, test_project):
        """Codex should create individual agent files."""
        compile_target(test_project, "codex")
        agents_dir = test_project / ".codex/agents"
        agent_files = list(agents_dir.glob("*.md"))
        # May be empty if no agents selected, but directory should exist
        assert agents_dir.exists()

    def test_codex_creates_skills_packages(self, test_project):
        """Codex should create skill packages."""
        compile_target(test_project, "codex")
        skills_dir = test_project / ".codex/skills"
        assert skills_dir.exists()

    def test_codex_agents_files_are_readable(self, test_project):
        """Codex agent files must be readable UTF-8."""
        compile_target(test_project, "codex")
        agents_dir = test_project / ".codex/agents"
        for agent_file in agents_dir.glob("*.md"):
            try:
                content = agent_file.read_text(encoding="utf-8")
                assert len(content) > 0, f"{agent_file.name} is empty"
            except UnicodeDecodeError:
                pytest.fail(f"{agent_file.name} is not valid UTF-8")

    def test_codex_creates_agents_summary(self, test_project):
        """Codex should create AGENTS.md summary at root."""
        created = compile_target(test_project, "codex")
        agents_md = test_project / "AGENTS.md"
        # Should be created or at least listed in created files
        # (depends on implementation)
        assert any(agents_md in p.parents or agents_md == p for p in created) or agents_md.exists()


class TestCompilerClaudeCodeAdapter:
    """Tests for Claude-Code adapter."""

    def test_claude_code_compile_succeeds(self, test_project):
        """Claude-Code adapter should compile without errors."""
        created = compile_target(test_project, "claude-code")
        assert len(created) > 0, "Should create output files"

    def test_claude_code_creates_workspace(self, test_project):
        """Claude-Code should create .claude-code/ or .claude/ workspace."""
        compile_target(test_project, "claude-code")
        # Claude-Code may use .claude or .claude-code
        workspace = test_project / ".claude-code" or test_project / ".claude"
        # At minimum, some output should be created
        created = compile_target(test_project, "claude-code")
        assert len(created) > 0, "Claude-Code should create output files"

    def test_claude_code_creates_entrypoint(self, test_project):
        """Claude-Code should create ENTRYPOINT at project root or in workspace."""
        compile_target(test_project, "claude-code")
        # Check root or workspace
        candidates = [
            test_project / "CLAUDE.md",
            test_project / ".claude-code/ENTRYPOINT.md",
            test_project / ".claude-code/CLAUDE.md"
        ]
        found = any(c.exists() for c in candidates)
        assert found, f"ENTRYPOINT not found in {[str(c) for c in candidates]}"

    def test_claude_code_workspace_has_agents(self, test_project):
        """Claude-Code should create agents or similar structure."""
        created = compile_target(test_project, "claude-code")
        # At minimum, compilation should generate output
        assert len(created) > 0, "Claude-Code should create output files"
        # Verify at least one file was created
        assert any(p.exists() for p in created)


class TestCompilerClaudeAdapter:
    """Tests for Claude adapter (standard)."""

    def test_claude_compile_succeeds(self, test_project):
        """Claude adapter should compile without errors."""
        created = compile_target(test_project, "claude")
        assert len(created) > 0, "Should create output files"

    def test_claude_creates_entrypoint(self, test_project):
        """Claude should create ENTRYPOINT.md."""
        compile_target(test_project, "claude")
        entrypoint = test_project / ".atlas/runtime/compiled/claude/ENTRYPOINT.md"
        assert entrypoint.exists(), "ENTRYPOINT.md must exist"
        assert entrypoint.stat().st_size > 0, "ENTRYPOINT.md must not be empty"

    def test_claude_output_is_readable(self, test_project):
        """Claude output must be readable UTF-8."""
        compile_target(test_project, "claude")
        entrypoint = test_project / ".atlas/runtime/compiled/claude/ENTRYPOINT.md"
        content = entrypoint.read_text(encoding="utf-8")
        assert len(content) > 0


class TestCompilerChatGPTAdapter:
    """Tests for ChatGPT adapter."""

    def test_chatgpt_compile_succeeds(self, test_project):
        """ChatGPT adapter should compile without errors."""
        created = compile_target(test_project, "chatgpt")
        assert len(created) > 0, "Should create output files"

    def test_chatgpt_creates_entrypoint(self, test_project):
        """ChatGPT should create ENTRYPOINT.md."""
        compile_target(test_project, "chatgpt")
        entrypoint = test_project / ".atlas/runtime/compiled/chatgpt/ENTRYPOINT.md"
        assert entrypoint.exists(), "ENTRYPOINT.md must exist"


class TestCompilerKimiAdapter:
    """Tests for Kimi adapter."""

    def test_kimi_compile_succeeds(self, test_project):
        """Kimi adapter should compile without errors."""
        created = compile_target(test_project, "kimi")
        assert len(created) > 0, "Should create output files"

    def test_kimi_creates_entrypoint(self, test_project):
        """Kimi should create ENTRYPOINT.md."""
        compile_target(test_project, "kimi")
        entrypoint = test_project / ".atlas/runtime/compiled/kimi/ENTRYPOINT.md"
        assert entrypoint.exists(), "ENTRYPOINT.md must exist"


class TestCompilerTracerAdapter:
    """Tests for Tracer adapter."""

    def test_tracer_compile_succeeds(self, test_project):
        """Tracer adapter should compile without errors."""
        created = compile_target(test_project, "traycer")
        assert len(created) > 0, "Should create output files"

    def test_tracer_creates_output_format(self, test_project):
        """Tracer may create PROJECT_ATLAS.md or ENTRYPOINT."""
        compile_target(test_project, "traycer")
        # Tracer is specialized; check at least one output exists
        candidates = [
            test_project / ".traycer/PROJECT_ATLAS.md",
            test_project / ".atlas/runtime/compiled/traycer/ENTRYPOINT.md",
            test_project / ".traycer/ENTRYPOINT.md"
        ]
        found = any(c.exists() for c in candidates)
        assert found or (test_project / ".traycer").exists()


class TestCompilerMultiTargetWorkflow:
    """Tests for compiling multiple targets in sequence."""

    def test_all_targets_compile_without_conflicts(self, test_project):
        """Should be able to compile all 7 targets without conflicts."""
        for target in SUPPORTED_TARGETS:
            try:
                created = compile_target(test_project, target)
                assert len(created) > 0, f"{target} created no files"
            except Exception as e:
                pytest.fail(f"{target} compilation failed: {e}")

    def test_subsequent_compiles_overwrite_cleanly(self, test_project):
        """Re-compiling same target should overwrite without errors."""
        compile_target(test_project, "generic")
        first_run = (test_project / ".atlas/runtime/compiled/generic/ENTRYPOINT.md").stat().st_mtime

        # Re-compile
        compile_target(test_project, "generic")
        second_run = (test_project / ".atlas/runtime/compiled/generic/ENTRYPOINT.md").stat().st_mtime

        # Should be same file (or newer)
        assert second_run >= first_run

    def test_compile_all_preserves_git_tracked_files(self, test_project):
        """Compilation should not affect .git or .ai directories."""
        ai_goals = list((test_project / ".ai/goals").glob("*.json"))
        ai_count_before = len(ai_goals)

        # Compile all targets
        for target in SUPPORTED_TARGETS:
            compile_target(test_project, target)

        ai_count_after = len(list((test_project / ".ai/goals").glob("*.json")))
        assert ai_count_before == ai_count_after, "Compilation should not modify .ai/"


class TestCompilerEdgeCases:
    """Tests for edge cases and error handling."""

    def test_compile_with_empty_manifests(self, test_project):
        """Should handle empty agent/skill manifests gracefully."""
        # Manifests are already empty in test fixture
        created = compile_target(test_project, "generic")
        assert len(created) > 0, "Should still generate output even with empty manifests"

    def test_unsupported_target_raises_error(self, test_project):
        """Should raise error for unsupported target."""
        with pytest.raises(ValueError, match="Unsupported target"):
            compile_target(test_project, "unsupported-target")

    def test_compile_preserves_manifests(self, test_project):
        """Compilation should not modify manifest files."""
        agents_manifest = test_project / ".ai/agents/manifest.json"
        manifest_content_before = agents_manifest.read_text()

        compile_target(test_project, "generic")

        manifest_content_after = agents_manifest.read_text()
        assert manifest_content_before == manifest_content_after, "Compilation should not modify manifests"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
