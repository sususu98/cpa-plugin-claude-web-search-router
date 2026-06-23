package main

import "testing"

func TestConfigurePreservesDefaultBooleansWhenConfigIsPartial(t *testing.T) {
	raw := mustJSON(t, lifecycleRequest{ConfigYAML: []byte("route: codex_web_search\n")})

	if errConfigure := configure(raw); errConfigure != nil {
		t.Fatalf("configure() error = %v", errConfigure)
	}

	cfg := loadedConfig()
	if !cfg.Enabled {
		t.Fatal("Enabled = false, want default true")
	}
	if !cfg.RequireWebSearchOnly {
		t.Fatal("RequireWebSearchOnly = false, want default true")
	}
	if cfg.Route.SingleRouteString() != string(backendCodexWebSearch) {
		t.Fatalf("Route = %#v, want codex_web_search", cfg.Route)
	}
}

func TestConfigureRouteYAMLListOrder(t *testing.T) {
	yaml := `route:
  - tavily
  - codex_web_search
  - xai_web_search
`
	raw := mustJSON(t, lifecycleRequest{ConfigYAML: []byte(yaml)})
	if errConfigure := configure(raw); errConfigure != nil {
		t.Fatalf("configure() error = %v", errConfigure)
	}
	cfg := loadedConfig()
	got := webSearchFallbackChain(cfg)
	want := []routeBackend{backendTavily, backendCodexWebSearch, backendXAIWebSearch}
	if len(got) != len(want) {
		t.Fatalf("chain len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
