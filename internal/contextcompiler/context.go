package contextcompiler

import (
	"sort"
)

type Source struct {
	Ref       string `json:"ref"`
	Authority string `json:"authority"`
	Freshness string `json:"freshness"`
	TokenCost int    `json:"token_cost"`
	Included  bool   `json:"included"`
	Reason    string `json:"reason"`
}
type Manifest struct {
	Version         int      `json:"version"`
	RunID           string   `json:"run_id"`
	Sources         []Source `json:"sources"`
	EstimatedTokens int      `json:"estimated_tokens"`
	Pressure        string   `json:"pressure"`
}

func Compile(runID string, sources []Source, budget int) Manifest {
	sort.Slice(sources, func(i, j int) bool {
		if sources[i].Authority != sources[j].Authority {
			return sources[i].Authority < sources[j].Authority
		}
		if sources[i].TokenCost != sources[j].TokenCost {
			return sources[i].TokenCost < sources[j].TokenCost
		}
		return sources[i].Ref < sources[j].Ref
	})
	seen := map[string]bool{}
	used := 0
	for i := range sources {
		if seen[sources[i].Ref] || used+sources[i].TokenCost > budget {
			sources[i].Included = false
			continue
		}
		seen[sources[i].Ref] = true
		sources[i].Included = true
		sources[i].Reason = "deterministic priority pack"
		used += sources[i].TokenCost
	}
	pressure := "healthy"
	if used > budget*80/100 {
		pressure = "pressure"
	}
	if used >= budget {
		pressure = "critical"
	}
	return Manifest{Version: 1, RunID: runID, Sources: sources, EstimatedTokens: used, Pressure: pressure}
}
