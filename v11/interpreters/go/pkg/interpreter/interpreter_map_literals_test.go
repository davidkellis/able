package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestEvaluateMapLiteral(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	literal := ast.MapLit([]ast.MapLiteralElement{
		ast.MapEntry(ast.NewStringLiteral("host"), ast.NewStringLiteral("api")),
		ast.MapEntry(ast.NewStringLiteral("port"), ast.NewIntegerLiteral(big.NewInt(443), nil)),
	})

	value, err := interp.evaluateExpression(literal, env)
	if err != nil {
		t.Fatalf("evaluate map literal failed: %v", err)
	}
	hm, ok := value.(*runtime.HashMapValue)
	if !ok {
		t.Fatalf("expected hash map value, got %T", value)
	}
	if len(hm.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(hm.Entries))
	}
	if entry := hm.Entries[0]; entry.Key.(runtime.StringValue).Val != "host" {
		t.Fatalf("unexpected first key %#v", entry.Key)
	}
}

func TestEvaluateMapLiteralWithSpread(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

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
	hm, ok := value.(*runtime.HashMapValue)
	if !ok {
		t.Fatalf("expected hash map value, got %T", value)
	}
	if len(hm.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(hm.Entries))
	}
}
