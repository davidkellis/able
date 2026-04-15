package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/typechecker"
)

func residualCarrierSynthesisSource() string {
	return strings.Join([]string{
		"package demo",
		"",
		"interface Reader <T> for Self {",
		"  fn read(self: Self) -> T",
		"}",
		"",
		"struct First { value: i32 }",
		"struct Thing { value: i32 }",
		"struct Boom { message_text: String }",
		"",
		"impl Reader i32 for First {",
		"  fn read(self: Self) -> i32 { self.value }",
		"}",
		"",
		"impl Error for Boom {",
		"  fn message(self: Self) -> String { self.message_text }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"struct Box T {",
		"  value: T",
		"}",
		"",
		"type MaybeReader = ?(Reader i32)",
		"type Choice = (Reader i32) | String | i32",
		"type Outcome = Error | (() -> Thing) | String",
		"",
		"fn read_maybe(box: Box MaybeReader) -> i32 {",
		"  box.value match {",
		"    case reader: Reader i32 => reader.read(),",
		"    case nil => 0",
		"  }",
		"}",
		"",
		"fn read_choice(box: Box Choice) -> i32 {",
		"  box.value match {",
		"    case reader: Reader i32 => reader.read(),",
		"    case _: String => -1,",
		"    case n: i32 => n",
		"  }",
		"}",
		"",
		"fn read_outcome(box: Box Outcome) -> i32 {",
		"  box.value match {",
		"    case build: () -> Thing => build().value,",
		"    case _: String => -1,",
		"    case _: Error => 0",
		"  }",
		"}",
		"",
		"fn read_nested_choice(box: Box !(Choice)) -> i32 {",
		"  box.value match {",
		"    case reader: Reader i32 => reader.read(),",
		"    case _: String => -1,",
		"    case n: i32 => n,",
		"    case _: Error => 0",
		"  }",
		"}",
		"",
		"fn read_nested_outcome(box: Box !(Outcome)) -> i32 {",
		"  box.value match {",
		"    case build: () -> Thing => build().value,",
		"    case _: String => -1,",
		"    case _: Error => 0",
		"  }",
		"}",
		"",
		"fn main() -> i32 {",
		"  maybe_box: Box MaybeReader = Box { value: First { value: 7 } }",
		"  choice_box: Box Choice = Box { value: First { value: 8 } }",
		"  outcome_box: Box Outcome = Box { value: fn() -> Thing { Thing { value: 9 } } }",
		"  nested_choice_box: Box !(Choice) = Box { value: First { value: 10 } }",
		"  nested_outcome_box: Box !(Outcome) = Box { value: fn() -> Thing { Thing { value: 11 } } }",
		"  read_maybe(maybe_box) + read_choice(choice_box) + read_outcome(outcome_box) + read_nested_choice(nested_choice_box) + read_nested_outcome(nested_outcome_box)",
		"}",
		"",
	}, "\n")
}

func residualCarrierSynthesisGenerator(t *testing.T) *generator {
	t.Helper()
	root := t.TempDir()
	entryPath := filepath.Join(root, "main.able")
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	if err := os.WriteFile(entryPath, []byte(residualCarrierSynthesisSource()), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
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
	checker := typechecker.NewProgramChecker()
	check, err := checker.Check(program)
	if err != nil {
		t.Fatalf("typecheck: %v", err)
	}
	gen := newGenerator(Options{PackageName: "main", EmitMain: true})
	gen.setTypecheckInference(check.Inferred)
	if err := gen.collect(program); err != nil {
		t.Fatalf("collect: %v", err)
	}
	report, err := DetectDynamicFeatures(program)
	if err != nil {
		t.Fatalf("dynamic features: %v", err)
	}
	gen.setDynamicFeatureReport(report)
	gen.resolveCompileabilityFixedPoint()
	return gen
}

func TestCompilerGenericNominalNormalizedCarrierFamiliesStayNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-residual-carrier-synthesis", residualCarrierSynthesisSource())
	compiledSrc := compiledSourceText(t, result)

	for _, signature := range []string{
		"func __able_compiled_fn_read_maybe(box *Box_",
		"func __able_compiled_fn_read_choice(box *Box_",
		"func __able_compiled_fn_read_outcome(box *Box_",
		"func __able_compiled_fn_read_nested_choice(box *Box_",
		"func __able_compiled_fn_read_nested_outcome(box *Box_",
	} {
		if !strings.Contains(compiledSrc, signature) {
			t.Fatalf("expected specialized generic nominal carrier signature %q:\n%s", signature, compiledSrc)
		}
	}

	for _, fnName := range []string{
		"__able_compiled_fn_read_maybe",
		"__able_compiled_fn_read_choice",
		"__able_compiled_fn_read_outcome",
		"__able_compiled_fn_read_nested_choice",
		"__able_compiled_fn_read_nested_outcome",
	} {
		body, ok := findCompiledFunction(result, fnName)
		if !ok {
			t.Fatalf("could not find compiled %s function", fnName)
		}
		for _, fragment := range []string{
			"runtime.Value",
			" any",
			"__able_try_cast(",
			"bridge.MatchType(",
			"__able_member_get(",
			"__able_method_call_node(",
			"var box *Box =",
		} {
			if strings.Contains(body, fragment) {
				t.Fatalf("expected %s to avoid %q:\n%s", fnName, fragment, body)
			}
		}
		if !strings.Contains(body, "box.Value") {
			t.Fatalf("expected %s to keep direct specialized field access:\n%s", fnName, body)
		}
	}
}

func TestCompilerSharedMapperRecoversGenericNominalNormalizedCarrierFamilies(t *testing.T) {
	gen := residualCarrierSynthesisGenerator(t)
	pkgName := gen.entryPackage
	cases := []struct {
		name string
		expr ast.TypeExpression
	}{
		{"box_maybe_reader", ast.Gen(ast.Ty("Box"), ast.Ty("MaybeReader"))},
		{"box_choice", ast.Gen(ast.Ty("Box"), ast.Ty("Choice"))},
		{"box_outcome", ast.Gen(ast.Ty("Box"), ast.Ty("Outcome"))},
		{"box_nested_choice", ast.Gen(ast.Ty("Box"), ast.Result(ast.Ty("Choice")))},
		{"box_nested_outcome", ast.Gen(ast.Ty("Box"), ast.Result(ast.Ty("Outcome")))},
	}

	for _, tc := range cases {
		mapped, ok := gen.lowerCarrierTypeInPackage(pkgName, tc.expr)
		if !ok || mapped == "" || mapped == "runtime.Value" || mapped == "any" || !strings.HasPrefix(mapped, "*Box_") {
			t.Fatalf("%s: expected specialized native Box carrier, got ok=%t mapped=%q normalized=%q", tc.name, ok, mapped, typeExpressionToString(normalizeTypeExprForPackage(gen, pkgName, tc.expr)))
		}
		info, ok := gen.ensureSpecializedStructInfo(pkgName, tc.expr)
		if !ok || info == nil || !info.Supported || info.GoName == "" || info.GoName == "Box" {
			t.Fatalf("%s: expected specialized struct info, got ok=%t info=%+v", tc.name, ok, info)
		}
		if len(info.Fields) != 1 {
			t.Fatalf("%s: expected one specialized field, got %d", tc.name, len(info.Fields))
		}
		field := info.Fields[0]
		if field.GoType == "" || field.GoType == "runtime.Value" || field.GoType == "any" {
			t.Fatalf("%s: expected specialized field carrier, got %q", tc.name, field.GoType)
		}
		expectedFieldType, ok := gen.lowerCarrierTypeInPackage(info.Package, field.TypeExpr)
		if !ok || expectedFieldType == "" || expectedFieldType == "runtime.Value" || expectedFieldType == "any" {
			t.Fatalf("%s: expected normalized field type to map natively, got ok=%t mapped=%q expr=%q", tc.name, ok, expectedFieldType, typeExpressionToString(field.TypeExpr))
		}
		if expectedFieldType != field.GoType {
			t.Fatalf("%s: expected specialized field carrier %q to match canonical mapper %q for %q", tc.name, field.GoType, expectedFieldType, typeExpressionToString(field.TypeExpr))
		}
	}
}
