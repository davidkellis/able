package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestHashMapBuiltins(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	val, err := global.Get("HashMap")
	if err != nil {
		t.Fatalf("HashMap package not registered: %v", err)
	}

	var pkg *runtime.PackageValue
	switch v := val.(type) {
	case *runtime.PackageValue:
		pkg = v
	case runtime.PackageValue:
		pkg = &v
	default:
		t.Fatalf("unexpected HashMap binding type %T", val)
	}

	newSym, ok := pkg.Public["new"]
	if !ok {
		t.Fatalf("HashMap.new missing from package")
	}

	var newFn runtime.NativeFunctionValue
	switch fn := newSym.(type) {
	case runtime.NativeFunctionValue:
		newFn = fn
	case *runtime.NativeFunctionValue:
		newFn = *fn
	default:
		t.Fatalf("HashMap.new unexpected type %T", newSym)
	}

	ctx := &runtime.NativeCallContext{Env: global}
	mapVal, err := newFn.Impl(ctx, nil)
	if err != nil {
		t.Fatalf("HashMap.new failed: %v", err)
	}

	hm, ok := mapVal.(*runtime.HashMapValue)
	if !ok {
		t.Fatalf("HashMap.new returned unexpected value %T", mapVal)
	}
	if len(hm.Entries) != 0 {
		t.Fatalf("expected new hash map to be empty")
	}

	// set
	setVal, err := interp.hashMapMember(hm, ast.NewIdentifier("set"))
	if err != nil {
		t.Fatalf("set lookup failed: %v", err)
	}
	setFn := setVal.(*runtime.NativeBoundMethodValue)
	key := runtime.StringValue{Val: "hello"}
	value := runtime.IntegerValue{Val: big.NewInt(1), TypeSuffix: runtime.IntegerI32}
	if _, err := setFn.Method.Impl(ctx, []runtime.Value{setFn.Receiver, key, value}); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if len(hm.Entries) != 1 {
		t.Fatalf("set should add entry")
	}

	boolKey := runtime.BoolValue{Val: true}
	if _, err := setFn.Method.Impl(ctx, []runtime.Value{setFn.Receiver, boolKey, runtime.StringValue{Val: "yes"}}); err != nil {
		t.Fatalf("set bool key failed: %v", err)
	}
	intKey := runtime.IntegerValue{Val: big.NewInt(-42), TypeSuffix: runtime.IntegerI32}
	if _, err := setFn.Method.Impl(ctx, []runtime.Value{setFn.Receiver, intKey, runtime.StringValue{Val: "neg"}}); err != nil {
		t.Fatalf("set int key failed: %v", err)
	}
	if len(hm.Entries) != 3 {
		t.Fatalf("expected three entries after inserting bool/int keys")
	}

	// contains
	containsVal, err := interp.hashMapMember(hm, ast.NewIdentifier("contains"))
	if err != nil {
		t.Fatalf("contains lookup failed: %v", err)
	}
	containsFn := containsVal.(*runtime.NativeBoundMethodValue)
	containsRes, err := containsFn.Method.Impl(ctx, []runtime.Value{containsFn.Receiver, key})
	if err != nil {
		t.Fatalf("contains failed: %v", err)
	}
	if v, ok := containsRes.(runtime.BoolValue); !ok || !v.Val {
		t.Fatalf("contains should return true, got %#v", containsRes)
	}
	containsRes, err = containsFn.Method.Impl(ctx, []runtime.Value{containsFn.Receiver, boolKey})
	if err != nil {
		t.Fatalf("contains bool key failed: %v", err)
	}
	if v, ok := containsRes.(runtime.BoolValue); !ok || !v.Val {
		t.Fatalf("contains bool key should return true")
	}

	// get
	getVal, err := interp.hashMapMember(hm, ast.NewIdentifier("get"))
	if err != nil {
		t.Fatalf("get lookup failed: %v", err)
	}
	getFn := getVal.(*runtime.NativeBoundMethodValue)
	got, err := getFn.Method.Impl(ctx, []runtime.Value{getFn.Receiver, key})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if v, ok := got.(runtime.IntegerValue); !ok || v.Val.Int64() != 1 {
		t.Fatalf("get returned unexpected value %#v", got)
	}
	boolValue, err := getFn.Method.Impl(ctx, []runtime.Value{getFn.Receiver, boolKey})
	if err != nil {
		t.Fatalf("get bool key failed: %v", err)
	}
	if v, ok := boolValue.(runtime.StringValue); !ok || v.Val != "yes" {
		t.Fatalf("get bool key returned unexpected value %#v", boolValue)
	}
	intValue, err := getFn.Method.Impl(ctx, []runtime.Value{getFn.Receiver, intKey})
	if err != nil {
		t.Fatalf("get int key failed: %v", err)
	}
	if v, ok := intValue.(runtime.StringValue); !ok || v.Val != "neg" {
		t.Fatalf("get int key returned unexpected value %#v", intValue)
	}
	missing, err := getFn.Method.Impl(ctx, []runtime.Value{getFn.Receiver, runtime.StringValue{Val: "missing"}})
	if err != nil {
		t.Fatalf("get missing failed: %v", err)
	}
	if _, ok := missing.(runtime.NilValue); !ok {
		t.Fatalf("get missing should return nil, got %#v", missing)
	}

	// size
	sizeVal, err := interp.hashMapMember(hm, ast.NewIdentifier("size"))
	if err != nil {
		t.Fatalf("size lookup failed: %v", err)
	}
	sizeFn := sizeVal.(*runtime.NativeBoundMethodValue)
	sizeRes, err := sizeFn.Method.Impl(ctx, []runtime.Value{sizeFn.Receiver})
	if err != nil {
		t.Fatalf("size failed: %v", err)
	}
	if v, ok := sizeRes.(runtime.IntegerValue); !ok || v.Val.Int64() != 3 {
		t.Fatalf("size returned unexpected value %#v", sizeRes)
	}

	// remove
	removeVal, err := interp.hashMapMember(hm, ast.NewIdentifier("remove"))
	if err != nil {
		t.Fatalf("remove lookup failed: %v", err)
	}
	removeFn := removeVal.(*runtime.NativeBoundMethodValue)
	removed, err := removeFn.Method.Impl(ctx, []runtime.Value{removeFn.Receiver, key})
	if err != nil {
		t.Fatalf("remove failed: %v", err)
	}
	if v, ok := removed.(runtime.IntegerValue); !ok || v.Val.Int64() != 1 {
		t.Fatalf("remove returned unexpected value %#v", removed)
	}
	if len(hm.Entries) != 2 {
		t.Fatalf("remove should delete one entry")
	}
	if _, err := removeFn.Method.Impl(ctx, []runtime.Value{removeFn.Receiver, boolKey}); err != nil {
		t.Fatalf("remove bool key failed: %v", err)
	}
	if _, err := removeFn.Method.Impl(ctx, []runtime.Value{removeFn.Receiver, intKey}); err != nil {
		t.Fatalf("remove int key failed: %v", err)
	}
	if len(hm.Entries) != 0 {
		t.Fatalf("all entries should be removed")
	}
	removed, err = removeFn.Method.Impl(ctx, []runtime.Value{removeFn.Receiver, key})
	if err != nil {
		t.Fatalf("remove missing failed: %v", err)
	}
	if _, ok := removed.(runtime.NilValue); !ok {
		t.Fatalf("remove missing should return nil")
	}

	// clear
	if _, err := setFn.Method.Impl(ctx, []runtime.Value{setFn.Receiver, key, value}); err != nil {
		t.Fatalf("set before clear failed: %v", err)
	}
	clearVal, err := interp.hashMapMember(hm, ast.NewIdentifier("clear"))
	if err != nil {
		t.Fatalf("clear lookup failed: %v", err)
	}
	clearFn := clearVal.(*runtime.NativeBoundMethodValue)
	if _, err := clearFn.Method.Impl(ctx, []runtime.Value{clearFn.Receiver}); err != nil {
		t.Fatalf("clear failed: %v", err)
	}
	if len(hm.Entries) != 0 {
		t.Fatalf("clear should remove all entries")
	}

	// unsupported key type should error
	unsupported := runtime.NativeFunctionValue{Name: "fn", Arity: 0}
	if _, err := setFn.Method.Impl(ctx, []runtime.Value{setFn.Receiver, unsupported, runtime.NilValue{}}); err == nil {
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
	val, err := global.Get("HashMap")
	if err != nil {
		t.Fatalf("HashMap package not registered: %v", err)
	}

	var pkg *runtime.PackageValue
	switch v := val.(type) {
	case *runtime.PackageValue:
		pkg = v
	case runtime.PackageValue:
		pkg = &v
	default:
		t.Fatalf("unexpected HashMap binding type %T", val)
	}

	newSym, ok := pkg.Public["new"]
	if !ok {
		t.Fatalf("HashMap.new missing from package")
	}

	var newFn runtime.NativeFunctionValue
	switch fn := newSym.(type) {
	case runtime.NativeFunctionValue:
		newFn = fn
	case *runtime.NativeFunctionValue:
		newFn = *fn
	default:
		t.Fatalf("HashMap.new unexpected type %T", newSym)
	}

	ctx := &runtime.NativeCallContext{Env: global}
	mapVal, err := newFn.Impl(ctx, nil)
	if err != nil {
		t.Fatalf("HashMap.new failed: %v", err)
	}

	hm, ok := mapVal.(*runtime.HashMapValue)
	if !ok {
		t.Fatalf("HashMap.new returned unexpected value %T", mapVal)
	}

	setVal, err := interp.hashMapMember(hm, ast.NewIdentifier("set"))
	if err != nil {
		t.Fatalf("set lookup failed: %v", err)
	}
	setFn := setVal.(*runtime.NativeBoundMethodValue)
	if _, err := setFn.Method.Impl(ctx, []runtime.Value{setFn.Receiver, keyInst, runtime.StringValue{Val: "value"}}); err != nil {
		t.Fatalf("set invocation failed: %v", err)
	}

	if len(hm.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(hm.Entries))
	}
	entry := hm.Entries[0]
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

	getVal, err := interp.hashMapMember(hm, ast.NewIdentifier("get"))
	if err != nil {
		t.Fatalf("get lookup failed: %v", err)
	}
	getFn := getVal.(*runtime.NativeBoundMethodValue)
	retrieved, err := getFn.Method.Impl(ctx, []runtime.Value{getFn.Receiver, keyVal2})
	if err != nil {
		t.Fatalf("get with second key failed: %v", err)
	}
	strVal, ok := retrieved.(runtime.StringValue)
	if !ok || strVal.Val != "value" {
		t.Fatalf("get returned unexpected value %#v", retrieved)
	}
}
