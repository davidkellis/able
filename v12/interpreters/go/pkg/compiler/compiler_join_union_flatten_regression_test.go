package compiler

import (
	"strings"
	"testing"
)

func TestCompilerJoinFlattensExistingNativeUnionMembers(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn render(flag: bool) -> i32 | String {",
		"  inner := if flag { 1 } else { \"x\" }",
		"  if flag { inner } else { \"y\" }",
		"}",
		"",
		"fn main() -> void {",
		"  print(render(false))",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-join-union-flatten", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if strings.TrimSpace(stdout) != "y" {
		t.Fatalf("expected joined union regression to print y, got %q", stdout)
	}
}
