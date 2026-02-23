package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileLocalFunctionDefinitionStatement(ctx *compileContext, def *ast.FunctionDefinition) ([]string, bool) {
	if def == nil || def.ID == nil || def.ID.Name == "" {
		ctx.setReason("unsupported local function definition")
		return nil, false
	}
	if def.IsMethodShorthand {
		ctx.setReason("unsupported local function definition")
		return nil, false
	}
	name := def.ID.Name
	current, hadCurrent := ctx.lookupCurrent(name)
	reuseBinding := hadCurrent && current.GoType == "runtime.Value" && current.GoName != ""
	binding := paramInfo{
		Name:      name,
		GoName:    sanitizeIdent(name),
		GoType:    "runtime.Value",
		Supported: true,
	}
	if reuseBinding {
		binding.GoName = current.GoName
	} else if hadCurrent {
		// Existing non-runtime bindings are shadowed with a fresh runtime local.
		binding.GoName = ctx.newTemp()
	}
	// Bind first so the function body can recursively reference itself.
	ctx.locals[name] = binding

	lambda := ast.NewLambdaExpression(def.Params, def.Body, def.ReturnType, def.GenericParams, def.WhereClause, true)
	valueExpr, valueType, ok := g.compileLambdaExpression(ctx, lambda, "runtime.Value")
	if !ok {
		if ctx.reason == "" {
			ctx.setReason("unsupported local function body")
		}
		return nil, false
	}
	if valueType != "runtime.Value" {
		ctx.setReason("unsupported local function body")
		return nil, false
	}
	if reuseBinding {
		// If the name already exists in the current scope, reuse the same binding.
		return []string{fmt.Sprintf("%s = %s", binding.GoName, valueExpr)}, true
	}
	return []string{fmt.Sprintf("var %s runtime.Value = %s", binding.GoName, valueExpr)}, true
}
