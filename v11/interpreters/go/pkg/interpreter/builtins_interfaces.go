package interpreter

import (
	"able/interpreter-go/pkg/ast"
)

type interfaceBundle struct {
	interfaces      []*ast.InterfaceDefinition
	implementations []*ast.ImplementationDefinition
}

func (i *Interpreter) initInterfaceBuiltins() {
	if i.interfaceBuiltinsReady {
		return
	}
	bundle := buildInterfaceBundle()
	for _, iface := range bundle.interfaces {
		if iface == nil {
			continue
		}
		_, _ = i.evaluateInterfaceDefinition(iface, i.global)
	}
	for _, impl := range bundle.implementations {
		if impl == nil {
			continue
		}
		_, _ = i.evaluateImplementationDefinition(impl, i.global, true)
	}
	i.interfaceBuiltinsReady = true
}

func buildInterfaceBundle() interfaceBundle {
	displaySig := ast.NewFunctionSignature(
		ast.NewIdentifier("to_string"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.NewIdentifier("self"), ast.NewSimpleTypeExpression(ast.NewIdentifier("Self"))),
		},
		ast.NewSimpleTypeExpression(ast.NewIdentifier("String")),
		nil,
		nil,
		nil,
	)
	cloneSig := ast.NewFunctionSignature(
		ast.NewIdentifier("clone"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.NewIdentifier("self"), ast.NewSimpleTypeExpression(ast.NewIdentifier("Self"))),
		},
		ast.NewSimpleTypeExpression(ast.NewIdentifier("Self")),
		nil,
		nil,
		nil,
	)
	errorMessageSig := ast.NewFunctionSignature(
		ast.NewIdentifier("message"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.NewIdentifier("self"), ast.NewSimpleTypeExpression(ast.NewIdentifier("Self"))),
		},
		ast.NewSimpleTypeExpression(ast.NewIdentifier("String")),
		nil,
		nil,
		nil,
	)
	errorCauseSig := ast.NewFunctionSignature(
		ast.NewIdentifier("cause"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.NewIdentifier("self"), ast.NewSimpleTypeExpression(ast.NewIdentifier("Self"))),
		},
		ast.NewNullableTypeExpression(ast.NewSimpleTypeExpression(ast.NewIdentifier("Error"))),
		nil,
		nil,
		nil,
	)

	displayDef := ast.NewInterfaceDefinition(ast.NewIdentifier("Display"), []*ast.FunctionSignature{displaySig}, nil, nil, nil, nil, false)
	cloneDef := ast.NewInterfaceDefinition(ast.NewIdentifier("Clone"), []*ast.FunctionSignature{cloneSig}, nil, nil, nil, nil, false)
	errorDef := ast.NewInterfaceDefinition(ast.NewIdentifier("Error"), []*ast.FunctionSignature{errorMessageSig, errorCauseSig}, nil, nil, nil, nil, false)

	var implementations []*ast.ImplementationDefinition
	for _, typeName := range []string{"String", "i32", "bool", "char", "f64"} {
		implementations = append(implementations, makeDisplayImpl(typeName))
		implementations = append(implementations, makeCloneImpl(typeName))
	}
	implementations = append(implementations, makeErrorImpl())

	return interfaceBundle{
		interfaces:      []*ast.InterfaceDefinition{displayDef, cloneDef, errorDef},
		implementations: implementations,
	}
}

func makeDisplayImpl(typeName string) *ast.ImplementationDefinition {
	selfType := ast.NewSimpleTypeExpression(ast.NewIdentifier(typeName))
	var bodyExpr ast.Expression
	if typeName == "String" {
		bodyExpr = ast.NewIdentifier("self")
	} else {
		bodyExpr = ast.NewStringInterpolation([]ast.Expression{ast.NewIdentifier("self")})
	}
	fn := ast.NewFunctionDefinition(
		ast.NewIdentifier("to_string"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.NewIdentifier("self"), selfType),
		},
		ast.NewBlockExpression([]ast.Statement{ast.NewReturnStatement(bodyExpr)}),
		ast.NewSimpleTypeExpression(ast.NewIdentifier("String")),
		nil,
		nil,
		false,
		false,
	)
	return ast.NewImplementationDefinition(
		ast.NewIdentifier("Display"),
		selfType,
		[]*ast.FunctionDefinition{fn},
		nil,
		nil,
		nil,
		nil,
		false,
	)
}

func makeCloneImpl(typeName string) *ast.ImplementationDefinition {
	selfType := ast.NewSimpleTypeExpression(ast.NewIdentifier(typeName))
	fn := ast.NewFunctionDefinition(
		ast.NewIdentifier("clone"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.NewIdentifier("self"), selfType),
		},
		ast.NewBlockExpression([]ast.Statement{ast.NewReturnStatement(ast.NewIdentifier("self"))}),
		selfType,
		nil,
		nil,
		false,
		false,
	)
	return ast.NewImplementationDefinition(
		ast.NewIdentifier("Clone"),
		selfType,
		[]*ast.FunctionDefinition{fn},
		nil,
		nil,
		nil,
		nil,
		false,
	)
}

func makeErrorImpl() *ast.ImplementationDefinition {
	selfType := ast.NewSimpleTypeExpression(ast.NewIdentifier("ProcError"))

	messageFn := ast.NewFunctionDefinition(
		ast.NewIdentifier("message"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.NewIdentifier("self"), selfType),
		},
		ast.NewBlockExpression([]ast.Statement{
			ast.NewReturnStatement(ast.NewMemberAccessExpression(ast.NewIdentifier("self"), ast.NewIdentifier("details"))),
		}),
		ast.NewSimpleTypeExpression(ast.NewIdentifier("String")),
		nil,
		nil,
		false,
		false,
	)

	causeFn := ast.NewFunctionDefinition(
		ast.NewIdentifier("cause"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.NewIdentifier("self"), selfType),
		},
		ast.NewBlockExpression([]ast.Statement{
			ast.NewReturnStatement(ast.NewNilLiteral()),
		}),
		ast.NewNullableTypeExpression(ast.NewSimpleTypeExpression(ast.NewIdentifier("Error"))),
		nil,
		nil,
		false,
		false,
	)

	return ast.NewImplementationDefinition(
		ast.NewIdentifier("Error"),
		selfType,
		[]*ast.FunctionDefinition{messageFn, causeFn},
		nil,
		nil,
		nil,
		nil,
		false,
	)
}
