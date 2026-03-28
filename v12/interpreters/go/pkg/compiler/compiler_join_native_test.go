package compiler

import (
	"strings"
	"testing"
)

func TestCompilerIfExpressionMixedBranchesInferNativeUnion(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  from_if := if true {",
		"    1",
		"  } else {",
		"    \"ok\"",
		"  }",
		"  from_if match {",
		"    case n: i32 => n",
		"    case _ => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var from_if __able_union_") {
		t.Fatalf("expected if-expression join to infer a native union carrier:\n%s", body)
	}
	if strings.Contains(body, "var from_if runtime.Value") {
		t.Fatalf("expected if-expression join to avoid runtime.Value local:\n%s", body)
	}
}

func TestCompilerMatchExpressionMixedClausesInferNativeUnion(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  from_match := true match {",
		"    case true => 1",
		"    case _ => \"ok\"",
		"  }",
		"  from_match match {",
		"    case n: i32 => n",
		"    case _ => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var from_match __able_union_") {
		t.Fatalf("expected match-expression join to infer a native union carrier:\n%s", body)
	}
	if strings.Contains(body, "var from_match runtime.Value") {
		t.Fatalf("expected match-expression join to avoid runtime.Value local:\n%s", body)
	}
}

func TestCompilerMatchExpressionInterfaceJoinCoercesNativeArrayBranches(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-match-join-iterable-array", strings.Join([]string{
		"package demo",
		"",
		"import able.kernel.{Array}",
		"import able.core.iteration.{Iterable}",
		"",
		"fn pick(value: ?(Array String)) -> (Iterable String) {",
		"  value match {",
		"    case nil => Array.new(),",
		"    case items: Array String => items",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_pick")
	if !ok {
		t.Fatalf("could not find compiled pick function")
	}
	if !strings.Contains(body, "var __able_tmp_") || !strings.Contains(body, "__able_iface_Iterable_String") {
		t.Fatalf("expected match join to use the native Iterable carrier:\n%s", body)
	}
	if !strings.Contains(body, "__able_iface_Iterable_String_from_value(__able_runtime,") {
		t.Fatalf("expected match join branches to coerce onto the native interface carrier:\n%s", body)
	}
}

func TestCompilerRescueExpressionMixedBranchesInferNativeUnion(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  from_rescue := do {",
		"    raise(\"boom\")",
		"    \"ok\"",
		"  } rescue {",
		"    case _: Error => 1",
		"  }",
		"  from_rescue match {",
		"    case n: i32 => n",
		"    case _ => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var from_rescue __able_union_") {
		t.Fatalf("expected rescue-expression join to infer a native union carrier:\n%s", body)
	}
	if strings.Contains(body, "var from_rescue runtime.Value") {
		t.Fatalf("expected rescue-expression join to avoid runtime.Value local:\n%s", body)
	}
}

func TestCompilerOrElseOnNullableMixedBranchesInferNativeUnion(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn maybe(flag: bool) -> ?i32 {",
		"  if flag { 1 } else { nil }",
		"}",
		"",
		"fn main() -> i32 {",
		"  from_or := maybe(true) or { \"missing\" }",
		"  from_or match {",
		"    case n: i32 => n",
		"    case _ => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var from_or __able_union_") {
		t.Fatalf("expected nullable or-else join to infer a native union carrier:\n%s", body)
	}
	if strings.Contains(body, "var from_or runtime.Value") {
		t.Fatalf("expected nullable or-else join to avoid runtime.Value local:\n%s", body)
	}
}

func TestCompilerOrElseOnErrorUnionMixedBranchesInferNativeUnion(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct MyError { message: String }",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn value(ok: bool) -> String | MyError {",
		"  if ok { \"ok\" } else { MyError { message: \"bad\" } }",
		"}",
		"",
		"fn main() -> i32 {",
		"  from_or := value(false) or { 7 }",
		"  from_or match {",
		"    case n: i32 => n",
		"    case _ => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var from_or __able_union_") {
		t.Fatalf("expected error-union or-else join to infer a native union carrier:\n%s", body)
	}
	if strings.Contains(body, "var from_or runtime.Value") {
		t.Fatalf("expected error-union or-else join to avoid runtime.Value local:\n%s", body)
	}
}

func TestCompilerLoopExpressionBreakValuesInferNativeUnion(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  from_loop := loop {",
		"    n := 1",
		"    if true { break n }",
		"    break \"ok\"",
		"  }",
		"  from_loop match {",
		"    case n: i32 => n",
		"    case _ => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var from_loop __able_union_") {
		t.Fatalf("expected loop-expression join to infer a native union carrier:\n%s", body)
	}
	if strings.Contains(body, "var from_loop runtime.Value") {
		t.Fatalf("expected loop-expression join to avoid runtime.Value local:\n%s", body)
	}
}

func TestCompilerBreakpointExpressionMixedExitsInferNativeUnion(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  from_breakpoint := breakpoint 'mix {",
		"    n := 1",
		"    if true { break 'mix n }",
		"    \"ok\"",
		"  }",
		"  from_breakpoint match {",
		"    case n: i32 => n",
		"    case _ => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var from_breakpoint __able_union_") {
		t.Fatalf("expected breakpoint-expression join to infer a native union carrier:\n%s", body)
	}
	if strings.Contains(body, "var from_breakpoint runtime.Value") {
		t.Fatalf("expected breakpoint-expression join to avoid runtime.Value local:\n%s", body)
	}
}

func TestCompilerJoinExpressionConcreteImplementersInferNativeInterface(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Speak for Self {",
		"  fn speak(self: Self) -> String",
		"}",
		"",
		"struct Cat {}",
		"struct Dog {}",
		"",
		"impl Speak for Cat { fn speak(self: Self) -> String { \"cat\" } }",
		"impl Speak for Dog { fn speak(self: Self) -> String { \"dog\" } }",
		"",
		"fn main() -> String {",
		"  speaker := if true { Cat {} } else { Dog {} }",
		"  speaker.speak()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var speaker __able_iface_Speak =") {
		t.Fatalf("expected concrete-implementer join to infer the shared native interface carrier:\n%s", body)
	}
	if strings.Contains(body, "var speaker __able_union_") {
		t.Fatalf("expected concrete-implementer join to avoid a native union local when a shared interface exists:\n%s", body)
	}
	if !strings.Contains(body, "speaker.speak()") {
		t.Fatalf("expected concrete-implementer join to dispatch through the native interface carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete-implementer join to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerJoinExpressionConcreteGenericImplementersInferNativeInterface(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Echo for Self {",
		"  fn pass<T>(self: Self, value: T) -> T",
		"}",
		"",
		"struct Box {}",
		"struct Mirror {}",
		"",
		"impl Echo for Box {",
		"  fn pass<T>(self: Self, value: T) -> T { value }",
		"}",
		"",
		"impl Echo for Mirror {",
		"  fn pass<T>(self: Self, value: T) -> T { value }",
		"}",
		"",
		"fn main() -> i32 {",
		"  value := if true { Box {} } else { Mirror {} }",
		"  value.pass<i32>(7)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var value __able_iface_Echo =") {
		t.Fatalf("expected concrete generic-implementer join to infer the shared native interface carrier:\n%s", body)
	}
	if strings.Contains(body, "var value __able_union_") {
		t.Fatalf("expected concrete generic-implementer join to avoid a native union local when a shared interface exists:\n%s", body)
	}
	if !strings.Contains(body, "__able_compiled_iface_Echo_pass_dispatch(") {
		t.Fatalf("expected concrete generic-implementer join to dispatch through the compiled generic interface helper:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete generic-implementer join to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerJoinExpressionConcreteErrorsInferNativeErrorCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct FirstError { message: String }",
		"struct SecondError { message: String }",
		"",
		"impl Error for FirstError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"impl Error for SecondError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn main() -> String {",
		"  err := if true { FirstError { message: \"first\" } } else { SecondError { message: \"second\" } }",
		"  err.message()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var err runtime.ErrorValue =") {
		t.Fatalf("expected concrete error join to infer the native Error carrier:\n%s", body)
	}
	if strings.Contains(body, "var err __able_union_") || strings.Contains(body, "var err runtime.Value") {
		t.Fatalf("expected concrete error join to avoid union/runtime.Value fallback:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete error join to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerIfExpressionInterfaceAndNilInferNativeCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Reader <T> for Self {",
		"  fn read(self: Self) -> T",
		"}",
		"",
		"struct Left {}",
		"struct Right {}",
		"",
		"impl Reader i32 for Left { fn read(self: Self) -> i32 { 1 } }",
		"impl Reader i32 for Right { fn read(self: Self) -> i32 { 2 } }",
		"",
		"fn main() -> i32 {",
		"  left: Reader i32 = Left {}",
		"  reader := if true { left } else { nil }",
		"  reader.read()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var reader __able_iface_Reader_i32 =") {
		t.Fatalf("expected interface-and-nil join to infer the native interface carrier:\n%s", body)
	}
	if strings.Contains(body, "var reader runtime.Value") {
		t.Fatalf("expected interface-and-nil join to avoid runtime.Value fallback:\n%s", body)
	}
}

func TestCompilerMatchExpressionCallableAndNilInferNativeCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  inc: (i32 -> i32) = fn(value: i32) -> i32 { value + 1 }",
		"  fnc := true match {",
		"    case true => inc",
		"    case _ => nil",
		"  }",
		"  fnc(41)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var fnc __able_fn_int32_to_int32 =") {
		t.Fatalf("expected callable-and-nil join to infer the native callable carrier:\n%s", body)
	}
	if strings.Contains(body, "var fnc runtime.Value") {
		t.Fatalf("expected callable-and-nil join to avoid runtime.Value fallback:\n%s", body)
	}
}

func TestCompilerRescueExpressionErrorAndNilInferNativeNullableError(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct FirstError { message: String }",
		"struct SecondError { message: String }",
		"",
		"impl Error for FirstError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"impl Error for SecondError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn main() -> i32 {",
		"  first: Error = FirstError { message: \"first\" }",
		"  second: Error = SecondError { message: \"second\" }",
		"  err := do {",
		"    raise(\"boom\")",
		"    if true { first } else { second }",
		"  } rescue {",
		"    case _: Error => nil",
		"  }",
		"  if err == nil { 1 } else { 0 }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var err *runtime.ErrorValue =") {
		t.Fatalf("expected error-and-nil rescue join to infer a native nullable error carrier:\n%s", body)
	}
	if strings.Contains(body, "var err runtime.Value") {
		t.Fatalf("expected error-and-nil rescue join to avoid runtime.Value fallback:\n%s", body)
	}
}

func TestCompilerJoinExpressionsExecuteWithoutRuntimeCarrierFallback(t *testing.T) {
	source := strings.Join([]string{
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"struct MyError { message: String }",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn maybe(flag: bool) -> ?i32 {",
		"  if flag { 1 } else { nil }",
		"}",
		"",
		"fn value(ok: bool) -> String | MyError {",
		"  if ok { \"ok\" } else { MyError { message: \"bad\" } }",
		"}",
		"",
		"fn main() {",
		"  from_if := if true {",
		"    1",
		"  } else {",
		"    \"ok\"",
		"  }",
		"  ok_if := from_if match {",
		"    case n: i32 => n == 1",
		"    case s: String => s == \"ok\"",
		"  }",
		"",
		"  from_match := true match {",
		"    case true => 1",
		"    case _ => \"ok\"",
		"  }",
		"  ok_match := from_match match {",
		"    case n: i32 => n == 1",
		"    case s: String => s == \"ok\"",
		"  }",
		"",
		"  from_rescue := do {",
		"    raise(\"boom\")",
		"    \"ok\"",
		"  } rescue {",
		"    case _: Error => 1",
		"  }",
		"  ok_rescue := from_rescue match {",
		"    case n: i32 => n == 1",
		"    case s: String => s == \"ok\"",
		"  }",
		"",
		"  from_or_nullable := maybe(false) or { \"missing\" }",
		"  ok_or_nullable := from_or_nullable match {",
		"    case n: i32 => n == 1",
		"    case s: String => s == \"missing\"",
		"  }",
		"",
		"  from_or_error := value(false) or { 7 }",
		"  ok_or_error := from_or_error match {",
		"    case n: i32 => n == 7",
		"    case s: String => s == \"ok\"",
		"  }",
		"",
		"  from_loop := loop {",
		"    n := 1",
		"    if true { break n }",
		"    break \"bad\"",
		"  }",
		"  ok_loop := from_loop match {",
		"    case n: i32 => n == 1",
		"    case s: String => s == \"bad\"",
		"  }",
		"",
		"  from_breakpoint := breakpoint 'mix {",
		"    n := 1",
		"    if false { break 'mix n }",
		"    \"ok\"",
		"  }",
		"  ok_breakpoint := from_breakpoint match {",
		"    case n: i32 => n == 1",
		"    case s: String => s == \"ok\"",
		"  }",
		"",
		"  if ok_if && ok_match && ok_rescue {",
		"    if ok_or_nullable && ok_or_error && ok_loop && ok_breakpoint {",
		"    __able_os_exit(0)",
		"    }",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n")

	compileAndRunSourceWithOptions(t, "ablec-native-join-", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
}

func TestCompilerCommonExistentialJoinsExecuteWithoutDynamicFallback(t *testing.T) {
	source := strings.Join([]string{
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"interface Speak for Self {",
		"  fn speak(self: Self) -> String",
		"}",
		"",
		"interface Echo for Self {",
		"  fn pass<T>(self: Self, value: T) -> T",
		"}",
		"",
		"struct Cat {}",
		"struct Dog {}",
		"struct Box {}",
		"struct Mirror {}",
		"struct FirstError { message: String }",
		"struct SecondError { message: String }",
		"",
		"impl Speak for Cat { fn speak(self: Self) -> String { \"cat\" } }",
		"impl Speak for Dog { fn speak(self: Self) -> String { \"dog\" } }",
		"impl Echo for Box { fn pass<T>(self: Self, value: T) -> T { value } }",
		"impl Echo for Mirror { fn pass<T>(self: Self, value: T) -> T { value } }",
		"",
		"impl Error for FirstError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"impl Error for SecondError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn main() {",
		"  speaker := if true { Cat {} } else { Dog {} }",
		"  echo := if true { Box {} } else { Mirror {} }",
		"  err := if true { FirstError { message: \"first\" } } else { SecondError { message: \"second\" } }",
		"  if speaker.speak() == \"cat\" && echo.pass<i32>(7) == 7 && err.message() == \"first\" {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n")

	compileAndRunSourceWithOptions(t, "ablec-native-join-existential-", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
}

func TestCompilerNilCapableJoinExpressionsExecuteWithoutRuntimeCarrierFallback(t *testing.T) {
	source := strings.Join([]string{
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"interface Reader <T> for Self {",
		"  fn read(self: Self) -> T",
		"}",
		"",
		"struct Left {}",
		"struct Right {}",
		"",
		"impl Reader i32 for Left { fn read(self: Self) -> i32 { 1 } }",
		"impl Reader i32 for Right { fn read(self: Self) -> i32 { 2 } }",
		"",
		"fn main() {",
		"  left: Reader i32 = Left {}",
		"  reader := if true { left } else { nil }",
		"  inc: (i32 -> i32) = fn(value: i32) -> i32 { value + 1 }",
		"  fnc := true match {",
		"    case true => inc",
		"    case _ => nil",
		"  }",
		"  if reader.read() == 1 && fnc(41) == 42 {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n")

	compileAndRunSourceWithOptions(t, "ablec-native-join-nil-", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
}

func TestCompilerJoinExpressionConcreteParameterizedImplementersInferBoundNativeInterface(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Reader <T> for Self {",
		"  fn read(self: Self) -> T",
		"}",
		"",
		"struct Left {}",
		"struct Right {}",
		"",
		"impl Reader i32 for Left { fn read(self: Self) -> i32 { 1 } }",
		"impl Reader i32 for Right { fn read(self: Self) -> i32 { 2 } }",
		"",
		"fn main() -> i32 {",
		"  reader := if true { Left {} } else { Right {} }",
		"  reader.read()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var reader __able_iface_Reader_i32 =") {
		t.Fatalf("expected concrete parameterized-implementer join to infer the shared bound native interface carrier:\n%s", body)
	}
	if strings.Contains(body, "var reader __able_union_") || strings.Contains(body, "var reader runtime.Value") {
		t.Fatalf("expected concrete parameterized-implementer join to avoid union/runtime.Value fallback:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete parameterized-implementer join to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerJoinExpressionParameterizedInheritedImplementersInferSharedParentInterface(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Reader <T> for Self {",
		"  fn read(self: Self) -> T",
		"}",
		"",
		"interface LeftReader <T> for Self: Reader T {",
		"  fn left(self: Self) -> T",
		"}",
		"",
		"interface RightReader <T> for Self: Reader T {",
		"  fn right(self: Self) -> T",
		"}",
		"",
		"struct Left {}",
		"struct Right {}",
		"",
		"impl LeftReader i32 for Left {",
		"  fn read(self: Self) -> i32 { 1 }",
		"  fn left(self: Self) -> i32 { 1 }",
		"}",
		"",
		"impl RightReader i32 for Right {",
		"  fn read(self: Self) -> i32 { 2 }",
		"  fn right(self: Self) -> i32 { 2 }",
		"}",
		"",
		"fn main() -> i32 {",
		"  reader := if true { Left {} } else { Right {} }",
		"  reader.read()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var reader __able_iface_Reader_i32 =") {
		t.Fatalf("expected inherited parameterized-implementer join to infer the shared bound parent interface carrier:\n%s", body)
	}
	if strings.Contains(body, "var reader __able_union_") || strings.Contains(body, "var reader runtime.Value") {
		t.Fatalf("expected inherited parameterized-implementer join to avoid union/runtime.Value fallback:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected inherited parameterized-implementer join to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerParameterizedInterfaceJoinsExecuteWithoutDynamicFallback(t *testing.T) {
	source := strings.Join([]string{
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"interface Reader <T> for Self {",
		"  fn read(self: Self) -> T",
		"}",
		"",
		"interface LeftReader <T> for Self: Reader T {",
		"  fn left(self: Self) -> T",
		"}",
		"",
		"interface RightReader <T> for Self: Reader T {",
		"  fn right(self: Self) -> T",
		"}",
		"",
		"struct Left {}",
		"struct Right {}",
		"",
		"impl LeftReader i32 for Left {",
		"  fn read(self: Self) -> i32 { 1 }",
		"  fn left(self: Self) -> i32 { 1 }",
		"}",
		"",
		"impl RightReader i32 for Right {",
		"  fn read(self: Self) -> i32 { 2 }",
		"  fn right(self: Self) -> i32 { 2 }",
		"}",
		"",
		"fn main() {",
		"  direct := if true { Left {} } else { Right {} }",
		"  inherited := if false { Left {} } else { Right {} }",
		"  if direct.read() == 1 && inherited.read() == 2 {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n")

	compileAndRunSourceWithOptions(t, "ablec-native-join-parameterized-", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
}

func TestCompilerIfJoinRecoversTypeExprBackedInterfaceCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Reader <T> for Self {",
		"  fn read(self: Self) -> T",
		"}",
		"",
		"struct First {}",
		"struct Second {}",
		"",
		"impl Reader i32 for First { fn read(self: Self) -> i32 { 1 } }",
		"impl Reader i32 for Second { fn read(self: Self) -> i32 { 2 } }",
		"",
		"fn fail() -> First {",
		"  raise(First {})",
		"  First {}",
		"}",
		"",
		"fn main() -> i32 {",
		"  recovered := fail() rescue {",
		"    case reader: Reader i32 =>",
		"      if true { reader } else { Second {} }",
		"  }",
		"  recovered.read()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var recovered __able_iface_Reader_i32") {
		t.Fatalf("expected typed-pattern if-join to recover the native interface carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var recovered runtime.Value",
		"var recovered __able_union_",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected typed-pattern if-join to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerIfJoinRecoversTypeExprBackedErrorCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct MyError { message: String }",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn fail() -> MyError {",
		"  raise(MyError { message: \"boom\" })",
		"  MyError { message: \"ok\" }",
		"}",
		"",
		"fn main() -> String {",
		"  recovered := fail() rescue {",
		"    case err: Error =>",
		"      if true { err } else { MyError { message: \"fallback\" } }",
		"  }",
		"  recovered.message()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var recovered runtime.ErrorValue") {
		t.Fatalf("expected typed-pattern if-join to recover the native error carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var recovered runtime.Value",
		"var recovered __able_union_",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected typed-pattern error join to avoid %q:\n%s", fragment, body)
		}
	}
}
