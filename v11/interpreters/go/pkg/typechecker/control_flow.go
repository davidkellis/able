package typechecker

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

func (c *Checker) checkWhileLoop(env *Environment, loop *ast.WhileLoop) ([]Diagnostic, Type) {
	if loop == nil {
		return nil, PrimitiveType{Kind: PrimitiveNil}
	}
	var diags []Diagnostic

	condDiags, condType := c.checkExpression(env, loop.Condition)
	diags = append(diags, condDiags...)
	if !typeAssignable(condType, PrimitiveType{Kind: PrimitiveBool}) && !isUnknownType(condType) {
		diags = append(diags, Diagnostic{
			Message: "typechecker: while condition must be bool",
			Node:    loop.Condition,
		})
	}

	bodyType := Type(UnknownType{})
	c.pushLoopContext()
	if loop.Body != nil {
		bodyEnv := env.Extend()
		bodyDiags, inferred := c.checkExpression(bodyEnv, loop.Body)
		diags = append(diags, bodyDiags...)
		bodyType = inferred
	}
	breakType := c.popLoopContext()
	if bodyType == nil {
		bodyType = UnknownType{}
	}
	if breakType == nil {
		breakType = UnknownType{}
	}
	resultType := mergeTypesAllowUnion(bodyType, breakType)
	if resultType == nil {
		resultType = UnknownType{}
	}
	c.infer.set(loop, resultType)
	return diags, resultType
}

func (c *Checker) checkForLoop(env *Environment, loop *ast.ForLoop) ([]Diagnostic, Type) {
	if loop == nil {
		return nil, PrimitiveType{Kind: PrimitiveNil}
	}

	var diags []Diagnostic

	iterDiags, iterableType := c.checkExpression(env, loop.Iterable)
	diags = append(diags, iterDiags...)
	elementType, ok := iterableElementType(iterableType)
	if !ok {
		if iterableType == nil || isUnknownType(iterableType) {
			elementType = UnknownType{}
		} else {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: for-loop iterable must be array, range, or iterator, got %s", typeName(iterableType)),
				Node:    loop.Iterable,
			})
			elementType = UnknownType{}
		}
	}
	if elementType == nil {
		elementType = UnknownType{}
	}

	loopEnv := env.Extend()
	diags = append(diags, c.validateForLoopPattern(loop.Pattern, elementType)...)
	if target, ok := loop.Pattern.(ast.AssignmentTarget); ok {
		diags = append(diags, c.bindPattern(loopEnv, target, elementType, true, nil)...)
	} else if loop.Pattern != nil {
		if node, ok := loop.Pattern.(ast.Node); ok {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: unsupported loop pattern %T", loop.Pattern),
				Node:    node,
			})
		}
	}

	c.pushLoopContext()
	bodyType := Type(UnknownType{})
	if loop.Body != nil {
		bodyDiags, inferred := c.checkExpression(loopEnv, loop.Body)
		diags = append(diags, bodyDiags...)
		bodyType = inferred
	}
	breakType := c.popLoopContext()
	if bodyType == nil {
		bodyType = UnknownType{}
	}
	if breakType == nil {
		breakType = UnknownType{}
	}
	resultType := mergeTypesAllowUnion(bodyType, breakType)
	if resultType == nil {
		resultType = UnknownType{}
	}
	c.infer.set(loop, resultType)
	return diags, resultType
}

func (c *Checker) validateForLoopPattern(pattern ast.Pattern, elementType Type) []Diagnostic {
	if pattern == nil || elementType == nil || isUnknownType(elementType) {
		return nil
	}
	typed, ok := pattern.(*ast.TypedPattern)
	if !ok || typed.TypeAnnotation == nil {
		return nil
	}
	expected := c.resolveTypeReference(typed.TypeAnnotation)
	if expected == nil || isUnknownType(expected) {
		return nil
	}
	if typeAssignable(elementType, expected) && typeAssignable(expected, elementType) {
		return nil
	}
	return []Diagnostic{{
		Message: fmt.Sprintf(
			"typechecker: for-loop pattern expects type %s, got %s",
			typeName(expected),
			typeName(elementType),
		),
		Node: typed,
	}}
}

func (c *Checker) checkBreakpointExpression(env *Environment, expr *ast.BreakpointExpression) ([]Diagnostic, Type) {
	if expr == nil {
		return nil, UnknownType{}
	}
	var diags []Diagnostic
	if expr.Label == nil {
		diags = append(diags, Diagnostic{
			Message: "typechecker: breakpoint requires a label",
			Node:    expr,
		})
		c.infer.set(expr, UnknownType{})
		return diags, UnknownType{}
	}
	label := expr.Label.Name
	if label == "" {
		diags = append(diags, Diagnostic{
			Message: "typechecker: breakpoint label cannot be empty",
			Node:    expr,
		})
	}
	c.pushBreakpointLabel(label)
	bodyDiags, bodyType := c.checkExpression(env, expr.Body)
	c.popBreakpointLabel()
	diags = append(diags, bodyDiags...)
	c.infer.set(expr, bodyType)
	return diags, bodyType
}

func (c *Checker) checkBreakStatement(env *Environment, stmt *ast.BreakStatement) []Diagnostic {
	if stmt == nil {
		return nil
	}
	var diags []Diagnostic
	inLoop := c.inLoopContext()
	hasLabel := stmt.Label != nil && stmt.Label.Name != ""
	if !inLoop && !hasLabel {
		diags = append(diags, Diagnostic{
			Message: "typechecker: break statement must appear inside a loop",
			Node:    stmt,
		})
	}
	if hasLabel {
		label := stmt.Label.Name
		if !c.hasBreakpointLabel(label) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: unknown break label '%s'", label),
				Node:    stmt,
			})
		}
	}
	breakType := Type(PrimitiveType{Kind: PrimitiveNil})
	if stmt.Value != nil {
		valueDiags, valueType := c.checkExpression(env, stmt.Value)
		diags = append(diags, valueDiags...)
		if valueType != nil {
			breakType = valueType
		}
	}
	if inLoop {
		c.recordBreakType(breakType)
	}
	c.infer.set(stmt, PrimitiveType{Kind: PrimitiveNil})
	return diags
}

func (c *Checker) checkContinueStatement(stmt *ast.ContinueStatement) []Diagnostic {
	if stmt == nil {
		return nil
	}
	var diags []Diagnostic
	if !c.inLoopContext() {
		diags = append(diags, Diagnostic{
			Message: "typechecker: continue statement must appear inside a loop",
			Node:    stmt,
		})
	}
	if stmt.Label != nil {
		diags = append(diags, Diagnostic{
			Message: "typechecker: labeled continue is not supported",
			Node:    stmt,
		})
	}
	c.infer.set(stmt, PrimitiveType{Kind: PrimitiveNil})
	return diags
}
