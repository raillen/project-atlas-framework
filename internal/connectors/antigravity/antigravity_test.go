package antigravity_test

import (
	"testing"

	"github.com/raillen/prumo/internal/connectors/antigravity"
	"github.com/raillen/prumo/internal/connectors/testkit"
)

func TestAntigravityConnector(t *testing.T) {
	c := antigravity.NewConnector()
	testkit.RunAll(t, c)
}
