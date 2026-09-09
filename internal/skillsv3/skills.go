package skillsv3

type Manifest struct {
	ID            string   `json:"id"`
	Version       string   `json:"version"`
	SchemaVersion int      `json:"schema_version"`
	Description   string   `json:"description"`
	Maturity      string   `json:"maturity"`
	Activation    string   `json:"activation"`
	Capabilities  []string `json:"capabilities"`
	Requires      []string `json:"requires"`
	Conflicts     []string `json:"conflicts"`
	Permissions   []string `json:"permissions"`
	ContextBudget int      `json:"context_budget"`
	Outputs       []string `json:"outputs"`
}
type Evaluation struct {
	SkillID  string             `json:"skill_id"`
	Baseline string             `json:"baseline"`
	Result   string             `json:"result"`
	Metrics  map[string]float64 `json:"metrics"`
}

func (m Manifest) CanActivate(mode string) bool {
	return m.Activation == mode || m.Activation == "auto" || m.Activation == "contextual"
}
