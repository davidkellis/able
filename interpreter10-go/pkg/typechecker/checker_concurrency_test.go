package typechecker

import (
	"able/interpreter10-go/pkg/ast"
	"strings"
	"testing"
)

func TestProcExpressionReturnsProcType(t *testing.T) {
	checker := New()
	expr := ast.Proc(ast.Int(42))
	module := ast.NewModule([]ast.Statement{expr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	procType, ok := checker.infer[expr]
	if !ok {
		t.Fatalf("expected proc inference entry")
	}
	pt, ok := procType.(ProcType)
	if !ok {
		t.Fatalf("expected ProcType, got %#v", procType)
	}
	if pt.Result == nil || typeName(pt.Result) != "i32" {
		t.Fatalf("expected proc result i32, got %#v", pt.Result)
	}
}
func TestProcHandleMethodsHaveExpectedTypes(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.ID("handle"),
		ast.Proc(ast.Int(1)),
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
	if !ok || typeName(fn.Return) != "ProcStatus" {
		t.Fatalf("expected status to return ProcStatus function, got %#v", statusType)
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
	if union.Members[1] == nil || typeName(union.Members[1]) != "ProcError" {
		t.Fatalf("expected union second member ProcError, got %#v", union.Members[1])
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
func TestFutureCancelProducesDiagnostic(t *testing.T) {
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
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for future cancel()")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "future handles do not support cancel") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected future cancel diagnostic, got %v", diags)
	}
}
func TestProcCancelledRequiresAsyncContext(t *testing.T) {
	checker := New()
	call := ast.Call("proc_cancelled")
	module := ast.NewModule([]ast.Statement{call}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for proc_cancelled outside async context")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "proc_cancelled must be called inside an asynchronous task") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected proc_cancelled context diagnostic, got %v", diags)
	}
}
func TestProcCancelledAllowedInsideProc(t *testing.T) {
	checker := New()
	procExpr := ast.Proc(ast.Call("proc_cancelled"))
	module := ast.NewModule([]ast.Statement{procExpr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for proc_cancelled inside proc, got %v", diags)
	}
	typ, ok := checker.infer[procExpr]
	if !ok {
		t.Fatalf("expected inference entry for proc expression")
	}
	procType, ok := typ.(ProcType)
	if !ok {
		t.Fatalf("expected ProcType, got %#v", typ)
	}
	if procType.Result == nil || typeName(procType.Result) != "bool" {
		t.Fatalf("expected proc result bool, got %#v", procType.Result)
	}
}
func TestProcYieldRequiresAsyncContext(t *testing.T) {
	checker := New()
	call := ast.Call("proc_yield")
	module := ast.NewModule([]ast.Statement{call}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for proc_yield outside async context")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "proc_yield() may only be called") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected proc_yield context diagnostic, got %v", diags)
	}
}
func TestProcYieldAllowedInsideProc(t *testing.T) {
	checker := New()
	procExpr := ast.Proc(ast.Call("proc_yield"))
	module := ast.NewModule([]ast.Statement{procExpr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for proc_yield inside proc, got %v", diags)
	}
	typ, ok := checker.infer[procExpr]
	if !ok {
		t.Fatalf("expected inference entry for proc expression")
	}
	procType, ok := typ.(ProcType)
	if !ok {
		t.Fatalf("expected ProcType, got %#v", typ)
	}
	if procType.Result == nil || typeName(procType.Result) != "nil" {
		t.Fatalf("expected proc result nil, got %#v", procType.Result)
	}
}
func TestProcFlushReturnsNil(t *testing.T) {
	checker := New()
	call := ast.Call("proc_flush")
	module := ast.NewModule([]ast.Statement{call}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for proc_flush call, got %v", diags)
	}
	typ, ok := checker.infer[call]
	if !ok {
		t.Fatalf("expected inference entry for proc_flush call")
	}
	if typ == nil || typeName(typ) != "nil" {
		t.Fatalf("expected proc_flush to return nil, got %#v", typ)
	}
}
