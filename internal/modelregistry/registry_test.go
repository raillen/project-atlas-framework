package modelregistry

import "testing"

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
