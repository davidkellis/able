package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func completeMonomorphicFunctionType(fn FunctionType) bool {
	return len(fn.TypeParams) == 0 &&
		len(fn.Where) == 0 &&
		exactTypeEquivalent(fn, fn)
}

func (c *Checker) constrainLocalLambdaArgument(
	env *Environment,
	identifier *ast.Identifier,
	expected FunctionType,
) ([]Diagnostic, Type, bool) {
	if env == nil || identifier == nil || identifier.Name == "" ||
		!completeMonomorphicFunctionType(expected) {
		return nil, nil, false
	}
	binding, ok := env.lookupLocalLambda(identifier.Name)
	if !ok || binding == nil {
		return nil, nil, false
	}
	diags, constrained := c.constrainLocalLambdaBinding(identifier.Name, binding, expected, identifier)
	return diags, constrained, true
}

func (c *Checker) inferLocalLambdaInvocation(
	env *Environment,
	identifier *ast.Identifier,
	arguments []ast.Expression,
	argumentTypes []Type,
	expectedReturn Type,
) ([]Diagnostic, Type, bool) {
	if env == nil || identifier == nil || identifier.Name == "" {
		return nil, nil, false
	}
	binding, ok := env.lookupLocalLambda(identifier.Name)
	if !ok || binding == nil || binding.expression == nil ||
		len(arguments) != len(binding.expression.Params) ||
		len(argumentTypes) != len(binding.expression.Params) {
		return nil, nil, false
	}
	for _, argumentType := range argumentTypes {
		if argumentType == nil || !exactTypeEquivalent(argumentType, argumentType) {
			return nil, nil, false
		}
	}
	provisional := FunctionType{
		Params: argumentTypes,
		Return: expectedReturn,
	}
	lambdaDiags, inferred := c.checkLambdaExpressionWithExpectedType(
		binding.declarationEnv,
		binding.expression,
		&provisional,
	)
	inferredFunction, ok := inferred.(FunctionType)
	if !ok || !completeMonomorphicFunctionType(inferredFunction) {
		return lambdaDiags, nil, false
	}
	if binding.signature == nil {
		signature := inferredFunction
		binding.signature = &signature
		binding.constraintNode = identifier
		binding.declarationEnv.Assign(identifier.Name, signature)
		return lambdaDiags, signature, true
	}
	constraintDiags, constrained := c.constrainLocalLambdaBinding(
		identifier.Name,
		binding,
		inferredFunction,
		identifier,
	)
	lambdaDiags = append(lambdaDiags, constraintDiags...)
	return lambdaDiags, constrained, true
}

func (c *Checker) constrainLocalLambdaBinding(
	name string,
	binding *localLambdaBinding,
	expected FunctionType,
	node ast.Node,
) ([]Diagnostic, Type) {
	if binding == nil {
		return nil, expected
	}
	if binding.signature != nil {
		if exactTypeEquivalent(*binding.signature, expected) {
			return nil, *binding.signature
		}
		diagnostic := Diagnostic{
			Severity: SeverityError,
			Code:     DiagnosticCodeCallableSignatureMismatch,
			Message: fmt.Sprintf(
				"typechecker: local lambda '%s' has conflicting callable constraints %s and %s",
				name,
				formatTypeForReturnDiagnostic(*binding.signature),
				formatTypeForReturnDiagnostic(expected),
			),
			Node: node,
		}
		if binding.constraintNode != nil {
			diagnostic.Notes = append(diagnostic.Notes, DiagnosticNote{
				Message: "first static callable constraint established here",
				Node:    binding.constraintNode,
			})
		}
		return []Diagnostic{diagnostic}, expected
	}

	lambdaDiags, inferred := c.checkLambdaExpressionWithExpectedType(
		binding.declarationEnv,
		binding.expression,
		&expected,
	)
	inferredFunction, ok := inferred.(FunctionType)
	if !ok || !completeMonomorphicFunctionType(inferredFunction) {
		return lambdaDiags, inferred
	}
	signature := inferredFunction
	binding.signature = &signature
	binding.constraintNode = node
	binding.declarationEnv.Assign(name, signature)
	return lambdaDiags, signature
}
