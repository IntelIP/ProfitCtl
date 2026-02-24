package llm

import (
	"fmt"
	"sort"
	"strings"
)

// BuildDetectionPrompt formats collected code files into a markdown prompt for LLM analysis.
func BuildDetectionPrompt(ctx CodeContext) string {
	var sb strings.Builder

	sb.WriteString("# Codebase Analysis\n\n")
	sb.WriteString("Analyze the following configuration files to detect cloud services, dependencies, and usage patterns.\n\n")

	paths := make([]string, 0, len(ctx.Files))
	for path := range ctx.Files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		sb.WriteString(fmt.Sprintf("## %s\n\n", path))
		sb.WriteString("```\n")
		sb.WriteString(truncate(ctx.Files[path], 2000))
		sb.WriteString("\n```\n\n")
	}

	sb.WriteString("## Analysis Instructions\n\n")
	sb.WriteString("Please identify:\n")
	sb.WriteString("- Cloud services and their providers (for example AWS RDS, Google Cloud Storage)\n")
	sb.WriteString("- Infrastructure dependencies (for example databases, caches, message queues)\n")
	sb.WriteString("- Cost implications and pricing models\n")
	sb.WriteString("- Usage patterns that might affect scaling or costs\n\n")
	sb.WriteString("Return only a valid JSON object with the following structure (no markdown, no code fences):\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"services\": [\n")
	sb.WriteString("    {\n")
	sb.WriteString("      \"name\": \"service name\",\n")
	sb.WriteString("      \"type\": \"database|storage|compute|etc\",\n")
	sb.WriteString("      \"provider\": \"aws|gcp|azure|etc\",\n")
	sb.WriteString("      \"fixed_cost\": 100.0,\n")
	sb.WriteString("      \"variable_cost\": 0.10\n")
	sb.WriteString("    }\n")
	sb.WriteString("  ],\n")
	sb.WriteString("  \"dependencies\": [\n")
	sb.WriteString("    {\"name\": \"dependency name\", \"type\": \"library|framework\"}\n")
	sb.WriteString("  ],\n")
	sb.WriteString("  \"patterns\": [\n")
	sb.WriteString("    {\"description\": \"usage pattern description\", \"confidence\": 0.8}\n")
	sb.WriteString("  ]\n")
	sb.WriteString("}\n")

	return sb.String()
}

// truncate limits a string to maxLength characters, preserving word boundaries where possible.
func truncate(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}

	truncated := content[:maxLength]
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxLength/2 {
		return truncated[:lastSpace]
	}

	return truncated
}
