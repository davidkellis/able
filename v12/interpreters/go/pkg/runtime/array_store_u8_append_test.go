package runtime

import "testing"

func TestArrayStoreAppendU8StringPromoteAppendsToMonoU8(t *testing.T) {
	arr := ArrayStoreMonoValueFromU8String("ab")
	if arr == nil || arr.Handle == 0 {
		t.Fatalf("mono u8 array missing handle: %#v", arr)
	}
	ok, err := ArrayStoreAppendU8StringPromote(arr.Handle, "cd")
	if err != nil {
		t.Fatalf("ArrayStoreAppendU8StringPromote: %v", err)
	}
	if !ok {
		t.Fatalf("expected mono u8 string append to succeed")
	}
	got, ok, err := ArrayStoreMonoBorrowedU8BytesIfAvailable(arr.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoBorrowedU8BytesIfAvailable: %v", err)
	}
	if !ok || string(got) != "abcd" {
		t.Fatalf("mono u8 bytes = %q/%v, want %q/true", string(got), ok, "abcd")
	}
}

func TestArrayStoreAppendU8BytesPromoteConvertsDynamicArray(t *testing.T) {
	handle := ArrayStoreNew()
	if err := ArrayStoreWrite(handle, 0, NewSmallInt(65, IntegerU8)); err != nil {
		t.Fatalf("ArrayStoreWrite(0): %v", err)
	}
	if err := ArrayStoreWrite(handle, 1, NewSmallInt(66, IntegerU8)); err != nil {
		t.Fatalf("ArrayStoreWrite(1): %v", err)
	}
	ok, err := ArrayStoreAppendU8BytesPromote(handle, []byte{67, 68})
	if err != nil {
		t.Fatalf("ArrayStoreAppendU8BytesPromote: %v", err)
	}
	if !ok {
		t.Fatalf("expected dynamic-to-mono u8 append to succeed")
	}
	got, ok, err := ArrayStoreMonoBorrowedU8BytesIfAvailable(handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoBorrowedU8BytesIfAvailable: %v", err)
	}
	if !ok || string(got) != "ABCD" {
		t.Fatalf("mono u8 bytes = %q/%v, want %q/true", string(got), ok, "ABCD")
	}
}
