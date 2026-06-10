package compiler

import (
	"strings"
	"testing"
)

func singletonErrorRescueSource() string {
	return strings.Join([]string{
		"package demo",
		"",
		"union Choice T = nil | T",
		"",
		"struct MarkerError {}",
		"",
		"impl Error for MarkerError {",
		"  fn message(self: Self) -> String { \"marker\" }",
		"  fn cause(self: Self) -> ?Error { nil }",
		"}",
		"",
		"methods Choice T {",
		"  fn fail(self: Self) -> void {",
		"    raise(MarkerError {})",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  present: Choice i32 = 1_i32",
		"  do { present.fail() } rescue {",
		"    case _: MarkerError => print(\"caught\")",
		"  }",
		"}",
		"",
	}, "\n")
}

func TestCompilerSingletonErrorRescueRecoversNominalPayload(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-singleton-error-rescue-", singletonErrorRescueSource(), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if strings.TrimSpace(stdout) != "caught" {
		t.Fatalf("expected singleton error rescue output caught, got %q", stdout)
	}
}
