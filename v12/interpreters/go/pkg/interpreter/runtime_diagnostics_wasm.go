//go:build js && wasm

package interpreter

import (
	"errors"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type runtimeCallFrame struct {
	node *ast.FunctionCall
}

type runtimeDiagnosticContext struct {
	node      ast.Node
	callStack []runtimeCallFrame
}

type runtimeDiagnosticError struct {
	err     error
	context *runtimeDiagnosticContext
}

func (e runtimeDiagnosticError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e runtimeDiagnosticError) Unwrap() error {
	return e.err
}

type RuntimeDiagnosticLocation struct {
	Path      string
	Line      int
	Column    int
	EndLine   int
	EndColumn int
}

type RuntimeDiagnosticNote struct {
	Message  string
	Location RuntimeDiagnosticLocation
}

type RuntimeDiagnostic struct {
	Severity string
	Message  string
	Location RuntimeDiagnosticLocation
	Notes    []RuntimeDiagnosticNote
}

func (i *Interpreter) BuildRuntimeDiagnostic(err error) RuntimeDiagnostic {
	return RuntimeDiagnostic{
		Severity: "error",
		Message:  runtimeMessageFromError(err),
		Location: RuntimeDiagnosticLocation{},
	}
}

// AttachRuntimeContext attaches diagnostic context to an error for compiled/native callers.
func (i *Interpreter) AttachRuntimeContext(err error, node ast.Node, env *runtime.Environment) error {
	if i == nil {
		return err
	}
	state := i.stateFromEnv(env)
	return i.attachRuntimeContext(err, node, state)
}

func DescribeRuntimeDiagnostic(diag RuntimeDiagnostic) string {
	message := strings.TrimSpace(diag.Message)
	if strings.HasPrefix(message, "runtime:") {
		message = strings.TrimSpace(strings.TrimPrefix(message, "runtime:"))
	}
	if message == "" {
		message = "unknown runtime error"
	}
	return fmt.Sprintf("runtime: %s", message)
}

func (i *Interpreter) attachRuntimeContext(err error, node ast.Node, state *evalState) error {
	if err == nil || node == nil {
		return err
	}
	switch err.(type) {
	case returnSignal, breakSignal, continueSignal, generatorStopSignal:
		return err
	}
	if runtimeContextFromError(err) != nil {
		return err
	}
	context := &runtimeDiagnosticContext{
		node:      node,
		callStack: state.snapshotCallStack(),
	}
	if sig, ok := err.(raiseSignal); ok {
		sig.context = context
		return sig
	}
	if diagErr, ok := err.(runtimeDiagnosticError); ok {
		diagErr.context = context
		return diagErr
	}
	return runtimeDiagnosticError{err: err, context: context}
}

func runtimeContextFromError(err error) *runtimeDiagnosticContext {
	if err == nil {
		return nil
	}
	var sig raiseSignal
	if errors.As(err, &sig) {
		if sig.context != nil {
			return sig.context
		}
	}
	var diagErr runtimeDiagnosticError
	if errors.As(err, &diagErr) {
		if diagErr.context != nil {
			return diagErr.context
		}
	}
	return nil
}

func runtimeMessageFromError(err error) string {
	if err == nil {
		return ""
	}
	var sig raiseSignal
	if errors.As(err, &sig) {
		if msg := runtimeErrorValueMessage(sig.value); msg != "" {
			return msg
		}
	}
	var diagErr runtimeDiagnosticError
	if errors.As(err, &diagErr) {
		if diagErr.err != nil && diagErr.err != err {
			return runtimeMessageFromError(diagErr.err)
		}
	}
	return err.Error()
}

func runtimeErrorValueMessage(val runtime.Value) string {
	switch v := val.(type) {
	case runtime.ErrorValue:
		return v.Message
	case *runtime.ErrorValue:
		if v != nil {
			return v.Message
		}
	}
	return ""
}
