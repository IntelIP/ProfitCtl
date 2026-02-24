package llm

import stdcontext "context"

// CodeContext holds collected file contents for LLM analysis.
type CodeContext struct {
	Files map[string]string
}

// AnalysisRequest represents a request for LLM analysis.
type AnalysisRequest struct {
	Context CodeContext
	Prompt  string
}

// AnalysisResponse represents LLM analysis results.
type AnalysisResponse struct {
	Services     []DetectedService    `json:"services"`
	Dependencies []DetectedDependency `json:"dependencies"`
	Patterns     []UsagePattern       `json:"patterns"`
}

// DetectedService represents a service detected from config files.
type DetectedService struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Provider     string   `json:"provider"`
	FixedCost    *float64 `json:"fixed_cost,omitempty"`
	VariableCost *float64 `json:"variable_cost,omitempty"`
}

// DetectedDependency represents a dependency detected from config files.
type DetectedDependency struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// UsagePattern represents a usage pattern detected from config files.
type UsagePattern struct {
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// LLMProvider defines the contract for LLM providers.
type LLMProvider interface {
	Chat(ctx stdcontext.Context, messages []Message) (string, error)
	IsAvailable(ctx stdcontext.Context) bool
}

// Message represents a single message in a conversation with an LLM.
type Message struct {
	Role    string
	Content string
}
