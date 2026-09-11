package model

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	if got := EstimateTokens("", "x"); got != 1 {
		t.Fatalf("empty must be 1, got %d", got)
	}
	if got := EstimateTokens("abcdefgh", "gpt-4"); got != 2 {
		t.Fatalf("8 chars @4.0 must be 2, got %d", got)
	}
	if got := EstimateTokens("abcdefgh", "claude-3"); got != 2 {
		t.Fatalf("8 chars @3.5 must be 2, got %d", got)
	}
	if EstimateTokens("x", "unknown-model-zzz") != 1 {
		t.Fatal("unknown model must use default ratio")
	}
	if CappedEstimate("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "gpt", 3) != 3 {
		t.Fatal("cap must bind")
	}
	if EstimatorVersion == "" {
		t.Fatal("estimator version pinned")
	}
}

func TestOpenAIDiscovery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/models" {
			_, _ = w.Write([]byte(`{"data":[{"id":"m1"},{"id":"m2"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	o := NewOpenAICompat(srv.URL, "", "fallback")
	models, err := o.Models(context.Background())
	if err != nil || len(models) != 2 || models[0] != "m1" {
		t.Fatalf("discovery failed: %v %v", models, err)
	}
	down := NewOpenAICompat("http://127.0.0.1:9", "", "fallback")
	models, err = down.Models(context.Background())
	if err != nil || len(models) != 1 || models[0] != "fallback" {
		t.Fatalf("unreachable must fall back: %v %v", models, err)
	}
}
