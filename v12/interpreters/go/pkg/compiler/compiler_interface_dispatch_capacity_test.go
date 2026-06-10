package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCompilerPreallocatesGeneratedInterfaceDispatchRegistration(t *testing.T) {
	iface := ast.NewInterfaceDefinition(
		ast.ID("Readable"),
		[]*ast.FunctionSignature{
			ast.NewFunctionSignature(ast.ID("read"), []*ast.FunctionParameter{
				ast.NewFunctionParameter(ast.ID("self"), ast.Ty("Self")),
			}, ast.Ty("i32"), nil, nil, nil),
		},
		nil, nil, nil, nil, false,
	)
	box := ast.NewStructDefinition(ast.ID("Box"), nil, ast.StructKindSingleton, nil, nil, false)
	read := ast.NewFunctionDefinition(ast.ID("read"), []*ast.FunctionParameter{
		ast.NewFunctionParameter(ast.ID("self"), ast.Ty("Box")),
	}, ast.NewBlockExpression([]ast.Statement{ast.Int(1)}), ast.Ty("i32"), nil, nil, false, false)
	impl := ast.NewImplementationDefinition(ast.ID("Readable"), ast.Ty("Box"), []*ast.FunctionDefinition{read}, nil, nil, nil, nil, false)
	mainFn := ast.NewFunctionDefinition(ast.ID("main"), nil, ast.NewBlockExpression(nil), ast.Ty("void"), nil, nil, false, false)
	module := ast.NewModule([]ast.Statement{iface, box, impl, mainFn}, nil, ast.NewPackageStatement([]*ast.Identifier{ast.ID("demo")}, false))

	result, err := New(Options{PackageName: "compiled"}).Compile(testProgramFromModule("demo", module))
	if err != nil {
		t.Fatalf("compile interface dispatch capacity fixture: %v", err)
	}
	source := combinedGeneratedSource(result)
	for _, want := range []string{
		"__able_interface_dispatch = __able_new_interface_dispatch()",
		"perIface = __able_new_interface_dispatch_methods(ifaceName)",
		"methods[\"read\"] = make([]__able_interface_dispatch_entry, 0, 1)",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source does not contain %q", want)
		}
	}
}
