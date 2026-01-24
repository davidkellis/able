package typechecker

import "able/interpreter-go/pkg/ast"

func (c *Checker) finalizeMemberAccessType(expr *ast.MemberAccessExpression, objectType Type, memberType Type) Type {
	final := memberType
	if expr != nil && expr.Safe && typeCanBeNil(objectType) && !typeCanBeNil(memberType) {
		if fn, ok := memberType.(FunctionType); ok {
			wrapped := fn
			if wrapped.Return == nil {
				wrapped.Return = UnknownType{}
			}
			wrapped.Return = NullableType{Inner: wrapped.Return}
			final = wrapped
		} else if memberType == nil {
			final = NullableType{Inner: UnknownType{}}
		} else {
			final = NullableType{Inner: memberType}
		}
	}
	if expr != nil {
		c.infer.set(expr, final)
	}
	return final
}

func (c *Checker) appendUfcsCandidate(
	candidates []FunctionType,
	ufcs Type,
	expr *ast.MemberAccessExpression,
	wrapType Type,
) ([]FunctionType, Type, bool) {
	switch fn := ufcs.(type) {
	case FunctionType:
		return append(candidates, fn), nil, false
	case FunctionOverloadType:
		if len(candidates) == 0 {
			final := c.finalizeMemberAccessType(expr, wrapType, fn)
			return candidates, final, true
		}
	}
	return candidates, nil, false
}

func makeValueUnion(success Type) Type {
	futureErr := StructType{StructName: "FutureError"}
	members := []Type{success, futureErr}
	return UnionLiteralType{Members: members}
}

func stripNilFromUnion(u UnionLiteralType) Type {
	nonNil := make([]Type, 0, len(u.Members))
	for _, member := range u.Members {
		if prim, ok := member.(PrimitiveType); ok && prim.Kind == PrimitiveNil {
			continue
		}
		nonNil = append(nonNil, member)
	}
	switch len(nonNil) {
	case 0:
		return nil
	case 1:
		return nonNil[0]
	default:
		return UnionLiteralType{Members: nonNil}
	}
}

func isErrorStructType(ty StructType) bool {
	return ty.StructName == "Error"
}

func isErrorStructInstanceType(ty StructInstanceType) bool {
	return ty.StructName == "Error"
}

func isCallableType(t Type) bool {
	if t == nil {
		return false
	}
	switch t.(type) {
	case FunctionType:
		return true
	default:
		return false
	}
}

func (c *Checker) errorMemberType(memberName string) (Type, bool) {
	switch memberName {
	case "value":
		return UnknownType{}, true
	case "message":
		return FunctionType{
			Params: nil,
			Return: PrimitiveType{Kind: PrimitiveString},
		}, true
	case "cause":
		var inner Type = StructType{StructName: "Error"}
		if c != nil {
			inner = c.lookupErrorType()
		}
		return FunctionType{
			Params: nil,
			Return: NullableType{Inner: inner},
		}, true
	default:
		return nil, false
	}
}
