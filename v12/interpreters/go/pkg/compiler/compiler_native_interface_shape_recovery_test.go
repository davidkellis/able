package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func nativeInterfaceShapeRecoveryPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, Thing::RemoteThing, Choice, Outcome, First}",
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
			"  readers: Keeper (Choice (RemoteReader i32)) = Box {}",
			"  reader_value := readers.keep(First {})",
			"  builders: Keeper (Outcome (() -> RemoteThing)) = Box {}",
			"  build_value := builders.keep(fn() -> RemoteThing { RemoteThing { remote: 7 } })",
			"  read_part := read_value match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: String => 0",
			"  }",
			"  build_part := build_value match {",
			"    case build: (() -> RemoteThing) => build().remote,",
			"    case _: Error => 0",
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
			"type Choice T = T | String",
			"type Outcome T = Error | T",
			"",
		}, "\n"),
	}
}

func nativeInterfaceShapeRecoveryGenerator(t *testing.T) *generator {
	t.Helper()
	program := loadShadowedInterfaceProgramFromFiles(t, nativeInterfaceShapeRecoveryPackageFiles())
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

func nativeInterfaceMethodByName(info *nativeInterfaceInfo, name string) *nativeInterfaceMethod {
	if info == nil {
		return nil
	}
	for _, method := range info.Methods {
		if method != nil && method.Name == name {
			return method
		}
	}
	return nil
}

func TestCompilerNativeInterfaceMethodShapesRecoverImportedShadowedAliasCarriers(t *testing.T) {
	gen := nativeInterfaceShapeRecoveryGenerator(t)
	readersInfo, ok := gen.ensureNativeInterfaceInfo("demo", ast.Gen(ast.Ty("Keeper"), ast.Gen(ast.Ty("Choice"), ast.Gen(ast.Ty("RemoteReader"), ast.Ty("i32")))))
	if !ok || readersInfo == nil {
		t.Fatalf("expected Keeper<Choice<RemoteReader i32>> interface info to compile natively")
	}
	buildersInfo, ok := gen.ensureNativeInterfaceInfo("demo", ast.Gen(ast.Ty("Keeper"), ast.NewResultTypeExpression(ast.NewFunctionTypeExpression(nil, ast.Ty("RemoteThing")))))
	if !ok || buildersInfo == nil {
		t.Fatalf("expected Keeper<Outcome<() -> RemoteThing>> interface info to compile natively")
	}

	readMethod := nativeInterfaceMethodByName(readersInfo, "keep")
	buildMethod := nativeInterfaceMethodByName(buildersInfo, "keep")
	if readMethod == nil || buildMethod == nil {
		t.Fatalf("missing Keeper methods: read=%v build=%v", readMethod != nil, buildMethod != nil)
	}
	if len(readMethod.ParamGoTypes) != 1 || len(buildMethod.ParamGoTypes) != 1 {
		t.Fatalf("expected Keeper.keep to have one param, got read=%d build=%d", len(readMethod.ParamGoTypes), len(buildMethod.ParamGoTypes))
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
	}
	if !strings.HasPrefix(readMethod.ParamGoTypes[0], "__able_union_") {
		t.Fatalf("expected Keeper<Choice<RemoteReader>> param to stay on a native union carrier, got %q", readMethod.ParamGoTypes[0])
	}
	if !strings.HasPrefix(readMethod.ReturnGoType, "__able_union_") {
		t.Fatalf("expected Keeper<Choice<RemoteReader>> return to stay on a native union carrier, got %q", readMethod.ReturnGoType)
	}
	if !strings.HasPrefix(buildMethod.ParamGoTypes[0], "__able_union_") {
		t.Fatalf("expected Keeper<Outcome<() -> RemoteThing>> param to stay on a native result carrier, got %q", buildMethod.ParamGoTypes[0])
	}
	if !strings.HasPrefix(buildMethod.ReturnGoType, "__able_union_") {
		t.Fatalf("expected Keeper<Outcome<() -> RemoteThing>> return to stay on a native result/callable carrier, got %q", buildMethod.ReturnGoType)
	}
}

func TestCompilerNativeInterfaceMethodShapesCompileImportedShadowedAliasCallsNatively(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", nativeInterfaceShapeRecoveryPackageFiles())

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	for _, fragment := range []string{
		"var readers __able_iface_Keeper_",
		"var builders __able_iface_Keeper_",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected Keeper locals to stay on native interface carriers via %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"readers.keep(",
		"builders.keep(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected Keeper calls to dispatch through %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
		"var readers runtime.Value",
		"var readers any",
		"var builders runtime.Value",
		"var builders any",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected Keeper calls to avoid %q:\n%s", fragment, mainBody)
		}
	}

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type __able_iface_Keeper_",
		"(value runtime.Value) (runtime.Value, *__ableControl)",
		"(value any) (any, *__ableControl)",
	} {
		if fragment == "type __able_iface_Keeper_" {
			if !strings.Contains(compiledSrc, fragment) {
				t.Fatalf("expected Keeper native interface signatures to include %q:\n%s", fragment, compiledSrc)
			}
			continue
		}
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected Keeper native interface signatures to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}
