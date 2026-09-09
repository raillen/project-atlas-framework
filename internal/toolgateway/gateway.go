package toolgateway

import "fmt"

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
	Kind            Kind     `json:"kind"`
	Trust           string   `json:"trust"`
	FilesystemScope []string `json:"filesystem_scope,omitempty"`
	NetworkEgress   []string `json:"network_egress,omitempty"`
	CredentialScope []string `json:"credential_scope,omitempty"`
	TimeoutMS       int      `json:"timeout_ms,omitempty"`
	OutputLimit     int      `json:"output_limit,omitempty"`
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
		allowed := false
		for _, root := range d.FilesystemScope {
			if root == projectRoot || root == "project-root" {
				allowed = true
			}
		}
		if target != "" && !allowed {
			return Decision{Reason: fmt.Sprintf("tool target outside allowed scope: %s", target), Descriptor: d}
		}
	}
	return Decision{Allowed: true, Reason: "tool permitted by descriptor and policy", Descriptor: d}
}
