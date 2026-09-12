package mcpserver

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCalibrationEvidenceTracksConsumedFiles(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "scenarios", "a.yml"), "calibration_file: ../usage.json\n")
	writeTestFile(t, filepath.Join(root, "scenarios", "b.yml"), "simulation:\n  iterations: 1\n")
	server := newTestServer(t, root, &fakeRunner{})
	tools := map[string]func() ResultEnvelope{
		"validate": func() ResultEnvelope {
			return server.validateScenario(context.Background(), "scenarios/a.yml", time.Now())
		},
		"simulate": func() ResultEnvelope {
			return server.simulateScenario(context.Background(), "scenarios/a.yml", time.Now())
		},
		"compare": func() ResultEnvelope {
			return server.compareScenarios(context.Background(), []string{"scenarios/a.yml", "scenarios/b.yml"}, time.Now())
		},
		"judge": func() ResultEnvelope {
			return server.judgeStandards(context.Background(), "scenarios/a.yml", time.Now())
		},
	}
	for _, content := range []string{`{"period":"2026-08"}`, `{"period":"2026-09"}`} {
		writeTestFile(t, filepath.Join(root, "usage.json"), content)
		want := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
		for name, call := range tools {
			got := call().Input
			if len(got.Paths) < 2 || got.Paths[1] != "usage.json" || got.SHA256[1] != want {
				t.Fatalf("%s calibration evidence = %#v", name, got)
			}
		}
	}
}

func TestUnreadableCalibrationStopsChild(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "a.yml"), "calibration_file: usage.json\n")
	writeTestFile(t, filepath.Join(root, "usage.json"), strings.Repeat("x", maxScenarioBytes+1))
	runner := &fakeRunner{}
	server := newTestServer(t, root, runner)
	got := server.simulateScenario(context.Background(), "a.yml", time.Now())
	if got.Error == nil || got.Error.Code != "calibration_file_unavailable" || len(runner.Calls()) != 0 {
		t.Fatalf("oversized calibration did not fail closed: %#v", got)
	}
}

func TestStandardsExecutableName(t *testing.T) {
	for goos, want := range map[string]string{"windows": "profitctl-standards.exe", "linux": "profitctl-standards", "darwin": "profitctl-standards"} {
		if got := standardsExecutableName(goos); got != want {
			t.Fatalf("%s: got %q, want %q", goos, got, want)
		}
	}
}
