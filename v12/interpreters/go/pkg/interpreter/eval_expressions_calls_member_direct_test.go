package interpreter

import (
	"fmt"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestEvaluateFunctionCall_DirectInterfaceMemberCallSkipsBoundWrapper(t *testing.T) {
	interp := New()
	env := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 1)
	receiver := runtime.StringValue{Val: "beta"}
	iface := &runtime.InterfaceValue{
		Interface: &runtime.InterfaceDefinitionValue{
			Node: ast.Iface(
				"Probe",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"probe",
						[]*ast.FunctionParameter{ast.Param("delta", ast.Ty("i32"))},
						ast.Ty("i32"),
						nil,
						nil,
						nil,
					),
				},
				nil,
				nil,
				nil,
				nil,
				false,
			),
		},
		Underlying: receiver,
		Methods: map[string]runtime.Value{
			"probe": runtime.NativeFunctionValue{
				Name:       "probe",
				Arity:      1,
				BorrowArgs: true,
				Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
					if len(args) != 2 {
						return nil, fmt.Errorf("probe args = %d, want 2", len(args))
					}
					if !valuesEqual(args[0], receiver) {
						return nil, fmt.Errorf("receiver = %#v, want %#v", args[0], receiver)
					}
					delta, ok := args[1].(runtime.IntegerValue)
					if !ok {
						return nil, fmt.Errorf("delta type = %T", args[1])
					}
					value, ok := delta.ToInt64()
					if !ok {
						return nil, fmt.Errorf("delta not small int: %#v", delta)
					}
					return runtime.NewSmallInt(value+1, runtime.IntegerI32), nil
				},
			},
		},
	}
	env.Define("iface", iface)

	result, err := interp.evaluateFunctionCall(
		ast.CallExpr(ast.Member(ast.ID("iface"), "probe"), ast.Int(5)),
		env,
	)
	if err != nil {
		t.Fatalf("direct interface member call failed: %v", err)
	}
	if !valuesEqual(result, runtime.NewSmallInt(6, runtime.IntegerI32)) {
		t.Fatalf("result = %#v, want 6", result)
	}
	if iface.BoundMethod != nil || iface.BoundMethodName != "" {
		t.Fatalf("direct interface member call should not materialize bound cache, got %q %#v", iface.BoundMethodName, iface.BoundMethod)
	}
}
