package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

type lambdaConstraint struct {
	GenericName string
	Interface   ast.TypeExpression
}

func (g *generator) lambdaConstraintLines(expr *ast.LambdaExpression, valueVars map[string]string, zeroExpr string) ([]string, bool) {
	if g == nil || expr == nil {
		return nil, true
	}
	constraints := collectLambdaConstraints(expr)
	if len(constraints) == 0 {
		return nil, true
	}
	lines := []string{}
	addRuntimeCheck := func() {
		if len(lines) == 0 {
			lines = append(lines, fmt.Sprintf("if __able_runtime == nil { return %s, __able_control_from_error(fmt.Errorf(\"compiler: missing runtime\")) }", zeroExpr))
		}
	}
	for _, constraint := range constraints {
		valueVar, ok := valueVars[constraint.GenericName]
		if !ok || valueVar == "" {
			continue
		}
		typeExpr, ok := g.renderTypeExpression(constraint.Interface)
		if !ok {
			return nil, false
		}
		addRuntimeCheck()
		lines = append(lines, fmt.Sprintf("if _, ok, err := bridge.MatchType(__able_runtime, %s, %s); err != nil { return %s, __able_control_from_error(err) } else if !ok { return %s, __able_control_from_error(fmt.Errorf(\"Type argument %s does not satisfy constraint\")) }", typeExpr, valueVar, zeroExpr, zeroExpr, constraint.GenericName))
	}
	return lines, true
}

func collectLambdaConstraints(expr *ast.LambdaExpression) []lambdaConstraint {
	if expr == nil {
		return nil
	}
	constraints := []lambdaConstraint{}
	for _, gp := range expr.GenericParams {
		if gp == nil || gp.Name == nil {
			continue
		}
		for _, constraint := range gp.Constraints {
			if constraint == nil || constraint.InterfaceType == nil {
				continue
			}
			constraints = append(constraints, lambdaConstraint{
				GenericName: gp.Name.Name,
				Interface:   constraint.InterfaceType,
			})
		}
	}
	for _, clause := range expr.WhereClause {
		if clause == nil || clause.TypeParam == nil {
			continue
		}
		typeParam, ok := clause.TypeParam.(*ast.SimpleTypeExpression)
		if !ok || typeParam == nil || typeParam.Name == nil {
			continue
		}
		name := typeParam.Name.Name
		for _, constraint := range clause.Constraints {
			if constraint == nil || constraint.InterfaceType == nil {
				continue
			}
			constraints = append(constraints, lambdaConstraint{
				GenericName: name,
				Interface:   constraint.InterfaceType,
			})
		}
	}
	return constraints
}
