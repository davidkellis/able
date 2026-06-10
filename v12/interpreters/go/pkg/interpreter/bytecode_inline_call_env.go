package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) inlineResolvedCallEnvForBindings(fn *runtime.FunctionValue, prog *bytecodeProgram, layout *bytecodeFrameLayout, injectedReceiver runtime.Value, hasInjectedReceiver bool, argBase int, argCount int, callNode *ast.FunctionCall) (*runtime.Environment, error) {
	if vm == nil || vm.interp == nil || fn == nil {
		return nil, nil
	}
	decl, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || decl == nil {
		return fn.Closure, nil
	}
	callPlan := vm.interp.functionRuntimeGenericBindingPlan(fn)
	needsCallLocalTypeBindings := callPlan != nil && callPlan.callLocalUsed
	needsConstraintCheck := callNode != nil && len(decl.GenericParams) > 0 && callPlan != nil && callPlan.hasGenericConstraints
	needsTypeArgs := (callNode != nil && len(callNode.TypeArguments) > 0) || needsConstraintCheck
	needsMethodSetConstraintCheck := false
	receiverNeedsSelf := functionDefinitionExpectsSelf(decl)
	receiver, hasReceiver := runtime.Value(nil), false
	if receiverNeedsSelf && (needsCallLocalTypeBindings || fn.MethodSet != nil) && (hasInjectedReceiver || argCount > 0) {
		receiver = bytecodeResolvedCallArgAt(vm.stack, argBase, 0, injectedReceiver, hasInjectedReceiver)
		hasReceiver = true
	}
	if hasReceiver && fn.MethodSet != nil {
		if plan := vm.interp.methodSetConstraintPlan(fn.MethodSet); plan != nil && len(plan.constraints) > 0 {
			needsMethodSetConstraintCheck = true
		}
	}
	needsReturnBindings := prog != nil && len(bytecodeInlineReturnGenericNames(fn, prog)) > 0
	if !needsCallLocalTypeBindings && !needsConstraintCheck && !needsTypeArgs && !needsMethodSetConstraintCheck {
		return fn.Closure, nil
	}
	if !needsCallLocalTypeBindings && !needsTypeArgs && !needsReturnBindings && !needsMethodSetConstraintCheck {
		return fn.Closure, nil
	}
	if needsTypeArgs {
		if err := vm.interp.populateCallTypeArgumentsFromBytecodeResolvedCallArgs(decl, callNode, vm.stack, argBase, argCount, injectedReceiver, hasInjectedReceiver); err != nil {
			return nil, err
		}
	}
	if needsMethodSetConstraintCheck {
		if err := vm.interp.enforceMethodSetConstraints(fn, receiver); err != nil {
			return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
		}
	}
	if needsConstraintCheck {
		if err := vm.interp.enforceGenericConstraintsIfAny(decl, callNode); err != nil {
			return nil, err
		}
	}
	callTypeBindings := vm.interp.functionCallTypeBindingSetWithPlanAndEnv(fn, decl, callNode, receiver, hasReceiver && needsCallLocalTypeBindings, callPlan, vm.env)
	if callTypeBindings.empty() {
		return fn.Closure, nil
	}
	if reusableEnv, ok := vm.interp.reusableBytecodeCallEnvForResolvedBindings(fn, decl, callNode, prog, callTypeBindings); ok {
		return reusableEnv, nil
	}
	return runtime.NewEnvironmentWithBindingSets(
		fn.Closure,
		callTypeBindings.envValueCapacity(functionLocalBindingCapacityForLayout(decl, callNode, layout)),
		callTypeBindings.explicit,
		callTypeBindings.callLocal,
	), nil
}
