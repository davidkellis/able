package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) inferHandledFailureTypeExpr(ctx *compileContext, expr ast.Expression) ast.TypeExpression {
	return g.inferHandledFailureTypeExprSeen(ctx, expr, make(map[string]struct{}))
}

func (g *generator) inferHandledFailureTypeExprSeen(ctx *compileContext, expr ast.Expression, functionSeen map[string]struct{}) ast.TypeExpression {
	if g == nil || ctx == nil || expr == nil {
		return nil
	}
	candidates := make([]ast.TypeExpression, 0, 4)
	seen := make(map[string]struct{}, 4)
	g.collectHandledFailureTypeExprsFromExpr(ctx, expr, seen, functionSeen, &candidates)
	return g.joinHandledFailureTypeExprs(ctx, candidates)
}

func (g *generator) appendHandledFailureTypeExpr(ctx *compileContext, expr ast.TypeExpression, seen map[string]struct{}, out *[]ast.TypeExpression) {
	if g == nil || expr == nil || out == nil {
		return
	}
	normalized := g.lowerNormalizedTypeExpr(ctx, expr)
	key := ""
	if ctx != nil {
		key = normalizeTypeExprIdentityKey(g, ctx.packageName, normalized)
	}
	if key == "" {
		key = typeExpressionToString(normalized)
	}
	if key == "" {
		return
	}
	if _, ok := seen[key]; ok {
		return
	}
	seen[key] = struct{}{}
	*out = append(*out, normalized)
}

func (g *generator) joinHandledFailureTypeExprs(ctx *compileContext, candidates []ast.TypeExpression) ast.TypeExpression {
	if g == nil || len(candidates) == 0 {
		return nil
	}
	if len(candidates) == 1 {
		return g.lowerNormalizedTypeExpr(ctx, candidates[0])
	}
	goTypes := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		recovered, ok := g.joinCarrierTypeFromTypeExpr(ctx, candidate)
		if !ok || recovered == "" {
			return nil
		}
		goTypes = append(goTypes, recovered)
	}
	joined, ok := g.lowerJoinCarrier(ctx, goTypes...)
	if !ok || joined == "" {
		return nil
	}
	typeExpr, ok := g.typeExprForGoType(joined)
	if !ok || typeExpr == nil {
		return nil
	}
	return g.lowerNormalizedTypeExpr(ctx, typeExpr)
}

func (g *generator) failureTypeExprFromPropagatedSource(ctx *compileContext, source ast.Expression) ast.TypeExpression {
	return g.failureTypeExprFromPropagatedSourceSeen(ctx, source, make(map[string]struct{}))
}

func (g *generator) failureTypeExprFromPropagatedSourceSeen(ctx *compileContext, source ast.Expression, functionSeen map[string]struct{}) ast.TypeExpression {
	if g == nil || ctx == nil || source == nil {
		return nil
	}
	sourceTypeExpr, ok := g.inferExpressionTypeExpr(ctx, source, "")
	if !ok || sourceTypeExpr == nil {
		return nil
	}
	sourceTypeExpr = g.lowerNormalizedTypeExpr(ctx, sourceTypeExpr)
	if _, ok := sourceTypeExpr.(*ast.ResultTypeExpression); ok {
		return g.lowerNormalizedTypeExpr(ctx, ast.Ty("Error"))
	}
	if unionPkg, members, ok := g.expandedUnionMembersInPackage(ctx.packageName, sourceTypeExpr); ok {
		if info, ok := g.ensureNativeUnionInfo(unionPkg, members); ok && info != nil {
			if _, failure, ok := g.nativeUnionOrElseMembers(info.GoType); ok && failure != nil && failure.TypeExpr != nil {
				return normalizeTypeExprForPackage(g, unionPkg, failure.TypeExpr)
			}
		}
	}
	if call, ok := source.(*ast.FunctionCall); ok {
		return g.failureTypeExprFromFunctionCallSeen(ctx, call, functionSeen)
	}
	return nil
}

func (g *generator) failureTypeExprFromRaisedExpr(ctx *compileContext, expr ast.Expression) ast.TypeExpression {
	if g == nil || ctx == nil || expr == nil {
		return nil
	}
	exprType := ast.TypeExpression(nil)
	if lit, ok := expr.(*ast.StructLiteral); ok && lit != nil {
		// Prefer the compiler's syntax-aware struct-literal reconstruction here.
		// Typechecker-side nominal inference is intentionally package-agnostic and
		// can collapse imported selector aliases onto shadowing local names.
		exprType = g.lowerNormalizedTypeExpr(ctx, g.staticStructLiteralTypeExpr(ctx, lit, ""))
	}
	if exprType == nil {
		if inferred, ok := g.inferExpressionTypeExpr(ctx, expr, ""); ok && inferred != nil {
			exprType = g.lowerNormalizedTypeExpr(ctx, inferred)
		}
	}
	if exprType != nil {
		if goType, ok := g.lowerCarrierType(ctx, exprType); ok && goType != "" {
			switch g.typeCategory(goType) {
			case "struct", "interface", "union", "callable", "monoarray":
				return exprType
			}
			if innerType, nullable := g.nativeNullableValueInnerType(goType); nullable {
				switch g.typeCategory(innerType) {
				case "struct", "interface", "union", "callable", "monoarray":
					return exprType
				}
			}
		}
	}
	return g.lowerNormalizedTypeExpr(ctx, ast.Ty("Error"))
}

func (g *generator) failureTypeExprFromFunctionCallSeen(ctx *compileContext, call *ast.FunctionCall, functionSeen map[string]struct{}) ast.TypeExpression {
	if g == nil || ctx == nil || call == nil {
		return nil
	}
	if info, overload, ok := g.staticCallableForFailureInference(ctx, call); ok {
		return g.failureTypeExprFromResolvedCallableSeen(ctx, info, overload, functionSeen)
	}
	if lambda, ok := call.Callee.(*ast.LambdaExpression); ok {
		return g.failureTypeExprFromLambdaSeen(ctx, lambda, functionSeen)
	}
	return nil
}

func (g *generator) staticCallableForFailureInference(ctx *compileContext, call *ast.FunctionCall) (*functionInfo, *overloadInfo, bool) {
	if g == nil || ctx == nil || call == nil || call.Callee == nil {
		return nil, nil, false
	}
	if ident, ok := call.Callee.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
		return g.resolveStaticCallable(ctx, ident.Name)
	}
	if callee, ok := call.Callee.(*ast.MemberAccessExpression); ok && callee != nil && !callee.Safe {
		objectIdent, ok := callee.Object.(*ast.Identifier)
		if !ok || objectIdent == nil || objectIdent.Name == "" {
			return nil, nil, false
		}
		memberIdent, ok := callee.Member.(*ast.Identifier)
		if !ok || memberIdent == nil || memberIdent.Name == "" {
			return nil, nil, false
		}
		if _, found := ctx.lookup(objectIdent.Name); found {
			return nil, nil, false
		}
		return g.resolveStaticCallable(ctx, objectIdent.Name+"."+memberIdent.Name)
	}
	return nil, nil, false
}

func (g *generator) failureTypeExprFromResolvedCallableSeen(ctx *compileContext, info *functionInfo, overload *overloadInfo, functionSeen map[string]struct{}) ast.TypeExpression {
	if g == nil || ctx == nil {
		return nil
	}
	if info != nil {
		return g.failureTypeExprFromFunctionInfoSeen(ctx, info, functionSeen)
	}
	if overload == nil {
		return nil
	}
	candidates := make([]ast.TypeExpression, 0, len(overload.Entries))
	seen := make(map[string]struct{}, len(overload.Entries))
	for _, entry := range overload.Entries {
		g.appendHandledFailureTypeExpr(ctx, g.failureTypeExprFromFunctionInfoSeen(ctx, entry, functionSeen), seen, &candidates)
	}
	return g.joinHandledFailureTypeExprs(ctx, candidates)
}

func (g *generator) failureTypeExprFromFunctionInfoSeen(ctx *compileContext, info *functionInfo, functionSeen map[string]struct{}) ast.TypeExpression {
	if g == nil || ctx == nil || info == nil || info.Definition == nil || info.Definition.Body == nil {
		return nil
	}
	key := info.QualifiedName
	if key == "" {
		key = info.Name
	}
	if key == "" {
		key = info.GoName
	}
	if key == "" {
		return nil
	}
	if info.Package != "" {
		key = info.Package + "::" + key
	}
	if _, ok := functionSeen[key]; ok {
		return nil
	}
	functionSeen[key] = struct{}{}
	defer delete(functionSeen, key)

	pkgName := info.Package
	if pkgName == "" {
		pkgName = ctx.packageName
	}
	fnCtx := newCompileContext(g, info, g.functions[pkgName], g.overloads[pkgName], pkgName, genericNameSet(info.Definition.GenericParams))
	inferred := g.inferHandledFailureTypeExprSeen(fnCtx, info.Definition.Body, functionSeen)
	returnTypeExpr := g.functionReturnTypeExpr(info)
	if inferred == nil || returnTypeExpr == nil {
		return inferred
	}
	returnTypeExpr = g.lowerNormalizedTypeExpr(fnCtx, returnTypeExpr)
	goType, ok := g.joinCarrierTypeFromTypeExpr(fnCtx, inferred)
	if !ok || goType == "" {
		return inferred
	}
	if g.typeExprCompatibleWithCarrier(fnCtx, returnTypeExpr, goType) {
		return returnTypeExpr
	}
	return inferred
}

func (g *generator) failureTypeExprFromLambdaSeen(ctx *compileContext, expr *ast.LambdaExpression, functionSeen map[string]struct{}) ast.TypeExpression {
	if g == nil || ctx == nil || expr == nil || expr.Body == nil {
		return nil
	}
	lambdaCtx := ctx.closureChild()
	for idx, param := range expr.Params {
		if param == nil {
			continue
		}
		ident, ok := param.Name.(*ast.Identifier)
		if !ok || ident == nil || ident.Name == "" {
			continue
		}
		goType := "runtime.Value"
		if param.ParamType != nil {
			if mapped, ok := g.lowerCarrierType(ctx, param.ParamType); ok && mapped != "" {
				goType = mapped
			}
		}
		lambdaCtx.locals[ident.Name] = paramInfo{
			Name:     ident.Name,
			GoName:   safeParamName(ident.Name, idx),
			GoType:   goType,
			TypeExpr: param.ParamType,
		}
	}
	return g.inferHandledFailureTypeExprSeen(lambdaCtx, expr.Body, functionSeen)
}

func (g *generator) collectHandledFailureTypeExprsFromStmt(ctx *compileContext, stmt ast.Statement, seen map[string]struct{}, functionSeen map[string]struct{}, out *[]ast.TypeExpression) {
	if g == nil || stmt == nil {
		return
	}
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		if s.Argument != nil {
			g.collectHandledFailureTypeExprsFromExpr(ctx, s.Argument, seen, functionSeen, out)
		}
	case *ast.WhileLoop:
		g.collectHandledFailureTypeExprsFromExpr(ctx, s.Condition, seen, functionSeen, out)
		g.collectHandledFailureTypeExprsFromExpr(ctx, s.Body, seen, functionSeen, out)
	case *ast.ForLoop:
		g.collectHandledFailureTypeExprsFromExpr(ctx, s.Iterable, seen, functionSeen, out)
		g.collectHandledFailureTypeExprsFromExpr(ctx, s.Body, seen, functionSeen, out)
	case *ast.BreakStatement:
		if s.Value != nil {
			g.collectHandledFailureTypeExprsFromExpr(ctx, s.Value, seen, functionSeen, out)
		}
	case *ast.RaiseStatement:
		g.appendHandledFailureTypeExpr(ctx, g.failureTypeExprFromRaisedExpr(ctx, s.Expression), seen, out)
		g.collectHandledFailureTypeExprsFromExpr(ctx, s.Expression, seen, functionSeen, out)
	case *ast.YieldStatement:
		if s.Expression != nil {
			g.collectHandledFailureTypeExprsFromExpr(ctx, s.Expression, seen, functionSeen, out)
		}
	default:
		if expr, ok := stmt.(ast.Expression); ok {
			g.collectHandledFailureTypeExprsFromExpr(ctx, expr, seen, functionSeen, out)
		}
	}
}

func (g *generator) collectHandledFailureTypeExprsFromExpr(ctx *compileContext, expr ast.Expression, seen map[string]struct{}, functionSeen map[string]struct{}, out *[]ast.TypeExpression) {
	if g == nil || expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.BlockExpression:
		for _, stmt := range e.Body {
			g.collectHandledFailureTypeExprsFromStmt(ctx, stmt, seen, functionSeen, out)
		}
	case *ast.FunctionCall:
		g.appendHandledFailureTypeExpr(ctx, g.failureTypeExprFromPropagatedSourceSeen(ctx, e, functionSeen), seen, out)
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Callee, seen, functionSeen, out)
		for _, arg := range e.Arguments {
			g.collectHandledFailureTypeExprsFromExpr(ctx, arg, seen, functionSeen, out)
		}
	case *ast.MemberAccessExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Object, seen, functionSeen, out)
		if memberExpr, ok := e.Member.(ast.Expression); ok {
			g.collectHandledFailureTypeExprsFromExpr(ctx, memberExpr, seen, functionSeen, out)
		}
	case *ast.IndexExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Object, seen, functionSeen, out)
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Index, seen, functionSeen, out)
	case *ast.ArrayLiteral:
		for _, element := range e.Elements {
			g.collectHandledFailureTypeExprsFromExpr(ctx, element, seen, functionSeen, out)
		}
	case *ast.StructLiteral:
		for _, field := range e.Fields {
			if field != nil && field.Value != nil {
				g.collectHandledFailureTypeExprsFromExpr(ctx, field.Value, seen, functionSeen, out)
			}
		}
		for _, source := range e.FunctionalUpdateSources {
			g.collectHandledFailureTypeExprsFromExpr(ctx, source, seen, functionSeen, out)
		}
	case *ast.MapLiteral:
		for _, element := range e.Elements {
			switch item := element.(type) {
			case *ast.MapLiteralEntry:
				g.collectHandledFailureTypeExprsFromExpr(ctx, item.Key, seen, functionSeen, out)
				g.collectHandledFailureTypeExprsFromExpr(ctx, item.Value, seen, functionSeen, out)
			case *ast.MapLiteralSpread:
				g.collectHandledFailureTypeExprsFromExpr(ctx, item.Expression, seen, functionSeen, out)
			}
		}
	case *ast.TypeCastExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Expression, seen, functionSeen, out)
	case *ast.UnaryExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Operand, seen, functionSeen, out)
	case *ast.BinaryExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Left, seen, functionSeen, out)
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Right, seen, functionSeen, out)
	case *ast.RangeExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Start, seen, functionSeen, out)
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.End, seen, functionSeen, out)
	case *ast.StringInterpolation:
		for _, part := range e.Parts {
			g.collectHandledFailureTypeExprsFromExpr(ctx, part, seen, functionSeen, out)
		}
	case *ast.AssignmentExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Right, seen, functionSeen, out)
		switch target := e.Left.(type) {
		case *ast.IndexExpression:
			g.collectHandledFailureTypeExprsFromExpr(ctx, target.Object, seen, functionSeen, out)
			g.collectHandledFailureTypeExprsFromExpr(ctx, target.Index, seen, functionSeen, out)
		case *ast.MemberAccessExpression:
			g.collectHandledFailureTypeExprsFromExpr(ctx, target.Object, seen, functionSeen, out)
		default:
			if targetExpr, ok := target.(ast.Expression); ok {
				g.collectHandledFailureTypeExprsFromExpr(ctx, targetExpr, seen, functionSeen, out)
			}
		}
	case *ast.IfExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.IfCondition, seen, functionSeen, out)
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.IfBody, seen, functionSeen, out)
		for _, clause := range e.ElseIfClauses {
			if clause == nil {
				continue
			}
			g.collectHandledFailureTypeExprsFromExpr(ctx, clause.Condition, seen, functionSeen, out)
			g.collectHandledFailureTypeExprsFromExpr(ctx, clause.Body, seen, functionSeen, out)
		}
		if e.ElseBody != nil {
			g.collectHandledFailureTypeExprsFromExpr(ctx, e.ElseBody, seen, functionSeen, out)
		}
	case *ast.MatchExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Subject, seen, functionSeen, out)
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil {
				g.collectHandledFailureTypeExprsFromExpr(ctx, clause.Guard, seen, functionSeen, out)
			}
			if clause.Body != nil {
				g.collectHandledFailureTypeExprsFromExpr(ctx, clause.Body, seen, functionSeen, out)
			}
		}
	case *ast.RescueExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.MonitoredExpression, seen, functionSeen, out)
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil {
				g.collectHandledFailureTypeExprsFromExpr(ctx, clause.Guard, seen, functionSeen, out)
			}
			if clause.Body != nil {
				g.collectHandledFailureTypeExprsFromExpr(ctx, clause.Body, seen, functionSeen, out)
			}
		}
	case *ast.EnsureExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.TryExpression, seen, functionSeen, out)
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.EnsureBlock, seen, functionSeen, out)
	case *ast.OrElseExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Handler, seen, functionSeen, out)
	case *ast.PropagationExpression:
		g.appendHandledFailureTypeExpr(ctx, g.failureTypeExprFromPropagatedSourceSeen(ctx, e.Expression, functionSeen), seen, out)
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Expression, seen, functionSeen, out)
	case *ast.BreakpointExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Body, seen, functionSeen, out)
	case *ast.LoopExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Body, seen, functionSeen, out)
	case *ast.AwaitExpression:
		g.collectHandledFailureTypeExprsFromExpr(ctx, e.Expression, seen, functionSeen, out)
	}
}
