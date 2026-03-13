package compiler

import (
	"strings"
	"testing"
)

func TestCompilerMultiMemberUnionMatchStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Red {}",
		"struct Green {}",
		"struct Blue {}",
		"",
		"union Color = Red | Green | Blue",
		"",
		"fn color_name(color: Color) -> String {",
		"  color match {",
		"    case Red => \"red\",",
		"    case Green => \"green\",",
		"    case Blue => \"blue\"",
		"  }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_color_name(color __able_union_") {
		t.Fatalf("expected multi-member union param to use a native union carrier:\n%s", compiledSrc)
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_color_name")
	if !ok {
		t.Fatalf("could not find compiled color_name function")
	}
	if !strings.Contains(body, "_as_") {
		t.Fatalf("expected multi-member union match lowering to use native branch extractors:\n%s", body)
	}
	for _, fragment := range []string{"__able_try_cast(", "bridge.MatchType("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected multi-member union match lowering to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerGenericUnionAliasesStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"union Option T = nil | T",
		"union Result T = Error | T",
		"",
		"struct MyError { message: String }",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn describe_option(value: Option i32) -> String {",
		"  value match {",
		"    case nil => \"none\",",
		"    case n: i32 => `some ${n}`",
		"  }",
		"}",
		"",
		"fn describe_result(value: Result i32) -> String {",
		"  value match {",
		"    case err: Error => err.message(),",
		"    case n: i32 => `ok ${n}`",
		"  }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_describe_option(value *int32) (string, *__ableControl)") {
		t.Fatalf("expected Option i32 alias to normalize to native nullable pointer carrier:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_describe_result(value __able_union_") {
		t.Fatalf("expected Result i32 alias to normalize to native union carrier:\n%s", compiledSrc)
	}

	optionBody, ok := findCompiledFunction(result, "__able_compiled_fn_describe_option")
	if !ok {
		t.Fatalf("could not find compiled describe_option function")
	}
	for _, fragment := range []string{"__able_try_cast(", "bridge.MatchType("} {
		if strings.Contains(optionBody, fragment) {
			t.Fatalf("expected Option alias nullable lowering to avoid %q:\n%s", fragment, optionBody)
		}
	}

	resultBody, ok := findCompiledFunction(result, "__able_compiled_fn_describe_result")
	if !ok {
		t.Fatalf("could not find compiled describe_result function")
	}
	if !strings.Contains(resultBody, "_as_") {
		t.Fatalf("expected Result alias match lowering to use native union extractors:\n%s", resultBody)
	}
	for _, fragment := range []string{"__able_try_cast(", "bridge.MatchType("} {
		if strings.Contains(resultBody, fragment) {
			t.Fatalf("expected Result alias union lowering to avoid %q:\n%s", fragment, resultBody)
		}
	}
}

func TestCompilerInterfaceBranchUnionStaysOnNativeCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Tag for Self {",
		"  fn tag(self: Self) -> String",
		"}",
		"",
		"struct Alpha {}",
		"",
		"impl Tag for Alpha {",
		"  fn tag(self: Self) -> String { \"alpha\" }",
		"}",
		"",
		"fn describe(value: String | Tag) -> String {",
		"  value match {",
		"    case tag: Tag => tag.tag(),",
		"    case text: String => text",
		"  }",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_describe(value __able_union_") {
		t.Fatalf("expected interface-branch union param to use native union carrier:\n%s", compiledSrc)
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_describe")
	if !ok {
		t.Fatalf("could not find compiled describe function")
	}
	if !strings.Contains(body, "_as___able_iface_Tag(") {
		t.Fatalf("expected interface union branch to unwrap through native interface carrier helper:\n%s", body)
	}
	if !strings.Contains(body, "tag.tag()") {
		t.Fatalf("expected interface union branch payload to dispatch through the native interface carrier:\n%s", body)
	}
	for _, fragment := range []string{"__able_try_cast(", "bridge.MatchType(", "__able_method_call_node("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected interface union discrimination to avoid %q in compiled body:\n%s", fragment, body)
		}
	}
}

func TestCompilerSingletonStructBoundaryAcceptsRuntimeDefinition(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Blue {}",
		"",
		"fn describe(value: Blue) -> String {",
		"  \"blue\"",
		"}",
		"",
	}, "\n"))

	fromBody, ok := findCompiledFunction(result, "__able_struct_Blue_from")
	if !ok {
		t.Fatalf("could not find compiled Blue from converter")
	}
	if !strings.Contains(fromBody, "__able_runtime_struct_definition_value(current)") {
		t.Fatalf("expected singleton struct converter to accept runtime struct definitions:\n%s", fromBody)
	}

	applyBody, ok := findCompiledFunction(result, "__able_struct_Blue_apply")
	if !ok {
		t.Fatalf("could not find compiled Blue apply converter")
	}
	if !strings.Contains(applyBody, "__able_runtime_struct_definition_value(targetCurrent)") {
		t.Fatalf("expected singleton struct apply converter to accept runtime struct definitions:\n%s", applyBody)
	}
}

func TestCompilerBroadNativeUnionExecutes(t *testing.T) {
	source := strings.Join([]string{
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"interface Tag for Self {",
		"  fn tag(self: Self) -> String",
		"}",
		"",
		"struct Red {}",
		"struct Green {}",
		"struct Blue {}",
		"struct Alpha {}",
		"struct MyError { message: String }",
		"",
		"union Color = Red | Green | Blue",
		"union Option T = nil | T",
		"union Result T = Error | T",
		"",
		"impl Tag for Alpha {",
		"  fn tag(self: Self) -> String { \"alpha\" }",
		"}",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn color_name(color: Color) -> String {",
		"  color match {",
		"    case Red => \"red\",",
		"    case Green => \"green\",",
		"    case Blue => \"blue\"",
		"  }",
		"}",
		"",
		"fn describe_option(value: Option i32) -> String {",
		"  value match {",
		"    case nil => \"none\",",
		"    case n: i32 => `some ${n}`",
		"  }",
		"}",
		"",
		"fn describe_result(value: Result i32) -> String {",
		"  value match {",
		"    case err: Error => err.message(),",
		"    case n: i32 => `ok ${n}`",
		"  }",
		"}",
		"",
		"fn describe_union(value: String | Tag) -> String {",
		"  value match {",
		"    case tag: Tag => tag.tag(),",
		"    case text: String => text",
		"  }",
		"}",
		"",
		"fn main() {",
		"  if color_name(Blue) == \"blue\" &&",
		"     describe_option(4) == \"some 4\" &&",
		"     describe_option(nil) == \"none\" &&",
		"     describe_result(9) == \"ok 9\" &&",
		"     describe_result(MyError { message: \"bad\" }) == \"bad\" &&",
		"     describe_union(Alpha {}) == \"alpha\" &&",
		"     describe_union(\"text\") == \"text\" {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n")
	compileAndRunSource(t, "ablec-broad-native-union-", source)
}
