package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func nativeInterfaceShapeNestedResultPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, Thing::RemoteThing, Choice3, Outcome3, First}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct Thing { local: i32 }",
			"",
			"interface Keeper <T> for Self {",
			"  fn keep(self: Self, value: T) -> T",
			"}",
			"",
			"struct Box {}",
			"",
			"impl Keeper T for Box {",
			"  fn keep(self: Self, value: T) -> T { value }",
			"}",
			"",
			"fn main() -> i32 {",
			"  readers: Keeper (!(Choice3 (RemoteReader i32))) = Box {}",
			"  reader_value := readers.keep(First {})",
			"  builders: Keeper (!(Outcome3 (() -> RemoteThing))) = Box {}",
			"  build_value := builders.keep(fn() -> RemoteThing { RemoteThing { remote: 7 } })",
			"  read_part := reader_value match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: String => 0,",
			"    case n: i32 => n,",
			"    case _: Error => -1",
			"  }",
			"  build_part := build_value match {",
			"    case build: (() -> RemoteThing) => build().remote,",
			"    case _: String => 0,",
			"    case _: Error => -1",
			"  }",
			"  read_part + build_part",
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
			"type Choice3 T = T | String | i32",
			"type Outcome3 T = Error | T | String",
			"",
		}, "\n"),
	}
}

func TestCompilerNativeInterfaceMethodShapesRecoverNestedResultMultiMemberAliasCarriers(t *testing.T) {
	gen := sharedMapperRecoveryGeneratorFromFiles(t, nativeInterfaceShapeNestedResultPackageFiles())

	readersInfo, ok := gen.ensureNativeInterfaceInfo("demo", ast.Gen(
		ast.Ty("Keeper"),
		ast.NewResultTypeExpression(ast.Gen(ast.Ty("Choice3"), ast.Gen(ast.Ty("RemoteReader"), ast.Ty("i32")))),
	))
	if !ok || readersInfo == nil {
		t.Fatalf("expected Keeper<!(Choice3<RemoteReader i32>)> interface info to compile natively")
	}
	buildersInfo, ok := gen.ensureNativeInterfaceInfo("demo", ast.Gen(
		ast.Ty("Keeper"),
		ast.NewResultTypeExpression(ast.Gen(ast.Ty("Outcome3"), ast.NewFunctionTypeExpression(nil, ast.Ty("RemoteThing")))),
	))
	if !ok || buildersInfo == nil {
		t.Fatalf("expected Keeper<!(Outcome3<() -> RemoteThing>)> interface info to compile natively")
	}

	readMethod := nativeInterfaceMethodByName(readersInfo, "keep")
	buildMethod := nativeInterfaceMethodByName(buildersInfo, "keep")
	if readMethod == nil || buildMethod == nil {
		t.Fatalf("missing Keeper methods: read=%v build=%v", readMethod != nil, buildMethod != nil)
	}
	for label, goType := range map[string]string{
		"read_param":   readMethod.ParamGoTypes[0],
		"read_return":  readMethod.ReturnGoType,
		"build_param":  buildMethod.ParamGoTypes[0],
		"build_return": buildMethod.ReturnGoType,
	} {
		if goType == "" || goType == "runtime.Value" || goType == "any" {
			t.Fatalf("expected %s to stay on a native carrier, got %q", label, goType)
		}
		if !strings.HasPrefix(goType, "__able_union_") {
			t.Fatalf("expected %s to stay on a native union/result carrier, got %q", label, goType)
		}
	}
}

func TestCompilerNativeInterfaceMethodShapesCompileNestedResultMultiMemberAliasCallsNatively(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", nativeInterfaceShapeNestedResultPackageFiles())

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	for _, fragment := range []string{
		"var readers __able_iface_Keeper_",
		"var builders __able_iface_Keeper_",
		"readers.keep(",
		"builders.keep(",
		"var reader_value __able_union_",
		"var build_value __able_union_",
		"reader.read()",
		".Remote",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected nested-result Keeper slice to include %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
		"__able_try_cast(",
		"var readers runtime.Value",
		"var readers any",
		"var builders runtime.Value",
		"var builders any",
		"var reader_value runtime.Value",
		"var reader_value any",
		"var build_value runtime.Value",
		"var build_value any",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected nested-result Keeper slice to avoid %q:\n%s", fragment, mainBody)
		}
	}

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_iface_Keeper_",
		"_variant_runtime_Value",
		"_wrap_runtime_Value",
		"_as_runtime_Value",
		"(value runtime.Value) (runtime.Value, *__ableControl)",
		"(value any) (any, *__ableControl)",
	} {
		if fragment == "type __able_iface_Keeper_" {
			if !strings.Contains(compiledSrc, fragment) {
				t.Fatalf("expected nested-result Keeper native interface signatures to include %q:\n%s", fragment, compiledSrc)
			}
			continue
		}
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected nested-result Keeper native interface signatures to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}
