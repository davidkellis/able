package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type functionCallGenericPlan struct {
	expectedCount           int
	namesByIndex            []string
	genericNames            map[string]struct{}
	genericIndex            map[string]int
	constraints             []constraintSpec
	functionName            string
	callingCtx              string
	skipLeadingRuntimeArgs  int
	inferenceRelevantParams []functionCallGenericInferenceParam
}

type methodSetConstraintPlan struct {
	genericNames map[string]struct{}
	constraints  []constraintSpec
}

type functionCallGenericInferenceParam struct {
	argIndex  int
	paramType ast.TypeExpression
}

func buildGenericParamNamesByIndex(generics []*ast.GenericParameter) []string {
	if len(generics) == 0 {
		return nil
	}
	names := make([]string, len(generics))
	for idx, gp := range generics {
		if gp == nil || gp.Name == nil {
			continue
		}
		names[idx] = gp.Name.Name
	}
	return names
}

func buildGenericParamIndex(namesByIndex []string) map[string]int {
	if len(namesByIndex) == 0 {
		return nil
	}
	index := make(map[string]int, len(namesByIndex))
	for idx, name := range namesByIndex {
		if name == "" {
			continue
		}
		index[name] = idx
	}
	return index
}

func buildFunctionCallGenericPlan(funcNode ast.Node) *functionCallGenericPlan {
	generics, whereClause := extractFunctionGenerics(funcNode)
	name := functionNameForErrors(funcNode)
	genericNames := genericNameSet(generics)
	namesByIndex := buildGenericParamNamesByIndex(generics)
	params := extractFunctionParams(funcNode)
	skipLeadingRuntimeArgs := 0
	if def, ok := funcNode.(*ast.FunctionDefinition); ok && def.IsMethodShorthand {
		skipLeadingRuntimeArgs = 1
	}
	return &functionCallGenericPlan{
		expectedCount:           len(generics),
		namesByIndex:            namesByIndex,
		genericNames:            genericNames,
		genericIndex:            buildGenericParamIndex(namesByIndex),
		constraints:             collectConstraintSpecs(generics, whereClause),
		functionName:            name,
		callingCtx:              "calling " + name,
		skipLeadingRuntimeArgs:  skipLeadingRuntimeArgs,
		inferenceRelevantParams: buildFunctionCallGenericInferenceParams(params, genericNames),
	}
}

func buildFunctionCallGenericInferenceParams(params []*ast.FunctionParameter, genericNames map[string]struct{}) []functionCallGenericInferenceParam {
	if len(params) == 0 || len(genericNames) == 0 {
		return nil
	}
	relevant := make([]functionCallGenericInferenceParam, 0, len(params))
	for idx, param := range params {
		if param == nil || param.ParamType == nil {
			continue
		}
		if !typeExpressionUsesGenerics(param.ParamType, genericNames) {
			continue
		}
		relevant = append(relevant, functionCallGenericInferenceParam{
			argIndex:  idx,
			paramType: param.ParamType,
		})
	}
	if len(relevant) == 0 {
		return nil
	}
	return relevant
}

func buildMethodSetConstraintPlan(methodSet *runtime.MethodSet) *methodSetConstraintPlan {
	if methodSet == nil {
		return &methodSetConstraintPlan{}
	}
	return &methodSetConstraintPlan{
		genericNames: genericNameSet(methodSet.GenericParams),
		constraints:  collectConstraintSpecs(methodSet.GenericParams, methodSet.WhereClause),
	}
}

func hasAnyGenericConstraints(generics []*ast.GenericParameter, whereClause []*ast.WhereClauseConstraint) bool {
	for _, gp := range generics {
		if gp == nil {
			continue
		}
		for _, constraint := range gp.Constraints {
			if constraint != nil && constraint.InterfaceType != nil {
				return true
			}
		}
	}
	for _, clause := range whereClause {
		if clause == nil {
			continue
		}
		for _, constraint := range clause.Constraints {
			if constraint != nil && constraint.InterfaceType != nil {
				return true
			}
		}
	}
	return false
}

func (i *Interpreter) functionCallGenericPlan(funcNode ast.Node) *functionCallGenericPlan {
	if i == nil || funcNode == nil {
		return nil
	}
	if i.envSingleThread {
		if plan, ok := i.functionCallGenericPlanCache[funcNode]; ok {
			return plan
		}
		plan := buildFunctionCallGenericPlan(funcNode)
		i.functionCallGenericPlanCache[funcNode] = plan
		return plan
	}
	i.functionCallGenericPlanCacheMu.RLock()
	plan, ok := i.functionCallGenericPlanCache[funcNode]
	i.functionCallGenericPlanCacheMu.RUnlock()
	if ok {
		return plan
	}
	plan = buildFunctionCallGenericPlan(funcNode)
	i.functionCallGenericPlanCacheMu.Lock()
	if existing, ok := i.functionCallGenericPlanCache[funcNode]; ok {
		i.functionCallGenericPlanCacheMu.Unlock()
		return existing
	}
	i.functionCallGenericPlanCache[funcNode] = plan
	i.functionCallGenericPlanCacheMu.Unlock()
	return plan
}

func (i *Interpreter) methodSetConstraintPlan(methodSet *runtime.MethodSet) *methodSetConstraintPlan {
	if i == nil || methodSet == nil {
		return nil
	}
	if i.envSingleThread {
		if plan, ok := i.methodSetConstraintPlanCache[methodSet]; ok {
			return plan
		}
		plan := buildMethodSetConstraintPlan(methodSet)
		i.methodSetConstraintPlanCache[methodSet] = plan
		return plan
	}
	i.methodSetConstraintPlanCacheMu.RLock()
	plan, ok := i.methodSetConstraintPlanCache[methodSet]
	i.methodSetConstraintPlanCacheMu.RUnlock()
	if ok {
		return plan
	}
	plan = buildMethodSetConstraintPlan(methodSet)
	i.methodSetConstraintPlanCacheMu.Lock()
	if existing, ok := i.methodSetConstraintPlanCache[methodSet]; ok {
		i.methodSetConstraintPlanCacheMu.Unlock()
		return existing
	}
	i.methodSetConstraintPlanCache[methodSet] = plan
	i.methodSetConstraintPlanCacheMu.Unlock()
	return plan
}
