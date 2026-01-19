package llm

import (
	"strings"
	"testing"
)

func TestBuildDetectionPrompt(t *testing.T) {
	ctx := CodeContext{
		Files: map[string]string{
			"go.mod": `module github.com/example/app

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/lib/pq v1.10.9
)`,
			"package.json": `{
  "name": "example-app",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.18.2",
    "redis": "^4.6.8"
  }
}`,
		},
	}

	prompt := BuildDetectionPrompt(ctx)

	// Check that prompt contains expected sections
	if !strings.Contains(prompt, "# Codebase Analysis") {
		t.Error("Prompt should contain header")
	}

	if !strings.Contains(prompt, "## go.mod") {
		t.Error("Prompt should contain go.mod section")
	}

	if !strings.Contains(prompt, "## package.json") {
		t.Error("Prompt should contain package.json section")
	}

	if !strings.Contains(prompt, "## Analysis Instructions") {
		t.Error("Prompt should contain analysis instructions")
	}

	// Check that code blocks are present
	if !strings.Contains(prompt, "```") {
		t.Error("Prompt should contain code blocks")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "short content",
			input:     "short",
			maxLength: 10,
			expected:  "short",
		},
		{
			name:      "exact length",
			input:     "exactly ten",
			maxLength: 11,
			expected:  "exactly ten",
		},
		{
			name:      "truncate at word boundary",
			input:     "this is a long sentence that should be truncated",
			maxLength: 20,
			expected:  "this is a long",
		},
		{
			name:      "truncate mid-word when no boundary",
			input:     "verylongwordwithoutspaces",
			maxLength: 10,
			expected:  "verylongwo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLength)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLength, result, tt.expected)
			}
		})
	}
}

func TestBuildDetectionPromptWithTruncation(t *testing.T) {
	longContent := strings.Repeat("a", 3000)
	ctx := CodeContext{
		Files: map[string]string{
			"long-file.txt": longContent,
		},
	}

	prompt := BuildDetectionPrompt(ctx)

	// Find the code block content
	lines := strings.Split(prompt, "\n")
	var codeBlock string
	inCodeBlock := false
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				break
			}
			inCodeBlock = true
			continue
		}
		if inCodeBlock {
			codeBlock += line
		}
	}

	// Should be truncated to 2000 chars
	if len(codeBlock) > 2000 {
		t.Errorf("Code block content should be truncated to 2000 chars, got %d", len(codeBlock))
	}
}