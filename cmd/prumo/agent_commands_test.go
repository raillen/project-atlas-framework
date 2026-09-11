package main

import (
	"strings"
	"testing"
)

func TestAgentRunResumeHandoff(t *testing.T) {
	dir := t.TempDir()
	code, out := captureOutput(func() int {
		return run([]string{"agent", "run", "--goal", "headless eval", "--path", dir, "--run", "R-harness-eval", "--max-turns", "2"})
	})
	if code != 0 || !strings.Contains(out, "R-harness-eval") {
		t.Fatalf("agent run failed: code=%d out=%s", code, out)
	}
	code, out = captureOutput(func() int {
		return run([]string{"--json", "agent", "resume", "--run", "R-harness-eval", "--path", dir})
	})
	if code != 0 || !strings.Contains(out, "R-harness-eval") {
		t.Fatalf("agent resume failed: code=%d out=%s", code, out)
	}
	code, out = captureOutput(func() int {
		return run([]string{"agent", "handoff", "--run", "R-harness-eval", "--path", dir, "--to", "codex"})
	})
	if code != 0 || !strings.Contains(out, "Handoff") {
		t.Fatalf("agent handoff failed: code=%d out=%s", code, out)
	}
}
