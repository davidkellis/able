package typechecker

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func makeIndexInterfaces() (*ast.InterfaceDefinition, *ast.InterfaceDefinition) {
	idx := ast.GenericParam("Idx", nil)
	val := ast.GenericParam("Val", nil)
	index := ast.Iface(
		"Index",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"get",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
					ast.Param("idx", ast.Ty("Idx")),
				},
				ast.Result(ast.Ty("Val")),
				nil,
				nil,
				nil,
			),
		},
		[]*ast.GenericParameter{idx, val},
		nil,
		nil,
		nil,
		false,
	)
	indexMut := ast.Iface(
		"IndexMut",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"set",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
					ast.Param("idx", ast.Ty("Idx")),
					ast.Param("value", ast.Ty("Val")),
				},
				ast.Result(ast.Ty("void")),
				nil,
				nil,
				nil,
			),
		},
		[]*ast.GenericParameter{idx, val},
		nil,
		nil,
		nil,
		false,
	)
	return index, indexMut
}

func makeBoxStruct() *ast.StructDefinition {
	return ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("i32"), "value")},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
}

func TestIndexAssignmentUsesIndexMutImplementation(t *testing.T) {
	checker := New()
	index, indexMut := makeIndexInterfaces()
	boxStruct := makeBoxStruct()

	getFn := ast.Fn(
		"get",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("idx", ast.Ty("i32")),
		},
		[]ast.Statement{ast.Ret(ast.ImplicitMember("value"))},
		ast.Result(ast.Ty("i32")),
		nil,
		nil,
		true,
		false,
	)
	setFn := ast.Fn(
		"set",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("idx", ast.Ty("i32")),
			ast.Param("value", ast.Ty("i32")),
		},
		[]ast.Statement{ast.Ret(ast.Nil())},
		ast.Result(ast.Ty("void")),
		nil,
		nil,
		true,
		false,
	)

	indexImpl := ast.Impl("Index", ast.Ty("Box"), []*ast.FunctionDefinition{getFn}, nil, nil, []ast.TypeExpression{ast.Ty("i32"), ast.Ty("i32")}, nil, false)
	indexMutImpl := ast.Impl("IndexMut", ast.Ty("Box"), []*ast.FunctionDefinition{setFn}, nil, nil, []ast.TypeExpression{ast.Ty("i32"), ast.Ty("i32")}, nil, false)

	assignBox := ast.Assign(ast.ID("b"), ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(1), "value")}, false, "Box", nil, nil))
	writeIndex := ast.AssignIndex(ast.ID("b"), ast.Int(0), ast.Int(9))

	module := ast.NewModule([]ast.Statement{index, indexMut, boxStruct, indexImpl, indexMutImpl, assignBox, writeIndex}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestIndexAssignmentRequiresIndexMut(t *testing.T) {
	checker := New()
	index, _ := makeIndexInterfaces()
	boxStruct := makeBoxStruct()

	getFn := ast.Fn(
		"get",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("idx", ast.Ty("i32")),
		},
		[]ast.Statement{ast.Ret(ast.ImplicitMember("value"))},
		ast.Result(ast.Ty("i32")),
		nil,
		nil,
		true,
		false,
	)
	indexImpl := ast.Impl("Index", ast.Ty("Box"), []*ast.FunctionDefinition{getFn}, nil, nil, []ast.TypeExpression{ast.Ty("i32"), ast.Ty("i32")}, nil, false)

	assignBox := ast.Assign(ast.ID("b"), ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(1), "value")}, false, "Box", nil, nil))
	writeIndex := ast.AssignIndex(ast.ID("b"), ast.Int(0), ast.Int(2))

	module := ast.NewModule([]ast.Statement{index, boxStruct, indexImpl, assignBox, writeIndex}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics for missing IndexMut, got none")
	}
	found := false
	for _, diag := range diags {
		if diag.Node == writeIndex || (diag.Message != "" && strings.Contains(diag.Message, "IndexMut")) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected diagnostic referencing IndexMut, got %v", diags)
	}
}

func TestIndexAssignmentRejectsDeclareOperator(t *testing.T) {
	checker := New()
	index, indexMut := makeIndexInterfaces()
	boxStruct := makeBoxStruct()
	assignBox := ast.Assign(ast.ID("b"), ast.StructLit([]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(1), "value")}, false, "Box", nil, nil))
	assignIndex := ast.AssignOp(ast.AssignmentDeclare, ast.Index(ast.ID("b"), ast.Int(0)), ast.Int(1))

	module := ast.NewModule([]ast.Statement{index, indexMut, boxStruct, assignBox, assignIndex}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, diag := range diags {
		if diag.Message != "" && strings.Contains(diag.Message, "cannot use :=") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected := index assignment diagnostic, got %v", diags)
	}
}
