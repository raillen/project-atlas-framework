package protocol

import (
	"encoding/json"
	"testing"
)

func TestOkEnvelope(t *testing.T) {
	env := OkEnvelope(map[string]string{"version": "0.4.0"})
	if !env.Ok {
		t.Fatalf("expected ok=true, got false")
	}
	if env.ProtocolVersion != ProtocolVersion {
		t.Fatalf("expected protocol %s, got %s", ProtocolVersion, env.ProtocolVersion)
	}
	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("empty json output")
	}
}

func TestErrEnvelope(t *testing.T) {
	env := ErrEnvelope(Diagnostic{Code: "ERR_TEST", Message: "test error"})
	if env.Ok {
		t.Fatalf("expected ok=false, got true")
	}
	if len(env.Diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(env.Diagnostics))
	}
}
