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
	return g.inferredBodyTypeExpr(info.Package, info.Definition.Body)
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
