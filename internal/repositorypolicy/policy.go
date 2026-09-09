package repositorypolicy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Policy struct {
	SchemaVersion int             `json:"schema_version"`
	Repository    Repository      `json:"repository"`
	Branches      Branches        `json:"branches"`
	Commits       Commits         `json:"commits"`
	Pushes        Pushes          `json:"pushes"`
	Issues        Issues          `json:"issues"`
	PullRequests  PullRequests    `json:"pull_requests"`
	Reviews       Reviews         `json:"reviews"`
	Merge         Merge           `json:"merge"`
	Agents        Agents          `json:"agents"`
	Releases      Releases        `json:"releases"`
	Tags          Tags            `json:"tags"`
	GitHub        GitHubPolicy    `json:"github"`
	Emergency     EmergencyBypass `json:"emergency_bypass"`
}

type Repository struct {
	DefaultBranch     string `json:"default_branch"`
	Provider          string `json:"provider"`
	GovernanceProfile string `json:"governance_profile"`
}

type Branches struct {
	AllowedCategories  []string          `json:"allowed_categories"`
	Patterns           map[string]string `json:"patterns"`
	AutomationPrefixes []string          `json:"automation_prefixes"`
}

func (b Branches) Category(branch string) string {
	if branch == "" {
		return ""
	}
	if idx := strings.Index(branch, "/"); idx > 0 {
		return branch[:idx]
	}
	return ""
}

func (b Branches) Kind(branch, defaultBranch string) string {
	if branch == defaultBranch {
		return "default"
	}
	for _, prefix := range b.AutomationPrefixes {
		if strings.HasPrefix(branch, prefix) {
			return "automation"
		}
	}
	category := b.Category(branch)
	for _, allowed := range b.AllowedCategories {
		if category == allowed {
			return "work"
		}
	}
	return "unknown"
}

func (b Branches) Validate(branch, defaultBranch string) error {
	kind := b.Kind(branch, defaultBranch)
	if kind == "default" || kind == "automation" {
		return nil
	}
	if kind != "work" {
		return fmt.Errorf("branch %q does not match allowed work categories", branch)
	}
	pattern, ok := b.Patterns[b.Category(branch)]
	if !ok {
		return fmt.Errorf("branch %q has no pattern for its category", branch)
	}
	matched, err := regexp.MatchString(pattern, branch)
	if err != nil || !matched {
		return fmt.Errorf("branch %q does not match pattern %q", branch, pattern)
	}
	return nil
}

type Commits struct {
	Types           []string `json:"types"`
	SubjectRequired bool     `json:"subject_required"`
	Format          string   `json:"format"`
}

var commitPattern = regexp.MustCompile(`^([a-z]+)(\([^)]*\))?:\s+\S`)

func (c Commits) Validate(subject string) error {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		if c.SubjectRequired {
			return fmt.Errorf("commit subject is required")
		}
		return nil
	}
	match := commitPattern.FindStringSubmatch(subject)
	if match == nil {
		return fmt.Errorf("commit subject %q must use type(scope): summary", subject)
	}
	commitType := match[1]
	for _, allowed := range c.Types {
		if commitType == allowed {
			return nil
		}
	}
	return fmt.Errorf("commit type %q is not allowed", commitType)
}

type Pushes struct {
	DirectMain               string `json:"direct_main"`
	ForceMain                string `json:"force_main"`
	WorkBranchForceWithLease string `json:"work_branch_force_with_lease"`
}

type Issues struct {
	BlankAllowed      bool   `json:"blank_allowed"`
	SecurityReporting string `json:"security_reporting"`
}

type PullRequests struct {
	RequiredForMain        bool   `json:"required_for_main"`
	TitleFormat            string `json:"title_format"`
	AutoMerge              string `json:"auto_merge"`
	DeleteBranchAfterMerge bool   `json:"delete_branch_after_merge"`
}

type Reviews struct {
	Profile                string   `json:"profile"`
	RequiredApprovals      int      `json:"required_approvals"`
	IndependentVerifierFor []string `json:"independent_verifier_for"`
}

func (r Reviews) ApprovalRequirement(risk string) int {
	if risk == "high" || risk == "critical" {
		return maxInt(r.RequiredApprovals, 1)
	}
	return r.RequiredApprovals
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

type Merge struct {
	Strategy               string `json:"strategy"`
	LinearHistory          bool   `json:"linear_history"`
	ConversationResolution bool   `json:"conversation_resolution"`
}

type Agents struct {
	DefaultPermissions  map[string]string `json:"default_permissions"`
	ForbiddenOperations []string          `json:"forbidden_operations"`
}

func (a Agents) IsForbidden(operation string) bool {
	for _, forbidden := range a.ForbiddenOperations {
		if operation == forbidden {
			return true
		}
	}
	return false
}

type Releases struct {
	TagPattern string `json:"tag_pattern"`
	Immutable  bool   `json:"immutable"`
}

func (r Releases) ValidateTag(tag string) error {
	matched, err := regexp.MatchString(r.TagPattern, tag)
	if err != nil || !matched {
		return fmt.Errorf("tag %q does not match release pattern", tag)
	}
	return nil
}

type Tags struct {
	ReleasePatterns []string `json:"release_patterns"`
	Deletion        string   `json:"deletion"`
	Rewrite         string   `json:"rewrite"`
}

func (t Tags) IsRelease(tag string) bool {
	for _, pattern := range t.ReleasePatterns {
		matched, err := filepath.Match(pattern, tag)
		if err == nil && matched {
			return true
		}
	}
	return false
}

type GitHubPolicy struct {
	Repository     string   `json:"repository"`
	Rulesets       bool     `json:"rulesets"`
	RequiredChecks []string `json:"required_checks"`
}

type EmergencyBypass struct {
	Enabled        bool     `json:"enabled"`
	RequiredFields []string `json:"required_fields"`
}

func (e EmergencyBypass) Validate(record map[string]any) error {
	if !e.Enabled {
		return fmt.Errorf("emergency bypass is disabled")
	}
	for _, field := range e.RequiredFields {
		value, ok := record[field]
		if !ok || fmt.Sprint(value) == "" {
			return fmt.Errorf("emergency bypass record is missing %q", field)
		}
	}
	return nil
}

func Load(path string) (Policy, error) {
	var policy Policy
	data, err := os.ReadFile(path)
	if err != nil {
		return policy, err
	}
	if err := json.Unmarshal(data, &policy); err != nil {
		return policy, err
	}
	if err := Validate(policy); err != nil {
		return policy, err
	}
	return policy, nil
}

func Validate(policy Policy) error {
	if policy.SchemaVersion != 1 {
		return fmt.Errorf("unsupported repository policy schema version %d", policy.SchemaVersion)
	}
	if policy.Repository.DefaultBranch == "" {
		return fmt.Errorf("repository.default_branch is required")
	}
	if policy.Repository.Provider != "github" {
		return fmt.Errorf("repository.provider must be github")
	}
	if policy.Repository.GovernanceProfile != "solo" && policy.Repository.GovernanceProfile != "team" && policy.Repository.GovernanceProfile != "strict" {
		return fmt.Errorf("unknown governance profile %q", policy.Repository.GovernanceProfile)
	}
	if len(policy.Branches.AllowedCategories) == 0 {
		return fmt.Errorf("branches.allowed_categories must not be empty")
	}
	if len(policy.Commits.Types) == 0 {
		return fmt.Errorf("commits.types must not be empty")
	}
	if policy.Merge.Strategy != "squash" {
		return fmt.Errorf("merge.strategy must be squash")
	}
	if policy.Tags.Deletion != "deny" || policy.Tags.Rewrite != "deny" {
		return fmt.Errorf("release tags must deny deletion and rewrite")
	}
	if policy.Pushes.ForceMain != "deny" {
		return fmt.Errorf("pushes.force_main must be deny")
	}
	return nil
}
