package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func shadowedImportedCallableAliasPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Thing::RemoteThing, MaybeBuilder, Choice, Outcome, fail}",
			"",
			"struct Thing { local: i32 }",
			"",
			"fn read_nullable(value: MaybeBuilder) -> i32 {",
			"  local: MaybeBuilder = value",
			"  local match {",
			"    case build: (() -> RemoteThing) => build().remote,",
			"    case nil => 0",
			"  }",
			"}",
			"",
			"fn read_union(value: Choice) -> i32 {",
			"  local: Choice = value",
			"  local match {",
			"    case build: (() -> RemoteThing) => build().remote,",
			"    case _text: String => 0",
			"  }",
			"}",
			"",
			"fn read_result(value: Outcome) -> i32 {",
			"  local: Outcome = value",
			"  local match {",
			"    case build: (() -> RemoteThing) => build().remote,",
			"    case _: Error => 0",
			"  }",
			"}",
			"",
			"fn read_rescue() -> i32 {",
			"  fail() rescue {",
			"    case build: (() -> RemoteThing) => build().remote",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"struct Thing { remote: i32 }",
			"",
			"type MaybeBuilder = Option (() -> Thing)",
			"type Choice = (() -> Thing) | String",
			"type Outcome = Result (() -> Thing)",
			"",
			"fn fail() -> Outcome {",
			"  raise(fn() -> Thing { Thing { remote: 7 } })",
			"  fn() -> Thing { Thing { remote: 7 } }",
			"}",
			"",
		}, "\n"),
	}
}

func assertShadowedImportedCallableAliasFunctionStaysNative(t *testing.T, result *Result, fnName string) {
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
		"__able_call_value(",
		"__able_member_get(",
		"unresolved static call",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected %s to avoid %q:\n%s", fnName, fragment, body)
		}
	}
	if !strings.Contains(body, "var build __able_fn_") {
		t.Fatalf("expected %s to keep a native callable binding:\n%s", fnName, body)
	}
	if !strings.Contains(body, ".Remote") {
		t.Fatalf("expected %s to keep direct native field access on the callable result:\n%s", fnName, body)
	}
}

func TestCompilerImportedSemanticOptionAliasWithShadowedCallableStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedCallableAliasPackageFiles())
	assertShadowedImportedCallableAliasFunctionStaysNative(t, result, "__able_compiled_fn_read_nullable")
}

func TestCompilerImportedUnionAliasWithShadowedCallableStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedCallableAliasPackageFiles())
	assertShadowedImportedCallableAliasFunctionStaysNative(t, result, "__able_compiled_fn_read_union")
}

func TestCompilerImportedSemanticResultAliasWithShadowedCallableStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedCallableAliasPackageFiles())
	assertShadowedImportedCallableAliasFunctionStaysNative(t, result, "__able_compiled_fn_read_result")
	assertShadowedImportedCallableAliasFunctionStaysNative(t, result, "__able_compiled_fn_read_rescue")
}

func shadowedImportedCallablePlaceholderAliasPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Thing::RemoteThing, Choice, Outcome}",
			"",
			"struct Thing { local: i32 }",
			"",
			"fn build_choice() -> Choice {",
			"  RemoteThing { remote: @ }",
			"}",
			"",
			"fn build_result() -> Outcome {",
			"  RemoteThing { remote: @ }",
			"}",
			"",
			"fn read_choice(value: Choice) -> i32 {",
			"  value match {",
			"    case build: (i32 -> RemoteThing) => build(7).remote,",
			"    case _text: String => 0",
			"  }",
			"}",
			"",
			"fn read_result(value: Outcome) -> i32 {",
			"  value match {",
			"    case build: (i32 -> RemoteThing) => build(7).remote,",
			"    case _: Error => 0",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"struct Thing { remote: i32 }",
			"",
			"type Choice = (i32 -> Thing) | String",
			"type Outcome = Result (i32 -> Thing)",
			"",
		}, "\n"),
	}
}

func shadowedImportedNestedCallableResultAliasPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Thing::RemoteThing, Outcome}",
			"",
			"struct Thing { local: i32 }",
			"",
			"fn read_nested(value: !(Outcome)) -> i32 {",
			"  value match {",
			"    case build: (() -> RemoteThing) => build().remote,",
			"    case _: Error => 0",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"struct Thing { remote: i32 }",
			"",
			"type Outcome = Result (() -> Thing)",
			"",
		}, "\n"),
	}
}

func TestCompilerImportedUnionAliasWithShadowedCallablePlaceholderStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedCallablePlaceholderAliasPackageFiles())
	assertShadowedImportedCallableAliasFunctionStaysNative(t, result, "__able_compiled_fn_read_choice")

	body, ok := findCompiledFunction(result, "__able_compiled_fn_build_choice")
	if !ok {
		t.Fatalf("could not find compiled build_choice function")
	}
	for _, fragment := range []string{"runtime.Value", " any", "__able_try_cast(", "bridge.MatchType(", "__able_call_value("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected build_choice to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_union_") || !strings.Contains(body, "__able_fn_int32_to_") {
		t.Fatalf("expected build_choice to keep a native callable-union return:\n%s", body)
	}
}

func TestCompilerImportedSemanticResultAliasWithShadowedCallablePlaceholderStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedCallablePlaceholderAliasPackageFiles())
	assertShadowedImportedCallableAliasFunctionStaysNative(t, result, "__able_compiled_fn_read_result")

	body, ok := findCompiledFunction(result, "__able_compiled_fn_build_result")
	if !ok {
		t.Fatalf("could not find compiled build_result function")
	}
	for _, fragment := range []string{"runtime.Value", " any", "__able_try_cast(", "bridge.MatchType(", "__able_call_value("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected build_result to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_union_") || !strings.Contains(body, "__able_fn_int32_to_") {
		t.Fatalf("expected build_result to keep a native callable-result return:\n%s", body)
	}
}

func TestCompilerImportedSemanticResultAliasNestedResultWithShadowedCallableStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedNestedCallableResultAliasPackageFiles())
	assertShadowedImportedCallableAliasFunctionStaysNative(t, result, "__able_compiled_fn_read_nested")
}

func TestCompilerImportedSemanticResultAliasNestedResultWithShadowedCallableResolvesNativePatternTarget(t *testing.T) {
	program := loadShadowedInterfaceProgramFromFiles(t, shadowedImportedNestedCallableResultAliasPackageFiles())
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

	var info *functionInfo
	var pattern *ast.TypedPattern
	for _, candidate := range gen.allFunctionInfos() {
		if candidate == nil || candidate.Definition == nil || candidate.Definition.ID == nil {
			continue
		}
		if candidate.Definition.ID.Name == "read_nested" {
			info = candidate
			break
		}
	}
	if info == nil || len(info.Params) == 0 {
		t.Fatalf("missing read_nested function info")
	}
	for _, mod := range program.Modules {
		for _, stmt := range mod.AST.Body {
			fn, ok := stmt.(*ast.FunctionDefinition)
			if !ok || fn == nil || fn.ID == nil || fn.ID.Name != "read_nested" || fn.Body == nil {
				continue
			}
			for _, bodyStmt := range fn.Body.Body {
				matchExpr, ok := bodyStmt.(*ast.MatchExpression)
				if !ok || matchExpr == nil || len(matchExpr.Clauses) == 0 {
					continue
				}
				typed, ok := matchExpr.Clauses[0].Pattern.(*ast.TypedPattern)
				if ok && typed != nil {
					pattern = typed
					break
				}
			}
		}
	}
	if pattern == nil || pattern.TypeAnnotation == nil {
		t.Fatalf("missing read_nested typed pattern")
	}
	ctx := &compileContext{
		locals:      make(map[string]paramInfo),
		packageName: info.Package,
		returnType:  "int32",
	}
	ctx.setLocalBinding("value", info.Params[0])
	ctx.matchSubjectTypeExpr = info.Params[0].TypeExpr
	target, ok := gen.resolveNativeUnionTypedPatternInContext(ctx, info.Params[0].GoType, pattern.TypeAnnotation)
	if !ok {
		mapped, mappedOK := gen.lowerCarrierType(ctx, pattern.TypeAnnotation)
		normalized := gen.lowerNormalizedTypeExpr(ctx, pattern.TypeAnnotation)
		resolvedPkg := gen.resolvedTypeExprPackage(ctx.packageName, normalized)
		retExpr := ast.TypeExpression(nil)
		retResolvedPkg := ""
		retMapped := ""
		retMappedOK := false
		if fnExpr, ok := normalized.(*ast.FunctionTypeExpression); ok && fnExpr != nil {
			retExpr = fnExpr.ReturnType
			retResolvedPkg = gen.resolvedTypeExprPackage(resolvedPkg, retExpr)
			retMapped, retMappedOK = gen.lowerCarrierTypeInPackage(retResolvedPkg, retExpr)
		}
		t.Fatalf("expected nested callable result alias pattern to resolve natively: subject=%q subjectExpr=%q pattern=%q mapped=%q mappedOK=%t normalized=%q resolvedPkg=%q returnExpr=%q returnPkg=%q returnMapped=%q returnMappedOK=%t", info.Params[0].GoType, typeExpressionToString(info.Params[0].TypeExpr), typeExpressionToString(pattern.TypeAnnotation), mapped, mappedOK, typeExpressionToString(normalized), resolvedPkg, typeExpressionToString(retExpr), retResolvedPkg, retMapped, retMappedOK)
	}
	if target.GoType == "" || target.Member == nil {
		t.Fatalf("expected nested callable result alias pattern to resolve a concrete callable member, got %+v", target)
	}
}
