package repositorypolicy

import (
	"path/filepath"
	"testing"
)

func TestLoadDogfoodPolicy(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("root: %v", err)
	}
	policy, err := Load(filepath.Join(root, ".prumo", "repository", "policy.json"))
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	if policy.Repository.DefaultBranch != "main" || policy.Merge.Strategy != "squash" {
		t.Fatalf("unexpected policy: %#v", policy)
	}
}
