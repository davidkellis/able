package main

import (
	"able/interpreter-go/pkg/runtime"
	testclipkg "able/interpreter-go/pkg/testcli"
)

type TestReporterFormat string

const (
	reporterDoc      TestReporterFormat = "doc"
	reporterProgress TestReporterFormat = "progress"
	reporterTap      TestReporterFormat = "tap"
	reporterJSON     TestReporterFormat = "json"
)

type TestCliFilters struct {
	IncludePaths []string
	ExcludePaths []string
	IncludeNames []string
	ExcludeNames []string
	IncludeTags  []string
	ExcludeTags  []string
}

type TestRunOptions struct {
	FailFast    bool
	Repeat      int
	Parallelism int
	ShuffleSeed *int64
}

type TestCliConfig struct {
	Targets        []string
	Filters        TestCliFilters
	Run            TestRunOptions
	ReporterFormat TestReporterFormat
	ListOnly       bool
	DryRun         bool
	Compiled       bool
}

type TestEventState = testclipkg.EventState

type testCliModule struct {
	discoverAll      runtime.Value
	runPlan          runtime.Value
	docReporter      runtime.Value
	progressReporter runtime.Value
	cliReporter      runtime.Value
	cliComposite     runtime.Value
	discoveryDef     *runtime.StructDefinitionValue
	runOptionsDef    *runtime.StructDefinitionValue
	testPlanDef      *runtime.StructDefinitionValue
}

type testTypecheckMode int

const (
	testTypecheckOff testTypecheckMode = iota
	testTypecheckWarn
	testTypecheckStrict
)

type reporterBundle struct {
	reporter runtime.Value
	finish   func()
}

type harnessFailure struct {
	message string
	details *string
}

type testDescriptor = testclipkg.TestDescriptor
type metadataEntry = testclipkg.MetadataEntry
type sourceLocation = testclipkg.SourceLocation
type failureData = testclipkg.FailureData
type testEvent = testclipkg.TestEvent

const testCliModuleSource = `
package able_test_cli

import able.test.harness.{discover_all, run_plan}
import able.test.reporters.{DocReporter, ProgressReporter}
import able.test.protocol.{DiscoveryRequest, RunOptions, TestPlan, Reporter, TestEvent}

struct CliReporter { emit_fn: TestEvent -> void }
struct CliCompositeReporter { inner: Reporter, emit_fn: TestEvent -> void }

fn CliReporter(emit_fn: TestEvent -> void) -> CliReporter {
  CliReporter { emit_fn }
}

fn CliCompositeReporter(inner: Reporter, emit_fn: TestEvent -> void) -> CliCompositeReporter {
  CliCompositeReporter { inner, emit_fn }
}

impl Reporter for CliReporter {
  fn emit(self: Self, event: TestEvent) -> void {
    self.emit_fn(event)
  }
}

impl Reporter for CliCompositeReporter {
  fn emit(self: Self, event: TestEvent) -> void {
    self.inner.emit(event)
    self.emit_fn(event)
  }
}
`
