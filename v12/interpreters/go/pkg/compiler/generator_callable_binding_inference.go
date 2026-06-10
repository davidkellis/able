package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) forwardFreshLambdaBindingCarrier(
	ctx *compileContext,
	name string,
	right ast.Expression,
	declaring bool,
	explicit ast.TypeExpression,
) (ast.TypeExpression, string) {
	if g == nil || !declaring || explicit != nil {
		return nil, ""
	}
	lambda, ok := right.(*ast.LambdaExpression)
	if !ok || lambda == nil {
		return nil, ""
	}
	inferred, ok := g.forwardFreshLambdaTypeExpr(ctx, name, lambda)
	if !ok || inferred == nil {
		return nil, ""
	}
	goType, ok := g.lowerCarrierType(ctx, inferred)
	if !ok {
		return nil, ""
	}
	return inferred, goType
}

func inferredAssignmentTypeExpr(
	explicit ast.TypeExpression,
	forward ast.TypeExpression,
	existing paramInfo,
	exists bool,
) ast.TypeExpression {
	if explicit != nil {
		return explicit
	}
	if forward != nil {
		return forward
	}
	if exists {
		return existing.TypeExpr
	}
	return nil
}

// forwardFreshLambdaTypeExpr preserves the concrete callable carrier when an
// unannotated local lambda is later passed directly to a statically resolved
// callable parameter. Dynamic, indirect, and conflicting uses remain erased.
func (g *generator) forwardFreshLambdaTypeExpr(
	ctx *compileContext,
	name string,
	lambda *ast.LambdaExpression,
) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || name == "" || lambda == nil || len(ctx.blockStatements) == 0 {
		return nil, false
	}
	var inferred ast.TypeExpression
	var inferredKey string
	for idx := ctx.statementIndex + 1; idx < len(ctx.blockStatements); idx++ {
		constraints, unsafe := g.forwardCallableConstraintsFromStatement(ctx, name, ctx.blockStatements[idx])
		if unsafe {
			return nil, false
		}
		for _, constraint := range constraints {
			normalized := g.lowerNormalizedTypeExpr(ctx, constraint)
			fnType, ok := normalized.(*ast.FunctionTypeExpression)
			if !ok || fnType == nil ||
				!g.typeExprFullyBound(ctx.packageName, normalized) ||
				!g.lambdaExpressionMatchesExpectedFunctionType(ctx, lambda, fnType) {
				return nil, false
			}
			key := normalizeTypeExprIdentityKey(g, ctx.packageName, normalized)
			if key == "" {
				return nil, false
			}
			if inferred == nil {
				inferred = normalized
				inferredKey = key
				continue
			}
			if inferredKey != key {
				return nil, false
			}
		}
	}
	if inferred == nil {
		return nil, false
	}
	return inferred, true
}

// forwardCallableConstraintsFromStatement accepts only direct identifier
// arguments whose static parameter type is known. Any other use is kept on the
// existing runtime.Value path because it may be dynamic or may escape.
func (g *generator) forwardCallableConstraintsFromStatement(
	ctx *compileContext,
	name string,
	stmt ast.Statement,
) ([]ast.TypeExpression, bool) {
	if g == nil || ctx == nil || name == "" || stmt == nil {
		return nil, false
	}
	allowed := make(map[*ast.Identifier]struct{})
	constraints := make([]ast.TypeExpression, 0, 1)
	invalidDirectUse := false
	ast.Walk(stmt, func(node ast.Node) bool {
		switch node.(type) {
		case *ast.FunctionDefinition, *ast.LambdaExpression, *ast.IteratorLiteral:
			return false
		}
		call, ok := node.(*ast.FunctionCall)
		if !ok || call == nil {
			return true
		}
		for argIdx, arg := range call.Arguments {
			ident, ok := arg.(*ast.Identifier)
			if !ok || ident == nil || ident.Name != name {
				continue
			}
			paramType, ok := g.forwardStaticCallArgumentTypeExpr(ctx, call, argIdx)
			if !ok || paramType == nil {
				invalidDirectUse = true
				continue
			}
			allowed[ident] = struct{}{}
			constraints = append(constraints, paramType)
		}
		return true
	})
	if invalidDirectUse {
		return nil, true
	}
	unknownUse := false
	ast.Walk(stmt, func(node ast.Node) bool {
		switch node.(type) {
		case *ast.FunctionDefinition, *ast.LambdaExpression, *ast.IteratorLiteral:
			return false
		}
		ident, ok := node.(*ast.Identifier)
		if !ok || ident == nil || ident.Name != name {
			return true
		}
		if _, ok := allowed[ident]; !ok {
			unknownUse = true
		}
		return true
	})
	return constraints, unknownUse
}
