package compiler

import (
	"strings"
	"testing"
)

func importedInterfaceGenericShadowedMultiMemberAliasPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Echo, Tagger, Box, Labeler, Reader::RemoteReader, Thing::RemoteThing, Choice3, Outcome3, First}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct Thing { local: i32 }",
			"",
			"struct MyError { message: String }",
			"",
			"impl Error for MyError {",
			"  fn message(self: Self) -> String { self.message }",
			"  fn cause(self: Self) -> ?Error { nil }",
			"}",
			"",
			"fn main() -> i32 {",
			"  echo: Echo = Box {}",
			"  tagged: Tagger = Labeler { label: \"L\" }",
			"  first_input: Choice3 (RemoteReader i32) = if true {",
			"    First {}",
			"  } else {",
			"    if false { \"ok\" } else { 9 }",
			"  }",
			"  second_input: Outcome3 (() -> RemoteThing) = if true {",
			"    fn() -> RemoteThing { RemoteThing { remote: 7 } }",
			"  } else {",
			"    if false { \"bad\" } else { MyError { message: \"oops\" } }",
			"  }",
			"  first := echo.pass<Choice3 (RemoteReader i32)>(first_input)",
			"  second := tagged.tagged<Outcome3 (() -> RemoteThing)>(second_input)",
			"  read_part := first match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: String => 0,",
			"    case n: i32 => n",
			"  }",
			"  build_part := second.value match {",
			"    case build: (() -> RemoteThing) => build().remote,",
			"    case _: String => 0,",
			"    case _: Error => -1",
			"  }",
			"  if second.tag == \"L\" { read_part + build_part } else { 0 }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
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
			"struct Labeler { label: String }",
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
			"type Choice3 T = T | String | i32",
			"type Outcome3 T = Error | T | String",
			"",
		}, "\n"),
	}
}

func TestCompilerImportedInterfaceGenericShadowedMultiMemberAliasCallsStayNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", importedInterfaceGenericShadowedMultiMemberAliasPackageFiles())

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	for _, fragment := range []string{
		"var echo runtime.Value",
		"var echo any",
		"var tagged runtime.Value",
		"var tagged any",
		"var first_input runtime.Value",
		"var first_input any",
		"var second_input runtime.Value",
		"var second_input any",
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
			t.Fatalf("expected imported generic-interface multi-member shadowed-alias calls to avoid %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"var echo __able_iface_Echo = __able_iface_Echo_wrap_ptr_Box(",
		"var tagged __able_iface_Tagger = __able_iface_Tagger_wrap_ptr_Labeler(",
		"var first_input __able_union_",
		"var second_input __able_union_",
		"__able_compiled_iface_Echo_pass_dispatch(",
		"__able_compiled_iface_Tagger_tagged_default(",
		"var first __able_union_",
		"var second *Tagged_",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected imported generic-interface multi-member shadowed-alias calls to include %q:\n%s", fragment, mainBody)
		}
	}

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"func __able_compiled_iface_Echo_pass_dispatch(",
		"func __able_compiled_iface_Tagger_tagged_default(",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected imported generic-interface multi-member shadowed-alias helpers to include %q:\n%s", fragment, compiledSrc)
		}
	}
	for _, fragment := range []string{
		"func __able_compiled_iface_Echo_pass_dispatch(self __able_iface_Echo, value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_iface_Echo_pass_dispatch(self __able_iface_Echo, value any) (any, *__ableControl)",
		"func __able_compiled_iface_Tagger_tagged_default(self __able_iface_Tagger, value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_iface_Tagger_tagged_default(self __able_iface_Tagger, value any) (any, *__ableControl)",
		"_variant_runtime_Value",
		"_wrap_runtime_Value",
		"_as_runtime_Value",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected imported generic-interface multi-member shadowed-alias helpers to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}
