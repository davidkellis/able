package runtime

import "testing"

func TestArrayStoreMonoValueFromU8BytesCopiesInput(t *testing.T) {
	source := []byte{1, 2, 3}
	arr := ArrayStoreMonoValueFromU8Bytes(source)
	source[1] = 9

	got, ok, err := ArrayStoreMonoReadU8IfAvailable(arr.Handle, 1)
	if err != nil {
		t.Fatalf("ArrayStoreMonoReadU8IfAvailable: %v", err)
	}
	if !ok {
		t.Fatalf("expected mono u8 handle")
	}
	if got != 2 {
		t.Fatalf("copied mono u8[1] = %d, want 2", got)
	}
}

func TestArrayStoreMonoValueFromOwnedU8BytesReusesInput(t *testing.T) {
	source := []byte{1, 2, 3}
	arr := ArrayStoreMonoValueFromOwnedU8Bytes(source)
	source[1] = 9

	got, ok, err := ArrayStoreMonoReadU8IfAvailable(arr.Handle, 1)
	if err != nil {
		t.Fatalf("ArrayStoreMonoReadU8IfAvailable: %v", err)
	}
	if !ok {
		t.Fatalf("expected mono u8 handle")
	}
	if got != 9 {
		t.Fatalf("owned mono u8[1] = %d, want 9", got)
	}
}

func TestArrayStoreMonoBorrowedU8BytesIfAvailableReusesBacking(t *testing.T) {
	source := []byte{1, 2, 3}
	arr := ArrayStoreMonoValueFromOwnedU8Bytes(source)

	borrowed, ok, err := ArrayStoreMonoBorrowedU8BytesIfAvailable(arr.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoBorrowedU8BytesIfAvailable: %v", err)
	}
	if !ok {
		t.Fatalf("expected mono u8 handle")
	}
	if len(borrowed) != len(source) || &borrowed[0] != &source[0] {
		t.Fatalf("expected borrowed mono bytes to reuse backing storage")
	}
}
