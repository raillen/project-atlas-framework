// Package gateway implements the internal Model Gateway: routing, fallback,
// retry and circuit breaker over ModelProviders. It never routes external
// agent runtimes through the model path.
package gateway

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/raillen/prumo/internal/harness/agent"
	"github.com/raillen/prumo/internal/harness/model"
)

// RouteTarget is one selectable provider+model.
type RouteTarget struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Weight   int    `json:"weight,omitempty"`
}

// ModelRoute is the chosen route with fallbacks.
type ModelRoute struct {
	Primary   RouteTarget   `json:"primary"`
	Fallbacks []RouteTarget `json:"fallbacks,omitempty"`
	Reason    string        `json:"reason,omitempty"`
}

// Health tracks per-provider circuit state.
type Health struct {
	Status       string `json:"status"` // healthy|degraded|open
	Failures     int    `json:"failures"`
	LastFailure  string `json:"last_failure,omitempty"`
	CooldownUntil string `json:"cooldown_until,omitempty"`
}

// Gateway routes model requests across registered providers.
type Gateway struct {
	mu        sync.Mutex
	providers map[string]model.Provider
	health    map[string]*Health
	// AfterSideEffects=false allows transparent fallback; once a Run has
	// observable effects, callers must use explicit Handoff instead.
}

func New() *Gateway {
	return &Gateway{providers: map[string]model.Provider{}, health: map[string]*Health{}}
}

func (g *Gateway) Register(p model.Provider) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.providers[p.Name()] = p
	g.health[p.Name()] = &Health{Status: "healthy"}
}

// Select picks primary+fallbacks: healthy first, deterministic order.
func (g *Gateway) Select(targets []RouteTarget) ModelRoute {
	g.mu.Lock()
	defer g.mu.Unlock()
	healthy, degraded := []RouteTarget{}, []RouteTarget{}
	for _, t := range targets {
		h := g.health[t.Provider]
		if h == nil || h.Status == "healthy" {
			healthy = append(healthy, t)
		} else if h.Status == "degraded" {
			degraded = append(degraded, t)
		}
	}
	sort.Slice(healthy, func(i, j int) bool { return healthy[i].Provider < healthy[j].Provider })
	ordered := append(healthy, degraded...)
	if len(ordered) == 0 {
		return ModelRoute{}
	}
	return ModelRoute{Primary: ordered[0], Fallbacks: ordered[1:], Reason: "healthy-first deterministic"}
}

func (g *Gateway) recordFailure(provider string) {
	h := g.health[provider]
	if h == nil {
		return
	}
	h.Failures++
	h.LastFailure = time.Now().UTC().Format(time.RFC3339Nano)
	if h.Failures >= 3 {
		h.Status = "open"
		h.CooldownUntil = time.Now().UTC().Add(30 * time.Second).Format(time.RFC3339Nano)
	} else if h.Failures >= 1 {
		h.Status = "degraded"
	}
}

func (g *Gateway) recordSuccess(provider string) {
	h := g.health[provider]
	if h == nil {
		return
	}
	h.Failures = 0
	h.Status = "healthy"
}

// StreamWithFallback tries primary then fallbacks on retryable errors.
// afterSideEffects=true disables transparent fallback (caller must Handoff).
func (g *Gateway) StreamWithFallback(ctx context.Context, route ModelRoute, req agent.ModelRequest, afterSideEffects bool) (<-chan agent.ModelEvent, string, error) {
	chain := append([]RouteTarget{route.Primary}, route.Fallbacks...)
	var lastErr error
	for i, t := range chain {
		if i > 0 && afterSideEffects {
			return nil, "", fmt.Errorf("transparent fallback denied after side effects: use explicit Handoff (would try %s)", t.Provider)
		}
		g.mu.Lock()
		p := g.providers[t.Provider]
		g.mu.Unlock()
		if p == nil {
			lastErr = fmt.Errorf("unknown provider %s", t.Provider)
			continue
		}
		r := req
		if t.Model != "" {
			r.Model = t.Model
		}
		ch, err := p.Stream(ctx, r)
		if err != nil {
			g.mu.Lock()
			g.recordFailure(t.Provider)
			g.mu.Unlock()
			lastErr = err
			continue
		}
		// Peek first event for immediate retryable error.
		select {
		case ev, ok := <-ch:
			if !ok {
				return nil, "", fmt.Errorf("provider %s closed stream", t.Provider)
			}
			if ev.Kind == agent.EventError && ev.Retryable {
				g.mu.Lock()
				g.recordFailure(t.Provider)
				g.mu.Unlock()
				lastErr = fmt.Errorf("provider %s: %s", t.Provider, ev.Error)
				continue
			}
			g.mu.Lock()
			g.recordSuccess(t.Provider)
			g.mu.Unlock()
			out := make(chan agent.ModelEvent, 64)
			out <- ev
			go func() {
				defer close(out)
				for e := range ch {
					out <- e
				}
			}()
			return out, t.Provider, nil
		case <-ctx.Done():
			return nil, "", ctx.Err()
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no route targets")
	}
	return nil, "", lastErr
}
