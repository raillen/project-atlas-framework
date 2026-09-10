package egress

import (
	"os"
	"strings"
	"testing"
)

func TestRestrictedExternalDenied(t *testing.T) {
	if Allow(Restricted, "external", false).Allowed {
		t.Fatal("restricted external egress allowed")
	}
}

func TestPublicExternalAllowed(t *testing.T) {
	if !Allow(Public, "external", false).Allowed {
		t.Fatal("public external egress denied")
	}
}

func TestAllowEgressPolicyRules(t *testing.T) {
	// Restricted
	dRestr := AllowEgress(Restricted, "https://api.openai.com", true, true)
	if dRestr.Allowed {
		t.Fatal("restricted data egress should be denied externally")
	}

	// Confidential without TLS
	dConfNoTLS := AllowEgress(Confidential, "http://trusted-host.com", false, true)
	if dConfNoTLS.Allowed {
		t.Fatal("confidential data without TLS should be denied")
	}

	// Confidential with TLS and trusted host
	dConfOK := AllowEgress(Confidential, "https://trusted-host.com", true, true)
	if !dConfOK.Allowed {
		t.Fatalf("confidential data with TLS and trusted host should be allowed: %s", dConfOK.Reason)
	}

	// Confidential with untrusted host
	dConfUntrusted := AllowEgress(Confidential, "https://unknown.com", true, false)
	if dConfUntrusted.Allowed {
		t.Fatal("confidential data with untrusted host should be denied")
	}
}

func TestSecretProviders(t *testing.T) {
	staticProv := NewStaticSecretProvider(map[string]string{
		"GITHUB_TOKEN": "ghp_secret12345678901234567890123456",
	})
	val, err := staticProv.Resolve(SecretReference{ID: "gh", Name: "GITHUB_TOKEN", Provider: "static"})
	if err != nil || val != "ghp_secret12345678901234567890123456" {
		t.Fatalf("static secret resolution failed: %v, %s", err, val)
	}

	os.Setenv("PRUMO_TEST_SECRET", "my-env-secret-val")
	defer os.Unsetenv("PRUMO_TEST_SECRET")

	envProv := EnvSecretProvider{}
	valEnv, err := envProv.Resolve(SecretReference{ID: "test", Name: "PRUMO_TEST_SECRET", Provider: "env"})
	if err != nil || valEnv != "my-env-secret-val" {
		t.Fatalf("env secret resolution failed: %v, %s", err, valEnv)
	}
}

func TestRedactor(t *testing.T) {
	r := MustNewRedactor()

	input := "Bearer abcdef12345678901234567890 and key sk-123456789012345678901234567890 and normal text"
	redacted := r.Redact(input)
	if strings.Contains(redacted, "sk-1234567890") {
		t.Fatalf("API key was not redacted: %s", redacted)
	}
	if strings.Contains(redacted, "abcdef1234567890") {
		t.Fatalf("Bearer token was not redacted: %s", redacted)
	}
	if !strings.Contains(redacted, "normal text") {
		t.Fatalf("normal text should not be redacted: %s", redacted)
	}

	// Redact exact values
	exactRedacted := r.RedactValues("connect with super-secret-password in connection string", "super-secret-password")
	if strings.Contains(exactRedacted, "super-secret-password") {
		t.Fatalf("exact secret value was not redacted: %s", exactRedacted)
	}
}
