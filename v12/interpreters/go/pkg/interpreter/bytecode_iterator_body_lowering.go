package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func iteratorLiteralBindingNames(expr *ast.IteratorLiteral) []string {
	bindingName := "gen"
	if expr != nil && expr.Binding != nil && expr.Binding.Name != "" {
		bindingName = expr.Binding.Name
	}
	if bindingName == "gen" {
		return []string{"gen"}
	}
	return []string{bindingName, "gen"}
}

func (i *Interpreter) lowerIteratorLiteralBodyToBytecode(expr *ast.IteratorLiteral, env *runtime.Environment) (*bytecodeProgram, []string, error) {
	if expr == nil {
		return nil, nil, nil
	}
	bindingNames := iteratorLiteralBindingNames(expr)
	params := make([]*ast.FunctionParameter, len(bindingNames))
	for idx, name := range bindingNames {
		params[idx] = ast.Param(name, nil)
	}
	def := ast.Fn("__able_iterator_body", params, expr.Body, nil, nil, nil, false, true)
	program, err := i.lowerFunctionDefinitionBytecodeWithEnv(def, env)
	if err != nil {
		return nil, nil, err
	}
	return program, bindingNames, nil
}
