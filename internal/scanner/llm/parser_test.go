package llm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAnalysisResponse_PlainJSON(t *testing.T) {
	raw := `{
  "services": [{"name":"postgres","type":"database","provider":"aws"}],
  "dependencies": [{"name":"gorm","type":"library"}],
  "patterns": [{"description":"read-heavy","confidence":0.8}]
}`

	analysis, err := ParseAnalysisResponse(raw)
	assert.NoError(t, err)
	assert.Len(t, analysis.Services, 1)
	assert.Equal(t, "postgres", analysis.Services[0].Name)
	assert.Len(t, analysis.Dependencies, 1)
	assert.Len(t, analysis.Patterns, 1)
}

func TestParseAnalysisResponse_FencedJSON(t *testing.T) {
	raw := "```json\n{\n  \"services\": [],\n  \"dependencies\": [],\n  \"patterns\": []\n}\n```"

	analysis, err := ParseAnalysisResponse(raw)
	assert.NoError(t, err)
	assert.NotNil(t, analysis.Services)
	assert.NotNil(t, analysis.Dependencies)
	assert.NotNil(t, analysis.Patterns)
}

func TestParseAnalysisResponse_EmptyResponse(t *testing.T) {
	_, err := ParseAnalysisResponse("   ")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrEmptyResponse))
}

func TestParseAnalysisResponse_InvalidJSON(t *testing.T) {
	_, err := ParseAnalysisResponse("not-json")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidJSON))
}

func TestParseAnalysisResponse_InvalidSchema(t *testing.T) {
	raw := `{
  "services": [{"name":"","type":"database","provider":"aws"}],
  "dependencies": [],
  "patterns": []
}`
	_, err := ParseAnalysisResponse(raw)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidSchema))
}

func TestParseAnalysisResponse_InvalidConfidence(t *testing.T) {
	raw := `{
  "services": [],
  "dependencies": [],
  "patterns": [{"description":"bad","confidence":2.2}]
}`
	_, err := ParseAnalysisResponse(raw)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidSchema))
}
