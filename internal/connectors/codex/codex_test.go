package codex_test

import (
	"testing"

	"github.com/raillen/prumo/internal/connectors/codex"
	"github.com/raillen/prumo/internal/connectors/testkit"
)

func TestCodexConnector(t *testing.T) {
	c := codex.NewConnector()
	testkit.RunAll(t, c)
}
