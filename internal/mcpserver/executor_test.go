package mcpserver

import (
	"context"
	"testing"
)

func TestExecRunnerScrubsInheritedProviderCredentials(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "test-only-secret")
	t.Setenv("OPENAI_API_KEY", "test-only-secret")

	result := (execRunner{}).Run(
		context.Background(),
		"/bin/sh",
		[]string{"-c", "printf '%s:%s' \"$OPENROUTER_API_KEY\" \"$OPENAI_API_KEY\""},
		t.TempDir(),
		1024,
	)
	if result.err != nil || result.exitCode != 0 {
		t.Fatalf("run child: exit=%d err=%v", result.exitCode, result.err)
	}
	if got := string(result.stdout); got != ":" {
		t.Fatalf("child inherited provider credentials: %q", got)
	}
}
