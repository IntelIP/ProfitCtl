package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenRouterProvider_UsesTerraByDefault(t *testing.T) {
	provider, err := NewOpenRouterProvider("test-key", "")

	require.NoError(t, err)
	assert.Equal(t, DefaultOpenRouterModel, provider.model)
}
