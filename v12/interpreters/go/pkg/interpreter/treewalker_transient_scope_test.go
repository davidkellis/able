package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestAcquireTransientRuntimeScopeEnvWithCapacityClearsBindingsBetweenUses(t *testing.T) {
	interp := New()
	base := runtime.NewEnvironment(nil)

	first := interp.acquireTransientRuntimeScopeEnvWithCapacity(base, 3)
	first.Define("x", runtime.NewSmallInt(7, runtime.IntegerI32))
	interp.releaseTransientRuntimeScopeEnv(first)

	second := interp.acquireTransientRuntimeScopeEnvWithCapacity(base, 3)
	defer interp.releaseTransientRuntimeScopeEnv(second)

	if second.Parent() != base {
		t.Fatalf("expected transient runtime scope env parent to reset to base")
	}
	if _, ok := second.Lookup("x"); ok {
		t.Fatalf("expected reused transient runtime scope env to clear prior bindings")
	}
}

func TestTreewalkerTransientBlockScopeKeepsEscapingLambdaCapture(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(
			ast.ID("captured"),
			ast.IfExpr(
				ast.Bool(true),
				ast.Block(
					ast.Assign(ast.ID("x"), ast.Int(7)),
					ast.Lam(nil, ast.ID("x")),
				),
			),
		),
		ast.IfExpr(
			ast.Bool(true),
			ast.Block(
				ast.Assign(ast.ID("y"), ast.Int(9)),
				ast.ID("y"),
			),
		),
		ast.CallExpr(ast.ID("captured")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.BigInt().Cmp(bigInt(7)) != 0 {
		t.Fatalf("expected captured lambda result 7, got %#v", result)
	}
}

func TestEvaluateBlockTransientScopeHotPathAllocatesLessThanFreshScopePath(t *testing.T) {
	interp := New()
	base := runtime.NewEnvironment(nil)

	eligible := ast.Block(
		ast.Assign(ast.ID("a"), ast.Int(1)),
		ast.Assign(ast.ID("b"), ast.Int(2)),
		ast.Assign(ast.ID("c"), ast.Int(3)),
		ast.Bin("+", ast.ID("a"), ast.ID("c")),
	)
	ineligible := ast.Block(
		ast.Iff(ast.Bool(false), ast.Lam(nil, ast.Int(0))),
		ast.Assign(ast.ID("a"), ast.Int(1)),
		ast.Assign(ast.ID("b"), ast.Int(2)),
		ast.Assign(ast.ID("c"), ast.Int(3)),
		ast.Bin("+", ast.ID("a"), ast.ID("c")),
	)

	if _, err := interp.evaluateBlock(eligible, base); err != nil {
		t.Fatalf("expected eligible block warmup to succeed: %v", err)
	}

	eligibleAllocs := testing.AllocsPerRun(1000, func() {
		result, err := interp.evaluateBlock(eligible, base)
		if err != nil {
			panic(err)
		}
		intVal, ok := result.(runtime.IntegerValue)
		if !ok || intVal.BigInt().Cmp(bigInt(4)) != 0 {
			panic("unexpected eligible block result")
		}
	})
	ineligibleAllocs := testing.AllocsPerRun(1000, func() {
		result, err := interp.evaluateBlock(ineligible, base)
		if err != nil {
			panic(err)
		}
		intVal, ok := result.(runtime.IntegerValue)
		if !ok || intVal.BigInt().Cmp(bigInt(4)) != 0 {
			panic("unexpected ineligible block result")
		}
	})
	if eligibleAllocs >= ineligibleAllocs {
		t.Fatalf("expected eligible transient treewalker block scope path to allocate less than fresh scope path, got eligible=%.2f ineligible=%.2f", eligibleAllocs, ineligibleAllocs)
	}
}
