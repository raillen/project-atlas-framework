package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/raillen/prumo/internal/harness/agent"
	"github.com/raillen/prumo/internal/harness/daemon"
	"github.com/raillen/prumo/internal/harness/model"
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

func TestAgentProtocolManifest(t *testing.T) {
	code, out := captureOutput(func() int {
		return run([]string{"--json", "agent", "protocol", "--manifest"})
	})
	if code != 0 {
		t.Fatalf("protocol manifest failed: code=%d", code)
	}
	for _, want := range []string{`"start"`, `"cancel"`, `"protocol"`, `"0.1.0"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("manifest missing %s:\n%s", want, out)
		}
	}
}

func TestAgentEventsAndProtocol(t *testing.T) {
	dir := t.TempDir()
	code, out := captureOutput(func() int {
		return run([]string{"agent", "run", "--goal", "eval events", "--path", dir, "--run", "R-ev", "--max-turns", "1"})
	})
	if code != 0 {
		t.Fatalf("agent run failed: code=%d out=%s", code, out)
	}
	code, out = captureOutput(func() int {
		return run([]string{"--json", "agent", "events", "--run", "R-ev", "--path", dir})
	})
	if code != 0 || !strings.Contains(out, "R-ev") {
		t.Fatalf("agent events failed: code=%d out=%s", code, out)
	}
	code, out = captureOutput(func() int {
		return run([]string{"agent", "protocol", "--client", "0.1.0"})
	})
	if code != 0 || !strings.Contains(out, "protocol 0.1.0") {
		t.Fatalf("agent protocol failed: code=%d out=%s", code, out)
	}
	code, _ = captureOutput(func() int {
		return run([]string{"agent", "events", "--run", "R-missing", "--path", dir})
	})
	if code == 0 {
		t.Fatal("events for unknown run must fail")
	}
}

type cliStubTools struct{}

func (cliStubTools) Execute(_ context.Context, call agent.ToolCall) (agent.ToolResult, error) {
	return agent.ToolResult{ToolCallID: call.ID, Output: "ok"}, nil
}
func (cliStubTools) KindOf(string) string { return "read-only" }

func TestAgentPsLogsAgainstDaemon(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "agentd.sock")
	srv := daemon.New(sock, filepath.Join(dir, "store"), daemon.Deps{
		NewProvider: func(name, baseURL, apiKey, mdl string) (model.Provider, error) {
			return model.NewFake(map[string][]model.ScriptStep{"*": {{Kind: "text", Text: "hi"}, {Kind: "complete"}}}), nil
		},
		Tools: cliStubTools{},
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx) }()
	c := daemon.Client{SocketPath: sock}
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := c.Protocol(); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("daemon did not come up")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := c.Start("daemon cli goal", "fake", "R-dcli", 1); err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(10 * time.Second)
	for {
		st, err := c.Status("R-dcli")
		if err != nil {
			t.Fatal(err)
		}
		if st["ok"] == true && st["status"] == "complete" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("run did not complete: %v", st)
		}
		time.Sleep(20 * time.Millisecond)
	}
	code, out := captureOutput(func() int {
		return run([]string{"agent", "ps", "--socket", sock})
	})
	if code != 0 || !strings.Contains(out, "R-dcli") {
		t.Fatalf("agent ps failed: code=%d out=%s", code, out)
	}
	code, out = captureOutput(func() int {
		return run([]string{"--json", "agent", "logs", "--run", "R-dcli", "--socket", sock})
	})
	if code != 0 || !strings.Contains(out, "R-dcli") {
		t.Fatalf("agent logs failed: code=%d out=%s", code, out)
	}
}
func TestAgentProvidersListsMatrix(t *testing.T) {
	// Hermetic PATH: no real binaries execute; wiring (not versions) is
	// under test here — parsing is covered in the extagent package.
	t.Setenv("PATH", t.TempDir())
	code, out := captureOutput(func() int {
		return run([]string{"agent", "providers"})
	})
	if code != 0 {
		t.Fatalf("providers failed: code=%d", code)
	}
	for _, want := range []string{"fake", "openai-compat", "anthropic", "opencode-cli", "opencode-server", "codex-cli", "acp-generic"} {
		if !strings.Contains(out, want) {
			t.Fatalf("matrix missing %q:\n%s", want, out)
		}
	}
	code, out = captureOutput(func() int {
		return run([]string{"--json", "agent", "providers"})
	})
	if code != 0 || !strings.Contains(out, `"providers"`) {
		t.Fatalf("providers json failed: code=%d out=%s", code, out)
	}
}
func TestAgentRunPersistsContextManifest(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _ := captureOutput(func() int {
		return run([]string{"agent", "run", "--goal", "ship it", "--path", dir, "--run", "R-ctx", "--max-turns", "1", "--context-budget", "4000"})
	})
	if code != 0 {
		t.Fatalf("agent run failed: code=%d", code)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".prumo", "runtime", "harness", "context-R-ctx.json"))
	if err != nil {
		t.Fatalf("context manifest not persisted: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["version"] != float64(2) {
		t.Fatalf("expected manifest v2: %v", m["version"])
	}
	included, _ := m["included"].([]any)
	if len(included) == 0 {
		t.Fatal("expected packed sources")
	}
	first, _ := included[0].(map[string]any)
	if first["ref"] != "goal" {
		t.Fatalf("goal must pack first: %v", first["ref"])
	}
}
func TestAgentSandboxFlagsFailFast(t *testing.T) {
	dir := t.TempDir()
	code, _ := captureOutput(func() int {
		return run([]string{"agent", "run", "--goal", "x", "--path", dir, "--run", "R-sbx", "--sandbox", "bogus"})
	})
	if code == 0 {
		t.Fatal("bogus sandbox must fail")
	}
	code, _ = captureOutput(func() int {
		return run([]string{"agent", "run", "--goal", "x", "--path", dir, "--run", "R-sbx", "--sandbox", "container"})
	})
	if code == 0 {
		t.Fatal("container without image must fail fast")
	}
	if _, err := agentTools(dir, map[string]string{"sandbox": "container"}); err == nil || !strings.Contains(err.Error(), "sandbox-image") {
		t.Fatalf("expected sandbox-image error, got %v", err)
	}
}
func TestAgentRunAnthropicAgainstStub(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"ok\"}}\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"))
	}))
	defer srv.Close()
	dir := t.TempDir()
	code, out := captureOutput(func() int {
		return run([]string{"agent", "run", "--goal", "anthropic stub", "--path", dir, "--run", "R-anthropic", "--provider", "anthropic", "--base-url", srv.URL, "--model", "stub", "--max-turns", "1"})
	})
	if code != 0 || !strings.Contains(out, "R-anthropic") {
		t.Fatalf("anthropic run failed: code=%d out=%s", code, out)
	}
}
