package runtime

import "testing"

func TestArrayStoreMonoPrimitiveReadInfoIfAvailableChar(t *testing.T) {
	handle := ArrayStoreMonoNewWithCapacityChar(2)
	if err := ArrayStoreMonoWriteChar(handle, 0, 'a'); err != nil {
		t.Fatalf("ArrayStoreMonoWriteChar(0): %v", err)
	}
	if err := ArrayStoreMonoWriteChar(handle, 1, 'z'); err != nil {
		t.Fatalf("ArrayStoreMonoWriteChar(1): %v", err)
	}
	info, ok, err := ArrayStoreMonoPrimitiveReadInfoIfAvailable(handle, 1)
	if err != nil || !ok {
		t.Fatalf("ArrayStoreMonoPrimitiveReadInfoIfAvailable char = (%#v, %v, %v), want info/true/nil", info, ok, err)
	}
	if info.Kind != ArrayStoreMonoPrimitiveReadChar || !info.InBounds || info.Size != 2 || rune(info.Int64) != 'z' {
		t.Fatalf("char info = %#v, want kind=char inBounds=true size=2 value='z'", info)
	}
}

func TestArrayStoreMonoPrimitiveReadInfoIfAvailableU32OutOfBounds(t *testing.T) {
	handle := ArrayStoreMonoNewWithCapacityU32(1)
	if err := ArrayStoreMonoWriteU32(handle, 0, 7); err != nil {
		t.Fatalf("ArrayStoreMonoWriteU32: %v", err)
	}
	info, ok, err := ArrayStoreMonoPrimitiveReadInfoIfAvailable(handle, 3)
	if err != nil || !ok {
		t.Fatalf("ArrayStoreMonoPrimitiveReadInfoIfAvailable u32 = (%#v, %v, %v), want info/true/nil", info, ok, err)
	}
	if info.Kind != ArrayStoreMonoPrimitiveReadU32 || info.InBounds || info.Size != 1 {
		t.Fatalf("u32 oob info = %#v, want kind=u32 inBounds=false size=1", info)
	}
}

func TestArrayStoreMonoPrimitiveReadInfoIntoClearsDestination(t *testing.T) {
	handle := ArrayStoreMonoNewWithCapacityChar(1)
	if err := ArrayStoreMonoWriteChar(handle, 0, 'k'); err != nil {
		t.Fatalf("ArrayStoreMonoWriteChar: %v", err)
	}
	info := ArrayStoreMonoPrimitiveReadInfo{
		Kind:     ArrayStoreMonoPrimitiveReadU64,
		Size:     99,
		InBounds: true,
		Uint64:   123,
	}
	ok, err := ArrayStoreMonoPrimitiveReadInfoInto(handle, 0, &info)
	if err != nil || !ok {
		t.Fatalf("ArrayStoreMonoPrimitiveReadInfoInto char = (%#v, %v, %v), want info/true/nil", info, ok, err)
	}
	if info.Kind != ArrayStoreMonoPrimitiveReadChar || !info.InBounds || info.Size != 1 || rune(info.Int64) != 'k' || info.Uint64 != 0 {
		t.Fatalf("char info after Into = %#v, want cleared char info for 'k'", info)
	}
	ok, err = ArrayStoreMonoPrimitiveReadInfoInto(0, 0, &info)
	if err != nil || ok {
		t.Fatalf("ArrayStoreMonoPrimitiveReadInfoInto zero handle = (%#v, %v, %v), want unavailable/nil", info, ok, err)
	}
	if info != (ArrayStoreMonoPrimitiveReadInfo{}) {
		t.Fatalf("zero-handle Into left destination = %#v, want cleared", info)
	}
}

func TestArrayStoreMonoPrimitiveReadInfoIntoFreshReadsZeroedDestination(t *testing.T) {
	handle := ArrayStoreMonoNewWithCapacityChar(1)
	if err := ArrayStoreMonoWriteChar(handle, 0, 'm'); err != nil {
		t.Fatalf("ArrayStoreMonoWriteChar: %v", err)
	}
	var info ArrayStoreMonoPrimitiveReadInfo
	ok, err := ArrayStoreMonoPrimitiveReadInfoIntoFresh(handle, 0, &info)
	if err != nil || !ok {
		t.Fatalf("ArrayStoreMonoPrimitiveReadInfoIntoFresh = (%#v, %v, %v), want info/true/nil", info, ok, err)
	}
	if info.Kind != ArrayStoreMonoPrimitiveReadChar || !info.InBounds || info.Size != 1 || rune(info.Int64) != 'm' || info.Uint64 != 0 {
		t.Fatalf("fresh char info = %#v, want char info for 'm'", info)
	}
}

func TestArrayStoreMonoPrimitiveReadInfoIfAvailableDynamicReturnsUnavailable(t *testing.T) {
	handle := ArrayStoreNew()
	if err := ArrayStoreWrite(handle, 0, CharValue{Val: 'q'}); err != nil {
		t.Fatalf("ArrayStoreWrite: %v", err)
	}
	info, ok, err := ArrayStoreMonoPrimitiveReadInfoIfAvailable(handle, 0)
	if err != nil {
		t.Fatalf("ArrayStoreMonoPrimitiveReadInfoIfAvailable dynamic err: %v", err)
	}
	if ok {
		t.Fatalf("dynamic info = %#v, want unavailable", info)
	}
}

func TestArrayStoreMonoPrimitiveReadInfoIfAvailableTracksDynamicToMonoPromotion(t *testing.T) {
	handle := ArrayStoreNew()
	if err := ArrayStoreWrite(handle, 0, CharValue{Val: 'a'}); err != nil {
		t.Fatalf("ArrayStoreWrite: %v", err)
	}

	promoted, err := ArrayStoreAppendCharPromote(handle, 'z')
	if err != nil {
		t.Fatalf("ArrayStoreAppendCharPromote: %v", err)
	}
	if !promoted {
		t.Fatalf("ArrayStoreAppendCharPromote returned promoted=false, want true")
	}

	info, ok, err := ArrayStoreMonoPrimitiveReadInfoIfAvailable(handle, 1)
	if err != nil || !ok {
		t.Fatalf("ArrayStoreMonoPrimitiveReadInfoIfAvailable promoted = (%#v, %v, %v), want info/true/nil", info, ok, err)
	}
	if info.Kind != ArrayStoreMonoPrimitiveReadChar || !info.InBounds || info.Size != 2 || rune(info.Int64) != 'z' {
		t.Fatalf("promoted char info = %#v, want kind=char inBounds=true size=2 value='z'", info)
	}
}

func TestArrayStoreMonoPrimitiveReadInfoIfAvailableClearsAfterMonoToDynamicDeopt(t *testing.T) {
	handle := ArrayStoreMonoNewWithCapacityChar(1)
	if err := ArrayStoreMonoWriteChar(handle, 0, 'x'); err != nil {
		t.Fatalf("ArrayStoreMonoWriteChar: %v", err)
	}
	info, ok, err := ArrayStoreMonoPrimitiveReadInfoIfAvailable(handle, 0)
	if err != nil || !ok {
		t.Fatalf("ArrayStoreMonoPrimitiveReadInfoIfAvailable mono = (%#v, %v, %v), want info/true/nil", info, ok, err)
	}
	if _, err := ArrayStoreState(handle); err != nil {
		t.Fatalf("ArrayStoreState: %v", err)
	}
	info, ok, err = ArrayStoreMonoPrimitiveReadInfoIfAvailable(handle, 0)
	if err != nil {
		t.Fatalf("ArrayStoreMonoPrimitiveReadInfoIfAvailable dynamic err: %v", err)
	}
	if ok {
		t.Fatalf("dynamic info after deopt = %#v, want unavailable", info)
	}
}

func TestArrayStoreMonoPrimitiveReadInfoHotStateCacheTracksDeoptAndPromotion(t *testing.T) {
	handle := ArrayStoreMonoNewWithCapacityChar(1)
	if err := ArrayStoreMonoWriteChar(handle, 0, 'x'); err != nil {
		t.Fatalf("ArrayStoreMonoWriteChar: %v", err)
	}
	info, ok, err := ArrayStoreMonoPrimitiveReadInfoIfAvailable(handle, 0)
	if err != nil || !ok || !info.InBounds || rune(info.Int64) != 'x' {
		t.Fatalf("first primitive read = (%#v, %v, %v), want char x", info, ok, err)
	}
	firstHot := monoArrayCharReadHot
	if monoArrayCharReadHotHandle != handle || firstHot == nil {
		t.Fatalf("char read hot cache = handle %d state %p, want handle %d non-nil", monoArrayCharReadHotHandle, firstHot, handle)
	}
	if monoArrayPrimitiveReadInfoHotChar == nil || monoArrayPrimitiveReadInfoHotChar != firstHot {
		t.Fatalf("primitive read info hot char state = %p, want %p", monoArrayPrimitiveReadInfoHotChar, firstHot)
	}
	if !monoArrayPrimitiveReadInfoHotOK ||
		monoArrayPrimitiveReadInfoHotHandle != handle ||
		monoArrayPrimitiveReadInfoHotKind != monoArrayKindChar {
		t.Fatalf("primitive read info hot kind = (ok %v handle %d kind %v), want char for handle %d",
			monoArrayPrimitiveReadInfoHotOK, monoArrayPrimitiveReadInfoHotHandle, monoArrayPrimitiveReadInfoHotKind, handle)
	}

	if _, err := ArrayStoreState(handle); err != nil {
		t.Fatalf("ArrayStoreState deopt: %v", err)
	}
	if monoArrayCharReadHotHandle == handle || monoArrayCharReadHot != nil {
		t.Fatalf("char read hot cache after deopt = handle %d state %p, want cleared", monoArrayCharReadHotHandle, monoArrayCharReadHot)
	}
	if monoArrayPrimitiveReadInfoHotChar != nil {
		t.Fatalf("primitive read info hot char state after deopt = %p, want cleared", monoArrayPrimitiveReadInfoHotChar)
	}
	if monoArrayPrimitiveReadInfoHotOK && monoArrayPrimitiveReadInfoHotHandle == handle {
		t.Fatalf("primitive read info hot kind survived deopt for handle %d", handle)
	}
	info, ok, err = ArrayStoreMonoPrimitiveReadInfoIfAvailable(handle, 0)
	if err != nil {
		t.Fatalf("dynamic primitive read err: %v", err)
	}
	if ok {
		t.Fatalf("dynamic primitive read = %#v, want unavailable", info)
	}
	if !monoArrayPrimitiveReadInfoHotOK ||
		monoArrayPrimitiveReadInfoHotHandle != handle ||
		monoArrayPrimitiveReadInfoHotKind != monoArrayKindDynamic {
		t.Fatalf("primitive read info hot kind after dynamic read = (ok %v handle %d kind %v), want dynamic for handle %d",
			monoArrayPrimitiveReadInfoHotOK, monoArrayPrimitiveReadInfoHotHandle, monoArrayPrimitiveReadInfoHotKind, handle)
	}

	if err := ArrayStoreWrite(handle, 0, CharValue{Val: 'y'}); err != nil {
		t.Fatalf("ArrayStoreWrite dynamic char: %v", err)
	}
	promoted, err := ArrayStoreAppendCharPromote(handle, 'z')
	if err != nil {
		t.Fatalf("ArrayStoreAppendCharPromote: %v", err)
	}
	if !promoted {
		t.Fatalf("ArrayStoreAppendCharPromote returned false, want true")
	}
	info, ok, err = ArrayStoreMonoPrimitiveReadInfoIfAvailable(handle, 1)
	if err != nil || !ok || !info.InBounds || rune(info.Int64) != 'z' {
		t.Fatalf("promoted primitive read = (%#v, %v, %v), want char z", info, ok, err)
	}
	if monoArrayCharReadHotHandle != handle || monoArrayCharReadHot == nil || monoArrayCharReadHot == firstHot {
		t.Fatalf("char read hot cache after promotion = handle %d state %p, want handle %d new non-nil state",
			monoArrayCharReadHotHandle, monoArrayCharReadHot, handle)
	}
	if monoArrayPrimitiveReadInfoHotChar == nil || monoArrayPrimitiveReadInfoHotChar != monoArrayCharReadHot {
		t.Fatalf("primitive read info hot char state after promotion = %p, want %p",
			monoArrayPrimitiveReadInfoHotChar, monoArrayCharReadHot)
	}
	if !monoArrayPrimitiveReadInfoHotOK ||
		monoArrayPrimitiveReadInfoHotHandle != handle ||
		monoArrayPrimitiveReadInfoHotKind != monoArrayKindChar {
		t.Fatalf("primitive read info hot kind after promotion = (ok %v handle %d kind %v), want char for handle %d",
			monoArrayPrimitiveReadInfoHotOK, monoArrayPrimitiveReadInfoHotHandle, monoArrayPrimitiveReadInfoHotKind, handle)
	}
}

func TestArrayStoreMonoTypeInfoTracksKindCacheAcrossDeoptAndPromotion(t *testing.T) {
	handle := ArrayStoreMonoNewWithCapacityChar(1)
	if err := ArrayStoreMonoWriteChar(handle, 0, 'x'); err != nil {
		t.Fatalf("ArrayStoreMonoWriteChar: %v", err)
	}

	typeName, ok, err := ArrayStoreMonoElementTypeNameIfKnown(handle)
	if err != nil || !ok || typeName != "char" {
		t.Fatalf("mono type info = (%q, %v, %v), want char/true/nil", typeName, ok, err)
	}
	if monoArrayElementTypeNameHotHandle != handle || monoArrayElementTypeNameHot != "char" || !monoArrayElementTypeNameHotOK {
		t.Fatalf("mono type info hot cache = (handle %d, %q, %v), want handle %d char/true",
			monoArrayElementTypeNameHotHandle, monoArrayElementTypeNameHot, monoArrayElementTypeNameHotOK, handle)
	}

	if _, err := ArrayStoreState(handle); err != nil {
		t.Fatalf("ArrayStoreState: %v", err)
	}
	if monoArrayElementTypeNameHotHandle == handle {
		t.Fatalf("mono type info hot cache survived deopt for handle %d", handle)
	}
	typeName, ok, err = ArrayStoreMonoElementTypeNameIfKnown(handle)
	if err != nil {
		t.Fatalf("dynamic type info err: %v", err)
	}
	if ok {
		t.Fatalf("dynamic type info after deopt = (%q, %v), want unavailable", typeName, ok)
	}

	if err := ArrayStoreWrite(handle, 0, CharValue{Val: 'y'}); err != nil {
		t.Fatalf("ArrayStoreWrite dynamic char: %v", err)
	}
	promoted, err := ArrayStoreAppendCharPromote(handle, 'z')
	if err != nil {
		t.Fatalf("ArrayStoreAppendCharPromote: %v", err)
	}
	if !promoted {
		t.Fatalf("ArrayStoreAppendCharPromote returned false, want true")
	}
	typeName, ok, err = ArrayStoreMonoElementTypeNameIfKnown(handle)
	if err != nil || !ok || typeName != "char" {
		t.Fatalf("promoted type info = (%q, %v, %v), want char/true/nil", typeName, ok, err)
	}
	if monoArrayElementTypeNameHotHandle != handle || monoArrayElementTypeNameHot != "char" || !monoArrayElementTypeNameHotOK {
		t.Fatalf("promoted type info hot cache = (handle %d, %q, %v), want handle %d char/true",
			monoArrayElementTypeNameHotHandle, monoArrayElementTypeNameHot, monoArrayElementTypeNameHotOK, handle)
	}
}
