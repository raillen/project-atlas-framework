package gemini_test

import (
	"testing"

	"github.com/raillen/project-atlas-framework/internal/connectors/gemini"
	"github.com/raillen/project-atlas-framework/internal/connectors/testkit"
)

func TestGeminiConnector(t *testing.T) {
	c := gemini.NewConnector()
	testkit.RunAll(t, c)
}
