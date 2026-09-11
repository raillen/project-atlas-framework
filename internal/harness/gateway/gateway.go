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
	Status        string `json:"status"` // healthy|degraded|open
	Failures      int    `json:"failures"`
	LastFailure   string `json:"last_failure,omitempty"`
	CooldownUntil string `json:"cooldown_until,omitempty"`
}

// RetryPolicy bounds same-provider retries for retryable failures.
// Attempts counts total tries (1 = no retry); Backoff scales linearly
// with jitter omitted for determinism in tests (callers may wrap sleep).
type RetryPolicy struct {
	Attempts int
	Backoff  time.Duration
}

// DefaultRetryPolicy retries twice with a 200ms base backoff.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{Attempts: 3, Backoff: 200 * time.Millisecond}
}

// Gateway routes model requests across registered providers.
type Gateway struct {
	mu        sync.Mutex
	providers map[string]model.Provider
	health    map[string]*Health
	Retry     RetryPolicy
	// AfterSideEffects=false allows transparent fallback; once a Run has
	// observable effects, callers must use explicit Handoff instead.
}

func New() *Gateway {
	return &Gateway{providers: map[string]model.Provider{}, health: map[string]*Health{}, Retry: DefaultRetryPolicy()}
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
		ch, err := g.streamWithRetry(ctx, p, t.Provider, r)
		if err != nil {
			lastErr = err
			continue
		}
		return ch, t.Provider, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no route targets")
	}
	return nil, "", lastErr
}

// streamWithRetry opens the provider stream and peeks the first event,
// retrying retryable failures per policy with linear backoff. Successes
// reset the circuit; terminal provider exhaustion records one failure.
func (g *Gateway) streamWithRetry(ctx context.Context, p model.Provider, name string, req agent.ModelRequest) (<-chan agent.ModelEvent, error) {
	attempts := g.Retry.Attempts
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for a := 0; a < attempts; a++ {
		if a > 0 {
			backoff := time.Duration(a) * g.Retry.Backoff
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
		ch, err := p.Stream(ctx, req)
		if err != nil {
			lastErr = err
			continue
		}
		select {
		case ev, ok := <-ch:
			if !ok {
				lastErr = fmt.Errorf("provider %s closed stream", name)
				continue
			}
			if ev.Kind == agent.EventError && ev.Retryable {
				lastErr = fmt.Errorf("provider %s: %s", name, ev.Error)
				continue
			}
			g.mu.Lock()
			g.recordSuccess(name)
			g.mu.Unlock()
			out := make(chan agent.ModelEvent, 64)
			out <- ev
			go func() {
				defer close(out)
				for e := range ch {
					out <- e
				}
			}()
			return out, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	g.mu.Lock()
	g.recordFailure(name)
	g.mu.Unlock()
	if lastErr == nil {
		lastErr = fmt.Errorf("provider %s exhausted retries", name)
	}
	return nil, lastErr
}
