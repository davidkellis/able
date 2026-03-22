package compiler

import "able/interpreter-go/pkg/ast"

func normalizeKernelBuiltinTypeName(name string) string {
	switch name {
	case "KernelArray":
		return "Array"
	case "KernelChannel":
		return "Channel"
	case "KernelHashMap":
		return "HashMap"
	case "KernelMutex":
		return "Mutex"
	case "KernelRange":
		return "Range"
	case "KernelRangeFactory":
		return "RangeFactory"
	case "KernelRatio":
		return "Ratio"
	case "KernelAwaitable":
		return "Awaitable"
	case "KernelAwaitWaker":
		return "AwaitWaker"
	case "KernelAwaitRegistration":
		return "AwaitRegistration"
	case "KernelLess":
		return "Less"
	case "KernelGreater":
		return "Greater"
	case "KernelEqual":
		return "Equal"
	case "KernelOrdering":
		return "Ordering"
	default:
		return name
	}
}

func normalizeKernelBuiltinTypeExpr(expr ast.TypeExpression) ast.TypeExpression {
	if expr == nil {
		return nil
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return expr
		}
		name := normalizeKernelBuiltinTypeName(t.Name.Name)
		if name == t.Name.Name {
			return expr
		}
		return ast.Ty(name)
	case *ast.GenericTypeExpression:
		if t == nil {
			return expr
		}
		base := normalizeKernelBuiltinTypeExpr(t.Base)
		changed := base != t.Base
		args := make([]ast.TypeExpression, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			next := normalizeKernelBuiltinTypeExpr(arg)
			args = append(args, next)
			if next != arg {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return ast.NewGenericTypeExpression(base, args)
	case *ast.NullableTypeExpression:
		if t == nil {
			return expr
		}
		inner := normalizeKernelBuiltinTypeExpr(t.InnerType)
		if inner == t.InnerType {
			return expr
		}
		return ast.NewNullableTypeExpression(inner)
	case *ast.ResultTypeExpression:
		if t == nil {
			return expr
		}
		inner := normalizeKernelBuiltinTypeExpr(t.InnerType)
		if inner == t.InnerType {
			return expr
		}
		return ast.NewResultTypeExpression(inner)
	case *ast.UnionTypeExpression:
		if t == nil {
			return expr
		}
		changed := false
		members := make([]ast.TypeExpression, 0, len(t.Members))
		for _, member := range t.Members {
			next := normalizeKernelBuiltinTypeExpr(member)
			members = append(members, next)
			if next != member {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return ast.NewUnionTypeExpression(members)
	case *ast.FunctionTypeExpression:
		if t == nil {
			return expr
		}
		ret := normalizeKernelBuiltinTypeExpr(t.ReturnType)
		changed := ret != t.ReturnType
		params := make([]ast.TypeExpression, 0, len(t.ParamTypes))
		for _, param := range t.ParamTypes {
			next := normalizeKernelBuiltinTypeExpr(param)
			params = append(params, next)
			if next != param {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return ast.NewFunctionTypeExpression(params, ret)
	default:
		return expr
	}
}
