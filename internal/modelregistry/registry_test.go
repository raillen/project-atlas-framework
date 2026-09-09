package modelregistry

import (
	"testing"
)

func TestRestrictedRequiresLocal(t *testing.T) {
	m := Descriptor{Privacy: External, ContextTokens: 1000}
	if Compatible(m, RouteRequest{DataClass: "restricted"}) {
		t.Fatal("external model accepted restricted data")
	}
}

func TestToolRoute(t *testing.T) {
	m := Descriptor{Privacy: Trusted, ContextTokens: 2000, ToolUse: true, StructuredOutput: true}
	if !Compatible(m, RouteRequest{NeedTools: true, NeedStructuredOutput: true, MinimumContext: 1000}) {
		t.Fatal("compatible model rejected")
	}
}

func TestModelRouterAndFallbacks(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Descriptor{
		ID:               "local-llama",
		Provider:         "ollama",
		Privacy:          Local,
		ContextTokens:    8000,
		ToolUse:          true,
		StructuredOutput: true,
		LatencyClass:     "standard",
	})
	reg.Register(Descriptor{
		ID:               "cloud-fast",
		Provider:         "openai",
		Privacy:          External,
		ContextTokens:    16000,
		ToolUse:          true,
		StructuredOutput: true,
		LatencyClass:     "fast",
	})
	reg.Register(Descriptor{
		ID:               "cloud-deep",
		Provider:         "anthropic",
		Privacy:          External,
		ContextTokens:    32000,
		ToolUse:          true,
		StructuredOutput: true,
		LatencyClass:     "deep",
	})

	// Restricted data route -> only local-llama should qualify
	resp, err := reg.Route(RouteRequest{DataClass: "restricted", MinimumContext: 4000, NeedTools: true})
	if err != nil {
		t.Fatalf("unexpected route error: %v", err)
	}
	if resp.Primary.ID != "local-llama" {
		t.Fatalf("expected local-llama, got: %s", resp.Primary.ID)
	}
	if len(resp.Fallbacks) != 0 {
		t.Fatalf("expected 0 fallbacks for restricted route, got %d", len(resp.Fallbacks))
	}

	// Provider unavailable -> filtered out
	reg.SetHealth("openai", ProviderHealth{Provider: "openai", Status: HealthUnavailable})
	resp2, err := reg.Route(RouteRequest{LatencyPreference: "fast", MinimumContext: 1000})
	if err != nil {
		t.Fatalf("unexpected route error: %v", err)
	}
	if resp2.Primary.ID == "cloud-fast" {
		t.Fatalf("unavailable provider was chosen as primary route")
	}
}

func TestCanaryDriftEvaluation(t *testing.T) {
	eval := EvaluateDrift("model-v1", 0.95, 0.80, 0.10)
	if !eval.DriftDetected {
		t.Fatalf("expected drift detected when drop exceeds tolerance")
	}

	evalNoDrift := EvaluateDrift("model-v1", 0.95, 0.92, 0.10)
	if evalNoDrift.DriftDetected {
		t.Fatalf("did not expect drift detected within tolerance")
	}
}
