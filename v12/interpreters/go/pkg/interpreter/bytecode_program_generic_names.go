package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func buildFunctionGenericNameSet(node ast.Node, methodSet *runtime.MethodSet) map[string]struct{} {
	var names map[string]struct{}
	add := func(name string) {
		if name == "" {
			return
		}
		if names == nil {
			names = make(map[string]struct{}, 4)
		}
		names[name] = struct{}{}
	}
	var genericParams []*ast.GenericParameter
	switch decl := node.(type) {
	case *ast.FunctionDefinition:
		genericParams = decl.GenericParams
	case *ast.LambdaExpression:
		genericParams = decl.GenericParams
	}
	for _, gp := range genericParams {
		if gp == nil || gp.Name == nil {
			continue
		}
		add(gp.Name.Name)
	}
	if methodSet != nil {
		for _, gp := range methodSet.GenericParams {
			if gp == nil || gp.Name == nil {
				continue
			}
			add(gp.Name.Name)
		}
	}
	return names
}

func setFunctionBytecodeProgram(fn *runtime.FunctionValue, program *bytecodeProgram) {
	if fn == nil {
		return
	}
	if program != nil {
		// Programs cached for repeated lambda evaluations are shared by several
		// FunctionValue instances. Callable metadata is immutable for a given
		// declaration, so initialize it once before publishing the program and
		// never rewrite it on later attachments.
		if !program.returnGenericNamesCached {
			program.returnGenericNames = buildFunctionGenericNameSet(fn.Declaration, fn.MethodSet)
			program.returnGenericNamesCached = true
		}
		if !program.returnTypeMetadataCached {
			setBytecodeProgramReturnTypeMetadata(fn, program)
		}
	}
	fn.Bytecode = program
}

func callableReturnType(node ast.Node) ast.TypeExpression {
	switch decl := node.(type) {
	case *ast.FunctionDefinition:
		if decl != nil {
			return decl.ReturnType
		}
	case *ast.LambdaExpression:
		if decl != nil {
			return decl.ReturnType
		}
	}
	return nil
}

func setBytecodeProgramReturnTypeMetadata(fn *runtime.FunctionValue, program *bytecodeProgram) {
	if fn == nil || program == nil {
		return
	}
	returnType := callableReturnType(fn.Declaration)
	program.returnType = returnType
	program.returnSimpleType = cachedSimpleTypeName(returnType)
	program.returnSimpleCheck = bytecodeSimpleTypeCheckForName(program.returnSimpleType)
	program.returnNullableSimple = cachedNullableSimpleTypeName(returnType)
	program.returnTypeUsesGenerics = typeExpressionUsesGenerics(returnType, program.returnGenericNames)
	program.returnTypeMetadataCached = true
}

func bytecodeProgramReturnGenericNames(fn *runtime.FunctionValue, program *bytecodeProgram) map[string]struct{} {
	if program != nil && program.returnGenericNamesCached {
		return program.returnGenericNames
	}
	if fn == nil {
		return nil
	}
	return fn.GenericNameSet(nil)
}

func bytecodeInlineReturnGenericNames(fn *runtime.FunctionValue, program *bytecodeProgram) map[string]struct{} {
	if program != nil && program.frameLayout != nil && !program.frameLayout.returnTypeUsesGenerics {
		return nil
	}
	if program != nil && program.returnTypeMetadataCached && !program.returnTypeUsesGenerics {
		return nil
	}
	if program != nil && program.returnGenericNamesCached {
		return program.returnGenericNames
	}
	if fn == nil {
		return nil
	}
	return fn.GenericNameSet(nil)
}

func bytecodeFunctionReturnGenericNames(fn *runtime.FunctionValue) map[string]struct{} {
	if fn == nil {
		return nil
	}
	if program, ok := fn.Bytecode.(*bytecodeProgram); ok && program != nil {
		return bytecodeProgramReturnGenericNames(fn, program)
	}
	return fn.GenericNameSet(nil)
}
