//go:build !able_bytecode_box_reuse

package interpreter

import "able/interpreter-go/pkg/runtime"

// The false compile-time constant erases every diagnostic guard at normal VM,
// CLI, and compiler build time. The opt-in implementation is compiled only by
// profiling with -tags able_bytecode_box_reuse.
const bytecodeDynamicIntBoxReuseEnabled = false

func bytecodeRecordDynamicIntBoxCacheEvent(runtime.IntegerType, bytecodeDynamicIntBoxCacheEvent) {}

func bytecodeResetDynamicIntBoxCacheReuseForTest() {}

func bytecodeDynamicIntBoxCacheReuseForTest() map[string]bytecodeDynamicIntBoxCacheReuse {
	return nil
}
