package interpreter

import (
	"fmt"
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestHashMapBuiltins(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	ctx := &runtime.NativeCallContext{Env: global}
	newFn := mustNativeFunction(t, global, "__able_hash_map_new")
	setFn := mustNativeFunction(t, global, "__able_hash_map_set")
	getFn := mustNativeFunction(t, global, "__able_hash_map_get")
	removeFn := mustNativeFunction(t, global, "__able_hash_map_remove")
	containsFn := mustNativeFunction(t, global, "__able_hash_map_contains")
	sizeFn := mustNativeFunction(t, global, "__able_hash_map_size")
	clearFn := mustNativeFunction(t, global, "__able_hash_map_clear")
	cloneFn := mustNativeFunction(t, global, "__able_hash_map_clone")
	forEachFn := mustNativeFunction(t, global, "__able_hash_map_for_each")

	handleVal, err := newFn.Impl(ctx, nil)
	if err != nil {
		t.Fatalf("__able_hash_map_new failed: %v", err)
	}
	handle := mustInt64Value(t, handleVal)
	state, err := interp.hashMapStateForHandle(handle)
	if err != nil {
		t.Fatalf("missing hash map state: %v", err)
	}
	if len(state.Entries) != 0 {
		t.Fatalf("expected new hash map to be empty")
	}

	// set
	key := runtime.StringValue{Val: "hello"}
	value := runtime.IntegerValue{Val: big.NewInt(1), TypeSuffix: runtime.IntegerI32}
	if _, err := setFn.Impl(ctx, []runtime.Value{handleVal, key, value}); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if len(state.Entries) != 1 {
		t.Fatalf("set should add entry")
	}

	boolKey := runtime.BoolValue{Val: true}
	if _, err := setFn.Impl(ctx, []runtime.Value{handleVal, boolKey, runtime.StringValue{Val: "yes"}}); err != nil {
		t.Fatalf("set bool key failed: %v", err)
	}
	intKey := runtime.IntegerValue{Val: big.NewInt(-42), TypeSuffix: runtime.IntegerI32}
	if _, err := setFn.Impl(ctx, []runtime.Value{handleVal, intKey, runtime.StringValue{Val: "neg"}}); err != nil {
		t.Fatalf("set int key failed: %v", err)
	}
	if len(state.Entries) != 3 {
		t.Fatalf("expected three entries after inserting bool/int keys")
	}

	// contains
	containsRes, err := containsFn.Impl(ctx, []runtime.Value{handleVal, key})
	if err != nil {
		t.Fatalf("contains failed: %v", err)
	}
	if v, ok := containsRes.(runtime.BoolValue); !ok || !v.Val {
		t.Fatalf("contains should return true, got %#v", containsRes)
	}
	containsRes, err = containsFn.Impl(ctx, []runtime.Value{handleVal, boolKey})
	if err != nil {
		t.Fatalf("contains bool key failed: %v", err)
	}
	if v, ok := containsRes.(runtime.BoolValue); !ok || !v.Val {
		t.Fatalf("contains bool key should return true")
	}

	// get
	got, err := getFn.Impl(ctx, []runtime.Value{handleVal, key})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if v, ok := got.(runtime.IntegerValue); !ok || v.Val.Int64() != 1 {
		t.Fatalf("get returned unexpected value %#v", got)
	}
	boolValue, err := getFn.Impl(ctx, []runtime.Value{handleVal, boolKey})
	if err != nil {
		t.Fatalf("get bool key failed: %v", err)
	}
	if v, ok := boolValue.(runtime.StringValue); !ok || v.Val != "yes" {
		t.Fatalf("get bool key returned unexpected value %#v", boolValue)
	}
	intValue, err := getFn.Impl(ctx, []runtime.Value{handleVal, intKey})
	if err != nil {
		t.Fatalf("get int key failed: %v", err)
	}
	if v, ok := intValue.(runtime.StringValue); !ok || v.Val != "neg" {
		t.Fatalf("get int key returned unexpected value %#v", intValue)
	}
	missing, err := getFn.Impl(ctx, []runtime.Value{handleVal, runtime.StringValue{Val: "missing"}})
	if err != nil {
		t.Fatalf("get missing failed: %v", err)
	}
	if _, ok := missing.(runtime.NilValue); !ok {
		t.Fatalf("get missing should return nil, got %#v", missing)
	}

	// size
	sizeRes, err := sizeFn.Impl(ctx, []runtime.Value{handleVal})
	if err != nil {
		t.Fatalf("size failed: %v", err)
	}
	if v, ok := sizeRes.(runtime.IntegerValue); !ok || v.Val.Int64() != 3 {
		t.Fatalf("size returned unexpected value %#v", sizeRes)
	}

	// for_each
	var seen int
	visitor := runtime.NativeFunctionValue{
		Name:  "visit",
		Arity: 2,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("visit expects key and value")
			}
			seen++
			return runtime.NilValue{}, nil
		},
	}
	if _, err := forEachFn.Impl(ctx, []runtime.Value{handleVal, visitor}); err != nil {
		t.Fatalf("for_each failed: %v", err)
	}
	if seen != 3 {
		t.Fatalf("expected for_each to visit 3 entries, got %d", seen)
	}

	// clone
	cloneVal, err := cloneFn.Impl(ctx, []runtime.Value{handleVal})
	if err != nil {
		t.Fatalf("clone failed: %v", err)
	}
	cloneHandle := mustInt64Value(t, cloneVal)
	if cloneHandle == handle {
		t.Fatalf("clone should return a new handle")
	}
	cloneState, err := interp.hashMapStateForHandle(cloneHandle)
	if err != nil {
		t.Fatalf("missing clone state: %v", err)
	}
	if len(cloneState.Entries) != 3 {
		t.Fatalf("clone should copy entries, got %d", len(cloneState.Entries))
	}

	// remove
	removed, err := removeFn.Impl(ctx, []runtime.Value{handleVal, key})
	if err != nil {
		t.Fatalf("remove failed: %v", err)
	}
	if v, ok := removed.(runtime.IntegerValue); !ok || v.Val.Int64() != 1 {
		t.Fatalf("remove returned unexpected value %#v", removed)
	}
	if len(state.Entries) != 2 {
		t.Fatalf("remove should delete one entry")
	}
	if _, err := removeFn.Impl(ctx, []runtime.Value{handleVal, boolKey}); err != nil {
		t.Fatalf("remove bool key failed: %v", err)
	}
	if _, err := removeFn.Impl(ctx, []runtime.Value{handleVal, intKey}); err != nil {
		t.Fatalf("remove int key failed: %v", err)
	}
	if len(state.Entries) != 0 {
		t.Fatalf("all entries should be removed")
	}
	removed, err = removeFn.Impl(ctx, []runtime.Value{handleVal, key})
	if err != nil {
		t.Fatalf("remove missing failed: %v", err)
	}
	if _, ok := removed.(runtime.NilValue); !ok {
		t.Fatalf("remove missing should return nil")
	}

	// clear
	if _, err := setFn.Impl(ctx, []runtime.Value{handleVal, key, value}); err != nil {
		t.Fatalf("set before clear failed: %v", err)
	}
	if _, err := clearFn.Impl(ctx, []runtime.Value{handleVal}); err != nil {
		t.Fatalf("clear failed: %v", err)
	}
	if len(state.Entries) != 0 {
		t.Fatalf("clear should remove all entries")
	}

	// unsupported key type should error
	unsupported := runtime.NativeFunctionValue{Name: "fn", Arity: 0}
	if _, err := setFn.Impl(ctx, []runtime.Value{handleVal, unsupported, runtime.NilValue{}}); err == nil {
		t.Fatalf("expected error for unsupported key type")
	}
}

func TestHasherNativeMethods(t *testing.T) {
	interp := New()
	hasher := runtime.NewHasherValue()

	finishVal, err := interp.hasherMember(hasher, ast.NewIdentifier("finish"))
	if err != nil {
		t.Fatalf("finish lookup failed: %v", err)
	}
	finishBound, ok := finishVal.(*runtime.NativeBoundMethodValue)
	if !ok {
		t.Fatalf("finish binding unexpected type %T", finishVal)
	}

	ctx := &runtime.NativeCallContext{Env: interp.GlobalEnvironment()}
	initialRes, err := finishBound.Method.Impl(ctx, []runtime.Value{finishBound.Receiver})
	if err != nil {
		t.Fatalf("finish invocation failed: %v", err)
	}
	initialInt, ok := initialRes.(runtime.IntegerValue)
	if !ok || initialInt.Val == nil {
		t.Fatalf("finish returned unexpected value %#v", initialRes)
	}
	if initialInt.Val.Uint64() != runtime.NewHasherValue().Finish() {
		t.Fatalf("finish should return initial FNV offset, got %s", initialInt.Val.String())
	}

	writeBytesVal, err := interp.hasherMember(hasher, ast.NewIdentifier("write_bytes"))
	if err != nil {
		t.Fatalf("write_bytes lookup failed: %v", err)
	}
	writeBytesBound, ok := writeBytesVal.(*runtime.NativeBoundMethodValue)
	if !ok {
		t.Fatalf("write_bytes binding unexpected type %T", writeBytesVal)
	}
	if _, err := writeBytesBound.Method.Impl(ctx, []runtime.Value{writeBytesBound.Receiver, runtime.StringValue{Val: "abc"}}); err != nil {
		t.Fatalf("write_bytes invocation failed: %v", err)
	}

	writeU64Val, err := interp.hasherMember(hasher, ast.NewIdentifier("write_u64"))
	if err != nil {
		t.Fatalf("write_u64 lookup failed: %v", err)
	}
	writeU64Bound, ok := writeU64Val.(*runtime.NativeBoundMethodValue)
	if !ok {
		t.Fatalf("write_u64 binding unexpected type %T", writeU64Val)
	}
	if _, err := writeU64Bound.Method.Impl(ctx, []runtime.Value{
		writeU64Bound.Receiver,
		runtime.IntegerValue{Val: big.NewInt(5), TypeSuffix: runtime.IntegerU64},
	}); err != nil {
		t.Fatalf("write_u64 invocation failed: %v", err)
	}

	writeI64Val, err := interp.hasherMember(hasher, ast.NewIdentifier("write_i64"))
	if err != nil {
		t.Fatalf("write_i64 lookup failed: %v", err)
	}
	writeI64Bound, ok := writeI64Val.(*runtime.NativeBoundMethodValue)
	if !ok {
		t.Fatalf("write_i64 binding unexpected type %T", writeI64Val)
	}
	if _, err := writeI64Bound.Method.Impl(ctx, []runtime.Value{
		writeI64Bound.Receiver,
		runtime.IntegerValue{Val: big.NewInt(-2), TypeSuffix: runtime.IntegerI64},
	}); err != nil {
		t.Fatalf("write_i64 invocation failed: %v", err)
	}

	writeBoolVal, err := interp.hasherMember(hasher, ast.NewIdentifier("write_bool"))
	if err != nil {
		t.Fatalf("write_bool lookup failed: %v", err)
	}
	writeBoolBound, ok := writeBoolVal.(*runtime.NativeBoundMethodValue)
	if !ok {
		t.Fatalf("write_bool binding unexpected type %T", writeBoolVal)
	}
	if _, err := writeBoolBound.Method.Impl(ctx, []runtime.Value{
		writeBoolBound.Receiver,
		runtime.BoolValue{Val: true},
	}); err != nil {
		t.Fatalf("write_bool invocation failed: %v", err)
	}

	finalRes, err := finishBound.Method.Impl(ctx, []runtime.Value{finishBound.Receiver})
	if err != nil {
		t.Fatalf("finish after writes failed: %v", err)
	}
	finalInt, ok := finalRes.(runtime.IntegerValue)
	if !ok || finalInt.Val == nil {
		t.Fatalf("finish returned unexpected value %#v", finalRes)
	}

	control := runtime.NewHasherValue()
	control.WriteBytes([]byte("abc"))
	control.WriteUint64(5)
	control.WriteInt64(-2)
	control.WriteBool(true)
	if finalInt.Val.Uint64() != control.Finish() {
		t.Fatalf("finish produced %s, want %d", finalInt.Val.String(), control.Finish())
	}
}

func TestHashMapCustomHash(t *testing.T) {
	interp := New()

	u64Type := ast.IntegerTypeU64
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Key",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "id"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Key"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"hash",
					[]*ast.FunctionParameter{
						ast.Param("self", nil),
						ast.Param("hasher", nil),
					},
					[]ast.Statement{
						ast.Ret(ast.IntTyped(123, &u64Type)),
					},
					ast.Ty("u64"),
					nil,
					nil,
					false,
					false,
				),
				ast.Fn(
					"eq",
					[]*ast.FunctionParameter{
						ast.Param("self", nil),
						ast.Param("other", ast.Ty("Key")),
					},
					[]ast.Statement{
						ast.Ret(ast.Bin("==",
							ast.Member(ast.ID("self"), ast.ID("id")),
							ast.Member(ast.ID("other"), ast.ID("id")))),
					},
					ast.Ty("bool"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
	}, nil, nil)

	if _, _, err := interp.EvaluateModule(module); err != nil {
		t.Fatalf("module evaluation failed: %v", err)
	}

	keyExpr := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(7), "id"),
		},
		false,
		"Key",
		nil,
		nil,
	)
	keyVal, err := interp.evaluateExpression(keyExpr, interp.GlobalEnvironment())
	if err != nil {
		t.Fatalf("struct literal evaluation failed: %v", err)
	}
	keyInst, ok := keyVal.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct instance, got %T", keyVal)
	}

	hash, err := interp.hashMapHashValue(keyInst)
	if err != nil {
		t.Fatalf("hash computation failed: %v", err)
	}
	if hash != 123 {
		t.Fatalf("hash = %d, want 123", hash)
	}

	global := interp.GlobalEnvironment()
	ctx := &runtime.NativeCallContext{Env: global}
	newFn := mustNativeFunction(t, global, "__able_hash_map_new")
	setFn := mustNativeFunction(t, global, "__able_hash_map_set")
	getFn := mustNativeFunction(t, global, "__able_hash_map_get")

	handleVal, err := newFn.Impl(ctx, nil)
	if err != nil {
		t.Fatalf("__able_hash_map_new failed: %v", err)
	}
	handle := mustInt64Value(t, handleVal)
	if _, err := setFn.Impl(ctx, []runtime.Value{handleVal, keyInst, runtime.StringValue{Val: "value"}}); err != nil {
		t.Fatalf("set invocation failed: %v", err)
	}

	state, err := interp.hashMapStateForHandle(handle)
	if err != nil {
		t.Fatalf("missing hash map state: %v", err)
	}
	if len(state.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(state.Entries))
	}
	entry := state.Entries[0]
	if entry.Hash != hash {
		t.Fatalf("entry hash = %d, want %d", entry.Hash, hash)
	}
	if _, ok := entry.Value.(runtime.StringValue); !ok {
		t.Fatalf("entry value unexpected type %T", entry.Value)
	}
	if entry.Key != keyInst {
		t.Fatalf("entry key does not match inserted instance")
	}

	// A second instance with the same id should compare equal via eq().
	keyVal2, err := interp.evaluateExpression(keyExpr, interp.GlobalEnvironment())
	if err != nil {
		t.Fatalf("second struct literal evaluation failed: %v", err)
	}

	retrieved, err := getFn.Impl(ctx, []runtime.Value{handleVal, keyVal2})
	if err != nil {
		t.Fatalf("get with second key failed: %v", err)
	}
	strVal, ok := retrieved.(runtime.StringValue)
	if !ok || strVal.Val != "value" {
		t.Fatalf("get returned unexpected value %#v", retrieved)
	}
}

func mustNativeFunction(t *testing.T, env *runtime.Environment, name string) runtime.NativeFunctionValue {
	t.Helper()
	val, err := env.Get(name)
	if err != nil {
		t.Fatalf("missing native function %s: %v", name, err)
	}
	switch fn := val.(type) {
	case runtime.NativeFunctionValue:
		return fn
	case *runtime.NativeFunctionValue:
		return *fn
	default:
		t.Fatalf("unexpected native function binding %s: %T", name, val)
	}
	return runtime.NativeFunctionValue{}
}

func mustInt64Value(t *testing.T, value runtime.Value) int64 {
	t.Helper()
	switch v := value.(type) {
	case runtime.IntegerValue:
		if v.Val == nil || !v.Val.IsInt64() {
			t.Fatalf("expected int64 value, got %#v", value)
		}
		return v.Val.Int64()
	case *runtime.IntegerValue:
		if v == nil || v.Val == nil || !v.Val.IsInt64() {
			t.Fatalf("expected int64 value, got %#v", value)
		}
		return v.Val.Int64()
	default:
		t.Fatalf("expected integer value, got %T", value)
	}
	return 0
}
