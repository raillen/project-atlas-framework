package egress

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Class string

const (
	Public       Class = "public"
	Internal     Class = "internal"
	Confidential Class = "confidential"
	Restricted   Class = "restricted"
)

type Decision struct {
	Allowed     bool   `json:"allowed"`
	Reason      string `json:"reason"`
	DataClass   Class  `json:"data_class"`
	Destination string `json:"destination"`
}

func Allow(data Class, destination string, local bool) Decision {
	if data == Restricted && !local {
		return Decision{Allowed: false, Reason: "restricted data requires local destination", DataClass: data, Destination: destination}
	}
	return Decision{Allowed: true, Reason: "egress permitted by data policy", DataClass: data, Destination: destination}
}

func AllowEgress(data Class, destination string, secureTLS bool, trustedHost bool) Decision {
	switch data {
	case Restricted:
		return Decision{Allowed: false, Reason: "restricted data cannot egress externally", DataClass: data, Destination: destination}
	case Confidential:
		if !secureTLS {
			return Decision{Allowed: false, Reason: "confidential data requires secure TLS connection", DataClass: data, Destination: destination}
		}
		if !trustedHost {
			return Decision{Allowed: false, Reason: "confidential data requires trusted destination host", DataClass: data, Destination: destination}
		}
	case Internal:
		if !secureTLS && !strings.HasPrefix(destination, "localhost") && !strings.HasPrefix(destination, "127.0.0.1") {
			return Decision{Allowed: false, Reason: "internal data requires secure TLS connection for external hosts", DataClass: data, Destination: destination}
		}
	case Public:
		// Public data is always permitted
	}
	return Decision{Allowed: true, Reason: "egress permitted by policy", DataClass: data, Destination: destination}
}

// Secret Reference and Providers
type SecretReference struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"` // e.g. "env", "vault", "keychain"
	Description string `json:"description,omitempty"`
}

type SecretProvider interface {
	Resolve(ref SecretReference) (string, error)
}

type EnvSecretProvider struct{}

func (p EnvSecretProvider) Resolve(ref SecretReference) (string, error) {
	val, ok := os.LookupEnv(ref.Name)
	if !ok || val == "" {
		return "", fmt.Errorf("secret not found in environment: %s", ref.Name)
	}
	return val, nil
}

type StaticSecretProvider struct {
	secrets map[string]string
}

func NewStaticSecretProvider(secrets map[string]string) *StaticSecretProvider {
	return &StaticSecretProvider{secrets: secrets}
}

func (p *StaticSecretProvider) Resolve(ref SecretReference) (string, error) {
	val, ok := p.secrets[ref.Name]
	if !ok {
		return "", fmt.Errorf("secret not found: %s", ref.Name)
	}
	return val, nil
}

// Redactor detects sensitive values and patterns and sanitizes them
type Redactor struct {
	regexes []*regexp.Regexp
}

var defaultSensitivePatterns = []string{
	`(?i)sk-[a-zA-Z0-9_-]{20,}`,                                          // OpenAI/OpenCode style API keys
	`(?i)ghp_[a-zA-Z0-9]{36}`,                                            // GitHub Personal Access Token
	`(?i)AKIA[0-9A-Z]{16}`,                                               // AWS Access Key ID
	`(?i)bearer\s+[a-zA-Z0-9\-_\.=]{20,}`,                                // Bearer tokens
	`(?i)(?:password|secret|apikey|token)\s*[:=]\s*["']?([^\s"']+)["']?`, // Generic key=val
}

func NewRedactor(extraPatterns ...string) (*Redactor, error) {
	patterns := append([]string{}, defaultSensitivePatterns...)
	patterns = append(patterns, extraPatterns...)

	regexes := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, errors.New("invalid redaction pattern: " + p)
		}
		regexes = append(regexes, re)
	}
	return &Redactor{regexes: regexes}, nil
}

func MustNewRedactor(extraPatterns ...string) *Redactor {
	r, err := NewRedactor(extraPatterns...)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *Redactor) Redact(text string) string {
	result := text
	for _, re := range r.regexes {
		result = re.ReplaceAllString(result, "[REDACTED]")
	}
	return result
}

func (r *Redactor) RedactValues(text string, secretValues ...string) string {
	result := text
	for _, s := range secretValues {
		if len(s) > 0 {
			result = strings.ReplaceAll(result, s, "[REDACTED_SECRET]")
		}
	}
	return r.Redact(result)
}
