package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) populateCallTypeArguments(funcNode ast.Node, call *ast.FunctionCall, args []runtime.Value) error {
	if funcNode == nil || call == nil {
		return nil
	}
	plan := i.functionCallGenericPlan(funcNode)
	if plan == nil || plan.expectedCount == 0 {
		return nil
	}
	if len(call.TypeArguments) > 0 {
		if !i.callHasExplicitTypeArguments(call) {
			goto infer
		}
		if len(call.TypeArguments) != plan.expectedCount {
			return fmt.Errorf("Type arguments count mismatch calling %s: expected %d, got %d", plan.functionName, plan.expectedCount, len(call.TypeArguments))
		}
		return nil
	}
infer:
	bindArgs := args
	if plan.skipLeadingRuntimeArgs > 0 {
		if len(bindArgs) > plan.skipLeadingRuntimeArgs {
			bindArgs = bindArgs[plan.skipLeadingRuntimeArgs:]
		} else {
			bindArgs = nil
		}
	}
	max := 0
	for _, param := range plan.inferenceRelevantParams {
		if param.argIndex >= len(bindArgs) {
			break
		}
		max++
	}
	i.inferAndSetCallTypeArgumentsFromValues(plan, funcNode, call, max, func(idx int) runtime.Value {
		return bindArgs[plan.inferenceRelevantParams[idx].argIndex]
	})
	return nil
}

func (i *Interpreter) enforceGenericConstraintsIfAny(funcNode ast.Node, call *ast.FunctionCall) error {
	if funcNode == nil || call == nil {
		return nil
	}
	plan := i.functionCallGenericPlan(funcNode)
	if plan == nil || plan.expectedCount == 0 {
		return nil
	}
	if len(call.TypeArguments) != plan.expectedCount {
		return fmt.Errorf("Type arguments count mismatch calling %s: expected %d, got %d", plan.functionName, plan.expectedCount, len(call.TypeArguments))
	}
	if len(plan.constraints) == 0 {
		return nil
	}
	cacheKey := functionCallConstraintResultCacheKey{
		function: funcNode,
		call:     call,
		version:  i.inferredCallTypeArgumentVersion(call),
	}
	if cached, ok := i.lookupFunctionCallConstraintResultCache(cacheKey); ok {
		return cached.err
	}
	err := i.enforceConstraintSpecsWithTypeArgs(plan.constraints, plan.namesByIndex, plan.expectedCount, call.TypeArguments, plan.callingCtx)
	i.storeFunctionCallConstraintResultCache(cacheKey, err)
	return err
}

func (i *Interpreter) bindTypeArgumentsIfAny(funcNode ast.Node, call *ast.FunctionCall, env *runtime.Environment) {
	if env == nil {
		return
	}
	env.DefineWithoutMergeBindings(i.explicitCallTypeBindingValuesIfAny(funcNode, call))
}
