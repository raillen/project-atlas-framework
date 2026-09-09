package docengine

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Impact struct {
	ContractID string   `json:"contract_id"`
	Documents  []string `json:"documents"`
	Reason     string   `json:"reason"`
	Severity   string   `json:"severity"`
}
type Delta struct {
	State     string   `json:"state"`
	Source    string   `json:"source"`
	Contracts []string `json:"contracts"`
	Documents []string `json:"documents"`
	Reason    string   `json:"reason"`
	Evidence  []string `json:"evidence"`
}
type Finding struct {
	ID          string   `json:"id"`
	Sources     []string `json:"sources"`
	Contract    string   `json:"contract,omitempty"`
	Type        string   `json:"type"`
	Severity    string   `json:"severity"`
	Authority   string   `json:"authority"`
	Description string   `json:"description"`
	State       string   `json:"state"`
}

func AnalyzeImpact(root string, changed []string) ([]Impact, error) {
	registry, err := LoadRegistry(root)
	if err != nil {
		return nil, err
	}
	bindings, err := LoadBindings(root)
	if err != nil {
		return nil, err
	}
	impacts := []Impact{}
	for _, contract := range registry.Contracts {
		for _, trigger := range contract.UpdateTriggers {
			if triggerMatches(trigger, changed) {
				docs := []string{}
				for _, b := range bindings {
					if b.ContractID == contract.ID {
						docs = append(docs, b.Sources...)
					}
				}
				sort.Strings(docs)
				impacts = append(impacts, Impact{ContractID: contract.ID, Documents: docs, Reason: "update trigger: " + trigger, Severity: "medium"})
				break
			}
		}
	}
	sort.Slice(impacts, func(i, j int) bool { return impacts[i].ContractID < impacts[j].ContractID })
	return impacts, nil
}
func triggerMatches(trigger string, changed []string) bool {
	token := strings.ToLower(strings.Split(trigger, ".")[0])
	for _, path := range changed {
		if strings.Contains(strings.ToLower(path), token) {
			return true
		}
	}
	return false
}
func MakeDelta(goal string, impacts []Impact) Delta {
	contracts := []string{}
	docs := []string{}
	for _, impact := range impacts {
		contracts = append(contracts, impact.ContractID)
		docs = append(docs, impact.Documents...)
	}
	sort.Strings(contracts)
	sort.Strings(docs)
	return Delta{State: "proposed", Source: goal, Contracts: unique(contracts), Documents: unique(docs), Reason: "documentation impact analysis", Evidence: []string{}}
}
func unique(values []string) []string {
	out := []string{}
	for _, v := range values {
		if len(out) == 0 || out[len(out)-1] != v {
			out = append(out, v)
		}
	}
	return out
}
func DetectContradictions(root string) ([]Finding, error) {
	values := map[string][]string{}
	pattern := regexp.MustCompile(`(?i)(?:port|default_port)\s*(?:=|:)\s*([0-9]{2,5})`)
	for _, source := range []string{"atlas.json", "docs/manual/installation.md", "docs/reference/cli.md", "docs/architecture/overview.md"} {
		data, err := os.ReadFile(filepath.Join(root, source))
		if err != nil {
			continue
		}
		for _, match := range pattern.FindAllStringSubmatch(string(data), -1) {
			if len(match) > 1 {
				values[match[1]] = append(values[match[1]], source)
			}
		}
	}
	if len(values) < 2 {
		return []Finding{}, nil
	}
	paths := []string{}
	ports := []string{}
	for port, sources := range values {
		ports = append(ports, port)
		paths = append(paths, sources...)
	}
	sort.Strings(ports)
	sort.Strings(paths)
	return []Finding{{ID: "DOC_CONTRADICTION", Sources: unique(paths), Type: "syntactic contradiction", Severity: "medium", Authority: "canonical-documentation", Description: "conflicting port values: " + strings.Join(ports, ", "), State: "open"}}, nil
}
func DetectStaleness(root string, changed []string) ([]Finding, error) {
	impacts, err := AnalyzeImpact(root, changed)
	if err != nil {
		return nil, err
	}
	out := []Finding{}
	for _, impact := range impacts {
		out = append(out, Finding{ID: "DOC_SOURCE_STALE", Sources: impact.Documents, Contract: impact.ContractID, Type: "trigger-based", Severity: impact.Severity, Authority: "canonical-documentation", Description: impact.Reason, State: "open"})
	}
	return out, nil
}
func (d Delta) Validate() error {
	valid := map[string]bool{"proposed": true, "reviewed": true, "accepted": true, "rejected": true, "applied": true}
	if !valid[d.State] || d.Source == "" {
		return fmt.Errorf("documentation delta requires valid state and source")
	}
	if d.State == "applied" && len(d.Evidence) == 0 {
		return fmt.Errorf("applied documentation delta requires evidence")
	}
	return nil
}
