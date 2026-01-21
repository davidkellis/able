package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/parser"
	"able/interpreter-go/pkg/runtime"
)

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
	return diags
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
	origins := make(map[ast.Node]string)
	for _, mod := range modules {
		if mod == nil || len(mod.NodeOrigins) == 0 {
			continue
		}
		for node, origin := range mod.NodeOrigins {
			if node == nil {
				continue
			}
			origins[node] = origin
		}
	}
	if len(origins) > 0 {
		interp.SetNodeOrigins(origins)
	}
	for _, mod := range modules {
		if mod == nil || mod.AST == nil {
			continue
		}
		if _, _, err := interp.EvaluateModule(mod.AST); err != nil {
			if code, ok := interpreter.ExitCodeFromError(err); ok {
				return false, code
			}
			fmt.Fprintln(os.Stderr, interpreter.DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(err)))
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
