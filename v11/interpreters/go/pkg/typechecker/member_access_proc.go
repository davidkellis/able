package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func procMemberFunction(name string, proc ProcType, node ast.Node) (Type, []Diagnostic) {
	var diags []Diagnostic
	switch name {
	case "status":
		return FunctionType{
			Params: nil,
			Return: StructType{StructName: "ProcStatus"},
		}, diags
	case "value":
		return FunctionType{
			Params: nil,
			Return: makeValueUnion(proc.Result),
		}, diags
	case "cancel":
		return FunctionType{
			Params: nil,
			Return: PrimitiveType{Kind: PrimitiveNil},
		}, diags
	default:
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: proc handle has no member '%s'", name),
			Node:    node,
		})
		return UnknownType{}, diags
	}
}

func futureMemberFunction(name string, future FutureType, node ast.Node) (Type, []Diagnostic) {
	var diags []Diagnostic
	switch name {
	case "status":
		return FunctionType{
			Params: nil,
			Return: StructType{StructName: "ProcStatus"},
		}, diags
	case "value":
		return FunctionType{
			Params: nil,
			Return: makeValueUnion(future.Result),
		}, diags
	case "cancel":
		diags = append(diags, Diagnostic{
			Message: "typechecker: future handles do not support cancel()",
			Node:    node,
		})
		return UnknownType{}, diags
	default:
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: future handle has no member '%s'", name),
			Node:    node,
		})
		return UnknownType{}, diags
	}
}
