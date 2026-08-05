package llm

import (
	"context"
	"fmt"

	"github.com/revrost/go-openrouter"
)

const DefaultOpenRouterModel = "openai/gpt-5.6-terra"

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
		model = DefaultOpenRouterModel
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

	// Convert our Message type to OpenRouter format
	var openRouterMessages []openrouter.ChatCompletionMessage
	for _, msg := range messages {
		openRouterMessages = append(openRouterMessages, openrouter.ChatCompletionMessage{
			Role:    msg.Role,
			Content: openrouter.Content{Text: msg.Content},
		})
	}

	// Create chat completion request
	req := openrouter.ChatCompletionRequest{
		Model:    p.model,
		Messages: openRouterMessages,
	}

	// Make the API call
	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenRouter API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	// Extract content from the response
	content := resp.Choices[0].Message.Content.Text
	return content, nil
}

// IsAvailable checks if the OpenRouter service is accessible
func (p *OpenRouterProvider) IsAvailable(ctx context.Context) bool {
	// Try a simple chat completion to test connectivity
	testMessages := []Message{
		{Role: "user", Content: "Hello"},
	}

	_, err := p.Chat(ctx, testMessages)
	return err == nil
}
