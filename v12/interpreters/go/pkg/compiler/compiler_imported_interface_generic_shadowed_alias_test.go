package compiler

import (
	"strings"
	"testing"
)

func importedInterfaceGenericShadowedAliasPackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Echo, Tagger, Box, Labeler, Reader::RemoteReader, Thing::RemoteThing, Choice, Outcome, First}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct Thing { local: i32 }",
			"",
			"fn main() -> i32 {",
			"  echo: Echo = Box {}",
			"  tagged: Tagger = Labeler { label: \"L\" }",
			"  first := echo.pass<Choice (RemoteReader i32)>(First {})",
			"  second := tagged.tagged<Outcome (() -> RemoteThing)>(fn() -> RemoteThing {",
			"    RemoteThing { remote: 7 }",
			"  })",
			"  read_part := first match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: String => 0",
			"  }",
			"  build_part := second.value match {",
			"    case build: (() -> RemoteThing) => build().remote,",
			"    case _: Error => 0",
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
			"type Choice T = T | String",
			"type Outcome T = Error | T",
			"",
		}, "\n"),
	}
}

func TestCompilerImportedInterfaceGenericShadowedAliasCallsStayNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", importedInterfaceGenericShadowedAliasPackageFiles())

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	for _, fragment := range []string{
		"var echo runtime.Value",
		"var echo any",
		"var tagged runtime.Value",
		"var tagged any",
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
			t.Fatalf("expected imported generic interface shadowed-alias calls to avoid %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"var echo __able_iface_Echo = __able_iface_Echo_wrap_ptr_Box(",
		"var tagged __able_iface_Tagger = __able_iface_Tagger_wrap_ptr_Labeler(",
		"__able_compiled_iface_Echo_pass_dispatch(",
		"__able_compiled_iface_Tagger_tagged_default(",
		"var first __able_union_",
		"var second *Tagged_",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected imported generic interface shadowed-alias calls to include %q:\n%s", fragment, mainBody)
		}
	}

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"func __able_compiled_iface_Echo_pass_dispatch(",
		"func __able_compiled_iface_Tagger_tagged_default(",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected imported generic interface shadowed-alias helpers to include %q:\n%s", fragment, compiledSrc)
		}
	}
	for _, fragment := range []string{
		"func __able_compiled_iface_Echo_pass_dispatch(self __able_iface_Echo, value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_iface_Echo_pass_dispatch(self __able_iface_Echo, value any) (any, *__ableControl)",
		"func __able_compiled_iface_Tagger_tagged_default(self __able_iface_Tagger, value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_iface_Tagger_tagged_default(self __able_iface_Tagger, value any) (any, *__ableControl)",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected imported generic interface shadowed-alias helpers to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}
