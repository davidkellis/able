package interpreter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/driver"
)

var (
	repoRoot               = repositoryRoot()
	stdlibRoot             = fallbackPath(filepath.Join(repoRoot, "v11", "stdlib", "src"), filepath.Join("..", "..", "stdlib", "src"))
	kernelRoot             = fallbackPath(filepath.Join(repoRoot, "v11", "kernel", "src"), filepath.Join("..", "..", "kernel", "src"))
	stdlibStringEntry      = filepath.Join(stdlibRoot, "text", "string.able")
	typecheckBaselineOnce  sync.Once
	typecheckBaselineData  map[string][]string
	typecheckBaselineErr   error
	stdlibConcurrencyEntry = filepath.Join(stdlibRoot, "concurrency", "await.able")
	stdlibLoader           *driver.Loader
)

// runFixtureWithExecutor replays a fixture directory using the provided executor.
func runFixtureWithExecutor(t testingT, dir string, rel string, executor Executor) {
	t.Helper()
	underlying, ok := t.(*testing.T)
	if !ok {
		panic("runFixtureWithExecutor expects *testing.T")
	}
	manifest := readManifest(underlying, dir)
	if shouldSkipTarget(manifest.SkipTargets, "go") {
		return
	}
	skipTS := shouldSkipTarget(manifest.SkipTargets, "ts")
	entry := manifest.Entry
	if entry == "" {
		entry = "module.json"
	}
	modulePath := filepath.Join(dir, entry)
	module, moduleOrigin := readModule(underlying, modulePath)

	interp := NewWithExecutor(executor)
	mode := configureFixtureTypechecker(interp)
	var stdout []string
	registerPrint(interp, &stdout)

	var programModules []*driver.Module
	added := make(map[string]bool)
	imports := make(map[string]struct{})
	for _, setup := range manifest.Setup {
		setupPath := filepath.Join(dir, setup)
		setupModule, setupOrigin := readModule(underlying, setupPath)
		setupDriverModule := fixtureDriverModule(setupModule, setupOrigin)
		programModules = append(programModules, setupDriverModule)
		recordImports(imports, setupDriverModule.Imports)
		if setupModule != nil && setupModule.Package != nil {
			parts := make([]string, 0, len(setupModule.Package.NamePath))
			for _, id := range setupModule.Package.NamePath {
				if id == nil || id.Name == "" {
					continue
				}
				parts = append(parts, id.Name)
			}
			if len(parts) > 0 {
				added[strings.Join(parts, ".")] = true
			}
		}
	}

	entryModule := fixtureDriverModule(module, moduleOrigin)
	added[entryModule.Package] = true
		recordImports(imports, entryModule.Imports)
	if hasImport(imports, "able.text.string") {
		if stdlibLoader == nil {
			loader, err := driver.NewLoader([]driver.SearchPath{
				{Path: stdlibRoot, Kind: driver.RootStdlib},
				{Path: kernelRoot, Kind: driver.RootStdlib},
			})
			if err != nil {
				t.Fatalf("stdlib loader init: %v", err)
			}
			stdlibLoader = loader
		}
		stdlibProgram, err := stdlibLoader.Load(stdlibStringEntry)
		if err != nil {
			t.Fatalf("load stdlib string: %v", err)
		}
		for _, mod := range stdlibProgram.Modules {
			if mod == nil || added[mod.Package] {
				continue
			}
			programModules = append(programModules, mod)
			added[mod.Package] = true
		}
		if stdlibProgram.Entry != nil && !added[stdlibProgram.Entry.Package] {
			programModules = append(programModules, stdlibProgram.Entry)
			added[stdlibProgram.Entry.Package] = true
		}
	}
	if hasImportWithPrefix(imports, "able.concurrency") {
		if stdlibLoader == nil {
			loader, err := driver.NewLoader([]driver.SearchPath{
				{Path: stdlibRoot, Kind: driver.RootStdlib},
				{Path: kernelRoot, Kind: driver.RootStdlib},
			})
			if err != nil {
				t.Fatalf("stdlib loader init: %v", err)
			}
			stdlibLoader = loader
		}
		stdlibProgram, err := stdlibLoader.Load(stdlibConcurrencyEntry)
		if err != nil {
			t.Fatalf("load stdlib concurrency: %v", err)
		}
		for _, mod := range stdlibProgram.Modules {
			if mod == nil || added[mod.Package] {
				continue
			}
			programModules = append(programModules, mod)
			added[mod.Package] = true
		}
		if stdlibProgram.Entry != nil && !added[stdlibProgram.Entry.Package] {
			programModules = append(programModules, stdlibProgram.Entry)
			added[stdlibProgram.Entry.Package] = true
		}
	}
	programModules = append(programModules, entryModule)
	program := &driver.Program{
		Entry:   entryModule,
		Modules: programModules,
	}

	value, _, check, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{
		SkipTypecheck:    mode == typecheckModeOff,
		AllowDiagnostics: mode != typecheckModeOff,
	})

	checkerDiags := append([]ModuleDiagnostic(nil), check.Diagnostics...)
	checkerDiags = filterStdlibDiagnostics(checkerDiags)
	actualDiagnostics := checkFixtureTypecheckDiagnostics(underlying, mode, manifest.Expect.TypecheckDiagnostics, checkerDiags, skipTS)
	enforceTypecheckBaseline(underlying, rel, mode, actualDiagnostics, skipTS)

	if len(manifest.Expect.Errors) > 0 {
		if err == nil {
			t.Fatalf("expected evaluation error")
		}
		msg := extractErrorMessage(err)
		if !contains(manifest.Expect.Errors, msg) {
			t.Fatalf("expected error in %v, got %s", manifest.Expect.Errors, msg)
		}
		return
	}
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}
	assertResult(underlying, dir, manifest, value, stdout)
}

func shouldSkipTarget(skip []string, target string) bool {
	if len(skip) == 0 {
		return false
	}
	target = strings.ToLower(target)
	for _, entry := range skip {
		if strings.ToLower(strings.TrimSpace(entry)) == target {
			return true
		}
	}
	return false
}

// testingT captures the subset of testing.T used by fixture helpers.
type testingT interface {
	Helper()
	Fatalf(format string, args ...interface{})
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

func hasImport(imports map[string]struct{}, target string) bool {
	_, ok := imports[target]
	return ok
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

func fallbackPath(primary string, fallback string) string {
	if primary != "" {
		if info, err := os.Stat(primary); err == nil && info.IsDir() {
			return primary
		}
	}
	return fallback
}

func fixtureDriverModule(module *ast.Module, file string) *driver.Module {
	var pkgName string
	if module != nil && module.Package != nil {
		parts := make([]string, 0, len(module.Package.NamePath))
		for _, id := range module.Package.NamePath {
			if id == nil || id.Name == "" {
				continue
			}
			parts = append(parts, id.Name)
		}
		pkgName = strings.Join(parts, ".")
	}
	importSet := make(map[string]struct{})
	for _, imp := range module.Imports {
		if imp == nil {
			continue
		}
		parts := make([]string, 0, len(imp.PackagePath))
		for _, id := range imp.PackagePath {
			if id == nil || id.Name == "" {
				continue
			}
			parts = append(parts, id.Name)
		}
		if len(parts) == 0 {
			continue
		}
		importSet[strings.Join(parts, ".")] = struct{}{}
	}
	imports := make([]string, 0, len(importSet))
	for name := range importSet {
		imports = append(imports, name)
	}
	sort.Strings(imports)

	files := []string{}
	if file != "" {
		files = []string{file}
	}
	origins := make(map[ast.Node]string)
	ast.AnnotateOrigins(module, file, origins)
	return &driver.Module{
		Package:     pkgName,
		AST:         module,
		Files:       files,
		Imports:     imports,
		NodeOrigins: origins,
	}
}

func enforceTypecheckBaseline(t *testing.T, rel string, mode fixtureTypecheckMode, actual []string, skipTS bool) {
	if mode == typecheckModeOff || skipTS {
		return
	}
	baseline := getTypecheckBaseline(t)
	if baseline == nil {
		return
	}
	key := filepath.ToSlash(rel)
	expected, ok := baseline[key]
	if !ok {
		if len(actual) == 0 {
			return
		}
		t.Fatalf("typechecker baseline missing entry for %s (actual %v)", key, actual)
	}
	expectedKeys := diagnosticKeys(expected)
	actualKeys := diagnosticKeys(actual)
	if len(actualKeys) < len(expectedKeys) {
		t.Fatalf("typechecker diagnostics mismatch for %s: expected %v, got %v", key, expected, actual)
	}
	for _, expectedKey := range expectedKeys {
		found := false
		for _, actualKey := range actualKeys {
			expMessage := expectedKey.message
			actMessage := actualKey.message
			if !strings.HasPrefix(expMessage, "typechecker:") && strings.HasPrefix(actMessage, "typechecker:") {
				expMessage = "typechecker: " + expMessage
			}
			if actMessage != expMessage {
				continue
			}
			if expectedKey.path != "" && expectedKey.path != actualKey.path {
				continue
			}
			if expectedKey.line != 0 && expectedKey.line != actualKey.line {
				continue
			}
			found = true
			break
		}
		if !found {
			fmt.Printf("expected key %+v unmatched in actual %v\n", expectedKey, actualKeys)
			t.Fatalf("typechecker diagnostics mismatch for %s: expected %v, got %v", key, expected, actual)
		}
	}
}

func getTypecheckBaseline(t testingT) map[string][]string {
	t.Helper()
	typecheckBaselineOnce.Do(func() {
		path := filepath.Join(repoRoot, "v11", "fixtures", "ast", "typecheck-baseline.json")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			path = filepath.Join("..", "..", "fixtures", "ast", "typecheck-baseline.json")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				typecheckBaselineData = nil
				return
			}
			typecheckBaselineErr = fmt.Errorf("read baseline %s: %w", path, err)
			return
		}
		var baseline map[string][]string
		if err := json.Unmarshal(data, &baseline); err != nil {
			typecheckBaselineErr = fmt.Errorf("parse baseline %s: %w", path, err)
			return
		}
		typecheckBaselineData = baseline
	})
	if typecheckBaselineErr != nil {
		t.Fatalf(typecheckBaselineErr.Error())
	}
	return typecheckBaselineData
}

func filterStdlibDiagnostics(diags []ModuleDiagnostic) []ModuleDiagnostic {
	if len(diags) == 0 {
		return diags
	}
	prefix := filepath.ToSlash(filepath.Join("stdlib", "src")) + "/"
	out := diags[:0]
	for _, diag := range diags {
		if diagnosticFromStdlib(diag, prefix) {
			continue
		}
		out = append(out, diag)
	}
	return out
}

func diagnosticFromStdlib(diag ModuleDiagnostic, prefix string) bool {
	if prefix == "" {
		return false
	}
	paths := []string{diag.Source.Path}
	paths = append(paths, diag.Files...)
	for _, path := range paths {
		normalized := filepath.ToSlash(normalizeSourcePath(path))
		if normalized == "" {
			continue
		}
		if strings.HasPrefix(normalized, prefix) || strings.Contains(normalized, "stdlib/src/") {
			return true
		}
	}
	return false
}

type diagKey struct {
	path    string
	line    int
	message string
}

func diagnosticKeys(entries []string) []diagKey {
	if len(entries) == 0 {
		return nil
	}
	keys := make([]diagKey, 0, len(entries))
	for _, entry := range entries {
		key := parseDiagnosticKey(entry)
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].path != keys[j].path {
			return keys[i].path < keys[j].path
		}
		if keys[i].line != keys[j].line {
			return keys[i].line < keys[j].line
		}
		return keys[i].message < keys[j].message
	})
	return keys
}

func parseDiagnosticKey(entry string) diagKey {
	trimmed := strings.TrimPrefix(entry, "typechecker: ")
	parts := strings.SplitN(trimmed, " ", 2)
	if len(parts) != 2 {
		return diagKey{message: trimmed}
	}
	location := parts[0]
	message := parts[1]
	if !strings.HasPrefix(message, "typechecker:") {
		message = "typechecker: " + message
	}
	path := location
	line := 0
	segments := strings.Split(location, ":")
	if len(segments) >= 2 {
		lineIndex := len(segments) - 1
		columnCandidate := segments[lineIndex]
		if _, err := strconv.Atoi(columnCandidate); err == nil {
			// Treat last segment as column; drop it and parse preceding as line if available
			segments = segments[:lineIndex]
			lineIndex--
			if lineIndex >= 0 {
				lineCandidate := segments[lineIndex]
				if parsed, err := strconv.Atoi(lineCandidate); err == nil {
					line = parsed
					segments = segments[:lineIndex]
				}
			}
		} else if parsed, err := strconv.Atoi(columnCandidate); err == nil {
			line = parsed
			segments = segments[:len(segments)-1]
		}
		path = strings.Join(segments, ":")
	}
	path = filepath.ToSlash(path)
	for strings.HasPrefix(path, "../") {
		path = strings.TrimPrefix(path, "../")
	}
	if path == "typechecker" || path == "typechecker:" || path == "" {
		path = ""
	}
	return diagKey{path: path, line: line, message: message}
}
