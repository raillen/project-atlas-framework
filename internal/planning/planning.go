package planning

type OpenQuestion struct {
	ID       string   `json:"id"`
	Scope    string   `json:"scope"`
	Contract string   `json:"contract,omitempty"`
	Priority string   `json:"priority"`
	Status   string   `json:"status"`
	Question string   `json:"question"`
	Answer   string   `json:"answer,omitempty"`
	Evidence []string `json:"evidence,omitempty"`
}
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
