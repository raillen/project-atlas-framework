package repositorypolicy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRepositoryConformanceFixture(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("root: %v", err)
	}
	policy, err := Load(filepath.Join(root, ".atlas", "repository", "policy.json"))
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "conformance", "fixtures", "repository-policy.json"))
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	var fixture struct {
		ValidBranches   []string `json:"valid_branches"`
		InvalidBranches []string `json:"invalid_branches"`
		ValidCommits    []string `json:"valid_commits"`
		InvalidCommits  []string `json:"invalid_commits"`
		Forbidden       []string `json:"forbidden_operations"`
		ReleaseTags     []string `json:"release_tags"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	for _, branch := range fixture.ValidBranches {
		if err := policy.Branches.Validate(branch, "main"); err != nil {
			t.Fatalf("valid branch %s: %v", branch, err)
		}
	}
	for _, branch := range fixture.InvalidBranches {
		if err := policy.Branches.Validate(branch, "main"); err == nil {
			t.Fatalf("invalid branch accepted %s", branch)
		}
	}
	for _, subject := range fixture.ValidCommits {
		if err := policy.Commits.Validate(subject); err != nil {
			t.Fatalf("valid commit %s: %v", subject, err)
		}
	}
	for _, subject := range fixture.InvalidCommits {
		if err := policy.Commits.Validate(subject); err == nil {
			t.Fatalf("invalid commit accepted %s", subject)
		}
	}
	for _, operation := range fixture.Forbidden {
		if !policy.Agents.IsForbidden(operation) {
			t.Fatalf("operation should be forbidden %s", operation)
		}
	}
	for _, tag := range fixture.ReleaseTags {
		if err := policy.Releases.ValidateTag(tag); err != nil {
			t.Fatalf("release tag %s: %v", tag, err)
		}
	}
}
