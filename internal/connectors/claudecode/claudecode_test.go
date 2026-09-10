package claudecode_test

import (
	"testing"

	"github.com/raillen/prumo/internal/connectors/claudecode"
	"github.com/raillen/prumo/internal/connectors/testkit"
)

func TestClaudeCodeConnector(t *testing.T) {
	c := claudecode.NewConnector()
	testkit.RunAll(t, c)
}
