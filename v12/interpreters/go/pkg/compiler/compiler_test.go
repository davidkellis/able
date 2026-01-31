package compiler

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

func TestCompilerEmitsStructsAndWrappers(t *testing.T) {
	point := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("f64"), "x"),
			ast.FieldDef(ast.Ty("f64"), "y"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	makePoint := ast.Fn(
		"make_point",
		nil,
		[]ast.Statement{
			ast.Ret(ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Flt(1.5), "x"),
					ast.FieldInit(ast.Flt(2.5), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			)),
		},
		ast.Ty("Point"),
		nil,
		nil,
		false,
		false,
	)
	getX := ast.Fn(
		"get_x",
		[]*ast.FunctionParameter{
			ast.Param("p", ast.Ty("Point")),
		},
		[]ast.Statement{
			ast.Ret(ast.Member(ast.ID("p"), "x")),
		},
		ast.Ty("f64"),
		nil,
		nil,
		false,
		false,
	)
	makeNumbers := ast.Fn(
		"make_numbers",
		nil,
		[]ast.Statement{
			ast.Ret(ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3))),
		},
		ast.Gen(ast.Ty("Array"), ast.Ty("i32")),
		nil,
		nil,
		false,
		false,
	)
	makePoints := ast.Fn(
		"make_points",
		nil,
		[]ast.Statement{
			ast.Ret(ast.Arr(
				ast.StructLit(
					[]*ast.StructFieldInitializer{
						ast.FieldInit(ast.Flt(1.0), "x"),
						ast.FieldInit(ast.Flt(2.0), "y"),
					},
					false,
					"Point",
					nil,
					nil,
				),
				ast.StructLit(
					[]*ast.StructFieldInitializer{
						ast.FieldInit(ast.Flt(3.0), "x"),
						ast.FieldInit(ast.Flt(4.0), "y"),
					},
					false,
					"Point",
					nil,
					nil,
				),
			)),
		},
		ast.Gen(ast.Ty("Array"), ast.Ty("Point")),
		nil,
		nil,
		false,
		false,
	)
	peek := ast.Fn(
		"peek",
		[]*ast.FunctionParameter{
			ast.Param("nums", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("idx", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentDeclare, ast.ID("val"), ast.Index(ast.ID("nums"), ast.ID("idx"))),
			ast.ID("nums"),
		},
		ast.Gen(ast.Ty("Array"), ast.Ty("i32")),
		nil,
		nil,
		false,
		false,
	)
	first := ast.Fn(
		"first",
		[]*ast.FunctionParameter{
			ast.Param("nums", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
		},
		[]ast.Statement{
			ast.Ret(ast.Index(ast.ID("nums"), ast.Int(0))),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	setFirst := ast.Fn(
		"set_first",
		[]*ast.FunctionParameter{
			ast.Param("nums", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
		},
		[]ast.Statement{
			ast.AssignIndex(ast.ID("nums"), ast.Int(0), ast.Int(5)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	add := ast.Fn(
		"add",
		[]*ast.FunctionParameter{
			ast.Param("x", ast.Ty("i64")),
			ast.Param("y", ast.Ty("i64")),
		},
		[]ast.Statement{
			ast.Ret(ast.Bin("+", ast.ID("x"), ast.ID("y"))),
		},
		ast.Ty("i64"),
		nil,
		nil,
		false,
		false,
	)
	isPositive := ast.Fn(
		"is_positive",
		[]*ast.FunctionParameter{
			ast.Param("x", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Ret(ast.Bin(">", ast.ID("x"), ast.Int(0))),
		},
		ast.Ty("bool"),
		nil,
		nil,
		false,
		false,
	)
	negate := ast.Fn(
		"negate",
		[]*ast.FunctionParameter{
			ast.Param("x", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Ret(ast.Un(ast.UnaryOperatorNegate, ast.ID("x"))),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	notFlag := ast.Fn(
		"not_flag",
		[]*ast.FunctionParameter{
			ast.Param("flag", ast.Ty("bool")),
		},
		[]ast.Statement{
			ast.Ret(ast.Un(ast.UnaryOperatorNot, ast.ID("flag"))),
		},
		ast.Ty("bool"),
		nil,
		nil,
		false,
		false,
	)
	bitFlip := ast.Fn(
		"bit_flip",
		[]*ast.FunctionParameter{
			ast.Param("mask", ast.Ty("u32")),
		},
		[]ast.Statement{
			ast.Ret(ast.Un(ast.UnaryOperatorBitNot, ast.ID("mask"))),
		},
		ast.Ty("u32"),
		nil,
		nil,
		false,
		false,
	)
	choose := ast.Fn(
		"choose",
		[]*ast.FunctionParameter{
			ast.Param("flag", ast.Ty("bool")),
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Ret(ast.NewIfExpression(
				ast.ID("flag"),
				ast.Block(ast.ID("a")),
				nil,
				ast.Block(ast.ID("b")),
			)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	chooseThree := ast.Fn(
		"choose_three",
		[]*ast.FunctionParameter{
			ast.Param("flag", ast.Ty("bool")),
			ast.Param("mid", ast.Ty("bool")),
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Ty("i32")),
			ast.Param("c", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Ret(ast.NewIfExpression(
				ast.ID("flag"),
				ast.Block(ast.ID("a")),
				[]*ast.ElseIfClause{
					ast.ElseIf(ast.Block(ast.ID("b")), ast.ID("mid")),
				},
				ast.Block(ast.ID("c")),
			)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	logic := ast.Fn(
		"logic",
		[]*ast.FunctionParameter{
			ast.Param("a", ast.Ty("bool")),
			ast.Param("b", ast.Ty("bool")),
		},
		[]ast.Statement{
			ast.Ret(ast.Bin("||", ast.Bin("&&", ast.ID("a"), ast.ID("b")), ast.ID("a"))),
		},
		ast.Ty("bool"),
		nil,
		nil,
		false,
		false,
	)
	isGreeting := ast.Fn(
		"is_greeting",
		[]*ast.FunctionParameter{
			ast.Param("name", ast.Ty("String")),
		},
		[]ast.Statement{
			ast.Ret(ast.Bin("==", ast.ID("name"), ast.Str("hi"))),
		},
		ast.Ty("bool"),
		nil,
		nil,
		false,
		false,
	)
	plusOne := ast.Fn(
		"plus_one",
		[]*ast.FunctionParameter{
			ast.Param("x", ast.Ty("i64")),
		},
		[]ast.Statement{
			ast.Ret(ast.CallExpr(ast.ID("add"), ast.ID("x"), ast.Int(1))),
		},
		ast.Ty("i64"),
		nil,
		nil,
		false,
		false,
	)
	increment := ast.Fn(
		"increment",
		[]*ast.FunctionParameter{
			ast.Param("x", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ID("y"), ast.ID("x")),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("y"), ast.Bin("+", ast.ID("y"), ast.Int(1))),
			ast.ID("y"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	shadow := ast.Fn(
		"shadow",
		[]*ast.FunctionParameter{
			ast.Param("x", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentDeclare, ast.ID("y"), ast.ID("x")),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("y"), ast.Bin("+", ast.ID("y"), ast.Int(2))),
			ast.ID("y"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	fallback := ast.Fn(
		"complex",
		nil,
		[]ast.Statement{
			ast.Ret(ast.Bin("/", ast.Int(1), ast.Int(2))),
		},
		ast.Ty("i64"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod(
		[]ast.Statement{point, makePoint, getX, makeNumbers, makePoints, peek, first, setFirst, add, isPositive, negate, notFlag, bitFlip, choose, chooseThree, logic, isGreeting, plusOne, increment, shadow, fallback},
		nil,
		ast.Pkg([]interface{}{"app"}, false),
	)
	entry := annotatedModule("app", module, "app.able", nil)
	program := &driver.Program{Entry: entry, Modules: []*driver.Module{entry}}

	comp := New(Options{PackageName: "compiled"})
	result, err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	src, ok := result.Files["compiled.go"]
	if !ok {
		t.Fatalf("expected compiled.go output")
	}
	code := string(src)
	if !strings.Contains(code, "type Point struct") {
		t.Fatalf("expected Point struct definition")
	}
	if !strings.Contains(code, "__able_compiled_fn_make_point") {
		t.Fatalf("expected compiled function for make_point")
	}
	if !strings.Contains(code, "__able_compiled_fn_get_x") {
		t.Fatalf("expected compiled function for get_x")
	}
	if !strings.Contains(code, "__able_compiled_fn_make_numbers") {
		t.Fatalf("expected compiled function for make_numbers")
	}
	if !strings.Contains(code, "__able_compiled_fn_make_points") {
		t.Fatalf("expected compiled function for make_points")
	}
	if !strings.Contains(code, "__able_compiled_fn_peek") {
		t.Fatalf("expected compiled function for peek")
	}
	if !strings.Contains(code, "__able_compiled_fn_first") {
		t.Fatalf("expected compiled function for first")
	}
	if !strings.Contains(code, "__able_compiled_fn_set_first") {
		t.Fatalf("expected compiled function for set_first")
	}
	if !strings.Contains(code, "__able_compiled_fn_add") {
		t.Fatalf("expected compiled function for add")
	}
	if !strings.Contains(code, "__able_compiled_fn_is_positive") {
		t.Fatalf("expected compiled function for is_positive")
	}
	if !strings.Contains(code, "__able_compiled_fn_negate") {
		t.Fatalf("expected compiled function for negate")
	}
	if !strings.Contains(code, "__able_compiled_fn_not_flag") {
		t.Fatalf("expected compiled function for not_flag")
	}
	if !strings.Contains(code, "__able_compiled_fn_bit_flip") {
		t.Fatalf("expected compiled function for bit_flip")
	}
	if !strings.Contains(code, "__able_compiled_fn_choose") {
		t.Fatalf("expected compiled function for choose")
	}
	if !strings.Contains(code, "__able_compiled_fn_choose_three") {
		t.Fatalf("expected compiled function for choose_three")
	}
	if !strings.Contains(code, "__able_compiled_fn_logic") {
		t.Fatalf("expected compiled function for logic")
	}
	if !strings.Contains(code, "__able_compiled_fn_is_greeting") {
		t.Fatalf("expected compiled function for is_greeting")
	}
	if !strings.Contains(code, "__able_compiled_fn_plus_one") {
		t.Fatalf("expected compiled function for plus_one")
	}
	if !strings.Contains(code, "__able_compiled_fn_increment") {
		t.Fatalf("expected compiled function for increment")
	}
	if !strings.Contains(code, "__able_compiled_fn_shadow") {
		t.Fatalf("expected compiled function for shadow")
	}
	if !strings.Contains(code, "CallOriginal(\"complex\"") {
		t.Fatalf("expected fallback wrapper for complex")
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "compiled.go", code, parser.AllErrors); err != nil {
		t.Fatalf("generated code parse error: %v", err)
	}
}

func annotatedModule(pkg string, module *ast.Module, file string, imports []string) *driver.Module {
	files := []string{}
	if file != "" {
		files = []string{file}
	}
	origins := make(map[ast.Node]string)
	ast.AnnotateOrigins(module, file, origins)
	return &driver.Module{
		Package:     pkg,
		AST:         module,
		Files:       files,
		Imports:     imports,
		NodeOrigins: origins,
	}
}
