package compiler

import (
	"strings"
	"testing"
)

func shadowedImportedGenericStructInterfaceCallablePackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, Thing::RemoteThing, Box, Outcome, Choice, First}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct Thing { local: i32 }",
			"",
			"fn read_result(value: Outcome (RemoteReader i32)) -> i32 {",
			"  value match {",
			"    case box: Box (RemoteReader i32) => box.value.read(),",
			"    case _: Error => 0",
			"  }",
			"}",
			"",
			"fn read_union(value: Choice (() -> RemoteThing)) -> i32 {",
			"  value match {",
			"    case box: Box (() -> RemoteThing) => box.value().remote,",
			"    case _: String => -1",
			"  }",
			"}",
			"",
			"fn main() -> i32 {",
			"  first_part := read_result(Box { value: First {} })",
			"  second_part := read_union(Box { value: fn() -> RemoteThing { RemoteThing { remote: 9 } } })",
			"  first_part + second_part",
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
			"struct Box T { value: T }",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
			"type Outcome T = Result (Box T)",
			"type Choice T = (Box T) | String",
			"",
		}, "\n"),
	}
}

func assertShadowedImportedGenericStructInterfaceCallableFunctionStaysNative(t *testing.T, result *Result, fnName string, directFragment string) {
	t.Helper()
	body, ok := findCompiledFunction(result, fnName)
	if !ok {
		t.Fatalf("could not find compiled %s function", fnName)
	}
	for _, fragment := range []string{
		"runtime.Value",
		" any",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_call_value(",
		"__able_member_get(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected %s to avoid %q:\n%s", fnName, fragment, body)
		}
	}
	if !strings.Contains(body, directFragment) {
		t.Fatalf("expected %s to keep direct native access via %q:\n%s", fnName, directFragment, body)
	}
}

func TestCompilerImportedGenericStructResultAliasWithShadowedInterfaceArgumentStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedGenericStructInterfaceCallablePackageFiles())
	assertShadowedImportedGenericStructInterfaceCallableFunctionStaysNative(t, result, "__able_compiled_fn_read_result", ".Value.read()")
}

func TestCompilerImportedGenericStructUnionAliasWithShadowedCallableArgumentStaysNative(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", shadowedImportedGenericStructInterfaceCallablePackageFiles())
	assertShadowedImportedGenericStructInterfaceCallableFunctionStaysNative(t, result, "__able_compiled_fn_read_union", ".Remote")
}
