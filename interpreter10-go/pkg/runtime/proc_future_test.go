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

func TestFutureValueRunner(t *testing.T) {
	calls := 0
	future := NewFutureValue(func() (Value, Value) {
		calls++
		return IntegerValue{Val: bigInt(5), TypeSuffix: IntegerI32}, nil
	})

	val, err := future.force()
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}
	iv, ok := val.(IntegerValue)
	if !ok || iv.Val.Cmp(bigInt(5)) != 0 {
		t.Fatalf("unexpected future value: %#v", val)
	}

	// Second force should reuse cached result.
	val2, err2 := future.force()
	if err2 != nil {
		t.Fatalf("unexpected error on second force: %#v", err2)
	}
	if val2.Kind() != KindInteger {
		t.Fatalf("unexpected kind on second force: %v", val2.Kind())
	}
	if calls != 1 {
		t.Fatalf("expected runner called once, got %d", calls)
	}
}
