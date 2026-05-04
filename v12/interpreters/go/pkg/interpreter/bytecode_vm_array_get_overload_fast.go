package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

const bytecodeArrayGetCallHotEntries = 4

type bytecodeArrayGetCallCacheEntry struct {
	env                *runtime.Environment
	envVersion         uint64
	globalRevision     uint64
	methodCacheVersion uint64
}

type bytecodeInlineArrayGetCallCacheEntry struct {
	valid              bool
	program            *bytecodeProgram
	ip                 int
	env                *runtime.Environment
	envVersion         uint64
	globalRevision     uint64
	methodCacheVersion uint64
}

func (vm *bytecodeVM) execCanonicalArrayGetOverloadMemberFast(callable runtime.Value, instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if instr.name != "get" || instr.argCount != 1 {
		return nil, false, nil
	}
	if !vm.isCanonicalNullableArrayGetOverload(callable) {
		return nil, false, nil
	}
	return vm.execArrayGetMemberFast(instr, receiverIndex, argBase, callNode)
}

func (vm *bytecodeVM) canUseCanonicalArrayGetCallCache(instr bytecodeInstruction, receiver runtime.Value) bool {
	if vm == nil || vm.interp == nil || vm.env == nil || instr.name != "get" || instr.argCount != 1 {
		return false
	}
	arr, ok := receiver.(*runtime.ArrayValue)
	if !ok || arr == nil {
		return false
	}
	return vm.env.RuntimeData() == nil
}

func (entry bytecodeInlineArrayGetCallCacheEntry) matchesCanonicalArrayGetCallIdentity(program *bytecodeProgram, ip int, env *runtime.Environment) bool {
	return entry.valid &&
		entry.program == program &&
		entry.ip == ip &&
		entry.env == env
}

func (entry bytecodeInlineArrayGetCallCacheEntry) matchesCanonicalArrayGetCallVersions(envVersion uint64, globalRev uint64, methodVersion uint64) bool {
	return entry.envVersion == envVersion &&
		entry.globalRevision == globalRev &&
		entry.methodCacheVersion == methodVersion
}

func (vm *bytecodeVM) promoteCanonicalArrayGetCallHot(entry bytecodeInlineArrayGetCallCacheEntry) {
	if vm == nil || !entry.valid {
		return
	}
	for i := 0; i < len(vm.arrayGetCallHot); i++ {
		if vm.arrayGetCallHot[i].matchesCanonicalArrayGetCallIdentity(entry.program, entry.ip, entry.env) {
			copy(vm.arrayGetCallHot[1:i+1], vm.arrayGetCallHot[0:i])
			vm.arrayGetCallHot[0] = entry
			return
		}
	}
	copy(vm.arrayGetCallHot[1:], vm.arrayGetCallHot[:len(vm.arrayGetCallHot)-1])
	vm.arrayGetCallHot[0] = entry
}

func (vm *bytecodeVM) canonicalArrayGetCallVersions(env *runtime.Environment) (uint64, uint64, uint64) {
	return vm.bytecodeEnvRevision(env), vm.bytecodeGlobalRevision(), vm.bytecodeMethodCacheVersion()
}

func (vm *bytecodeVM) lookupCachedCanonicalArrayGetCall(program *bytecodeProgram, ip int, instr bytecodeInstruction, receiver runtime.Value) bool {
	if program == nil || !vm.canUseCanonicalArrayGetCallCache(instr, receiver) {
		return false
	}
	env := vm.env
	var (
		envVersion    uint64
		globalRev     uint64
		methodVersion uint64
		haveVersions  bool
	)
	for i := 0; i < len(vm.arrayGetCallHot); i++ {
		hot := vm.arrayGetCallHot[i]
		if !hot.matchesCanonicalArrayGetCallIdentity(program, ip, env) {
			continue
		}
		if !haveVersions {
			envVersion, globalRev, methodVersion = vm.canonicalArrayGetCallVersions(env)
			haveVersions = true
		}
		if !hot.matchesCanonicalArrayGetCallVersions(envVersion, globalRev, methodVersion) {
			return false
		}
		return true
	}
	if vm.arrayGetCallCache == nil {
		return false
	}
	if !haveVersions {
		envVersion, globalRev, methodVersion = vm.canonicalArrayGetCallVersions(env)
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	entry, ok := vm.arrayGetCallCache[key]
	if !ok ||
		entry.env != env ||
		entry.envVersion != envVersion ||
		entry.globalRevision != globalRev ||
		entry.methodCacheVersion != methodVersion {
		return false
	}
	vm.promoteCanonicalArrayGetCallHot(bytecodeInlineArrayGetCallCacheEntry{
		valid:              true,
		program:            program,
		ip:                 ip,
		env:                entry.env,
		envVersion:         entry.envVersion,
		globalRevision:     entry.globalRevision,
		methodCacheVersion: entry.methodCacheVersion,
	})
	return true
}

func (vm *bytecodeVM) storeCachedCanonicalArrayGetCall(program *bytecodeProgram, ip int, instr bytecodeInstruction, receiver runtime.Value) {
	if program == nil || !vm.canUseCanonicalArrayGetCallCache(instr, receiver) {
		return
	}
	envVersion, globalRev, methodVersion := vm.canonicalArrayGetCallVersions(vm.env)
	entry := bytecodeArrayGetCallCacheEntry{
		env:                vm.env,
		envVersion:         envVersion,
		globalRevision:     globalRev,
		methodCacheVersion: methodVersion,
	}
	if vm.arrayGetCallCache == nil {
		vm.arrayGetCallCache = make(map[bytecodeGlobalLookupCacheKey]bytecodeArrayGetCallCacheEntry, 8)
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	vm.arrayGetCallCache[key] = entry
	vm.promoteCanonicalArrayGetCallHot(bytecodeInlineArrayGetCallCacheEntry{
		valid:              true,
		program:            program,
		ip:                 ip,
		env:                entry.env,
		envVersion:         entry.envVersion,
		globalRevision:     entry.globalRevision,
		methodCacheVersion: entry.methodCacheVersion,
	})
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
	nullableFn, resultFn, hasPair := bytecodeArrayGetOverloadFunctionPair(overload)
	if vm != nil && hasPair &&
		vm.arrayGetOverloadPairNullable == nullableFn &&
		vm.arrayGetOverloadPairResult == resultFn &&
		vm.arrayGetOverloadPairVersion == version {
		vm.arrayGetOverloadHot = overload
		vm.arrayGetOverloadHotVersion = version
		vm.arrayGetOverloadHotOK = vm.arrayGetOverloadPairOK
		return vm.arrayGetOverloadPairOK
	}
	ok := vm.isCanonicalNullableArrayGetOverloadSlow(overload)
	if vm != nil {
		vm.arrayGetOverloadHot = overload
		vm.arrayGetOverloadHotVersion = version
		vm.arrayGetOverloadHotOK = ok
		if hasPair {
			vm.arrayGetOverloadPairNullable = nullableFn
			vm.arrayGetOverloadPairResult = resultFn
			vm.arrayGetOverloadPairVersion = version
			vm.arrayGetOverloadPairOK = ok
		}
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

func bytecodeArrayGetOverloadFunctionPair(overload *runtime.FunctionOverloadValue) (*runtime.FunctionValue, *runtime.FunctionValue, bool) {
	if overload == nil || len(overload.Overloads) != 2 {
		return nil, nil, false
	}
	var nullableFn *runtime.FunctionValue
	var resultFn *runtime.FunctionValue
	for _, fn := range overload.Overloads {
		def, ok := bytecodeArrayGetOverloadFunctionDefinition(fn)
		if !ok {
			return nil, nil, false
		}
		switch def.ReturnType.(type) {
		case *ast.NullableTypeExpression:
			if fn.MethodPriority < 0 || nullableFn != nil {
				return nil, nil, false
			}
			nullableFn = fn
		case *ast.ResultTypeExpression:
			if fn.MethodPriority >= 0 || resultFn != nil {
				return nil, nil, false
			}
			resultFn = fn
		default:
			return nil, nil, false
		}
	}
	return nullableFn, resultFn, nullableFn != nil && resultFn != nil
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

func bytecodeArrayGetOverloadFunctionDefinition(fn *runtime.FunctionValue) (*ast.FunctionDefinition, bool) {
	if fn == nil {
		return nil, false
	}
	def, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || def == nil || def.ID == nil || def.ID.Name != "get" {
		return nil, false
	}
	if len(def.Params) != 2 || !bytecodeArrayGetParamIsI32(def.Params[1]) {
		return nil, false
	}
	return def, true
}

func (vm *bytecodeVM) isCanonicalArrayGetOverloadFunction(fn *runtime.FunctionValue) bool {
	if vm == nil || vm.interp == nil || fn == nil {
		return false
	}
	def, ok := bytecodeArrayGetOverloadFunctionDefinition(fn)
	if !ok {
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
