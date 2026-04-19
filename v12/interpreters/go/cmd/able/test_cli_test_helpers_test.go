package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func enterWorkingDir(t *testing.T, dir string) {
	t.Helper()

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	})
}

func enterTempWorkingDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	enterWorkingDir(t, dir)
	return dir
}

func writeMinimalTestCliWorkspace(t *testing.T, root string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(root, "tests"), 0o755); err != nil {
		t.Fatalf("mkdir tests: %v", err)
	}
	writeFile(t, filepath.Join(root, "tests", "example.test.able"), `
package tests
`)
}

func writeMinimalTestCliStdlib(t *testing.T, root string) string {
	t.Helper()

	stdlibRoot := filepath.Join(root, "stdlib")
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
  SourceLocation,
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
    location: SourceLocation {
      module_path: "pkg/tests/example.test.able",
      line: 7,
      column: 3
    },
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
	writeFile(t, filepath.Join(stdlibSrc, "os.able"), `
package os

import able.kernel.{Array, __able_os_args, __able_os_exit}

fn args() -> Array String {
  __able_os_args()
}

fn exit(code: i32) -> void {
  __able_os_exit(code)
}
`)

	return stdlibSrc
}

func writeMinimalTestCliReporterEventsStdlib(t *testing.T, root string) string {
	t.Helper()

	stdlibRoot := filepath.Join(root, "stdlib-events")
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
  SourceLocation,
  TestDescriptor,
  TestPlan,
  case_started,
  case_failed,
  case_skipped
}

fn skipped_descriptor() -> TestDescriptor {
  TestDescriptor {
    framework_id: "demo.framework",
    module_path: "pkg",
    test_id: "skip-1",
    display_name: "skipped example",
    location: SourceLocation {
      module_path: "pkg/tests/skipped.test.able",
      line: 4,
      column: 2
    },
    tags: ["slow"],
    metadata: [MetadataEntry { key: "kind", value: "skip" }]
  }
}

fn failed_descriptor() -> TestDescriptor {
  TestDescriptor {
    framework_id: "demo.framework",
    module_path: "pkg",
    test_id: "fail-1",
    display_name: "failed example",
    location: SourceLocation {
      module_path: "pkg/tests/failed.test.able",
      line: 9,
      column: 5
    },
    tags: ["focus"],
    metadata: [MetadataEntry { key: "kind", value: "fail" }]
  }
}

fn discover_all(_request: DiscoveryRequest) -> Array TestDescriptor | Failure {
  [skipped_descriptor(), failed_descriptor()]
}

fn run_plan(plan: TestPlan, _options: RunOptions, reporter: Reporter) -> ?Failure {
  if plan.descriptors.len() < 2 {
    return Failure { message: "expected two descriptors", details: nil, location: nil }
  }

  plan.descriptors.get(0) match {
    case nil => {},
    case descriptor: TestDescriptor => {
      reporter.emit(case_started { descriptor })
      reporter.emit(case_skipped { descriptor, reason: "pending" })
    }
  }

  plan.descriptors.get(1) match {
    case nil => {},
    case descriptor: TestDescriptor => {
      reporter.emit(case_started { descriptor })
      reporter.emit(case_failed {
        descriptor,
        duration_ms: 7,
        failure: Failure {
          message: "boom",
          details: "extra detail",
          location: SourceLocation {
            module_path: "pkg/tests/failed.test.able",
            line: 11,
            column: 9
          }
        }
      })
    }
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
	writeFile(t, filepath.Join(stdlibSrc, "os.able"), `
package os

import able.kernel.{Array, __able_os_args, __able_os_exit}

fn args() -> Array String {
  __able_os_args()
}

fn exit(code: i32) -> void {
  __able_os_exit(code)
}
`)

	return stdlibSrc
}

func writeMinimalTestCliFrameworkErrorStdlib(t *testing.T, root string) string {
	t.Helper()

	stdlibRoot := filepath.Join(root, "stdlib-framework-error")
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
  SourceLocation,
  Reporter,
  RunOptions,
  TestDescriptor,
  TestPlan,
  framework_error
}

fn discover_all(_request: DiscoveryRequest) -> Array TestDescriptor | Failure {
  [TestDescriptor {
    framework_id: "demo.framework",
    module_path: "pkg",
    test_id: "framework-1",
    display_name: "framework example",
    location: SourceLocation {
      module_path: "pkg/tests/framework.test.able",
      line: 2,
      column: 1
    },
    tags: Array.new(),
    metadata: Array.new()
  }]
}

fn run_plan(_plan: TestPlan, _options: RunOptions, reporter: Reporter) -> ?Failure {
  reporter.emit(framework_error { message: "broken harness" })
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
	writeFile(t, filepath.Join(stdlibSrc, "os.able"), `
package os

import able.kernel.{Array, __able_os_args, __able_os_exit}

fn args() -> Array String {
  __able_os_args()
}

fn exit(code: i32) -> void {
  __able_os_exit(code)
}
`)

	return stdlibSrc
}

func requireGoToolchain(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}
}

func configureMinimalTestCliEnv(t *testing.T, stdlibSrc string, kernelInModulePaths bool) {
	t.Helper()

	if kernelInModulePaths {
		t.Setenv("ABLE_MODULE_PATHS", stdlibSrc+string(os.PathListSeparator)+repoKernelPath(t))
	} else {
		t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
		t.Setenv("ABLE_PATH", repoKernelPath(t))
	}
	t.Setenv("ABLE_TYPECHECK_FIXTURES", "off")
}

func configureRepoCompiledEnv(t *testing.T) {
	t.Helper()

	requireGoToolchain(t)
	t.Setenv("ABLE_MODULE_PATHS", repoStdlibPath(t))
	t.Setenv("ABLE_PATH", repoKernelPath(t))
}

func runCLIExpectSuccess(t *testing.T, args ...string) string {
	t.Helper()

	code, stdout, stderr := captureCLI(t, args)
	if code != 0 {
		t.Fatalf("expected exit code 0 for %q, got %d (stderr=%q)", strings.Join(args, " "), code, stderr)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty for %q, got %q", strings.Join(args, " "), stderr)
	}
	return stdout
}

func runCLIExpectSuccessAllowingStderr(t *testing.T, args ...string) (string, string) {
	t.Helper()

	code, stdout, stderr := captureCLI(t, args)
	if code != 0 {
		t.Fatalf("expected exit code 0 for %q, got %d (stderr=%q)", strings.Join(args, " "), code, stderr)
	}
	return stdout, stderr
}

func runCLIExpectFailure(t *testing.T, args ...string) (int, string, string) {
	t.Helper()

	code, stdout, stderr := captureCLI(t, args)
	if code == 0 {
		t.Fatalf("expected non-zero exit code for %q", strings.Join(args, " "))
	}
	return code, stdout, stderr
}

func runCLIExpectFailureCode(t *testing.T, expectedCode int, args ...string) (string, string) {
	t.Helper()

	code, stdout, stderr := runCLIExpectFailure(t, args...)
	if code != expectedCode {
		t.Fatalf("expected exit code %d for %q, got %d (stderr=%q)", expectedCode, strings.Join(args, " "), code, stderr)
	}
	return stdout, stderr
}

func runCompiledStdlibTests(t *testing.T, label string, relPaths ...string) string {
	t.Helper()

	enterTempWorkingDir(t)

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	targets := make([]string, 0, len(relPaths)+2)
	targets = append(targets, "test", "--compiled")
	for _, relPath := range relPaths {
		targets = append(targets, filepath.Join(stdlibTests, filepath.FromSlash(relPath)))
	}

	configureRepoCompiledEnv(t)
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	stdout := runCLIExpectSuccess(t, targets...)
	return stdout
}

type compiledStdlibCase struct {
	label    string
	relPaths []string
	expected []string
}

func runCompiledStdlibCase(t *testing.T, tc compiledStdlibCase) {
	t.Helper()

	stdout := runCompiledStdlibTests(t, tc.label, tc.relPaths...)
	assertOutputContainsAll(t, stdout, tc.expected...)
}

func writeCompiledSampleTestModule(t *testing.T, root string) string {
	t.Helper()

	path := filepath.Join(root, "example.test.able")
	writeFile(t, path, `
package sample_tests

import able.spec.*

describe("math") { suite =>
  suite.it("adds") { _ctx =>
    expect(1 + 1).to(eq(2))
  }
}
`)
	return path
}

func runCompiledSampleTests(t *testing.T, target string) string {
	t.Helper()

	configureRepoCompiledEnv(t)
	return runCLIExpectSuccess(t, "test", "--compiled", target)
}

func assertCompiledSampleSuccess(t *testing.T, stdout string) {
	t.Helper()

	assertOutputContainsAll(t, stdout, "adds", "ok")
}

func assertTextContainsAll(t *testing.T, text string, substrings ...string) {
	t.Helper()

	for _, substring := range substrings {
		if !strings.Contains(text, substring) {
			t.Fatalf("expected text to contain %q, got %q", substring, text)
		}
	}
}

func assertAnyTextContainsAll(t *testing.T, texts []string, substrings ...string) {
	t.Helper()

	for _, text := range texts {
		matched := true
		for _, substring := range substrings {
			if !strings.Contains(text, substring) {
				matched = false
				break
			}
		}
		if matched {
			return
		}
	}
	t.Fatalf("expected at least one entry to contain %q, got %#v", strings.Join(substrings, ", "), texts)
}

func assertTextContainsAny(t *testing.T, text string, substrings ...string) {
	t.Helper()

	for _, substring := range substrings {
		if strings.Contains(text, substring) {
			return
		}
	}
	t.Fatalf("expected text to contain at least one of %#v, got %q", substrings, text)
}

func assertOutputContainsAll(t *testing.T, stdout string, substrings ...string) {
	t.Helper()

	assertTextContainsAll(t, stdout, substrings...)
}
