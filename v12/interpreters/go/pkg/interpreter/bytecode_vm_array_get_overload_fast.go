package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execCanonicalArrayGetOverloadMemberFast(callable runtime.Value, instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.name != "get" || instr.argCount != 1 {
		return nil, false, nil
	}
	if !vm.isCanonicalNullableArrayGetOverload(callable) {
		return nil, false, nil
	}
	return vm.execArrayGetMemberFast(instr, receiverIndex, argBase, callNode)
}

func (vm *bytecodeVM) isCanonicalNullableArrayGetOverload(callable runtime.Value) bool {
	overload := bytecodeArrayGetOverloadCallable(callable)
	if overload == nil {
		return false
	}
	version := vm.bytecodeMethodCacheVersion()
	if vm != nil && vm.arrayGetOverloadHot == overload && vm.arrayGetOverloadHotVersion == version {
		return vm.arrayGetOverloadHotOK
	}
	ok := vm.isCanonicalNullableArrayGetOverloadSlow(overload)
	if vm != nil {
		vm.arrayGetOverloadHot = overload
		vm.arrayGetOverloadHotVersion = version
		vm.arrayGetOverloadHotOK = ok
	}
	return ok
}

func bytecodeArrayGetOverloadCallable(callable runtime.Value) *runtime.FunctionOverloadValue {
	switch fn := callable.(type) {
	case *runtime.FunctionOverloadValue:
		return fn
	case runtime.BoundMethodValue:
		if method, ok := fn.Method.(*runtime.FunctionOverloadValue); ok {
			return method
		}
	case *runtime.BoundMethodValue:
		if fn != nil {
			if method, ok := fn.Method.(*runtime.FunctionOverloadValue); ok {
				return method
			}
		}
	}
	return nil
}

func (vm *bytecodeVM) isCanonicalNullableArrayGetOverloadSlow(overload *runtime.FunctionOverloadValue) bool {
	if overload == nil || len(overload.Overloads) != 2 {
		return false
	}

	nullableCount := 0
	resultCount := 0
	for _, fn := range overload.Overloads {
		if !vm.isCanonicalArrayGetOverloadFunction(fn) {
			return false
		}
		def := fn.Declaration.(*ast.FunctionDefinition)
		switch def.ReturnType.(type) {
		case *ast.NullableTypeExpression:
			if fn.MethodPriority < 0 {
				return false
			}
			nullableCount++
		case *ast.ResultTypeExpression:
			if fn.MethodPriority >= 0 {
				return false
			}
			resultCount++
		default:
			return false
		}
	}
	return nullableCount == 1 && resultCount == 1
}

func (vm *bytecodeVM) isCanonicalArrayGetOverloadFunction(fn *runtime.FunctionValue) bool {
	if vm == nil || vm.interp == nil || fn == nil {
		return false
	}
	def, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || def == nil || def.ID == nil || def.ID.Name != "get" {
		return false
	}
	if len(def.Params) != 2 || !bytecodeArrayGetParamIsI32(def.Params[1]) {
		return false
	}
	origin := vm.interp.nodeOrigins[def]
	return isCanonicalAbleStdlibOrigin(origin, "collections/array.able")
}

func bytecodeArrayGetParamIsI32(param *ast.FunctionParameter) bool {
	if param == nil {
		return false
	}
	simple, ok := param.ParamType.(*ast.SimpleTypeExpression)
	return ok && simple != nil && simple.Name != nil && simple.Name.Name == "i32"
}
