package repositorypolicy

import "testing"

func testPolicy() Policy {
	return Policy{SchemaVersion: 1, Repository: Repository{DefaultBranch: "main", Provider: "github", GovernanceProfile: "solo"}, Branches: Branches{AllowedCategories: []string{"feat", "fix"}, Patterns: map[string]string{"feat": `^feat/[a-z0-9._-]+$`, "fix": `^fix/[a-z0-9._-]+$`}, AutomationPrefixes: []string{"dependabot/"}}, Commits: Commits{Types: []string{"feat", "fix"}, SubjectRequired: true}, Pushes: Pushes{DirectMain: "require_pr", ForceMain: "deny", WorkBranchForceWithLease: "conditional"}, Merge: Merge{Strategy: "squash", LinearHistory: true, ConversationResolution: true}, Tags: Tags{Deletion: "deny", Rewrite: "deny"}}
}

func TestBranchPolicy(t *testing.T) {
	policy := testPolicy()
	for _, branch := range []string{"main", "feat/g042-policy", "dependabot/go"} {
		if err := policy.Branches.Validate(branch, "main"); err != nil {
			t.Fatalf("expected valid branch %s: %v", branch, err)
		}
	}
	if err := policy.Branches.Validate("feature/bad", "main"); err == nil {
		t.Fatalf("expected invalid branch")
	}
}

func TestCommitPolicy(t *testing.T) {
	policy := testPolicy()
	if err := policy.Commits.Validate("feat(repo): add governance"); err != nil {
		t.Fatalf("valid commit rejected: %v", err)
	}
	if err := policy.Commits.Validate("wip"); err == nil {
		t.Fatalf("expected invalid commit")
	}
}

func TestForbiddenOperations(t *testing.T) {
	policy := testPolicy()
	policy.Agents.ForbiddenOperations = []string{"force_push_main"}
	if !policy.Agents.IsForbidden("force_push_main") {
		t.Fatalf("expected forbidden operation")
	}
}
