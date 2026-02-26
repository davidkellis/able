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

// RaiseWithContext raises a value with attached runtime diagnostics.
func RaiseWithContext(rt *Runtime, node ast.Node, value runtime.Value) {
	if rt == nil || rt.interp == nil {
		panic(value)
	}
	env := rt.currentEnv()
	err := interpreter.Raise(rt.interp, value, env)
	if node != nil {
		err = rt.interp.AttachRuntimeContext(err, node, env)
	}
	panic(err)
}

// RaiseRuntimeErrorWithContext attaches runtime diagnostics to an error and panics.
func RaiseRuntimeErrorWithContext(rt *Runtime, node ast.Node, err error) {
	if err == nil {
		return
	}
	if rt == nil || rt.interp == nil {
		panic(err)
	}
	env := rt.currentEnv()
	panic(rt.interp.AttachRuntimeContext(err, node, env))
}

// RegisterNodeOrigin wires a node origin path for compiled diagnostics.
func RegisterNodeOrigin(rt *Runtime, node ast.Node, origin string) {
	if rt == nil || rt.interp == nil || node == nil || origin == "" {
		return
	}
	rt.interp.AddNodeOrigin(node, origin)
}

// PushCallFrame records a call expression in the interpreter's runtime state.
func PushCallFrame(rt *Runtime, call *ast.FunctionCall) {
	if rt == nil || rt.interp == nil || call == nil {
		return
	}
	env := rt.currentEnv()
	rt.interp.PushCallFrame(env, call)
}

// PopCallFrame removes the most recent call expression frame.
func PopCallFrame(rt *Runtime) {
	if rt == nil || rt.interp == nil {
		return
	}
	env := rt.currentEnv()
	rt.interp.PopCallFrame(env)
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
		var env *runtime.Environment
		if ctx != nil {
			env = ctx.Env
		}
		if rt == nil || rt.interp == nil {
			return fmt.Errorf("panic: %v", val)
		}
		return interpreter.Raise(rt.interp, val, env)
	}
	return fmt.Errorf("panic: %v", recovered)
}
