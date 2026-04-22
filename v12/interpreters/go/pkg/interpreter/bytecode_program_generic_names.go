package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func buildFunctionGenericNameSet(def *ast.FunctionDefinition, methodSet *runtime.MethodSet) map[string]struct{} {
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
	if def != nil {
		for _, gp := range def.GenericParams {
			if gp == nil || gp.Name == nil {
				continue
			}
			add(gp.Name.Name)
		}
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
		switch decl := fn.Declaration.(type) {
		case *ast.FunctionDefinition:
			program.returnGenericNames = buildFunctionGenericNameSet(decl, fn.MethodSet)
		default:
			program.returnGenericNames = nil
		}
		program.returnGenericNamesCached = true
	}
	fn.Bytecode = program
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
