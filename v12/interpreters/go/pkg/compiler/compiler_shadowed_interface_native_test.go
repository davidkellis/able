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

func shadowedImportedInterfacePackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, MaybeReader, Choice, First, fail}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn read_nullable(value: MaybeReader) -> i32 {",
			"  value match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case nil => 0",
			"  }",
			"}",
			"",
			"fn read_union(value: Choice) -> i32 {",
			"  value match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case text: String => 0",
			"  }",
			"}",
			"",
			"fn read_result() -> i32 {",
			"  fail() rescue {",
			"    case reader: RemoteReader i32 => reader.read()",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct First {}",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
			"type MaybeReader = ?(Reader i32)",
			"type Choice = (Reader i32) | String",
			"",
			"fn fail() -> !(Reader i32) {",
			"  raise(First {})",
			"  First {}",
			"}",
			"",
		}, "\n"),
	}
}

func shadowedImportedGenericInterfaceAliasPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, Choice, Outcome, First, fail}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn read_union(value: Choice (RemoteReader i32)) -> i32 {",
			"  value match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: String => 0",
			"  }",
			"}",
			"",
			"fn read_result(value: Outcome (RemoteReader i32)) -> i32 {",
			"  value match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: Error => 0",
			"  }",
			"}",
			"",
			"fn read_direct() -> i32 {",
			"  read_union(First {}) + read_result(fail())",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct First {}",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
			"type Choice T = T | String",
			"type Outcome T = Error | T",
			"",
			"fn fail() -> Outcome (Reader i32) {",
			"  First {}",
			"}",
			"",
		}, "\n"),
	}
}

func loadShadowedImportedInterfaceProgram(t *testing.T) *driver.Program {
	t.Helper()
	return loadShadowedInterfaceProgramFromFiles(t, shadowedImportedInterfacePackageFiles())
}

func loadShadowedInterfaceProgramFromFiles(t *testing.T, files map[string]string) *driver.Program {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
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
	program, err := loader.Load(filepath.Join(root, "main.able"))
	if err != nil {
		t.Fatalf("load program: %v", err)
	}
	return program
}

func shadowedImportedInterfaceUnionParts(t *testing.T) (*generator, *compileContext, string, *ast.TypedPattern) {
	t.Helper()
	program := loadShadowedImportedInterfaceProgram(t)
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
	for _, candidate := range gen.allFunctionInfos() {
		if candidate == nil || candidate.Definition == nil || candidate.Definition.ID == nil {
			continue
		}
		if candidate.Definition.ID.Name == "read_union" {
			info = candidate
			break
		}
	}
	if info == nil || len(info.Params) == 0 {
		t.Fatalf("missing read_union function info")
	}
	var matchExpr *ast.MatchExpression
	for _, mod := range program.Modules {
		for _, stmt := range mod.AST.Body {
			fn, ok := stmt.(*ast.FunctionDefinition)
			if !ok || fn == nil || fn.ID == nil || fn.ID.Name != "read_union" || fn.Body == nil {
				continue
			}
			for _, bodyStmt := range fn.Body.Body {
				if typed, ok := bodyStmt.(*ast.MatchExpression); ok {
					matchExpr = typed
					break
				}
			}
		}
	}
	if matchExpr == nil || len(matchExpr.Clauses) == 0 {
		t.Fatalf("missing read_union match expression")
	}
	pattern, ok := matchExpr.Clauses[0].Pattern.(*ast.TypedPattern)
	if !ok || pattern == nil || pattern.TypeAnnotation == nil {
		t.Fatalf("missing read_union typed pattern")
	}
	temp := 0
	ctx := &compileContext{
		locals:      make(map[string]paramInfo),
		packageName: info.Package,
		temps:       &temp,
		returnType:  "int32",
	}
	ctx.setLocalBinding("value", info.Params[0])
	ctx.matchSubjectTypeExpr = info.Params[0].TypeExpr
	return gen, ctx, info.Params[0].GoType, pattern
}

func shadowedImportedGenericInterfaceClauseParts(t *testing.T, fnName string, clauseIndex int) (*generator, *compileContext, string, *ast.TypedPattern) {
	t.Helper()
	program := loadShadowedInterfaceProgramFromFiles(t, shadowedImportedGenericInterfaceAliasPackageFiles())
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
	for _, candidate := range gen.allFunctionInfos() {
		if candidate == nil || candidate.Definition == nil || candidate.Definition.ID == nil {
			continue
		}
		if candidate.Definition.ID.Name == fnName {
			info = candidate
			break
		}
	}
	if info == nil || len(info.Params) == 0 {
		t.Fatalf("missing %s function info", fnName)
	}
	var matchExpr *ast.MatchExpression
	for _, mod := range program.Modules {
		for _, stmt := range mod.AST.Body {
			fn, ok := stmt.(*ast.FunctionDefinition)
			if !ok || fn == nil || fn.ID == nil || fn.ID.Name != fnName || fn.Body == nil {
				continue
			}
			for _, bodyStmt := range fn.Body.Body {
				if typed, ok := bodyStmt.(*ast.MatchExpression); ok {
					matchExpr = typed
					break
				}
			}
		}
	}
	if matchExpr == nil || len(matchExpr.Clauses) == 0 {
		t.Fatalf("missing %s match expression", fnName)
	}
	if clauseIndex < 0 || clauseIndex >= len(matchExpr.Clauses) {
		t.Fatalf("missing %s clause %d", fnName, clauseIndex)
	}
	pattern, ok := matchExpr.Clauses[clauseIndex].Pattern.(*ast.TypedPattern)
	if !ok || pattern == nil || pattern.TypeAnnotation == nil {
		t.Fatalf("missing %s typed pattern at clause %d", fnName, clauseIndex)
	}
	temp := 0
	ctx := &compileContext{
		locals:      make(map[string]paramInfo),
		packageName: info.Package,
		temps:       &temp,
		returnType:  "int32",
	}
	ctx.setLocalBinding("value", info.Params[0])
	ctx.matchSubjectTypeExpr = info.Params[0].TypeExpr
	return gen, ctx, info.Params[0].GoType, pattern
}

func assertShadowedImportedInterfaceFunctionStaysNative(t *testing.T, result *Result, fnName string) {
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
		"__able_method_call_node(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected %s to avoid %q:\n%s", fnName, fragment, body)
		}
	}
	if !strings.Contains(body, "reader.read()") {
		t.Fatalf("expected %s to keep direct native interface dispatch:\n%s", fnName, body)
	}
}

func TestCompilerImportedNullableAliasWithShadowedInterfaceStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedInterfacePackageFiles())
	assertShadowedImportedInterfaceFunctionStaysNative(t, result, "__able_compiled_fn_read_nullable")
}

func TestCompilerImportedUnionAliasWithShadowedInterfaceStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedInterfacePackageFiles())
	assertShadowedImportedInterfaceFunctionStaysNative(t, result, "__able_compiled_fn_read_union")
}

func TestCompilerImportedResultAliasWithShadowedInterfaceStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedInterfacePackageFiles())
	assertShadowedImportedInterfaceFunctionStaysNative(t, result, "__able_compiled_fn_read_result")
}

func TestCompilerImportedGenericUnionAliasWithShadowedInterfaceResolvesNativePatternTarget(t *testing.T) {
	gen, ctx, subjectType, pattern := shadowedImportedGenericInterfaceClauseParts(t, "read_union", 0)
	normalized := gen.lowerNormalizedTypeExpr(ctx, pattern.TypeAnnotation)
	mapped, mappedOK := gen.lowerCarrierType(ctx, pattern.TypeAnnotation)
	if !mappedOK || mapped == "" || mapped == "runtime.Value" || mapped == "any" {
		t.Fatalf("expected generic imported union pattern carrier to stay native, got normalized=%q mapped=%q mappedOK=%t", typeExpressionToString(normalized), mapped, mappedOK)
	}
	target, ok := gen.resolveNativeUnionTypedPatternInContext(ctx, subjectType, pattern.TypeAnnotation)
	if !ok {
		localIface, _ := gen.ensureNativeInterfaceInfo(ctx.packageName, pattern.TypeAnnotation)
		remoteIface, _ := gen.ensureNativeInterfaceInfo("demo.remote", ast.Gen(ast.Ty("Reader"), ast.Ty("i32")))
		t.Fatalf("resolve generic imported union typed pattern failed: subject=%q pattern=%q normalized=%q mapped=%q mappedOK=%t subjectExpr=%q localIface=%v remoteIface=%v", subjectType, typeExpressionToString(pattern.TypeAnnotation), typeExpressionToString(normalized), mapped, mappedOK, typeExpressionToString(ctx.matchSubjectTypeExpr), localIface, remoteIface)
	}
	if target.GoType == "" || target.Member == nil {
		t.Fatalf("expected imported generic union alias pattern to resolve a concrete member, got %+v", target)
	}
}

func TestCompilerImportedGenericResultAliasWithShadowedInterfaceResolvesNativePatternTarget(t *testing.T) {
	gen, ctx, subjectType, pattern := shadowedImportedGenericInterfaceClauseParts(t, "read_result", 0)
	target, ok := gen.resolveNativeUnionTypedPatternInContext(ctx, subjectType, pattern.TypeAnnotation)
	if !ok {
		normalized := gen.lowerNormalizedTypeExpr(ctx, pattern.TypeAnnotation)
		mapped, mappedOK := gen.lowerCarrierType(ctx, pattern.TypeAnnotation)
		pkgMapped, pkgMappedOK := gen.lowerCarrierTypeInPackage(ctx.packageName, pattern.TypeAnnotation)
		sourcePkg, sourceName := gen.importedSelectorSourceTypeAlias(ctx.packageName, "RemoteReader")
		aliasPkg, aliasName, aliasTarget, _, aliasOK := gen.typeAliasTargetForPackage(ctx.packageName, "RemoteReader")
		localIface, _ := gen.ensureNativeInterfaceInfo(ctx.packageName, pattern.TypeAnnotation)
		remoteIface, _ := gen.ensureNativeInterfaceInfo("demo.remote", ast.Gen(ast.Ty("Reader"), ast.Ty("i32")))
		t.Fatalf("resolve generic imported result typed pattern failed: subject=%q pattern=%q normalized=%q mapped=%q mappedOK=%t pkgMapped=%q pkgMappedOK=%t subjectExpr=%q typeBindings=%#v import=%s.%s alias=%t aliasPkg=%q aliasName=%q aliasTarget=%q localIface=%v remoteIface=%v", subjectType, typeExpressionToString(pattern.TypeAnnotation), typeExpressionToString(normalized), mapped, mappedOK, pkgMapped, pkgMappedOK, typeExpressionToString(ctx.matchSubjectTypeExpr), ctx.typeBindings, sourcePkg, sourceName, aliasOK, aliasPkg, aliasName, typeExpressionToString(aliasTarget), localIface, remoteIface)
	}
	if target.GoType == "" || target.Member == nil {
		t.Fatalf("expected imported generic result alias pattern to resolve a concrete member, got %+v", target)
	}
}

func TestCompilerImportedGenericUnionAliasWithShadowedInterfaceClausesCompileNatively(t *testing.T) {
	gen, ctx, subjectType, first := shadowedImportedGenericInterfaceClauseParts(t, "read_union", 0)
	if _, _, _, ok := gen.compileMatchPattern(ctx, first, "value", subjectType); !ok {
		t.Fatalf("expected first imported generic union clause to compile natively, got reason %q", ctx.reason)
	}
	narrowedType := gen.narrowedNativeUnionSubjectType(ctx, subjectType, first)
	gen, ctx, _, second := shadowedImportedGenericInterfaceClauseParts(t, "read_union", 1)
	if _, _, _, ok := gen.compileMatchPattern(ctx, second, "value", narrowedType); !ok {
		t.Fatalf("expected narrowed imported generic union clause to compile natively, got reason %q", ctx.reason)
	}
}

func TestCompilerImportedGenericResultAliasWithShadowedInterfaceClausesCompileNatively(t *testing.T) {
	gen, ctx, subjectType, first := shadowedImportedGenericInterfaceClauseParts(t, "read_result", 0)
	if _, _, _, ok := gen.compileMatchPattern(ctx, first, "value", subjectType); !ok {
		t.Fatalf("expected first imported generic result clause to compile natively, got reason %q", ctx.reason)
	}
	narrowedType := gen.narrowedNativeUnionSubjectType(ctx, subjectType, first)
	gen, ctx, _, second := shadowedImportedGenericInterfaceClauseParts(t, "read_result", 1)
	if _, _, _, ok := gen.compileMatchPattern(ctx, second, "value", narrowedType); !ok {
		t.Fatalf("expected narrowed imported generic result clause to compile natively, got reason %q", ctx.reason)
	}
}

func TestCompilerImportedGenericUnionAliasWithShadowedInterfaceStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedGenericInterfaceAliasPackageFiles())
	assertShadowedImportedInterfaceFunctionStaysNative(t, result, "__able_compiled_fn_read_union")
}

func TestCompilerImportedGenericResultAliasWithShadowedInterfaceStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedGenericInterfaceAliasPackageFiles())
	assertShadowedImportedInterfaceFunctionStaysNative(t, result, "__able_compiled_fn_read_result")
}

func TestCompilerImportedUnionAliasWithShadowedInterfaceResolvesNativePatternTarget(t *testing.T) {
	gen, ctx, subjectType, pattern := shadowedImportedInterfaceUnionParts(t)
	target, ok := gen.resolveNativeUnionTypedPatternInContext(ctx, subjectType, pattern.TypeAnnotation)
	if !ok {
		normalized := gen.lowerNormalizedTypeExpr(ctx, pattern.TypeAnnotation)
		resolvedPkg := gen.resolvedTypeExprPackage(ctx.packageName, normalized)
		carrier, carrierOK := gen.recoverTypedPatternCarrier(ctx, pattern.TypeAnnotation)
		sourcePkg, sourceName := gen.importedSelectorSourceTypeAlias(ctx.packageName, "RemoteReader")
		aliasPkg, aliasName, aliasTarget, _, aliasOK := gen.typeAliasTargetForPackage(ctx.packageName, "RemoteReader")
		t.Fatalf("resolve union typed pattern failed: subject=%q pattern=%q normalized=%q pkg=%q carrier=%q carrierOK=%t import=%s.%s alias=%t aliasPkg=%q aliasName=%q aliasTarget=%q staticImports=%#v", subjectType, typeExpressionToString(pattern.TypeAnnotation), typeExpressionToString(normalized), resolvedPkg, carrier, carrierOK, sourcePkg, sourceName, aliasOK, aliasPkg, aliasName, typeExpressionToString(aliasTarget), gen.staticImports)
	}
	if target.GoType == "" || target.Member == nil {
		t.Fatalf("expected imported shadowed interface union pattern to resolve a concrete member, got %+v", target)
	}
}

func TestCompilerImportedSemanticResultAliasWithShadowedInterfaceStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, Outcome, First, fail}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn read_result() -> i32 {",
			"  fail() rescue {",
			"    case reader: RemoteReader i32 => reader.read()",
			"  }",
			"}",
			"",
			"fn read_local(value: Outcome) -> i32 {",
			"  local: Outcome = value",
			"  local match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: Error => 0",
			"  }",
			"}",
			"",
			"fn read_direct() -> i32 {",
			"  read_local(First {})",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct First {}",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
			"type Outcome = Result (Reader i32)",
			"",
			"fn fail() -> Outcome {",
			"  raise(First {})",
			"  First {}",
			"}",
			"",
		}, "\n"),
	})

	assertShadowedImportedInterfaceFunctionStaysNative(t, result, "__able_compiled_fn_read_result")
	assertShadowedImportedInterfaceFunctionStaysNative(t, result, "__able_compiled_fn_read_local")
}

func TestCompilerImportedSemanticOptionAliasWithShadowedInterfaceStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, MaybeReader, First}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn read_nullable(value: MaybeReader) -> i32 {",
			"  local: MaybeReader = value",
			"  local match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case nil => 0",
			"  }",
			"}",
			"",
			"fn read_direct() -> i32 {",
			"  read_nullable(First {})",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct First {}",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
			"type MaybeReader = Option (Reader i32)",
			"",
		}, "\n"),
	})

	assertShadowedImportedInterfaceFunctionStaysNative(t, result, "__able_compiled_fn_read_nullable")
}

func TestCompilerImportedShadowedInterfaceGenericReturnStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, First}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn id<T>(value: T) -> T { value }",
			"",
			"fn main() -> i32 {",
			"  reader: RemoteReader i32 = First {}",
			"  picked := id(reader)",
			"  picked.read()",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct First {}",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
		}, "\n"),
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var picked runtime.Value",
		"var picked any",
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected imported shadowed interface generic return to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var picked __able_iface_") || !strings.Contains(body, "picked.read()") {
		t.Fatalf("expected imported shadowed interface generic return to stay on the native interface carrier:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_id_spec(value __able_iface_") {
		t.Fatalf("expected imported shadowed interface generic helper to use a native interface signature:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_id_spec(value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_fn_id_spec(value any) (any, *__ableControl)",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected imported shadowed interface generic helper to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerImportedShadowedInterfaceDuplicateGenericUnionCollapsesToNativeCarrier(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, First}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"type Choice T = T | RemoteReader i32",
			"",
			"fn choose<T>(flag: bool, value: T) -> Choice T {",
			"  if flag { value } else { First {} }",
			"}",
			"",
			"fn main() -> i32 {",
			"  reader: RemoteReader i32 = First {}",
			"  picked := choose(true, reader)",
			"  picked.read()",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct First {}",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
		}, "\n"),
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var picked runtime.Value",
		"var picked any",
		"var picked __able_union_",
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected imported shadowed interface duplicate-generic local to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var picked __able_iface_") || !strings.Contains(body, "picked.read()") {
		t.Fatalf("expected imported shadowed interface duplicate-generic local to collapse to the native interface carrier:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_choose_spec(flag bool, value __able_iface_") {
		t.Fatalf("expected imported shadowed interface duplicate-generic helper to use a native interface signature:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_choose_spec(flag bool, value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_fn_choose_spec(flag bool, value any) (any, *__ableControl)",
		"func __able_compiled_fn_choose_spec(flag bool, value __able_iface_Reader_i32) (__able_union_",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected imported shadowed interface duplicate-generic helper to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerImportedShadowedInterfaceGenericBindingPreservesSourcePackage(t *testing.T) {
	program := loadShadowedInterfaceProgramFromFiles(t, map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, First}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"type Choice T = T | RemoteReader i32",
			"",
			"fn choose<T>(flag: bool, value: T) -> Choice T {",
			"  if flag { value } else { First {} }",
			"}",
			"",
			"fn main() -> i32 {",
			"  reader: RemoteReader i32 = First {}",
			"  picked := choose(true, reader)",
			"  picked.read()",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct First {}",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
		}, "\n"),
	})
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

	var chooseInfo *functionInfo
	var mainInfo *functionInfo
	for _, info := range gen.allFunctionInfos() {
		if info == nil || info.Definition == nil || info.Definition.ID == nil {
			continue
		}
		switch info.Definition.ID.Name {
		case "choose":
			chooseInfo = info
		case "main":
			mainInfo = info
		}
	}
	if chooseInfo == nil || mainInfo == nil {
		t.Fatalf("missing function infos: choose=%v main=%v", chooseInfo != nil, mainInfo != nil)
	}

	var readerAssign *ast.AssignmentExpression
	var chooseAssign *ast.AssignmentExpression
	for _, mod := range program.Modules {
		for _, stmt := range mod.AST.Body {
			fn, ok := stmt.(*ast.FunctionDefinition)
			if !ok || fn == nil || fn.ID == nil || fn.ID.Name != "main" || fn.Body == nil {
				continue
			}
			if len(fn.Body.Body) < 2 {
				t.Fatalf("main body too short")
			}
			readerAssign, _ = fn.Body.Body[0].(*ast.AssignmentExpression)
			chooseAssign, _ = fn.Body.Body[1].(*ast.AssignmentExpression)
		}
	}
	if readerAssign == nil || chooseAssign == nil {
		t.Fatalf("missing main assignments")
	}
	call, ok := chooseAssign.Right.(*ast.FunctionCall)
	if !ok || call == nil {
		t.Fatalf("expected choose assignment rhs to be a function call")
	}

	temp := 0
	ctx := &compileContext{
		locals:      make(map[string]paramInfo),
		packageName: mainInfo.Package,
		temps:       &temp,
		returnType:  "int32",
	}
	if _, _, _, ok := gen.compileAssignment(ctx, readerAssign); !ok {
		t.Fatalf("compile reader assignment: %s", ctx.reason)
	}
	binding, ok := ctx.lookup("reader")
	if !ok {
		t.Fatalf("missing reader binding after assignment compile")
	}
	boundPkg := gen.resolvedTypeExprPackage(ctx.packageName, binding.TypeExpr)
	boundType := normalizeTypeExprString(gen, ctx.packageName, binding.TypeExpr)

	bindings, ok := gen.specializedFunctionBindings(ctx, call, chooseInfo, "")
	if !ok || len(bindings) == 0 {
		t.Fatalf("specialized bindings failed: ok=%t len=%d readerType=%q readerPkg=%q", ok, len(bindings), boundType, boundPkg)
	}
	tBinding := bindings["T"]
	if tBinding == nil {
		t.Fatalf("missing T binding: %#v", bindings)
	}
	if gotPkg := gen.resolvedTypeExprPackage(ctx.packageName, tBinding); gotPkg != "demo.remote" {
		t.Fatalf("expected T binding to preserve remote package, got pkg=%q type=%q readerPkg=%q readerType=%q", gotPkg, normalizeTypeExprString(gen, ctx.packageName, tBinding), boundPkg, boundType)
	}
	specialized, ok := gen.ensureSpecializedFunctionInfo(chooseInfo, bindings)
	if !ok || specialized == nil {
		t.Fatalf("ensure specialized function info failed")
	}
	if specialized.TypeBindings["T"] == nil {
		t.Fatalf("specialized function missing T binding: %#v", specialized.TypeBindings)
	}
	if gotPkg := gen.resolvedTypeExprPackage(ctx.packageName, specialized.TypeBindings["T"]); gotPkg != "demo.remote" {
		t.Fatalf("expected specialized function T binding to stay remote, got pkg=%q type=%q", gotPkg, normalizeTypeExprString(gen, ctx.packageName, specialized.TypeBindings["T"]))
	}
	if len(specialized.Params) != 2 {
		t.Fatalf("expected specialized choose params, got %d", len(specialized.Params))
	}
	if gotPkg := gen.resolvedTypeExprPackage(ctx.packageName, specialized.Params[1].TypeExpr); gotPkg != "demo.remote" {
		t.Fatalf("expected specialized choose value param to stay remote, got pkg=%q type=%q goType=%q", gotPkg, normalizeTypeExprString(gen, ctx.packageName, specialized.Params[1].TypeExpr), specialized.Params[1].GoType)
	}
}
