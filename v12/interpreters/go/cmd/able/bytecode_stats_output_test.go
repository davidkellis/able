package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/interpreter"
)

func TestWriteBytecodeStatsSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "stats.json")
	want := interpreter.BytecodeStatsSnapshot{Enabled: true, ValueStackMaxDepth: 7, CallFrameMaxDepth: 3}
	if err := writeBytecodeStatsSnapshot(path, want); err != nil {
		t.Fatalf("writeBytecodeStatsSnapshot() error = %v", err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	var got interpreter.BytecodeStatsSnapshot
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if got.Enabled != want.Enabled || got.ValueStackMaxDepth != want.ValueStackMaxDepth || got.CallFrameMaxDepth != want.CallFrameMaxDepth {
		t.Fatalf("snapshot = %+v, want %+v", got, want)
	}
}

func TestBytecodeStatsMainOnlyEnabled(t *testing.T) {
	t.Setenv(bytecodeStatsMainOnlyEnv, "")
	if bytecodeStatsMainOnlyEnabled() {
		t.Fatalf("main-only stats unexpectedly enabled")
	}
	t.Setenv(bytecodeStatsMainOnlyEnv, "1")
	if !bytecodeStatsMainOnlyEnabled() {
		t.Fatalf("main-only stats not enabled")
	}
}
