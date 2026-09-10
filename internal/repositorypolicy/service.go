package repositorypolicy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raillen/prumo/internal/scm/github"
)

type RepositoryService struct {
	Root string
}

func NewService(root string) *RepositoryService { return &RepositoryService{Root: root} }

func (s *RepositoryService) Load() (Policy, error) {
	return Load(filepath.Join(s.Root, ".prumo", "repository", "policy.json"))
}

func (s *RepositoryService) CheckRemote(policy Policy, token string) (map[string]any, error) {
	adapter := github.NewAdapter(policy.GitHub.Repository, token)
	remote, err := adapter.RepositoryState()
	if err != nil {
		return nil, err
	}
	rulesets, err := adapter.Rulesets()
	if err != nil {
		return nil, err
	}
	return map[string]any{"repository": remote, "rulesets": rulesets, "checks": CompareRemote(policy, remote, rulesets)}, nil
}

func status(match, available bool) Status {
	if !available {
		return Unavailable
	}
	if match {
		return Compliant
	}
	return NonCompliant
}

func CompareRemote(policy Policy, remote github.RepositoryState, rules []github.RulesetState) []Check {
	return []Check{
		{ID: "github.repository.default_branch", Status: status(remote.DefaultBranch == policy.Repository.DefaultBranch, remote.DefaultBranch != ""), Expected: policy.Repository.DefaultBranch, Actual: remote.DefaultBranch, Severity: "high", Remediable: false, Source: "github", Explanation: "default branch must match repository policy"},
		{ID: "github.merge.squash_only", Status: status(remote.SquashMerge && !remote.MergeCommits && !remote.RebaseMerge, true), Expected: "squash-only", Actual: map[string]bool{"squash": remote.SquashMerge, "merge": remote.MergeCommits, "rebase": remote.RebaseMerge}, Severity: "medium", Remediable: true, Source: "github", Explanation: "only squash merge should be enabled"},
		{ID: "github.branches.delete_after_merge", Status: status(remote.DeleteAfterMerge, true), Expected: true, Actual: remote.DeleteAfterMerge, Severity: "low", Remediable: true, Source: "github", Explanation: "merged branches should be deleted automatically"},
		{ID: "github.ruleset.main", Status: status(hasMainRuleset(rules), true), Expected: "active ruleset for main", Actual: len(rules), Severity: "high", Remediable: true, Source: "github", Explanation: "main must be protected by a pull-request ruleset"},
	}
}

func hasMainRuleset(rules []github.RulesetState) bool {
	for _, ruleset := range rules {
		if ruleset.Target != "tag" && containsFold(ruleset.Name, "main") {
			return true
		}
	}
	return false
}

func containsFold(value, part string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(part))
}

func (s *RepositoryService) Plan(policy Policy, token string) (map[string]any, error) {
	local, err := InspectLocal(s.Root)
	if err != nil {
		return nil, err
	}
	checks := EvaluateLocal(policy, local)
	remoteAvailable := token != ""
	var remote any
	if remoteAvailable {
		remoteResult, err := s.CheckRemote(policy, token)
		if err != nil {
			remote = map[string]any{"status": "unavailable", "error": err.Error()}
		} else {
			remote = remoteResult
		}
	}
	return map[string]any{"policy": policy, "local": local, "remote": remote, "checks": checks, "actions": Plan(policy, checks, remoteAvailable)}, nil
}

func (s *RepositoryService) Apply(policy Policy, token string, dryRun bool) (map[string]any, error) {
	applied := []string{}
	skipped := []string{}
	adapter := github.NewAdapter(policy.GitHub.Repository, token)
	current, err := adapter.RepositoryState()
	if err != nil {
		return nil, err
	}
	settings := map[string]any{}
	if current.MergeCommits {
		settings["allow_merge_commit"] = false
	}
	if current.RebaseMerge {
		settings["allow_rebase_merge"] = false
	}
	if !current.SquashMerge {
		settings["allow_squash_merge"] = true
	}
	if !current.DeleteAfterMerge {
		settings["delete_branch_on_merge"] = true
	}
	if len(settings) > 0 {
		if dryRun {
			skipped = append(skipped, "github.repository.settings")
		} else {
			fresh, err := adapter.RepositoryState()
			if err != nil {
				return nil, err
			}
			if fresh.MergeCommits != current.MergeCommits || fresh.RebaseMerge != current.RebaseMerge || fresh.SquashMerge != current.SquashMerge || fresh.DeleteAfterMerge != current.DeleteAfterMerge {
				return nil, fmt.Errorf("remote repository state changed during apply")
			}
			if err := adapter.UpdateRepositorySettings(settings); err != nil {
				return nil, err
			}
			applied = append(applied, "github.repository.settings")
		}
	}
	rules, err := adapter.Rulesets()
	if err != nil {
		return nil, err
	}
	mainProtected := hasMainRuleset(rules)
	if !mainProtected {
		if dryRun {
			skipped = append(skipped, "github.ruleset.main")
		} else {
			fresh, err := adapter.Rulesets()
			if err != nil {
				return nil, err
			}
			if len(fresh) != len(rules) {
				return nil, fmt.Errorf("remote ruleset state changed during apply")
			}
			if err := adapter.CreateRuleset(MainRulesetPayload(policy)); err != nil {
				return nil, err
			}
			applied = append(applied, "github.ruleset.main")
		}
	}
	return map[string]any{"applied": applied, "skipped": skipped, "dry_run": dryRun, "idempotent": true}, nil
}

func MainRulesetPayload(policy Policy) map[string]any {
	return map[string]any{
		"name":        "Prumo main governance",
		"target":      "branch",
		"enforcement": "active",
		"conditions":  map[string]any{"ref_name": map[string]any{"include": []string{"~DEFAULT_BRANCH"}, "exclude": []string{}}},
		"rules": []map[string]any{
			{"type": "deletion"},
			{"type": "non_fast_forward"},
			{"type": "required_linear_history"},
			{"type": "pull_request", "parameters": map[string]any{"required_approving_review_count": policy.Reviews.RequiredApprovals, "require_code_owner_review": false, "require_last_push_approval": false, "required_review_thread_resolution": policy.Merge.ConversationResolution, "allowed_merge_methods": []string{"squash"}}},
		},
	}
}

func SaveDefault(root string) error {
	path := filepath.Join(root, ".prumo", "repository", "policy.json")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return fmt.Errorf("repository policy not found: %s", path)
}

func Marshal(value any) ([]byte, error) { return json.MarshalIndent(value, "", "  ") }
