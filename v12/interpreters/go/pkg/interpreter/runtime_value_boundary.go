package interpreter

import "able/interpreter-go/pkg/runtime"

// MaterializeRuntimeValue returns a stable runtime value for use outside the
// bytecode VM. VM-only raw scalar carriers are boxed; already stable values
// are returned unchanged.
func MaterializeRuntimeValue(value runtime.Value) runtime.Value {
	return bytecodeMaterializeRawValue(value)
}

// MaterializeRuntimeValues materializes VM-only raw scalar carriers in a
// runtime-value slice. Stable slices are returned without allocation.
func MaterializeRuntimeValues(values []runtime.Value) []runtime.Value {
	if !bytecodeCallArgsNeedMaterialization(values) {
		return values
	}
	materialized := make([]runtime.Value, len(values))
	bytecodeCopyMaterializedCallArgs(materialized, values)
	return materialized
}
