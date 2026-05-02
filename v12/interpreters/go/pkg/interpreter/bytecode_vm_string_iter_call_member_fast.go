package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execCanonicalStringByteIteratorNextCallMemberFast(instr bytecodeInstruction, receiverIndex int, callNode *ast.FunctionCall) (*bytecodeProgram, bool, error) {
	if vm == nil || vm.interp == nil || instr.name != "next" || instr.argCount != 0 || receiverIndex < 0 || receiverIndex >= len(vm.stack) {
		return nil, false, nil
	}
	if !vm.isCanonicalStringByteIteratorInterfaceReceiver(vm.stack[receiverIndex]) {
		return nil, false, nil
	}
	if _, ok, err := vm.canonicalStringBytesIteratorNextMethod(); err != nil || !ok {
		return nil, err != nil, err
	}
	return vm.execStringByteIteratorNextMemberFast(instr, receiverIndex, callNode)
}

func (vm *bytecodeVM) isCanonicalStringByteIteratorInterfaceReceiver(value runtime.Value) bool {
	switch iface := value.(type) {
	case *runtime.InterfaceValue:
		return vm.isCanonicalStringByteIteratorInterfaceValue(iface)
	case runtime.InterfaceValue:
		return vm.isCanonicalStringByteIteratorInterfaceValue(&iface)
	default:
		return false
	}
}

func (vm *bytecodeVM) isCanonicalStringByteIteratorInterfaceValue(iface *runtime.InterfaceValue) bool {
	if iface == nil || !vm.isCanonicalIteratorU8Interface(iface) {
		return false
	}
	inst, ok := iface.Underlying.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return false
	}
	return vm.isCanonicalRawStringByteIteratorInstance(inst)
}

func (vm *bytecodeVM) isCanonicalIteratorU8Interface(iface *runtime.InterfaceValue) bool {
	if iface == nil {
		return false
	}
	def, ok := vm.canonicalIteratorInterfaceDefinition()
	if !ok || iface.Interface != def {
		return false
	}
	return bytecodeIsCanonicalU8InterfaceArgs(iface.InterfaceArgs)
}

func (vm *bytecodeVM) isCanonicalRawStringByteIteratorInstance(inst *runtime.StructInstanceValue) bool {
	if inst == nil {
		return false
	}
	def, ok := vm.canonicalStringBytesIteratorDefinition()
	return ok && inst.Definition == def
}

func bytecodeIsCanonicalU8InterfaceArgs(args []ast.TypeExpression) bool {
	return len(args) == 1 && args[0] == bytecodeStringBytesIteratorTypeArgs[0]
}
