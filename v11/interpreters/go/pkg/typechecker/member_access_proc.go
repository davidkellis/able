package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (c *Checker) futureStatusType() Type {
	if c == nil || c.global == nil {
		return StructType{StructName: "FutureStatus"}
	}
	if typ, ok := c.global.Lookup("FutureStatus"); ok && typ != nil {
		return typ
	}
	return StructType{StructName: "FutureStatus"}
}

func (c *Checker) futureMemberFunction(name string, future FutureType, node ast.Node) (Type, []Diagnostic) {
	var diags []Diagnostic
	switch name {
	case "status":
		return FunctionType{
			Params: nil,
			Return: c.futureStatusType(),
		}, diags
	case "value":
		return FunctionType{
			Params: nil,
			Return: makeValueUnion(future.Result),
		}, diags
	case "cancel":
		return FunctionType{
			Params: nil,
			Return: PrimitiveType{Kind: PrimitiveNil},
		}, diags
	default:
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: future handle has no member '%s'", name),
			Node:    node,
		})
		return UnknownType{}, diags
	}
}
