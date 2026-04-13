package compiler

import (
	"strings"
	"testing"
)

func assertNestedUnionLiteralMainStaysNative(t *testing.T, result *Result) string {
	t.Helper()
	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var value runtime.Value",
		"var value any",
		"__able_try_cast(",
		"bridge.MatchType(",
		"__able_any_to_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected nested union-member literal path to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "var value __able_union_") {
		t.Fatalf("expected nested union-member literal path to keep a native outer union carrier:\n%s", body)
	}
	return body
}

func TestCompilerNestedResultUnionStringLiteralStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"type Outer = !String | bool",
		"",
		"fn main() -> String {",
		"  value: Outer = \"ok\"",
		"  value match {",
		"    case text: String => text,",
		"    case _: Error => \"err\",",
		"    case flag: bool => if flag { \"true\" } else { \"false\" }",
		"  }",
		"}",
		"",
	}, "\n"))

	body := assertNestedUnionLiteralMainStaysNative(t, result)
	if !strings.Contains(body, "\"ok\"") {
		t.Fatalf("expected nested result-member string literal to stay on the compiled path:\n%s", body)
	}
}

func TestCompilerNestedNullableUnionStringLiteralStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"type Outer = ?String | bool",
		"",
		"fn main() -> String {",
		"  value: Outer = \"ok\"",
		"  value match {",
		"    case text: String => text,",
		"    case nil => \"nil\",",
		"    case flag: bool => if flag { \"true\" } else { \"false\" }",
		"  }",
		"}",
		"",
	}, "\n"))

	body := assertNestedUnionLiteralMainStaysNative(t, result)
	if !strings.Contains(body, "\"ok\"") {
		t.Fatalf("expected nested nullable-member string literal to stay on the compiled path:\n%s", body)
	}
}

func TestCompilerNestedUnionIntegerLiteralRetargetsUniqueInnerCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"type Inner = String | i64",
		"type Outer = Inner | bool",
		"",
		"fn main() -> i64 {",
		"  value: Outer = 7",
		"  value match {",
		"    case n: i64 => n,",
		"    case _: String => 0,",
		"    case _: bool => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	body := assertNestedUnionLiteralMainStaysNative(t, result)
	if !strings.Contains(body, "int64(7)") {
		t.Fatalf("expected nested union integer literal to retarget onto the unique inner i64 carrier:\n%s", body)
	}
}

func TestCompilerNestedResultUnionStructLiteralStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Box { value: i32 }",
		"",
		"type Outer = !Box | bool",
		"",
		"fn main() -> i32 {",
		"  value: Outer = Box { value: 7 }",
		"  value match {",
		"    case box: Box => box.value,",
		"    case _: Error => 0,",
		"    case _: bool => 1",
		"  }",
		"}",
		"",
	}, "\n"))

	body := assertNestedUnionLiteralMainStaysNative(t, result)
	if !strings.Contains(body, "&Box{Value: int32(7)}") {
		t.Fatalf("expected nested result-member struct literal to stay on the native Box carrier:\n%s", body)
	}
}

func TestCompilerNestedResultUnionCarrierFlattensRepresentableMembers(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"type Outer = !String | bool",
		"",
		"fn main(value: Outer) -> String {",
		"  value match {",
		"    case text: String => text,",
		"    case _: Error => \"err\",",
		"    case flag: bool => if flag { \"true\" } else { \"false\" }",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"value __able_union___able_union_",
		"_as___able_union_",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected representable nested result members to flatten into one native union and avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "value __able_union_") {
		t.Fatalf("expected flattened nested result members to keep a native union carrier:\n%s", body)
	}
}

func TestCompilerImportedNestedResultInterfaceUnionFlattensRepresentableMembers(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader, Outer}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn main(value: Outer (RemoteReader i32)) -> i32 {",
			"  value match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: Error => 0,",
			"    case _: bool => 1",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct First {}",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
			"type Outcome T = Error | T",
			"type Outer T = Outcome T | bool",
			"",
		}, "\n"),
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"value __able_union___able_union_",
		"_as___able_union_",
		"__able_try_cast(",
		"bridge.MatchType(",
		"runtime.Value",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected imported nested result/interface carrier to flatten and avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "reader.read()") {
		t.Fatalf("expected imported nested result/interface carrier to keep direct native interface dispatch:\n%s", body)
	}
}

func TestCompilerNestedResultCarrierFlattensRepresentableMembers(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main(value: !!String) -> String {",
		"  value match {",
		"    case text: String => text,",
		"    case _: Error => \"err\"",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"value __able_union___able_union_",
		"_as___able_union_",
		"runtime.Value",
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected nested result carrier to flatten into one native union and avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "value __able_union_") {
		t.Fatalf("expected nested result carrier to keep a flattened native union parameter:\n%s", body)
	}
}

func TestCompilerNestedResultImportedInterfaceCarrierFlattensRepresentableMembers(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Reader::RemoteReader}",
			"",
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"fn main(value: !!(RemoteReader i32)) -> i32 {",
			"  value match {",
			"    case reader: RemoteReader i32 => reader.read(),",
			"    case _: Error => 0",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"interface Reader <T> for Self {",
			"  fn read(self: Self) -> T",
			"}",
			"",
			"struct First {}",
			"",
			"impl Reader i32 for First {",
			"  fn read(self: Self) -> i32 { 7 }",
			"}",
			"",
		}, "\n"),
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"value __able_union___able_union_",
		"_as___able_union_",
		"runtime.Value",
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected nested imported result/interface carrier to flatten and avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "reader.read()") {
		t.Fatalf("expected nested imported result/interface carrier to keep direct native interface dispatch:\n%s", body)
	}
}
