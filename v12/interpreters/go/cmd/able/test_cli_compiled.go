package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"able/interpreter-go/pkg/compiler"
	"able/interpreter-go/pkg/driver"
)

func runCompiledTests(config TestCliConfig, testFiles []string) int {
	searchPaths, err := resolveTestSearchPaths(testFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: %v\n", err)
		return 2
	}
	packages, err := collectTestPackages(testFiles, searchPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: %v\n", err)
		return 2
	}

	moduleRoot, err := findGoModuleRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: %v\n", err)
		return 2
	}
	tmpRoot := filepath.Join(moduleRoot, "tmp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: %v\n", err)
		return 2
	}
	workDir, err := os.MkdirTemp(tmpRoot, "able-test-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: %v\n", err)
		return 2
	}
	keepWorkDir := os.Getenv("ABLE_TEST_KEEP_WORKDIR") != ""
	if !keepWorkDir {
		defer func() { _ = os.RemoveAll(workDir) }()
	}

	entryPath := filepath.Join(workDir, "runner.able")
	if err := os.WriteFile(entryPath, []byte(compiledTestRunnerSource(config)), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: write runner: %v\n", err)
		return 2
	}
	if keepWorkDir {
		fmt.Fprintf(os.Stderr, "able test --compiled: keeping workdir %s\n", workDir)
	}

	searchPaths = appendSearchPath(searchPaths, driver.SearchPath{Path: workDir, Kind: driver.RootUser})

	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: %v\n", err)
		return 2
	}
	defer loader.Close()

	loadOptions := driver.LoadOptions{IncludeTests: true}
	if len(packages) > 0 {
		loadOptions.IncludePackages = packages
	}
	program, err := loader.LoadWithOptions(entryPath, loadOptions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: load program: %v\n", err)
		return 2
	}

	requireNoFallbacks, err := resolveCompilerRequireNoFallbacksFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: %v\n", err)
		return 2
	}
	experimentalMonoArrays, err := resolveCompilerExperimentalMonoArraysFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: %v\n", err)
		return 2
	}
	result, err := compiler.New(compiler.Options{
		PackageName:               "main",
		RequireNoFallbacks:        requireNoFallbacks,
		ExperimentalMonoArrays:    experimentalMonoArrays,
		ExperimentalMonoArraysSet: true,
	}).Compile(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: compile: %v\n", err)
		return 2
	}
	if err := result.Write(workDir); err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: write output: %v\n", err)
		return 2
	}

	harness := compiledTestHarnessSource(config, entryPath, searchPaths, packages)
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte(harness), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: write harness: %v\n", err)
		return 2
	}

	binPath := filepath.Join(workDir, "able-test")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = workDir
	if output, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: go build failed: %v\n%s\n", err, string(output))
		return 2
	}

	cmd := exec.Command(binPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "able test --compiled: %v\n", err)
		return 2
	}
	return 0
}

func collectTestPackages(testFiles []string, searchPaths []driver.SearchPath) ([]string, error) {
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		return nil, err
	}
	defer loader.Close()

	seen := make(map[string]struct{})
	for _, file := range testFiles {
		program, err := loader.LoadWithOptions(file, driver.LoadOptions{IncludeTests: true})
		if err != nil {
			return nil, err
		}
		if program != nil && program.Entry != nil && program.Entry.Package != "" {
			seen[program.Entry.Package] = struct{}{}
		}
	}
	packages := make([]string, 0, len(seen))
	for pkg := range seen {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)
	return packages, nil
}

func appendSearchPath(paths []driver.SearchPath, sp driver.SearchPath) []driver.SearchPath {
	if sp.Path == "" {
		return paths
	}
	clean := filepath.Clean(sp.Path)
	for _, existing := range paths {
		if filepath.Clean(existing.Path) == clean {
			return paths
		}
	}
	return append(paths, sp)
}

func compiledTestRunnerSource(config TestCliConfig) string {
	var buf strings.Builder
	buf.WriteString("package compiled_tests\n")
	buf.WriteString("import able.kernel.{Array}\n")
	buf.WriteString("import able.test.harness.{discover_all, run_plan}\n")
	buf.WriteString("import able.test.protocol.{DiscoveryRequest, RunOptions, TestPlan, Reporter, TestEvent, Failure, TestDescriptor, SourceLocation, MetadataEntry, case_started, case_passed, case_failed, case_skipped, framework_error}\n")
	buf.WriteString("import able.test.reporters.{DocReporter, ProgressReporter}\n")
	buf.WriteString("import able.os\n")
	buf.WriteString("\n")
	buf.WriteString("extern go fn __able_test_cli_emit(kind: String, framework_id: String, module_path: String, test_id: String, display_name: String, tags: Array String, metadata_keys: Array String, metadata_values: Array String, descriptor_location_present: bool, descriptor_location_module: String, descriptor_location_line: i32, descriptor_location_column: i32, duration_ms: i64, reason_present: bool, reason: String, message: String, failure_message: String, failure_details_present: bool, failure_details: String, failure_location_present: bool, failure_location_module: String, failure_location_line: i32, failure_location_column: i32) -> void {}\n\n")

	buf.WriteString("fn metadata_keys(entries: Array MetadataEntry) -> Array String {\n")
	buf.WriteString("  keys: Array String := Array.new()\n")
	buf.WriteString("  idx := 0\n")
	buf.WriteString("  loop {\n")
	buf.WriteString("    if idx >= entries.len() { break }\n")
	buf.WriteString("    entries.get(idx) match {\n")
	buf.WriteString("      case nil => {},\n")
	buf.WriteString("      case entry: MetadataEntry => keys.push(entry.key)\n")
	buf.WriteString("    }\n")
	buf.WriteString("    idx = idx + 1\n")
	buf.WriteString("  }\n")
	buf.WriteString("  keys\n")
	buf.WriteString("}\n\n")

	buf.WriteString("fn metadata_values(entries: Array MetadataEntry) -> Array String {\n")
	buf.WriteString("  values: Array String := Array.new()\n")
	buf.WriteString("  idx := 0\n")
	buf.WriteString("  loop {\n")
	buf.WriteString("    if idx >= entries.len() { break }\n")
	buf.WriteString("    entries.get(idx) match {\n")
	buf.WriteString("      case nil => {},\n")
	buf.WriteString("      case entry: MetadataEntry => values.push(entry.value)\n")
	buf.WriteString("    }\n")
	buf.WriteString("    idx = idx + 1\n")
	buf.WriteString("  }\n")
	buf.WriteString("  values\n")
	buf.WriteString("}\n\n")

	buf.WriteString("struct CliReporter {}\n")
	buf.WriteString("fn CliReporter() -> CliReporter { CliReporter {} }\n")
	buf.WriteString("impl Reporter for CliReporter {\n")
	buf.WriteString("  fn emit(self: Self, event: TestEvent) -> void {\n")
	buf.WriteString("    event match {\n")
	buf.WriteString("      case case_failed { descriptor, duration_ms, failure } => {\n")
	buf.WriteString("        descriptor.location match {\n")
	buf.WriteString("          case nil => {\n")
	buf.WriteString("            failure.location match {\n")
	buf.WriteString("              case nil => __able_test_cli_emit(\"case_failed\", descriptor.framework_id, descriptor.module_path, descriptor.test_id, descriptor.display_name, descriptor.tags, metadata_keys(descriptor.metadata), metadata_values(descriptor.metadata), false, \"\", 0, 0, duration_ms, false, \"\", \"\", failure.message, failure.details != nil, failure.details match { case nil => \"\", case details: String => details }, false, \"\", 0, 0),\n")
	buf.WriteString("              case location: SourceLocation => __able_test_cli_emit(\"case_failed\", descriptor.framework_id, descriptor.module_path, descriptor.test_id, descriptor.display_name, descriptor.tags, metadata_keys(descriptor.metadata), metadata_values(descriptor.metadata), false, \"\", 0, 0, duration_ms, false, \"\", \"\", failure.message, failure.details != nil, failure.details match { case nil => \"\", case details: String => details }, true, location.module_path, location.line, location.column)\n")
	buf.WriteString("            }\n")
	buf.WriteString("          },\n")
	buf.WriteString("          case descriptor_location: SourceLocation => {\n")
	buf.WriteString("            failure.location match {\n")
	buf.WriteString("              case nil => __able_test_cli_emit(\"case_failed\", descriptor.framework_id, descriptor.module_path, descriptor.test_id, descriptor.display_name, descriptor.tags, metadata_keys(descriptor.metadata), metadata_values(descriptor.metadata), true, descriptor_location.module_path, descriptor_location.line, descriptor_location.column, duration_ms, false, \"\", \"\", failure.message, failure.details != nil, failure.details match { case nil => \"\", case details: String => details }, false, \"\", 0, 0),\n")
	buf.WriteString("              case location: SourceLocation => __able_test_cli_emit(\"case_failed\", descriptor.framework_id, descriptor.module_path, descriptor.test_id, descriptor.display_name, descriptor.tags, metadata_keys(descriptor.metadata), metadata_values(descriptor.metadata), true, descriptor_location.module_path, descriptor_location.line, descriptor_location.column, duration_ms, false, \"\", \"\", failure.message, failure.details != nil, failure.details match { case nil => \"\", case details: String => details }, true, location.module_path, location.line, location.column)\n")
	buf.WriteString("            }\n")
	buf.WriteString("          }\n")
	buf.WriteString("        }\n")
	buf.WriteString("      },\n")
	buf.WriteString("      case case_skipped { descriptor, reason } => {\n")
	buf.WriteString("        descriptor.location match {\n")
	buf.WriteString("          case nil => __able_test_cli_emit(\"case_skipped\", descriptor.framework_id, descriptor.module_path, descriptor.test_id, descriptor.display_name, descriptor.tags, metadata_keys(descriptor.metadata), metadata_values(descriptor.metadata), false, \"\", 0, 0, 0, reason != nil, reason match { case nil => \"\", case text: String => text }, \"\", \"\", false, \"\", false, \"\", 0, 0),\n")
	buf.WriteString("          case location: SourceLocation => __able_test_cli_emit(\"case_skipped\", descriptor.framework_id, descriptor.module_path, descriptor.test_id, descriptor.display_name, descriptor.tags, metadata_keys(descriptor.metadata), metadata_values(descriptor.metadata), true, location.module_path, location.line, location.column, 0, reason != nil, reason match { case nil => \"\", case text: String => text }, \"\", \"\", false, \"\", false, \"\", 0, 0)\n")
	buf.WriteString("        }\n")
	buf.WriteString("      },\n")
	buf.WriteString("      case framework_error { message } => {\n")
	buf.WriteString("        __able_test_cli_emit(\"framework_error\", \"\", \"\", \"\", \"\", Array.new(), Array.new(), Array.new(), false, \"\", 0, 0, 0, false, \"\", message, \"\", false, \"\", false, \"\", 0, 0)\n")
	buf.WriteString("      },\n")
	buf.WriteString("      case case_started { descriptor } => {\n")
	buf.WriteString("        descriptor.location match {\n")
	buf.WriteString("          case nil => __able_test_cli_emit(\"case_started\", descriptor.framework_id, descriptor.module_path, descriptor.test_id, descriptor.display_name, descriptor.tags, metadata_keys(descriptor.metadata), metadata_values(descriptor.metadata), false, \"\", 0, 0, 0, false, \"\", \"\", \"\", false, \"\", false, \"\", 0, 0),\n")
	buf.WriteString("          case location: SourceLocation => __able_test_cli_emit(\"case_started\", descriptor.framework_id, descriptor.module_path, descriptor.test_id, descriptor.display_name, descriptor.tags, metadata_keys(descriptor.metadata), metadata_values(descriptor.metadata), true, location.module_path, location.line, location.column, 0, false, \"\", \"\", \"\", false, \"\", false, \"\", 0, 0)\n")
	buf.WriteString("        }\n")
	buf.WriteString("      },\n")
	buf.WriteString("      case case_passed { descriptor, duration_ms } => {\n")
	buf.WriteString("        descriptor.location match {\n")
	buf.WriteString("          case nil => __able_test_cli_emit(\"case_passed\", descriptor.framework_id, descriptor.module_path, descriptor.test_id, descriptor.display_name, descriptor.tags, metadata_keys(descriptor.metadata), metadata_values(descriptor.metadata), false, \"\", 0, 0, duration_ms, false, \"\", \"\", \"\", false, \"\", false, \"\", 0, 0),\n")
	buf.WriteString("          case location: SourceLocation => __able_test_cli_emit(\"case_passed\", descriptor.framework_id, descriptor.module_path, descriptor.test_id, descriptor.display_name, descriptor.tags, metadata_keys(descriptor.metadata), metadata_values(descriptor.metadata), true, location.module_path, location.line, location.column, duration_ms, false, \"\", \"\", \"\", false, \"\", false, \"\", 0, 0)\n")
	buf.WriteString("        }\n")
	buf.WriteString("      }\n")
	buf.WriteString("    }\n")
	buf.WriteString("  }\n")
	buf.WriteString("}\n\n")

	buf.WriteString("struct CountingReporter { inner: Reporter, failed: i32, skipped: i32, framework_errors: i32 }\n")
	buf.WriteString("fn CountingReporter(inner: Reporter) -> CountingReporter { CountingReporter { inner, failed: 0, skipped: 0, framework_errors: 0 } }\n")
	buf.WriteString("impl Reporter for CountingReporter {\n")
	buf.WriteString("  fn emit(self: Self, event: TestEvent) -> void {\n")
	buf.WriteString("    self.inner.emit(event)\n")
	buf.WriteString("    event match {\n")
	buf.WriteString("      case case_failed { descriptor::_, duration_ms::_, failure::_ } => { self.failed = self.failed + 1 },\n")
	buf.WriteString("      case case_skipped { descriptor::_, reason::_ } => { self.skipped = self.skipped + 1 },\n")
	buf.WriteString("      case framework_error { message::_ } => { self.framework_errors = self.framework_errors + 1 },\n")
	buf.WriteString("      case _ => {}\n")
	buf.WriteString("    }\n")
	buf.WriteString("  }\n")
	buf.WriteString("}\n\n")

	buf.WriteString("fn main() -> void {\n")
	buf.WriteString(fmt.Sprintf("  request := DiscoveryRequest { include_paths: %s, exclude_paths: %s, include_names: %s, exclude_names: %s, include_tags: %s, exclude_tags: %s, list_only: %s }\n",
		ableStringArrayLiteral(config.Filters.IncludePaths),
		ableStringArrayLiteral(config.Filters.ExcludePaths),
		ableStringArrayLiteral(config.Filters.IncludeNames),
		ableStringArrayLiteral(config.Filters.ExcludeNames),
		ableStringArrayLiteral(config.Filters.IncludeTags),
		ableStringArrayLiteral(config.Filters.ExcludeTags),
		ableBoolLiteral(false),
	))
	buf.WriteString(fmt.Sprintf("  options := RunOptions { shuffle_seed: %s, fail_fast: %s, parallelism: %d, repeat: %d }\n",
		ableOptionalIntLiteral(config.Run.ShuffleSeed),
		ableBoolLiteral(config.Run.FailFast),
		config.Run.Parallelism,
		config.Run.Repeat,
	))
	buf.WriteString("  discover_all(request) match {\n")
	buf.WriteString("    case failure: Failure => {\n")
	buf.WriteString("      print(`framework error: ${failure.message}`)\n")
	buf.WriteString("      os.exit(2)\n")
	buf.WriteString("    },\n")
	buf.WriteString("    case descriptors: Array TestDescriptor => {\n")
	buf.WriteString("      if descriptors.len() == 0 {\n")
	buf.WriteString("        print(\"able test: no tests to run\")\n")
	buf.WriteString("        os.exit(0)\n")
	buf.WriteString("      }\n")
	buf.WriteString("      reporter: Reporter := DocReporter({ line => print(line) })\n")
	if config.ReporterFormat == reporterJSON || config.ReporterFormat == reporterTap {
		buf.WriteString("      reporter = CliReporter()\n")
	} else if config.ReporterFormat == reporterProgress {
		buf.WriteString("      progress := ProgressReporter({ line => print(line) })\n")
		buf.WriteString("      reporter = progress\n")
		buf.WriteString("      counter := CountingReporter(reporter)\n")
		buf.WriteString("      failure := run_plan(TestPlan { descriptors }, options, counter)\n")
		buf.WriteString("      progress.finish()\n")
		buf.WriteString("      failure match {\n")
		buf.WriteString("        case nil => {},\n")
		buf.WriteString("        case err: Failure => { print(`framework error: ${err.message}`); os.exit(2) }\n")
		buf.WriteString("      }\n")
		buf.WriteString("      if counter.framework_errors > 0 { os.exit(2) }\n")
		buf.WriteString("      if counter.failed > 0 { os.exit(1) }\n")
		buf.WriteString("      os.exit(0)\n")
		buf.WriteString("    }\n")
		buf.WriteString("  }\n")
		buf.WriteString("}\n")
		return buf.String()
	}

	buf.WriteString("      counter := CountingReporter(reporter)\n")
	buf.WriteString("      failure := run_plan(TestPlan { descriptors }, options, counter)\n")
	buf.WriteString("      failure match {\n")
	buf.WriteString("        case nil => {},\n")
	buf.WriteString("        case err: Failure => { print(`framework error: ${err.message}`); os.exit(2) }\n")
	buf.WriteString("      }\n")
	buf.WriteString("      if counter.framework_errors > 0 { os.exit(2) }\n")
	buf.WriteString("      if counter.failed > 0 { os.exit(1) }\n")
	buf.WriteString("      os.exit(0)\n")
	buf.WriteString("    }\n")
	buf.WriteString("  }\n")
	buf.WriteString("}\n")
	return buf.String()
}

func compiledTestHarnessSource(config TestCliConfig, entryPath string, searchPaths []driver.SearchPath, packages []string) string {
	var buf strings.Builder
	_ = entryPath
	_ = searchPaths
	_ = packages
	buf.WriteString("package main\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"fmt\"\n")
	buf.WriteString("\t\"os\"\n")
	buf.WriteString("\t\"strings\"\n")
	buf.WriteString("\t\"able/interpreter-go/pkg/interpreter\"\n")
	buf.WriteString("\t\"able/interpreter-go/pkg/testcli\"\n")
	buf.WriteString("\t\"able/interpreter-go/pkg/runtime\"\n")
	buf.WriteString(")\n\n")
	buf.WriteString("func main() {\n")
	buf.WriteString("\tinterp := interpreter.New()\n")
	buf.WriteString("\tinterp.SetArgs(os.Args[1:])\n")
	buf.WriteString("\tentryEnv := runtime.NewEnvironment(nil)\n")
	buf.WriteString("\tregisterPrintInEnv(entryEnv, interp)\n")
	buf.WriteString("\tregisterOSBuiltinsInEnv(entryEnv, os.Args[1:])\n")
	buf.WriteString("\tregisterTestReporterBuiltinsInEnv(entryEnv, interp)\n")
	buf.WriteString("\trt, err := RegisterIn(interp, entryEnv)\n")
	buf.WriteString("\tif err != nil {\n")
	buf.WriteString("\t\tif code, ok := interpreter.ExitCodeFromError(err); ok {\n\t\t\tos.Exit(code)\n\t\t}\n")
	buf.WriteString("\t\tfmt.Fprintln(os.Stderr, err)\n")
	buf.WriteString("\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("\tif err := RunRegisteredMain(rt, interp, entryEnv); err != nil {\n")
	buf.WriteString("\t\tif code, ok := interpreter.ExitCodeFromError(err); ok {\n\t\t\tos.Exit(code)\n\t\t}\n")
	buf.WriteString("\t\tfmt.Fprintln(os.Stderr, err)\n")
	buf.WriteString("\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("}\n\n")
	buf.WriteString("func registerPrintInEnv(env *runtime.Environment, interp *interpreter.Interpreter) {\n")
	buf.WriteString("\tif env == nil {\n\t\treturn\n\t}\n")
	buf.WriteString("\tprintFn := runtime.NativeFunctionValue{\n")
	buf.WriteString("\t\tName:  \"print\",\n\t\tArity: 1,\n")
	buf.WriteString("\t\tImpl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
	buf.WriteString("\t\t\tvar parts []string\n")
	buf.WriteString("\t\t\tfor _, arg := range args {\n")
	buf.WriteString("\t\t\t\trendered, err := interp.Stringify(arg, nil)\n")
	buf.WriteString("\t\t\t\tif err != nil || rendered == \"\" {\n\t\t\t\t\trendered = fmt.Sprintf(\"<%s>\", arg.Kind())\n\t\t\t\t}\n")
	buf.WriteString("\t\t\t\tparts = append(parts, rendered)\n\t\t\t}\n")
	buf.WriteString("\t\t\tfmt.Fprintln(os.Stdout, strings.Join(parts, \" \"))\n")
	buf.WriteString("\t\t\treturn runtime.VoidValue{}, nil\n\t\t},\n\t}\n")
	buf.WriteString("\tenv.Define(\"print\", printFn)\n}\n\n")
	buf.WriteString("func registerOSBuiltinsInEnv(env *runtime.Environment, osArgs []string) {\n")
	buf.WriteString("\tif env == nil {\n\t\treturn\n\t}\n")
	buf.WriteString("\tif _, err := env.Get(\"__able_os_args\"); err != nil {\n")
	buf.WriteString("\t\tenv.Define(\"__able_os_args\", runtime.NativeFunctionValue{\n")
	buf.WriteString("\t\t\tName:  \"__able_os_args\",\n\t\t\tArity: 0,\n")
	buf.WriteString("\t\t\tImpl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
	buf.WriteString("\t\t\t\tif len(args) != 0 {\n\t\t\t\t\treturn nil, fmt.Errorf(\"__able_os_args expects no arguments\")\n\t\t\t\t}\n")
	buf.WriteString("\t\t\t\tvalues := make([]runtime.Value, 0, len(osArgs))\n")
	buf.WriteString("\t\t\t\tfor _, arg := range osArgs {\n")
	buf.WriteString("\t\t\t\t\tvalues = append(values, runtime.StringValue{Val: arg})\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t\treturn &runtime.ArrayValue{Elements: values}, nil\n")
	buf.WriteString("\t\t\t},\n\t\t})\n\t}\n")
	buf.WriteString("\tif _, err := env.Get(\"__able_os_exit\"); err != nil {\n")
	buf.WriteString("\t\tenv.Define(\"__able_os_exit\", runtime.NativeFunctionValue{\n")
	buf.WriteString("\t\t\tName:  \"__able_os_exit\",\n\t\t\tArity: 1,\n")
	buf.WriteString("\t\t\tImpl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
	buf.WriteString("\t\t\t\tif len(args) != 1 {\n\t\t\t\t\treturn nil, fmt.Errorf(\"__able_os_exit expects one argument\")\n\t\t\t\t}\n")
	buf.WriteString("\t\t\t\tvar code int64\n")
	buf.WriteString("\t\t\t\tswitch typed := args[0].(type) {\n")
	buf.WriteString("\t\t\t\tcase runtime.IntegerValue:\n")
	buf.WriteString("\t\t\t\t\tif n, ok := typed.ToInt64(); ok { code = n } else { code = typed.BigInt().Int64() }\n")
	buf.WriteString("\t\t\t\tcase *runtime.IntegerValue:\n")
	buf.WriteString("\t\t\t\t\tif typed == nil { return nil, fmt.Errorf(\"__able_os_exit expects integer argument\") }\n")
	buf.WriteString("\t\t\t\t\tif n, ok := typed.ToInt64(); ok { code = n } else { code = typed.BigInt().Int64() }\n")
	buf.WriteString("\t\t\t\tdefault:\n")
	buf.WriteString("\t\t\t\t\treturn nil, fmt.Errorf(\"__able_os_exit expects integer argument\")\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t\tif code < 0 { return nil, fmt.Errorf(\"exit code must be non-negative\") }\n")
	buf.WriteString("\t\t\t\tif code > int64(^uint(0)>>1) { return nil, fmt.Errorf(\"exit code is out of range\") }\n")
	buf.WriteString("\t\t\t\tos.Exit(int(code))\n")
	buf.WriteString("\t\t\t\treturn runtime.VoidValue{}, nil\n")
	buf.WriteString("\t\t\t},\n\t\t})\n\t}\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")
	buf.WriteString("func registerTestReporterBuiltinsInEnv(env *runtime.Environment, interp *interpreter.Interpreter) {\n")
	buf.WriteString("\tif env == nil {\n\t\treturn\n\t}\n")
	buf.WriteString(fmt.Sprintf("\treporterFormat := testcli.ReporterFormat(%q)\n", config.ReporterFormat))
	buf.WriteString("\temitter := testcli.NewEventEmitter(reporterFormat, os.Stdout, nil)\n")
	buf.WriteString("\tif reporterFormat == testcli.ReporterTap {\n")
	buf.WriteString("\t\temitter.EmitHeader()\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tenv.Define(\"__able_test_cli_emit\", runtime.NativeFunctionValue{\n")
	buf.WriteString("\t\tName:  \"__able_test_cli_emit\",\n")
	buf.WriteString("\t\tArity: 23,\n")
	buf.WriteString("\t\tImpl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
	buf.WriteString("\t\t\tif len(args) != 23 {\n")
	buf.WriteString("\t\t\t\treturn nil, fmt.Errorf(\"__able_test_cli_emit expects twenty-three arguments\")\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tevent, err := decodeCompiledTestEventArgs(interp, args)\n")
	buf.WriteString("\t\t\tif err != nil {\n")
	buf.WriteString("\t\t\t\treturn nil, err\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tif err := emitter.Emit(event); err != nil {\n")
	buf.WriteString("\t\t\t\tfmt.Fprintf(os.Stderr, \"able test: %v\\n\", err)\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn runtime.VoidValue{}, nil\n")
	buf.WriteString("\t\t},\n")
	buf.WriteString("\t})\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")
	buf.WriteString("func decodeCompiledTestEventArgs(interp *interpreter.Interpreter, args []runtime.Value) (*testcli.TestEvent, error) {\n")
	buf.WriteString("\tkind, err := compiledTestStringArg(args[0], \"kind\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tframeworkID, err := compiledTestStringArg(args[1], \"framework_id\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tmodulePath, err := compiledTestStringArg(args[2], \"module_path\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\ttestID, err := compiledTestStringArg(args[3], \"test_id\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tdisplayName, err := compiledTestStringArg(args[4], \"display_name\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\ttags, err := compiledTestStringArrayArg(interp, args[5], \"tags\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tmetadataKeys, err := compiledTestStringArrayArg(interp, args[6], \"metadata_keys\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tmetadataValues, err := compiledTestStringArrayArg(interp, args[7], \"metadata_values\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tdescriptorLocationPresent, err := compiledTestBoolArg(args[8], \"descriptor_location_present\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tdescriptorLocationModule, err := compiledTestStringArg(args[9], \"descriptor_location_module\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tdescriptorLocationLine, err := compiledTestInt64Arg(args[10], \"descriptor_location_line\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tdescriptorLocationColumn, err := compiledTestInt64Arg(args[11], \"descriptor_location_column\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tdurationMs, err := compiledTestInt64Arg(args[12], \"duration_ms\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\treasonPresent, err := compiledTestBoolArg(args[13], \"reason_present\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\treasonText, err := compiledTestStringArg(args[14], \"reason\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tmessage, err := compiledTestStringArg(args[15], \"message\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tfailureMessage, err := compiledTestStringArg(args[16], \"failure_message\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tfailureDetailsPresent, err := compiledTestBoolArg(args[17], \"failure_details_present\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tfailureDetailsText, err := compiledTestStringArg(args[18], \"failure_details\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tfailureLocationPresent, err := compiledTestBoolArg(args[19], \"failure_location_present\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tfailureLocationModule, err := compiledTestStringArg(args[20], \"failure_location_module\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tfailureLocationLine, err := compiledTestInt64Arg(args[21], \"failure_location_line\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tfailureLocationColumn, err := compiledTestInt64Arg(args[22], \"failure_location_column\")\n")
	buf.WriteString("\tif err != nil { return nil, err }\n")
	buf.WriteString("\tvar descriptor *testcli.TestDescriptor\n")
	buf.WriteString("\tif frameworkID != \"\" || modulePath != \"\" || testID != \"\" || displayName != \"\" {\n")
	buf.WriteString("\t\tdescriptor = &testcli.TestDescriptor{FrameworkID: frameworkID, ModulePath: modulePath, TestID: testID, DisplayName: displayName, Tags: tags, Metadata: compiledTestMetadataEntries(metadataKeys, metadataValues)}\n")
	buf.WriteString("\t\tif descriptorLocationPresent {\n")
	buf.WriteString("\t\t\tdescriptor.Location = &testcli.SourceLocation{ModulePath: descriptorLocationModule, Line: int(descriptorLocationLine), Column: int(descriptorLocationColumn)}\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tvar reason *string\n")
	buf.WriteString("\tif reasonPresent {\n")
	buf.WriteString("\t\treason = &reasonText\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tvar failure *testcli.FailureData\n")
	buf.WriteString("\tif failureMessage != \"\" || failureDetailsPresent || failureLocationPresent {\n")
	buf.WriteString("\t\tfailure = &testcli.FailureData{Message: failureMessage}\n")
	buf.WriteString("\t\tif failureDetailsPresent {\n")
	buf.WriteString("\t\t\tfailure.Details = &failureDetailsText\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tif failureLocationPresent {\n")
	buf.WriteString("\t\t\tfailure.Location = &testcli.SourceLocation{ModulePath: failureLocationModule, Line: int(failureLocationLine), Column: int(failureLocationColumn)}\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\treturn &testcli.TestEvent{Kind: kind, Descriptor: descriptor, DurationMs: durationMs, Failure: failure, Reason: reason, Message: message}, nil\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")
	buf.WriteString("func compiledTestStringArrayArg(interp *interpreter.Interpreter, value runtime.Value, label string) ([]string, error) {\n")
	buf.WriteString("\tif value == nil {\n")
	buf.WriteString("\t\treturn nil, fmt.Errorf(\"missing %s value\", label)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tvar arrayVal *runtime.ArrayValue\n")
	buf.WriteString("\tswitch typed := value.(type) {\n")
	buf.WriteString("\tcase *runtime.ArrayValue:\n")
	buf.WriteString("\t\tarrayVal = typed\n")
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\tif interp == nil {\n")
	buf.WriteString("\t\t\treturn nil, fmt.Errorf(\"%s must be Array String\", label)\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tcoerced, err := interp.CoerceArrayValue(value)\n")
	buf.WriteString("\t\tif err != nil {\n")
	buf.WriteString("\t\t\treturn nil, fmt.Errorf(\"%s must be Array String\", label)\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tarrayVal = coerced\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tif arrayVal == nil {\n")
	buf.WriteString("\t\treturn nil, fmt.Errorf(\"missing %s value\", label)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tout := make([]string, 0, len(arrayVal.Elements))\n")
	buf.WriteString("\tfor _, entry := range arrayVal.Elements {\n")
	buf.WriteString("\t\ttext, err := compiledTestStringArg(entry, label+\" entry\")\n")
	buf.WriteString("\t\tif err != nil {\n")
	buf.WriteString("\t\t\treturn nil, err\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tout = append(out, text)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\treturn out, nil\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")
	buf.WriteString("func compiledTestMetadataEntries(keys []string, values []string) []testcli.MetadataEntry {\n")
	buf.WriteString("\tif len(keys) == 0 && len(values) == 0 {\n")
	buf.WriteString("\t\treturn nil\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tcount := len(keys)\n")
	buf.WriteString("\tif len(values) > count {\n")
	buf.WriteString("\t\tcount = len(values)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tout := make([]testcli.MetadataEntry, 0, count)\n")
	buf.WriteString("\tfor idx := 0; idx < count; idx++ {\n")
	buf.WriteString("\t\tkey := \"\"\n")
	buf.WriteString("\t\tif idx < len(keys) { key = keys[idx] }\n")
	buf.WriteString("\t\tvalue := \"\"\n")
	buf.WriteString("\t\tif idx < len(values) { value = values[idx] }\n")
	buf.WriteString("\t\tout = append(out, testcli.MetadataEntry{Key: key, Value: value})\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\treturn out\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")
	buf.WriteString("func compiledTestStringArg(value runtime.Value, label string) (string, error) {\n")
	buf.WriteString("\tswitch typed := value.(type) {\n")
	buf.WriteString("\tcase runtime.StringValue:\n")
	buf.WriteString("\t\treturn typed.Val, nil\n")
	buf.WriteString("\tcase *runtime.StringValue:\n")
	buf.WriteString("\t\tif typed == nil { return \"\", fmt.Errorf(\"missing %s value\", label) }\n")
	buf.WriteString("\t\treturn typed.Val, nil\n")
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\treturn \"\", fmt.Errorf(\"%s must be String\", label)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")
	buf.WriteString("func compiledTestBoolArg(value runtime.Value, label string) (bool, error) {\n")
	buf.WriteString("\tswitch typed := value.(type) {\n")
	buf.WriteString("\tcase runtime.BoolValue:\n")
	buf.WriteString("\t\treturn typed.Val, nil\n")
	buf.WriteString("\tcase *runtime.BoolValue:\n")
	buf.WriteString("\t\tif typed == nil { return false, fmt.Errorf(\"missing %s value\", label) }\n")
	buf.WriteString("\t\treturn typed.Val, nil\n")
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\treturn false, fmt.Errorf(\"%s must be bool\", label)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")
	buf.WriteString("func compiledTestInt64Arg(value runtime.Value, label string) (int64, error) {\n")
	buf.WriteString("\tswitch typed := value.(type) {\n")
	buf.WriteString("\tcase runtime.IntegerValue:\n")
	buf.WriteString("\t\tif n, ok := typed.ToInt64(); ok { return n, nil }\n")
	buf.WriteString("\t\treturn typed.BigInt().Int64(), nil\n")
	buf.WriteString("\tcase *runtime.IntegerValue:\n")
	buf.WriteString("\t\tif typed == nil { return 0, fmt.Errorf(\"missing %s value\", label) }\n")
	buf.WriteString("\t\tif n, ok := typed.ToInt64(); ok { return n, nil }\n")
	buf.WriteString("\t\treturn typed.BigInt().Int64(), nil\n")
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\treturn 0, fmt.Errorf(\"%s must be integer\", label)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
	return buf.String()
}

func ableStringArrayLiteral(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, ableStringLiteral(value))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func ableStringLiteral(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	return "\"" + replacer.Replace(value) + "\""
}

func goStringArrayLiteral(values []string) string {
	if len(values) == 0 {
		return "nil"
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("%q", value))
	}
	return "[]string{" + strings.Join(parts, ", ") + "}"
}

func ableBoolLiteral(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func ableOptionalIntLiteral(value *int64) string {
	if value == nil {
		return "nil"
	}
	return fmt.Sprintf("%d_i64", *value)
}

func searchPathKindLiteral(kind driver.RootKind) string {
	switch kind {
	case driver.RootStdlib:
		return "driver.RootStdlib"
	default:
		return "driver.RootUser"
	}
}
