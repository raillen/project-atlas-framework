package modelregistry

type Privacy string

const (
	Local    Privacy = "local"
	Trusted  Privacy = "trusted"
	External Privacy = "external"
	Unknown  Privacy = "unknown"
)

type Descriptor struct {
	ID               string         `json:"id"`
	Version          int            `json:"version"`
	Provider         string         `json:"provider"`
	Privacy          Privacy        `json:"privacy"`
	ContextTokens    int            `json:"context_tokens"`
	StructuredOutput bool           `json:"structured_output"`
	ToolUse          bool           `json:"tool_use"`
	Pricing          map[string]any `json:"pricing,omitempty"`
	RateLimit        map[string]any `json:"rate_limit,omitempty"`
}
type RouteRequest struct {
	Risk                 string
	DataClass            string
	NeedTools            bool
	NeedStructuredOutput bool
	MinimumContext       int
}

func Compatible(model Descriptor, request RouteRequest) bool {
	if request.DataClass == "restricted" && model.Privacy != Local {
		return false
	}
	if request.NeedTools && !model.ToolUse {
		return false
	}
	if request.NeedStructuredOutput && !model.StructuredOutput {
		return false
	}
	return model.ContextTokens >= request.MinimumContext
}
