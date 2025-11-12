package runtime

import (
	"testing"
)

func TestProcHandleResolve(t *testing.T) {
	h := NewProcHandle()

	if status := h.Status(); status != ProcPending {
		t.Fatalf("expected pending status, got %v", status)
	}

	result := StringValue{Val: "done"}
	h.Resolve(result)

	val, err, status := h.Await()
	if status != ProcResolved {
		t.Fatalf("expected resolved status, got %v", status)
	}
	if err != nil {
		t.Fatalf("expected nil error, got %#v", err)
	}
	if got, ok := val.(StringValue); !ok || got.Val != "done" {
		t.Fatalf("unexpected result %#v", val)
	}
}

func TestProcHandleCancel(t *testing.T) {
	h := NewProcHandle()
	errVal := ErrorValue{Message: "cancelled"}
	h.Cancel(errVal)
	_, err, status := h.Await()
	if status != ProcCancelled {
		t.Fatalf("expected cancelled status, got %v", status)
	}
	if err == nil {
		t.Fatalf("expected error value")
	}
}

func TestFutureValueAwaitResolved(t *testing.T) {
	handle := NewProcHandle()
	future := NewFutureFromHandle(handle)

	expected := IntegerValue{Val: bigInt(5), TypeSuffix: IntegerI32}
	handle.Resolve(expected)

	val, err, status := future.Await()
	if status != ProcResolved {
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
