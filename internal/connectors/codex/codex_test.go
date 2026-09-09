package codex_test

import (
	"testing"

	"github.com/raillen/project-atlas-framework/internal/connectors/codex"
	"github.com/raillen/project-atlas-framework/internal/connectors/testkit"
)

func TestCodexConnector(t *testing.T) {
	c := codex.NewConnector()
	testkit.RunAll(t, c)
}
