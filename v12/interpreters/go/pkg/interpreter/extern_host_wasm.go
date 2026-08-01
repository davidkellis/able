//go:build js && wasm

package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// externHostPackage keeps the shared interpreter state shape portable. Go
// plugin-based extern modules are unavailable on js/wasm; a future browser
// host ABI can replace this placeholder without altering AST evaluation.
type externHostPackage struct{}

type externHostInvoker func(i *Interpreter, args []runtime.Value) (runtime.Value, error)

func (i *Interpreter) isKernelExtern(name string) bool {
	return strings.HasPrefix(name, "__able_")
}

func (i *Interpreter) registerExternStatements(_ *ast.Module) {
}

func (i *Interpreter) invokeExternHostFunction(_ string, def *ast.ExternFunctionBody, _ []runtime.Value) (runtime.Value, error) {
	name := "<unknown>"
	if def != nil && def.Signature != nil && def.Signature.ID != nil && def.Signature.ID.Name != "" {
		name = def.Signature.ID.Name
	}
	return nil, fmt.Errorf("extern function %s is unavailable on js/wasm; browser host callbacks are not implemented", name)
}

func typeKey(expr ast.TypeExpression) string {
	if expr == nil {
		return "void"
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return "_"
		}
		return t.Name.Name
	case *ast.GenericTypeExpression:
		if t == nil {
			return "?"
		}
		args := make([]string, len(t.Arguments))
		for idx, arg := range t.Arguments {
			args[idx] = typeKey(arg)
		}
		return fmt.Sprintf("%s<%s>", typeKey(t.Base), strings.Join(args, ","))
	case *ast.NullableTypeExpression:
		return "?" + typeKey(t.InnerType)
	case *ast.ResultTypeExpression:
		return "!" + typeKey(t.InnerType)
	case *ast.UnionTypeExpression:
		if t == nil {
			return "?"
		}
		members := make([]string, len(t.Members))
		for idx, member := range t.Members {
			members[idx] = typeKey(member)
		}
		return strings.Join(members, "|")
	case *ast.FunctionTypeExpression:
		if t == nil {
			return "?"
		}
		params := make([]string, len(t.ParamTypes))
		for idx, param := range t.ParamTypes {
			params[idx] = typeKey(param)
		}
		return fmt.Sprintf("(%s)->%s", strings.Join(params, ","), typeKey(t.ReturnType))
	case *ast.WildcardTypeExpression:
		return "_"
	default:
		return fmt.Sprintf("%T", expr)
	}
}
