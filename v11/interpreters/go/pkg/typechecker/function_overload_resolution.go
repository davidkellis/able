package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

type overloadCandidate struct {
	fn               FunctionType
	inst             FunctionType
	params           []Type
	optionalLast     bool
	score            float64
	isGeneric        bool
	specificity      int
	inferredTypeArgs []ast.TypeExpression
	diags            []Diagnostic
}

func cloneFunctionCall(call *ast.FunctionCall) *ast.FunctionCall {
	if call == nil {
		return nil
	}
	clone := *call
	if call.Arguments != nil {
		clone.Arguments = append([]ast.Expression(nil), call.Arguments...)
	}
	if call.TypeArguments != nil {
		clone.TypeArguments = append([]ast.TypeExpression(nil), call.TypeArguments...)
	}
	return &clone
}

func hasOptionalLastParam(params []Type) bool {
	if len(params) == 0 {
		return false
	}
	_, ok := params[len(params)-1].(NullableType)
	return ok
}

func arityMatches(params []Type, argCount int, optionalLast bool) bool {
	return len(params) == argCount || (optionalLast && argCount == len(params)-1)
}

func dropOptionalParam(params []Type, argCount int, optionalLast bool) []Type {
	if optionalLast && argCount == len(params)-1 {
		return params[:len(params)-1]
	}
	return params
}

func overloadLabel(call *ast.FunctionCall) string {
	if call == nil {
		return "function"
	}
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		if callee != nil && callee.Name != "" {
			return callee.Name
		}
	case *ast.MemberAccessExpression:
		if callee != nil {
			if member, ok := callee.Member.(*ast.Identifier); ok && member != nil && member.Name != "" {
				return member.Name
			}
		}
	}
	return "function"
}

func (c *Checker) instantiateOverloadCandidate(fn FunctionType, call *ast.FunctionCall, argTypes []Type, expectedReturn Type) (FunctionType, []ast.TypeExpression, []Diagnostic) {
	clone := cloneFunctionCall(call)
	inst, diags := c.instantiateFunctionCall(fn, clone, argTypes, expectedReturn)
	var inferred []ast.TypeExpression
	if call != nil && len(call.TypeArguments) == 0 && clone != nil && len(clone.TypeArguments) > 0 {
		inferred = clone.TypeArguments
	}
	return inst, inferred, diags
}

func (c *Checker) scoreCompatibility(params []Type, argTypes []Type, allowUnknown bool) float64 {
	compareCount := len(argTypes)
	if len(params) < compareCount {
		compareCount = len(params)
	}
	score := 0.0
	for i := 0; i < compareCount; i++ {
		expected := params[i]
		actual := argTypes[i]
		if expected == nil || actual == nil || isUnknownType(expected) || isUnknownType(actual) {
			continue
		}
		if !allowUnknown && containsUnknownType(actual) && !containsUnknownType(expected) {
			return -1
		}
		if _, ok := literalMismatchMessage(actual, expected); ok {
			return -1
		}
		if !typeAssignable(actual, expected) {
			if iface, args, ok := interfaceFromType(expected); ok {
				if okImpl, _ := c.typeImplementsInterface(actual, iface, args); okImpl {
					score += 1
					continue
				}
			}
			if nullable, ok := expected.(NullableType); ok {
				if iface, args, ok := interfaceFromType(nullable.Inner); ok {
					if okImpl, _ := c.typeImplementsInterface(actual, iface, args); okImpl {
						score += 1
						continue
					}
				}
			}
			return -1
		}
		if _, ok := expected.(NullableType); ok {
			score += 1
		} else {
			score += 2
		}
	}
	return score
}

func signatureSpecificityScore(params []Type) int {
	total := 0
	for _, param := range params {
		total += overloadTypeSpecificity(param)
	}
	return total
}

func containsUnknownType(t Type) bool {
	if t == nil {
		return false
	}
	if isUnknownType(t) {
		return true
	}
	switch v := t.(type) {
	case ArrayType:
		return containsUnknownType(v.Element)
	case MapType:
		return containsUnknownType(v.Key) || containsUnknownType(v.Value)
	case RangeType:
		if containsUnknownType(v.Element) {
			return true
		}
		for _, bound := range v.Bounds {
			if containsUnknownType(bound) {
				return true
			}
		}
		return false
	case IteratorType:
		return containsUnknownType(v.Element)
	case NullableType:
		return containsUnknownType(v.Inner)
	case UnionLiteralType:
		for _, member := range v.Members {
			if containsUnknownType(member) {
				return true
			}
		}
		return false
	case UnionType:
		for _, member := range v.Variants {
			if containsUnknownType(member) {
				return true
			}
		}
		return false
	case AppliedType:
		if containsUnknownType(v.Base) {
			return true
		}
		for _, arg := range v.Arguments {
			if containsUnknownType(arg) {
				return true
			}
		}
		return false
	case StructInstanceType:
		for _, arg := range v.TypeArgs {
			if containsUnknownType(arg) {
				return true
			}
		}
		return false
	case FunctionType:
		if containsUnknownType(v.Return) {
			return true
		}
		for _, param := range v.Params {
			if containsUnknownType(param) {
				return true
			}
		}
		return false
	case ProcType:
		return containsUnknownType(v.Result)
	case FutureType:
		return containsUnknownType(v.Result)
	default:
		return false
	}
}

func overloadTypeSpecificity(t Type) int {
	if t == nil || isUnknownType(t) {
		return 0
	}
	switch v := t.(type) {
	case TypeParameterType:
		return 0
	case PrimitiveType, IntegerType, FloatType, StructType, StructInstanceType, InterfaceType, UnionType, PackageType, ImplementationNamespaceType:
		score := 1
		if inst, ok := t.(StructInstanceType); ok {
			for _, arg := range inst.TypeArgs {
				score += overloadTypeSpecificity(arg)
			}
		}
		return score
	case AliasType:
		return 1 + overloadTypeSpecificity(v.Target)
	case AppliedType:
		score := 1 + overloadTypeSpecificity(v.Base)
		for _, arg := range v.Arguments {
			score += overloadTypeSpecificity(arg)
		}
		return score
	case ArrayType:
		return 1 + overloadTypeSpecificity(v.Element)
	case MapType:
		return 1 + overloadTypeSpecificity(v.Key) + overloadTypeSpecificity(v.Value)
	case RangeType:
		score := 1 + overloadTypeSpecificity(v.Element)
		for _, bound := range v.Bounds {
			score += overloadTypeSpecificity(bound)
		}
		return score
	case IteratorType:
		return 1 + overloadTypeSpecificity(v.Element)
	case NullableType:
		return 1 + overloadTypeSpecificity(v.Inner)
	case UnionLiteralType:
		score := 1
		for _, member := range v.Members {
			score += overloadTypeSpecificity(member)
		}
		return score
	case FunctionType:
		score := 1 + overloadTypeSpecificity(v.Return)
		for _, param := range v.Params {
			score += overloadTypeSpecificity(param)
		}
		return score
	case ProcType:
		return 1 + overloadTypeSpecificity(v.Result)
	case FutureType:
		return 1 + overloadTypeSpecificity(v.Result)
	default:
		return 0
	}
}

func compareOverloadCandidates(a, b overloadCandidate) int {
	if a.score > b.score+1e-9 {
		return 1
	}
	if b.score > a.score+1e-9 {
		return -1
	}
	if a.isGeneric != b.isGeneric {
		if !a.isGeneric && b.isGeneric {
			return 1
		}
		return -1
	}
	if a.isGeneric && b.isGeneric {
		if a.specificity > b.specificity {
			return 1
		}
		if a.specificity < b.specificity {
			return -1
		}
	}
	return 0
}

func (c *Checker) selectBestOverload(candidates []overloadCandidate) (overloadCandidate, bool) {
	best := candidates[0]
	ambiguous := false
	for i := 1; i < len(candidates); i++ {
		cand := candidates[i]
		switch compareOverloadCandidates(cand, best) {
		case 1:
			best = cand
			ambiguous = false
		case 0:
			ambiguous = true
		}
	}
	return best, ambiguous
}

func (c *Checker) selectPartialOverload(
	overloads []FunctionType,
	call *ast.FunctionCall,
	argTypes []Type,
	expectedReturn Type,
	explicitTypeArgs int,
) (overloadCandidate, bool) {
	allowUnknown := len(overloads) <= 1
	argCount := len(argTypes)
	var best overloadCandidate
	bestSet := false
	for _, fn := range overloads {
		if explicitTypeArgs > 0 && len(fn.TypeParams) != explicitTypeArgs {
			continue
		}
		inst, inferred, instDiags := c.instantiateOverloadCandidate(fn, call, argTypes, expectedReturn)
		optionalLast := hasOptionalLastParam(inst.Params)
		minArgs := len(inst.Params)
		if optionalLast && minArgs > 0 {
			minArgs = minArgs - 1
		}
		if argCount >= minArgs {
			continue
		}
		prefix := inst.Params
		if argCount < len(prefix) {
			prefix = prefix[:argCount]
		}
		score := c.scoreCompatibility(prefix, argTypes, allowUnknown)
		if score < 0 {
			continue
		}
		remaining := len(inst.Params) - argCount
		if !bestSet || remaining < len(best.inst.Params)-argCount {
			best = overloadCandidate{
				fn:               fn,
				inst:             inst,
				params:           inst.Params,
				optionalLast:     optionalLast,
				score:            score,
				inferredTypeArgs: inferred,
				diags:            instDiags,
			}
			bestSet = true
		}
	}
	return best, bestSet
}

func (c *Checker) applyFunctionCallSignature(
	call *ast.FunctionCall,
	fn FunctionType,
	args []ast.Expression,
	argTypes []Type,
	diags []Diagnostic,
) ([]Diagnostic, Type) {
	expectedParams := fn.Params
	optionalLast := hasOptionalLastParam(expectedParams)
	argMatches := func(actual Type, expected Type) bool {
		if typeAssignable(actual, expected) {
			return true
		}
		if expected != nil {
			if iface, args, ok := interfaceFromType(expected); ok {
				if okImpl, _ := c.typeImplementsInterface(actual, iface, args); okImpl {
					return true
				}
			}
			if nullable, ok := expected.(NullableType); ok {
				if iface, args, ok := interfaceFromType(nullable.Inner); ok {
					if okImpl, _ := c.typeImplementsInterface(actual, iface, args); okImpl {
						return true
					}
				}
			}
		}
		return false
	}
	argCount := len(argTypes)
	paramCount := len(expectedParams)
	minArgs := paramCount
	if optionalLast && paramCount > 0 {
		minArgs = paramCount - 1
	}
	if argCount > paramCount {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: function expects %d arguments, got %d", paramCount, argCount),
			Node:    call,
		})
		return diags, fn.Return
	}
	if argCount < minArgs {
		compareCount := argCount
		for i := 0; i < compareCount; i++ {
			expected := expectedParams[i]
			if !argMatches(argTypes[i], expected) {
				if msg, ok := literalMismatchMessage(argTypes[i], expected); ok {
					diags = append(diags, Diagnostic{
						Message: fmt.Sprintf("typechecker: %s", msg),
						Node:    args[i],
					})
				} else {
					diags = append(diags, Diagnostic{
						Message: fmt.Sprintf("typechecker: argument %d has type %s, expected %s", i+1, typeName(argTypes[i]), typeName(expected)),
						Node:    args[i],
					})
				}
			}
		}
		remaining := expectedParams[argCount:]
		resultType := FunctionType{Params: remaining, Return: fn.Return}
		return diags, resultType
	}
	if optionalLast && argCount == paramCount-1 {
		expectedParams = expectedParams[:len(expectedParams)-1]
		paramCount = len(expectedParams)
	}
	compareCount := len(argTypes)
	if paramCount < compareCount {
		compareCount = paramCount
	}
	for i := 0; i < compareCount; i++ {
		expected := expectedParams[i]
		if !argMatches(argTypes[i], expected) {
			if msg, ok := literalMismatchMessage(argTypes[i], expected); ok {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: %s", msg),
					Node:    args[i],
				})
			} else {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: argument %d has type %s, expected %s", i+1, typeName(argTypes[i]), typeName(expected)),
					Node:    args[i],
				})
			}
		}
	}
	return diags, fn.Return
}

func (c *Checker) checkOverloadedFunctionCall(
	call *ast.FunctionCall,
	overloads []FunctionType,
	args []ast.Expression,
	argTypes []Type,
	expectedReturn Type,
) ([]Diagnostic, Type) {
	argCount := len(argTypes)
	explicitTypeArgs := 0
	if call != nil {
		explicitTypeArgs = len(call.TypeArguments)
	}
	allowUnknown := len(overloads) <= 1
	candidates := make([]overloadCandidate, 0, len(overloads))
	for _, fn := range overloads {
		if explicitTypeArgs > 0 && len(fn.TypeParams) != explicitTypeArgs {
			continue
		}
		inst, inferred, instDiags := c.instantiateOverloadCandidate(fn, call, argTypes, expectedReturn)
		optionalLast := hasOptionalLastParam(inst.Params)
		if !arityMatches(inst.Params, argCount, optionalLast) {
			continue
		}
		params := dropOptionalParam(inst.Params, argCount, optionalLast)
		score := c.scoreCompatibility(params, argTypes, allowUnknown)
		if score < 0 {
			continue
		}
		if optionalLast && len(params) != len(inst.Params) {
			score -= 0.5
		}
		specificity := 0
		isGeneric := len(fn.TypeParams) > 0
		if isGeneric {
			sigParams := dropOptionalParam(fn.Params, argCount, optionalLast)
			specificity = signatureSpecificityScore(sigParams)
		}
		candidates = append(candidates, overloadCandidate{
			fn:               fn,
			inst:             inst,
			params:           params,
			optionalLast:     optionalLast,
			score:            score,
			isGeneric:        isGeneric,
			specificity:      specificity,
			inferredTypeArgs: inferred,
			diags:            instDiags,
		})
	}
	if len(candidates) > 0 {
		best, ambiguous := c.selectBestOverload(candidates)
		if ambiguous {
			return []Diagnostic{{
				Message: fmt.Sprintf("typechecker: ambiguous overload for %s", overloadLabel(call)),
				Node:    call,
			}}, UnknownType{}
		}
		diags := append([]Diagnostic{}, best.diags...)
		if call != nil && len(call.TypeArguments) == 0 && len(best.inferredTypeArgs) > 0 {
			call.TypeArguments = best.inferredTypeArgs
		}
		if len(best.inst.Obligations) > 0 {
			c.obligations = append(c.obligations, best.inst.Obligations...)
		}
		return c.applyFunctionCallSignature(call, best.inst, args, argTypes, diags)
	}
	if partial, ok := c.selectPartialOverload(overloads, call, argTypes, expectedReturn, explicitTypeArgs); ok {
		diags := append([]Diagnostic{}, partial.diags...)
		if call != nil && len(call.TypeArguments) == 0 && len(partial.inferredTypeArgs) > 0 {
			call.TypeArguments = partial.inferredTypeArgs
		}
		if len(partial.inst.Obligations) > 0 {
			c.obligations = append(c.obligations, partial.inst.Obligations...)
		}
		return c.applyFunctionCallSignature(call, partial.inst, args, argTypes, diags)
	}
	return []Diagnostic{{
		Message: fmt.Sprintf("typechecker: no overloads of %s match provided arguments", overloadLabel(call)),
		Node:    call,
	}}, UnknownType{}
}
