package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestRegisterCompiledImplMethodOverloadCanonicalizesStoredInterfaceAlias(t *testing.T) {
	interp := newInterpreter(nil, execModeTreewalker)
	cloneNode := ast.NewInterfaceDefinition(ast.ID("Clone"), nil, nil, nil, nil, nil, false)
	bootstrapClone := &runtime.InterfaceDefinitionValue{
		Node:          cloneNode,
		QualifiedName: "Clone",
	}
	canonicalClone := &runtime.InterfaceDefinitionValue{
		Node:          cloneNode,
		QualifiedName: "able.kernel.Clone",
	}
	interp.interfaces["Clone"] = bootstrapClone
	interp.interfaces["able.kernel.Clone"] = canonicalClone
	interp.typeAliases["Clone"] = ast.NewTypeAliasDefinition(
		ast.ID("Clone"),
		ast.NewSimpleTypeExpression(ast.ID("able.kernel.Clone")),
		nil,
		nil,
		false,
	)

	typeParam := ast.NewSimpleTypeExpression(ast.ID("T"))
	arrayType := ast.NewGenericTypeExpression(ast.NewSimpleTypeExpression(ast.ID("Array")), []ast.TypeExpression{typeParam})
	cloneMethod := ast.NewFunctionDefinition(
		ast.ID("clone"),
		[]*ast.FunctionParameter{ast.NewFunctionParameter(ast.ID("self"), arrayType)},
		nil,
		arrayType,
		nil,
		nil,
		false,
		false,
	)
	cloneFunction := &runtime.FunctionValue{Declaration: cloneMethod}
	implDefinition := ast.NewImplementationDefinition(
		ast.ID("Clone"),
		arrayType,
		[]*ast.FunctionDefinition{cloneMethod},
		nil,
		[]*ast.GenericParameter{ast.NewGenericParameter(ast.ID("T"), nil)},
		nil,
		nil,
		false,
	)
	interp.genericImpls = append(interp.genericImpls, implEntry{
		interfaceName:      "Clone",
		methods:            map[string]runtime.Value{"clone": cloneFunction},
		definition:         implDefinition,
		registrationTarget: arrayType,
		genericParams:      implDefinition.GenericParams,
	})

	thunk := CompiledThunk(func(_ *runtime.Environment, _ []runtime.Value) (runtime.Value, error) {
		return runtime.NilValue{}, nil
	})
	if err := interp.RegisterCompiledImplMethodOverload(
		"able.kernel.Clone",
		arrayType,
		nil,
		"<none>",
		"",
		"clone",
		[]ast.TypeExpression{arrayType},
		thunk,
	); err != nil {
		t.Fatalf("register compiled generic Array Clone impl through canonical interface identity: %v", err)
	}
	if cloneFunction.Bytecode == nil {
		t.Fatal("expected compiled thunk to be attached to aliased generic impl method")
	}
}

func TestRegisterCompiledImplMethodOverloadDoesNotMergeUnrelatedShortInterfaceName(t *testing.T) {
	interp := newInterpreter(nil, execModeTreewalker)
	markerNode := ast.NewInterfaceDefinition(ast.ID("Marker"), nil, nil, nil, nil, nil, false)
	interp.interfaces["Marker"] = &runtime.InterfaceDefinitionValue{
		Node:          markerNode,
		QualifiedName: "Marker",
	}
	interp.interfaces["other.Marker"] = &runtime.InterfaceDefinitionValue{
		Node:          markerNode,
		QualifiedName: "other.Marker",
	}

	targetType := ast.NewSimpleTypeExpression(ast.ID("Box"))
	markerMethod := ast.NewFunctionDefinition(
		ast.ID("mark"),
		[]*ast.FunctionParameter{ast.NewFunctionParameter(ast.ID("self"), targetType)},
		nil,
		targetType,
		nil,
		nil,
		false,
		false,
	)
	markerFunction := &runtime.FunctionValue{Declaration: markerMethod}
	implDefinition := ast.NewImplementationDefinition(
		ast.ID("Marker"),
		targetType,
		[]*ast.FunctionDefinition{markerMethod},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	interp.implMethods["Box"] = []implEntry{{
		interfaceName:      "Marker",
		methods:            map[string]runtime.Value{"mark": markerFunction},
		definition:         implDefinition,
		registrationTarget: targetType,
	}}

	err := interp.RegisterCompiledImplMethodOverload(
		"other.Marker",
		targetType,
		nil,
		"<none>",
		"",
		"mark",
		[]ast.TypeExpression{targetType},
		func(_ *runtime.Environment, _ []runtime.Value) (runtime.Value, error) {
			return runtime.NilValue{}, nil
		},
	)
	if err == nil {
		t.Fatal("expected unrelated qualified interface to remain distinct")
	}
	if markerFunction.Bytecode != nil {
		t.Fatal("unrelated short-name impl received compiled thunk")
	}
}
