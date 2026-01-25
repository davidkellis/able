package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTestCommandReportsEmptyWorkspaceInListMode(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	}()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	code, stdout, stderr := captureCLI(t, []string{"test", "--list"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr=%q)", code, stderr)
	}
	if !strings.Contains(stdout, "able test: no test modules found") {
		t.Fatalf("expected stdout to mention empty workspace, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandParsesFiltersAndTargets(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	}()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatalf("mkdir tests: %v", err)
	}
	writeFile(t, filepath.Join(dir, "tests", "example.test.able"), `
package tests
`)

	stdlibRoot := filepath.Join(dir, "stdlib")
	stdlibSrc := filepath.Join(stdlibRoot, "src")
	if err := os.MkdirAll(filepath.Join(stdlibSrc, "test"), 0o755); err != nil {
		t.Fatalf("mkdir stdlib test: %v", err)
	}
	writeFile(t, filepath.Join(stdlibSrc, "package.yml"), `
name: able
`)
	writeFile(t, filepath.Join(stdlibSrc, "test", "protocol.able"), `
package protocol

import able.kernel.{Array}

struct SourceLocation {
  module_path: String,
  line: i32,
  column: i32
}

struct DiscoveryRequest {
  include_paths: Array String,
  exclude_paths: Array String,
  include_names: Array String,
  exclude_names: Array String,
  include_tags: Array String,
  exclude_tags: Array String,
  list_only: bool
}

struct RunOptions {
  shuffle_seed: ?i64,
  fail_fast: bool,
  parallelism: i32,
  repeat: i32
}

struct MetadataEntry {
  key: String,
  value: String
}

struct TestDescriptor {
  framework_id: String,
  module_path: String,
  test_id: String,
  display_name: String,
  location: ?SourceLocation,
  tags: Array String,
  metadata: Array MetadataEntry
}

struct TestPlan { descriptors: Array TestDescriptor }

struct Failure {
  message: String,
  details: ?String,
  location: ?SourceLocation
}

struct case_started { descriptor: TestDescriptor }
struct case_passed { descriptor: TestDescriptor, duration_ms: i64 }
struct case_failed { descriptor: TestDescriptor, duration_ms: i64, failure: Failure }
struct case_skipped { descriptor: TestDescriptor, reason: ?String }
struct framework_error { message: String }

union TestEvent = case_started | case_passed | case_failed | case_skipped | framework_error

interface Reporter {
  fn emit(self: Self, event: TestEvent) -> void
}

interface Framework {
  fn id(self: Self) -> String
  fn discover(
    self: Self,
    request: DiscoveryRequest,
    register: TestDescriptor -> void
  ) -> ?Failure
  fn run(
    self: Self,
    plan: TestPlan,
    options: RunOptions,
    reporter: Reporter
  ) -> ?Failure
}
`)
	writeFile(t, filepath.Join(stdlibSrc, "test", "harness.able"), `
package harness

import able.kernel.{Array}
import able.test.protocol.{
  DiscoveryRequest,
  Failure,
  MetadataEntry,
  Reporter,
  RunOptions,
  TestDescriptor,
  TestPlan,
  case_passed
}

fn discover_all(_request: DiscoveryRequest) -> Array TestDescriptor | Failure {
  tags := ["fast"]
  metadata := [MetadataEntry { key: "kind", value: "demo" }]
  [TestDescriptor {
    framework_id: "demo.framework",
    module_path: "pkg",
    test_id: "demo-1",
    display_name: "example works",
    location: nil,
    tags,
    metadata
  }]
}

fn run_plan(plan: TestPlan, _options: RunOptions, reporter: Reporter) -> ?Failure {
  idx := 0
  loop {
    if idx >= plan.descriptors.len() { break }
    plan.descriptors.get(idx) match {
      case nil => {},
      case descriptor: TestDescriptor => reporter.emit(case_passed { descriptor, duration_ms: 0 })
    }
    idx = idx + 1
  }
  nil
}
`)
	writeFile(t, filepath.Join(stdlibSrc, "test", "reporters.able"), `
package reporters

import able.test.protocol.{Reporter, TestEvent}

struct DocReporterState {}
struct ProgressReporterState {}

fn DocReporter(_output: String -> void) -> DocReporterState { DocReporterState {} }
fn ProgressReporter(_output: String -> void) -> ProgressReporterState { ProgressReporterState {} }

impl Reporter for DocReporterState {
  fn emit(self: Self, _event: TestEvent) -> void {}
}

impl Reporter for ProgressReporterState {
  fn emit(self: Self, _event: TestEvent) -> void {}
}

methods ProgressReporterState {
  fn finish(self: Self) -> void {}
}
`)

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc+string(os.PathListSeparator)+repoKernelPath(t))
	t.Setenv("ABLE_TYPECHECK_FIXTURES", "off")

	code, stdout, stderr := captureCLI(t, []string{
		"test",
		"--path", "pkg",
		"--exclude-path", "tmp",
		"--name", "example works",
		"--exclude-name", "skip",
		"--tag", "fast",
		"--exclude-tag", "flaky",
		"--format", "progress",
		"--fail-fast",
		"--repeat", "3",
		"--parallel", "2",
		"--shuffle", "123",
		"--dry-run",
		".",
	})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr=%q)", code, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
	if !strings.Contains(stdout, "demo.framework") {
		t.Fatalf("expected stdout to contain framework id, got %q", stdout)
	}
	if !strings.Contains(stdout, "example") {
		t.Fatalf("expected stdout to mention test name, got %q", stdout)
	}
	if !strings.Contains(stdout, "tags=") {
		t.Fatalf("expected stdout to mention tags, got %q", stdout)
	}
	if !strings.Contains(stdout, "metadata=") {
		t.Fatalf("expected stdout to include metadata when dry-run is set, got %q", stdout)
	}
}
