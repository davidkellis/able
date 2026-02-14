package main

import (
	"os"
	"os/exec"
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

func TestTestCommandCompiledRuns(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "example.test.able"), `
package sample_tests

import able.spec.*

describe("math") { suite =>
  suite.it("adds") { _ctx =>
    expect(1 + 1).to(eq(2))
  }
}
`)

	t.Setenv("ABLE_MODULE_PATHS", repoStdlibPath(t))
	t.Setenv("ABLE_PATH", repoKernelPath(t))

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", dir})
	if code != 0 {
		t.Fatalf("able test --compiled returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "adds") {
		t.Fatalf("expected stdout to include test name, got %q", stdout)
	}
	if !strings.Contains(stdout, "ok") {
		t.Fatalf("expected stdout to include ok, got %q", stdout)
	}
}

func TestTestCommandCompiledRunsStdlibBigintAndBiguintSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	bigintSpec := filepath.Join(stdlibTests, "bigint.test.able")
	biguintSpec := filepath.Join(stdlibTests, "biguint.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", bigintSpec, biguintSpec})
	if code != 0 {
		t.Fatalf("able test --compiled stdlib bigint/biguint returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "BigInt") {
		t.Fatalf("expected stdout to include BigInt suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "BigUint") {
		t.Fatalf("expected stdout to include BigUint suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "raises on underflow") {
		t.Fatalf("expected stdout to include BigUint underflow test, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibExtendedNumericSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	int128Spec := filepath.Join(stdlibTests, "int128.test.able")
	uint128Spec := filepath.Join(stdlibTests, "uint128.test.able")
	rationalSpec := filepath.Join(stdlibTests, "rational.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", int128Spec, uint128Spec, rationalSpec})
	if code != 0 {
		t.Fatalf("able test --compiled stdlib int128/uint128/rational returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "Int128") {
		t.Fatalf("expected stdout to include Int128 suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "UInt128") {
		t.Fatalf("expected stdout to include UInt128 suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "Rational") {
		t.Fatalf("expected stdout to include Rational suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "round-trips through display helpers") {
		t.Fatalf("expected stdout to include Rational display helper coverage, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibNumbersNumericSuite(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	numericSpec := filepath.Join(stdlibTests, "numbers_numeric.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", numericSpec})
	if code != 0 {
		t.Fatalf("able test --compiled stdlib numbers_numeric returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "Numeric primitives") {
		t.Fatalf("expected stdout to include numeric suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "covers f64 fractional helpers") {
		t.Fatalf("expected stdout to include fractional helper coverage, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibFoundationalSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	simpleSpec := filepath.Join(stdlibTests, "simple.test.able")
	assertionsSpec := filepath.Join(stdlibTests, "assertions.test.able")
	enumerableSpec := filepath.Join(stdlibTests, "enumerable.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", simpleSpec, assertionsSpec, enumerableSpec})
	if code != 0 {
		t.Fatalf("able test --compiled foundational stdlib suites returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "simple suite verifies addition works") {
		t.Fatalf("expected stdout to include simple suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "able.spec assertions passes equality matcher") {
		t.Fatalf("expected stdout to include assertions suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "Enumerable helpers maps and filters arrays") {
		t.Fatalf("expected stdout to include enumerable suite output, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibCollectionsListVectorSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	listSpec := filepath.Join(stdlibTests, "list.test.able")
	vectorSpec := filepath.Join(stdlibTests, "vector.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", listSpec, vectorSpec})
	if code != 0 {
		t.Fatalf("able test --compiled collections list/vector suites returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "List supports prepend/head/tail with structural sharing") {
		t.Fatalf("expected stdout to include list suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "Vector supports set without mutating prior versions") {
		t.Fatalf("expected stdout to include vector suite output, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibCollectionsTreeMapTreeSetSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	treeMapSpec := filepath.Join(stdlibTests, "tree_map.test.able")
	treeSetSpec := filepath.Join(stdlibTests, "tree_set.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", treeMapSpec, treeSetSpec})
	if code != 0 {
		t.Fatalf("able test --compiled collections tree_map/tree_set suites returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "TreeMap inserts, updates, and retrieves entries") {
		t.Fatalf("expected stdout to include tree_map suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "TreeSet inserts unique values and iterates in order") {
		t.Fatalf("expected stdout to include tree_set suite output, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibCollectionsPersistentMapPersistentSetSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	persistentMapSpec := filepath.Join(stdlibTests, "persistent_map.test.able")
	persistentSetSpec := filepath.Join(stdlibTests, "persistent_set.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", persistentMapSpec, persistentSetSpec})
	if code != 0 {
		t.Fatalf("able test --compiled collections persistent_map/persistent_set suites returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "PersistentMap stores, reads, and updates entries") {
		t.Fatalf("expected stdout to include persistent_map suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "PersistentSet unions and intersects") {
		t.Fatalf("expected stdout to include persistent_set suite output, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibCollectionsPersistentSortedSetAndQueueSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	persistentSortedSetSpec := filepath.Join(stdlibTests, "persistent_sorted_set.test.able")
	persistentQueueSpec := filepath.Join(stdlibTests, "persistent_queue.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", persistentSortedSetSpec, persistentQueueSpec})
	if code != 0 {
		t.Fatalf("able test --compiled collections persistent_sorted_set/persistent_queue suites returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "PersistentSortedSet keeps values ordered and unique") {
		t.Fatalf("expected stdout to include persistent_sorted_set suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "PersistentQueue iterates values in FIFO order") {
		t.Fatalf("expected stdout to include persistent_queue suite output, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibCollectionsLinkedListAndLazySeqSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	linkedListSpec := filepath.Join(stdlibTests, "linked_list.test.able")
	lazySeqSpec := filepath.Join(stdlibTests, "lazy_seq.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", linkedListSpec, lazySeqSpec})
	if code != 0 {
		t.Fatalf("able test --compiled collections linked_list/lazy_seq suites returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "LinkedList pushes and pops from both ends") {
		t.Fatalf("expected stdout to include linked_list suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "LazySeq iterates with caching and produces arrays") {
		t.Fatalf("expected stdout to include lazy_seq suite output, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibCollectionsHashMapSmokeAndHashSetSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	hashMapSmokeSpec := filepath.Join(stdlibTests, "collections", "hash_map_smoke.test.able")
	hashSetSpec := filepath.Join(stdlibTests, "collections", "hash_set.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", hashMapSmokeSpec, hashSetSpec})
	if code != 0 {
		t.Fatalf("able test --compiled collections hash_map_smoke/hash_set suites returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "HashSet adds, removes, and checks membership") {
		t.Fatalf("expected stdout to include hash_set suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "HashSet subset, superset, and disjoint checks") {
		t.Fatalf("expected stdout to include hash_set relation coverage, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibCollectionsDequeAndQueueSmokeSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	dequeSmokeSpec := filepath.Join(stdlibTests, "collections", "deque_smoke.test.able")
	queueSmokeSpec := filepath.Join(stdlibTests, "collections", "queue_smoke.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", dequeSmokeSpec, queueSmokeSpec})
	if code != 0 {
		t.Fatalf("able test --compiled collections deque_smoke/queue_smoke suites returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "able test: no tests to run") {
		t.Fatalf("expected stdout to report no tests for smoke suites, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibCollectionsBitSetAndHeapSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	bitSetSpec := filepath.Join(stdlibTests, "bit_set.test.able")
	heapSpec := filepath.Join(stdlibTests, "heap.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", bitSetSpec, heapSpec})
	if code != 0 {
		t.Fatalf("able test --compiled collections bit_set/heap suites returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "BitSet sets, checks, and resets bits") {
		t.Fatalf("expected stdout to include bit_set suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "Heap pushes and pops smallest values first") {
		t.Fatalf("expected stdout to include heap suite output, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibCollectionsArrayAndRangeSmokeSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	arraySmokeSpec := filepath.Join(stdlibTests, "collections", "array_smoke.test.able")
	rangeSmokeSpec := filepath.Join(stdlibTests, "collections", "range_smoke.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", arraySmokeSpec, rangeSmokeSpec})
	if code != 0 {
		t.Fatalf("able test --compiled collections array_smoke/range_smoke suites returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "able test: no tests to run") {
		t.Fatalf("expected stdout to report no tests for smoke suites, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibConcurrencyChannelMutexAndQueueSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	channelMutexSpec := filepath.Join(stdlibTests, "concurrency", "channel_mutex.test.able")
	concurrentQueueSpec := filepath.Join(stdlibTests, "concurrency", "concurrent_queue.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", channelMutexSpec, concurrentQueueSpec})
	if code != 0 {
		t.Fatalf("able test --compiled concurrency channel_mutex/concurrent_queue suites returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "Channel supports send/receive/close operations") {
		t.Fatalf("expected stdout to include channel_mutex suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "ConcurrentQueue supports try operations and close") {
		t.Fatalf("expected stdout to include concurrent_queue suite output, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibMathAndCoreNumericSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	mathSpec := filepath.Join(stdlibTests, "math.test.able")
	numericSmokeSpec := filepath.Join(stdlibTests, "core", "numeric_smoke.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", mathSpec, numericSmokeSpec})
	if code != 0 {
		t.Fatalf("able test --compiled math/core numeric suites returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "able.math computes gcd/lcm for integers") {
		t.Fatalf("expected stdout to include math suite output, got %q", stdout)
	}
	if !strings.Contains(stdout, "able.math offers rounding helpers") {
		t.Fatalf("expected stdout to include math rounding coverage, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibFsAndPathSmokeSuites(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	fsSmokeSpec := filepath.Join(stdlibTests, "fs_smoke.test.able")
	pathSmokeSpec := filepath.Join(stdlibTests, "path_smoke.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", fsSmokeSpec, pathSmokeSpec})
	if code != 0 {
		t.Fatalf("able test --compiled fs/path smoke suites returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "able test: no tests to run") {
		t.Fatalf("expected stdout to report no tests for smoke suites, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibIoSmokeSuite(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	ioSmokeSpec := filepath.Join(stdlibTests, "io_smoke.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", ioSmokeSpec})
	if code != 0 {
		t.Fatalf("able test --compiled io smoke suite returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "able test: no tests to run") {
		t.Fatalf("expected stdout to report no tests for smoke suite, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibOsSmokeSuite(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	osSmokeSpec := filepath.Join(stdlibTests, "os_smoke.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", osSmokeSpec})
	if code != 0 {
		t.Fatalf("able test --compiled os smoke suite returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "able test: no tests to run") {
		t.Fatalf("expected stdout to report no tests for smoke suite, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibProcessSmokeSuite(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	processSmokeSpec := filepath.Join(stdlibTests, "process_smoke.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", processSmokeSpec})
	if code != 0 {
		t.Fatalf("able test --compiled process smoke suite returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "able test: no tests to run") {
		t.Fatalf("expected stdout to report no tests for smoke suite, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibTermSmokeSuite(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	termSmokeSpec := filepath.Join(stdlibTests, "term_smoke.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", termSmokeSpec})
	if code != 0 {
		t.Fatalf("able test --compiled term smoke suite returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "able test: no tests to run") {
		t.Fatalf("expected stdout to report no tests for smoke suite, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRunsStdlibHarnessReportersSmokeSuite(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	stdlibSrc := repoStdlibPath(t)
	stdlibTests := filepath.Join(filepath.Dir(stdlibSrc), "tests")
	smokeSpec := filepath.Join(stdlibTests, "harness_reporters_smoke.test.able")

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")

	code, stdout, stderr := captureCLI(t, []string{"test", "--compiled", smokeSpec})
	if code != 0 {
		t.Fatalf("able test --compiled harness/reporters smoke suite returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "able test: no tests to run") {
		t.Fatalf("expected stdout to report no tests for smoke suite, got %q", stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestTestCommandCompiledRejectsInvalidRequireNoFallbacksEnv(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "example.test.able"), `
package sample_tests

import able.spec.*

describe("math") { suite =>
  suite.it("adds") { _ctx =>
    expect(1 + 1).to(eq(2))
  }
}
`)

	t.Setenv("ABLE_MODULE_PATHS", repoStdlibPath(t))
	t.Setenv("ABLE_PATH", repoKernelPath(t))
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "sometimes")

	code, _, stderr := captureCLI(t, []string{"test", "--compiled", dir})
	if code != 2 {
		t.Fatalf("able test --compiled expected exit code 2, got %d (stderr=%q)", code, stderr)
	}
	if !strings.Contains(stderr, "invalid ABLE_COMPILER_REQUIRE_NO_FALLBACKS value") {
		t.Fatalf("expected invalid env error, got %q", stderr)
	}
}
