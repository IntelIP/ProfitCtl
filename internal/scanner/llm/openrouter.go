package llm

import (
	"context"
	"fmt"

	"github.com/revrost/go-openrouter"
)

// OpenRouterProvider implements the LLMProvider interface using OpenRouter API
type OpenRouterProvider struct {
	client *openrouter.Client
	model  string
}

// NewOpenRouterProvider creates a new OpenRouter provider instance
func NewOpenRouterProvider(apiKey, model string) (*OpenRouterProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if model == "" {
		model = "anthropic/claude-3-haiku" // Default cost-effective model
	}

	client := openrouter.NewClient(apiKey)

	return &OpenRouterProvider{
		client: client,
		model:  model,
	}, nil
}

// Chat sends a conversation to OpenRouter and returns the response
func (p *OpenRouterProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("messages cannot be empty")
	}

	// For now, return not implemented until API is clarified
	return "", fmt.Errorf("OpenRouter provider not yet implemented - awaiting API clarification")
}

// IsAvailable checks if the OpenRouter service is accessible
func (p *OpenRouterProvider) IsAvailable(ctx context.Context) bool {
	// Temporarily return false until implementation is complete
	return false
}