package interpreter

import (
	"testing"
	"weak"

	"able/interpreter-go/pkg/ast"
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
	if tracking.single.Value() != arr {
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
	if tracking.single.Value() != nil {
		t.Fatalf("expected alias promotion to tracking set, got single=%#v", tracking.single.Value())
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
	if tracking.single.Value() != first {
		t.Fatalf("expected demotion back to single tracked alias, got %#v", tracking)
	}
	if tracking.many != nil {
		t.Fatalf("expected promoted tracking set to collapse after untracking second alias")
	}
	if first.TrackedAliases {
		t.Fatalf("expected remaining alias to return to exclusive tracking")
	}
}

func TestInterpreterArrayTrackerPrunesStaleWeakAliasesDuringSync(t *testing.T) {
	interp := New()
	arr := &runtime.ArrayValue{}
	state, handle, err := runtime.ArrayStoreEnsure(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	interp.trackArrayValue(handle, arr)
	arr.TrackedAliases = true
	interp.arraysByHandle[handle] = arrayHandleTracking{
		many: map[weak.Pointer[runtime.ArrayValue]]struct{}{
			weak.Make(arr):                     {},
			weak.Pointer[runtime.ArrayValue]{}: {},
		},
	}
	interp.syncArrayValues(handle, state)

	tracking, ok := interp.arraysByHandle[handle]
	if !ok {
		t.Fatalf("expected tracking entry for handle %d", handle)
	}
	if tracking.single.Value() != arr || tracking.many != nil {
		t.Fatalf("stale weak aliases were not collapsed: %#v", tracking)
	}
	if arr.TrackedAliases {
		t.Fatal("remaining live view should return to exclusive tracking")
	}
}

// Array aliases remain aliases after the value crosses a function boundary.
// This uses the canonical kernel Array representation rather than directly
// constructing runtime.ArrayValue views, so an ArrayStore lifetime change
// cannot release a raw storage_handle when one of its copies is still live.
func TestArrayHandleAliasingThroughFunctionReturnParity(t *testing.T) {
	module := mustParseModuleSource(t, `
import able.kernel.{Array}
import able.collections.array

fn extend(values: Array i32) -> Array i32 {
  alias := values
  alias.push(30)
  values
}

fn main() -> i32 {
  values: Array i32 = Array.new()
  values.push(10)
  alias := values
  returned := extend(alias)
  values.push(20)
  values.read_slot(0) * 10000 + alias.read_slot(1) * 100 + returned.read_slot(2)
}

main()
`)

	treeInterp := New()
	preloadArrayStdlibForTest(t, treeInterp)
	want := mustEvalModule(t, treeInterp, module)
	expected := runtime.NewSmallInt(103020, runtime.IntegerI32)
	if !valuesEqual(want, expected) {
		t.Fatalf("tree-walker Array alias result = %#v, want %#v", want, expected)
	}

	byteInterp := NewBytecode()
	preloadArrayStdlibForTest(t, byteInterp)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)
	if !valuesEqual(got, expected) {
		t.Fatalf("bytecode Array alias result = %#v, want %#v", got, expected)
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

func TestInterpreterEnsureArrayStateMaterializesTrackedRawValues(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue(nil, 0)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	vm.appendTrackedArrayValueFast(arr, state, bytecodeRawI32SlotCachedValue(7))
	if !state.ValuesMaterialized {
		t.Fatalf("tracked array append must materialize a raw VM value at the aggregate boundary")
	}

	fastState, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure tracked raw array state: %v", err)
	}
	if fastState != state {
		t.Fatalf("expected tracked array fast path to reuse state pointer")
	}
	if !state.ValuesMaterialized {
		t.Fatalf("ensureArrayState should materialize tracked raw values")
	}
	if len(state.Values) != 1 {
		t.Fatalf("tracked array length = %d, want 1", len(state.Values))
	}
	got, err := arrayIndexFromValue(state.Values[0])
	if err != nil || got != 7 {
		t.Fatalf("materialized tracked value = %#v (%v), want 7", state.Values[0], err)
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
	if !state.CachedI32ValuesKnown || len(state.CachedI32Values) != 1 || state.CachedI32Values[0] != 1 {
		t.Fatalf("expected initial cached i32 values [1], got known=%v values=%v", state.CachedI32ValuesKnown, state.CachedI32Values)
	}

	state.Values[0] = runtime.StringValue{Val: "x"}
	interp.syncArrayValues(arr.Handle, state)

	if !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeString {
		t.Fatalf("expected cached element token string after sync, got known=%v token=%d", state.ElementTypeTokenKnown, state.ElementTypeToken)
	}
	if state.CachedI32ValuesKnown || state.CachedI32Values != nil {
		t.Fatalf("expected cached i32 values to be cleared after string sync, got known=%v values=%v", state.CachedI32ValuesKnown, state.CachedI32Values)
	}
}

func TestInterpreterNestedArrayTypeInspectionUsesTrackedState(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	row := interp.newArrayValue(nil, 0)
	rowState, err := interp.ensureArrayState(row, 0)
	if err != nil {
		t.Fatalf("ensure row array state: %v", err)
	}
	vm.appendTrackedArrayValueFast(row, rowState, bytecodeRawFloatSlotValue(1.25, runtime.FloatF64))
	row.Elements = nil

	outer := interp.newArrayValue(nil, 0)
	outerState, err := interp.ensureArrayState(outer, 0)
	if err != nil {
		t.Fatalf("ensure outer array state: %v", err)
	}
	vm.appendTrackedArrayValueFast(outer, outerState, row)
	outer.Elements = nil

	got := typeExpressionToString(interp.typeExpressionFromValue(outer))
	if got != "Array<Array<f64>>" {
		t.Fatalf("typeExpressionFromValue(outer) = %s, want Array<Array<f64>>", got)
	}
	if !interp.matchesType(ast.Gen(ast.Ty("Array"), ast.Gen(ast.Ty("Array"), ast.Ty("f64"))), outer) {
		t.Fatalf("matchesType(Array<Array<f64>>, outer) = false, want true")
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
	if state.CachedI32ValuesKnown || state.CachedI32Values != nil {
		t.Fatalf("expected tracked write to clear cached i32 values, got known=%v values=%v", state.CachedI32ValuesKnown, state.CachedI32Values)
	}
	if got, ok := first.Elements[0].(runtime.StringValue); !ok || got.Val != "x" {
		t.Fatalf("expected first alias to observe synced write, got %#v", first.Elements[0])
	}
	if got, ok := second.Elements[0].(runtime.StringValue); !ok || got.Val != "x" {
		t.Fatalf("expected second alias to observe synced write, got %#v", second.Elements[0])
	}
}

func TestInterpreterSyncTrackedArrayWriteMarksRawStateUnmaterialized(t *testing.T) {
	interp := NewBytecode()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}, 1)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	if !state.ValuesMaterialized {
		t.Fatalf("expected initial tracked array state to be materialized")
	}

	written := bytecodeRawI32SlotCachedValue(9)
	state.Values[0] = written
	interp.syncTrackedArrayWrite(arr, state, 0, written)

	if state.ValuesMaterialized {
		t.Fatalf("tracked raw write should mark array state as needing materialization")
	}
}

func TestInterpreterEnsureArrayStateForMetadataPreservesMaterializedValues(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue(nil, 0)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	vm.appendTrackedArrayValueFast(arr, state, bytecodeRawI32SlotCachedValue(7))
	if !state.ValuesMaterialized {
		t.Fatalf("tracked array append must materialize raw values")
	}

	metaState, err := interp.ensureArrayStateForMetadata(arr, 0)
	if err != nil {
		t.Fatalf("ensure metadata state: %v", err)
	}
	if metaState != state {
		t.Fatalf("metadata state pointer changed: got %p want %p", metaState, state)
	}
	if !state.ValuesMaterialized {
		t.Fatalf("metadata-only state access must preserve materialized aggregate values")
	}
}

func TestBytecodeVMArrayLengthAssignmentPreservesMaterializedValues(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue(nil, 0)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	vm.appendTrackedArrayValueFast(arr, state, bytecodeRawI32SlotCachedValue(7))
	if !state.ValuesMaterialized {
		t.Fatalf("tracked array append must materialize raw values")
	}

	result, err := vm.assignMemberValue(arr, &ast.Identifier{Name: "length"}, runtime.NewSmallInt(0, runtime.IntegerI32), ast.AssignmentAssign, "", false)
	if err != nil {
		t.Fatalf("assign member value: %v", err)
	}
	if _, ok := result.(runtime.IntegerValue); !ok {
		t.Fatalf("length assignment result type = %T, want runtime.IntegerValue", result)
	}
	if !state.ValuesMaterialized {
		t.Fatalf("array length assignment must preserve materialized aggregate values")
	}
	if len(state.Values) != 0 {
		t.Fatalf("array length after trim = %d, want 0", len(state.Values))
	}
}

func TestInterpreterMemberAssignArrayCapacityPreservesMaterializedValues(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue(nil, 0)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	vm.appendTrackedArrayValueFast(arr, state, bytecodeRawI32SlotCachedValue(7))
	if !state.ValuesMaterialized {
		t.Fatalf("tracked array append must materialize raw values")
	}

	member := runtime.StringValue{Val: "capacity"}
	result, err := interp.MemberAssign(arr, member, runtime.NewSmallInt(8, runtime.IntegerI32), interp.GlobalEnvironment())
	if err != nil {
		t.Fatalf("member assign: %v", err)
	}
	if _, ok := result.(runtime.IntegerValue); !ok {
		t.Fatalf("capacity assignment result type = %T, want runtime.IntegerValue", result)
	}
	if !state.ValuesMaterialized {
		t.Fatalf("array capacity assignment must preserve materialized aggregate values")
	}
	if state.Capacity < 8 {
		t.Fatalf("array capacity after assignment = %d, want >= 8", state.Capacity)
	}
}

func TestInterpreterSyncTrackedArrayWriteUpdatesCachedI32Values(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}, 2)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	if !state.CachedI32ValuesKnown || len(state.CachedI32Values) != 2 {
		t.Fatalf("expected initial cached i32 values, got known=%v values=%v", state.CachedI32ValuesKnown, state.CachedI32Values)
	}

	written := runtime.NewSmallInt(7, runtime.IntegerI32)
	state.Values[1] = written
	interp.syncTrackedArrayWrite(arr, state, 1, written)

	if !state.CachedI32ValuesKnown || len(state.CachedI32Values) != 2 {
		t.Fatalf("expected cached i32 values after tracked i32 write, got known=%v values=%v", state.CachedI32ValuesKnown, state.CachedI32Values)
	}
	if state.CachedI32Values[0] != 1 || state.CachedI32Values[1] != 7 {
		t.Fatalf("cached i32 values = %v, want [1 7]", state.CachedI32Values)
	}
}

func TestInterpreterSyncTrackedArrayWriteBuildsPartialCachedI32ValuesForHoleyArray(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue([]runtime.Value{nil, nil, nil}, 3)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	if state.ElementTypeTokenKnown {
		t.Fatalf("expected initial element token to stay unknown for holey array")
	}

	first := runtime.NewSmallInt(7, runtime.IntegerI32)
	state.Values[0] = first
	interp.syncTrackedArrayWrite(arr, state, 0, first)

	if !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeI32 {
		t.Fatalf("expected tracked holey write to infer i32 token, got known=%v token=%d", state.ElementTypeTokenKnown, state.ElementTypeToken)
	}
	if state.CachedI32ValuesKnown {
		t.Fatalf("expected partial cached i32 values after first holey write")
	}
	if len(state.CachedI32Values) != 3 || len(state.CachedI32ValuesValid) != 3 || state.CachedI32ValuesCount != 1 {
		t.Fatalf("partial cached i32 metadata = values=%v valid=%v count=%d", state.CachedI32Values, state.CachedI32ValuesValid, state.CachedI32ValuesCount)
	}
	if raw, ok := trackedArrayCachedI32RawAt(state, 0); !ok || raw != 7 {
		t.Fatalf("cached raw i32 at slot 0 = %d/%v, want 7/true", raw, ok)
	}
	if _, ok := trackedArrayCachedI32RawAt(state, 1); ok {
		t.Fatalf("expected unwritten hole slot 1 to remain uncached")
	}

	last := runtime.NewSmallInt(9, runtime.IntegerI32)
	state.Values[2] = last
	interp.syncTrackedArrayWrite(arr, state, 2, last)

	if state.CachedI32ValuesKnown {
		t.Fatalf("expected holey cache to remain partial before all slots are written")
	}
	if state.CachedI32ValuesCount != 2 {
		t.Fatalf("partial cached i32 count after second write = %d, want 2", state.CachedI32ValuesCount)
	}
	if raw, ok := trackedArrayCachedI32RawAt(state, 2); !ok || raw != 9 {
		t.Fatalf("cached raw i32 at slot 2 = %d/%v, want 9/true", raw, ok)
	}

	middle := runtime.NewSmallInt(8, runtime.IntegerI32)
	state.Values[1] = middle
	interp.syncTrackedArrayWrite(arr, state, 1, middle)

	if !state.CachedI32ValuesKnown || state.CachedI32ValuesCount != 3 {
		t.Fatalf("expected full cached i32 coverage after filling holey array, got known=%v count=%d", state.CachedI32ValuesKnown, state.CachedI32ValuesCount)
	}
	if raw, ok := trackedArrayCachedI32RawAt(state, 1); !ok || raw != 8 {
		t.Fatalf("cached raw i32 at slot 1 = %d/%v, want 8/true", raw, ok)
	}
}

func TestRefreshTrackedArrayI32RawCacheTreatsNilValuesAsHoles(t *testing.T) {
	state := &arrayState{
		Values: []runtime.Value{
			runtime.NewSmallInt(1, runtime.IntegerI32),
			runtime.NilValue{},
			runtime.NewSmallInt(3, runtime.IntegerI32),
		},
		ElementTypeToken:      bytecodeIndexTypeI32,
		ElementTypeTokenKnown: true,
	}

	refreshTrackedArrayI32RawCache(state)

	if state.CachedI32ValuesKnown {
		t.Fatalf("expected nil slot to keep cache partial, got known=%v", state.CachedI32ValuesKnown)
	}
	if len(state.CachedI32Values) != 3 || len(state.CachedI32ValuesValid) != 3 || state.CachedI32ValuesCount != 2 {
		t.Fatalf("cached nil-hole i32 metadata = values=%v valid=%v count=%d", state.CachedI32Values, state.CachedI32ValuesValid, state.CachedI32ValuesCount)
	}
	if raw, ok := trackedArrayCachedI32RawAt(state, 0); !ok || raw != 1 {
		t.Fatalf("cached raw i32 at slot 0 = %d/%v, want 1/true", raw, ok)
	}
	if _, ok := trackedArrayCachedI32RawAt(state, 1); ok {
		t.Fatalf("expected typed nil slot 1 to remain uncached")
	}
	if raw, ok := trackedArrayCachedI32RawAt(state, 2); !ok || raw != 3 {
		t.Fatalf("cached raw i32 at slot 2 = %d/%v, want 3/true", raw, ok)
	}
}

func TestInterpreterSyncTrackedArrayWriteNilInvalidatesOnlyTargetCachedI32Slot(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
		runtime.NewSmallInt(3, runtime.IntegerI32),
	}, 3)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	written := runtime.NilValue{}
	state.Values[1] = written
	interp.syncTrackedArrayWrite(arr, state, 1, written)

	if !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeI32 {
		t.Fatalf("expected typed nil write to keep i32 token, got known=%v token=%d", state.ElementTypeTokenKnown, state.ElementTypeToken)
	}
	if state.CachedI32ValuesKnown {
		t.Fatalf("expected typed nil write to leave cache partial")
	}
	if len(state.CachedI32Values) != 3 || len(state.CachedI32ValuesValid) != 3 || state.CachedI32ValuesCount != 2 {
		t.Fatalf("cached i32 metadata after typed nil write = values=%v valid=%v count=%d", state.CachedI32Values, state.CachedI32ValuesValid, state.CachedI32ValuesCount)
	}
	if _, ok := trackedArrayCachedI32RawAt(state, 1); ok {
		t.Fatalf("expected typed nil slot 1 to be uncached after write")
	}
}

func TestInterpreterSyncArrayHandleLengthKeepsCachedI32ValuesOnTrim(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
		runtime.NewSmallInt(3, runtime.IntegerI32),
	}, 3)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	revisionBefore := state.Revision

	setArrayLength(state, 2)
	if state.Revision != revisionBefore+1 {
		t.Fatalf("revision after trim = %d, want %d", state.Revision, revisionBefore+1)
	}

	interp.syncArrayHandleLength(arr.Handle, state)

	if state.Revision != revisionBefore+1 {
		t.Fatalf("length sync should not add another revision, got %d want %d", state.Revision, revisionBefore+1)
	}
	if !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeI32 {
		t.Fatalf("element token after trim = known=%v token=%d, want i32", state.ElementTypeTokenKnown, state.ElementTypeToken)
	}
	if !state.CachedI32ValuesKnown || len(state.CachedI32Values) != 2 || state.CachedI32ValuesCount != 2 {
		t.Fatalf("cached i32 values after trim = known=%v count=%d values=%v", state.CachedI32ValuesKnown, state.CachedI32ValuesCount, state.CachedI32Values)
	}
	if state.CachedI32Values[0] != 1 || state.CachedI32Values[1] != 2 {
		t.Fatalf("cached i32 values after trim = %v, want [1 2]", state.CachedI32Values)
	}
}

func TestInterpreterSyncArrayHandleLengthFromEmptyUsesNilToken(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue(nil, 0)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	revisionBefore := state.Revision

	setArrayLength(state, 2)
	if state.Revision != revisionBefore+1 {
		t.Fatalf("revision after grow = %d, want %d", state.Revision, revisionBefore+1)
	}

	interp.syncArrayHandleLength(arr.Handle, state)

	if state.Revision != revisionBefore+1 {
		t.Fatalf("length sync should not add another revision, got %d want %d", state.Revision, revisionBefore+1)
	}
	if !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeNil {
		t.Fatalf("element token after empty grow = known=%v token=%d, want nil", state.ElementTypeTokenKnown, state.ElementTypeToken)
	}
	if state.CachedI32ValuesKnown || state.CachedI32Values != nil || state.CachedI32ValuesValid != nil {
		t.Fatalf("expected i32 cache to stay clear after nil-filled grow, got known=%v values=%v valid=%v", state.CachedI32ValuesKnown, state.CachedI32Values, state.CachedI32ValuesValid)
	}
}

func TestInterpreterSyncArrayHandleMetadataPreservesRevisionAndCachedI32Values(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(4, runtime.IntegerI32),
		runtime.NewSmallInt(5, runtime.IntegerI32),
	}, 2)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	revisionBefore := state.Revision

	interp.syncArrayHandleMetadata(arr.Handle, state)

	if state.Revision != revisionBefore {
		t.Fatalf("metadata sync revision = %d, want %d", state.Revision, revisionBefore)
	}
	if !state.CachedI32ValuesKnown || state.CachedI32ValuesCount != 2 || len(state.CachedI32Values) != 2 {
		t.Fatalf("cached i32 values after metadata sync = known=%v count=%d values=%v", state.CachedI32ValuesKnown, state.CachedI32ValuesCount, state.CachedI32Values)
	}
	if state.CachedI32Values[0] != 4 || state.CachedI32Values[1] != 5 {
		t.Fatalf("cached i32 values after metadata sync = %v, want [4 5]", state.CachedI32Values)
	}
}

func TestInterpreterSyncArrayHandleWriteAppendKeepsCachedI32Values(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}, 2)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	appended := runtime.NewSmallInt(3, runtime.IntegerI32)
	if err := runtime.ArrayStoreWrite(arr.Handle, 2, appended); err != nil {
		t.Fatalf("ArrayStoreWrite append: %v", err)
	}
	handleState, err := runtime.ArrayStoreState(arr.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreState: %v", err)
	}

	interp.syncArrayHandleWrite(arr.Handle, handleState, 2, appended)

	if handleState != state {
		t.Fatalf("handle state pointer changed: got %p want %p", handleState, state)
	}
	if !state.CachedI32ValuesKnown || state.CachedI32ValuesCount != 3 {
		t.Fatalf("cached i32 values after handle append = known=%v count=%d", state.CachedI32ValuesKnown, state.CachedI32ValuesCount)
	}
	if len(state.CachedI32Values) != 3 || len(state.CachedI32ValuesValid) != 3 {
		t.Fatalf("cached i32 lengths after handle append = (%d, %d), want (3, 3)", len(state.CachedI32Values), len(state.CachedI32ValuesValid))
	}
	if state.CachedI32Values[0] != 1 || state.CachedI32Values[1] != 2 || state.CachedI32Values[2] != 3 {
		t.Fatalf("cached i32 values after handle append = %v, want [1 2 3]", state.CachedI32Values)
	}
}

func TestInterpreterSyncTrackedArrayWriteAppendExtendsCachedI32Values(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}, 2)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	appended := runtime.NewSmallInt(3, runtime.IntegerI32)
	state.Values = append(state.Values, appended)

	interp.syncTrackedArrayWrite(arr, state, 2, appended)

	if !state.CachedI32ValuesKnown || state.CachedI32ValuesCount != 3 {
		t.Fatalf("cached i32 values after append = known=%v count=%d", state.CachedI32ValuesKnown, state.CachedI32ValuesCount)
	}
	if len(state.CachedI32Values) != 3 || len(state.CachedI32ValuesValid) != 3 {
		t.Fatalf("cached i32 lengths after append = (%d, %d), want (3, 3)", len(state.CachedI32Values), len(state.CachedI32ValuesValid))
	}
	if state.CachedI32Values[0] != 1 || state.CachedI32Values[1] != 2 || state.CachedI32Values[2] != 3 {
		t.Fatalf("cached i32 values after append = %v, want [1 2 3]", state.CachedI32Values)
	}
}

func TestInterpreterSyncTrackedArrayWriteExtendsCachedI32ValuesAcrossNonHoleSuffixGrowth(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}, 1)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	middle := runtime.NewSmallInt(2, runtime.IntegerI32)
	appended := runtime.NewSmallInt(3, runtime.IntegerI32)
	state.Values = append(state.Values, middle, appended)

	interp.syncTrackedArrayWrite(arr, state, 2, appended)

	if !state.CachedI32ValuesKnown || state.CachedI32ValuesCount != 3 {
		t.Fatalf("cached i32 values after non-hole suffix growth = known=%v count=%d", state.CachedI32ValuesKnown, state.CachedI32ValuesCount)
	}
	if len(state.CachedI32Values) != 3 || len(state.CachedI32ValuesValid) != 3 {
		t.Fatalf("cached i32 lengths after non-hole suffix growth = (%d, %d), want (3, 3)", len(state.CachedI32Values), len(state.CachedI32ValuesValid))
	}
	if state.CachedI32Values[0] != 1 || state.CachedI32Values[1] != 2 || state.CachedI32Values[2] != 3 {
		t.Fatalf("cached i32 values after non-hole suffix growth = %v, want [1 2 3]", state.CachedI32Values)
	}
}

func TestInterpreterSyncTrackedArrayWriteFirstAppendBuildsCachedI32ValuesFromEmpty(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue(nil, 0)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	appended := runtime.NewSmallInt(3, runtime.IntegerI32)
	state.Values = append(state.Values, appended)

	interp.syncTrackedArrayWrite(arr, state, 0, appended)

	if !state.CachedI32ValuesKnown || state.CachedI32ValuesCount != 1 {
		t.Fatalf("cached i32 values after first append = known=%v count=%d", state.CachedI32ValuesKnown, state.CachedI32ValuesCount)
	}
	if len(state.CachedI32Values) != 1 || len(state.CachedI32ValuesValid) != 1 {
		t.Fatalf("cached i32 lengths after first append = (%d, %d), want (1, 1)", len(state.CachedI32Values), len(state.CachedI32ValuesValid))
	}
	if state.CachedI32Values[0] != 3 || !state.CachedI32ValuesValid[0] {
		t.Fatalf("cached i32 values after first append = values=%v valid=%v, want [3]/[true]", state.CachedI32Values, state.CachedI32ValuesValid)
	}
}

func TestInterpreterSyncTrackedArrayWriteExtendsCachedI32ValuesAcrossTypedNilGrowth(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}, 1)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	appended := runtime.NewSmallInt(4, runtime.IntegerI32)
	state.Values = append(state.Values, runtime.NilValue{}, runtime.NilValue{}, appended)

	interp.syncTrackedArrayWrite(arr, state, 3, appended)

	if state.CachedI32ValuesKnown {
		t.Fatalf("expected typed nil growth to keep cache partial")
	}
	if len(state.CachedI32Values) != 4 || len(state.CachedI32ValuesValid) != 4 || state.CachedI32ValuesCount != 2 {
		t.Fatalf("cached i32 metadata after typed nil growth = values=%v valid=%v count=%d", state.CachedI32Values, state.CachedI32ValuesValid, state.CachedI32ValuesCount)
	}
	if raw, ok := trackedArrayCachedI32RawAt(state, 0); !ok || raw != 1 {
		t.Fatalf("cached raw i32 at slot 0 = %d/%v, want 1/true", raw, ok)
	}
	if _, ok := trackedArrayCachedI32RawAt(state, 1); ok {
		t.Fatalf("expected typed nil slot 1 to remain uncached")
	}
	if _, ok := trackedArrayCachedI32RawAt(state, 2); ok {
		t.Fatalf("expected typed nil slot 2 to remain uncached")
	}
	if raw, ok := trackedArrayCachedI32RawAt(state, 3); !ok || raw != 4 {
		t.Fatalf("cached raw i32 at slot 3 = %d/%v, want 4/true", raw, ok)
	}
}

func TestInterpreterSyncTrackedArrayWriteShrinksCachedI32ValuesBeforeWrite(t *testing.T) {
	interp := New()
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
		runtime.NewSmallInt(3, runtime.IntegerI32),
	}, 3)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	replacement := runtime.NewSmallInt(7, runtime.IntegerI32)
	state.Values = state.Values[:2]
	state.Values[1] = replacement

	interp.syncTrackedArrayWrite(arr, state, 1, replacement)

	if !state.CachedI32ValuesKnown || state.CachedI32ValuesCount != 2 {
		t.Fatalf("cached i32 values after shrink/write = known=%v count=%d", state.CachedI32ValuesKnown, state.CachedI32ValuesCount)
	}
	if len(state.CachedI32Values) != 2 || len(state.CachedI32ValuesValid) != 2 {
		t.Fatalf("cached i32 lengths after shrink/write = (%d, %d), want (2, 2)", len(state.CachedI32Values), len(state.CachedI32ValuesValid))
	}
	if state.CachedI32Values[0] != 1 || state.CachedI32Values[1] != 7 {
		t.Fatalf("cached i32 values after shrink/write = %v, want [1 7]", state.CachedI32Values)
	}
}

func TestRefreshTrackedArrayI32RawCacheReusesBacking(t *testing.T) {
	state := &arrayState{
		Values: []runtime.Value{
			runtime.NewSmallInt(1, runtime.IntegerI32),
			runtime.NewSmallInt(2, runtime.IntegerI32),
			runtime.NewSmallInt(3, runtime.IntegerI32),
		},
		ElementTypeToken:      bytecodeIndexTypeI32,
		ElementTypeTokenKnown: true,
		CachedI32Values:       make([]int32, 3, 8),
		CachedI32ValuesValid:  make([]bool, 3, 8),
	}

	allocs := testing.AllocsPerRun(1000, func() {
		refreshTrackedArrayI32RawCache(state)
	})
	if allocs > 0.1 {
		t.Fatalf("refreshTrackedArrayI32RawCache allocs = %.2f, want <= 0.1", allocs)
	}
}
