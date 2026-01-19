package llm

import (
	stdcontext "context"
)

// CodeContext holds collected file contents for LLM analysis
type CodeContext struct {
	Files map[string]string
}

// AnalysisRequest represents a request for LLM analysis
type AnalysisRequest struct {
	Context CodeContext
	Prompt  string
}

// AnalysisResponse represents LLM analysis results
type AnalysisResponse struct {
	Services     []DetectedService
	Dependencies []DetectedDependency
	Patterns     []UsagePattern
}

// DetectedService represents a service detected from config files
type DetectedService struct {
	Name         string
	Type         string
	Provider     string
	FixedCost   *float64 `json:"fixed_cost,omitempty"`
	VariableCost *float64 `json:"variable_cost,omitempty"`
}

// DetectedDependency represents a dependency detected from config files
type DetectedDependency struct {
	Name string
	Type string
}

// UsagePattern represents a usage pattern detected from config files
type UsagePattern struct {
	Description string
	Confidence  float64
}

// LLMProvider interface defines contract for LLM providers
type LLMProvider interface {
	Chat(ctx stdcontext.Context, messages []Message) (string, error)
	IsAvailable(ctx stdcontext.Context) bool
}

// Message represents a single message in a conversation with LLM
type Message struct {
	Role    string
	Content string
}
