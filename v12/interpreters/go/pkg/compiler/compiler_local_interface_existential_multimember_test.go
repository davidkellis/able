package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func localInterfaceExistentialMultimemberPackageFiles() map[string]string {
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
			"interface Keeper <T> for Self {",
			"  fn keep(self: Self, value: T) -> T",
			"}",
			"",
			"struct First {}",
			"struct Thing { remote: i32 }",
			"struct Box {}",
			"struct Labeler { label: String }",
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
			"impl Tagger for Labeler {",
			"  fn tag(self: Self) -> String { self.label }",
			"}",
			"",
			"impl Keeper T for Box {",
			"  fn keep(self: Self, value: T) -> T { value }",
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
			"fn use_choice(value: Choice3) -> i32 {",
			"  value match {",
			"    case reader: Reader i32 => reader.read(),",
			"    case echo: Echo => echo.pass<i32>(9),",
			"    case _: String => 1",
			"  }",
			"}",
			"",
			"fn use_outcome(value: Outcome3) -> i32 {",
			"  value match {",
			"    case build: (() -> Thing) => build().remote,",
			"    case _: String => 2,",
			"    case _: Error => -1",
			"  }",
			"}",
			"",
			"fn main() -> i32 {",
			"  mirror: Echo = Box {}",
			"  tagger: Tagger = Labeler { label: \"L\" }",
			"  keeper_choice: Keeper Choice3 = Box {}",
			"  keeper_outcome: Keeper Outcome3 = Box {}",
			"  choice_input: Choice3 = if true {",
			"    First {}",
			"  } else {",
			"    if false { Box {} } else { \"ok\" }",
			"  }",
			"  outcome_input: Outcome3 = if true {",
			"    fn() -> Thing { Thing { remote: 7 } }",
			"  } else {",
			"    if false { \"bad\" } else { MyError { message: \"oops\" } }",
			"  }",
			"  first := mirror.pass<Choice3>(choice_input)",
			"  second := tagger.tagged<Outcome3>(outcome_input)",
			"  kept_choice := keeper_choice.keep(first)",
			"  kept_outcome := keeper_outcome.keep(second.value)",
			"  if second.tag == \"L\" {",
			"    use_choice(kept_choice) + use_outcome(kept_outcome)",
			"  } else {",
			"    0",
			"  }",
			"}",
			"",
		}, "\n"),
	}
}

func localInterfaceExistentialMultimemberGenerator(t *testing.T) *generator {
	t.Helper()
	program := loadShadowedInterfaceProgramFromFiles(t, localInterfaceExistentialMultimemberPackageFiles())
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

func localInterfaceExistentialMultimemberResolvedPackage(t *testing.T, gen *generator) string {
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
	t.Fatalf("could not resolve local multi-member interface existential package from aliases: %#v", gen.typeAliases)
	return ""
}

func TestCompilerLocalInterfaceExistentialMultiMemberFamilyStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", localInterfaceExistentialMultimemberPackageFiles())

	choiceBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_use_choice")
	for _, fragment := range []string{
		"func __able_compiled_fn_use_choice(value runtime.Value)",
		"func __able_compiled_fn_use_choice(value any)",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(choiceBody, fragment) {
			t.Fatalf("expected local multi-member interface existential choice body to avoid %q:\n%s", fragment, choiceBody)
		}
	}
	for _, fragment := range []string{
		"value __able_union_",
		"reader.read()",
		"__able_compiled_iface_Echo_pass_dispatch",
	} {
		if !strings.Contains(choiceBody, fragment) {
			t.Fatalf("expected local multi-member interface existential choice body to include %q:\n%s", fragment, choiceBody)
		}
	}

	outcomeBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_use_outcome")
	for _, fragment := range []string{
		"func __able_compiled_fn_use_outcome(value runtime.Value)",
		"func __able_compiled_fn_use_outcome(value any)",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(outcomeBody, fragment) {
			t.Fatalf("expected local multi-member interface existential outcome body to avoid %q:\n%s", fragment, outcomeBody)
		}
	}
	for _, fragment := range []string{
		"value __able_union_",
		"var build __able_fn_",
		".Remote",
	} {
		if !strings.Contains(outcomeBody, fragment) {
			t.Fatalf("expected local multi-member interface existential outcome body to include %q:\n%s", fragment, outcomeBody)
		}
	}

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	for _, fragment := range []string{
		"var mirror __able_iface_Echo = __able_iface_Echo_wrap_ptr_Box(",
		"var tagger __able_iface_Tagger = __able_iface_Tagger_wrap_ptr_Labeler(",
		"var keeper_choice __able_iface_Keeper_",
		"var keeper_outcome __able_iface_Keeper_",
		"var choice_input __able_union_",
		"var outcome_input __able_union_",
		"var first __able_union_",
		"var second *Tagged_",
		"__able_compiled_iface_Echo_pass_dispatch",
		"__able_compiled_iface_Tagger_tagged_default",
		"keeper_choice.keep(",
		"keeper_outcome.keep(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected local multi-member interface existential main body to include %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"var mirror runtime.Value",
		"var mirror any",
		"var tagger runtime.Value",
		"var tagger any",
		"var keeper_choice runtime.Value",
		"var keeper_choice any",
		"var keeper_outcome runtime.Value",
		"var keeper_outcome any",
		"var choice_input runtime.Value",
		"var choice_input any",
		"var outcome_input runtime.Value",
		"var outcome_input any",
		"var first runtime.Value",
		"var first any",
		"var second runtime.Value",
		"var second any",
		"var kept_choice runtime.Value",
		"var kept_choice any",
		"var kept_outcome runtime.Value",
		"var kept_outcome any",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_method_call_node(",
		"__able_call_value(",
		"_variant_runtime_Value",
		"_wrap_runtime_Value",
		"_as_runtime_Value",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected local multi-member interface existential main body to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerNativeInterfaceShapesRecoverLocalInterfaceExistentialMultiMemberCarriers(t *testing.T) {
	gen := localInterfaceExistentialMultimemberGenerator(t)
	pkgName := localInterfaceExistentialMultimemberResolvedPackage(t, gen)

	normalizedChoice := normalizeTypeExprForPackage(gen, pkgName, ast.Ty("Choice3"))
	normalizedOutcome := normalizeTypeExprForPackage(gen, pkgName, ast.Ty("Outcome3"))
	if typeExpressionToString(normalizedChoice) == "Choice3" {
		t.Fatalf("expected local Choice3 alias to normalize in package %q", pkgName)
	}
	if typeExpressionToString(normalizedOutcome) == "Outcome3" {
		t.Fatalf("expected local Outcome3 alias to normalize in package %q", pkgName)
	}

	choiceInfo, ok := gen.ensureNativeInterfaceInfo(pkgName, ast.Gen(ast.Ty("Keeper"), ast.Ty("Choice3")))
	if !ok || choiceInfo == nil {
		t.Fatalf("expected Keeper<Choice3> interface info to compile natively")
	}
	outcomeInfo, ok := gen.ensureNativeInterfaceInfo(pkgName, ast.Gen(ast.Ty("Keeper"), ast.Ty("Outcome3")))
	if !ok || outcomeInfo == nil {
		t.Fatalf("expected Keeper<Outcome3> interface info to compile natively")
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
