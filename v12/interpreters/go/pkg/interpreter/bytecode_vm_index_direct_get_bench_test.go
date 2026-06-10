package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

var benchmarkBytecodeResolveValueSink runtime.Value
var benchmarkBytecodeResolveTokenSink uint16
var benchmarkBytecodeResolveKnownSink bool

func benchmarkBytecodeMonoCharHandle(b *testing.B) int64 {
	b.Helper()
	handle := runtime.ArrayStoreMonoNewWithCapacityChar(64)
	for i := 0; i < 64; i++ {
		if err := runtime.ArrayStoreMonoWriteChar(handle, i, rune('a'+i%26)); err != nil {
			b.Fatalf("ArrayStoreMonoWriteChar(%d): %v", i, err)
		}
	}
	return handle
}

func BenchmarkBytecodeVMResolveMonoPrimitiveArrayGetAtHandle(b *testing.B) {
	handle := benchmarkBytecodeMonoCharHandle(b)
	vm := &bytecodeVM{interp: &Interpreter{}}
	if value, token, known, err := vm.resolveDirectArrayIndexGetAtHandleReadyWithToken(handle, 0); err != nil {
		b.Fatalf("warm resolve err: %v", err)
	} else if _, ok := value.(runtime.CharValue); !ok || token != bytecodeIndexTypeChar || !known {
		b.Fatalf("warm resolve = (%#v, %d, %v), want char token", value, token, known)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		value, token, known, err := vm.resolveDirectArrayIndexGetAtHandleReadyWithToken(handle, i&63)
		if err != nil {
			b.Fatalf("resolveDirectArrayIndexGetAtHandleReadyWithToken: %v", err)
		}
		benchmarkBytecodeResolveValueSink = value
		benchmarkBytecodeResolveTokenSink = token
		benchmarkBytecodeResolveKnownSink = known
	}
}
