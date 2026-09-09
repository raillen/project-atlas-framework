package connectors

import (
	"fmt"
	"sort"
)

// NegotiationRequest specifies capabilities requested by an agent or workflow.
type NegotiationRequest struct {
	TargetHarness        string   `json:"target_harness"`
	RequiredCapabilities []string `json:"required_capabilities"`
	OptionalCapabilities []string `json:"optional_capabilities,omitempty"`
	Strict               bool     `json:"strict,omitempty"`
}

// NegotiationResult reports compatibility, native support, and graceful degradations.
type NegotiationResult struct {
	ConnectorID string            `json:"connector_id"`
	Compatible  bool              `json:"compatible"`
	Supported   []string          `json:"supported"`
	Degraded    map[string]string `json:"degraded"`
	Unsupported []string          `json:"unsupported"`
	Warnings    []string          `json:"warnings,omitempty"`
}

// Graceful fallback mappings when a host harness does not natively support a capability.
var fallbackModes = map[string]string{
	CapPreToolBlock:     "degraded to post-execution advisory review (advise mode)",
	CapSessionHooks:     "degraded to periodic checkpoint and manual handoff synthesis",
	CapPostToolVerify:   "degraded to offline verification via doctor/validation quality gates",
	CapIsolateSubagents: "degraded to single-context execution with explicit role prompting",
	CapRestrictTools:    "degraded to system-prompt constraints and audit trail verification",
	CapNativePlugin:     "degraded to external CLI invocation",
}

// Negotiate checks contract capabilities against requested capabilities, computing supported, degraded, and unsupported sets.
func Negotiate(contract Contract, req NegotiationRequest) (*NegotiationResult, error) {
	result := &NegotiationResult{
		ConnectorID: contract.ID,
		Compatible:  true,
		Supported:   []string{},
		Degraded:    map[string]string{},
		Unsupported: []string{},
		Warnings:    []string{},
	}

	capMap := map[string]bool{}
	for _, c := range contract.Capabilities {
		capMap[c] = true
	}

	// Evaluate required capabilities
	for _, reqCap := range req.RequiredCapabilities {
		if capMap[reqCap] {
			result.Supported = append(result.Supported, reqCap)
		} else if fallback, hasFallback := fallbackModes[reqCap]; hasFallback && !req.Strict {
			result.Degraded[reqCap] = fallback
			result.Warnings = append(result.Warnings, fmt.Sprintf("Capability %q not natively supported by %s; %s", reqCap, contract.ID, fallback))
		} else {
			result.Unsupported = append(result.Unsupported, reqCap)
			result.Compatible = false
			result.Warnings = append(result.Warnings, fmt.Sprintf("Required capability %q is unsupported by %s", reqCap, contract.ID))
		}
	}

	// Evaluate optional capabilities
	for _, optCap := range req.OptionalCapabilities {
		if capMap[optCap] {
			result.Supported = append(result.Supported, optCap)
		} else if fallback, hasFallback := fallbackModes[optCap]; hasFallback {
			result.Degraded[optCap] = fallback
		} else {
			result.Unsupported = append(result.Unsupported, optCap)
		}
	}

	sort.Strings(result.Supported)
	sort.Strings(result.Unsupported)
	sort.Strings(result.Warnings)

	if req.Strict && !result.Compatible {
		return result, fmt.Errorf("strict negotiation failed for %s: unsupported capabilities: %v", contract.ID, result.Unsupported)
	}

	return result, nil
}
