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

// forwardFreshLambdaTypeExpr preserves the concrete callable carrier when all
// later statically known uses of an unannotated local lambda agree. Dynamic,
// escaping, insufficiently typed, and conflicting uses remain erased.
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
		constraints, unsafe := g.forwardCallableConstraintsFromStatement(ctx, name, lambda, ctx.blockStatements[idx])
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

// forwardCallableConstraintsFromStatement accepts direct identifier arguments
// and direct invocations whose signatures are fully static. It descends into
// an inline nested lambda only when the surrounding call proves that lambda's
// complete signature. Any other use remains on the runtime.Value path.
func (g *generator) forwardCallableConstraintsFromStatement(
	ctx *compileContext,
	name string,
	source *ast.LambdaExpression,
	stmt ast.Statement,
) ([]ast.TypeExpression, bool) {
	if g == nil || ctx == nil || name == "" || source == nil || stmt == nil {
		return nil, false
	}
	return g.forwardCallableConstraintsFromNode(ctx, ctx, name, source, stmt)
}

func (g *generator) forwardCallableConstraintsFromNode(
	useCtx *compileContext,
	declarationCtx *compileContext,
	name string,
	source *ast.LambdaExpression,
	node ast.Node,
) ([]ast.TypeExpression, bool) {
	if g == nil || useCtx == nil || declarationCtx == nil || name == "" || source == nil || node == nil {
		return nil, false
	}
	if block, ok := node.(*ast.BlockExpression); ok && block != nil {
		return g.forwardCallableConstraintsFromBlock(useCtx, declarationCtx, name, source, block)
	}
	nestedTypes := make(map[*ast.LambdaExpression]*ast.FunctionTypeExpression)
	ast.Walk(node, func(current ast.Node) bool {
		switch current.(type) {
		case *ast.BlockExpression, *ast.FunctionDefinition, *ast.LambdaExpression, *ast.IteratorLiteral:
			return false
		}
		call, ok := current.(*ast.FunctionCall)
		if !ok || call == nil {
			return true
		}
		for argIdx, arg := range call.Arguments {
			nested, ok := arg.(*ast.LambdaExpression)
			if !ok || nested == nil || !astNodeReferencesIdentifier(nested.Body, name) {
				continue
			}
			expected, ok := g.forwardStaticCallArgumentTypeExpr(useCtx, call, argIdx)
			if !ok || expected == nil {
				continue
			}
			normalized := g.lowerNormalizedTypeExpr(useCtx, expected)
			fnType, ok := normalized.(*ast.FunctionTypeExpression)
			if !ok || fnType == nil ||
				!g.typeExprFullyBound(useCtx.packageName, normalized) ||
				!g.lambdaExpressionMatchesExpectedFunctionType(useCtx, nested, fnType) {
				continue
			}
			nestedTypes[nested] = fnType
		}
		return true
	})

	allowed := make(map[*ast.Identifier]struct{})
	constraints := make([]ast.TypeExpression, 0, 1)
	invalidDirectUse := false
	ast.Walk(node, func(current ast.Node) bool {
		switch value := current.(type) {
		case *ast.FunctionDefinition, *ast.IteratorLiteral:
			if astNodeReferencesIdentifier(current, name) {
				invalidDirectUse = true
			}
			return false
		case *ast.BlockExpression:
			blockConstraints, unsafe := g.forwardCallableConstraintsFromBlock(
				useCtx,
				declarationCtx,
				name,
				source,
				value,
			)
			if unsafe {
				invalidDirectUse = true
				return false
			}
			constraints = append(constraints, blockConstraints...)
			return false
		case *ast.LambdaExpression:
			if !astNodeReferencesIdentifier(value.Body, name) {
				return false
			}
			expected := nestedTypes[value]
			if expected == nil || lambdaBindsIdentifier(value, name) || astNodeDeclaresIdentifier(value.Body, name) {
				invalidDirectUse = true
				return false
			}
			nestedCtx, ok := g.forwardNestedLambdaContext(useCtx, value, expected)
			if !ok {
				invalidDirectUse = true
				return false
			}
			nestedConstraints, unsafe := g.forwardCallableConstraintsFromNode(
				nestedCtx,
				declarationCtx,
				name,
				source,
				value.Body,
			)
			if unsafe {
				invalidDirectUse = true
				return false
			}
			constraints = append(constraints, nestedConstraints...)
			return false
		}
		call, ok := current.(*ast.FunctionCall)
		if !ok || call == nil {
			return true
		}
		if callee, ok := call.Callee.(*ast.Identifier); ok && callee != nil && callee.Name == name {
			constraint, ok := g.forwardCallableInvocationTypeExpr(declarationCtx, useCtx, source, call)
			if !ok || constraint == nil {
				invalidDirectUse = true
			} else {
				allowed[callee] = struct{}{}
				constraints = append(constraints, constraint)
			}
		}
		for argIdx, arg := range call.Arguments {
			ident, ok := arg.(*ast.Identifier)
			if !ok || ident == nil || ident.Name != name {
				continue
			}
			paramType, ok := g.forwardStaticCallArgumentTypeExpr(useCtx, call, argIdx)
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
	ast.Walk(node, func(current ast.Node) bool {
		switch current.(type) {
		case *ast.BlockExpression, *ast.FunctionDefinition, *ast.LambdaExpression, *ast.IteratorLiteral:
			return false
		}
		ident, ok := current.(*ast.Identifier)
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

func (g *generator) forwardCallableConstraintsFromBlock(
	parent *compileContext,
	declarationCtx *compileContext,
	name string,
	source *ast.LambdaExpression,
	block *ast.BlockExpression,
) ([]ast.TypeExpression, bool) {
	if g == nil || parent == nil || declarationCtx == nil || name == "" || source == nil || block == nil {
		return nil, false
	}
	blockCtx := parent.child()
	if blockCtx == nil {
		return nil, true
	}
	blockCtx.blockStatements = block.Body
	constraints := make([]ast.TypeExpression, 0, 1)
	for idx, stmt := range block.Body {
		blockCtx.statementIndex = idx
		next, unsafe := g.forwardCallableConstraintsFromNode(blockCtx, declarationCtx, name, source, stmt)
		if unsafe {
			return nil, true
		}
		constraints = append(constraints, next...)
		g.forwardCallableRecordBinding(blockCtx, stmt)
	}
	return constraints, false
}

func (g *generator) forwardCallableRecordBinding(ctx *compileContext, stmt ast.Statement) {
	if g == nil || ctx == nil || stmt == nil {
		return
	}
	assign, ok := stmt.(*ast.AssignmentExpression)
	if !ok || assign == nil || assign.Right == nil {
		return
	}
	name, explicit, ok := g.assignmentTargetName(assign.Left)
	if !ok || name == "" {
		return
	}
	if assign.Operator != ast.AssignmentDeclare {
		if _, exists := ctx.lookup(name); exists {
			return
		}
	}
	typeExpr := explicit
	if typeExpr == nil {
		inferred, ok := g.inferExpressionTypeExpr(ctx, assign.Right, "")
		if !ok || inferred == nil {
			return
		}
		typeExpr = inferred
	}
	typeExpr = g.lowerNormalizedTypeExpr(ctx, typeExpr)
	if typeExpr == nil || !g.typeExprFullyBound(ctx.packageName, typeExpr) {
		return
	}
	goType, ok := g.lowerCarrierType(ctx, typeExpr)
	if !ok || goType == "" || goType == "runtime.Value" || goType == "any" {
		return
	}
	ctx.locals[name] = paramInfo{
		Name:     name,
		GoName:   sanitizeIdent(name),
		GoType:   goType,
		TypeExpr: typeExpr,
	}
}

func (g *generator) forwardNestedLambdaContext(
	ctx *compileContext,
	lambda *ast.LambdaExpression,
	expected *ast.FunctionTypeExpression,
) (*compileContext, bool) {
	if g == nil || ctx == nil || lambda == nil || expected == nil || len(lambda.Params) != len(expected.ParamTypes) {
		return nil, false
	}
	child := ctx.closureChild()
	if child == nil {
		return nil, false
	}
	child.expectedTypeExpr = nil
	for idx, param := range lambda.Params {
		if param == nil {
			return nil, false
		}
		ident, ok := param.Name.(*ast.Identifier)
		if !ok || ident == nil || ident.Name == "" {
			return nil, false
		}
		paramType := g.lowerNormalizedTypeExpr(ctx, expected.ParamTypes[idx])
		goType, ok := g.lowerCarrierType(ctx, paramType)
		if !ok || goType == "" || goType == "runtime.Value" || goType == "any" {
			return nil, false
		}
		child.locals[ident.Name] = paramInfo{
			Name:     ident.Name,
			GoName:   safeParamName(ident.Name, idx),
			GoType:   goType,
			TypeExpr: paramType,
		}
	}
	return child, true
}

func (g *generator) forwardCallableInvocationTypeExpr(
	declarationCtx *compileContext,
	useCtx *compileContext,
	source *ast.LambdaExpression,
	call *ast.FunctionCall,
) (ast.TypeExpression, bool) {
	if g == nil || declarationCtx == nil || useCtx == nil || source == nil || call == nil ||
		len(source.Params) != len(call.Arguments) {
		return nil, false
	}
	paramTypes := make([]ast.TypeExpression, len(call.Arguments))
	for idx, arg := range call.Arguments {
		inferred, ok := g.inferExpressionTypeExpr(useCtx, arg, "")
		if !ok || inferred == nil {
			return nil, false
		}
		inferred = g.lowerNormalizedTypeExpr(useCtx, inferred)
		goType, mapped := g.lowerCarrierType(useCtx, inferred)
		if !mapped || goType == "" || goType == "runtime.Value" || goType == "any" ||
			!g.typeExprFullyBound(useCtx.packageName, inferred) {
			return nil, false
		}
		paramTypes[idx] = inferred
	}
	returnType, ok := g.forwardFreshLambdaReturnTypeExpr(declarationCtx, source, paramTypes)
	if !ok || returnType == nil {
		return nil, false
	}
	signature := ast.NewFunctionTypeExpression(paramTypes, returnType)
	if !g.typeExprFullyBound(declarationCtx.packageName, signature) ||
		!g.lambdaExpressionMatchesExpectedFunctionType(declarationCtx, source, signature) {
		return nil, false
	}
	return signature, true
}

func (g *generator) forwardFreshLambdaReturnTypeExpr(
	ctx *compileContext,
	lambda *ast.LambdaExpression,
	paramTypes []ast.TypeExpression,
) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || lambda == nil || len(lambda.Params) != len(paramTypes) {
		return nil, false
	}
	if lambda.ReturnType != nil {
		result := g.lowerNormalizedTypeExpr(ctx, lambda.ReturnType)
		if result != nil && g.typeExprFullyBound(ctx.packageName, result) {
			return result, true
		}
		return nil, false
	}
	probe := detachedCallableInferenceContext(ctx)
	if probe == nil {
		return nil, false
	}
	lambdaCtx := probe.closureChild()
	if lambdaCtx == nil {
		return nil, false
	}
	lambdaCtx.environmentEffect = &compiledEnvironmentEffect{
		localIndependent: true,
		callees:          make(map[*functionInfo]struct{}),
	}
	for idx, param := range lambda.Params {
		if param == nil {
			return nil, false
		}
		ident, ok := param.Name.(*ast.Identifier)
		if !ok || ident == nil || ident.Name == "" {
			return nil, false
		}
		paramType := g.lowerNormalizedTypeExpr(ctx, paramTypes[idx])
		goType, ok := g.lowerCarrierType(ctx, paramType)
		if !ok || goType == "" || goType == "runtime.Value" || goType == "any" {
			return nil, false
		}
		lambdaCtx.locals[ident.Name] = paramInfo{
			Name:     ident.Name,
			GoName:   safeParamName(ident.Name, idx),
			GoType:   goType,
			TypeExpr: paramType,
		}
	}
	var bodyType string
	var ok bool
	if lambda.IsVerboseSyntax {
		block, isBlock := lambda.Body.(*ast.BlockExpression)
		if !isBlock || block == nil {
			return nil, false
		}
		_, _, bodyType, ok = g.compileLambdaBlockBody(lambdaCtx, "", block)
	} else {
		_, _, bodyType, ok = g.compileTailExpression(lambdaCtx, "", lambda.Body)
	}
	if !ok || bodyType == "" || bodyType == "runtime.Value" || bodyType == "any" || g.isVoidType(bodyType) {
		return nil, false
	}
	result, ok := g.typeExprForGoType(bodyType)
	if !ok || result == nil {
		return nil, false
	}
	result = g.lowerNormalizedTypeExpr(ctx, result)
	return result, result != nil && g.typeExprFullyBound(ctx.packageName, result)
}

func detachedCallableInferenceContext(ctx *compileContext) *compileContext {
	if ctx == nil {
		return nil
	}
	probe := ctx.child()
	if probe == nil {
		return nil
	}
	visible := make(map[string]paramInfo)
	var chain []*compileContext
	for current := ctx; current != nil; current = current.parent {
		chain = append(chain, current)
	}
	for idx := len(chain) - 1; idx >= 0; idx-- {
		for name, info := range chain[idx].params {
			visible[name] = info
		}
		for name, info := range chain[idx].locals {
			visible[name] = info
		}
	}
	probe.parent = nil
	probe.params = nil
	probe.locals = visible
	probe.reason = ""
	probe.temps = new(int)
	probe.blockStatements = nil
	probe.statementIndex = 0
	probe.analysisOnly = true
	probe.originExtractions = nil
	probe.coercedNominalOrigins = nil
	probe.environmentEffect = &compiledEnvironmentEffect{
		localIndependent: true,
		callees:          make(map[*functionInfo]struct{}),
	}
	return probe
}

func astNodeReferencesIdentifier(node ast.Node, name string) bool {
	if node == nil || name == "" {
		return false
	}
	found := false
	ast.Walk(node, func(current ast.Node) bool {
		ident, ok := current.(*ast.Identifier)
		if ok && ident != nil && ident.Name == name {
			found = true
			return false
		}
		return !found
	})
	return found
}

func astNodeDeclaresIdentifier(node ast.Node, name string) bool {
	if node == nil || name == "" {
		return false
	}
	found := false
	ast.Walk(node, func(current ast.Node) bool {
		assign, ok := current.(*ast.AssignmentExpression)
		if !ok || assign == nil || assign.Operator != ast.AssignmentDeclare {
			return !found
		}
		ident, ok := assign.Left.(*ast.Identifier)
		if ok && ident != nil && ident.Name == name {
			found = true
			return false
		}
		return !found
	})
	return found
}

func lambdaBindsIdentifier(lambda *ast.LambdaExpression, name string) bool {
	if lambda == nil || name == "" {
		return false
	}
	for _, param := range lambda.Params {
		if param == nil {
			continue
		}
		ident, ok := param.Name.(*ast.Identifier)
		if ok && ident != nil && ident.Name == name {
			return true
		}
	}
	return false
}
