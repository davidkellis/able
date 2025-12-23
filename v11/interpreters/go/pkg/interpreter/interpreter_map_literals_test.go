package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestEvaluateMapLiteral(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	seedHashMapStruct(t, interp, env)

	literal := ast.MapLit([]ast.MapLiteralElement{
		ast.MapEntry(ast.NewStringLiteral("host"), ast.NewIntegerLiteral(big.NewInt(1), nil)),
		ast.MapEntry(ast.NewStringLiteral("port"), ast.NewIntegerLiteral(big.NewInt(443), nil)),
	})

	value, err := interp.evaluateExpression(literal, env)
	if err != nil {
		t.Fatalf("evaluate map literal failed: %v", err)
	}
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected hash map struct instance, got %T", value)
	}
	if inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil || inst.Definition.Node.ID.Name != "HashMap" {
		t.Fatalf("expected HashMap instance, got %#v", inst.Definition)
	}
	handle, err := interp.hashMapHandleFromInstance(inst)
	if err != nil {
		t.Fatalf("missing hash map handle: %v", err)
	}
	state, err := interp.hashMapStateForHandle(handle)
	if err != nil {
		t.Fatalf("missing hash map state: %v", err)
	}
	if len(state.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(state.Entries))
	}
	if entry := state.Entries[0]; entry.Key.(runtime.StringValue).Val != "host" {
		t.Fatalf("unexpected first key %#v", entry.Key)
	}
	if len(inst.TypeArguments) != 2 {
		t.Fatalf("expected HashMap type arguments, got %#v", inst.TypeArguments)
	}
	if simple, ok := inst.TypeArguments[0].(*ast.SimpleTypeExpression); !ok || simple.Name.Name != "String" {
		t.Fatalf("expected key type String, got %#v", inst.TypeArguments[0])
	}
	if simple, ok := inst.TypeArguments[1].(*ast.SimpleTypeExpression); !ok || simple.Name.Name != "i32" {
		t.Fatalf("expected value type i32, got %#v", inst.TypeArguments[1])
	}
}

func TestEvaluateMapLiteralWithSpread(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	seedHashMapStruct(t, interp, env)

	defaultsLiteral := ast.MapLit([]ast.MapLiteralElement{
		ast.MapEntry(ast.NewStringLiteral("accept"), ast.NewStringLiteral("application/json")),
		ast.MapEntry(ast.NewStringLiteral("cache"), ast.NewStringLiteral("no-store")),
	})
	defaultsValue, err := interp.evaluateExpression(defaultsLiteral, env)
	if err != nil {
		t.Fatalf("evaluate defaults failed: %v", err)
	}
	env.Define("defaults", defaultsValue)

	literal := ast.MapLit([]ast.MapLiteralElement{
		ast.MapEntry(ast.NewStringLiteral("content-type"), ast.NewStringLiteral("application/json")),
		ast.MapSpread(ast.NewIdentifier("defaults")),
		ast.MapEntry(ast.NewStringLiteral("cache"), ast.NewStringLiteral("max-age=0")),
	})

	value, err := interp.evaluateExpression(literal, env)
	if err != nil {
		t.Fatalf("evaluate map literal with spread failed: %v", err)
	}
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected hash map struct instance, got %T", value)
	}
	handle, err := interp.hashMapHandleFromInstance(inst)
	if err != nil {
		t.Fatalf("missing hash map handle: %v", err)
	}
	state, err := interp.hashMapStateForHandle(handle)
	if err != nil {
		t.Fatalf("missing hash map state: %v", err)
	}
	if len(state.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(state.Entries))
	}
	cacheVal := mapStateValue(t, interp, state, runtime.StringValue{Val: "cache"})
	if str, ok := cacheVal.(runtime.StringValue); !ok || str.Val != "max-age=0" {
		t.Fatalf("expected cache override, got %#v", cacheVal)
	}
}

func seedHashMapStruct(t *testing.T, interp *Interpreter, env *runtime.Environment) {
	t.Helper()
	keyParam := ast.GenericParam("K", nil)
	valParam := ast.GenericParam("V", nil)
	def := ast.StructDef(
		"HashMap",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i64"), "handle"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{keyParam, valParam},
		nil,
		false,
	)
	if _, err := interp.evaluateStructDefinition(def, env); err != nil {
		t.Fatalf("define HashMap struct failed: %v", err)
	}
}

func mapStateValue(t *testing.T, interp *Interpreter, state *runtime.HashMapValue, key runtime.Value) runtime.Value {
	t.Helper()
	hash, err := interp.hashMapHashValue(key)
	if err != nil {
		t.Fatalf("hash key failed: %v", err)
	}
	idx, found, err := interp.hashMapFindEntryWithHash(state, hash, key)
	if err != nil {
		t.Fatalf("find key failed: %v", err)
	}
	if !found {
		t.Fatalf("expected key not found in map")
	}
	return state.Entries[idx].Value
}
