package bridge

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
)

// Raise panics with the provided value so compiled code can signal a runtime error.
func Raise(value runtime.Value) {
	panic(value)
}

// RaisedError converts a raised runtime value into an interpreter-compatible error.
func RaisedError(rt *Runtime, node ast.Node, value runtime.Value) error {
	if value == nil {
		value = runtime.NilValue{}
	}
	if rt == nil || rt.interp == nil {
		if errVal, ok := value.(runtime.ErrorValue); ok {
			return fmt.Errorf("%s", errVal.Message)
		}
		return fmt.Errorf("%v", value)
	}
	env := rt.currentEnv()
	err := interpreter.Raise(rt.interp, value, env)
	if node != nil {
		err = attachRuntimeContext(rt, err, node, env)
	}
	return err
}

// RaisedErrorIn converts a raised runtime value into an interpreter-compatible
// error using the native-call context environment when available.
func RaisedErrorIn(rt *Runtime, ctx *runtime.NativeCallContext, value runtime.Value) error {
	if value == nil {
		value = runtime.NilValue{}
	}
	var env *runtime.Environment
	if ctx != nil {
		env = ctx.Env
	}
	if env == nil && rt != nil {
		env = rt.currentEnv()
	}
	if rt == nil || rt.interp == nil {
		if errVal, ok := value.(runtime.ErrorValue); ok {
			return fmt.Errorf("%s", errVal.Message)
		}
		return fmt.Errorf("%v", value)
	}
	return interpreter.Raise(rt.interp, value, env)
}

// RaiseWithContext raises a value with attached runtime diagnostics.
func RaiseWithContext(rt *Runtime, node ast.Node, value runtime.Value) {
	panic(RaisedError(rt, node, value))
}

// RaiseRuntimeErrorWithContext attaches runtime diagnostics to an error and panics.
func RaiseRuntimeErrorWithContext(rt *Runtime, node ast.Node, err error) {
	panic(RuntimeErrorWithContext(rt, node, err))
}

// RuntimeErrorWithContext attaches runtime diagnostics to an error without panicking.
func RuntimeErrorWithContext(rt *Runtime, node ast.Node, err error) error {
	if err == nil {
		return nil
	}
	if rt == nil || rt.interp == nil {
		return err
	}
	env := rt.currentEnv()
	return attachRuntimeContext(rt, err, node, env)
}

// RegisterNodeOrigin wires a node origin path for compiled diagnostics.
func RegisterNodeOrigin(rt *Runtime, node ast.Node, origin string) {
	if rt == nil || rt.interp == nil || node == nil || origin == "" {
		return
	}
	rt.interp.AddNodeOrigin(node, origin)
}

// PushCallFrame records a compiled call expression for later diagnostics.
func PushCallFrame(rt *Runtime, call *ast.FunctionCall) {
	if rt == nil || rt.interp == nil || call == nil {
		return
	}
	rt.pushBridgeCallFrame(call)
}

// PopCallFrame removes the most recent compiled call expression frame.
func PopCallFrame(rt *Runtime) {
	if rt == nil || rt.interp == nil {
		return
	}
	rt.popBridgeCallFrame()
}

// AppendCallFrameError appends a compiled caller frame onto an existing
// runtime-diagnostic error on the slow error path.
func AppendCallFrameError(rt *Runtime, err error, call *ast.FunctionCall) error {
	if err == nil || rt == nil || rt.interp == nil || call == nil {
		return err
	}
	return rt.interp.AppendRuntimeCallFrame(err, call)
}

func attachRuntimeContext(rt *Runtime, err error, node ast.Node, env *runtime.Environment) error {
	if err == nil || rt == nil || rt.interp == nil {
		return err
	}
	return rt.interp.AttachRuntimeContextWithCallStack(err, node, env, rt.snapshotBridgeCallFrames())
}

// Recover converts a recovered panic into a runtime error compatible with the interpreter.
func Recover(rt *Runtime, ctx *runtime.NativeCallContext, recovered any) error {
	if recovered == nil {
		return nil
	}
	if err, ok := recovered.(error); ok {
		return err
	}
	if val, ok := recovered.(runtime.Value); ok {
		return RaisedErrorIn(rt, ctx, val)
	}
	return fmt.Errorf("panic: %v", recovered)
}
