package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_EvalExpressionReusesLoweredMatchPrograms(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	matchExpr := ast.Match(
		ast.Int(3),
		ast.Mc(ast.ID("x"), ast.Int(1), ast.Bin("<", ast.ID("x"), ast.Int(3))),
		ast.Mc(ast.ID("x"), ast.Bin("+", ast.ID("x"), ast.Int(2)), ast.Bin("==", ast.ID("x"), ast.Int(3))),
		ast.Mc(ast.Wc(), ast.Int(0)),
	)
	vm := interp.acquireBytecodeVM(interp.GlobalEnvironment())
	defer interp.releaseBytecodeVM(vm)

	interp.ResetBytecodeStats()

	first, err := vm.evalExpressionBytecode(matchExpr, interp.GlobalEnvironment())
	if err != nil {
		t.Fatalf("first evalExpressionBytecode: %v", err)
	}
	want := runtime.NewSmallInt(5, runtime.IntegerI32)
	if !valuesEqual(first, want) {
		t.Fatalf("first evalExpressionBytecode mismatch: got=%#v want=%#v", first, want)
	}
	afterFirst := interp.BytecodeStats()
	if afterFirst.ExprCacheMisses < 3 {
		t.Fatalf("expected initial match evaluation to lower subject/guard/body programs, got misses=%d", afterFirst.ExprCacheMisses)
	}
	if afterFirst.ExprCacheHits != 0 {
		t.Fatalf("expected no expression cache hits on first evaluation, got hits=%d", afterFirst.ExprCacheHits)
	}

	second, err := vm.evalExpressionBytecode(matchExpr, interp.GlobalEnvironment())
	if err != nil {
		t.Fatalf("second evalExpressionBytecode: %v", err)
	}
	if !valuesEqual(second, want) {
		t.Fatalf("second evalExpressionBytecode mismatch: got=%#v want=%#v", second, want)
	}
	afterSecond := interp.BytecodeStats()
	if afterSecond.ExprCacheMisses != afterFirst.ExprCacheMisses {
		t.Fatalf("expected second evaluation to reuse cached lowered programs, got misses=%d want=%d", afterSecond.ExprCacheMisses, afterFirst.ExprCacheMisses)
	}
	if afterSecond.ExprCacheHits < 3 {
		t.Fatalf("expected cached match/guard/body programs on second evaluation, got hits=%d", afterSecond.ExprCacheHits)
	}
}
