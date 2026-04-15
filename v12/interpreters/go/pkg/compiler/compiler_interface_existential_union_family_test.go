package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func interfaceExistentialUnionFamilyPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"interface Echo for Self {",
			"  fn pass<T>(self: Self, value: T) -> T",
			"}",
			"",
			"interface Keeper <T> for Self {",
			"  fn keep(self: Self, value: T) -> T",
			"}",
			"",
			"type Either = (Reader i32) | Echo",
			"type Outcome = !Either",
			"",
			"struct First {}",
			"struct Second {}",
			"struct Vault {}",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
			"impl Echo for Second {",
			"  fn pass<T>(self: Self, value: T) -> T { value }",
			"}",
			"",
			"impl Keeper T for Vault {",
			"  fn keep(self: Self, value: T) -> T { value }",
			"}",
			"",
			"fn use_union(value: Either) -> i32 {",
			"  value match {",
			"    case reader: Reader i32 => reader.read(),",
			"    case echo: Echo => echo.pass<i32>(7)",
			"  }",
			"}",
			"",
			"fn use_result(value: Outcome) -> i32 {",
			"  value match {",
			"    case reader: Reader i32 => reader.read(),",
			"    case echo: Echo => echo.pass<i32>(7),",
			"    case _: Error => -1",
			"  }",
			"}",
			"",
			"fn main() -> i32 {",
			"  keeper_either: Keeper Either = Vault {}",
			"  keeper_result: Keeper Outcome = Vault {}",
			"  either_value: Either = if true { First {} } else { Second {} }",
			"  result_value: Outcome = if false { First {} } else { Second {} }",
			"  use_union(keeper_either.keep(either_value)) + use_result(keeper_result.keep(result_value))",
			"}",
			"",
		}, "\n"),
	}
}

func interfaceExistentialUnionFamilyGenerator(t *testing.T) *generator {
	t.Helper()
	program := loadShadowedInterfaceProgramFromFiles(t, interfaceExistentialUnionFamilyPackageFiles())
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

func interfaceExistentialUnionFamilyResolvedPackage(t *testing.T, gen *generator) string {
	t.Helper()
	if gen == nil {
		t.Fatal("missing generator")
	}
	for pkgName, aliases := range gen.typeAliases {
		if aliases == nil {
			continue
		}
		if aliases["Either"] != nil && aliases["Outcome"] != nil {
			return pkgName
		}
	}
	t.Fatalf("could not resolve local interface existential union family package from aliases: %#v", gen.typeAliases)
	return ""
}

func TestCompilerInterfaceExistentialUnionFamilyAliasesStayNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", interfaceExistentialUnionFamilyPackageFiles())

	unionBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_use_union")
	for _, fragment := range []string{
		"func __able_compiled_fn_use_union(value runtime.Value)",
		"func __able_compiled_fn_use_union(value any)",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(unionBody, fragment) {
			t.Fatalf("expected interface existential union alias body to avoid %q:\n%s", fragment, unionBody)
		}
	}
	for _, fragment := range []string{
		"value __able_union_",
		"reader.read()",
		"__able_compiled_iface_Echo_pass_dispatch(",
	} {
		if !strings.Contains(unionBody, fragment) {
			t.Fatalf("expected interface existential union alias body to include %q:\n%s", fragment, unionBody)
		}
	}

	resultBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_use_result")
	for _, fragment := range []string{
		"func __able_compiled_fn_use_result(value runtime.Value)",
		"func __able_compiled_fn_use_result(value any)",
		"value __able_union___able_union_",
		"_as___able_union_",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(resultBody, fragment) {
			t.Fatalf("expected interface existential result alias body to avoid %q:\n%s", fragment, resultBody)
		}
	}
	for _, fragment := range []string{
		"value __able_union_",
		"reader.read()",
		"__able_compiled_iface_Echo_pass_dispatch(",
	} {
		if !strings.Contains(resultBody, fragment) {
			t.Fatalf("expected interface existential result alias body to include %q:\n%s", fragment, resultBody)
		}
	}

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	for _, fragment := range []string{
		"var keeper_either __able_iface_Keeper_",
		"var keeper_result __able_iface_Keeper_",
		"var either_value __able_union_",
		"var result_value __able_union_",
		"keeper_either.keep(",
		"keeper_result.keep(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected interface existential family main body to include %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"var keeper_either runtime.Value",
		"var keeper_result runtime.Value",
		"var either_value runtime.Value",
		"var result_value runtime.Value",
		"var keeper_either any",
		"var keeper_result any",
		"var either_value any",
		"var result_value any",
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
		"__able_try_cast(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected interface existential family main body to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerNativeInterfaceShapesRecoverInterfaceExistentialUnionFamilyCarriers(t *testing.T) {
	gen := interfaceExistentialUnionFamilyGenerator(t)
	pkgName := interfaceExistentialUnionFamilyResolvedPackage(t, gen)

	normalizedEither := normalizeTypeExprForPackage(gen, pkgName, ast.Ty("Either"))
	normalizedOutcome := normalizeTypeExprForPackage(gen, pkgName, ast.Ty("Outcome"))
	if typeExpressionToString(normalizedEither) == "Either" {
		t.Fatalf("expected local Either alias to normalize in package %q", pkgName)
	}
	if typeExpressionToString(normalizedOutcome) == "Outcome" {
		t.Fatalf("expected local Outcome alias to normalize in package %q", pkgName)
	}

	eitherInfo, ok := gen.ensureNativeInterfaceInfo(pkgName, ast.Gen(ast.Ty("Keeper"), ast.Ty("Either")))
	if !ok || eitherInfo == nil {
		t.Fatalf("expected Keeper<Either> interface info to compile natively")
	}
	outcomeInfo, ok := gen.ensureNativeInterfaceInfo(pkgName, ast.Gen(ast.Ty("Keeper"), ast.Ty("Outcome")))
	if !ok || outcomeInfo == nil {
		t.Fatalf("expected Keeper<Outcome> interface info to compile natively")
	}

	eitherKeep := nativeInterfaceMethodByName(eitherInfo, "keep")
	outcomeKeep := nativeInterfaceMethodByName(outcomeInfo, "keep")
	if eitherKeep == nil || outcomeKeep == nil {
		t.Fatalf("missing Keeper.keep methods: either=%v outcome=%v", eitherKeep != nil, outcomeKeep != nil)
	}
	for label, goType := range map[string]string{
		"either_param":   eitherKeep.ParamGoTypes[0],
		"either_return":  eitherKeep.ReturnGoType,
		"outcome_param":  outcomeKeep.ParamGoTypes[0],
		"outcome_return": outcomeKeep.ReturnGoType,
	} {
		if goType == "" || goType == "runtime.Value" || goType == "any" {
			t.Fatalf("expected %s to stay on a native carrier, got %q", label, goType)
		}
		if !strings.HasPrefix(goType, "__able_union_") {
			t.Fatalf("expected %s to stay on a native union/result carrier, got %q", label, goType)
		}
	}
}
