package egress

import "testing"

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
