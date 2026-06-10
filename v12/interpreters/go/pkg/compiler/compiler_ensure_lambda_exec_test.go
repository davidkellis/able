package compiler

import "testing"

func TestCompilerEnsureInsideNativeLambdaBuildsAndRuns(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiled ensure/lambda execution test in short mode")
	}

	source := `package demo

fn main() -> void {
  action := fn() -> i64 {
    do {
      7_i64 + 5_i64
    } ensure {
      marker := 1_i64
    }
  }
  print(action())
}
`

	if got := compileAndRunExecSourceWithOptions(t, "ablec-ensure-lambda", source, Options{PackageName: "main", EmitMain: true}); got != "12\n" {
		t.Fatalf("compiled ensure/lambda output = %q, want %q", got, "12\\n")
	}
}
