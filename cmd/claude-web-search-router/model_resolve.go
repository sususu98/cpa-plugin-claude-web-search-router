package main

import (
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy"
)

const (
	// Default Codex model for Claude web_search → Codex Responses (override with codex_model).
	defaultCodexWebSearchModel = "gpt-5.4-mini"
	// Default xAI model for server-side web_search per https://docs.x.ai/developers/tools/web-search
	defaultXAIWebSearchModel = "grok-4.3"
)

// resolveAntigravityWebSearchTargetModel picks an Antigravity model that can run native googleSearch.
// Config antigravity_model wins; otherwise antigravityWebSearchModelFor(requested) or the
// first available antigravity model with SupportsWebSearch.
func resolveAntigravityWebSearchTargetModel(configured, requested string) string {
	if m := strings.TrimSpace(configured); m != "" {
		return m
	}
	if m := antigravityWebSearchModelFor(strings.TrimSpace(requested)); m != "" {
		return m
	}
	for _, model := range cliproxy.GlobalModelRegistry().GetAvailableModelsByProvider("antigravity") {
		if model == nil || !model.SupportsWebSearch {
			continue
		}
		if id := strings.TrimSpace(model.ID); id != "" {
			return id
		}
	}
	return ""
}

// resolveCodexWebSearchTargetModel never forwards the client Claude model to Codex.
func resolveCodexWebSearchTargetModel(configured string) string {
	if m := strings.TrimSpace(configured); m != "" {
		return m
	}
	return defaultCodexWebSearchModel
}

// resolveXAIWebSearchTargetModel never forwards the client Claude model to xAI Responses.
func resolveXAIWebSearchTargetModel(configured string) string {
	if m := strings.TrimSpace(configured); m != "" {
		return m
	}
	return defaultXAIWebSearchModel
}

func antigravityWebSearchModelFor(modelID string) string {
	modelID = normalizeAntigravityCapabilityModelID(modelID)
	if modelID == "" {
		return ""
	}
	for _, model := range cliproxy.GlobalModelRegistry().GetAvailableModelsByProvider("antigravity") {
		if model == nil {
			continue
		}
		currentModelID := normalizeAntigravityCapabilityModelID(model.ID)
		if currentModelID == "" {
			continue
		}
		if currentModelID == modelID {
			if model.SupportsWebSearch {
				return currentModelID
			}
			return ""
		}
	}
	return ""
}

func normalizeAntigravityCapabilityModelID(modelID string) string {
	modelID = strings.ToLower(strings.TrimSpace(modelID))
	if open := strings.LastIndex(modelID, "("); open >= 0 && strings.HasSuffix(modelID, ")") {
		modelID = strings.TrimSpace(modelID[:open])
	}
	return modelID
}

