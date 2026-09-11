package main

import (
	"net/http"
	"net/http/httptest"
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
