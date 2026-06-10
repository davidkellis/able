package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

var (
	benchmarkIntegerInfoSink   integerInfo
	benchmarkIntegerInfoOKSink bool
)

func BenchmarkLookupIntegerInfo(b *testing.B) {
	for _, tc := range []struct {
		name string
		kind runtime.IntegerType
	}{
		{name: "i32", kind: runtime.IntegerI32},
		{name: "i128", kind: runtime.IntegerI128},
		{name: "isize", kind: runtime.IntegerIsize},
		{name: "miss", kind: runtime.IntegerType("not-integer")},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for idx := 0; idx < b.N; idx++ {
				benchmarkIntegerInfoSink, benchmarkIntegerInfoOKSink = lookupIntegerInfo(tc.kind)
			}
		})
	}
}
