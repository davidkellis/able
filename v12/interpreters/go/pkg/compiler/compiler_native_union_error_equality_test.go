package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNativeResultErrorEqualityMatchesReference(t *testing.T) {
	compileAndRunSource(t, "ablec-native-result-error-equality-", strings.Join([]string{
		"package demo",
		"",
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"struct MyError { message: String }",
		"",
		"impl Error for MyError {",
		"  fn message(self: Self) -> String { self.message }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"fn failure() -> !u32 { MyError { message: \"no\" } }",
		"",
		"fn main() {",
		"  if failure() == failure() { __able_os_exit(1) }",
		"  if failure() != failure() { __able_os_exit(0) }",
		"  __able_os_exit(2)",
		"}",
		"",
	}, "\n"))
}
