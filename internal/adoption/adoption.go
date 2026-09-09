package adoption

type Finding struct {
	Path       string `json:"path"`
	Concept    string `json:"concept"`
	Confidence string `json:"confidence"`
	Reason     string `json:"reason"`
}
type Report struct {
	Repository   string    `json:"repository"`
	Capabilities []string  `json:"capabilities"`
	Findings     []Finding `json:"findings"`
	Proposals    []string  `json:"proposals"`
}
