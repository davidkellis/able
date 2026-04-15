package compiler

import (
	"strings"
	"testing"
)

func nativeInterfaceGenericShadowedAliasPackageFiles() map[string]string {
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
			"struct Box {}",
			"",
			"interface Echo for Self {",
			"  fn pass<T>(self: Self, value: T) -> T",
			"}",
			"",
			"impl Echo for Box {",
			"  fn pass<T>(self: Self, value: T) -> T { value }",
			"}",
			"",
			"fn main() -> i32 {",
			"  echo: Echo = Box {}",
			"  first := echo.pass<Choice (RemoteReader i32)>(First {})",
			"  second := echo.pass<Outcome (() -> RemoteThing)>(fn() -> RemoteThing {",
			"    RemoteThing { remote: 7 }",
			"  })",
			"  read_part := first match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: String => 0",
			"  }",
			"  build_part := second match {",
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

func TestCompilerGenericInterfaceMethodImportedShadowedAliasActualsStayNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", nativeInterfaceGenericShadowedAliasPackageFiles())

	mainBody := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	for _, fragment := range []string{
		"var echo runtime.Value",
		"var echo any",
		"var first runtime.Value",
		"var first any",
		"var second runtime.Value",
		"var second any",
		"__able_method_call_node(",
		"bridge.MatchType(",
		"__able_try_cast(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected generic interface method imported shadowed alias path to avoid %q:\n%s", fragment, mainBody)
		}
	}
	if !strings.Contains(mainBody, "var echo __able_iface_Echo = __able_iface_Echo_wrap_ptr_Box(") {
		t.Fatalf("expected generic interface existential to stay native:\n%s", mainBody)
	}
	for _, fragment := range []string{
		"var first __able_union_",
		"var second __able_union_",
		"reader.read()",
		".Remote",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected generic interface method imported shadowed alias path to include %q:\n%s", fragment, mainBody)
		}
	}

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"func __able_compiled_iface_Echo_pass_dispatch(",
		"func __able_compiled_impl_Echo_pass_",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected generic interface method imported shadowed alias helpers to include %q:\n%s", fragment, compiledSrc)
		}
	}
	for _, fragment := range []string{
		"func __able_compiled_iface_Echo_pass_dispatch(self __able_iface_Echo, value runtime.Value) (runtime.Value, *__ableControl)",
		"func __able_compiled_iface_Echo_pass_dispatch(self __able_iface_Echo, value any) (any, *__ableControl)",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected generic interface method imported shadowed alias helper to avoid %q:\n%s", fragment, compiledSrc)
		}
	}
}
