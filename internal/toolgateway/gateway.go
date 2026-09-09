package toolgateway

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Kind string

const (
	ReadOnly      Kind = "read-only"
	Idempotent    Kind = "idempotent"
	SideEffecting Kind = "side-effecting"
	Destructive   Kind = "destructive"
)

type Descriptor struct {
	ID              string   `json:"id"`
	Version         int      `json:"version"`
	Description     string   `json:"description,omitempty"`
	Kind            Kind     `json:"kind"`
	Trust           string   `json:"trust"`
	Capabilities    []string `json:"capabilities,omitempty"`
	FilesystemScope []string `json:"filesystem_scope,omitempty"`
	NetworkEgress   []string `json:"network_egress,omitempty"`
	CredentialScope []string `json:"credential_scope,omitempty"`
	TimeoutMS       int      `json:"timeout_ms,omitempty"`
	OutputLimit     int      `json:"output_limit,omitempty"`
}

type Summary struct {
	ID           string   `json:"id"`
	Kind         Kind     `json:"kind"`
	Trust        string   `json:"trust"`
	Description  string   `json:"description,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

type Decision struct {
	Allowed    bool       `json:"allowed"`
	Reason     string     `json:"reason"`
	Descriptor Descriptor `json:"descriptor"`
}

func Evaluate(d Descriptor, projectRoot, target string, safe bool) Decision {
	if d.ID == "" {
		return Decision{Reason: "tool descriptor missing id", Descriptor: d}
	}
	if safe && d.Kind == Destructive {
		return Decision{Reason: "safe mode denies destructive tool", Descriptor: d}
	}
	if d.Kind == SideEffecting || d.Kind == Destructive {
		inScope := false
		for _, root := range d.FilesystemScope {
			effectiveRoot := root
			if root == "project-root" {
				effectiveRoot = projectRoot
			}
			if target != "" && (strings.HasPrefix(target, effectiveRoot) || effectiveRoot == ".") {
				inScope = true
				break
			}
		}
		if target != "" && !inScope {
			return Decision{Reason: fmt.Sprintf("tool target outside allowed scope: %s", target), Descriptor: d}
		}
	}
	return Decision{Allowed: true, Reason: "tool permitted by descriptor and policy", Descriptor: d}
}

type MCPServerDescriptor struct {
	ID            string   `json:"id"`
	Transport     string   `json:"transport"`
	Command       string   `json:"command,omitempty"`
	Args          []string `json:"args,omitempty"`
	URL           string   `json:"url,omitempty"`
	Trust         string   `json:"trust"` // "trusted" or "untrusted" (default)
	AllowedTools  []string `json:"allowed_tools,omitempty"`
	RootScopes    []string `json:"root_scopes,omitempty"`
	NetworkPolicy string   `json:"network_policy,omitempty"` // "deny", "restricted", "allow"
	SecretScopes  []string `json:"secret_scopes,omitempty"`
}

func EvaluateMCP(server MCPServerDescriptor, tool Descriptor, target string, safe bool) Decision {
	if server.ID == "" {
		return Decision{Reason: "mcp server descriptor missing id", Descriptor: tool}
	}
	// Untrusted by default
	trust := server.Trust
	if trust == "" {
		trust = "untrusted"
	}
	if trust == "untrusted" && tool.Kind == Destructive {
		return Decision{Reason: "untrusted mcp server cannot execute destructive tool", Descriptor: tool}
	}
	if len(server.AllowedTools) > 0 {
		allowed := false
		for _, name := range server.AllowedTools {
			if name == tool.ID || name == "*" {
				allowed = true
				break
			}
		}
		if !allowed {
			return Decision{Reason: fmt.Sprintf("tool %s is not in allowed list for mcp server %s", tool.ID, server.ID), Descriptor: tool}
		}
	}
	if len(server.RootScopes) > 0 && target != "" {
		inScope := false
		for _, root := range server.RootScopes {
			if strings.HasPrefix(target, root) || root == "." || root == "*" {
				inScope = true
				break
			}
		}
		if !inScope {
			return Decision{Reason: fmt.Sprintf("target %s outside mcp server root scopes", target), Descriptor: tool}
		}
	}
	return Evaluate(tool, ".", target, safe)
}

// Registry provides lazy discovery and tool lookup.
type Registry struct {
	tools map[string]Descriptor
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Descriptor)}
}

func (r *Registry) Register(d Descriptor) {
	r.tools[d.ID] = d
}

func (r *Registry) Get(id string) (Descriptor, bool) {
	d, ok := r.tools[id]
	return d, ok
}

func (r *Registry) List() []Descriptor {
	out := make([]Descriptor, 0, len(r.tools))
	for _, d := range r.tools {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Discover returns lightweight summaries filtered by capabilities (lazy discovery).
func (r *Registry) Discover(needCapabilities ...string) []Summary {
	out := make([]Summary, 0)
	for _, d := range r.tools {
		if len(needCapabilities) > 0 {
			matched := false
			for _, need := range needCapabilities {
				for _, cap := range d.Capabilities {
					if cap == need {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				continue
			}
		}
		out = append(out, Summary{
			ID:           d.ID,
			Kind:         d.Kind,
			Trust:        d.Trust,
			Description:  d.Description,
			Capabilities: d.Capabilities,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Output budgeting
func EnforceOutputBudget(output string, maxBytes int) (string, bool, string) {
	if maxBytes <= 0 || len(output) <= maxBytes {
		return output, false, ""
	}
	sum := sha256.Sum256([]byte(output))
	ptr := "sha256:" + hex.EncodeToString(sum[:])
	truncated := output[:maxBytes] + fmt.Sprintf("\n... [output truncated: %d bytes total, full content hash %s]", len(output), ptr)
	return truncated, true, ptr
}

// Side-effect enforcement
type Intent struct {
	ToolID         string `json:"tool_id"`
	IdempotencyKey string `json:"idempotency_key"`
	Target         string `json:"target"`
	PayloadHash    string `json:"payload_hash,omitempty"`
	Timestamp      string `json:"timestamp"`
}

type Outcome struct {
	ToolID         string `json:"tool_id"`
	IdempotencyKey string `json:"idempotency_key"`
	Success        bool   `json:"success"`
	OutputHash     string `json:"output_hash,omitempty"`
	Error          string `json:"error,omitempty"`
	Timestamp      string `json:"timestamp"`
}

func ValidateIntent(intent Intent) error {
	if intent.ToolID == "" {
		return errors.New("tool_id is required for side-effect intent")
	}
	if intent.IdempotencyKey == "" {
		return errors.New("idempotency_key is required for side-effect intent")
	}
	return nil
}

func RecordOutcome(intent Intent, success bool, output []byte, err error) Outcome {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	outHash := ""
	if len(output) > 0 {
		sum := sha256.Sum256(output)
		outHash = "sha256:" + hex.EncodeToString(sum[:])
	}
	return Outcome{
		ToolID:         intent.ToolID,
		IdempotencyKey: intent.IdempotencyKey,
		Success:        success,
		OutputHash:     outHash,
		Error:          errMsg,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	}
}
