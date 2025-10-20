package typechecker

import "able/interpreter10-go/pkg/ast"

func (c *Checker) checkBuiltinCallContext(name string, call *ast.FunctionCall) []Diagnostic {
	if name == "" || call == nil {
		return nil
	}

	switch name {
	case "proc_yield":
		if !c.inAsyncContext() {
			return []Diagnostic{{
				Message: "typechecker: proc_yield() may only be called from within proc or spawn bodies",
				Node:    call,
			}}
		}
	case "proc_cancelled":
		if !c.inAsyncContext() {
			return []Diagnostic{{
				Message: "typechecker: proc_cancelled must be called inside an asynchronous task",
				Node:    call,
			}}
		}
	case "proc_flush":
		return nil
	}

	// Unknown builtin name; nothing to report.
	return nil
}
