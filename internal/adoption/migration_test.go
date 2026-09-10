package adoption

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/raillen/prumo/internal/migrations"
)

func TestGenerateMigrationProposals(t *testing.T) {
	report := AdoptionReport{
		Repository: RepoSummary{
			Root: "/test/repo",
		},
		Classification: RepositoryClassification{
			AppTypes:   []string{"application"},
			Languages:  []string{"go"},
			Frameworks: []string{"gin"},
		},
		Profiles: []ProfileCandidate{
			{Profile: "service", Confidence: ConfidenceHigh},
		},
		Capabilities: []CapabilityProposal{
			{Capability: "ci-cd", Confidence: ConfidenceHigh},
		},
		PrumoArtifacts: []string{}, // no prumo.json
		DocBindings: []CandidateBinding{
			{
				ContractID: "architecture",
				Sources:    []string{"docs/arch.md"},
			},
		},
	}

	proposals := GenerateMigrationProposals(report)
	if len(proposals) != 2 {
		t.Fatalf("expected 2 migration proposals, got %d", len(proposals))
	}

	// 1. Manifest proposal
	initProp := proposals[0]
	if initProp.ID != "amp-init-prumo" {
		t.Errorf("expected proposal ID amp-init-prumo, got %s", initProp.ID)
	}
	if initProp.Status != ProposalStatusPendingReview {
		t.Errorf("expected pending-review status, got %s", initProp.Status)
	}
	if !initProp.ReviewRequired {
		t.Errorf("expected review_required=true")
	}
	if len(initProp.Actions) != 1 || initProp.Actions[0].TargetPath != "prumo.json" {
		t.Errorf("expected create action for prumo.json")
	}
	if !initProp.Contract.Reversible {
		t.Errorf("expected contract reversible=true")
	}

	// 2. Doc bindings proposal
	docProp := proposals[1]
	if docProp.ID != "amp-doc-bindings" {
		t.Errorf("expected proposal ID amp-doc-bindings, got %s", docProp.ID)
	}
	if len(docProp.Actions) != 1 || !strings.Contains(docProp.Actions[0].TargetPath, "bindings.json") {
		t.Errorf("expected action targeting bindings.json, got %v", docProp.Actions)
	}

	// Case where prumo.json already exists
	reportWithPrumo := report
	reportWithPrumo.PrumoArtifacts = []string{"prumo.json"}
	props2 := GenerateMigrationProposals(reportWithPrumo)
	for _, p := range props2 {
		if p.ID == "amp-init-prumo" {
			t.Errorf("should not propose amp-init-prumo when prumo.json already exists")
		}
	}
}

func TestReviewQueueWorkflow(t *testing.T) {
	q := NewReviewQueue()
	proposal := AdoptionMigrationProposal{
		ID:             "amp-test",
		Title:          "Test proposal",
		Status:         ProposalStatusPendingReview,
		ReviewRequired: true,
	}

	item := q.Submit(proposal)
	if item.ID != "rq-amp-test" {
		t.Errorf("unexpected queue item ID: %s", item.ID)
	}

	pending := q.ListPending()
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending item, got %d", len(pending))
	}

	err := q.Approve(item.ID, "reviewer@example.com", "Approved after security check")
	if err != nil {
		t.Fatalf("failed to approve: %v", err)
	}

	if len(q.ListPending()) != 0 {
		t.Errorf("expected 0 pending items after approval")
	}
	if q.Items[0].Proposal.ApprovedBy != "reviewer@example.com" {
		t.Errorf("expected approved_by reviewer@example.com, got %s", q.Items[0].Proposal.ApprovedBy)
	}

	// Test reject on another item
	p2 := proposal
	p2.ID = "amp-test-2"
	i2 := q.Submit(p2)
	err = q.Reject(i2.ID, "reviewer@example.com", "Rejected: unsafe layout")
	if err != nil {
		t.Fatalf("failed to reject: %v", err)
	}
	if q.Items[1].Status != ProposalStatusRejected {
		t.Errorf("expected status rejected, got %s", q.Items[1].Status)
	}
}

func TestDryRunAndApply(t *testing.T) {
	tmpDir := t.TempDir()

	proposal := AdoptionMigrationProposal{
		ID:             "amp-init-prumo",
		Title:          "Initialize Prumo manifest",
		Status:         ProposalStatusPendingReview,
		ReviewRequired: true,
		Actions: []MigrationAction{
			{
				Type:        ActionCreateFile,
				TargetPath:  "prumo.json",
				Content:     "{\"version\":1}\n",
				Description: "Create prumo.json",
				Reversible:  true,
			},
		},
		Contract: migrations.Contract{
			ID:                "contract-amp-init-prumo",
			FromVersion:       "0.0.0",
			ToVersion:         "0.5.0",
			Preconditions:     []string{"manifest_absent:prumo.json"},
			BackupStrategy:    "none",
			AffectedArtifacts: []string{"prumo.json"},
			Reversible:        true,
			Rollback:          "rm prumo.json",
		},
	}

	// 1. Dry run on empty dir
	dryRun, err := DryRun(tmpDir, proposal)
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}
	if !dryRun.PreconditionsPassed {
		t.Errorf("expected preconditions passed in dry run")
	}
	if len(dryRun.Plan) != 1 || !strings.Contains(dryRun.Plan[0].DiffPreview, "+++ prumo.json") {
		t.Errorf("unexpected dry run diff: %v", dryRun.Plan)
	}

	// 2. Invariant: applying unapproved proposal MUST fail
	_, err = Apply(tmpDir, proposal)
	if err == nil {
		t.Fatalf("expected Apply to fail on unapproved proposal, got nil")
	}

	// 3. Approve proposal and apply
	proposal.Status = ProposalStatusApproved
	proposal.ApprovedBy = "operator"

	applyRes, err := Apply(tmpDir, proposal)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if !applyRes.Success {
		t.Errorf("expected apply success")
	}
	if applyRes.JournalEntry.MigrationID != proposal.Contract.ID {
		t.Errorf("unexpected journal migration ID: %s", applyRes.JournalEntry.MigrationID)
	}
	if applyRes.JournalEntry.Hash == "" {
		t.Errorf("expected non-empty journal hash")
	}

	// Check file was created
	written, err := os.ReadFile(filepath.Join(tmpDir, "prumo.json"))
	if err != nil {
		t.Fatalf("failed reading created file: %v", err)
	}
	if string(written) != "{\"version\":1}\n" {
		t.Errorf("unexpected file content: %s", string(written))
	}

	// 4. Precondition check: running dry run now should fail because prumo.json exists
	dryRun2, err := DryRun(tmpDir, proposal)
	if err != nil {
		t.Fatalf("dry run 2 failed: %v", err)
	}
	if dryRun2.PreconditionsPassed {
		t.Errorf("expected preconditions to fail since prumo.json already exists")
	}
}
