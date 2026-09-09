package connectors

type Contract struct {
	ID            string         `json:"id"`
	Version       string         `json:"version"`
	ProtocolRange string         `json:"protocol_range"`
	Capabilities  []string       `json:"capabilities"`
	Enforcement   string         `json:"enforcement"`
	Hooks         []string       `json:"hooks,omitempty"`
	Install       map[string]any `json:"install,omitempty"`
	Cleanup       map[string]any `json:"cleanup,omitempty"`
}

func (c Contract) Supports(capability string) bool {
	for _, v := range c.Capabilities {
		if v == capability {
			return true
		}
	}
	return false
}
