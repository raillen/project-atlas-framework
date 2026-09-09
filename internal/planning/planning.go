package planning

type DecisionProposal struct {
	ID             string   `json:"id"`
	Statement      string   `json:"statement"`
	Classification string   `json:"classification"`
	Authority      string   `json:"authority"`
	Confidence     string   `json:"confidence"`
	Status         string   `json:"status"`
	Scope          string   `json:"scope,omitempty"`
	Evidence       []string `json:"evidence,omitempty"`
}

func AgentSuggestion(statement, scope string) DecisionProposal {
	return DecisionProposal{Statement: statement, Classification: "agent-suggestion", Authority: "agent-proposal", Confidence: "unknown", Status: "proposed", Scope: scope}
}
