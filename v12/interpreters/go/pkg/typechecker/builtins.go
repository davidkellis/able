package typechecker

import "able/interpreter-go/pkg/ast"

func (c *Checker) checkBuiltinCallContext(name string, call *ast.FunctionCall) []Diagnostic {
	if name == "" || call == nil {
		return nil
	}

	switch name {
	case "future_yield":
		if !c.inAsyncContext() {
			return []Diagnostic{{
				Message: "typechecker: future_yield() may only be called from within an asynchronous task",
				Node:    call,
			}}
		}
	case "future_cancelled":
		return nil
	case "future_flush":
		return nil
	}

	// Unknown builtin name; nothing to report.
	return nil
}
