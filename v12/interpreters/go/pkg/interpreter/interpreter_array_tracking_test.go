package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestInterpreterTrackArrayValueUsesSingleFastPath(t *testing.T) {
	interp := New()
	arr := &runtime.ArrayValue{}
	state, handle, err := runtime.ArrayStoreEnsure(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}

	interp.trackArrayValue(handle, arr)

	tracking, ok := interp.arraysByHandle[handle]
	if !ok {
		t.Fatalf("expected array tracking entry for handle %d", handle)
	}
	if tracking.single != arr {
		t.Fatalf("expected single tracked array fast path, got %#v", tracking)
	}
	if tracking.many != nil {
		t.Fatalf("expected no promoted tracking set for single array")
	}
	if arr.TrackedAliases {
		t.Fatalf("expected single tracked array to remain exclusive")
	}

	interp.syncArrayValues(handle, state)
	if arr.Handle != handle {
		t.Fatalf("expected tracked handle %d, got %d", handle, arr.Handle)
	}
	if arr.Elements == nil || len(arr.Elements) != len(state.Values) {
		t.Fatalf("expected synced array elements")
	}
}

func TestInterpreterTrackArrayValuePromotesAndDemotesAliases(t *testing.T) {
	interp := New()
	first := &runtime.ArrayValue{}
	state, handle, err := runtime.ArrayStoreEnsure(first, 0)
	if err != nil {
		t.Fatalf("ensure first array state: %v", err)
	}
	interp.trackArrayValue(handle, first)
	second, err := interp.arrayValueFromHandle(handle, 0, 0)
	if err != nil {
		t.Fatalf("arrayValueFromHandle: %v", err)
	}

	tracking := interp.arraysByHandle[handle]
	if tracking.single != nil {
		t.Fatalf("expected alias promotion to tracking set, got single=%#v", tracking.single)
	}
	if len(tracking.many) != 2 {
		t.Fatalf("expected 2 tracked aliases after promotion, got %d", len(tracking.many))
	}
	if !first.TrackedAliases || !second.TrackedAliases {
		t.Fatalf("expected both aliases to be marked shared after promotion")
	}

	state.Values = append(state.Values, runtime.NewSmallInt(7, runtime.IntegerI32))
	interp.syncArrayValues(handle, state)
	if len(first.Elements) != 1 || len(second.Elements) != 1 {
		t.Fatalf("expected synced values for both aliases, got first=%d second=%d", len(first.Elements), len(second.Elements))
	}

	interp.untrackArrayValue(handle, second)
	tracking = interp.arraysByHandle[handle]
	if tracking.single != first {
		t.Fatalf("expected demotion back to single tracked alias, got %#v", tracking)
	}
	if tracking.many != nil {
		t.Fatalf("expected promoted tracking set to collapse after untracking second alias")
	}
	if first.TrackedAliases {
		t.Fatalf("expected remaining alias to return to exclusive tracking")
	}
}

func TestInterpreterEnsureArrayStateUsesTrackedStateFastPath(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}, 1)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	if arr.State != state {
		t.Fatalf("expected array wrapper to retain tracked state pointer")
	}
	if arr.TrackedHandle != arr.Handle || arr.Handle == 0 {
		t.Fatalf("expected tracked handle to match live handle, got tracked=%d handle=%d", arr.TrackedHandle, arr.Handle)
	}

	state.Values = append(state.Values, runtime.NewSmallInt(2, runtime.IntegerI32))
	interp.syncArrayValues(arr.Handle, state)

	fastState, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state fast path: %v", err)
	}
	if fastState != state {
		t.Fatalf("expected fast path to reuse tracked state pointer")
	}
	if len(fastState.Values) != 2 || len(arr.Elements) != 2 {
		t.Fatalf("expected tracked state and array elements to stay synchronized")
	}
}

func TestInterpreterSyncArrayValuesUpdatesCachedElementTypeToken(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}, 1)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	if !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeI32 {
		t.Fatalf("expected initial cached element token i32, got known=%v token=%d", state.ElementTypeTokenKnown, state.ElementTypeToken)
	}

	state.Values[0] = runtime.StringValue{Val: "x"}
	interp.syncArrayValues(arr.Handle, state)

	if !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeString {
		t.Fatalf("expected cached element token string after sync, got known=%v token=%d", state.ElementTypeTokenKnown, state.ElementTypeToken)
	}
}

func TestInterpreterSyncTrackedArrayWriteUpdatesSharedAliasesAndToken(t *testing.T) {
	interp := New()
	first := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}, 1)

	state, err := interp.ensureArrayState(first, 0)
	if err != nil {
		t.Fatalf("ensure first array state: %v", err)
	}
	second, err := interp.arrayValueFromHandle(first.Handle, 0, 0)
	if err != nil {
		t.Fatalf("arrayValueFromHandle: %v", err)
	}
	if !first.TrackedAliases || !second.TrackedAliases {
		t.Fatalf("expected alias pair to be marked shared before write sync")
	}

	written := runtime.StringValue{Val: "x"}
	state.Values[0] = written
	interp.syncTrackedArrayWrite(first, state, 0, written)

	if !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeString {
		t.Fatalf("expected tracked write to refresh element token, got known=%v token=%d", state.ElementTypeTokenKnown, state.ElementTypeToken)
	}
	if got, ok := first.Elements[0].(runtime.StringValue); !ok || got.Val != "x" {
		t.Fatalf("expected first alias to observe synced write, got %#v", first.Elements[0])
	}
	if got, ok := second.Elements[0].(runtime.StringValue); !ok || got.Val != "x" {
		t.Fatalf("expected second alias to observe synced write, got %#v", second.Elements[0])
	}
}
