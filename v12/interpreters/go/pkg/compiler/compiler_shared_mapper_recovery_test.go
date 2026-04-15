package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func sharedMapperRecoveryGeneratorFromFiles(t *testing.T, files map[string]string) *generator {
	t.Helper()
	program := loadShadowedInterfaceProgramFromFiles(t, files)
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
	return gen
}

func TestCompilerSharedMapperRecoversImportedShadowedInterfaceAliasCarriers(t *testing.T) {
	gen := sharedMapperRecoveryGeneratorFromFiles(t, shadowedImportedGenericInterfaceAliasPackageFiles())
	pkgName := gen.entryPackage
	cases := []struct {
		name       string
		expr       ast.TypeExpression
		wantPrefix string
	}{
		{"remote_reader", ast.Gen(ast.Ty("RemoteReader"), ast.Ty("i32")), "__able_iface_"},
		{"choice_remote_reader", ast.Gen(ast.Ty("Choice"), ast.Gen(ast.Ty("RemoteReader"), ast.Ty("i32"))), "__able_union_"},
		{"outcome_remote_reader", ast.Gen(ast.Ty("Outcome"), ast.Gen(ast.Ty("RemoteReader"), ast.Ty("i32"))), "__able_union_"},
	}
	for _, tc := range cases {
		mapped, ok := gen.lowerCarrierTypeInPackage(pkgName, tc.expr)
		if !ok || mapped == "" || mapped == "runtime.Value" || mapped == "any" || !strings.HasPrefix(mapped, tc.wantPrefix) {
			baseName, _ := typeExprBaseName(tc.expr)
			sourcePkg, sourceName := gen.importedSelectorSourceTypeAlias(pkgName, baseName)
			aliasPkg, aliasName, aliasTarget, _, aliasOK := gen.typeAliasTargetForPackage(pkgName, baseName)
			t.Fatalf("%s: expected shared mapper to recover a native carrier with prefix %q, got ok=%t mapped=%q normalized=%q resolvedPkg=%q import=%s.%s alias=%t aliasPkg=%q aliasName=%q aliasTarget=%q staticImports=%#v", tc.name, tc.wantPrefix, ok, mapped, typeExpressionToString(normalizeTypeExprForPackage(gen, pkgName, tc.expr)), gen.resolvedTypeExprPackage(pkgName, normalizeTypeExprForPackage(gen, pkgName, tc.expr)), sourcePkg, sourceName, aliasOK, aliasPkg, aliasName, typeExpressionToString(aliasTarget), gen.staticImports)
		}
	}
}

func TestCompilerSharedMapperRecoversImportedShadowedCallableAliasCarriers(t *testing.T) {
	gen := sharedMapperRecoveryGeneratorFromFiles(t, shadowedImportedCallableAliasPackageFiles())
	pkgName := gen.entryPackage
	callableExpr := ast.NewFunctionTypeExpression(nil, ast.Ty("RemoteThing"))
	cases := []struct {
		name       string
		expr       ast.TypeExpression
		wantPrefix string
	}{
		{"callable_remote_thing", callableExpr, "__able_fn_"},
		{"choice_callable_remote_thing", ast.Ty("Choice"), "__able_union_"},
		{"outcome_callable_remote_thing", ast.Ty("Outcome"), "__able_union_"},
	}
	for _, tc := range cases {
		mapped, ok := gen.lowerCarrierTypeInPackage(pkgName, tc.expr)
		if !ok || mapped == "" || mapped == "runtime.Value" || mapped == "any" || !strings.HasPrefix(mapped, tc.wantPrefix) {
			baseName, _ := typeExprBaseName(tc.expr)
			sourcePkg, sourceName := gen.importedSelectorSourceTypeAlias(pkgName, baseName)
			aliasPkg, aliasName, aliasTarget, _, aliasOK := gen.typeAliasTargetForPackage(pkgName, baseName)
			t.Fatalf("%s: expected shared mapper to recover a native carrier with prefix %q, got ok=%t mapped=%q normalized=%q resolvedPkg=%q import=%s.%s alias=%t aliasPkg=%q aliasName=%q aliasTarget=%q staticImports=%#v", tc.name, tc.wantPrefix, ok, mapped, typeExpressionToString(normalizeTypeExprForPackage(gen, pkgName, tc.expr)), gen.resolvedTypeExprPackage(pkgName, normalizeTypeExprForPackage(gen, pkgName, tc.expr)), sourcePkg, sourceName, aliasOK, aliasPkg, aliasName, typeExpressionToString(aliasTarget), gen.staticImports)
		}
	}
}
