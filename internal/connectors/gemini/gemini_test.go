package gemini_test

import (
	"testing"

	"github.com/raillen/prumo/internal/connectors/gemini"
	"github.com/raillen/prumo/internal/connectors/testkit"
)

func TestGeminiConnector(t *testing.T) {
	c := gemini.NewConnector()
	testkit.RunAll(t, c)
}
