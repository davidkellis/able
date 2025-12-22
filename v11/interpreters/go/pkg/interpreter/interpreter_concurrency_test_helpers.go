package interpreter

import (
	"context"
	"testing"
	"time"

	"able/interpreter-go/pkg/runtime"
)

func waitForStatus(handle *runtime.ProcHandleValue, desired runtime.ProcStatus, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if handle.Status() == desired {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return handle.Status() == desired
}

func waitForEnvString(t *testing.T, env *runtime.Environment, name string, desired string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		current := mustGetString(t, env, name)
		if current == desired {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return mustGetString(t, env, name) == desired
}

func mustGetString(t *testing.T, env *runtime.Environment, name string) string {
	val, err := env.Get(name)
	if err != nil {
		t.Fatalf("failed to read %s: %v", name, err)
	}
	strVal, ok := val.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected %s to be string, got %#v", name, val)
	}
	return strVal.Val
}

func mustGetBool(t *testing.T, env *runtime.Environment, name string) bool {
	val, err := env.Get(name)
	if err != nil {
		t.Fatalf("failed to read %s: %v", name, err)
	}
	boolVal, ok := val.(runtime.BoolValue)
	if !ok {
		t.Fatalf("expected %s to be bool, got %#v", name, val)
	}
	return boolVal.Val
}

func intFromValue(t *testing.T, value runtime.Value) int {
	t.Helper()
	intVal, ok := value.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer value, got %#v", value)
	}
	return int(intVal.Val.Int64())
}

type stubExecutor struct {
	flushCalls int
}

func (s *stubExecutor) RunProc(task ProcTask) *runtime.ProcHandleValue {
	handle := runtime.NewProcHandle()
	go func() {
		if task != nil {
			if _, err := task(context.Background()); err != nil {
				handle.Fail(runtime.ErrorValue{Message: err.Error()})
				return
			}
		}
		handle.Resolve(runtime.NilValue{})
	}()
	return handle
}

func (s *stubExecutor) RunFuture(task ProcTask) *runtime.FutureValue {
	handle := s.RunProc(task)
	return runtime.NewFutureFromHandle(handle)
}

func (s *stubExecutor) Flush() {
	s.flushCalls++
}

func (s *stubExecutor) PendingTasks() int {
	return 0
}
