package compiler

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree(t *testing.T) {
	mainSrc := compileMainSource(t, "demo", strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  0",
		"}",
		"",
	}, "\n"))
	if strings.Contains(mainSrc, "EvaluateProgram(") {
		t.Fatalf("expected static launcher to skip interpreter program evaluation")
	}
	if strings.Contains(mainSrc, "interpreter.New()") {
		t.Fatalf("expected static launcher to avoid interpreter initialization")
	}
	if !strings.Contains(mainSrc, "RegisterIn(nil, entryEnv)") {
		t.Fatalf("expected static launcher to register compiled runtime in static entry environment")
	}
	if !strings.Contains(mainSrc, "RunRegisteredMain(rt, nil, entryEnv)") {
		t.Fatalf("expected static launcher to execute compiled main without interpreter fallback")
	}
}

func TestCompilerMainKeepsProgramEvaluationWhenDynamicFeaturesPresent(t *testing.T) {
	mainSrc := compileMainSource(t, "demo", strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  dyn.def_package(\"demo.dynamic\")",
		"}",
		"",
	}, "\n"))
	if !strings.Contains(mainSrc, "EvaluateProgram(") {
		t.Fatalf("expected dynamic launcher to evaluate program through interpreter bootstrap")
	}
}

func TestCompilerMainUsesInstalledStdlibDiscoveryBeforeSiblingLookup(t *testing.T) {
	mainSrc := compileMainSource(t, "demo", strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  0",
		"}",
		"",
	}, "\n"))
	if !strings.Contains(mainSrc, "stdlibpath.ResolveInstalledSrc()") {
		t.Fatalf("expected emitted main.go to consult installed stdlib discovery")
	}
}

func TestCompilerMainSkipsProgramEvaluationWhenStaticUsesHelpers(t *testing.T) {
	mainSrc := compileMainSource(t, "demo", strings.Join([]string{
		"package demo",
		"",
		"fn add(x: i32, y: i32) -> i32 {",
		"  x + y",
		"}",
		"",
		"fn main() -> i32 {",
		"  add(1, 2)",
		"}",
		"",
	}, "\n"))
	if strings.Contains(mainSrc, "EvaluateProgram(") {
		t.Fatalf("expected static launcher with helpers to skip interpreter program evaluation")
	}
}

func TestCompilerMainSkipsProgramEvaluationWhenStaticUsesStructMethodsAndImpls(t *testing.T) {
	mainSrc := compileMainSource(t, "demo", strings.Join([]string{
		"package demo",
		"",
		"interface Label {",
		"  fn label(self) -> String",
		"}",
		"",
		"struct Box {",
		"  value: i32",
		"}",
		"",
		"methods Box {",
		"  fn make(value: i32) -> Box {",
		"    Box { value: value }",
		"  }",
		"}",
		"",
		"impl Label for Box {",
		"  fn label(self) -> String {",
		"    \"box\"",
		"  }",
		"}",
		"",
		"fn main() -> String {",
		"  b := Box.make(1)",
		"  b.label()",
		"}",
		"",
	}, "\n"))
	if strings.Contains(mainSrc, "EvaluateProgram(") {
		t.Fatalf("expected static launcher with structs/methods/impls to skip interpreter program evaluation")
	}
}

func TestCompilerMainSkipsProgramEvaluationWhenStaticUsesMultiPackageImports(t *testing.T) {
	mainSrc, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.math.{double}",
			"import demo.shared::shared",
			"",
			"fn main() -> String {",
			"  shared.tag(double(2))",
			"}",
			"",
		}, "\n"),
		"math/helpers.able": strings.Join([]string{
			"fn double(value: i32) -> i32 {",
			"  value * 2",
			"}",
			"",
		}, "\n"),
		"shared/helpers.able": strings.Join([]string{
			"fn tag(value: i32) -> String {",
			"  `tag ${value}`",
			"}",
			"",
		}, "\n"),
	})
	if strings.Contains(mainSrc, "EvaluateProgram(") {
		t.Fatalf("expected static multi-package launcher to skip interpreter program evaluation")
	}
	if !strings.Contains(compiledSrc, "__able_make_pkg_callable") {
		t.Fatalf("expected RegisterIn to seed static import callables for no-bootstrap mode")
	}
	if !strings.Contains(compiledSrc, "entry := __able_lookup_compiled_call(pkgEnv, name)") {
		t.Fatalf("expected static import callable seeding to bind directly to compiled call table")
	}
	if strings.Contains(compiledSrc, "__able_call_named(name, args, nil)") {
		t.Fatalf("expected static import callable seeding to avoid __able_call_named bridge path")
	}
	if !strings.Contains(compiledSrc, "runtime.PackageValue{Name: \"demo.shared\"") {
		t.Fatalf("expected RegisterIn to seed package alias import for no-bootstrap mode")
	}
}

func TestCompilerMainSkipsProgramEvaluationWhenStaticUsesWildcardImport(t *testing.T) {
	mainSrc := compileMainSource(t, "demo", strings.Join([]string{
		"package demo",
		"",
		"import demo.tools.*",
		"",
		"fn main() -> String {",
		"  tag(double(4))",
		"}",
		"",
	}, "\n"), "tools.able", strings.Join([]string{
		"package tools",
		"",
		"fn double(value: i32) -> i32 {",
		"  value * 2",
		"}",
		"",
		"fn tag(value: i32) -> String {",
		"  `tag ${value}`",
		"}",
		"",
	}, "\n"))
	if strings.Contains(mainSrc, "EvaluateProgram(") {
		t.Fatalf("expected static wildcard-import launcher to skip interpreter program evaluation")
	}
}

func TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceImport(t *testing.T) {
	mainSrc, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.iface.{Describer}",
			"import demo.types.{Widget}",
			"import demo.impls.{Fancy}",
			"",
			"fn main() -> String {",
			"  Fancy.describe(Widget { id: 5 })",
			"}",
			"",
		}, "\n"),
		"iface.able": strings.Join([]string{
			"package iface",
			"",
			"interface Describer for Self {",
			"  fn describe(self: Self) -> String",
			"}",
			"",
		}, "\n"),
		"types.able": strings.Join([]string{
			"package types",
			"",
			"struct Widget {",
			"  id: i32",
			"}",
			"",
		}, "\n"),
		"impls.able": strings.Join([]string{
			"package impls",
			"",
			"import demo.iface.{Describer}",
			"import demo.types.{Widget}",
			"",
			"Fancy = impl Describer for Widget {",
			"  fn describe(self: Self) -> String {",
			"    `fancy-${self.id}`",
			"  }",
			"}",
			"",
		}, "\n"),
	})
	if strings.Contains(mainSrc, "EvaluateProgram(") {
		t.Fatalf("expected static named-impl launcher to skip interpreter program evaluation")
	}
	if !strings.Contains(compiledSrc, "runtime.ImplementationNamespaceValue{Name: ast.NewIdentifier(\"Fancy\")") {
		t.Fatalf("expected RegisterIn to seed named impl namespace value for no-bootstrap mode")
	}
}

func TestCompilerMainSkipsProgramEvaluationWhenStaticUsesNamedImplNamespaceOverloads(t *testing.T) {
	mainSrc, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.iface.{Formatter}",
			"import demo.types.{Widget}",
			"import demo.impls.{Fancy}",
			"",
			"fn main() -> String {",
			"  widget := Widget { id: 5 }",
			"  one := Fancy.format(widget, \"tag\")",
			"  two := Fancy.format(widget, \"tag\", 2)",
			"  `${one}|${two}`",
			"}",
			"",
		}, "\n"),
		"iface.able": strings.Join([]string{
			"package iface",
			"",
			"interface Formatter for Self {",
			"  fn format(self: Self, label: String) -> String",
			"  fn format(self: Self, label: String, count: i32) -> String",
			"}",
			"",
		}, "\n"),
		"types.able": strings.Join([]string{
			"package types",
			"",
			"struct Widget {",
			"  id: i32",
			"}",
			"",
		}, "\n"),
		"impls.able": strings.Join([]string{
			"package impls",
			"",
			"import demo.iface.{Formatter}",
			"import demo.types.{Widget}",
			"",
			"Fancy = impl Formatter for Widget {",
			"  fn format(self: Self, label: String) -> String {",
			"    `${label}-${self.id}`",
			"  }",
			"",
			"  fn format(self: Self, label: String, count: i32) -> String {",
			"    `${label}-${count}-${self.id}`",
			"  }",
			"}",
			"",
		}, "\n"),
	})
	if strings.Contains(mainSrc, "EvaluateProgram(") {
		t.Fatalf("expected static named-impl-overload launcher to skip interpreter program evaluation")
	}
	if !strings.Contains(compiledSrc, "No overloads of Fancy.format match provided arguments") {
		t.Fatalf("expected no-bootstrap seeder to emit overload dispatch for named impl namespaces")
	}
}

func TestCompilerMainSkipsProgramEvaluationWhenStaticUsesImportedPublicMethodSelector(t *testing.T) {
	mainSrc, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.spec.{use}",
			"",
			"fn main() -> i32 {",
			"  use()",
			"}",
			"",
		}, "\n"),
		"assertions.able": strings.Join([]string{
			"package assertions",
			"",
			"struct Expectation {",
			"  actual: i32",
			"}",
			"",
			"fn expect(value: i32) -> Expectation {",
			"  Expectation { actual: value }",
			"}",
			"",
			"methods Expectation {",
			"  fn to(self: Self, expected: i32) -> i32 {",
			"    if self.actual == expected { 1 } else { 0 }",
			"  }",
			"}",
			"",
		}, "\n"),
		"spec.able": strings.Join([]string{
			"package spec",
			"",
			"import demo.assertions.{Expectation, expect, to::assertions_to}",
			"",
			"to := assertions_to",
			"",
			"fn use() -> i32 {",
			"  to(expect(7), 7)",
			"}",
			"",
		}, "\n"),
	})
	if strings.Contains(mainSrc, "EvaluateProgram(") {
		t.Fatalf("expected static imported-public-method launcher to skip interpreter program evaluation")
	}
	if !strings.Contains(compiledSrc, "__able_make_compiled_wrapper_callable") {
		t.Fatalf("expected no-bootstrap seeder to synthesize wrapper-backed callables for imported public methods")
	}
	if !strings.Contains(compiledSrc, "__able_public_package_method_demo_assertions_to") {
		t.Fatalf("expected compiled source to emit a package-scoped public method dispatcher:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "\"to\", 2, 2, __able_public_package_method_demo_assertions_to") {
		t.Fatalf("expected no-bootstrap import seeding to bind the imported public method through the shared wrapper helper:\n%s", compiledSrc)
	}
}

func compileMainSource(t *testing.T, pkgName string, source string, extraPairs ...string) string {
	t.Helper()
	files := map[string]string{"main.able": source}
	for idx := 0; idx+1 < len(extraPairs); idx += 2 {
		files[extraPairs[idx]] = extraPairs[idx+1]
	}
	mainSrc, _ := compileOutputs(t, pkgName, files)
	return mainSrc
}

func compileOutputs(t *testing.T, pkgName string, files map[string]string) (string, string) {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: "+pkgName+"\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	entryPath := filepath.Join(root, "main.able")
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	result, err := New(Options{
		PackageName: "main",
		EmitMain:    true,
		EntryPath:   entryPath,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	mainSrc, ok := result.Files["main.go"]
	if !ok {
		t.Fatalf("expected main.go output")
	}
	compiledSrc, ok := result.Files["compiled.go"]
	if !ok {
		t.Fatalf("expected compiled.go output")
	}
	combinedCompiled := string(compiledSrc)
	compiledNames := make([]string, 0, len(result.Files))
	for name := range result.Files {
		if name == "compiled.go" {
			continue
		}
		if strings.HasPrefix(name, "compiled") && strings.HasSuffix(name, ".go") {
			compiledNames = append(compiledNames, name)
		}
	}
	sort.Strings(compiledNames)
	for _, name := range compiledNames {
		combinedCompiled += "\n" + string(result.Files[name])
	}
	return string(mainSrc), combinedCompiled
}
