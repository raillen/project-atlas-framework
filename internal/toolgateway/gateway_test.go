package toolgateway

import (
	"errors"
	"strings"
	"testing"
)

func TestEvaluateTool(t *testing.T) {
	d := Descriptor{
		ID:              "git_commit",
		Version:         1,
		Kind:            SideEffecting,
		Trust:           "trusted",
		FilesystemScope: []string{"project-root"},
	}

	// Permitted within scope
	dec := Evaluate(d, "/repo", "/repo/main.go", false)
	if !dec.Allowed {
		t.Fatalf("expected allowed, got: %s", dec.Reason)
	}

	// Outside scope
	decOut := Evaluate(d, "/repo", "/etc/passwd", false)
	if decOut.Allowed {
		t.Fatalf("expected denied for target outside scope")
	}

	// Destructive denied in safe mode
	destr := Descriptor{
		ID:      "delete_all",
		Version: 1,
		Kind:    Destructive,
		Trust:   "trusted",
	}
	decSafe := Evaluate(destr, "/repo", "", true)
	if decSafe.Allowed {
		t.Fatalf("expected destructive denied in safe mode")
	}
}

func TestEvaluateMCP(t *testing.T) {
	server := MCPServerDescriptor{
		ID:           "test-mcp",
		Transport:    "stdio",
		Trust:        "untrusted",
		AllowedTools: []string{"read_file"},
		RootScopes:   []string{"/workspace"},
	}

	toolRead := Descriptor{
		ID:              "read_file",
		Version:         1,
		Kind:            ReadOnly,
		Trust:           "untrusted",
		FilesystemScope: []string{"."},
	}

	// Allowed tool and in scope
	dec := EvaluateMCP(server, toolRead, "/workspace/doc.txt", false)
	if !dec.Allowed {
		t.Fatalf("expected allowed mcp call: %s", dec.Reason)
	}

	// Tool not in allowed tools
	toolExec := Descriptor{
		ID:      "execute_cmd",
		Version: 1,
		Kind:    SideEffecting,
		Trust:   "untrusted",
	}
	decDisallowed := EvaluateMCP(server, toolExec, "/workspace/doc.txt", false)
	if decDisallowed.Allowed {
		t.Fatalf("expected disallowed tool to be denied")
	}

	// Destructive tool denied for untrusted MCP server
	toolDestroy := Descriptor{
		ID:      "read_file", // in allowed list but destructive
		Version: 1,
		Kind:    Destructive,
		Trust:   "untrusted",
	}
	decDestr := EvaluateMCP(server, toolDestroy, "/workspace/doc.txt", false)
	if decDestr.Allowed {
		t.Fatalf("expected destructive tool denied for untrusted MCP server")
	}
}

func TestToolRegistryLazyDiscovery(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Descriptor{
		ID:           "git_commit",
		Version:      1,
		Description:  "Create a git commit",
		Kind:         SideEffecting,
		Trust:        "trusted",
		Capabilities: []string{"scm", "git"},
	})
	reg.Register(Descriptor{
		ID:           "read_file",
		Version:      1,
		Description:  "Read file contents",
		Kind:         ReadOnly,
		Trust:        "core",
		Capabilities: []string{"fs", "read"},
	})

	if len(reg.List()) != 2 {
		t.Fatalf("expected 2 tools in registry")
	}

	// Discover with capability filter
	found := reg.Discover("scm")
	if len(found) != 1 || found[0].ID != "git_commit" {
		t.Fatalf("expected git_commit discovered for scm, got: %+v", found)
	}

	// Discover all
	all := reg.Discover()
	if len(all) != 2 {
		t.Fatalf("expected all 2 tools, got: %d", len(all))
	}
}

func TestOutputBudgetEnforcement(t *testing.T) {
	text := strings.Repeat("a", 100)
	out, truncated, ptr := EnforceOutputBudget(text, 50)
	if !truncated {
		t.Fatalf("expected truncated")
	}
	if !strings.Contains(out, "output truncated: 100 bytes") {
		t.Fatalf("expected truncation message in output: %s", out)
	}
	if !strings.HasPrefix(ptr, "sha256:") {
		t.Fatalf("expected sha256 pointer, got: %s", ptr)
	}

	// Under limit
	out2, trunc2, _ := EnforceOutputBudget(text, 200)
	if trunc2 || out2 != text {
		t.Fatalf("expected not truncated")
	}
}

func TestSideEffectEnforcement(t *testing.T) {
	intent := Intent{
		ToolID:         "write_file",
		IdempotencyKey: "key-123",
		Target:         "/workspace/test.txt",
		Timestamp:      "2026-09-09T00:00:00Z",
	}

	if err := ValidateIntent(intent); err != nil {
		t.Fatalf("expected valid intent, got: %v", err)
	}

	badIntent := Intent{ToolID: "write_file"}
	if err := ValidateIntent(badIntent); err == nil {
		t.Fatalf("expected error for intent missing idempotency_key")
	}

	outcome := RecordOutcome(intent, true, []byte("result data"), nil)
	if !outcome.Success || !strings.HasPrefix(outcome.OutputHash, "sha256:") {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}

	errOutcome := RecordOutcome(intent, false, nil, errors.New("write failed"))
	if errOutcome.Success || errOutcome.Error != "write failed" {
		t.Fatalf("unexpected error outcome: %+v", errOutcome)
	}
}
