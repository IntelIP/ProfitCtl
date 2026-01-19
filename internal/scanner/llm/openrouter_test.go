package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenRouterProvider(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		model    string
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "valid provider",
			apiKey:  "test-key",
			model:   "anthropic/claude-3-haiku",
			wantErr: false,
		},
		{
			name:    "empty API key",
			apiKey:  "",
			model:   "test-model",
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name:    "default model",
			apiKey:  "test-key",
			model:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOpenRouterProvider(tt.apiKey, tt.model)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				if tt.model == "" {
					assert.Equal(t, "anthropic/claude-3-haiku", provider.model)
				} else {
					assert.Equal(t, tt.model, provider.model)
				}
			}
		})
	}
}

func TestOpenRouterProvider_Chat_EmptyMessages(t *testing.T) {
	provider, err := NewOpenRouterProvider("test-key", "test-model")
	require.NoError(t, err)

	ctx := context.Background()
	_, err = provider.Chat(ctx, []Message{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "messages cannot be empty")
}

func TestOpenRouterProvider_Chat_InvalidAPIKey(t *testing.T) {
	provider, err := NewOpenRouterProvider("invalid-key", "test-model")
	require.NoError(t, err)

	ctx := context.Background()
	messages := []Message{
		{Role: "user", Content: "Hello"},
	}

	_, err = provider.Chat(ctx, messages)

	// Should fail due to not implemented
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestOpenRouterProvider_IsAvailable_InvalidKey(t *testing.T) {
	provider, err := NewOpenRouterProvider("invalid-key", "test-model")
	require.NoError(t, err)

	ctx := context.Background()
	available := provider.IsAvailable(ctx)

	// Should return false due to invalid API key
	assert.False(t, available)
}

// Integration test - only runs if OPENROUTER_API_KEY environment variable is set
func TestOpenRouterProvider_Integration(t *testing.T) {
	apiKey := "test" // Replace with actual key for integration testing
	if apiKey == "test" {
		t.Skip("Skipping integration test - set OPENROUTER_API_KEY to run")
	}

	provider, err := NewOpenRouterProvider(apiKey, "anthropic/claude-3-haiku")
	require.NoError(t, err)

	ctx := context.Background()

	// Test availability
	available := provider.IsAvailable(ctx)
	assert.True(t, available)

	// Test chat
	messages := []Message{
		{Role: "user", Content: "Say 'Hello World' and nothing else."},
	}

	response, err := provider.Chat(ctx, messages)
	assert.NoError(t, err)
	assert.Contains(t, response, "Hello World")
}