package main

import (
	"encoding/json"
	"testing"
)

func TestTestCommandCompiledDryRunJSONUsesSharedDiscoveryPath(t *testing.T) {
	dir := enterTempWorkingDir(t)
	writeMinimalTestCliWorkspace(t, dir)

	stdlibSrc := writeMinimalTestCliStdlib(t, dir)
	configureMinimalTestCliEnv(t, stdlibSrc, true)

	stdout := runCLIExpectSuccess(t,
		"test",
		"--compiled",
		"--dry-run",
		"--format", "json",
		".",
	)

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout is not valid JSON array: %v (stdout=%q)", err, stdout)
	}
	if len(payload) != 1 {
		t.Fatalf("expected one descriptor, got %d (%v)", len(payload), payload)
	}
	if payload[0]["framework_id"] != "demo.framework" {
		t.Fatalf("unexpected framework_id %v", payload[0]["framework_id"])
	}
	if payload[0]["display_name"] != "example works" {
		t.Fatalf("unexpected display_name %v", payload[0]["display_name"])
	}
}
