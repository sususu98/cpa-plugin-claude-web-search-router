package main

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type routeModeKind int

const (
	routeModeBuiltinDefault routeModeKind = iota
	routeModeSingle
	routeModeOrdered
)

// routeMode is plugin `route`: scalar backend, `fallback`/empty for default chain, or YAML list for ordered fallback.
type routeMode struct {
	kind    routeModeKind
	single  string
	ordered []string
}

func routeBuiltinDefault() routeMode {
	return routeMode{kind: routeModeBuiltinDefault}
}

func (m *routeMode) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		*m = routeBuiltinDefault()
		return nil
	}
	switch value.Kind {
	case yaml.ScalarNode:
		s := strings.TrimSpace(value.Value)
		if s == "" || strings.EqualFold(s, string(backendFallback)) {
			*m = routeBuiltinDefault()
			return nil
		}
		*m = routeMode{kind: routeModeSingle, single: s}
		return nil
	case yaml.SequenceNode:
		var entries []string
		if err := value.Decode(&entries); err != nil {
			return err
		}
		*m = routeMode{kind: routeModeOrdered, ordered: normalizeRouteListEntries(entries)}
		return nil
	default:
		return fmt.Errorf("route: expected string or sequence, got %s", value.ShortTag())
	}
}

func (m routeMode) MarshalYAML() (any, error) {
	switch m.kind {
	case routeModeBuiltinDefault:
		return string(backendFallback), nil
	case routeModeSingle:
		if strings.TrimSpace(m.single) == "" {
			return string(backendFallback), nil
		}
		return m.single, nil
	case routeModeOrdered:
		if len(m.ordered) == 0 {
			return string(backendFallback), nil
		}
		return m.ordered, nil
	default:
		return string(backendFallback), nil
	}
}

func (m routeMode) OrchestratedFallback() bool {
	return m.kind == routeModeBuiltinDefault || m.kind == routeModeOrdered
}

func (m routeMode) SingleRouteString() string {
	if m.kind != routeModeSingle {
		return ""
	}
	return strings.TrimSpace(m.single)
}

func (m routeMode) userOrderedBackends() []routeBackend {
	if m.kind != routeModeOrdered || len(m.ordered) == 0 {
		return nil
	}
	return normalizeFallbackChainEntries(m.ordered)
}

func normalizeRouteListEntries(entries []string) []string {
	if len(entries) == 0 {
		return nil
	}
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		if trimmed := strings.TrimSpace(entry); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
