package compiler

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/typechecker"
)

func stdlibFsReadTextSource() string {
	return strings.Join([]string{
		"package demo",
		"",
		"import able.fs",
		"import able.io",
		"import able.io.path",
		"import able.io.temp",
		"",
		"fn main() -> void {",
		"  root := temp.dir(\"ablec-fs-read-text-\").path",
		"  file := path.parse(root).join(\"data.txt\").to_string()",
		"  do {",
		"    fs.mkdir(root, true)",
		"    writer := fs.open(file, fs.write_only(true, true), nil)",
		"    io.write_all(writer, io.string_to_bytes(\"alpha\"))",
		"    io.close(writer)",
		"    print(fs.read_text(file))",
		"  } ensure {",
		"    fs.remove(root, true)",
		"  }",
		"}",
		"",
	}, "\n")
}

func formatTypeBindings(bindings map[string]ast.TypeExpression) string {
	if len(bindings) == 0 {
		return "{}"
	}
	names := make([]string, 0, len(bindings))
	for name := range bindings {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, name+"="+typeExpressionToString(bindings[name]))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func TestCompilerStdlibIoWriteAllTempExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-stdlib-io-write-all-", strings.Join([]string{
		"package demo",
		"",
		"import able.io",
		"import able.io.temp",
		"",
		"fn main() -> void {",
		"  tmp := temp.file(\"ablec-io-write-all-\")",
		"  do {",
		"    io.write_all(tmp.handle, io.string_to_bytes(\"alpha\"))",
		"    io.close(tmp.handle)",
		"    print(\"ok\")",
		"  } ensure {",
		"    temp.cleanup(tmp)",
		"  }",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "ok" {
		t.Fatalf("expected stdlib io write_all temp output ok, got %q", stdout)
	}
}

func TestCompilerStdlibFsReadTextAfterWriteExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-stdlib-fs-read-text-", stdlibFsReadTextSource(), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "alpha" {
		t.Fatalf("expected stdlib fs read_text output alpha, got %q", stdout)
	}
}

func TestDebugStdlibFsReadTextCompiledOutput(t *testing.T) {
	moduleRoot, workDir := compilerTestWorkDirNoCleanup(t, "ablec-stdlib-fs-read-text-dump-")
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-stdlib-fs-read-text-debug-", stdlibFsReadTextSource(), Options{
		PackageName: "main",
	})
	outputDir := filepath.Join(workDir, "out")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output: %v", err)
	}
	if err := result.Write(outputDir); err != nil {
		t.Fatalf("write output: %v", err)
	}
	t.Fatalf("wrote compiled output to %s", filepath.Join(moduleRoot, "tmp", filepath.Base(workDir), "out"))
}

func TestDebugStdlibFsReadTextUnwrapSpecializations(t *testing.T) {
	_, workDir := compilerTestWorkDirNoCleanup(t, "ablec-stdlib-fs-read-text-specs-")
	entryPath := filepath.Join(workDir, "main.able")
	if err := os.WriteFile(filepath.Join(workDir, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	if err := os.WriteFile(entryPath, []byte(stdlibFsReadTextSource()), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	searchPaths, err := buildExecSearchPaths(entryPath, workDir, interpreter.FixtureManifest{})
	if err != nil {
		t.Fatalf("build search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	defer loader.Close()
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}
	checker := typechecker.NewProgramChecker()
	if _, err := checker.Check(program); err != nil {
		t.Fatalf("typecheck: %v", err)
	}
	gen := newGenerator(Options{PackageName: "main"})
	if err := gen.collect(program); err != nil {
		t.Fatalf("collect: %v", err)
	}
	dynamicReport, err := DetectDynamicFeatures(program)
	if err != nil {
		t.Fatalf("dynamic features: %v", err)
	}
	gen.setDynamicFeatureReport(dynamicReport)
	gen.resolveCompileableFunctions()
	gen.resolveCompileableMethods()
	if _, err := gen.render(); err != nil {
		t.Fatalf("render: %v", err)
	}
	var lines []string
	for _, info := range gen.specializedFunctions {
		if info == nil || info.Name != "unwrap" {
			continue
		}
		lines = append(lines, "go="+info.GoName+" bindings="+formatTypeBindings(info.TypeBindings)+" return="+typeExpressionToString(gen.functionReturnTypeExpr(info)))
	}
	if len(lines) == 0 {
		t.Fatalf("no unwrap specializations found")
	}
	t.Fatalf("%s", strings.Join(lines, "\n"))
}
