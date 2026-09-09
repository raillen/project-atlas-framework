package claudecode_test

import (
	"testing"

	"github.com/raillen/project-atlas-framework/internal/connectors/claudecode"
	"github.com/raillen/project-atlas-framework/internal/connectors/testkit"
)

func TestClaudeCodeConnector(t *testing.T) {
	c := claudecode.NewConnector()
	testkit.RunAll(t, c)
}
