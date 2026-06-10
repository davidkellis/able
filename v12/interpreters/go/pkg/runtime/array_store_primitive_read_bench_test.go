package runtime

import "testing"

var benchmarkArrayStorePrimitiveReadInfoSink ArrayStoreMonoPrimitiveReadInfo
var benchmarkArrayStorePrimitiveRuneSink rune
var benchmarkArrayStorePrimitiveBoolSink bool

func benchmarkArrayStoreMonoCharHandle(b *testing.B) int64 {
	b.Helper()
	handle := ArrayStoreMonoNewWithCapacityChar(64)
	for i := 0; i < 64; i++ {
		if err := ArrayStoreMonoWriteChar(handle, i, rune('a'+i%26)); err != nil {
			b.Fatalf("ArrayStoreMonoWriteChar(%d): %v", i, err)
		}
	}
	return handle
}

func BenchmarkArrayStoreMonoPrimitiveReadInfoChar(b *testing.B) {
	handle := benchmarkArrayStoreMonoCharHandle(b)
	var warm ArrayStoreMonoPrimitiveReadInfo
	if ok, err := ArrayStoreMonoPrimitiveReadInfoIntoFresh(handle, 0, &warm); err != nil || !ok {
		b.Fatalf("warm primitive read = (%#v, %v, %v), want ok/nil", warm, ok, err)
	}
	state, ok := cachedMonoArrayCharReadState(handle)
	if !ok {
		b.Fatalf("cachedMonoArrayCharReadState(%d) = false", handle)
	}

	b.Run("fresh_hot", func(b *testing.B) {
		b.ReportAllocs()
		var info ArrayStoreMonoPrimitiveReadInfo
		for i := 0; i < b.N; i++ {
			ok, err := ArrayStoreMonoPrimitiveReadInfoIntoFresh(handle, i&63, &info)
			if err != nil || !ok {
				b.Fatalf("ArrayStoreMonoPrimitiveReadInfoIntoFresh = (%v, %v), want ok/nil", ok, err)
			}
			benchmarkArrayStorePrimitiveReadInfoSink = info
		}
	})

	b.Run("clearing", func(b *testing.B) {
		b.ReportAllocs()
		var info ArrayStoreMonoPrimitiveReadInfo
		for i := 0; i < b.N; i++ {
			ok, err := ArrayStoreMonoPrimitiveReadInfoInto(handle, i&63, &info)
			if err != nil || !ok {
				b.Fatalf("ArrayStoreMonoPrimitiveReadInfoInto = (%v, %v), want ok/nil", ok, err)
			}
			benchmarkArrayStorePrimitiveReadInfoSink = info
		}
	})

	b.Run("value_return", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			info, ok, err := ArrayStoreMonoPrimitiveReadInfoIfAvailable(handle, i&63)
			if err != nil || !ok {
				b.Fatalf("ArrayStoreMonoPrimitiveReadInfoIfAvailable = (%v, %v), want ok/nil", ok, err)
			}
			benchmarkArrayStorePrimitiveReadInfoSink = info
		}
	})

	b.Run("char_helper", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			value, ok, err := ArrayStoreMonoReadCharIfAvailable(handle, i&63)
			if err != nil || !ok {
				b.Fatalf("ArrayStoreMonoReadCharIfAvailable = (%v, %v), want ok/nil", ok, err)
			}
			benchmarkArrayStorePrimitiveRuneSink = value
			benchmarkArrayStorePrimitiveBoolSink = ok
		}
	})

	b.Run("direct_state_fill", func(b *testing.B) {
		b.ReportAllocs()
		var info ArrayStoreMonoPrimitiveReadInfo
		for i := 0; i < b.N; i++ {
			fillMonoArrayCharPrimitiveReadInfo(state, i&63, &info)
			benchmarkArrayStorePrimitiveReadInfoSink = info
		}
	})
}
