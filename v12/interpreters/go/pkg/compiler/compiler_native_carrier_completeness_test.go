package compiler

import (
	"strings"
	"testing"
)

func TestCompilerDivModConcreteCarrierStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-divmod-native-carrier", strings.Join([]string{
		"package demo",
		"",
		"fn parts() -> DivMod i32 {",
		"  7 /% 3",
		"}",
		"",
		"fn main() -> i32 {",
		"  divmod := parts()",
		"  divmod.quotient + divmod.remainder",
		"}",
		"",
	}, "\n"))

	source := compiledSourceText(t, result)
	for _, fragment := range []string{
		"type DivMod_i32 struct",
		"func __able_compiled_fn_parts() (*DivMod_i32, *__ableControl)",
	} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("expected native DivMod carrier fragment %q:\n%s", fragment, source)
		}
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_parts")
	if !ok {
		t.Fatalf("could not find compiled parts function")
	}
	for _, fragment := range []string{
		"__able_divmod_result(",
		"runtime.Value",
		"__able_any_to_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete DivMod carrier path to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "&DivMod_i32{Quotient:") {
		t.Fatalf("expected concrete DivMod carrier construction in compiled body:\n%s", body)
	}
}

func TestCompilerParameterizedUnionAliasLocalStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-parameterized-union-alias-carrier", strings.Join([]string{
		"package demo",
		"",
		"type PairOrText T = DivMod T | String",
		"",
		"fn main() -> i32 {",
		"  local: PairOrText i32 = if true { 7 /% 3 } else { \"ok\" }",
		"  local match {",
		"    case pair: DivMod i32 => pair.remainder,",
		"    case _ => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var local __able_union_") {
		t.Fatalf("expected parameterized union alias local to use a native union carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var local runtime.Value",
		"var local any",
		"__able_divmod_result(",
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected parameterized union alias local to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerParameterizedResultAliasLocalStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-parameterized-result-alias-carrier", strings.Join([]string{
		"package demo",
		"",
		"struct MyError { message: String }",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"type CalcResult T = Error | T",
		"",
		"fn value(ok: bool) -> CalcResult (DivMod i32) {",
		"  if ok { 7 /% 3 } else { MyError { message: \"bad\" } }",
		"}",
		"",
		"fn main() -> i32 {",
		"  local: CalcResult (DivMod i32) = value(true)",
		"  local match {",
		"    case pair: DivMod i32 => pair.quotient,",
		"    case _: Error => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	source := compiledSourceText(t, result)
	if !strings.Contains(source, "func __able_compiled_fn_value(ok bool) (__able_union_") {
		t.Fatalf("expected parameterized result alias return to use a native union carrier:\n%s", source)
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var local __able_union_") {
		t.Fatalf("expected parameterized result alias local to use a native union carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var local runtime.Value",
		"var local any",
		"__able_divmod_result(",
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected parameterized result alias local to avoid %q:\n%s", fragment, body)
		}
	}
}
