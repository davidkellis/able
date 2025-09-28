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

func TestFutureValueRunnerError(t *testing.T) {
	errFuture := NewFutureValue(func() (Value, Value) {
		return nil, ErrorValue{Message: "boom"}
	})

	val, err := errFuture.force()
	if val != nil {
		t.Fatalf("expected nil value, got %#v", val)
	}
	if err == nil {
		t.Fatalf("expected error value")
	}
}
