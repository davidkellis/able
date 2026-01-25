package runtime

import (
	"testing"
)

func TestFutureResolve(t *testing.T) {
	h := NewFuture()

	if status := h.Status(); status != FuturePending {
		t.Fatalf("expected pending status, got %v", status)
	}

	result := StringValue{Val: "done"}
	h.Resolve(result)

	val, err, status := h.Await()
	if status != FutureResolved {
		t.Fatalf("expected resolved status, got %v", status)
	}
	if err != nil {
		t.Fatalf("expected nil error, got %#v", err)
	}
	if got, ok := val.(StringValue); !ok || got.Val != "done" {
		t.Fatalf("unexpected result %#v", val)
	}
}

func TestFutureCancel(t *testing.T) {
	h := NewFuture()
	errVal := ErrorValue{Message: "cancelled"}
	h.Cancel(errVal)
	_, err, status := h.Await()
	if status != FutureCancelled {
		t.Fatalf("expected cancelled status, got %v", status)
	}
	if err == nil {
		t.Fatalf("expected error value")
	}
}

func TestFutureAwaitResolved(t *testing.T) {
	future := NewFuture()

	expected := IntegerValue{Val: bigInt(5), TypeSuffix: IntegerI32}
	future.Resolve(expected)

	val, err, status := future.Await()
	if status != FutureResolved {
		t.Fatalf("expected resolved status, got %v", status)
	}
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}
	iv, ok := val.(IntegerValue)
	if !ok || iv.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("unexpected future value: %#v", val)
	}
}
