package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestCallableAllowsTransientCallEnvReuseMatchesBodyShape(t *testing.T) {
	simpleFn := ast.Fn(
		"simple",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("i32"))},
		[]ast.Statement{ast.Ret(ast.ID("x"))},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	if !callableAllowsTransientCallEnvReuse(simpleFn) {
		t.Fatalf("expected simple function body to allow transient call env reuse")
	}

	escapingFn := ast.Fn(
		"escaping",
		nil,
		[]ast.Statement{ast.Ret(ast.Lam(nil, ast.Int(1)))},
		nil,
		nil,
		nil,
		false,
		false,
	)
	if callableAllowsTransientCallEnvReuse(escapingFn) {
		t.Fatalf("expected closure-producing function body to reject transient call env reuse")
	}

	simpleLambda := ast.Lam([]*ast.FunctionParameter{ast.Param("x", ast.Ty("i32"))}, ast.ID("x"))
	if !callableAllowsTransientCallEnvReuse(simpleLambda) {
		t.Fatalf("expected simple lambda body to allow transient call env reuse")
	}

	escapingLambda := ast.Lam(nil, ast.IteratorLit(ast.Yield(ast.Int(1))))
	if callableAllowsTransientCallEnvReuse(escapingLambda) {
		t.Fatalf("expected iterator-producing lambda body to reject transient call env reuse")
	}
}

func TestInvokeFunctionTransientCallEnvResetsMutatedExplicitBindingsBytecode(t *testing.T) {
	interp := NewBytecode()
	closure := runtime.NewEnvironment(nil)
	decl := ast.Fn(
		"probe",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
		[]ast.Statement{
			ast.Assign(ast.ID("before"), ast.ID("T_type")),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("T_type"), ast.Str("dirty")),
			ast.Ret(ast.ID("before")),
		},
		ast.Ty("String"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	if !callableAllowsTransientCallEnvReuse(decl) {
		t.Fatalf("expected mutating non-escaping function body to allow transient call env reuse")
	}
	program, err := interp.lowerFunctionDefinitionBytecode(decl)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	fn := &runtime.FunctionValue{Declaration: decl, Closure: closure}
	setFunctionBytecodeProgram(fn, program)
	call := ast.CallT(ast.ID("probe"), []ast.TypeExpression{ast.Ty("i32")}, ast.Int(1))
	args := []runtime.Value{runtime.NewSmallInt(1, runtime.IntegerI32)}

	first, err := interp.callResolvedFunctionValue(fn, fn, args, closure, call, false)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	second, err := interp.callResolvedFunctionValue(fn, fn, args, closure, call, false)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	for idx, got := range []runtime.Value{first, second} {
		if str, ok := got.(runtime.StringValue); !ok || str.Val != "i32" {
			t.Fatalf("call %d result = %#v, want StringValue{i32}", idx+1, got)
		}
	}
	if got := len(interp.reusableBytecodeCallEnvCache); got != 0 {
		t.Fatalf("expected mutating runtime-binding function to avoid immutable reusable env cache, got %d entries", got)
	}
}

func TestInvokeFunctionTransientCallEnvHotPathAllocatesLessThanFreshEnvPath(t *testing.T) {
	buildFunction := func(body []ast.Statement) (*Interpreter, *runtime.FunctionValue, *runtime.Environment, *ast.FunctionCall, []runtime.Value) {
		interp := NewBytecode()
		closure := runtime.NewEnvironment(nil)
		decl := ast.Fn(
			"probe",
			[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
			body,
			nil,
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
			false,
		)
		fn := &runtime.FunctionValue{
			Declaration: decl,
			Closure:     closure,
			Bytecode: CompiledThunk(func(localEnv *runtime.Environment, args []runtime.Value) (runtime.Value, error) {
				if got, ok := localEnv.Lookup("T_type"); !ok {
					t.Fatalf("expected T_type binding in transient-call-env thunk")
				} else if str, ok := got.(runtime.StringValue); !ok || str.Val != "i32" {
					t.Fatalf("T_type = %#v, want i32", got)
				}
				return runtime.StringValue{Val: "ok"}, nil
			}),
		}
		call := ast.CallT(ast.ID("probe"), []ast.TypeExpression{ast.Ty("i32")}, ast.Int(1))
		args := []runtime.Value{runtime.NewSmallInt(1, runtime.IntegerI32)}
		return interp, fn, closure, call, args
	}

	eligibleInterp, eligibleFn, eligibleClosure, eligibleCall, eligibleArgs := buildFunction(
		[]ast.Statement{ast.ID("T_type")},
	)
	if !callableAllowsTransientCallEnvReuse(eligibleFn.Declaration) {
		t.Fatalf("expected eligible function body to allow transient call env reuse")
	}
	ineligibleInterp, ineligibleFn, ineligibleClosure, ineligibleCall, ineligibleArgs := buildFunction(
		[]ast.Statement{ast.Ret(ast.Lam(nil, ast.ID("T_type")))},
	)
	if callableAllowsTransientCallEnvReuse(ineligibleFn.Declaration) {
		t.Fatalf("expected closure-producing function body to reject transient call env reuse")
	}

	for _, tc := range []struct {
		fn      *runtime.FunctionValue
		interp  *Interpreter
		closure *runtime.Environment
		call    *ast.FunctionCall
		args    []runtime.Value
	}{
		{fn: eligibleFn, interp: eligibleInterp, closure: eligibleClosure, call: eligibleCall, args: eligibleArgs},
		{fn: ineligibleFn, interp: ineligibleInterp, closure: ineligibleClosure, call: ineligibleCall, args: ineligibleArgs},
	} {
		for iter := 0; iter < 2; iter++ {
			got, err := tc.interp.callResolvedFunctionValue(tc.fn, tc.fn, tc.args, tc.closure, tc.call, false)
			if err != nil {
				t.Fatalf("warmup call failed: %v", err)
			}
			if str, ok := got.(runtime.StringValue); !ok || str.Val != "ok" {
				t.Fatalf("warmup result = %#v, want ok", got)
			}
		}
	}

	eligibleAllocs := testing.AllocsPerRun(1000, func() {
		got, err := eligibleInterp.callResolvedFunctionValue(eligibleFn, eligibleFn, eligibleArgs, eligibleClosure, eligibleCall, false)
		if err != nil {
			panic(err)
		}
		if str, ok := got.(runtime.StringValue); !ok || str.Val != "ok" {
			panic("unexpected eligible transient-call-env result")
		}
	})
	ineligibleAllocs := testing.AllocsPerRun(1000, func() {
		got, err := ineligibleInterp.callResolvedFunctionValue(ineligibleFn, ineligibleFn, ineligibleArgs, ineligibleClosure, ineligibleCall, false)
		if err != nil {
			panic(err)
		}
		if str, ok := got.(runtime.StringValue); !ok || str.Val != "ok" {
			panic("unexpected ineligible transient-call-env result")
		}
	})
	if !(eligibleAllocs < ineligibleAllocs) {
		t.Fatalf("expected eligible transient call env path to allocate less than fresh env path, got eligible=%.2f ineligible=%.2f", eligibleAllocs, ineligibleAllocs)
	}
}
