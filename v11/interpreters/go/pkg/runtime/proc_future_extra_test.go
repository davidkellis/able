package runtime

import "testing"

func TestFutureCancelRequestedFlag(t *testing.T) {
	h := NewFuture()
	if h.CancelRequested() {
		t.Fatalf("expected cancelRequested to be false initially")
	}

	h.RequestCancel()
	if !h.CancelRequested() {
		t.Fatalf("expected cancelRequested to be true after request")
	}
}

func TestFutureAwaitFailure(t *testing.T) {
	future := NewFuture()

	errVal := ErrorValue{Message: "boom"}
	future.Fail(errVal)

	val, err, status := future.Await()
	if status != FutureFailed {
		t.Fatalf("expected failed status, got %v", status)
	}
	if val != nil {
		t.Fatalf("expected nil value, got %#v", val)
	}
	if err == nil {
		t.Fatalf("expected error value")
	}
}
