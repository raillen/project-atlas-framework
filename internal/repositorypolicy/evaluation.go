package repositorypolicy

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type LocalState struct {
	RepositoryRoot string `json:"repository_root"`
	Branch         string `json:"branch"`
	Clean          bool   `json:"clean"`
	HeadSubject    string `json:"head_subject"`
	Origin         string `json:"origin"`
}

func InspectLocal(root string) (LocalState, error) {
	state := LocalState{RepositoryRoot: root}
	if value, err := git(root, "rev-parse", "--show-toplevel"); err == nil {
		state.RepositoryRoot = strings.TrimSpace(value)
	} else {
		return state, fmt.Errorf("not a Git repository: %w", err)
	}
	var err error
	if state.Branch, err = git(root, "branch", "--show-current"); err != nil {
		return state, fmt.Errorf("inspect branch: %w", err)
	}
	status, err := git(root, "status", "--porcelain")
	if err != nil {
		return state, fmt.Errorf("inspect status: %w", err)
	}
	state.Clean = strings.TrimSpace(status) == ""
	state.HeadSubject, _ = git(root, "log", "-1", "--format=%s")
	state.Origin, _ = git(root, "remote", "get-url", "origin")
	state.Branch = strings.TrimSpace(state.Branch)
	state.HeadSubject = strings.TrimSpace(state.HeadSubject)
	state.Origin = strings.TrimSpace(state.Origin)
	return state, nil
}

func git(root string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

type Status string

const (
	Compliant       Status = "compliant"
	NonCompliant    Status = "non_compliant"
	Unavailable     Status = "unavailable"
	Unsupported     Status = "unsupported"
	EvaluationError Status = "error"
)

type Check struct {
	ID          string `json:"id"`
	Status      Status `json:"status"`
	Expected    any    `json:"expected"`
	Actual      any    `json:"actual"`
	Severity    string `json:"severity"`
	Remediable  bool   `json:"remediable"`
	Source      string `json:"source"`
	Explanation string `json:"explanation"`
}

type Action struct {
	ID                 string `json:"id"`
	Description        string `json:"description"`
	SideEffect         string `json:"side_effect"`
	Reversibility      string `json:"reversibility"`
	RequiredPermission string `json:"required_permission"`
	Risk               string `json:"risk"`
}

func PolicyPath(root string) string {
	return filepath.Join(root, ".prumo", "repository", "policy.json")
}

func LoadPolicy(root string) (Policy, error) {
	data, err := os.ReadFile(PolicyPath(root))
	if err != nil {
		return Policy{}, err
	}
	var policy Policy
	if err := json.Unmarshal(data, &policy); err != nil {
		return policy, err
	}
	if err := Validate(policy); err != nil {
		return policy, err
	}
	return policy, nil
}

func EvaluateLocal(policy Policy, state LocalState) []Check {
	checks := []Check{}
	branchStatus := Compliant
	branchExplanation := "current branch matches repository policy"
	if err := policy.Branches.Validate(state.Branch, policy.Repository.DefaultBranch); err != nil {
		branchStatus = NonCompliant
		branchExplanation = err.Error()
	}
	checks = append(checks, Check{ID: "local.branch.naming", Status: branchStatus, Expected: policy.Repository.DefaultBranch + " or allowed work/automation branch", Actual: state.Branch, Severity: "medium", Remediable: false, Source: "local.git", Explanation: branchExplanation})
	cleanStatus := Compliant
	cleanExplanation := "working tree is clean"
	if !state.Clean {
		cleanStatus = NonCompliant
		cleanExplanation = "working tree has uncommitted changes"
	}
	checks = append(checks, Check{ID: "local.working_tree.clean", Status: cleanStatus, Expected: true, Actual: state.Clean, Severity: "low", Remediable: false, Source: "local.git", Explanation: cleanExplanation})
	commitStatus := Compliant
	commitExplanation := "HEAD subject matches commit policy"
	if err := policy.Commits.Validate(state.HeadSubject); err != nil {
		commitStatus = NonCompliant
		commitExplanation = err.Error()
	}
	checks = append(checks, Check{ID: "local.commit.subject", Status: commitStatus, Expected: policy.Commits.Format, Actual: state.HeadSubject, Severity: "medium", Remediable: false, Source: "local.git", Explanation: commitExplanation})
	return checks
}

func Plan(policy Policy, checks []Check, remoteAvailable bool) []Action {
	actions := []Action{}
	for _, check := range checks {
		if check.Status != NonCompliant || !check.Remediable {
			continue
		}
		actions = append(actions, Action{ID: check.ID, Description: check.Explanation, SideEffect: "remote", Reversibility: "policy-reversible", RequiredPermission: "admin", Risk: check.Severity})
	}
	if remoteAvailable {
		for _, action := range remoteActions(policy) {
			actions = append(actions, action)
		}
	}
	return actions
}

func remoteActions(policy Policy) []Action {
	return []Action{
		{ID: "github.merge.squash_only", Description: "Enable squash merge and disable merge commits and rebase merges", SideEffect: "remote", Reversibility: "reversible", RequiredPermission: "admin", Risk: "medium"},
		{ID: "github.branches.delete_after_merge", Description: "Enable automatic deletion of merged branches", SideEffect: "remote", Reversibility: "reversible", RequiredPermission: "admin", Risk: "low"},
		{ID: "github.ruleset.main", Description: "Ensure main is protected by a pull-request ruleset", SideEffect: "remote", Reversibility: "reversible", RequiredPermission: "admin", Risk: "high"},
		{ID: "github.ruleset.release_tags", Description: "Protect semantic release tags from deletion and rewrite", SideEffect: "remote", Reversibility: "reversible", RequiredPermission: "admin", Risk: "high"},
	}
}
