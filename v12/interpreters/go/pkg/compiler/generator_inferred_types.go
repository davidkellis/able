package compiler

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func (g *generator) setTypecheckInference(inferred map[string]typechecker.InferenceMap) {
	if g == nil {
		return
	}
	g.inferredTypes = inferred
}

func (g *generator) inferredExpressionTypeExpr(ctx *compileContext, expr ast.Expression) ast.TypeExpression {
	if g == nil || ctx == nil || expr == nil || len(g.inferredTypes) == 0 {
		return nil
	}
	pkgInferred := g.inferredTypes[ctx.packageName]
	if len(pkgInferred) == 0 {
		return nil
	}
	typ, ok := pkgInferred[expr]
	if !ok || typ == nil {
		return nil
	}
	if _, unknown := typ.(typechecker.UnknownType); unknown {
		return nil
	}
	return g.lowerNormalizedTypeExpr(ctx, g.typeExprFromInferredType(typ))
}

func (g *generator) inferredBodyTypeExpr(pkgName string, body *ast.BlockExpression) ast.TypeExpression {
	if g == nil || body == nil || len(g.inferredTypes) == 0 {
		return nil
	}
	pkgInferred := g.inferredTypes[pkgName]
	if len(pkgInferred) == 0 {
		return nil
	}
	typ, ok := pkgInferred[body]
	if !ok || typ == nil {
		return nil
	}
	if _, unknown := typ.(typechecker.UnknownType); unknown {
		return nil
	}
	return normalizeTypeExprForPackage(g, pkgName, g.typeExprFromInferredType(typ))
}

func (g *generator) functionDeclaredOrInferredReturnTypeExpr(info *functionInfo) ast.TypeExpression {
	if g == nil || info == nil || info.Definition == nil {
		return nil
	}
	if info.Definition.ReturnType != nil {
		return normalizeTypeExprForPackage(g, info.Package, info.Definition.ReturnType)
	}
	inferred := g.inferredBodyTypeExpr(info.Package, info.Definition.Body)
	if isNilTypeExpression(inferred) && !functionBodyHasExplicitNilTail(info.Definition.Body) {
		hasValueReturn, hasNonNilValueReturn := functionBodyValueReturnShape(info.Definition.Body)
		switch {
		case hasNonNilValueReturn:
			// The checker can report the implicit fallthrough nil even when an
			// earlier branch returns a value. Keep the carrier open until those
			// branch return types are joined.
			return ast.NewWildcardTypeExpression()
		case !hasValueReturn && g.functionHasImplicitVoidConditionalTail(info):
			// A final if-without-else joins its side-effect-only branch with an
			// implicit nil in the checker. It has no observable value when every
			// branch is void, so keep recursive calls on the static void ABI.
			return ast.Ty("void")
		case hasValueReturn:
			// Explicit `return nil` is an observable nil result.
			return inferred
		}
		return ast.NewWildcardTypeExpression()
	}
	return inferred
}

func functionBodyHasExplicitNilTail(body *ast.BlockExpression) bool {
	if body == nil || len(body.Body) == 0 {
		return false
	}
	switch tail := body.Body[len(body.Body)-1].(type) {
	case *ast.NilLiteral:
		return tail != nil
	case *ast.ReturnStatement:
		_, ok := tail.Argument.(*ast.NilLiteral)
		return ok
	default:
		return false
	}
}

func functionBodyValueReturnShape(body *ast.BlockExpression) (hasValueReturn bool, hasNonNilValueReturn bool) {
	if body == nil {
		return false, false
	}
	ast.Walk(body, func(node ast.Node) bool {
		if hasNonNilValueReturn {
			return false
		}
		switch current := node.(type) {
		case *ast.FunctionDefinition, *ast.LambdaExpression, *ast.IteratorLiteral:
			// A nested callable owns its own return contract.
			return false
		case *ast.ReturnStatement:
			if current == nil || current.Argument == nil {
				return false
			}
			hasValueReturn = true
			if _, isNil := current.Argument.(*ast.NilLiteral); !isNil {
				hasNonNilValueReturn = true
			}
			return false
		default:
			return true
		}
	})
	return hasValueReturn, hasNonNilValueReturn
}

func (g *generator) functionHasImplicitVoidConditionalTail(info *functionInfo) bool {
	if g == nil || info == nil || info.Definition == nil || info.Definition.Body == nil {
		return false
	}
	body := info.Definition.Body.Body
	if len(body) == 0 {
		return false
	}
	tail, ok := body[len(body)-1].(*ast.IfExpression)
	if !ok || tail == nil || tail.ElseBody != nil {
		return false
	}
	if g.blockMayProduceNonNilTail(info, tail.IfBody) {
		return false
	}
	for _, clause := range tail.ElseIfClauses {
		if clause != nil && g.blockMayProduceNonNilTail(info, clause.Body) {
			return false
		}
	}
	return true
}

func (g *generator) blockMayProduceNonNilTail(info *functionInfo, body *ast.BlockExpression) bool {
	if body == nil || len(body.Body) == 0 {
		return false
	}
	if inferred := g.inferredBodyTypeExpr(info.Package, body); inferred != nil {
		return !isNilTypeExpression(inferred) && typeExpressionToString(inferred) != "void"
	}
	switch tail := body.Body[len(body.Body)-1].(type) {
	case *ast.NilLiteral:
		return false
	case *ast.ReturnStatement:
		if tail == nil || tail.Argument == nil {
			return false
		}
		_, isNil := tail.Argument.(*ast.NilLiteral)
		return !isNil
	case *ast.FunctionCall:
		ident, ok := tail.Callee.(*ast.Identifier)
		return !ok || ident == nil || info.Definition.ID == nil || ident.Name != info.Definition.ID.Name
	case *ast.IfExpression:
		if tail == nil || g.blockMayProduceNonNilTail(info, tail.IfBody) {
			return true
		}
		for _, clause := range tail.ElseIfClauses {
			if clause != nil && g.blockMayProduceNonNilTail(info, clause.Body) {
				return true
			}
		}
		return tail.ElseBody != nil && g.blockMayProduceNonNilTail(info, tail.ElseBody)
	case *ast.BlockExpression:
		return g.blockMayProduceNonNilTail(info, tail)
	default:
		return true
	}
}

func (g *generator) typeExprFromInferredType(typ typechecker.Type) ast.TypeExpression {
	if typ == nil {
		return ast.NewWildcardTypeExpression()
	}
	switch v := typ.(type) {
	case typechecker.UnknownType:
		return ast.NewWildcardTypeExpression()
	case typechecker.TypeParameterType:
		if v.ParameterName == "" {
			return ast.NewWildcardTypeExpression()
		}
		return ast.Ty(v.ParameterName)
	case typechecker.PrimitiveType:
		switch v.Kind {
		case typechecker.PrimitiveNil:
			return ast.Ty("nil")
		case typechecker.PrimitiveBool:
			return ast.Ty("bool")
		case typechecker.PrimitiveChar:
			return ast.Ty("char")
		case typechecker.PrimitiveString:
			return ast.Ty("String")
		case typechecker.PrimitiveInt:
			return ast.Ty("int")
		case typechecker.PrimitiveFloat:
			return ast.Ty("float")
		case typechecker.PrimitiveIoHandle:
			return ast.Ty("IoHandle")
		case typechecker.PrimitiveProcHandle:
			return ast.Ty("ProcHandle")
		default:
			return ast.Ty(v.Name())
		}
	case typechecker.IntegerType:
		return ast.Ty(v.Suffix)
	case typechecker.FloatType:
		return ast.Ty(v.Suffix)
	case typechecker.StructType:
		return typeExprWithWildcards(v.StructName, len(v.TypeParams))
	case typechecker.StructInstanceType:
		if len(v.TypeArgs) == 0 {
			return ast.Ty(v.StructName)
		}
		args := make([]ast.TypeExpression, len(v.TypeArgs))
		for idx, arg := range v.TypeArgs {
			args[idx] = g.typeExprFromInferredType(arg)
		}
		return ast.NewGenericTypeExpression(ast.Ty(v.StructName), args)
	case typechecker.InterfaceType:
		return typeExprWithWildcards(v.InterfaceName, len(v.TypeParams))
	case typechecker.ArrayType:
		return ast.Gen(ast.Ty("Array"), g.typeExprFromInferredType(v.Element))
	case typechecker.RangeType:
		// Range expressions are specified by their iterable behavior. Rebuild the
		// observable surface type here instead of forcing the internal checker
		// placeholder onto the nominal kernel Range struct carrier.
		return ast.Gen(ast.Ty("Iterable"), g.typeExprFromInferredType(v.Element))
	case typechecker.IteratorType:
		return ast.Gen(ast.Ty("Iterator"), g.typeExprFromInferredType(v.Element))
	case typechecker.FutureType:
		return ast.Gen(ast.Ty("Future"), g.typeExprFromInferredType(v.Result))
	case typechecker.NullableType:
		return ast.NewNullableTypeExpression(g.typeExprFromInferredType(v.Inner))
	case typechecker.UnionLiteralType:
		members := make([]ast.TypeExpression, len(v.Members))
		for idx, member := range v.Members {
			members[idx] = g.typeExprFromInferredType(member)
		}
		return ast.NewUnionTypeExpression(members)
	case typechecker.FunctionType:
		params := make([]ast.TypeExpression, len(v.Params))
		for idx, param := range v.Params {
			params[idx] = g.typeExprFromInferredType(param)
		}
		return ast.NewFunctionTypeExpression(params, g.typeExprFromInferredType(v.Return))
	case typechecker.AppliedType:
		base := g.typeExprFromInferredType(v.Base)
		if len(v.Arguments) == 0 {
			return base
		}
		args := make([]ast.TypeExpression, len(v.Arguments))
		for idx, arg := range v.Arguments {
			args[idx] = g.typeExprFromInferredType(arg)
		}
		return ast.NewGenericTypeExpression(base, args)
	case typechecker.AliasType:
		if v.Definition != nil && v.Definition.TargetType != nil {
			return v.Definition.TargetType
		}
		if v.AliasName != "" {
			return ast.Ty(v.AliasName)
		}
		return ast.NewWildcardTypeExpression()
	default:
		return ast.Ty(typ.Name())
	}
}

func typeExprWithWildcards(name string, count int) ast.TypeExpression {
	if name == "" {
		return ast.NewWildcardTypeExpression()
	}
	if count <= 0 {
		return ast.Ty(name)
	}
	args := make([]ast.TypeExpression, count)
	for idx := range args {
		args[idx] = ast.NewWildcardTypeExpression()
	}
	return ast.NewGenericTypeExpression(ast.Ty(name), args)
}
