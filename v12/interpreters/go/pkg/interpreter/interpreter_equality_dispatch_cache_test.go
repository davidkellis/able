package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestApplyEqualityInterface_EqualityDispatchCacheInvalidationAndResolver(t *testing.T) {
	interp := New()
	def := &runtime.StructDefinitionValue{Node: ast.StructDef("EqualityCacheProbe", nil, ast.StructKindNamed, nil, nil, false)}

	result, handled, err := interp.applyEqualityInterface("==", def, def)
	if err != nil {
		t.Fatalf("initial equality dispatch failed: %v", err)
	}
	if handled || result != nil {
		t.Fatalf("expected no equality method before resolver, got handled=%v result=%#v", handled, result)
	}
	if got := len(interp.equalityDispatchCache); got == 0 {
		t.Fatalf("expected no-method equality dispatch cache entry")
	}

	interp.invalidateMethodCache()
	if got := len(interp.equalityDispatchCache); got != 0 {
		t.Fatalf("expected method cache invalidation to clear equality dispatch cache, got %d entries", got)
	}

	_, handled, err = interp.applyEqualityInterface("==", def, def)
	if err != nil {
		t.Fatalf("second equality dispatch failed: %v", err)
	}
	if handled {
		t.Fatalf("expected no equality method after cache refill setup")
	}
	if got := len(interp.equalityDispatchCache); got == 0 {
		t.Fatalf("expected equality dispatch cache refill before resolver install")
	}

	resolverCalls := 0
	interp.SetInterfaceMethodResolver(func(receiver runtime.Value, interfaceName string, methodName string) (runtime.Value, bool) {
		resolverCalls++
		if interfaceName != "Eq" || methodName != "eq" {
			return nil, false
		}
		return runtime.NativeFunctionValue{
			Name:        "eq",
			Arity:       2,
			BorrowArgs:  true,
			SkipContext: true,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					t.Fatalf("resolver eq args = %d, want 2", len(args))
				}
				return runtime.BoolValue{Val: true}, nil
			},
		}, true
	})
	if got := len(interp.equalityDispatchCache); got != 0 {
		t.Fatalf("expected resolver install to clear equality dispatch cache, got %d entries", got)
	}

	result, handled, err = interp.applyEqualityInterface("==", def, def)
	if err != nil {
		t.Fatalf("resolver equality dispatch failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected resolver-provided equality method to handle comparison")
	}
	boolResult, ok := result.(runtime.BoolValue)
	if !ok || !boolResult.Val {
		t.Fatalf("expected resolver equality to return true bool, got %T (%#v)", result, result)
	}
	if resolverCalls == 0 {
		t.Fatalf("expected dynamic resolver to be called")
	}
	if got := len(interp.equalityDispatchCache); got != 0 {
		t.Fatalf("expected equality dispatch cache to stay disabled while resolver is installed, got %d entries", got)
	}
}

func TestApplyCachedEqualityDispatch_OwnedArgsAllowParameterCoercion(t *testing.T) {
	for _, test := range []struct {
		name   string
		interp *Interpreter
	}{
		{name: "treewalker", interp: New()},
		{name: "bytecode", interp: NewBytecode()},
	} {
		t.Run(test.name, func(t *testing.T) {
			module := ast.Mod([]ast.Statement{
				ast.Fn(
					"coerced_eq",
					[]*ast.FunctionParameter{
						ast.Param("left", ast.Ty("f64")),
						ast.Param("right", ast.Ty("f64")),
					},
					[]ast.Statement{ast.Bool(true)},
					ast.Ty("bool"),
					nil,
					nil,
					false,
					false,
				),
			}, nil, nil)
			if _, _, err := test.interp.EvaluateModule(module); err != nil {
				t.Fatalf("evaluate module: %v", err)
			}
			method, err := test.interp.GlobalEnvironment().Get("coerced_eq")
			if err != nil {
				t.Fatalf("lookup coerced_eq: %v", err)
			}

			result, handled, err := test.interp.applyCachedEqualityDispatch(
				"==",
				runtime.NewSmallInt(7, runtime.IntegerI32),
				runtime.NewSmallInt(7, runtime.IntegerI32),
				equalityDispatchCacheEntry{
					kind: equalityDispatchCacheMethod,
					dispatch: operatorDispatch{
						interfaceName: "Eq",
						methodName:    "eq",
					},
					method: method,
				},
			)
			if err != nil {
				t.Fatalf("cached equality dispatch: %v", err)
			}
			boolResult, ok := result.(runtime.BoolValue)
			if !handled || !ok || !boolResult.Val {
				t.Fatalf("cached equality result = %#v, handled=%v", result, handled)
			}
		})
	}
}

func TestCallCallableValue2Mutable_ReentrantBorrowedCallsKeepOuterArgsStable(t *testing.T) {
	interp := New()
	inner := runtime.NativeFunctionValue{
		Name:        "inner_eq",
		Arity:       2,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				t.Fatalf("inner args = %d, want 2", len(args))
			}
			return runtime.BoolValue{Val: true}, nil
		},
	}
	outer := runtime.NativeFunctionValue{
		Name:        "outer_eq",
		Arity:       2,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			first, second := args[0], args[1]
			if _, err := interp.callCallableValue2Mutable(inner, runtime.BoolValue{Val: false}, runtime.BoolValue{Val: false}, nil, nil); err != nil {
				return nil, err
			}
			if args[0] != first || args[1] != second {
				t.Fatalf("nested call changed outer args: %#v", args)
			}
			return runtime.BoolValue{Val: true}, nil
		},
	}

	result, err := interp.callCallableValue2Mutable(outer, runtime.BoolValue{Val: true}, runtime.BoolValue{Val: true}, nil, nil)
	if err != nil {
		t.Fatalf("outer call: %v", err)
	}
	if boolResult, ok := result.(runtime.BoolValue); !ok || !boolResult.Val {
		t.Fatalf("outer result = %#v", result)
	}
}

func TestCallCallableValue2Mutable_NonBorrowingNativeRetainsOwnedArgs(t *testing.T) {
	interp := New()
	var retained []runtime.Value
	native := runtime.NativeFunctionValue{
		Name:        "retaining",
		Arity:       2,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			retained = args
			return runtime.NilValue{}, nil
		},
	}
	first := runtime.StringValue{Val: "first"}
	second := runtime.StringValue{Val: "second"}
	if _, err := interp.callCallableValue2Mutable(native, first, second, nil, nil); err != nil {
		t.Fatalf("retaining call: %v", err)
	}
	if _, err := interp.callCallableValue2Mutable(
		runtime.NativeFunctionValue{
			Name:        "borrowed",
			Arity:       2,
			BorrowArgs:  true,
			SkipContext: true,
			Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
				return runtime.NilValue{}, nil
			},
		},
		runtime.NilValue{},
		runtime.NilValue{},
		nil,
		nil,
	); err != nil {
		t.Fatalf("borrowed call: %v", err)
	}
	if len(retained) != 2 || retained[0] != first || retained[1] != second {
		t.Fatalf("retained args changed: %#v", retained)
	}
}

func TestCallCallableValue2Mutable_PartialCopiesPooledArgs(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"three_args",
			[]*ast.FunctionParameter{
				ast.Param("first", ast.Ty("bool")),
				ast.Param("second", ast.Ty("bool")),
				ast.Param("third", ast.Ty("bool")),
			},
			[]ast.Statement{ast.ID("first")},
			ast.Ty("bool"),
			nil,
			nil,
			false,
			false,
		),
	}, nil, nil)
	if _, _, err := interp.EvaluateModule(module); err != nil {
		t.Fatalf("evaluate module: %v", err)
	}
	callable, err := interp.GlobalEnvironment().Get("three_args")
	if err != nil {
		t.Fatalf("lookup three_args: %v", err)
	}
	result, err := interp.callCallableValue2Mutable(
		callable,
		runtime.BoolValue{Val: true},
		runtime.BoolValue{Val: false},
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("partial call: %v", err)
	}
	partial, ok := result.(*runtime.PartialFunctionValue)
	if !ok {
		t.Fatalf("partial result = %T, want *runtime.PartialFunctionValue", result)
	}
	if _, err := interp.callCallableValue2Mutable(
		runtime.NativeFunctionValue{
			Name:        "borrowed",
			Arity:       2,
			BorrowArgs:  true,
			SkipContext: true,
			Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
				return runtime.NilValue{}, nil
			},
		},
		runtime.NilValue{},
		runtime.NilValue{},
		nil,
		nil,
	); err != nil {
		t.Fatalf("scratch overwrite call: %v", err)
	}
	if len(partial.BoundArgs) != 2 {
		t.Fatalf("partial bound args = %d, want 2", len(partial.BoundArgs))
	}
	first, firstOK := partial.BoundArgs[0].(runtime.BoolValue)
	second, secondOK := partial.BoundArgs[1].(runtime.BoolValue)
	if !firstOK || !secondOK || !first.Val || second.Val {
		t.Fatalf("partial bound args changed: %#v", partial.BoundArgs)
	}
}
