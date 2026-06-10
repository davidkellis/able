package compiler

import "testing"

func TestCompilerNullableNominalFieldCrossesRuntimeBoundary(t *testing.T) {
	compileAndRunSource(t, "ablec-nullable-nominal-boundary-", `
extern go fn __able_os_exit(code: i32) -> void {}

struct Child { value: i32 }
struct Parent { child: ?Child }

fn round_trip(value: any) -> any { value }

fn main() -> void {
  parent := Parent { child: nil }
  restored := round_trip(parent) as Parent
  restored.child match {
    case nil => __able_os_exit(0),
    case _: Child => __able_os_exit(1)
  }
}
`)
}
