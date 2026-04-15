package compiler

import (
	"regexp"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func requireNativeCallableCarrierPackage(t *testing.T, gen *generator, ctx *compileContext, goType string, wantPkg string) *nativeCallableInfo {
	t.Helper()
	info := gen.nativeCallableInfoForGoType(goType)
	if info == nil {
		t.Fatalf("expected native callable info for %q", goType)
	}
	if got := strings.TrimSpace(info.PackageName); got != wantPkg {
		t.Fatalf("expected callable carrier %q to keep package %q, got %q", goType, wantPkg, got)
	}
	if info.TypeExpr == nil {
		t.Fatalf("expected callable carrier %q to retain a type expression", goType)
	}
	if got := gen.resolvedTypeExprPackage(ctx.packageName, info.TypeExpr); got != wantPkg {
		t.Fatalf("expected callable carrier %q type expression to resolve to %q, got %q", goType, wantPkg, got)
	}
	if info.ReturnTypeExpr == nil {
		t.Fatalf("expected callable carrier %q to retain a return type expression", goType)
	}
	if got := gen.resolvedTypeExprPackage(ctx.packageName, info.ReturnTypeExpr); got != wantPkg {
		t.Fatalf("expected callable carrier %q return type to resolve to %q, got %q", goType, wantPkg, got)
	}
	return info
}

func requirePickedCarrier(t *testing.T, body string) string {
	t.Helper()
	match := regexp.MustCompile(`var picked ([A-Za-z0-9_]+) =`).FindStringSubmatch(body)
	if len(match) != 2 {
		t.Fatalf("expected compiled body to bind picked to a native carrier:\n%s", body)
	}
	return match[1]
}

func requireSpecializedIDHelperForCarrier(t *testing.T, compiledSrc string, carrier string) {
	t.Helper()
	pattern := `func __able_compiled_fn_id_spec(?:_[A-Za-z0-9]+)?\(value ` +
		regexp.QuoteMeta(carrier) +
		`\) \(` + regexp.QuoteMeta(carrier) + `, \*__ableControl\)`
	if !regexp.MustCompile(pattern).MatchString(compiledSrc) {
		t.Fatalf("expected specialized id helper to use native carrier %q:\n%s", carrier, compiledSrc)
	}
}

func shadowedImportedNullableSpecializationPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, Thing::RemoteThing, MaybeReader, MaybeBuilder, First}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct Thing { local: i32 }",
			"",
			"fn id<T>(value: T) -> T { value }",
			"",
			"fn read_nullable_interface() -> i32 {",
			"  local: MaybeReader = First {}",
			"  picked := id(local)",
			"  picked match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case nil => 0",
			"  }",
			"}",
			"",
			"fn read_nullable_callable() -> i32 {",
			"  local: MaybeBuilder = fn() -> RemoteThing { RemoteThing { remote: 9 } }",
			"  picked := id(local)",
			"  picked match {",
			"    case build: (() -> RemoteThing) => build().remote,",
			"    case nil => 0",
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
			"struct Thing { remote: i32 }",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
			"type MaybeReader = ?(Reader i32)",
			"type MaybeBuilder = Option (() -> Thing)",
			"",
		}, "\n"),
	}
}

func shadowedImportedNullableCallableParts(t *testing.T) (*generator, *functionInfo, *compileContext, string, *ast.TypedPattern, []ast.Statement) {
	t.Helper()
	program := loadShadowedInterfaceProgramFromFiles(t, shadowedImportedNullableSpecializationPackageFiles())
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
		if candidate == nil || candidate.Definition == nil || candidate.Definition.ID == nil || candidate.Definition.ID.Name != "read_nullable_callable" {
			continue
		}
		info = candidate
		break
	}
	if info == nil {
		t.Fatalf("missing read_nullable_callable function info")
	}
	var statements []ast.Statement
	for _, mod := range program.Modules {
		for _, stmt := range mod.AST.Body {
			fn, ok := stmt.(*ast.FunctionDefinition)
			if !ok || fn == nil || fn.ID == nil || fn.ID.Name != "read_nullable_callable" || fn.Body == nil {
				continue
			}
			statements = fn.Body.Body
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
		t.Fatalf("missing read_nullable_callable typed pattern")
	}
	ctx := newCompileContext(gen, info, gen.functions[info.Package], gen.overloads[info.Package], info.Package, gen.compileContextGenericNames(info))
	if len(info.Params) != 0 {
		t.Fatalf("expected read_nullable_callable to have no params")
	}
	pickedTypeExpr := ast.Ty("MaybeBuilder")
	pickedGoType, ok := gen.lowerCarrierType(ctx, pickedTypeExpr)
	if !ok || pickedGoType == "" {
		t.Fatalf("could not lower picked carrier")
	}
	ctx.setLocalBinding("picked", paramInfo{Name: "picked", GoName: "picked", GoType: pickedGoType, TypeExpr: pickedTypeExpr})
	ctx.matchSubjectTypeExpr = pickedTypeExpr
	if len(statements) == 0 {
		t.Fatalf("missing read_nullable_callable statements")
	}
	return gen, info, ctx, pickedGoType, pattern, statements
}

func TestCompilerImportedShadowedNullableCallablePatternBindingStaysNative(t *testing.T) {
	gen, info, ctx, subjectType, pattern, statements := shadowedImportedNullableCallableParts(t)
	ctx.setLocalBinding("local", paramInfo{
		Name:     "local",
		GoName:   "local",
		GoType:   subjectType,
		TypeExpr: ast.Ty("MaybeBuilder"),
	})
	normalizedMaybeBuilder := normalizeTypeExprForPackage(gen, ctx.packageName, ast.Ty("MaybeBuilder"))
	resolvedMaybeBuilderPkg := gen.resolvedTypeExprPackage(ctx.packageName, normalizedMaybeBuilder)
	callLines, callExpr, callType, ok := gen.compileExprLines(ctx, ast.NewFunctionCall(ast.NewIdentifier("id"), []ast.Expression{ast.NewIdentifier("local")}, nil, false), "")
	if !ok {
		t.Fatalf("compileExprLines(id(local)) failed: %s", ctx.reason)
	}
	if len(callLines) == 0 || callExpr == "" {
		t.Fatalf("expected id(local) to compile to a concrete callable expression")
	}
	if callType != subjectType {
		t.Fatalf("expected id(local) to keep native callable carrier %q, got %q (MaybeBuilder normalized=%q resolvedPkg=%q)", subjectType, callType, typeExpressionToString(normalizedMaybeBuilder), resolvedMaybeBuilderPkg)
	}
	mapped, ok := gen.lowerCarrierType(ctx, pattern.TypeAnnotation)
	if !ok {
		t.Fatalf("could not lower callable pattern type")
	}
	if mapped != subjectType {
		t.Fatalf("expected nullable callable pattern type to map to the picked native carrier, got pattern=%q subject=%q (pattern=%q)", mapped, subjectType, typeExpressionToString(pattern.TypeAnnotation))
	}
	lines, ok := gen.compileMatchPatternBindings(ctx, pattern, "picked", subjectType)
	if !ok {
		t.Fatalf("compileMatchPatternBindings failed: %s", ctx.reason)
	}
	if len(lines) == 0 {
		t.Fatalf("expected nullable callable typed pattern to bind a native callable local")
	}
	binding, ok := ctx.lookup("build")
	if !ok {
		t.Fatalf("expected nullable callable typed pattern to bind build")
	}
	if binding.GoType != subjectType {
		t.Fatalf("expected build binding to keep native callable carrier %q, got %q", subjectType, binding.GoType)
	}
	requireNativeCallableCarrierPackage(t, gen, ctx, binding.GoType, "demo.remote")

	gen, info, ctx, _, pattern, statements = shadowedImportedNullableCallableParts(t)
	if len(statements) < 3 {
		t.Fatalf("expected read_nullable_callable to contain assignment + match statements")
	}
	fullCtx := newCompileContext(gen, info, gen.functions[ctx.packageName], gen.overloads[ctx.packageName], ctx.packageName, gen.compileContextGenericNames(info))
	if _, ok := gen.compileStatement(fullCtx, statements[0]); !ok {
		t.Fatalf("compiling local statement failed: %s", fullCtx.reason)
	}
	localBinding, ok := fullCtx.lookup("local")
	if !ok {
		t.Fatalf("expected local binding after compiling local statement")
	}
	localInfo := requireNativeCallableCarrierPackage(t, gen, fullCtx, localBinding.GoType, "demo.remote")
	prePickedCallLines, prePickedCallExpr, prePickedCallType, prePickedCallOK := gen.compileExprLines(fullCtx, statements[1].(*ast.AssignmentExpression).Right, "")
	if !prePickedCallOK || prePickedCallExpr == "" || len(prePickedCallLines) == 0 {
		t.Fatalf("pre-picked compileExprLines failed: ok=%t expr=%q type=%q reason=%s", prePickedCallOK, prePickedCallExpr, prePickedCallType, fullCtx.reason)
	}
	if _, ok := gen.compileStatement(fullCtx, statements[1]); !ok {
		t.Fatalf("compiling picked statement failed: %s", fullCtx.reason)
	}
	pickedBinding, ok := fullCtx.lookup("picked")
	if !ok {
		t.Fatalf("expected picked binding after compiling picked assignment")
	}
	if pickedBinding.GoType != localBinding.GoType {
		t.Fatalf("expected generic specialization to preserve the remote native callable carrier, got local=%q picked=%q", localBinding.GoType, pickedBinding.GoType)
	}
	pickedInfo := requireNativeCallableCarrierPackage(t, gen, fullCtx, pickedBinding.GoType, "demo.remote")
	fullMapped, ok := gen.lowerCarrierType(fullCtx, pattern.TypeAnnotation)
	if !ok {
		t.Fatalf("could not lower callable pattern type in full context")
	}
	if fullMapped != pickedBinding.GoType {
		t.Fatalf("expected nullable callable typed pattern to resolve to the specialized picked carrier, got pattern=%q picked=%q", fullMapped, pickedBinding.GoType)
	}
	bindingLines, ok := gen.compileMatchPatternBindings(fullCtx, pattern, "picked", pickedBinding.GoType)
	if !ok {
		t.Fatalf("compileMatchPatternBindings in full context failed: %s", fullCtx.reason)
	}
	if len(bindingLines) == 0 {
		t.Fatalf("expected nullable callable typed pattern to bind a native local in full context")
	}
	buildBinding, ok := fullCtx.lookup("build")
	if !ok {
		t.Fatalf("expected full-context nullable callable typed pattern to bind build")
	}
	if buildBinding.GoType != pickedBinding.GoType {
		t.Fatalf("expected full-context build binding to keep picked carrier %q, got %q", pickedBinding.GoType, buildBinding.GoType)
	}
	requireNativeCallableCarrierPackage(t, gen, fullCtx, buildBinding.GoType, "demo.remote")
	if localInfo.GoType != pickedInfo.GoType {
		t.Fatalf("expected local and picked callable info to agree, got local=%q picked=%q", localInfo.GoType, pickedInfo.GoType)
	}
}

func TestCompilerSpecializedGenericImportedShadowedNullableInterfaceAliasReturnStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedNullableSpecializationPackageFiles())

	body, ok := findCompiledFunction(result, "__able_compiled_fn_read_nullable_interface")
	if !ok {
		t.Fatalf("could not find compiled read_nullable_interface function")
	}
	for _, fragment := range []string{
		"var picked runtime.Value",
		"var picked any",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected specialized generic imported shadowed nullable interface alias return to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "reader.read()") {
		t.Fatalf("expected specialized generic imported shadowed nullable interface alias return to keep direct native interface dispatch:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	requireSpecializedIDHelperForCarrier(t, compiledSrc, requirePickedCarrier(t, body))
}

func TestCompilerSpecializedGenericImportedShadowedNullableCallableAliasReturnStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedNullableSpecializationPackageFiles())

	body, ok := findCompiledFunction(result, "__able_compiled_fn_read_nullable_callable")
	if !ok {
		t.Fatalf("could not find compiled read_nullable_callable function")
	}
	for _, fragment := range []string{
		"var picked runtime.Value",
		"var picked any",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_call_value(",
		"__able_member_get(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected specialized generic imported shadowed nullable callable alias return to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, ".Remote") {
		t.Fatalf("expected specialized generic imported shadowed nullable callable alias return to keep direct native field access:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	requireSpecializedIDHelperForCarrier(t, compiledSrc, requirePickedCarrier(t, body))
}
