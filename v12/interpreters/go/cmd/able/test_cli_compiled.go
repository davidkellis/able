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
	if config.ReporterFormat == reporterJSON || config.ReporterFormat == reporterTap {
		fmt.Fprintf(os.Stderr, "able test --compiled: reporter format %q is not supported (use doc or progress)\n", config.ReporterFormat)
		return 1
	}

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
		PackageName:            "main",
		RequireNoFallbacks:     requireNoFallbacks,
		ExperimentalMonoArrays: experimentalMonoArrays,
	}).Compile(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: compile: %v\n", err)
		return 2
	}
	if err := result.Write(workDir); err != nil {
		fmt.Fprintf(os.Stderr, "able test --compiled: write output: %v\n", err)
		return 2
	}

	harness := compiledTestHarnessSource(entryPath, searchPaths, packages)
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
	buf.WriteString("import able.test.protocol.{DiscoveryRequest, RunOptions, TestPlan, Reporter, TestEvent, Failure, TestDescriptor, case_failed, case_skipped, framework_error}\n")
	buf.WriteString("import able.test.reporters.{DocReporter, ProgressReporter}\n")
	buf.WriteString("import able.os\n")
	buf.WriteString("\n")

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

	buf.WriteString("fn list_descriptors(descriptors: Array TestDescriptor) -> void {\n")
	buf.WriteString("  idx := 0\n")
	buf.WriteString("  loop {\n")
	buf.WriteString("    if idx >= descriptors.len() { break }\n")
	buf.WriteString("    descriptors.get(idx) match {\n")
	buf.WriteString("      case nil => {},\n")
	buf.WriteString("      case descriptor: TestDescriptor => { print(descriptor.display_name) }\n")
	buf.WriteString("    }\n")
	buf.WriteString("    idx = idx + 1\n")
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
		ableBoolLiteral(config.ListOnly || config.DryRun),
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
	buf.WriteString("      if request.list_only {\n")
	buf.WriteString("        list_descriptors(descriptors)\n")
	buf.WriteString("        os.exit(0)\n")
	buf.WriteString("      }\n")
	buf.WriteString("      if descriptors.len() == 0 {\n")
	buf.WriteString("        print(\"able test: no tests to run\")\n")
	buf.WriteString("        os.exit(0)\n")
	buf.WriteString("      }\n")
	buf.WriteString("      reporter: Reporter := DocReporter({ line => print(line) })\n")
	if config.ReporterFormat == reporterProgress {
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

func compiledTestHarnessSource(entryPath string, searchPaths []driver.SearchPath, packages []string) string {
	var buf strings.Builder
	buf.WriteString("package main\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"fmt\"\n")
	buf.WriteString("\t\"os\"\n")
	buf.WriteString("\t\"strings\"\n")
	buf.WriteString("\t\"able/interpreter-go/pkg/driver\"\n")
	buf.WriteString("\t\"able/interpreter-go/pkg/interpreter\"\n")
	buf.WriteString("\t\"able/interpreter-go/pkg/runtime\"\n")
	buf.WriteString(")\n\n")
	buf.WriteString("func main() {\n")
	buf.WriteString("\tsearchPaths := []driver.SearchPath{\n")
	for _, sp := range searchPaths {
		buf.WriteString(fmt.Sprintf("\t\t{Path: %q, Kind: %s},\n", sp.Path, searchPathKindLiteral(sp.Kind)))
	}
	buf.WriteString("\t}\n")
	buf.WriteString(fmt.Sprintf("\tentry := %q\n", entryPath))
	if len(packages) > 0 {
		buf.WriteString(fmt.Sprintf("\tincludePackages := %s\n", goStringArrayLiteral(packages)))
	}
	buf.WriteString("\tloader, err := driver.NewLoader(searchPaths)\n")
	buf.WriteString("\tif err != nil {\n\t\tfmt.Fprintln(os.Stderr, err)\n\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("\tdefer loader.Close()\n")
	if len(packages) > 0 {
		buf.WriteString("\tprogram, err := loader.LoadWithOptions(entry, driver.LoadOptions{IncludeTests: true, IncludePackages: includePackages})\n")
	} else {
		buf.WriteString("\tprogram, err := loader.LoadWithOptions(entry, driver.LoadOptions{IncludeTests: true})\n")
	}
	buf.WriteString("\tif err != nil {\n\t\tfmt.Fprintln(os.Stderr, err)\n\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("\tinterp := interpreter.New()\n")
	buf.WriteString("\tinterp.SetArgs(os.Args[1:])\n")
	buf.WriteString("\tregisterPrint(interp)\n")
	buf.WriteString("\tmode := resolveTestTypecheckMode()\n")
	buf.WriteString("\t_, entryEnv, _, err := interp.EvaluateProgram(program, interpreter.ProgramEvaluationOptions{SkipTypecheck: mode == typecheckModeOff, AllowDiagnostics: mode != typecheckModeOff})\n")
	buf.WriteString("\tif err != nil {\n")
	buf.WriteString("\t\tif code, ok := interpreter.ExitCodeFromError(err); ok {\n\t\t\tos.Exit(code)\n\t\t}\n")
	buf.WriteString("\t\tfmt.Fprintln(os.Stderr, interpreter.DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(err)))\n")
	buf.WriteString("\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("\trt, err := RegisterIn(interp, entryEnv)\n")
	buf.WriteString("\tif err != nil {\n\t\tfmt.Fprintln(os.Stderr, err)\n\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("\tif entryEnv == nil {\n\t\tentryEnv = interp.GlobalEnvironment()\n\t}\n")
	buf.WriteString("\tif err := RunRegisteredMain(rt, interp, entryEnv); err != nil {\n")
	buf.WriteString("\t\tif code, ok := interpreter.ExitCodeFromError(err); ok {\n\t\t\tos.Exit(code)\n\t\t}\n")
	buf.WriteString("\t\tfmt.Fprintln(os.Stderr, interpreter.DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(err)))\n")
	buf.WriteString("\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("}\n\n")
	buf.WriteString("func registerPrint(interp *interpreter.Interpreter) {\n")
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
	buf.WriteString("\tinterp.GlobalEnvironment().Define(\"print\", printFn)\n}\n\n")
	buf.WriteString("type fixtureTypecheckMode int\n\n")
	buf.WriteString("const (\n")
	buf.WriteString("\ttypecheckModeOff fixtureTypecheckMode = iota\n")
	buf.WriteString("\ttypecheckModeWarn\n")
	buf.WriteString("\ttypecheckModeStrict\n")
	buf.WriteString(")\n\n")
	buf.WriteString("func resolveTestTypecheckMode() fixtureTypecheckMode {\n")
	buf.WriteString("\tif _, ok := os.LookupEnv(\"ABLE_TYPECHECK_FIXTURES\"); !ok {\n")
	buf.WriteString("\t\treturn typecheckModeStrict\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tmode := strings.TrimSpace(strings.ToLower(os.Getenv(\"ABLE_TYPECHECK_FIXTURES\")))\n")
	buf.WriteString("\tswitch mode {\n")
	buf.WriteString("\tcase \"\", \"0\", \"off\", \"false\":\n")
	buf.WriteString("\t\treturn typecheckModeOff\n")
	buf.WriteString("\tcase \"strict\", \"fail\", \"error\", \"1\", \"true\":\n")
	buf.WriteString("\t\treturn typecheckModeStrict\n")
	buf.WriteString("\tcase \"warn\", \"warning\":\n")
	buf.WriteString("\t\treturn typecheckModeWarn\n")
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\treturn typecheckModeWarn\n")
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
