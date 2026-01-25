package typechecker

import (
	"able/interpreter-go/pkg/ast"
	"strings"
	"testing"
)

func TestSpawnExpressionReturnsFutureType(t *testing.T) {
	checker := New()
	expr := ast.Spawn(ast.Int(42))
	module := ast.NewModule([]ast.Statement{expr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	futureType, ok := checker.infer[expr]
	if !ok {
		t.Fatalf("expected future inference entry")
	}
	pt, ok := futureType.(FutureType)
	if !ok {
		t.Fatalf("expected FutureType, got %#v", futureType)
	}
	if pt.Result == nil || typeName(pt.Result) != "i32" {
		t.Fatalf("expected future result i32, got %#v", pt.Result)
	}
}
func TestSpawnHandleMethodsHaveExpectedTypes(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.ID("handle"),
		ast.Spawn(ast.Int(1)),
	)
	statusMember := ast.Member(ast.ID("handle"), "status")
	valueMember := ast.Member(ast.ID("handle"), "value")
	cancelMember := ast.Member(ast.ID("handle"), "cancel")
	module := ast.NewModule([]ast.Statement{
		assign,
		statusMember,
		valueMember,
		cancelMember,
	}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	statusType, ok := checker.infer[statusMember]
	if !ok {
		t.Fatalf("expected status member inference")
	}
	fn, ok := statusType.(FunctionType)
	if !ok || typeName(fn.Return) != "FutureStatus" {
		t.Fatalf("expected status to return FutureStatus function, got %#v", statusType)
	}
	valueType, ok := checker.infer[valueMember]
	if !ok {
		t.Fatalf("expected value member inference")
	}
	valueFn, ok := valueType.(FunctionType)
	if !ok {
		t.Fatalf("expected function type for value, got %#v", valueType)
	}
	union, ok := valueFn.Return.(UnionLiteralType)
	if !ok || len(union.Members) != 2 {
		t.Fatalf("expected value() to return union, got %#v", valueFn.Return)
	}
	if union.Members[0] == nil || typeName(union.Members[0]) != "i32" {
		t.Fatalf("expected union first member i32, got %#v", union.Members[0])
	}
	if union.Members[1] == nil || typeName(union.Members[1]) != "FutureError" {
		t.Fatalf("expected union second member FutureError, got %#v", union.Members[1])
	}
	cancelType, ok := checker.infer[cancelMember]
	if !ok {
		t.Fatalf("expected cancel member inference")
	}
	cancelFn, ok := cancelType.(FunctionType)
	if !ok || typeName(cancelFn.Return) != "nil" {
		t.Fatalf("expected cancel() to return nil, got %#v", cancelType)
	}
}
func TestFutureCancelAllowed(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.ID("future"),
		ast.Spawn(ast.Int(1)),
	)
	cancelMember := ast.Member(ast.ID("future"), "cancel")
	module := ast.NewModule([]ast.Statement{
		assign,
		cancelMember,
	}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for future cancel(), got %v", diags)
	}
}
func TestFutureCancelledRequiresAsyncContext(t *testing.T) {
	checker := New()
	call := ast.Call("future_cancelled")
	module := ast.NewModule([]ast.Statement{call}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for future_cancelled outside async context, got %v", diags)
	}
}
func TestFutureCancelledAllowedInsideSpawn(t *testing.T) {
	checker := New()
	spawnExpr := ast.Spawn(ast.Call("future_cancelled"))
	module := ast.NewModule([]ast.Statement{spawnExpr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for future_cancelled inside spawn, got %v", diags)
	}
	typ, ok := checker.infer[spawnExpr]
	if !ok {
		t.Fatalf("expected inference entry for spawn expression")
	}
	futureType, ok := typ.(FutureType)
	if !ok {
		t.Fatalf("expected FutureType, got %#v", typ)
	}
	if futureType.Result == nil || typeName(futureType.Result) != "bool" {
		t.Fatalf("expected future result bool, got %#v", futureType.Result)
	}
}
func TestFutureYieldRequiresAsyncContext(t *testing.T) {
	checker := New()
	call := ast.Call("future_yield")
	module := ast.NewModule([]ast.Statement{call}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for future_yield outside async context")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "future_yield() may only be called") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected future_yield context diagnostic, got %v", diags)
	}
}
func TestFutureYieldAllowedInsideSpawn(t *testing.T) {
	checker := New()
	spawnExpr := ast.Spawn(ast.Call("future_yield"))
	module := ast.NewModule([]ast.Statement{spawnExpr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for future_yield inside spawn, got %v", diags)
	}
	typ, ok := checker.infer[spawnExpr]
	if !ok {
		t.Fatalf("expected inference entry for spawn expression")
	}
	futureType, ok := typ.(FutureType)
	if !ok {
		t.Fatalf("expected FutureType, got %#v", typ)
	}
	if futureType.Result == nil || typeName(futureType.Result) != "nil" {
		t.Fatalf("expected future result nil, got %#v", futureType.Result)
	}
}
func TestFutureFlushReturnsNil(t *testing.T) {
	checker := New()
	call := ast.Call("future_flush")
	module := ast.NewModule([]ast.Statement{call}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for future_flush call, got %v", diags)
	}
	typ, ok := checker.infer[call]
	if !ok {
		t.Fatalf("expected inference entry for future_flush call")
	}
	if typ == nil || typeName(typ) != "nil" {
		t.Fatalf("expected future_flush to return nil, got %#v", typ)
	}
}
