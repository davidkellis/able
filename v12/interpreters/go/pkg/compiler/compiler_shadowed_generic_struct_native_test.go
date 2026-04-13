package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func shadowedImportedGenericStructCarrierPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Thing::RemoteThing, Box, Outcome, Choice}",
			"",
			"struct Thing { local: i32 }",
			"",
			"fn read_result(value: Outcome RemoteThing) -> i32 {",
			"  value match {",
			"    case box: Box RemoteThing => box.value.remote,",
			"    case _: Error => 0",
			"  }",
			"}",
			"",
			"fn read_nested_result(value: !(Outcome RemoteThing)) -> i32 {",
			"  value match {",
			"    case box: Box RemoteThing => box.value.remote,",
			"    case _: Error => 0",
			"  }",
			"}",
			"",
			"fn read_nested_union(value: !(Choice RemoteThing)) -> i32 {",
			"  value match {",
			"    case box: Box RemoteThing => box.value.remote,",
			"    case _: String => -1,",
			"    case _: Error => 0",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"struct Thing { remote: i32 }",
			"struct Box T { value: T }",
			"type Outcome T = Result (Box T)",
			"type Choice T = (Box T) | String",
			"",
		}, "\n"),
	}
}

func assertShadowedImportedGenericStructFunctionStaysNative(t *testing.T, result *Result, fnName string) {
	t.Helper()
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
		"var box *Box =",
		"_as_ptr_Box(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected %s to avoid %q:\n%s", fnName, fragment, body)
		}
	}
	if !strings.Contains(body, ".Value.Remote") {
		t.Fatalf("expected %s to keep direct specialized field access:\n%s", fnName, body)
	}
}

func TestCompilerImportedGenericStructResultAliasWithShadowedNominalStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedGenericStructCarrierPackageFiles())
	assertShadowedImportedGenericStructFunctionStaysNative(t, result, "__able_compiled_fn_read_result")
}

func TestCompilerImportedGenericStructNestedResultAliasWithShadowedNominalStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedGenericStructCarrierPackageFiles())
	assertShadowedImportedGenericStructFunctionStaysNative(t, result, "__able_compiled_fn_read_nested_result")
}

func TestCompilerImportedGenericStructNestedUnionAliasWithShadowedNominalStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedGenericStructCarrierPackageFiles())
	assertShadowedImportedGenericStructFunctionStaysNative(t, result, "__able_compiled_fn_read_nested_union")
}

func TestCompilerImportedGenericStructShadowedNominalSpecializesCarrier(t *testing.T) {
	program := loadShadowedInterfaceProgramFromFiles(t, shadowedImportedGenericStructCarrierPackageFiles())
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

	pkgName := "demo"
	for _, candidate := range gen.allFunctionInfos() {
		if candidate == nil || candidate.Definition == nil || candidate.Definition.ID == nil {
			continue
		}
		if candidate.Definition.ID.Name == "read_result" {
			pkgName = candidate.Package
			break
		}
	}
	expr := ast.NewGenericTypeExpression(ast.Ty("Box"), []ast.TypeExpression{ast.Ty("RemoteThing")})
	info, ok := gen.ensureSpecializedStructInfo(pkgName, expr)
	if !ok || info == nil || !info.Supported || info.GoName == "Box" {
		fieldSummary := []string{}
		if info != nil {
			for _, field := range info.Fields {
				fieldSummary = append(fieldSummary, field.Name+":"+field.GoType+"="+typeExpressionToString(field.TypeExpr))
			}
		}
		normalized := normalizeTypeExprForPackage(gen, pkgName, expr)
		baseInfo, baseOK := gen.structInfoForTypeName(pkgName, "Box")
		concreteArgs := false
		concreteExpr := false
		if generic, ok := normalized.(*ast.GenericTypeExpression); ok && generic != nil && baseInfo != nil && baseInfo.Node != nil {
			concreteArgs = gen.structSpecializationArgsConcrete(pkgName, generic.Arguments, baseInfo.Node.GenericParams)
		}
		if baseInfo != nil {
			concreteExpr = gen.typeExprIsConcreteInPackage(baseInfo.Package, normalized)
		}
		mapped, mappedOK := gen.lowerCarrierTypeInPackage(pkgName, expr)
		t.Fatalf("expected imported generic struct specialization to stay native: pkg=%q ok=%t info=%+v fields=%v mapped=%q mappedOK=%t normalized=%q baseOK=%t base=%+v concreteArgs=%t concreteExpr=%t", pkgName, ok, info, fieldSummary, mapped, mappedOK, typeExpressionToString(normalized), baseOK, baseInfo, concreteArgs, concreteExpr)
	}
}
