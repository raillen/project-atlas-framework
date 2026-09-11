package gateway

import (
	"context"
	"testing"

	"github.com/raillen/prumo/internal/harness/agent"
	"github.com/raillen/prumo/internal/harness/model"
)

func TestFallbackBeforeSideEffects(t *testing.T) {
	g := New()
	bad := model.NewFake(map[string][]model.ScriptStep{"*": []model.ScriptStep{{Kind: "error", Error: "boom", Retryable: true}}})
	good := model.NewFake(map[string][]model.ScriptStep{"*": []model.ScriptStep{{Kind: "text", Text: "ok"}, {Kind: "complete"}}})
	// Wrap names via custom providers
	g.Register(named{"bad", bad})
	g.Register(named{"good", good})
	route := g.Select([]RouteTarget{{Provider: "bad"}, {Provider: "good"}})
	ch, used, err := g.StreamWithFallback(context.Background(), route, agent.ModelRequest{RequestID: "r", TurnID: "t"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if used != "good" {
		t.Fatalf("expected fallback to good, got %s", used)
	}
	n := 0
	for range ch {
		n++
	}
	if n == 0 {
		t.Fatal("expected events")
	}
}

func TestNoTransparentFallbackAfterSideEffects(t *testing.T) {
	g := New()
	bad := model.NewFake(map[string][]model.ScriptStep{"*": []model.ScriptStep{{Kind: "error", Error: "boom", Retryable: true}}})
	good := model.NewFake(map[string][]model.ScriptStep{"*": []model.ScriptStep{{Kind: "complete"}}})
	g.Register(named{"bad", bad})
	g.Register(named{"good", good})
	route := g.Select([]RouteTarget{{Provider: "bad"}, {Provider: "good"}})
	_, _, err := g.StreamWithFallback(context.Background(), route, agent.ModelRequest{RequestID: "r"}, true)
	if err == nil {
		t.Fatal("expected Handoff-required error after side effects")
	}
}

type named struct {
	name string
	p    *model.FakeProvider
}

func (n named) Name() string                        { return n.name }
func (n named) Capabilities() model.Capabilities    { return n.p.Capabilities() }
func (n named) Models(ctx context.Context) ([]string, error) { return n.p.Models(ctx) }
func (n named) Health(ctx context.Context) (string, error)   { return n.p.Health(ctx) }
func (n named) Stream(ctx context.Context, r agent.ModelRequest) (<-chan agent.ModelEvent, error) {
	return n.p.Stream(ctx, r)
}
