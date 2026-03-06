package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_NativeCallArgsSliceStaysStable(t *testing.T) {
	interp := NewBytecode()

	var captured []runtime.Value
	interp.GlobalEnvironment().Define("capture_once", runtime.NativeFunctionValue{
		Name:  "capture_once",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if captured == nil {
				captured = args
			}
			return runtime.NilValue{}, nil
		},
	})

	module := ast.Mod([]ast.Statement{
		ast.Call("capture_once", ast.Int(41)),
		ast.Call("capture_once", ast.Int(42)),
		ast.Int(0),
	}, nil, nil)

	_ = runBytecodeModuleWithInterpreter(t, interp, module)

	if len(captured) != 1 {
		t.Fatalf("expected one captured argument, got %d", len(captured))
	}
	intVal, ok := captured[0].(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected captured arg to be integer, got %#v", captured[0])
	}
	if intVal.BigInt().Int64() != 41 {
		t.Fatalf("expected first captured arg to remain 41, got %d", intVal.BigInt().Int64())
	}
}
