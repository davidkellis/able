package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/parser"
	"able/interpreter-go/pkg/runtime"
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
}

type TestEventState struct {
	Total           int
	Failed          int
	Skipped         int
	FrameworkErrors int
}

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

type testDescriptor struct {
	FrameworkID string          `json:"framework_id"`
	ModulePath  string          `json:"module_path"`
	TestID      string          `json:"test_id"`
	DisplayName string          `json:"display_name"`
	Tags        []string        `json:"tags"`
	Metadata    []metadataEntry `json:"metadata"`
	Location    *sourceLocation `json:"location"`
}

type metadataEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type sourceLocation struct {
	ModulePath string `json:"module_path"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
}

type failureData struct {
	Message  string          `json:"message"`
	Details  *string         `json:"details"`
	Location *sourceLocation `json:"location"`
}

type testEvent struct {
	Kind       string          `json:"event"`
	Descriptor *testDescriptor `json:"descriptor,omitempty"`
	DurationMs int64           `json:"duration_ms,omitempty"`
	Failure    *failureData    `json:"failure,omitempty"`
	Reason     *string         `json:"reason,omitempty"`
	Message    string          `json:"message,omitempty"`
}

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

func runTest(args []string) int {
	config, err := parseTestArguments(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 1
	}

	targets, err := resolveTestTargets(config.Targets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 1
	}

	testFiles, err := collectTestFiles(targets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 1
	}

	if len(testFiles) == 0 {
		fmt.Fprintln(os.Stdout, "able test: no test modules found")
		return 0
	}

	loadResult, err := loadTestPrograms(testFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 2
	}

	mode := resolveTestTypecheckMode()
	if ok, code := typecheckTestModules(loadResult.modules, mode); !ok {
		return code
	}

	interp := interpreter.New()
	registerPrint(interp)

	if ok, code := evaluateTestModules(interp, loadResult.modules); !ok {
		return code
	}

	cliModule, err := loadTestCliModule(interp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 2
	}

	discoveryRequest, err := buildDiscoveryRequest(interp, cliModule, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 2
	}

	descriptors, ok := callHarnessDiscover(interp, cliModule, discoveryRequest)
	if !ok {
		return 2
	}

	if config.ListOnly || config.DryRun {
		emitTestPlanList(interp, descriptors, config)
		return 0
	}

	if arrayLength(interp, descriptors) == 0 {
		fmt.Fprintln(os.Stdout, "able test: no tests to run")
		return 0
	}

	state := &TestEventState{}
	reporter, err := createTestReporter(interp, cliModule, config.ReporterFormat, state)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 2
	}

	runOptions, err := buildRunOptions(interp, cliModule, config.Run)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 2
	}
	testPlan, err := buildTestPlan(cliModule, descriptors)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return 2
	}

	if callHarnessRun(interp, cliModule, testPlan, runOptions, reporter.reporter) != nil {
		return 2
	}

	if reporter.finish != nil {
		reporter.finish()
	}

	if state.FrameworkErrors > 0 {
		return 2
	}
	if state.Failed > 0 {
		return 1
	}
	return 0
}

func parseTestArguments(args []string) (TestCliConfig, error) {
	filters := TestCliFilters{
		IncludePaths: []string{},
		ExcludePaths: []string{},
		IncludeNames: []string{},
		ExcludeNames: []string{},
		IncludeTags:  []string{},
		ExcludeTags:  []string{},
	}
	run := TestRunOptions{
		FailFast:    false,
		Repeat:      1,
		Parallelism: 1,
	}
	format := reporterDoc
	listOnly := false
	dryRun := false
	var shuffleSeed *int64
	var targets []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--list":
			listOnly = true
		case "--dry-run":
			dryRun = true
			listOnly = true
		case "--path":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			filters.IncludePaths = append(filters.IncludePaths, val)
		case "--exclude-path":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			filters.ExcludePaths = append(filters.ExcludePaths, val)
		case "--name":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			filters.IncludeNames = append(filters.IncludeNames, val)
		case "--exclude-name":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			filters.ExcludeNames = append(filters.ExcludeNames, val)
		case "--tag":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			filters.IncludeTags = append(filters.IncludeTags, val)
		case "--exclude-tag":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			filters.ExcludeTags = append(filters.ExcludeTags, val)
		case "--format":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			parsed, err := parseReporterFormat(val)
			if err != nil {
				return TestCliConfig{}, err
			}
			format = parsed
		case "--fail-fast":
			run.FailFast = true
		case "--repeat":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			count, err := parsePositiveInt(val, arg, 1)
			if err != nil {
				return TestCliConfig{}, err
			}
			run.Repeat = count
		case "--parallel":
			val, err := expectFlagValue(arg, nextArg(args, &i))
			if err != nil {
				return TestCliConfig{}, err
			}
			count, err := parsePositiveInt(val, arg, 1)
			if err != nil {
				return TestCliConfig{}, err
			}
			run.Parallelism = count
		case "--shuffle":
			next := peekArg(args, i+1)
			if next != "" && !strings.HasPrefix(next, "-") {
				seed, err := parsePositiveInt(next, arg, 0)
				if err != nil {
					return TestCliConfig{}, err
				}
				seedVal := int64(seed)
				shuffleSeed = &seedVal
				i++
			} else {
				seed := generateShuffleSeed()
				shuffleSeed = &seed
			}
		default:
			if strings.HasPrefix(arg, "-") {
				return TestCliConfig{}, fmt.Errorf("unknown able test flag '%s'", arg)
			}
			targets = append(targets, arg)
		}
	}

	run.ShuffleSeed = shuffleSeed

	return TestCliConfig{
		Targets:        targets,
		Filters:        filters,
		Run:            run,
		ReporterFormat: format,
		ListOnly:       listOnly,
		DryRun:         dryRun,
	}, nil
}

func nextArg(args []string, index *int) string {
	*index = *index + 1
	if *index >= len(args) {
		return ""
	}
	return args[*index]
}

func peekArg(args []string, index int) string {
	if index < 0 || index >= len(args) {
		return ""
	}
	return args[index]
}

func expectFlagValue(flag string, value string) (string, error) {
	if value == "" || strings.HasPrefix(value, "-") {
		return "", fmt.Errorf("%s expects a value", flag)
	}
	return value, nil
}

func parseReporterFormat(value string) (TestReporterFormat, error) {
	switch value {
	case "doc":
		return reporterDoc, nil
	case "progress":
		return reporterProgress, nil
	case "tap":
		return reporterTap, nil
	case "json":
		return reporterJSON, nil
	default:
		return "", fmt.Errorf("unknown --format value '%s' (expected doc, progress, tap, or json)", value)
	}
}

func parsePositiveInt(value string, flag string, min int) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < min {
		return 0, fmt.Errorf("%s expects an integer >= %d", flag, min)
	}
	return parsed, nil
}

func generateShuffleSeed() int64 {
	now := time.Now().UnixMilli()
	str := fmt.Sprintf("%d", now)
	if len(str) > 9 {
		str = str[len(str)-9:]
	}
	parsed, _ := strconv.ParseInt(str, 10, 64)
	return parsed
}

func resolveTestTargets(targets []string) ([]string, error) {
	rawTargets := targets
	if len(rawTargets) == 0 {
		rawTargets = []string{"."}
	}
	seen := make(map[string]struct{})
	var resolved []string
	for _, target := range rawTargets {
		abs, err := filepath.Abs(target)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve %s: %w", target, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("unable to access %s: %w", abs, err)
		}
		if info.IsDir() {
			if _, ok := seen[abs]; !ok {
				seen[abs] = struct{}{}
				resolved = append(resolved, abs)
			}
			continue
		}
		if info.Mode().IsRegular() {
			if isTestFile(abs) {
				if _, ok := seen[abs]; !ok {
					seen[abs] = struct{}{}
					resolved = append(resolved, abs)
				}
			} else {
				dir := filepath.Dir(abs)
				if _, ok := seen[dir]; !ok {
					seen[dir] = struct{}{}
					resolved = append(resolved, dir)
				}
			}
			continue
		}
		return nil, fmt.Errorf("unsupported test target: %s", abs)
	}
	return resolved, nil
}

func collectTestFiles(targets []string) ([]string, error) {
	found := make(map[string]struct{})
	for _, target := range targets {
		info, err := os.Stat(target)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			if err := walkTestFiles(target, found); err != nil {
				return nil, err
			}
			continue
		}
		if isTestFile(target) {
			found[filepath.Clean(target)] = struct{}{}
		}
	}
	files := make([]string, 0, len(found))
	for file := range found {
		files = append(files, file)
	}
	sort.Strings(files)
	return files, nil
}

func walkTestFiles(root string, found map[string]struct{}) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case "quarantine", "node_modules", ".git":
				return fs.SkipDir
			default:
				return nil
			}
		}
		if d.Type().IsRegular() && isTestFile(d.Name()) {
			found[filepath.Clean(path)] = struct{}{}
		}
		return nil
	})
}

func isTestFile(path string) bool {
	return strings.HasSuffix(path, ".test.able") || strings.HasSuffix(path, ".spec.able")
}

type testLoadResult struct {
	programs []*driver.Program
	modules  []*driver.Module
}

func loadTestPrograms(testFiles []string) (*testLoadResult, error) {
	searchPaths, err := resolveTestSearchPaths(testFiles)
	if err != nil {
		return nil, err
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		return nil, err
	}
	defer loader.Close()

	include := []string{
		"able.test.harness",
		"able.test.protocol",
		"able.test.reporters",
	}
	programs := make([]*driver.Program, 0, len(testFiles))
	for _, file := range testFiles {
		program, err := loader.LoadWithOptions(file, driver.LoadOptions{IncludePackages: include})
		if err != nil {
			return nil, fmt.Errorf("failed to load tests from %s: %w", file, err)
		}
		programs = append(programs, program)
	}
	return &testLoadResult{
		programs: programs,
		modules:  mergeTestModules(programs),
	}, nil
}

func resolveTestSearchPaths(testFiles []string) ([]driver.SearchPath, error) {
	if len(testFiles) == 0 {
		return collectSearchPaths(""), nil
	}
	seen := make(map[string]struct{})
	var merged []driver.SearchPath
	for _, file := range testFiles {
		base := filepath.Dir(file)
		manifest, err := loadManifestFrom(base)
		if err != nil && !errors.Is(err, errManifestNotFound) {
			return nil, err
		}
		var lock *driver.Lockfile
		if manifest != nil {
			lock, err = loadLockfileForManifest(manifest)
			if err != nil {
				return nil, err
			}
		}
		extras, err := buildExecutionSearchPaths(manifest, lock)
		if err != nil {
			return nil, err
		}
		paths := collectSearchPaths(base, extras...)
		for _, sp := range paths {
			clean := filepath.Clean(sp.Path)
			if _, ok := seen[clean]; ok {
				continue
			}
			seen[clean] = struct{}{}
			merged = append(merged, sp)
		}
	}
	return merged, nil
}

func mergeTestModules(programs []*driver.Program) []*driver.Module {
	seen := make(map[string]struct{})
	var modules []*driver.Module
	for _, program := range programs {
		if program == nil {
			continue
		}
		for _, mod := range program.Modules {
			if mod == nil {
				continue
			}
			if _, ok := seen[mod.Package]; ok {
				continue
			}
			seen[mod.Package] = struct{}{}
			modules = append(modules, mod)
		}
	}
	return modules
}

func resolveTestTypecheckMode() testTypecheckMode {
	raw, ok := os.LookupEnv("ABLE_TYPECHECK_FIXTURES")
	if !ok {
		return testTypecheckStrict
	}
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" || normalized == "0" || normalized == "off" || normalized == "false" {
		return testTypecheckOff
	}
	if normalized == "strict" || normalized == "fail" || normalized == "error" || normalized == "1" || normalized == "true" {
		return testTypecheckStrict
	}
	if normalized == "warn" || normalized == "warning" {
		return testTypecheckWarn
	}
	return testTypecheckWarn
}

func typecheckTestModules(modules []*driver.Module, mode testTypecheckMode) (bool, int) {
	if mode == testTypecheckOff {
		return true, 0
	}
	if len(modules) == 0 {
		return true, 0
	}
	program := &driver.Program{Entry: modules[0], Modules: modules}
	result, err := interpreter.TypecheckProgram(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "typecheck error: %v\n", err)
		return false, 2
	}
	diagnostics := filterDiagnostics(result.Diagnostics)
	if len(diagnostics) == 0 {
		return true, 0
	}
	emitDiagnostics(diagnostics)
	printPackageSummaries(os.Stderr, result.Packages)
	if mode == testTypecheckStrict {
		return false, 2
	}
	fmt.Fprintln(os.Stderr, "typechecker: proceeding despite diagnostics because ABLE_TYPECHECK_FIXTURES=warn")
	return true, 0
}

func filterDiagnostics(diags []interpreter.ModuleDiagnostic) []interpreter.ModuleDiagnostic {
	out := make([]interpreter.ModuleDiagnostic, 0, len(diags))
	for _, diag := range diags {
		if isStdlibPackage(diag.Package) {
			continue
		}
		out = append(out, diag)
	}
	return out
}

func isStdlibPackage(name string) bool {
	if name == "able" {
		return true
	}
	return strings.HasPrefix(name, "able.")
}

func emitDiagnostics(diags []interpreter.ModuleDiagnostic) {
	seen := make(map[string]struct{})
	for _, diag := range diags {
		msg := interpreter.DescribeModuleDiagnostic(diag)
		if _, ok := seen[msg]; ok {
			continue
		}
		seen[msg] = struct{}{}
		fmt.Fprintln(os.Stderr, msg)
	}
}

func evaluateTestModules(interp *interpreter.Interpreter, modules []*driver.Module) (bool, int) {
	for _, mod := range modules {
		if mod == nil || mod.AST == nil {
			continue
		}
		if _, _, err := interp.EvaluateModule(mod.AST); err != nil {
			if code, ok := interpreter.ExitCodeFromError(err); ok {
				return false, code
			}
			fmt.Fprintf(os.Stderr, "runtime error: %v\n", err)
			return false, 2
		}
	}
	return true, 0
}

func loadTestCliModule(interp *interpreter.Interpreter) (*testCliModule, error) {
	parser, err := parser.NewModuleParser()
	if err != nil {
		return nil, err
	}
	defer parser.Close()
	moduleAst, err := parser.ParseModule([]byte(testCliModuleSource))
	if err != nil {
		return nil, err
	}
	_, env, err := interp.EvaluateModule(moduleAst)
	if err != nil {
		return nil, err
	}
	discoverAll, err := env.Get("discover_all")
	if err != nil {
		return nil, fmt.Errorf("missing discover_all: %w", err)
	}
	runPlan, err := env.Get("run_plan")
	if err != nil {
		return nil, fmt.Errorf("missing run_plan: %w", err)
	}
	docReporter, err := env.Get("DocReporter")
	if err != nil {
		return nil, fmt.Errorf("missing DocReporter: %w", err)
	}
	progressReporter, err := env.Get("ProgressReporter")
	if err != nil {
		return nil, fmt.Errorf("missing ProgressReporter: %w", err)
	}
	cliReporter, err := env.Get("CliReporter")
	if err != nil {
		return nil, fmt.Errorf("missing CliReporter: %w", err)
	}
	cliComposite, err := env.Get("CliCompositeReporter")
	if err != nil {
		return nil, fmt.Errorf("missing CliCompositeReporter: %w", err)
	}
	discoveryDef, err := getStructDef(env, "DiscoveryRequest")
	if err != nil {
		return nil, err
	}
	runOptionsDef, err := getStructDef(env, "RunOptions")
	if err != nil {
		return nil, err
	}
	testPlanDef, err := getStructDef(env, "TestPlan")
	if err != nil {
		return nil, err
	}

	return &testCliModule{
		discoverAll:      discoverAll,
		runPlan:          runPlan,
		docReporter:      docReporter,
		progressReporter: progressReporter,
		cliReporter:      cliReporter,
		cliComposite:     cliComposite,
		discoveryDef:     discoveryDef,
		runOptionsDef:    runOptionsDef,
		testPlanDef:      testPlanDef,
	}, nil
}

func getStructDef(env *runtime.Environment, name string) (*runtime.StructDefinitionValue, error) {
	value, err := env.Get(name)
	if err != nil {
		return nil, fmt.Errorf("missing %s: %w", name, err)
	}
	switch v := value.(type) {
	case runtime.StructDefinitionValue:
		return &v, nil
	case *runtime.StructDefinitionValue:
		if v == nil {
			return nil, fmt.Errorf("missing %s struct definition", name)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("expected %s struct definition", name)
	}
}

func buildDiscoveryRequest(interp *interpreter.Interpreter, cli *testCliModule, config TestCliConfig) (runtime.Value, error) {
	if cli == nil || cli.discoveryDef == nil {
		return nil, fmt.Errorf("missing DiscoveryRequest definition")
	}
	includePaths, _ := makeStringArray(interp, config.Filters.IncludePaths)
	excludePaths, _ := makeStringArray(interp, config.Filters.ExcludePaths)
	includeNames, _ := makeStringArray(interp, config.Filters.IncludeNames)
	excludeNames, _ := makeStringArray(interp, config.Filters.ExcludeNames)
	includeTags, _ := makeStringArray(interp, config.Filters.IncludeTags)
	excludeTags, _ := makeStringArray(interp, config.Filters.ExcludeTags)

	fields := map[string]runtime.Value{
		"include_paths": includePaths,
		"exclude_paths": excludePaths,
		"include_names": includeNames,
		"exclude_names": excludeNames,
		"include_tags":  includeTags,
		"exclude_tags":  excludeTags,
		"list_only":     runtime.BoolValue{Val: config.ListOnly},
	}
	return &runtime.StructInstanceValue{Definition: cli.discoveryDef, Fields: fields}, nil
}

func buildRunOptions(interp *interpreter.Interpreter, cli *testCliModule, run TestRunOptions) (runtime.Value, error) {
	if cli == nil || cli.runOptionsDef == nil {
		return nil, fmt.Errorf("missing RunOptions definition")
	}
	var shuffle runtime.Value = runtime.NilValue{}
	if run.ShuffleSeed != nil {
		shuffle = makeIntegerValue(runtime.IntegerI64, *run.ShuffleSeed)
	}
	fields := map[string]runtime.Value{
		"shuffle_seed": shuffle,
		"fail_fast":    runtime.BoolValue{Val: run.FailFast},
		"parallelism":  makeIntegerValue(runtime.IntegerI32, int64(run.Parallelism)),
		"repeat":       makeIntegerValue(runtime.IntegerI32, int64(run.Repeat)),
	}
	return &runtime.StructInstanceValue{Definition: cli.runOptionsDef, Fields: fields}, nil
}

func buildTestPlan(cli *testCliModule, descriptors runtime.Value) (runtime.Value, error) {
	if cli == nil || cli.testPlanDef == nil {
		return nil, fmt.Errorf("missing TestPlan definition")
	}
	fields := map[string]runtime.Value{
		"descriptors": descriptors,
	}
	return &runtime.StructInstanceValue{Definition: cli.testPlanDef, Fields: fields}, nil
}

func callHarnessDiscover(interp *interpreter.Interpreter, cli *testCliModule, request runtime.Value) (runtime.Value, bool) {
	if cli == nil {
		fmt.Fprintln(os.Stderr, "able test: missing CLI module")
		return nil, false
	}
	result, err := interp.CallFunction(cli.discoverAll, []runtime.Value{request})
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return nil, false
	}
	if failure := extractFailure(interp, result); failure != nil {
		fmt.Fprintf(os.Stderr, "able test: %s\n", formatFailure(failure))
		return nil, false
	}
	if _, err := coerceArrayValue(interp, result, "discovery result"); err != nil {
		fmt.Fprintln(os.Stderr, "able test: discovery returned unexpected result")
		return nil, false
	}
	return result, true
}

func callHarnessRun(
	interp *interpreter.Interpreter,
	cli *testCliModule,
	plan runtime.Value,
	options runtime.Value,
	reporter runtime.Value,
) *harnessFailure {
	if cli == nil {
		fmt.Fprintln(os.Stderr, "able test: missing CLI module")
		return &harnessFailure{message: "missing CLI module"}
	}
	result, err := interp.CallFunction(cli.runPlan, []runtime.Value{plan, options, reporter})
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return &harnessFailure{message: err.Error()}
	}
	if isNilValue(result) {
		return nil
	}
	if failure := extractFailure(interp, result); failure != nil {
		fmt.Fprintf(os.Stderr, "able test: %s\n", formatFailure(failure))
		return failure
	}
	fmt.Fprintln(os.Stderr, "able test: run_plan returned unexpected result")
	return &harnessFailure{message: "run_plan returned unexpected result"}
}

func createTestReporter(
	interp *interpreter.Interpreter,
	cli *testCliModule,
	format TestReporterFormat,
	state *TestEventState,
) (*reporterBundle, error) {
	if cli == nil {
		return nil, fmt.Errorf("missing CLI module")
	}
	emit := createEventHandler(interp, format, state)
	emitFn := runtime.NativeFunctionValue{
		Name:  "__able_test_cli_emit",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) > 0 && args[0] != nil {
				emit(args[0])
			}
			return runtime.NilValue{}, nil
		},
	}

	if format == reporterJSON || format == reporterTap {
		if format == reporterTap {
			fmt.Fprintln(os.Stdout, "TAP version 13")
		}
		reporter, err := interp.CallFunction(cli.cliReporter, []runtime.Value{emitFn})
		if err != nil {
			return nil, err
		}
		return &reporterBundle{reporter: reporter}, nil
	}

	inner, err := createStdlibReporter(interp, cli, format)
	if err != nil {
		return nil, err
	}
	reporter, err := interp.CallFunction(cli.cliComposite, []runtime.Value{inner, emitFn})
	if err != nil {
		return nil, err
	}
	var finish func()
	if format == reporterProgress {
		finish = func() { finishProgressReporter(interp, inner) }
	}
	return &reporterBundle{reporter: reporter, finish: finish}, nil
}

func createStdlibReporter(
	interp *interpreter.Interpreter,
	cli *testCliModule,
	format TestReporterFormat,
) (runtime.Value, error) {
	writeLine := runtime.NativeFunctionValue{
		Name:  "__able_test_cli_line",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) == 0 || args[0] == nil {
				return runtime.NilValue{}, nil
			}
			fmt.Fprintln(os.Stdout, runtimeValueToString(args[0]))
			return runtime.NilValue{}, nil
		},
	}
	var reporterFn runtime.Value
	switch format {
	case reporterDoc:
		reporterFn = cli.docReporter
	case reporterProgress:
		reporterFn = cli.progressReporter
	default:
		return nil, fmt.Errorf("unsupported reporter format %q", format)
	}
	return interp.CallFunction(reporterFn, []runtime.Value{writeLine})
}

func finishProgressReporter(interp *interpreter.Interpreter, reporter runtime.Value) {
	method, err := interp.LookupStructMethod(reporter, "finish")
	if err != nil || method == nil {
		return
	}
	if _, err := interp.CallFunction(method, []runtime.Value{reporter}); err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
	}
}

func emitTestPlanList(interp *interpreter.Interpreter, descriptors runtime.Value, config TestCliConfig) {
	items, err := decodeDescriptorArray(interp, descriptors)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test: %v\n", err)
		return
	}
	if config.ReporterFormat == reporterJSON {
		payload, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "able test: %v\n", err)
			return
		}
		fmt.Fprintln(os.Stdout, string(payload))
		return
	}
	if len(items) == 0 {
		fmt.Fprintln(os.Stdout, "able test: no tests found")
		return
	}
	for _, item := range items {
		tags := "-"
		if len(item.Tags) > 0 {
			tags = strings.Join(item.Tags, ",")
		}
		modulePath := item.ModulePath
		if modulePath == "" {
			modulePath = "-"
		}
		parts := []string{
			item.FrameworkID,
			modulePath,
			item.TestID,
			item.DisplayName,
			fmt.Sprintf("tags=%s", tags),
		}
		if config.DryRun {
			parts = append(parts, fmt.Sprintf("metadata=%s", formatMetadata(item.Metadata)))
		}
		fmt.Fprintln(os.Stdout, strings.Join(parts, " | "))
	}
}

func createEventHandler(
	interp *interpreter.Interpreter,
	format TestReporterFormat,
	state *TestEventState,
) func(runtime.Value) {
	tapIndex := 0
	return func(event runtime.Value) {
		decoded, err := decodeTestEvent(interp, event)
		if err != nil || decoded == nil {
			if err != nil {
				fmt.Fprintf(os.Stderr, "able test: %v\n", err)
			}
			return
		}
		recordTestEvent(state, decoded)

		switch format {
		case reporterJSON:
			payload, err := json.Marshal(decoded)
			if err != nil {
				fmt.Fprintf(os.Stderr, "able test: %v\n", err)
				return
			}
			fmt.Fprintln(os.Stdout, string(payload))
		case reporterTap:
			switch decoded.Kind {
			case "case_passed":
				tapIndex++
				fmt.Fprintf(os.Stdout, "ok %d - %s\n", tapIndex, decoded.Descriptor.DisplayName)
			case "case_failed":
				tapIndex++
				fmt.Fprintf(os.Stdout, "not ok %d - %s\n", tapIndex, decoded.Descriptor.DisplayName)
				emitTapFailure(decoded.Failure)
			case "case_skipped":
				tapIndex++
				reason := "skipped"
				if decoded.Reason != nil {
					reason = *decoded.Reason
				}
				fmt.Fprintf(os.Stdout, "ok %d - %s # SKIP %s\n", tapIndex, decoded.Descriptor.DisplayName, reason)
			case "framework_error":
				fmt.Fprintf(os.Stdout, "Bail out! %s\n", decoded.Message)
			}
		default:
		}
	}
}

func emitTapFailure(failure *failureData) {
	if failure == nil {
		return
	}
	lines := []string{
		"  ---",
		fmt.Sprintf("  message: %s", sanitizeTapValue(failure.Message)),
	}
	if failure.Details != nil {
		lines = append(lines, fmt.Sprintf("  details: %s", sanitizeTapValue(*failure.Details)))
	}
	if failure.Location != nil {
		lines = append(lines, fmt.Sprintf(
			"  location: %s",
			sanitizeTapValue(fmt.Sprintf("%s:%d:%d", failure.Location.ModulePath, failure.Location.Line, failure.Location.Column)),
		))
	}
	lines = append(lines, "  ...")
	for _, line := range lines {
		fmt.Fprintln(os.Stdout, line)
	}
}

func sanitizeTapValue(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\\n"), "\n", "\\n")
}

func recordTestEvent(state *TestEventState, event *testEvent) {
	if state == nil || event == nil {
		return
	}
	switch event.Kind {
	case "case_passed":
		state.Total++
	case "case_failed":
		state.Total++
		state.Failed++
	case "case_skipped":
		state.Total++
		state.Skipped++
	case "framework_error":
		state.FrameworkErrors++
	}
}

func decodeTestEvent(interp *interpreter.Interpreter, value runtime.Value) (*testEvent, error) {
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return nil, nil
	}
	switch structTag(inst) {
	case "case_started":
		descriptor, err := decodeDescriptor(interp, structField(inst, "descriptor"))
		if err != nil {
			return nil, err
		}
		return &testEvent{Kind: "case_started", Descriptor: descriptor}, nil
	case "case_passed":
		descriptor, err := decodeDescriptor(interp, structField(inst, "descriptor"))
		if err != nil {
			return nil, err
		}
		duration := decodeNumber(structField(inst, "duration_ms"))
		return &testEvent{Kind: "case_passed", Descriptor: descriptor, DurationMs: duration}, nil
	case "case_failed":
		descriptor, err := decodeDescriptor(interp, structField(inst, "descriptor"))
		if err != nil {
			return nil, err
		}
		duration := decodeNumber(structField(inst, "duration_ms"))
		failure, err := decodeFailure(interp, structField(inst, "failure"))
		if err != nil {
			return nil, err
		}
		return &testEvent{Kind: "case_failed", Descriptor: descriptor, DurationMs: duration, Failure: failure}, nil
	case "case_skipped":
		descriptor, err := decodeDescriptor(interp, structField(inst, "descriptor"))
		if err != nil {
			return nil, err
		}
		reason := decodeOptionalString(interp, structField(inst, "reason"))
		return &testEvent{Kind: "case_skipped", Descriptor: descriptor, Reason: reason}, nil
	case "framework_error":
		message := decodeString(interp, structField(inst, "message"))
		return &testEvent{Kind: "framework_error", Message: message}, nil
	default:
		return nil, nil
	}
}

func decodeDescriptorArray(interp *interpreter.Interpreter, value runtime.Value) ([]testDescriptor, error) {
	arrayVal, err := coerceArrayValue(interp, value, "descriptor array")
	if err != nil {
		return nil, err
	}
	out := make([]testDescriptor, 0, len(arrayVal.Elements))
	for _, entry := range arrayVal.Elements {
		desc, err := decodeDescriptor(interp, entry)
		if err != nil {
			return nil, err
		}
		out = append(out, *desc)
	}
	return out, nil
}

func decodeDescriptor(interp *interpreter.Interpreter, value runtime.Value) (*testDescriptor, error) {
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return nil, fmt.Errorf("expected TestDescriptor struct")
	}
	return &testDescriptor{
		FrameworkID: decodeString(interp, structField(inst, "framework_id")),
		ModulePath:  decodeString(interp, structField(inst, "module_path")),
		TestID:      decodeString(interp, structField(inst, "test_id")),
		DisplayName: decodeString(interp, structField(inst, "display_name")),
		Tags:        decodeStringArray(interp, structField(inst, "tags")),
		Metadata:    decodeMetadataArray(interp, structField(inst, "metadata")),
		Location:    decodeLocation(interp, structField(inst, "location")),
	}, nil
}

func decodeFailure(interp *interpreter.Interpreter, value runtime.Value) (*failureData, error) {
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return nil, fmt.Errorf("expected Failure struct")
	}
	return &failureData{
		Message:  decodeString(interp, structField(inst, "message")),
		Details:  decodeOptionalString(interp, structField(inst, "details")),
		Location: decodeLocation(interp, structField(inst, "location")),
	}, nil
}

func decodeLocation(interp *interpreter.Interpreter, value runtime.Value) *sourceLocation {
	if isNilValue(value) {
		return nil
	}
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return nil
	}
	return &sourceLocation{
		ModulePath: decodeString(interp, structField(inst, "module_path")),
		Line:       int(decodeNumber(structField(inst, "line"))),
		Column:     int(decodeNumber(structField(inst, "column"))),
	}
}

func decodeMetadataArray(interp *interpreter.Interpreter, value runtime.Value) []metadataEntry {
	if isNilValue(value) {
		return nil
	}
	arrayVal, err := coerceArrayValue(interp, value, "metadata array")
	if err != nil {
		return nil
	}
	out := make([]metadataEntry, 0, len(arrayVal.Elements))
	for _, entry := range arrayVal.Elements {
		inst, ok := entry.(*runtime.StructInstanceValue)
		if !ok || inst == nil {
			continue
		}
		out = append(out, metadataEntry{
			Key:   decodeString(interp, structField(inst, "key")),
			Value: decodeString(interp, structField(inst, "value")),
		})
	}
	return out
}

func decodeStringArray(interp *interpreter.Interpreter, value runtime.Value) []string {
	if isNilValue(value) {
		return nil
	}
	arrayVal, err := coerceArrayValue(interp, value, "string array")
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(arrayVal.Elements))
	for _, entry := range arrayVal.Elements {
		out = append(out, decodeString(interp, entry))
	}
	return out
}

func decodeString(_ *interpreter.Interpreter, value runtime.Value) string {
	switch v := value.(type) {
	case runtime.StringValue:
		return v.Val
	case *runtime.StringValue:
		if v != nil {
			return v.Val
		}
	}
	return runtimeValueToString(value)
}

func decodeOptionalString(interp *interpreter.Interpreter, value runtime.Value) *string {
	if isNilValue(value) {
		return nil
	}
	decoded := decodeString(interp, value)
	return &decoded
}

func decodeNumber(value runtime.Value) int64 {
	switch v := value.(type) {
	case runtime.IntegerValue:
		if v.Val != nil {
			return v.Val.Int64()
		}
	case *runtime.IntegerValue:
		if v != nil && v.Val != nil {
			return v.Val.Int64()
		}
	}
	return 0
}

func coerceArrayValue(interp *interpreter.Interpreter, value runtime.Value, label string) (*runtime.ArrayValue, error) {
	if value == nil {
		return nil, fmt.Errorf("expected %s", label)
	}
	switch v := value.(type) {
	case *runtime.ArrayValue:
		return v, nil
	default:
		if interp == nil {
			return nil, fmt.Errorf("expected %s", label)
		}
		return interp.CoerceArrayValue(value)
	}
}

func arrayLength(interp *interpreter.Interpreter, value runtime.Value) int {
	arr, err := coerceArrayValue(interp, value, "array")
	if err != nil || arr == nil {
		return 0
	}
	return len(arr.Elements)
}

func structField(inst *runtime.StructInstanceValue, name string) runtime.Value {
	if inst == nil || name == "" {
		return runtime.NilValue{}
	}
	if inst.Fields != nil {
		if value, ok := inst.Fields[name]; ok && value != nil {
			return value
		}
	}
	if inst.Definition != nil && inst.Definition.Node != nil && len(inst.Positional) > 0 {
		for idx, field := range inst.Definition.Node.Fields {
			if field != nil && field.Name != nil && field.Name.Name == name {
				if idx < len(inst.Positional) {
					return inst.Positional[idx]
				}
				break
			}
		}
	}
	return runtime.NilValue{}
}

func structTag(inst *runtime.StructInstanceValue) string {
	if inst == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
		return ""
	}
	return inst.Definition.Node.ID.Name
}

func makeStringArray(interp *interpreter.Interpreter, values []string) (runtime.Value, error) {
	elems := make([]runtime.Value, 0, len(values))
	for _, val := range values {
		elems = append(elems, runtime.StringValue{Val: val})
	}
	arr := &runtime.ArrayValue{Elements: elems}
	if interp == nil {
		return arr, nil
	}
	coerced, err := interp.CoerceArrayValue(arr)
	if err != nil {
		return arr, nil
	}
	return coerced, nil
}

func makeIntegerValue(suffix runtime.IntegerType, value int64) runtime.IntegerValue {
	return runtime.IntegerValue{Val: big.NewInt(value), TypeSuffix: suffix}
}

func extractFailure(interp *interpreter.Interpreter, value runtime.Value) *harnessFailure {
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return nil
	}
	if structTag(inst) != "Failure" {
		return nil
	}
	message := decodeString(interp, structField(inst, "message"))
	details := decodeOptionalString(interp, structField(inst, "details"))
	return &harnessFailure{message: message, details: details}
}

func formatFailure(failure *harnessFailure) string {
	if failure == nil {
		return ""
	}
	if failure.details == nil {
		return failure.message
	}
	return fmt.Sprintf("%s (%s)", failure.message, *failure.details)
}

func isNilValue(value runtime.Value) bool {
	switch value.(type) {
	case nil:
		return true
	case runtime.NilValue:
		return true
	case *runtime.NilValue:
		return true
	default:
		return false
	}
}

func runtimeValueToString(value runtime.Value) string {
	if value == nil {
		return ""
	}
	return formatRuntimeValue(value)
}

func formatMetadata(entries []metadataEntry) string {
	if len(entries) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		parts = append(parts, fmt.Sprintf("%s=%s", entry.Key, entry.Value))
	}
	return strings.Join(parts, ",")
}
