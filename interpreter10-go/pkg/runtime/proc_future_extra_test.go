package runtime

import "testing"

func TestProcHandleCancelRequestedFlag(t *testing.T) {
	h := NewProcHandle()
	if h.CancelRequested() {
		t.Fatalf("expected cancelRequested to be false initially")
	}

	h.RequestCancel()
	if !h.CancelRequested() {
		t.Fatalf("expected cancelRequested to be true after request")
	}
}

func TestFutureValueAwaitFailure(t *testing.T) {
	handle := NewProcHandle()
	future := NewFutureFromHandle(handle)

	errVal := ErrorValue{Message: "boom"}
	handle.Fail(errVal)

	val, err, status := future.Await()
	if status != ProcFailed {
		t.Fatalf("expected failed status, got %v", status)
	}
	if val != nil {
		t.Fatalf("expected nil value, got %#v", val)
	}
	if err == nil {
		t.Fatalf("expected error value")
	}
}
