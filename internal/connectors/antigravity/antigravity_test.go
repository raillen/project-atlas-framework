package antigravity_test

import (
	"testing"

	"github.com/raillen/project-atlas-framework/internal/connectors/antigravity"
	"github.com/raillen/project-atlas-framework/internal/connectors/testkit"
)

func TestAntigravityConnector(t *testing.T) {
	c := antigravity.NewConnector()
	testkit.RunAll(t, c)
}
