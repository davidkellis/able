package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
	"able/interpreter-go/pkg/typechecker"
)

type parityValue struct {
	Kind  string `json:"kind"`
	Value string `json:"value,omitempty"`
	Bool  *bool  `json:"bool,omitempty"`
}

type parityOutput struct {
	Result        *parityValue `json:"result,omitempty"`
	Stdout        []string     `json:"stdout,omitempty"`
	Error         string       `json:"error,omitempty"`
	Diagnostics   []string     `json:"diagnostics,omitempty"`
	TypecheckMode string       `json:"typecheckMode"`
	Skipped       bool         `json:"skipped"`
}

type typecheckMode int

const (
	modeOff typecheckMode = iota
	modeWarn
	modeStrict
)

func main() {
	dirFlag := flag.String("dir", "", "Path to fixture directory")
	entryFlag := flag.String("entry", "", "Override manifest entry file")
	execFlag := flag.String("executor", "serial", "Executor to use (serial|goroutine)")
	flag.Parse()

	if *dirFlag == "" {
		fmt.Fprintln(os.Stderr, "--dir is required")
		os.Exit(1)
	}

	manifest, err := interpreter.LoadFixtureManifest(*dirFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read manifest: %v\n", err)
		os.Exit(1)
	}

	if len(manifest.SkipTargets) > 0 {
		for _, target := range manifest.SkipTargets {
			if strings.EqualFold(target, "go") {
				writeJSON(parityOutput{
					TypecheckMode: os.Getenv("ABLE_TYPECHECK_FIXTURES"),
					Skipped:       true,
				})
				return
			}
		}
	}

	entry := manifest.Entry
	if entry == "" {
		entry = "module.json"
	}
	if *entryFlag != "" {
		entry = *entryFlag
	}

	executor, err := selectExecutor(*execFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid executor: %v\n", err)
		os.Exit(1)
	}

	outcome, err := runFixture(*dirFlag, entry, manifest.Setup, executor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "evaluation failed: %v\n", err)
		os.Exit(1)
	}

	writeJSON(outcome)
}

func selectExecutor(name string) (interpreter.Executor, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "serial":
		return interpreter.NewSerialExecutor(nil), nil
	case "goroutine":
		return interpreter.NewGoroutineExecutor(nil), nil
	default:
		return nil, fmt.Errorf("unknown executor %q", name)
	}
}

func runFixture(dir, entry string, setup []string, executor interpreter.Executor) (parityOutput, error) {
	var output parityOutput

	modulePath := filepath.Join(dir, entry)
	entryModule, entryOrigin, err := interpreter.LoadFixtureModule(modulePath)
	if err != nil {
		return output, fmt.Errorf("load module %s: %w", modulePath, err)
	}

	setupModules := make([]*driver.Module, 0, len(setup))
	for _, setupFile := range setup {
		setupPath := filepath.Join(dir, setupFile)
		setupModule, setupOrigin, err := interpreter.LoadFixtureModule(setupPath)
		if err != nil {
			return output, fmt.Errorf("load setup %s: %w", setupPath, err)
		}
		setupModules = append(setupModules, fixtureDriverModule(setupModule, setupOrigin))
	}

	entryDriver := fixtureDriverModule(entryModule, entryOrigin)

	// Collect imports to opportunistically load stdlib/kernel modules needed by the fixture.
	imports := make(map[string]struct{})
	for _, mod := range setupModules {
		recordImports(imports, mod.Imports)
	}
	recordImports(imports, entryDriver.Imports)

	modules := make([]*driver.Module, 0, len(setupModules)+4)
	added := make(map[string]bool)
	modules = append(modules, setupModules...)

	loadStdlib := func(entryFile string) error {
		repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
		stdlibRoot := filepath.Join(repoRoot, "v11", "stdlib", "src")
		kernelRoot := filepath.Join(repoRoot, "v11", "kernel", "src")
		loader, err := driver.NewLoader([]driver.SearchPath{
			{Path: stdlibRoot, Kind: driver.RootStdlib},
			{Path: kernelRoot, Kind: driver.RootStdlib},
		})
		if err != nil {
			return err
		}
		prog, err := loader.Load(filepath.Join(stdlibRoot, entryFile))
		if err != nil {
			return err
		}
		for _, mod := range prog.Modules {
			if mod == nil || added[mod.Package] {
				continue
			}
			modules = append(modules, mod)
			added[mod.Package] = true
		}
		if prog.Entry != nil && !added[prog.Entry.Package] {
			modules = append(modules, prog.Entry)
			added[prog.Entry.Package] = true
		}
		return nil
	}

	if hasImportWithPrefix(imports, "able.text.string") {
		if err := loadStdlib(filepath.Join("text", "string.able")); err != nil {
			return output, fmt.Errorf("load stdlib string: %w", err)
		}
	}
	if hasImportWithPrefix(imports, "able.concurrency") {
		if err := loadStdlib(filepath.Join("concurrency", "await.able")); err != nil {
			return output, fmt.Errorf("load stdlib concurrency: %w", err)
		}
	}

	modules = append(modules, entryDriver)
	program := &driver.Program{
		Entry:   entryDriver,
		Modules: modules,
	}

	interp := interpreter.NewWithExecutor(executor)
	mode := configureTypechecker(interp)
	output.TypecheckMode = modeString(mode)

	var stdout []string
	registerPrint(interp, &stdout)

	value, _, check, err := interp.EvaluateProgram(program, interpreter.ProgramEvaluationOptions{
		SkipTypecheck:    mode == modeOff,
		AllowDiagnostics: mode != modeOff,
	})
	output.Diagnostics = formatDiagnostics(check.Diagnostics)
	output.Stdout = stdout
	if err != nil {
		output.Error = interpreter.DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(err))
		return output, nil
	}
	if value != nil {
		normalized := normalizeValue(value)
		output.Result = &normalized
	}
	return output, nil
}

func configureTypechecker(interp *interpreter.Interpreter) typecheckMode {
	modeVal, ok := os.LookupEnv("ABLE_TYPECHECK_FIXTURES")
	if !ok {
		return modeOff
	}
	mode := strings.TrimSpace(strings.ToLower(modeVal))
	switch mode {
	case "", "0", "off", "false":
		return modeOff
	case "strict", "fail", "error", "1", "true":
		interp.EnableTypechecker(interpreter.TypecheckConfig{FailFast: true})
		return modeStrict
	case "warn", "warning":
		interp.EnableTypechecker(interpreter.TypecheckConfig{})
		return modeWarn
	default:
		interp.EnableTypechecker(interpreter.TypecheckConfig{})
		return modeWarn
	}
}

func recordImports(dst map[string]struct{}, imports []string) {
	if len(imports) == 0 {
		return
	}
	for _, name := range imports {
		if name == "" {
			continue
		}
		dst[name] = struct{}{}
	}
}

func hasImportWithPrefix(imports map[string]struct{}, prefix string) bool {
	if prefix == "" {
		return false
	}
	for name := range imports {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func modeString(mode typecheckMode) string {
	switch mode {
	case modeWarn:
		return "warn"
	case modeStrict:
		return "strict"
	default:
		return "off"
	}
}

func normalizeValue(val runtime.Value) parityValue {
	switch v := val.(type) {
	case runtime.StringValue:
		return parityValue{Kind: "String", Value: v.Val}
	case runtime.BoolValue:
		return parityValue{Kind: "bool", Bool: &v.Val}
	case runtime.CharValue:
		return parityValue{Kind: "char", Value: string(v.Val)}
	case runtime.IntegerValue:
		return parityValue{Kind: string(v.TypeSuffix), Value: v.Val.String()}
	case runtime.FloatValue:
		return parityValue{Kind: string(v.TypeSuffix), Value: fmt.Sprintf("%g", v.Val)}
	case runtime.NilValue:
		return parityValue{Kind: "nil"}
	default:
		return parityValue{Kind: v.Kind().String()}
	}
}

func formatDiagnostics(diags []interpreter.ModuleDiagnostic) []string {
	if len(diags) == 0 {
		return nil
	}
	msgs := make([]string, len(diags))
	for i, diag := range diags {
		msgs[i] = formatModuleDiagnostic(diag)
	}
	return msgs
}

func formatModuleDiagnostic(diag interpreter.ModuleDiagnostic) string {
	location := formatSourceHint(diag.Source)
	prefix := "typechecker: "
	if diag.Diagnostic.Severity == typechecker.SeverityWarning {
		prefix = "warning: typechecker: "
	}
	if location != "" {
		return fmt.Sprintf("%s%s %s", prefix, location, diag.Diagnostic.Message)
	}
	if diag.Package != "" {
		return fmt.Sprintf("%s%s %s", prefix, diag.Package, diag.Diagnostic.Message)
	}
	return fmt.Sprintf("%s%s", prefix, diag.Diagnostic.Message)
}

func formatSourceHint(hint typechecker.SourceHint) string {
	path := normalizeSourcePath(strings.TrimSpace(hint.Path))
	line := hint.Line
	column := hint.Column
	switch {
	case path != "" && line > 0 && column > 0:
		return fmt.Sprintf("%s:%d:%d", path, line, column)
	case path != "" && line > 0:
		return fmt.Sprintf("%s:%d", path, line)
	case path != "":
		return path
	case line > 0 && column > 0:
		return fmt.Sprintf("line %d, column %d", line, column)
	case line > 0:
		return fmt.Sprintf("line %d", line)
	default:
		return ""
	}
}

func normalizeSourcePath(raw string) string {
	if raw == "" {
		return ""
	}
	path := raw
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}
	return filepath.ToSlash(path)
}

func fixtureDriverModule(module *ast.Module, origin string) *driver.Module {
	var pkgName string
	if module != nil && module.Package != nil {
		segments := make([]string, 0, len(module.Package.NamePath))
		for _, id := range module.Package.NamePath {
			if id != nil && id.Name != "" {
				segments = append(segments, id.Name)
			}
		}
		pkgName = strings.Join(segments, ".")
	}

	importSet := make(map[string]struct{})
	for _, imp := range module.Imports {
		if imp == nil {
			continue
		}
		parts := make([]string, 0, len(imp.PackagePath))
		for _, id := range imp.PackagePath {
			if id != nil && id.Name != "" {
				parts = append(parts, id.Name)
			}
		}
		if len(parts) > 0 {
			importSet[strings.Join(parts, ".")] = struct{}{}
		}
	}
	imports := make([]string, 0, len(importSet))
	for name := range importSet {
		imports = append(imports, name)
	}

	files := []string{}
	if origin != "" {
		files = []string{origin}
	}

	origins := make(map[ast.Node]string)
	ast.AnnotateOrigins(module, origin, origins)

	return &driver.Module{
		Package:     pkgName,
		AST:         module,
		Files:       files,
		Imports:     imports,
		NodeOrigins: origins,
	}
}

func registerPrint(interp *interpreter.Interpreter, buffer *[]string) {
	printFn := runtime.NativeFunctionValue{
		Name:  "print",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			parts := make([]string, len(args))
			for i, arg := range args {
				parts[i] = formatRuntimeValue(arg)
			}
			*buffer = append(*buffer, strings.Join(parts, " "))
			return runtime.NilValue{}, nil
		},
	}
	interp.GlobalEnvironment().Define("print", printFn)
}

func formatRuntimeValue(val runtime.Value) string {
	switch v := val.(type) {
	case runtime.StringValue:
		return v.Val
	case runtime.BoolValue:
		if v.Val {
			return "true"
		}
		return "false"
	case runtime.IntegerValue:
		return v.Val.String()
	case runtime.FloatValue:
		return fmt.Sprintf("%g", v.Val)
	default:
		return fmt.Sprintf("[%s]", v.Kind())
	}
}

func writeJSON(out parityOutput) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}
