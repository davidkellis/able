package compiler

import "testing"

func TestCompilerLogicalOperatorsShortCircuitStaticBoolRHS(t *testing.T) {
	compileAndRunSource(t, "ablec-logical-short-circuit-static-", `
extern go fn __able_os_exit(code: i32) -> void {}

fn fail() -> bool {
  __able_os_exit(1)
  true
}

fn main() -> void {
  if false && fail() { __able_os_exit(2) }
  if true || fail() { __able_os_exit(0) }
  __able_os_exit(3)
}
`)
}

func TestCompilerLogicalOperatorsShortCircuitRuntimeValueRHS(t *testing.T) {
	compileAndRunSource(t, "ablec-logical-short-circuit-runtime-", `
extern go fn __able_os_exit(code: i32) -> void {}

fn fail() -> i32 {
  __able_os_exit(1)
  1
}

fn main() -> void {
  left: any = false
  if left && fail() { __able_os_exit(2) }
  left = true
  if left || fail() { __able_os_exit(0) }
  __able_os_exit(3)
}
`)
}
