//go:build !(js && wasm)

package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func bytecodeInferenceSimpleCheck(inferred bytecodeInferenceFacts, node ast.Node) bytecodeSimpleTypeCheck {
	if node == nil || len(inferred) == 0 {
		return bytecodeSimpleTypeCheckUnknown
	}
	return bytecodeTypecheckerSimpleCheck(inferred[node])
}

func bytecodeTypecheckerSimpleCheck(typ typechecker.Type) bytecodeSimpleTypeCheck {
	switch typed := typ.(type) {
	case typechecker.IntegerType:
		if typed.Suffix == "" {
			return bytecodeSimpleTypeCheckAnyInteger
		}
		return bytecodeSimpleTypeCheckForName(typed.Suffix)
	case typechecker.FloatType:
		if typed.Suffix == "" {
			return bytecodeSimpleTypeCheckAnyFloat
		}
		return bytecodeSimpleTypeCheckForName(typed.Suffix)
	case typechecker.PrimitiveType:
		switch typed.Kind {
		case typechecker.PrimitiveInt:
			return bytecodeSimpleTypeCheckAnyInteger
		case typechecker.PrimitiveFloat:
			return bytecodeSimpleTypeCheckAnyFloat
		case typechecker.PrimitiveBool:
			return bytecodeSimpleTypeCheckBool
		case typechecker.PrimitiveChar:
			return bytecodeSimpleTypeCheckChar
		case typechecker.PrimitiveString:
			return bytecodeSimpleTypeCheckString
		}
	case typechecker.AliasType:
		return bytecodeTypecheckerSimpleCheck(typed.Target)
	}
	return bytecodeSimpleTypeCheckUnknown
}
