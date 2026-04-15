package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func localJoinedInterfaceExistentialMultimemberPackageFiles() map[string]string {
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
			"struct Tagged T { tag: String, value: T }",
			"",
			"interface Tagger for Self {",
			"  fn tag(self: Self) -> String",
			"  fn tagged<T>(self: Self, value: T) -> Tagged T {",
			"    Tagged T { tag: self.tag(), value: value }",
			"  }",
			"}",
			"",
			"struct First {}",
			"struct Thing { remote: i32 }",
			"struct Box {}",
			"struct Mirror {}",
			"struct Labeler { label: String }",
			"struct Badge { label: String }",
			"struct MyError { message: String }",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
			"impl Echo for Box {",
			"  fn pass<T>(self: Self, value: T) -> T { value }",
			"}",
			"",
			"impl Echo for Mirror {",
			"  fn pass<T>(self: Self, value: T) -> T { value }",
			"}",
			"",
			"impl Tagger for Labeler {",
			"  fn tag(self: Self) -> String { self.label }",
			"}",
			"",
			"impl Tagger for Badge {",
			"  fn tag(self: Self) -> String { self.label }",
			"}",
			"",
			"impl Error for MyError {",
			"  fn message(self: Self) -> String { self.message }",
			"  fn cause(self: Self) -> ?Error { nil }",
			"}",
			"",
			"type Choice3 = (Reader i32) | Echo | String",
			"type Outcome3 = Error | (() -> Thing) | String",
			"",
			"fn main() -> i32 {",
			"  echo: Echo = if true { Box {} } else { Mirror {} }",
			"  tagger: Tagger = if true { Labeler { label: \"L\" } } else { Badge { label: \"B\" } }",
			"  first_input: Choice3 = if true {",
			"    First {}",
			"  } else {",
			"    if false { Box {} } else { \"ok\" }",
			"  }",
			"  second_input: Outcome3 = if true {",
			"    fn() -> Thing { Thing { remote: 7 } }",
			"  } else {",
			"    if false { \"bad\" } else { MyError { message: \"oops\" } }",
			"  }",
			"  first := echo.pass<Choice3>(first_input)",
			"  second := tagger.tagged<Outcome3>(second_input)",
			"  read_part := first match {",
			"    case reader: Reader i32 => reader.read(),",
			"    case nested_echo: Echo => nested_echo.pass<i32>(9),",
			"    case _: String => 1",
			"  }",
			"  build_part := second.value match {",
			"    case build: (() -> Thing) => build().remote,",
			"    case _: String => 2,",
			"    case _: Error => -1",
			"  }",
			"  if second.tag == \"L\" { read_part + build_part } else { 0 }",
			"}",
			"",
		}, "\n"),
	}
}

func TestCompilerLocalJoinedGenericInterfaceMultiMemberAliasCallsStayNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", localJoinedInterfaceExistentialMultimemberPackageFiles())

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	for _, fragment := range []string{
		"var echo __able_iface_Echo",
		"var tagger __able_iface_Tagger",
		"var first_input __able_union_",
		"var second_input __able_union_",
		"var first __able_union_",
		"var second *Tagged_",
		"__able_compiled_iface_Echo_pass_dispatch",
		"__able_compiled_iface_Tagger_tagged_default",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected local joined existential generic-interface body to include %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"var echo runtime.Value",
		"var echo any",
		"var tagger runtime.Value",
		"var tagger any",
		"var first runtime.Value",
		"var first any",
		"var second runtime.Value",
		"var second any",
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
		"__able_try_cast(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected local joined existential generic-interface body to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func localNestedResultInterfaceExistentialGenerator(t *testing.T) *generator {
	t.Helper()
	program := loadShadowedInterfaceProgramFromFiles(t, map[string]string{
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
			"struct First {}",
			"struct Thing { remote: i32 }",
			"struct Box {}",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
			"impl Echo for Box {",
			"  fn pass<T>(self: Self, value: T) -> T { value }",
			"}",
			"",
			"impl Keeper T for Box {",
			"  fn keep(self: Self, value: T) -> T { value }",
			"}",
			"",
			"type Choice3 = (Reader i32) | Echo | String",
			"type Outcome3 = Error | (() -> Thing) | String",
			"",
			"fn main() -> i32 {",
			"  readers: Keeper (!(Choice3)) = Box {}",
			"  builders: Keeper (!(Outcome3)) = Box {}",
			"  choice_value := readers.keep(First {})",
			"  outcome_value := builders.keep(fn() -> Thing { Thing { remote: 7 } })",
			"  left := choice_value match {",
			"    case reader: Reader i32 => reader.read(),",
			"    case echo: Echo => echo.pass<i32>(9),",
			"    case _: String => 1,",
			"    case _: Error => -1",
			"  }",
			"  right := outcome_value match {",
			"    case build: (() -> Thing) => build().remote,",
			"    case _: String => 2,",
			"    case _: Error => -2",
			"  }",
			"  left + right",
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
	return gen
}

func localNestedResultInterfaceExistentialResolvedPackage(t *testing.T, gen *generator) string {
	t.Helper()
	if gen == nil {
		t.Fatal("missing generator")
	}
	for pkgName, aliases := range gen.typeAliases {
		if aliases == nil {
			continue
		}
		if aliases["Choice3"] != nil && aliases["Outcome3"] != nil {
			return pkgName
		}
	}
	t.Fatalf("could not resolve local nested-result existential package from aliases: %#v", gen.typeAliases)
	return ""
}

func TestCompilerNativeInterfaceShapesRecoverLocalNestedResultMultiMemberCarriers(t *testing.T) {
	gen := localNestedResultInterfaceExistentialGenerator(t)
	pkgName := localNestedResultInterfaceExistentialResolvedPackage(t, gen)

	choiceInfo, ok := gen.ensureNativeInterfaceInfo(pkgName, ast.Gen(ast.Ty("Keeper"), ast.NewResultTypeExpression(ast.Ty("Choice3"))))
	if !ok || choiceInfo == nil {
		t.Fatalf("expected Keeper<!(Choice3)> interface info to compile natively")
	}
	outcomeInfo, ok := gen.ensureNativeInterfaceInfo(pkgName, ast.Gen(ast.Ty("Keeper"), ast.NewResultTypeExpression(ast.Ty("Outcome3"))))
	if !ok || outcomeInfo == nil {
		t.Fatalf("expected Keeper<!(Outcome3)> interface info to compile natively")
	}

	choiceKeep := nativeInterfaceMethodByName(choiceInfo, "keep")
	outcomeKeep := nativeInterfaceMethodByName(outcomeInfo, "keep")
	if choiceKeep == nil || outcomeKeep == nil {
		t.Fatalf("missing Keeper.keep methods: choice=%v outcome=%v", choiceKeep != nil, outcomeKeep != nil)
	}
	for label, goType := range map[string]string{
		"choice_param":   choiceKeep.ParamGoTypes[0],
		"choice_return":  choiceKeep.ReturnGoType,
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
