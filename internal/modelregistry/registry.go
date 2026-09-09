package modelregistry

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

type Privacy string

const (
	Local    Privacy = "local"
	Trusted  Privacy = "trusted"
	External Privacy = "external"
	Unknown  Privacy = "unknown"
)

type HealthStatus string

const (
	HealthHealthy     HealthStatus = "healthy"
	HealthDegraded    HealthStatus = "degraded"
	HealthUnavailable HealthStatus = "unavailable"
)

type ProviderHealth struct {
	Provider     string       `json:"provider"`
	Status       HealthStatus `json:"status"`
	LatencyMS    int          `json:"latency_ms"`
	LastChecked  string       `json:"last_checked"`
	FailureCount int          `json:"failure_count"`
}

type Descriptor struct {
	ID                    string         `json:"id"`
	Version               int            `json:"version"`
	Provider              string         `json:"provider"`
	Privacy               Privacy        `json:"privacy"`
	ContextTokens         int            `json:"context_tokens"`
	StructuredOutput      bool           `json:"structured_output"`
	ToolUse               bool           `json:"tool_use"`
	LatencyClass          string         `json:"latency_class,omitempty"` // "fast", "standard", "deep"
	CostPer1kInputTokens  float64        `json:"cost_per_1k_input,omitempty"`
	CostPer1kOutputTokens float64        `json:"cost_per_1k_output,omitempty"`
	CanaryScore           float64        `json:"canary_score,omitempty"`
	Pricing               map[string]any `json:"pricing,omitempty"`
	RateLimit             map[string]any `json:"rate_limit,omitempty"`
}

type RouteRequest struct {
	Risk                 string `json:"risk,omitempty"`
	DataClass            string `json:"data_class,omitempty"`
	NeedTools            bool   `json:"need_tools,omitempty"`
	NeedStructuredOutput bool   `json:"need_structured_output,omitempty"`
	MinimumContext       int    `json:"minimum_context,omitempty"`
	LatencyPreference    string `json:"latency_preference,omitempty"` // "fast", "standard", "any"
}

type RouteResponse struct {
	Primary     Descriptor   `json:"primary"`
	Fallbacks   []Descriptor `json:"fallbacks"`
	Explanation string       `json:"explanation"`
	Score       float64      `json:"score"`
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

type Registry struct {
	models map[string]Descriptor
	health map[string]ProviderHealth
}

func NewRegistry() *Registry {
	return &Registry{
		models: make(map[string]Descriptor),
		health: make(map[string]ProviderHealth),
	}
}

func (r *Registry) Register(d Descriptor) {
	r.models[d.ID] = d
}

func (r *Registry) Get(id string) (Descriptor, bool) {
	d, ok := r.models[id]
	return d, ok
}

func (r *Registry) List() []Descriptor {
	out := make([]Descriptor, 0, len(r.models))
	for _, d := range r.models {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (r *Registry) SetHealth(provider string, h ProviderHealth) {
	r.health[provider] = h
}

func (r *Registry) GetHealth(provider string) ProviderHealth {
	h, ok := r.health[provider]
	if !ok {
		return ProviderHealth{Provider: provider, Status: HealthHealthy}
	}
	return h
}

func (r *Registry) Route(req RouteRequest) (RouteResponse, error) {
	candidates := make([]Descriptor, 0)
	for _, m := range r.models {
		if !Compatible(m, req) {
			continue
		}
		h := r.GetHealth(m.Provider)
		if h.Status == HealthUnavailable {
			continue
		}
		candidates = append(candidates, m)
	}

	if len(candidates) == 0 {
		return RouteResponse{}, errors.New("no compatible model available for request")
	}

	// Score candidates: healthy +10, low latency +5, local privacy if confidential/internal +5
	type scoredModel struct {
		model Descriptor
		score float64
	}
	scored := make([]scoredModel, 0, len(candidates))
	for _, c := range candidates {
		s := 10.0
		h := r.GetHealth(c.Provider)
		if h.Status == HealthDegraded {
			s -= 5.0
		}
		if req.LatencyPreference == "fast" && c.LatencyClass == "fast" {
			s += 5.0
		}
		if (req.DataClass == "confidential" || req.DataClass == "internal") && c.Privacy == Local {
			s += 5.0
		}
		if c.CanaryScore > 0 {
			s += c.CanaryScore * 2.0
		}
		scored = append(scored, scoredModel{model: c, score: s})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	primary := scored[0].model
	fallbacks := make([]Descriptor, 0, len(scored)-1)
	for i := 1; i < len(scored); i++ {
		fallbacks = append(fallbacks, scored[i].model)
	}

	explanation := fmt.Sprintf("Selected %s (provider=%s, privacy=%s, score=%.1f) from %d candidate(s)",
		primary.ID, primary.Provider, primary.Privacy, scored[0].score, len(candidates))

	return RouteResponse{
		Primary:     primary,
		Fallbacks:   fallbacks,
		Explanation: explanation,
		Score:       scored[0].score,
	}, nil
}

// Canary & drift tracking
type CanaryEval struct {
	ModelID       string  `json:"model_id"`
	BaselineScore float64 `json:"baseline_score"`
	CurrentScore  float64 `json:"current_score"`
	DriftDetected bool    `json:"drift_detected"`
	EvaluatedAt   string  `json:"evaluated_at"`
}

func EvaluateDrift(modelID string, baseline, current, tolerance float64) CanaryEval {
	diff := baseline - current
	drift := diff > tolerance
	return CanaryEval{
		ModelID:       modelID,
		BaselineScore: baseline,
		CurrentScore:  current,
		DriftDetected: drift,
		EvaluatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
}
