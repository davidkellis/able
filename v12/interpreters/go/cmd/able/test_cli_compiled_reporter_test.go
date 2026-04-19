package main

import (
	"encoding/json"
	"strings"
	"testing"

	testclipkg "able/interpreter-go/pkg/testcli"
)

func TestTestCommandCompiledJSONReporterEmitsEvents(t *testing.T) {
	dir := enterTempWorkingDir(t)
	target := writeCompiledSampleTestModule(t, dir)

	configureRepoCompiledEnv(t)
	stdout := runCLIExpectSuccess(t,
		"test",
		"--compiled",
		"--format", "json",
		target,
	)

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) == 0 {
		t.Fatalf("expected JSON event lines, got %q", stdout)
	}

	for _, line := range lines {
		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			t.Fatalf("stdout line is not valid JSON: %v (line=%q)", err, line)
		}
		if payload["event"] != "case_passed" {
			continue
		}
		descriptor, ok := payload["descriptor"].(map[string]any)
		if !ok {
			t.Fatalf("missing descriptor payload: %v", payload["descriptor"])
		}
		if name, ok := descriptor["display_name"].(string); ok && strings.Contains(name, "adds") {
			return
		}
	}
	t.Fatalf("expected a case_passed event for the sample test, got %q", stdout)
}

func TestTestCommandCompiledJSONReporterPreservesDescriptorFields(t *testing.T) {
	requireGoToolchain(t)

	dir := enterTempWorkingDir(t)
	writeMinimalTestCliWorkspace(t, dir)

	stdlibSrc := writeMinimalTestCliStdlib(t, dir)
	configureMinimalTestCliEnv(t, stdlibSrc, true)

	stdout := runCLIExpectSuccess(t,
		"test",
		"--compiled",
		"--format", "json",
		".",
	)

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) == 0 {
		t.Fatalf("expected JSON event lines, got %q", stdout)
	}

	for _, line := range lines {
		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			t.Fatalf("stdout line is not valid JSON: %v (line=%q)", err, line)
		}
		if payload["event"] != "case_passed" {
			continue
		}
		descriptor, ok := payload["descriptor"].(map[string]any)
		if !ok {
			t.Fatalf("missing descriptor payload: %v", payload["descriptor"])
		}

		tags, ok := descriptor["tags"].([]any)
		if !ok || len(tags) != 1 || tags[0] != "fast" {
			t.Fatalf("unexpected tags payload %v", descriptor["tags"])
		}
		metadata, ok := descriptor["metadata"].([]any)
		if !ok || len(metadata) != 1 {
			t.Fatalf("unexpected metadata payload %v", descriptor["metadata"])
		}
		entry, ok := metadata[0].(map[string]any)
		if !ok || entry["key"] != "kind" || entry["value"] != "demo" {
			t.Fatalf("unexpected metadata entry %v", metadata[0])
		}
		location, ok := descriptor["location"].(map[string]any)
		if !ok {
			t.Fatalf("unexpected location payload %v", descriptor["location"])
		}
		if location["module_path"] != "pkg/tests/example.test.able" || location["line"] != float64(7) || location["column"] != float64(3) {
			t.Fatalf("unexpected location payload %v", location)
		}
		return
	}
	t.Fatalf("expected a case_passed event with descriptor fields, got %q", stdout)
}

func TestTestCommandCompiledTapReporterEmitsTAP(t *testing.T) {
	dir := enterTempWorkingDir(t)
	target := writeCompiledSampleTestModule(t, dir)

	configureRepoCompiledEnv(t)
	stdout := runCLIExpectSuccess(t,
		"test",
		"--compiled",
		"--format", "tap",
		target,
	)

	assertOutputContainsAll(t, stdout, "TAP version 13", "ok 1 -", "adds")
}

func TestTestCommandCompiledJSONReporterEmitsSkipAndFailureEvents(t *testing.T) {
	requireGoToolchain(t)

	dir := enterTempWorkingDir(t)
	writeMinimalTestCliWorkspace(t, dir)

	stdlibSrc := writeMinimalTestCliReporterEventsStdlib(t, dir)
	configureMinimalTestCliEnv(t, stdlibSrc, true)

	stdout, _ := runCLIExpectFailureCode(t, 1,
		"test",
		"--compiled",
		"--format", "json",
		".",
	)

	var sawSkipped bool
	var sawFailed bool
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			t.Fatalf("stdout line is not valid JSON: %v (line=%q)", err, line)
		}
		switch payload["event"] {
		case "case_skipped":
			descriptor, ok := payload["descriptor"].(map[string]any)
			if !ok {
				t.Fatalf("missing skip descriptor payload: %v", payload["descriptor"])
			}
			if descriptor["display_name"] != "skipped example" || payload["reason"] != "pending" {
				t.Fatalf("unexpected skipped payload %v", payload)
			}
			sawSkipped = true
		case "case_failed":
			descriptor, ok := payload["descriptor"].(map[string]any)
			if !ok {
				t.Fatalf("missing failure descriptor payload: %v", payload["descriptor"])
			}
			failure, ok := payload["failure"].(map[string]any)
			if !ok {
				t.Fatalf("missing failure payload: %v", payload["failure"])
			}
			if descriptor["display_name"] != "failed example" || failure["message"] != "boom" || failure["details"] != "extra detail" {
				t.Fatalf("unexpected failure payload %v", payload)
			}
			location, ok := failure["location"].(map[string]any)
			if !ok {
				t.Fatalf("missing failure location payload: %v", failure["location"])
			}
			if location["module_path"] != "pkg/tests/failed.test.able" || location["line"] != float64(11) || location["column"] != float64(9) {
				t.Fatalf("unexpected failure location payload %v", location)
			}
			sawFailed = true
		}
	}
	if !sawSkipped || !sawFailed {
		t.Fatalf("expected skipped and failed events, got %q", stdout)
	}
}

func TestTestCommandCompiledJSONReporterPreservesStartedEventOrder(t *testing.T) {
	requireGoToolchain(t)

	dir := enterTempWorkingDir(t)
	writeMinimalTestCliWorkspace(t, dir)

	stdlibSrc := writeMinimalTestCliReporterEventsStdlib(t, dir)
	configureMinimalTestCliEnv(t, stdlibSrc, true)

	stdout, _ := runCLIExpectFailureCode(t, 1,
		"test",
		"--compiled",
		"--format", "json",
		".",
	)

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	expected := []struct {
		kind string
		name string
	}{
		{kind: "case_started", name: "skipped example"},
		{kind: "case_skipped", name: "skipped example"},
		{kind: "case_started", name: "failed example"},
		{kind: "case_failed", name: "failed example"},
	}
	if len(lines) != len(expected) {
		t.Fatalf("expected %d JSON events, got %d: %q", len(expected), len(lines), stdout)
	}

	for idx, line := range lines {
		var event testclipkg.TestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("stdout line is not valid JSON: %v (line=%q)", err, line)
		}
		if event.Kind != expected[idx].kind {
			t.Fatalf("event %d kind = %q, want %q (stdout=%q)", idx, event.Kind, expected[idx].kind, stdout)
		}
		if event.Descriptor == nil || event.Descriptor.DisplayName != expected[idx].name {
			t.Fatalf("event %d descriptor = %#v, want %q", idx, event.Descriptor, expected[idx].name)
		}
	}
}

func TestTestCommandCompiledTapReporterEmitsSkipAndFailureDiagnostics(t *testing.T) {
	requireGoToolchain(t)

	dir := enterTempWorkingDir(t)
	writeMinimalTestCliWorkspace(t, dir)

	stdlibSrc := writeMinimalTestCliReporterEventsStdlib(t, dir)
	configureMinimalTestCliEnv(t, stdlibSrc, true)

	stdout, _ := runCLIExpectFailureCode(t, 1,
		"test",
		"--compiled",
		"--format", "tap",
		".",
	)

	assertOutputContainsAll(
		t,
		stdout,
		"TAP version 13",
		"ok 1 - skipped example # SKIP pending",
		"not ok 2 - failed example",
		"message: boom",
		"details: extra detail",
		"location: pkg/tests/failed.test.able:11:9",
	)
}

func TestTestCommandCompiledTapReporterCountsOnlyTerminalEvents(t *testing.T) {
	requireGoToolchain(t)

	dir := enterTempWorkingDir(t)
	writeMinimalTestCliWorkspace(t, dir)

	stdlibSrc := writeMinimalTestCliReporterEventsStdlib(t, dir)
	configureMinimalTestCliEnv(t, stdlibSrc, true)

	stdout, _ := runCLIExpectFailureCode(t, 1,
		"test",
		"--compiled",
		"--format", "tap",
		".",
	)

	var testPoints []string
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if strings.HasPrefix(line, "ok ") || strings.HasPrefix(line, "not ok ") {
			testPoints = append(testPoints, line)
		}
	}

	expected := []string{
		"ok 1 - skipped example # SKIP pending",
		"not ok 2 - failed example",
	}
	if len(testPoints) != len(expected) {
		t.Fatalf("expected %d TAP test points, got %d: %q", len(expected), len(testPoints), stdout)
	}
	for idx, line := range testPoints {
		if line != expected[idx] {
			t.Fatalf("TAP test point %d = %q, want %q", idx, line, expected[idx])
		}
	}
}

func TestTestCommandCompiledJSONReporterEmitsFrameworkErrorEvent(t *testing.T) {
	requireGoToolchain(t)

	dir := enterTempWorkingDir(t)
	writeMinimalTestCliWorkspace(t, dir)

	stdlibSrc := writeMinimalTestCliFrameworkErrorStdlib(t, dir)
	configureMinimalTestCliEnv(t, stdlibSrc, true)

	stdout, _ := runCLIExpectFailureCode(t, 2,
		"test",
		"--compiled",
		"--format", "json",
		".",
	)

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected one framework error event, got %q", stdout)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &payload); err != nil {
		t.Fatalf("stdout line is not valid JSON: %v (line=%q)", err, lines[0])
	}
	if payload["event"] != "framework_error" || payload["message"] != "broken harness" {
		t.Fatalf("unexpected framework error payload %v", payload)
	}
}

func TestTestCommandCompiledTapReporterEmitsFrameworkErrorBailOut(t *testing.T) {
	requireGoToolchain(t)

	dir := enterTempWorkingDir(t)
	writeMinimalTestCliWorkspace(t, dir)

	stdlibSrc := writeMinimalTestCliFrameworkErrorStdlib(t, dir)
	configureMinimalTestCliEnv(t, stdlibSrc, true)

	stdout, _ := runCLIExpectFailureCode(t, 2,
		"test",
		"--compiled",
		"--format", "tap",
		".",
	)

	assertOutputContainsAll(t, stdout, "TAP version 13", "Bail out! broken harness")
}
