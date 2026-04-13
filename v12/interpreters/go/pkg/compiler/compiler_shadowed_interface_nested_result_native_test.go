package compiler

import (
	"strconv"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func shadowedImportedNestedResultInterfaceAliasPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, Outcome}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn read_nested(value: !(Outcome (RemoteReader i32))) -> i32 {",
			"  value match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: Error => 0",
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
			"type Outcome T = Result T",
			"",
		}, "\n"),
	}
}

func shadowedImportedNestedOptionInterfaceAliasPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, MaybeReader}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn read_nested(value: !(MaybeReader (RemoteReader i32))) -> i32 {",
			"  value match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case nil => -1,",
			"    case _: Error => 0",
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
			"type MaybeReader T = Option T",
			"",
		}, "\n"),
	}
}

func shadowedImportedNestedUnionInterfaceAliasPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, Choice}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn read_nested(value: !(Choice (RemoteReader i32))) -> i32 {",
			"  value match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: String => -1,",
			"    case _: Error => 0",
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
			"type Choice T = T | String",
			"",
		}, "\n"),
	}
}

func TestCompilerImportedSemanticResultAliasNestedResultWithShadowedInterfaceStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedNestedResultInterfaceAliasPackageFiles())

	body, ok := findCompiledFunction(result, "__able_compiled_fn_read_nested")
	if !ok {
		t.Fatalf("could not find compiled read_nested function")
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_read_nested(value any)",
		"func __able_compiled_fn_read_nested(value runtime.Value)",
		"value __able_union___able_union_",
		"_as___able_union_",
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected nested imported semantic Result alias to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "reader.read()") {
		t.Fatalf("expected nested imported semantic Result alias to keep direct native interface dispatch:\n%s", body)
	}
}

func TestCompilerImportedSemanticResultAliasNestedResultWithShadowedInterfaceResolvesNativeCarrier(t *testing.T) {
	program := loadShadowedInterfaceProgramFromFiles(t, shadowedImportedNestedResultInterfaceAliasPackageFiles())

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
		if candidate.Definition.ID.Name == "read_nested" {
			info = candidate
			break
		}
	}
	if info == nil || len(info.Params) == 0 {
		t.Fatalf("missing read_nested function info")
	}
	ctx := &compileContext{packageName: info.Package}
	normalized := gen.lowerNormalizedTypeExpr(ctx, info.Params[0].TypeExpr)
	mapped, mappedOK := gen.lowerCarrierType(ctx, info.Params[0].TypeExpr)
	before := info.Params[0].GoType
	gen.refreshRepresentableFunctionInfo(info)
	afterRefresh := info.Params[0].GoType
	gen.discardRedundantImplFallbackSpecializations()
	appendDynamicFeatureWarnings(gen, report)
	_ = gen.collectFallbacks()
	afterCompilePasses := info.Params[0].GoType
	if program != nil && program.Entry != nil && len(program.Entry.Files) > 0 {
		gen.opts.EntryPath = program.Entry.Files[0]
	}
	gen.opts.EmitMain = true
	files, err := gen.render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	afterRender := info.Params[0].GoType
	compiledSrc := string(files["compiled.go"])
	freshProgram := loadShadowedInterfaceProgramFromFiles(t, shadowedImportedNestedResultInterfaceAliasPackageFiles())
	apiResult, err := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
		EmitMain:           true,
		EntryPath:          freshProgram.Entry.Files[0],
	}).Compile(freshProgram)
	if err != nil {
		t.Fatalf("api compile: %v", err)
	}
	apiCompiledSrc := string(apiResult.Files["compiled.go"])
	if !mappedOK || mapped == "" || mapped == "any" || mapped == "runtime.Value" ||
		afterRefresh == "" || afterRefresh == "any" || afterRefresh == "runtime.Value" ||
		afterCompilePasses == "" || afterCompilePasses == "any" || afterCompilePasses == "runtime.Value" ||
		strings.Contains(compiledSrc, "func __able_compiled_fn_read_nested(value any)") ||
		strings.Contains(apiCompiledSrc, "func __able_compiled_fn_read_nested(value any)") {
		t.Fatalf("expected nested imported semantic Result alias to resolve a native carrier: package=%q mapped=%q ok=%t before=%q afterRefresh=%q afterCompilePasses=%q afterRender=%q normalized=%q resolvedPkg=%q raw=%q\ndirect compiled:\n%s\n\napi compiled:\n%s", info.Package, mapped, mappedOK, before, afterRefresh, afterCompilePasses, afterRender, typeExpressionToString(normalized), gen.resolvedTypeExprPackage(ctx.packageName, normalized), typeExpressionToString(info.Params[0].TypeExpr), compiledSrc, apiCompiledSrc)
	}
}

func TestCompilerImportedSemanticResultAliasNestedResultWithShadowedInterfaceNormalizesInnerAlias(t *testing.T) {
	program := loadShadowedInterfaceProgramFromFiles(t, shadowedImportedNestedResultInterfaceAliasPackageFiles())

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
		if candidate.Definition.ID.Name == "read_nested" {
			info = candidate
			break
		}
	}
	if info == nil || len(info.Params) == 0 || info.Params[0].TypeExpr == nil {
		t.Fatalf("missing read_nested param type expr")
	}
	resultExpr, ok := info.Params[0].TypeExpr.(*ast.ResultTypeExpression)
	if !ok || resultExpr == nil || resultExpr.InnerType == nil {
		t.Fatalf("expected outer result type expr, got %T %q", info.Params[0].TypeExpr, typeExpressionToString(info.Params[0].TypeExpr))
	}
	expandedInner := gen.expandTypeAliasForPackage(info.Package, resultExpr.InnerType)
	normalizedInner := normalizeTypeExprForPackage(gen, info.Package, resultExpr.InnerType)
	if typeExpressionToString(expandedInner) == typeExpressionToString(resultExpr.InnerType) || strings.Contains(typeExpressionToString(normalizedInner), "Outcome") {
		sourcePkg, sourceName := gen.importedSelectorSourceTypeAlias(info.Package, "Outcome")
		aliasPkg, aliasName, aliasTarget, params, aliasOK := gen.typeAliasTargetForPackage(info.Package, "Outcome")
		t.Fatalf("expected inner imported generic result alias to expand: package=%q raw=%q expanded=%q normalized=%q import=%s.%s aliasOK=%t aliasPkg=%q aliasName=%q aliasTarget=%q params=%d staticImports=%#v", info.Package, typeExpressionToString(resultExpr.InnerType), typeExpressionToString(expandedInner), typeExpressionToString(normalizedInner), sourcePkg, sourceName, aliasOK, aliasPkg, aliasName, typeExpressionToString(aliasTarget), len(params), gen.staticImports)
	}
}

func TestCompilerImportedSemanticResultAliasNestedResultWithShadowedInterfaceResultMembersStayNative(t *testing.T) {
	program := loadShadowedInterfaceProgramFromFiles(t, shadowedImportedNestedResultInterfaceAliasPackageFiles())

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
		if candidate.Definition.ID.Name == "read_nested" {
			info = candidate
			break
		}
	}
	if info == nil || len(info.Params) == 0 || info.Params[0].TypeExpr == nil {
		t.Fatalf("missing read_nested param type expr")
	}
	resultExpr, ok := info.Params[0].TypeExpr.(*ast.ResultTypeExpression)
	if !ok || resultExpr == nil {
		t.Fatalf("expected outer result type expr, got %T %q", info.Params[0].TypeExpr, typeExpressionToString(info.Params[0].TypeExpr))
	}
	resultPkg, members, ok := gen.resultUnionMembersInPackage(info.Package, resultExpr)
	if !ok || len(members) == 0 {
		t.Fatalf("expected nested result members")
	}
	mapper := NewTypeMapper(gen, resultPkg)
	parts := make([]string, 0, len(members))
	for _, member := range members {
		mapped, mappedOK := mapper.Map(member)
		parts = append(parts, typeExpressionToString(member)+"@"+gen.resolvedTypeExprPackage(resultPkg, member)+"="+mapped+" ok="+strconv.FormatBool(mappedOK))
	}
	if _, ok := gen.ensureNativeResultUnionInfo(info.Package, resultExpr); !ok {
		t.Fatalf("expected nested result members to stay native: pkg=%q normalized=%q members=%v", resultPkg, typeExpressionToString(normalizeTypeExprForPackage(gen, info.Package, resultExpr)), parts)
	}
}

func TestCompilerImportedSemanticOptionAliasNestedResultWithShadowedInterfaceStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedNestedOptionInterfaceAliasPackageFiles())

	body, ok := findCompiledFunction(result, "__able_compiled_fn_read_nested")
	if !ok {
		t.Fatalf("could not find compiled read_nested function")
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_read_nested(value any)",
		"func __able_compiled_fn_read_nested(value runtime.Value)",
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected nested imported semantic Option alias to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "reader.read()") {
		t.Fatalf("expected nested imported semantic Option alias to keep direct native interface dispatch:\n%s", body)
	}
}

func TestCompilerImportedGenericUnionAliasNestedResultWithShadowedInterfaceStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedNestedUnionInterfaceAliasPackageFiles())

	body, ok := findCompiledFunction(result, "__able_compiled_fn_read_nested")
	if !ok {
		t.Fatalf("could not find compiled read_nested function")
	}
	for _, fragment := range []string{
		"func __able_compiled_fn_read_nested(value any)",
		"func __able_compiled_fn_read_nested(value runtime.Value)",
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected nested imported generic union alias to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "reader.read()") {
		t.Fatalf("expected nested imported generic union alias to keep direct native interface dispatch:\n%s", body)
	}
}
