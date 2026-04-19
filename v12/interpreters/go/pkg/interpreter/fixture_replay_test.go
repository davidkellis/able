//go:build !(js && wasm)

package interpreter

import (
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestReplayFixtureEvaluatesSourceModule(t *testing.T) {
	fixtureDir := t.TempDir()
	sourcePath := filepath.Join(fixtureDir, "source.able")
	if err := os.WriteFile(sourcePath, []byte("package sample\n\nprint(\"hello\")\n\"done\"\n"), 0o644); err != nil {
		t.Fatalf("write source fixture: %v", err)
	}

	result, err := ReplayFixture(fixtureDir, "source.able", nil, nil)
	if err != nil {
		t.Fatalf("ReplayFixture: %v", err)
	}
	if result.RuntimeError != "" {
		t.Fatalf("unexpected runtime error %q", result.RuntimeError)
	}
	if len(result.Stdout) != 1 || result.Stdout[0] != "hello" {
		t.Fatalf("stdout = %v, want [hello]", result.Stdout)
	}
	stringValue, ok := result.Value.(runtime.StringValue)
	if !ok {
		t.Fatalf("result.Value = %T, want runtime.StringValue", result.Value)
	}
	if stringValue.Val != "done" {
		t.Fatalf("string value = %q, want done", stringValue.Val)
	}
	if result.TypecheckMode == "" {
		t.Fatalf("expected non-empty typecheck mode")
	}
}
