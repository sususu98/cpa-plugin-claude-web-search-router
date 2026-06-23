package main

import "testing"

func TestWebSearchFallbackChainUsesRouteListOrder(t *testing.T) {
	cfg := pluginConfig{
		Route: routeMode{kind: routeModeOrdered, ordered: []string{"tavily", "codex_web_search", "xai_web_search"}},
	}
	got := webSearchFallbackChain(cfg)
	want := []routeBackend{backendTavily, backendCodexWebSearch, backendXAIWebSearch}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("[%d] = %q, want %q (full %v)", i, got[i], want[i], got)
		}
	}
}

func TestWebSearchFallbackChainDefaultWhenBuiltinFallback(t *testing.T) {
	cfg := pluginConfig{Route: routeBuiltinDefault()}
	got := webSearchFallbackChain(cfg)
	want := defaultWebSearchFallbackChain()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestNormalizeFallbackChainDedupes(t *testing.T) {
	got := normalizeFallbackChainEntries([]string{"tavily", "TAVILY", "codex", "unknown"})
	if len(got) != 2 || got[0] != backendTavily || got[1] != backendCodexWebSearch {
		t.Fatalf("got %v", got)
	}
}
