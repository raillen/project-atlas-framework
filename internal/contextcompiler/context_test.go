package contextcompiler

import "testing"

func TestCompileDeterministic(t *testing.T) {
	m := Compile("R1", []Source{{Ref: "b", Authority: "canonical", TokenCost: 60}, {Ref: "a", Authority: "canonical", TokenCost: 20}, {Ref: "a", Authority: "canonical", TokenCost: 20}}, 70)
	if m.EstimatedTokens != 20 || !m.Sources[0].Included {
		t.Fatalf("manifest: %#v", m)
	}
	if m.Pressure != "healthy" {
		t.Fatal(m.Pressure)
	}
}
