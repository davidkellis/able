package compiler

import (
	"strings"
	"testing"
)

func TestCompilerSpecNullableExpectationBuilds(t *testing.T) {
	_ = compileAndBuildStdlibSource(t, "ablec-spec-nullable-matcher-", strings.Join([]string{
		"package demo",
		"",
		"import able.spec.*",
		"",
		"fn fetch(flag: bool) -> ?i32 {",
		"  if flag {",
		"    1",
		"  } else {",
		"    nil",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  expect(fetch(true)).to(eq(1))",
		"  expect(fetch(false)).to(be_nil())",
		"}",
		"",
	}, "\n"))
}
