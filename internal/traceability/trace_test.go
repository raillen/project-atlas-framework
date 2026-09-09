package traceability

import (
	"testing"
)

func TestTraceGraph(t *testing.T) {
	g := NewGraph()

	// 1. Add nodes: Req -> Dec -> Goal -> Code -> Test -> Doc -> Evidence
	reqNode := Node{ID: "req-001", Kind: NodeRequirement, Title: "Provider Neutrality"}
	decNode := Node{ID: "dec-001", Kind: NodeDecision, Title: "Use Go Standard Library for HTTP"}
	goalNode := Node{ID: "goal-m8", Kind: NodeGoal, Title: "M8 History and Traceability"}
	codeNode := Node{ID: "code-trace", Kind: NodeCode, Title: "Traceability Engine", Ref: "internal/traceability/graph.go"}
	testNode := Node{ID: "test-trace", Kind: NodeTest, Title: "Traceability Tests", Ref: "internal/traceability/trace_test.go"}
	docNode := Node{ID: "doc-exit", Kind: NodeDoc, Title: "M8 Exit Gate", Ref: "docs/governance/m8-exit-gate.md"}
	eviNode := Node{ID: "evi-ci", Kind: NodeEvidence, Title: "CI Test Pass Evidence"}

	nodes := []Node{reqNode, decNode, goalNode, codeNode, testNode, docNode, eviNode}
	for _, n := range nodes {
		if err := g.AddNode(n); err != nil {
			t.Fatalf("failed to add node %s: %v", n.ID, err)
		}
	}

	// Invariant: Invalid node kind rejected
	if err := g.AddNode(Node{ID: "bad", Kind: "invalid-kind", Title: "Bad"}); err == nil {
		t.Errorf("expected error adding node with invalid kind")
	}

	// 2. Add edges
	edges := []Edge{
		{From: "dec-001", To: "req-001", Kind: string(EdgeSatisfies)},
		{From: "goal-m8", To: "dec-001", Kind: string(EdgeDerivesFrom)},
		{From: "code-trace", To: "goal-m8", Kind: string(EdgeImplements)},
		{From: "test-trace", To: "code-trace", Kind: string(EdgeVerifies)},
		{From: "doc-exit", To: "goal-m8", Kind: string(EdgeDocuments)},
		{From: "evi-ci", To: "test-trace", Kind: string(EdgeEvidencedBy)},
	}
	for _, e := range edges {
		if err := g.AddEdge(e); err != nil {
			t.Fatalf("failed to add edge %s: %v", e.ID, err)
		}
	}

	// Invariant: Edge to non-existent node rejected
	if err := g.AddEdge(Edge{From: "code-trace", To: "non-existent", Kind: string(EdgeImplements)}); err == nil {
		t.Errorf("expected error adding edge with non-existent target")
	}

	// 3. Trace from code-trace:
	// Upstream: goal-m8 -> dec-001 -> req-001
	// Downstream: test-trace -> evi-ci
	trace, err := g.Trace("internal/traceability/graph.go")
	if err != nil {
		t.Fatalf("trace failed: %v", err)
	}

	if trace.Root.ID != "code-trace" {
		t.Errorf("expected root code-trace, got %s", trace.Root.ID)
	}

	upMap := make(map[string]bool)
	for _, u := range trace.Upstream {
		upMap[u.ID] = true
	}
	if !upMap["goal-m8"] || !upMap["dec-001"] || !upMap["req-001"] {
		t.Errorf("missing expected upstream nodes: %v", trace.Upstream)
	}

	downMap := make(map[string]bool)
	for _, d := range trace.Downstream {
		downMap[d.ID] = true
	}
	if !downMap["test-trace"] || !downMap["evi-ci"] {
		t.Errorf("missing expected downstream nodes: %v", trace.Downstream)
	}
}

func TestImplementationJournal(t *testing.T) {
	j := NewJournal()

	entry := JournalEntry{
		ID:          "jour-001",
		Goal:        "M8",
		Title:       "Traceability Graph Implementation",
		Summary:     "Implemented typed graph linking requirements, decisions, code, tests, docs, and evidence.",
		Decisions:   []string{"dec-001"},
		CodeChanges: []string{"internal/traceability/graph.go", "internal/traceability/journal.go"},
		Tests:       []string{"internal/traceability/trace_test.go"},
		Evidence:    []string{"hash:abcd1234"},
	}

	if err := j.Add(entry); err != nil {
		t.Fatalf("failed to add journal entry: %v", err)
	}

	// Invariant: entry without ID or Summary rejected
	if err := j.Add(JournalEntry{ID: "", Summary: "test"}); err == nil {
		t.Errorf("expected error adding entry without ID")
	}
	if err := j.Add(JournalEntry{ID: "jour-bad", Summary: ""}); err == nil {
		t.Errorf("expected error adding entry without Summary")
	}

	// Query by Goal
	byGoal := j.Query(JournalFilter{Goal: "M8"})
	if len(byGoal) != 1 {
		t.Fatalf("expected 1 entry for goal M8, got %d", len(byGoal))
	}

	// Query by File
	byFile := j.Query(JournalFilter{File: "graph.go"})
	if len(byFile) != 1 {
		t.Fatalf("expected 1 entry for file graph.go, got %d", len(byFile))
	}
}

func TestRegisters(t *testing.T) {
	r := NewRegisters()

	// 1. Experiment
	err := r.RecordExperiment(ExperimentEntry{
		ID:         "exp-001",
		Goal:       "M8",
		Hypothesis: "BFS traversal performs < 5ms on graphs up to 10k nodes",
		Method:     "Synthetic graph benchmark",
		Outcome:    "Traversal completes in 0.8ms",
		Conclusion: "Accepted BFS for bidirectional graph trace",
	})
	if err != nil {
		t.Fatalf("failed to record experiment: %v", err)
	}

	// 2. Rejection
	err = r.RecordRejection(RejectionEntry{
		ID:          "rej-001",
		ProposalID:  "prop-raw-telemetry",
		Title:       "Persisting full raw LLM transcripts into journal",
		Alternative: "Structured synthesized summaries with evidence pointers",
		Rationale:   "Violates Lean Progressive Context and bloats repository with untracked tokens",
		RejectedBy:  "Architectural Governance",
	})
	if err != nil {
		t.Fatalf("failed to record rejection: %v", err)
	}

	// 3. Technical Debt
	err = r.RecordDebt(DebtEntry{
		ID:                "debt-001",
		Title:             "In-memory graph requires streaming for repositories > 100k files",
		Severity:          DebtSeverityLow,
		RemediationPlan:   "Introduce SQLite backing store in M12 if repository scale exceeds 100k artifacts",
		ImpactedContracts: []string{"traceability.scale"},
	})
	if err != nil {
		t.Fatalf("failed to record debt: %v", err)
	}

	if len(r.Experiments) != 1 || len(r.Rejections) != 1 || len(r.Debt) != 1 {
		t.Errorf("unexpected counts in registers: exp=%d, rej=%d, debt=%d",
			len(r.Experiments), len(r.Rejections), len(r.Debt))
	}
}

func TestStorePersistence(t *testing.T) {
	tmpDir := t.TempDir()

	g := NewGraph()
	_ = g.AddNode(Node{ID: "n1", Kind: NodeGoal, Title: "Goal 1"})
	_ = g.AddNode(Node{ID: "n2", Kind: NodeCode, Title: "Code 1"})
	_ = g.AddEdge(Edge{From: "n2", To: "n1", Kind: string(EdgeImplements)})

	j := NewJournal()
	_ = j.Add(JournalEntry{ID: "j1", Summary: "Synthesis summary", Goal: "G1"})

	r := NewRegisters()
	_ = r.RecordDebt(DebtEntry{ID: "d1", Title: "Sample debt", Severity: DebtSeverityLow})

	if err := SaveGraph(tmpDir, g); err != nil {
		t.Fatalf("SaveGraph failed: %v", err)
	}
	if err := SaveJournal(tmpDir, j); err != nil {
		t.Fatalf("SaveJournal failed: %v", err)
	}
	if err := SaveRegisters(tmpDir, r); err != nil {
		t.Fatalf("SaveRegisters failed: %v", err)
	}

	loadedG, err := LoadGraph(tmpDir)
	if err != nil {
		t.Fatalf("LoadGraph failed: %v", err)
	}
	if len(loadedG.Nodes) != 2 || len(loadedG.Edges) != 1 {
		t.Errorf("mismatched loaded graph: nodes=%d, edges=%d", len(loadedG.Nodes), len(loadedG.Edges))
	}

	loadedJ, err := LoadJournal(tmpDir)
	if err != nil {
		t.Fatalf("LoadJournal failed: %v", err)
	}
	if len(loadedJ.Entries) != 1 || loadedJ.Entries[0].ID != "j1" {
		t.Errorf("mismatched loaded journal")
	}

	loadedR, err := LoadRegisters(tmpDir)
	if err != nil {
		t.Fatalf("LoadRegisters failed: %v", err)
	}
	if len(loadedR.Debt) != 1 || loadedR.Debt[0].ID != "d1" {
		t.Errorf("mismatched loaded registers")
	}
}
