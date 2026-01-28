package interpreter

import (
	"errors"
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execDefineUnion(instr bytecodeInstruction) error {
	def, ok := instr.node.(*ast.UnionDefinition)
	if !ok || def == nil {
		return fmt.Errorf("bytecode define expects union definition")
	}
	val, err := vm.interp.evaluateUnionDefinition(def, vm.env)
	if err != nil {
		return vm.interp.attachRuntimeContext(err, def, vm.interp.stateFromEnv(vm.env))
	}
	if val == nil {
		val = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, val)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execDefineTypeAlias(instr bytecodeInstruction) error {
	def, ok := instr.node.(*ast.TypeAliasDefinition)
	if !ok || def == nil {
		return fmt.Errorf("bytecode define expects type alias definition")
	}
	if def.ID != nil {
		if def.ID.Name == "_" {
			err := errors.New("type alias name '_' is reserved")
			return vm.interp.attachRuntimeContext(err, def, vm.interp.stateFromEnv(vm.env))
		}
		alias := def
		if def.TargetType != nil {
			canonicalTarget := canonicalizeTypeExpression(def.TargetType, vm.env, vm.interp.typeAliases)
			clone := *def
			clone.TargetType = canonicalTarget
			alias = &clone
		}
		vm.interp.typeAliases[def.ID.Name] = alias
	}
	vm.stack = append(vm.stack, runtime.NilValue{})
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execDefineMethods(instr bytecodeInstruction) error {
	def, ok := instr.node.(*ast.MethodsDefinition)
	if !ok || def == nil {
		return fmt.Errorf("bytecode define expects methods definition")
	}
	val, err := vm.interp.evaluateMethodsDefinition(def, vm.env)
	if err != nil {
		return vm.interp.attachRuntimeContext(err, def, vm.interp.stateFromEnv(vm.env))
	}
	if val == nil {
		val = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, val)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execDefineInterface(instr bytecodeInstruction) error {
	def, ok := instr.node.(*ast.InterfaceDefinition)
	if !ok || def == nil {
		return fmt.Errorf("bytecode define expects interface definition")
	}
	val, err := vm.interp.evaluateInterfaceDefinition(def, vm.env)
	if err != nil {
		return vm.interp.attachRuntimeContext(err, def, vm.interp.stateFromEnv(vm.env))
	}
	if val == nil {
		val = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, val)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execDefineImplementation(instr bytecodeInstruction) error {
	def, ok := instr.node.(*ast.ImplementationDefinition)
	if !ok || def == nil {
		return fmt.Errorf("bytecode define expects implementation definition")
	}
	val, err := vm.interp.evaluateImplementationDefinition(def, vm.env, false)
	if err != nil {
		return vm.interp.attachRuntimeContext(err, def, vm.interp.stateFromEnv(vm.env))
	}
	if val == nil {
		val = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, val)
	vm.ip++
	return nil
}

func (vm *bytecodeVM) execDefineExtern(instr bytecodeInstruction) error {
	def, ok := instr.node.(*ast.ExternFunctionBody)
	if !ok || def == nil {
		return fmt.Errorf("bytecode define expects extern function body")
	}
	val, err := vm.interp.evaluateExternFunctionBody(def, vm.env)
	if err != nil {
		return vm.interp.attachRuntimeContext(err, def, vm.interp.stateFromEnv(vm.env))
	}
	if val == nil {
		val = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, val)
	vm.ip++
	return nil
}
