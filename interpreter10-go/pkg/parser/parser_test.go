package parser_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/interpreter"
	"able/interpreter10-go/pkg/parser"
)

type fixtureCase struct {
	name   string
	source string
}

func loadFixtureModule(t testing.TB, fixtureName string) *ast.Module {
	t.Helper()
	modulePath := filepath.Join("..", "..", "..", "fixtures", "ast", fixtureName, "module.json")
	data, err := os.ReadFile(modulePath)
	if err != nil {
		t.Fatalf("read fixture module %s: %v", modulePath, err)
	}
	mod, err := interpreter.DecodeModule(data)
	if err != nil {
		t.Fatalf("decode fixture module %s: %v", modulePath, err)
	}
	return mod
}

func assertModulesEqual(t testing.TB, expected, actual *ast.Module) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		return
	}
	wantJSON, _ := json.MarshalIndent(expected, "", "  ")
	gotJSON, _ := json.MarshalIndent(actual, "", "  ")
	t.Fatalf("module mismatch\nexpected: %s\n   actual: %s", wantJSON, gotJSON)
}

func runFixtureCases(t *testing.T, cases []fixtureCase) {
	t.Helper()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p, err := parser.NewModuleParser()
			if err != nil {
				t.Fatalf("NewModuleParser error: %v", err)
			}
			defer p.Close()

			mod, err := p.ParseModule([]byte(tc.source))
			if err != nil {
				t.Fatalf("ParseModule error for %s: %v", tc.name, err)
			}

			expected := loadFixtureModule(t, tc.name)
			assertModulesEqual(t, expected, mod)
		})
	}
}

func placeholder(n int) *ast.PlaceholderExpression {
	if n <= 0 {
		return ast.NewPlaceholderExpression(nil)
	}
	idx := n
	return ast.NewPlaceholderExpression(&idx)
}

func TestParseImplicitMethods(t *testing.T) {
	source := "struct Counter {\n  value: i32,\n}\n\nmethods Counter {\n  fn #increment() {\n    #value = #value + 1\n  }\n\n  fn #add(amount: i32) {\n    #value = #value + amount\n  }\n}\n\nimpl Display for Counter {\n  fn #to_string() -> string {\n    `Counter(${#value})`\n  }\n}\n"

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	structDef := ast.NewStructDefinition(
		ast.ID("Counter"),
		[]*ast.StructFieldDefinition{
			ast.NewStructFieldDefinition(ast.Ty("i32"), ast.ID("value")),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	incrementBody := ast.NewBlockExpression([]ast.Statement{
		ast.NewAssignmentExpression(
			ast.AssignmentAssign,
			ast.NewImplicitMemberExpression(ast.ID("value")),
			ast.NewBinaryExpression(
				"+",
				ast.NewImplicitMemberExpression(ast.ID("value")),
				ast.Int(1),
			),
		),
	})

	addBody := ast.NewBlockExpression([]ast.Statement{
		ast.NewAssignmentExpression(
			ast.AssignmentAssign,
			ast.NewImplicitMemberExpression(ast.ID("value")),
			ast.NewBinaryExpression(
				"+",
				ast.NewImplicitMemberExpression(ast.ID("value")),
				ast.ID("amount"),
			),
		),
	})

	methodsDef := ast.NewMethodsDefinition(
		ast.Ty("Counter"),
		[]*ast.FunctionDefinition{
			ast.NewFunctionDefinition(ast.ID("increment"), nil, incrementBody, nil, nil, nil, true, false),
			ast.NewFunctionDefinition(
				ast.ID("add"),
				[]*ast.FunctionParameter{
					ast.NewFunctionParameter(ast.ID("amount"), ast.Ty("i32")),
				},
				addBody,
				nil,
				nil,
				nil,
				true,
				false,
			),
		},
		nil,
		nil,
	)

	toStringBody := ast.NewBlockExpression([]ast.Statement{
		ast.NewStringInterpolation([]ast.Expression{
			ast.Str("Counter("),
			ast.NewImplicitMemberExpression(ast.ID("value")),
			ast.Str(")"),
		}),
	})

	implDef := ast.NewImplementationDefinition(
		ast.ID("Display"),
		ast.Ty("Counter"),
		[]*ast.FunctionDefinition{
			ast.NewFunctionDefinition(
				ast.ID("to_string"),
				nil,
				toStringBody,
				ast.Ty("string"),
				nil,
				nil,
				true,
				false,
			),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)

	expected := ast.NewModule([]ast.Statement{structDef, methodsDef, implDef}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}
	assertModulesEqual(t, expected, mod)
}

func TestParsePlaceholderExpressions(t *testing.T) {
	source := "fn partials(data, factor) {\n  add(@, 10)\n  merge(@, @2, @1)\n  5.add\n}\n"

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	callAdd := ast.NewFunctionCall(
		ast.ID("add"),
		[]ast.Expression{
			placeholder(0),
			ast.Int(10),
		},
		nil,
		false,
	)

	mergeCall := ast.NewFunctionCall(
		ast.ID("merge"),
		[]ast.Expression{
			placeholder(0),
			placeholder(2),
			placeholder(1),
		},
		nil,
		false,
	)

	memberAccess := ast.NewMemberAccessExpression(ast.Int(5), ast.ID("add"))

	fnBody := ast.NewBlockExpression([]ast.Statement{callAdd, mergeCall, memberAccess})

	expected := ast.NewModule([]ast.Statement{
		ast.NewFunctionDefinition(
			ast.ID("partials"),
			[]*ast.FunctionParameter{
				ast.NewFunctionParameter(ast.ID("data"), nil),
				ast.NewFunctionParameter(ast.ID("factor"), nil),
			},
			fnBody,
			nil,
			nil,
			nil,
			false,
			false,
		),
	}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseLiteralExpressions(t *testing.T) {
	source := `fn literals() {
  signed := 42_i64
  unsigned := 255_u8
  precise := 3.25_f32
  wide := 2.0
  truth := true
  lie := false
  ch := 'x'
  newline := '\n'
  text := "hello"
  nothing := nil
  arr := [1, 2, 3]
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	i64 := ast.IntegerTypeI64
	u8 := ast.IntegerTypeU8
	f32 := ast.FloatTypeF32

	body := ast.Block(
		ast.Assign(ast.ID("signed"), ast.IntTyped(42, &i64)),
		ast.Assign(ast.ID("unsigned"), ast.IntTyped(255, &u8)),
		ast.Assign(ast.ID("precise"), ast.FltTyped(3.25, &f32)),
		ast.Assign(ast.ID("wide"), ast.Flt(2.0)),
		ast.Assign(ast.ID("truth"), ast.Bool(true)),
		ast.Assign(ast.ID("lie"), ast.Bool(false)),
		ast.Assign(ast.ID("ch"), ast.Chr("x")),
		ast.Assign(ast.ID("newline"), ast.Chr("\n")),
		ast.Assign(ast.ID("text"), ast.Str("hello")),
		ast.Assign(ast.ID("nothing"), ast.Nil()),
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3))),
	)

	expected := ast.NewModule([]ast.Statement{
		ast.NewFunctionDefinition(
			ast.ID("literals"),
			nil,
			body,
			nil,
			nil,
			nil,
			false,
			false,
		),
	}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseStructAndArrayLiterals(t *testing.T) {
	source := `struct Point {
  x: i32,
  y: i32,
}

fn build() {
  point := Point{ x: 1, y: 2 }
  updated := Point{ ...point, x: 7 }
  coords := [point.x, updated.y]
  value := updated.x
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	pointStruct := ast.NewStructDefinition(
		ast.ID("Point"),
		[]*ast.StructFieldDefinition{
			ast.NewStructFieldDefinition(ast.Ty("i32"), ast.ID("x")),
			ast.NewStructFieldDefinition(ast.Ty("i32"), ast.ID("y")),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	pointLiteral := ast.NewStructLiteral(
		[]*ast.StructFieldInitializer{
			ast.NewStructFieldInitializer(ast.Int(1), ast.ID("x"), false),
			ast.NewStructFieldInitializer(ast.Int(2), ast.ID("y"), false),
		},
		false,
		ast.ID("Point"),
		nil,
		nil,
	)

	updatedLiteral := ast.NewStructLiteral(
		[]*ast.StructFieldInitializer{
			ast.NewStructFieldInitializer(ast.Int(7), ast.ID("x"), false),
		},
		false,
		ast.ID("Point"),
		[]ast.Expression{ast.ID("point")},
		nil,
	)

	coords := ast.Arr(
		ast.NewMemberAccessExpression(ast.ID("point"), ast.ID("x")),
		ast.NewMemberAccessExpression(ast.ID("updated"), ast.ID("y")),
	)

	valueExpr := ast.NewMemberAccessExpression(ast.ID("updated"), ast.ID("x"))

	body := ast.Block(
		ast.Assign(ast.ID("point"), pointLiteral),
		ast.Assign(ast.ID("updated"), updatedLiteral),
		ast.Assign(ast.ID("coords"), coords),
		ast.Assign(ast.ID("value"), valueExpr),
	)

	expected := ast.NewModule([]ast.Statement{
		pointStruct,
		ast.NewFunctionDefinition(
			ast.ID("build"),
			nil,
			body,
			nil,
			nil,
			nil,
			false,
			false,
		),
	}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseStructLiteralMultipleSpreads(t *testing.T) {
	source := `struct Point {
  x: i32,
  y: i32,
  z: i32,
}

fn merge(p1: Point, p2: Point) -> Point {
  Point{ ...p1, ...p2, z: 42 }
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	pointStruct := ast.NewStructDefinition(
		ast.ID("Point"),
		[]*ast.StructFieldDefinition{
			ast.NewStructFieldDefinition(ast.Ty("i32"), ast.ID("x")),
			ast.NewStructFieldDefinition(ast.Ty("i32"), ast.ID("y")),
			ast.NewStructFieldDefinition(ast.Ty("i32"), ast.ID("z")),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	mergedLiteral := ast.NewStructLiteral(
		[]*ast.StructFieldInitializer{
			ast.NewStructFieldInitializer(ast.Int(42), ast.ID("z"), false),
		},
		false,
		ast.ID("Point"),
		[]ast.Expression{ast.ID("p1"), ast.ID("p2")},
		nil,
	)

	fnBody := ast.Block(
		mergedLiteral,
	)

	expected := ast.NewModule([]ast.Statement{
		pointStruct,
		ast.NewFunctionDefinition(
			ast.ID("merge"),
			[]*ast.FunctionParameter{
				ast.NewFunctionParameter(ast.ID("p1"), ast.Ty("Point")),
				ast.NewFunctionParameter(ast.ID("p2"), ast.Ty("Point")),
			},
			fnBody,
			ast.Ty("Point"),
			nil,
			nil,
			false,
			false,
		),
	}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseMatchExpression(t *testing.T) {
	source := `struct Point {
  x: i32,
  y: i32,
}

fn classify(point: Point) -> i32 {
  point match {
    case Point { x: a, y: b } if b > a => a + b,
    case Point { x: value, y: value } => value,
    case _ => 0,
  }
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	pointStruct := ast.NewStructDefinition(
		ast.ID("Point"),
		[]*ast.StructFieldDefinition{
			ast.NewStructFieldDefinition(ast.Ty("i32"), ast.ID("x")),
			ast.NewStructFieldDefinition(ast.Ty("i32"), ast.ID("y")),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	firstPattern := ast.StructP([]*ast.StructPatternField{
		ast.FieldP(ast.ID("a"), "x", nil),
		ast.FieldP(ast.ID("b"), "y", nil),
	}, false, "Point")

	firstGuard := ast.NewBinaryExpression(
		">",
		ast.ID("b"),
		ast.ID("a"),
	)
	firstBody := ast.NewBinaryExpression("+", ast.ID("a"), ast.ID("b"))

	secondPattern := ast.StructP([]*ast.StructPatternField{
		ast.FieldP(ast.ID("value"), "x", nil),
		ast.FieldP(ast.ID("value"), "y", nil),
	}, false, "Point")

	matchExpr := ast.NewMatchExpression(
		ast.ID("point"),
		[]*ast.MatchClause{
			ast.NewMatchClause(firstPattern, firstBody, firstGuard),
			ast.NewMatchClause(secondPattern, ast.ID("value"), nil),
			ast.NewMatchClause(ast.Wc(), ast.Int(0), nil),
		},
	)

	body := ast.Block(
		matchExpr,
	)

	fn := ast.NewFunctionDefinition(
		ast.ID("classify"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("point"), ast.Ty("Point")),
		},
		body,
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	expected := ast.NewModule([]ast.Statement{
		pointStruct,
		fn,
	}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParsePropagationAndOrElse(t *testing.T) {
	source := `fn handlers(opt: ?i32, res: !i32) -> !i32 {
 	value := opt! else { 0 }
 	processed := res! else { |err| err }
 	processed
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	value := ast.Assign(
		ast.ID("value"),
		ast.NewOrElseExpression(
			ast.Prop(ast.ID("opt")),
			ast.Block(ast.Int(0)),
			nil,
		),
	)

	processed := ast.Assign(
		ast.ID("processed"),
		ast.NewOrElseExpression(
			ast.Prop(ast.ID("res")),
			ast.Block(ast.ID("err")),
			ast.ID("err"),
		),
	)

	body := ast.Block(
		value,
		processed,
		ast.ID("processed"),
	)

	fn := ast.NewFunctionDefinition(
		ast.ID("handlers"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("opt"), ast.Nullable(ast.Ty("i32"))),
			ast.NewFunctionParameter(ast.ID("res"), ast.Result(ast.Ty("i32"))),
		},
		body,
		ast.Result(ast.Ty("i32")),
		nil,
		nil,
		false,
		false,
	)

	expected := ast.NewModule([]ast.Statement{fn}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseRescueAndEnsure(t *testing.T) {
	source := `struct MyError {
  message: string,
}

impl Error for MyError {
  fn message(self: Self) -> string { "boom" }
  fn cause(self: Self) -> ?Error { nil }
}

fn guard() -> string {
  attempt := (do {
    raise MyError{ message: "boom" }
    "ok"
  } rescue {
    case _ => "rescued"
  }) ensure {
    print("cleanup")
  }
  attempt
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	myErrorStruct := ast.NewStructDefinition(
		ast.ID("MyError"),
		[]*ast.StructFieldDefinition{
			ast.NewStructFieldDefinition(ast.Ty("string"), ast.ID("message")),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	messageFn := ast.NewFunctionDefinition(
		ast.ID("message"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("self"), ast.Ty("Self")),
		},
		ast.Block(ast.Str("boom")),
		ast.Ty("string"),
		nil,
		nil,
		true,
		false,
	)

	causeFn := ast.NewFunctionDefinition(
		ast.ID("cause"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("self"), ast.Ty("Self")),
		},
		ast.Block(ast.Nil()),
		ast.Nullable(ast.Ty("Error")),
		nil,
		nil,
		true,
		false,
	)

	implDef := ast.NewImplementationDefinition(
		ast.ID("Error"),
		ast.Ty("MyError"),
		[]*ast.FunctionDefinition{messageFn, causeFn},
		nil,
		nil,
		nil,
		nil,
		false,
	)

	errorLiteral := ast.NewStructLiteral(
		[]*ast.StructFieldInitializer{
			ast.NewStructFieldInitializer(ast.Str("boom"), ast.ID("message"), false),
		},
		false,
		ast.ID("MyError"),
		nil,
		nil,
	)

	monitored := ast.Block(
		ast.NewRaiseStatement(errorLiteral),
		ast.Str("ok"),
	)

	rescueExpr := ast.NewRescueExpression(
		monitored,
		[]*ast.MatchClause{
			ast.NewMatchClause(ast.Wc(), ast.Str("rescued"), nil),
		},
	)

	ensureBlock := ast.Block(
		ast.NewFunctionCall(ast.ID("print"), []ast.Expression{ast.Str("cleanup")}, nil, false),
	)

	ensureExpr := ast.NewEnsureExpression(rescueExpr, ensureBlock)

	fnBody := ast.Block(
		ast.Assign(ast.ID("attempt"), ensureExpr),
		ast.ID("attempt"),
	)

	fn := ast.NewFunctionDefinition(
		ast.ID("guard"),
		nil,
		fnBody,
		ast.Ty("string"),
		nil,
		nil,
		false,
		false,
	)

	expected := ast.NewModule([]ast.Statement{
		myErrorStruct,
		implDef,
		fn,
	}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}
	t.Logf("expected imports=%v actual imports=%v", expected.Imports, mod.Imports)

	assertModulesEqual(t, expected, mod)
}

func TestParseBreakpointExpression(t *testing.T) {
	source := `fn debug(value: i32) -> i32 {
  breakpoint 'trace {
    log(value)
    value + 1
  }
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	breakBody := ast.Block(
		ast.NewFunctionCall(ast.ID("log"), []ast.Expression{ast.ID("value")}, nil, false),
		ast.NewBinaryExpression("+", ast.ID("value"), ast.Int(1)),
	)

	fnBody := ast.Block(
		ast.NewBreakpointExpression(ast.ID("trace"), breakBody),
	)

	fn := ast.NewFunctionDefinition(
		ast.ID("debug"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("value"), ast.Ty("i32")),
		},
		fnBody,
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	expected := ast.NewModule([]ast.Statement{fn}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseIfOrExpression(t *testing.T) {
	source := `fn grade(score: i32) -> string {
  if score >= 90 { "A" }
  or score >= 80 { "B" }
  or { "C" }
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	firstCond := ast.NewBinaryExpression(">=", ast.ID("score"), ast.Int(90))
	firstBody := ast.Block(ast.Str("A"))
	secondCond := ast.NewBinaryExpression(">=", ast.ID("score"), ast.Int(80))
	secondBody := ast.Block(ast.Str("B"))
	defaultBody := ast.Block(ast.Str("C"))

	ifExpr := ast.NewIfExpression(
		firstCond,
		firstBody,
		[]*ast.OrClause{
			ast.NewOrClause(secondBody, secondCond),
			ast.NewOrClause(defaultBody, nil),
		},
	)

	fnBody := ast.Block(ifExpr)

	fn := ast.NewFunctionDefinition(
		ast.ID("grade"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("score"), ast.Ty("i32")),
		},
		fnBody,
		ast.Ty("string"),
		nil,
		nil,
		false,
		false,
	)

	expected := ast.NewModule([]ast.Statement{fn}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParsePackageStatement(t *testing.T) {
	source := `package sample.core;
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	expected := ast.NewModule(
		[]ast.Statement{},
		nil,
		ast.NewPackageStatement(
			[]*ast.Identifier{ast.ID("sample"), ast.ID("core")},
			false,
		),
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseImportSelectors(t *testing.T) {
	source := `import alpha.beta.{Foo, Bar as B};
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	expected := ast.NewModule(
		[]ast.Statement{},
		[]*ast.ImportStatement{
			ast.NewImportStatement(
				[]*ast.Identifier{ast.ID("alpha"), ast.ID("beta")},
				false,
				[]*ast.ImportSelector{
					ast.NewImportSelector(ast.ID("Foo"), nil),
					ast.NewImportSelector(ast.ID("Bar"), ast.ID("B")),
				},
				nil,
			),
		},
		nil,
	)

	assertModulesEqual(t, expected, mod)
}

func TestParseWildcardImport(t *testing.T) {
	source := `import gamma.delta.*;
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	expected := ast.NewModule(
		[]ast.Statement{},
		[]*ast.ImportStatement{
			ast.NewImportStatement(
				[]*ast.Identifier{ast.ID("gamma"), ast.ID("delta")},
				true,
				nil,
				nil,
			),
		},
		nil,
	)

	assertModulesEqual(t, expected, mod)
}

func TestParseImportAlias(t *testing.T) {
	source := `import util.io as io_util;
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	expected := ast.NewModule(
		[]ast.Statement{},
		[]*ast.ImportStatement{
			ast.NewImportStatement(
				[]*ast.Identifier{ast.ID("util"), ast.ID("io")},
				false,
				nil,
				ast.ID("io_util"),
			),
		},
		nil,
	)

	assertModulesEqual(t, expected, mod)
}

func TestParseDynImportSelectors(t *testing.T) {
	source := `dynimport host.bindings as host;
dynimport host.bindings.{Device as HostDevice, Logger};
dynimport plugin.widgets.*;
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewDynImportStatement(
				[]*ast.Identifier{ast.ID("host"), ast.ID("bindings")},
				false,
				nil,
				ast.ID("host"),
			),
			ast.NewDynImportStatement(
				[]*ast.Identifier{ast.ID("host"), ast.ID("bindings")},
				false,
				[]*ast.ImportSelector{
					ast.NewImportSelector(ast.ID("Device"), ast.ID("HostDevice")),
					ast.NewImportSelector(ast.ID("Logger"), nil),
				},
				nil,
			),
			ast.NewDynImportStatement(
				[]*ast.Identifier{ast.ID("plugin"), ast.ID("widgets")},
				true,
				nil,
				nil,
			),
		},
		[]*ast.ImportStatement{},
		nil,
	)

	assertModulesEqual(t, expected, mod)
}

func TestParseProcExpressionForms(t *testing.T) {
	source := `fn worker(n: i32) -> i32 {
  n
}

handle := proc do {
  worker(1)
}

inline := proc {
  worker(2)
}

call := proc worker(3)
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	workerFn := ast.NewFunctionDefinition(
		ast.ID("worker"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("n"), ast.Ty("i32")),
		},
		ast.Block(ast.ID("n")),
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	handleAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("handle"),
		ast.NewProcExpression(
			ast.Block(
				ast.NewFunctionCall(
					ast.ID("worker"),
					[]ast.Expression{ast.Int(1)},
					nil,
					false,
				),
			),
		),
	)

	inlineAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("inline"),
		ast.NewProcExpression(
			ast.Block(
				ast.NewFunctionCall(
					ast.ID("worker"),
					[]ast.Expression{ast.Int(2)},
					nil,
					false,
				),
			),
		),
	)

	callAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("call"),
		ast.NewProcExpression(
			ast.NewFunctionCall(
				ast.ID("worker"),
				[]ast.Expression{ast.Int(3)},
				nil,
				false,
			),
		),
	)

	expected := ast.NewModule(
		[]ast.Statement{
			workerFn,
			handleAssign,
			inlineAssign,
			callAssign,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseSpawnExpressionForms(t *testing.T) {
	source := `fn task() -> i32 {
  1
}

fn run(n: i32) -> i32 {
  n
}

future := spawn do {
  task()
}

inline := spawn {
  task()
}

call := spawn run(3)
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	taskFn := ast.NewFunctionDefinition(
		ast.ID("task"),
		nil,
		ast.Block(ast.Int(1)),
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	runFn := ast.NewFunctionDefinition(
		ast.ID("run"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("n"), ast.Ty("i32")),
		},
		ast.Block(ast.ID("n")),
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	futureAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("future"),
		ast.NewSpawnExpression(
			ast.Block(
				ast.NewFunctionCall(
					ast.ID("task"),
					[]ast.Expression{},
					nil,
					false,
				),
			),
		),
	)

	inlineAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("inline"),
		ast.NewSpawnExpression(
			ast.Block(
				ast.NewFunctionCall(
					ast.ID("task"),
					[]ast.Expression{},
					nil,
					false,
				),
			),
		),
	)

	callAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("call"),
		ast.NewSpawnExpression(
			ast.NewFunctionCall(
				ast.ID("run"),
				[]ast.Expression{ast.Int(3)},
				nil,
				false,
			),
		),
	)

	expected := ast.NewModule(
		[]ast.Statement{
			taskFn,
			runFn,
			futureAssign,
			inlineAssign,
			callAssign,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseProcHelpers(t *testing.T) {
	source := `handle := proc do {
  proc_yield()
  0
}

proc_yield()
isCancelled := proc_cancelled()
proc_flush(handle)
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	handleAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("handle"),
		ast.NewProcExpression(
			ast.Block(
				ast.NewFunctionCall(
					ast.ID("proc_yield"),
					[]ast.Expression{},
					nil,
					false,
				),
				ast.Int(0),
			),
		),
	)

	yieldCall := ast.NewFunctionCall(
		ast.ID("proc_yield"),
		[]ast.Expression{},
		nil,
		false,
	)

	cancelAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("isCancelled"),
		ast.NewFunctionCall(
			ast.ID("proc_cancelled"),
			[]ast.Expression{},
			nil,
			false,
		),
	)

	flushCall := ast.NewFunctionCall(
		ast.ID("proc_flush"),
		[]ast.Expression{ast.ID("handle")},
		nil,
		false,
	)

	expected := ast.NewModule(
		[]ast.Statement{
			handleAssign,
			yieldCall,
			cancelAssign,
			flushCall,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseChannelAndMutexHelpers(t *testing.T) {
	source := `channel := __able_channel_new(2)
__able_channel_send(channel, 1)
value := __able_channel_receive(channel)
trySend := __able_channel_try_send(channel, 3)
maybeValue := __able_channel_try_receive(channel)
isClosed := __able_channel_is_closed(channel)
__able_channel_close(channel)

mutex := __able_mutex_new()
__able_mutex_lock(mutex)
__able_mutex_unlock(mutex)
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	channelAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("channel"),
		ast.NewFunctionCall(
			ast.ID("__able_channel_new"),
			[]ast.Expression{ast.Int(2)},
			nil,
			false,
		),
	)

	channelSend := ast.NewFunctionCall(
		ast.ID("__able_channel_send"),
		[]ast.Expression{ast.ID("channel"), ast.Int(1)},
		nil,
		false,
	)

	valueAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("value"),
		ast.NewFunctionCall(
			ast.ID("__able_channel_receive"),
			[]ast.Expression{ast.ID("channel")},
			nil,
			false,
		),
	)

	trySendAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("trySend"),
		ast.NewFunctionCall(
			ast.ID("__able_channel_try_send"),
			[]ast.Expression{ast.ID("channel"), ast.Int(3)},
			nil,
			false,
		),
	)

	tryReceiveAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("maybeValue"),
		ast.NewFunctionCall(
			ast.ID("__able_channel_try_receive"),
			[]ast.Expression{ast.ID("channel")},
			nil,
			false,
		),
	)

	isClosedAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("isClosed"),
		ast.NewFunctionCall(
			ast.ID("__able_channel_is_closed"),
			[]ast.Expression{ast.ID("channel")},
			nil,
			false,
		),
	)

	closeCall := ast.NewFunctionCall(
		ast.ID("__able_channel_close"),
		[]ast.Expression{ast.ID("channel")},
		nil,
		false,
	)

	mutexAssign := ast.NewAssignmentExpression(
		ast.AssignmentDeclare,
		ast.ID("mutex"),
		ast.NewFunctionCall(
			ast.ID("__able_mutex_new"),
			[]ast.Expression{},
			nil,
			false,
		),
	)

	mutexLock := ast.NewFunctionCall(
		ast.ID("__able_mutex_lock"),
		[]ast.Expression{ast.ID("mutex")},
		nil,
		false,
	)

	mutexUnlock := ast.NewFunctionCall(
		ast.ID("__able_mutex_unlock"),
		[]ast.Expression{ast.ID("mutex")},
		nil,
		false,
	)

	expected := ast.NewModule(
		[]ast.Statement{
			channelAssign,
			channelSend,
			valueAssign,
			trySendAssign,
			tryReceiveAssign,
			isClosedAssign,
			closeCall,
			mutexAssign,
			mutexLock,
			mutexUnlock,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseModuleImports(t *testing.T) {
	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	source := []byte(`package sample.core;

import alpha.beta.{Foo, Bar as B};
import gamma.delta.*;
import util.io;
dynimport host.bindings as host;

fn process(items) -> util.Strings {
  items + 1
}

fn use() {
  util.io.device
}

fn call_device() {
  util.io.device()
}

fn call_device_with_args(msg) {
  util.io.log(msg, 42)
}

fn transform(value) {
  identity<String>(value)
}

fn aggregate(items, seed) {
  items.reduce(seed) { acc, item => acc + item }
}

fn map_items(items) {
  items.map { item => item }
}
`)

	mod, err := p.ParseModule(source)
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewDynImportStatement(
				[]*ast.Identifier{ast.ID("host"), ast.ID("bindings")},
				false,
				nil,
				ast.ID("host"),
			),
			ast.NewFunctionDefinition(
				ast.ID("process"),
				[]*ast.FunctionParameter{
					ast.NewFunctionParameter(ast.ID("items"), nil),
				},
				ast.NewBlockExpression([]ast.Statement{
					ast.NewBinaryExpression("+", ast.ID("items"), ast.Int(1)),
				}),
				ast.Ty("util.Strings"),
				nil,
				nil,
				false,
				false,
			),
			ast.NewFunctionDefinition(
				ast.ID("use"),
				nil,
				ast.NewBlockExpression([]ast.Statement{
					ast.NewMemberAccessExpression(
						ast.NewMemberAccessExpression(ast.ID("util"), ast.ID("io")),
						ast.ID("device"),
					),
				}),
				nil,
				nil,
				nil,
				false,
				false,
			),
			ast.NewFunctionDefinition(
				ast.ID("call_device"),
				nil,
				ast.NewBlockExpression([]ast.Statement{
					ast.NewFunctionCall(
						ast.NewMemberAccessExpression(
							ast.NewMemberAccessExpression(ast.ID("util"), ast.ID("io")),
							ast.ID("device"),
						),
						[]ast.Expression{},
						nil,
						false,
					),
				}),
				nil,
				nil,
				nil,
				false,
				false,
			),
			ast.NewFunctionDefinition(
				ast.ID("call_device_with_args"),
				[]*ast.FunctionParameter{
					ast.NewFunctionParameter(ast.ID("msg"), nil),
				},
				ast.NewBlockExpression([]ast.Statement{
					ast.NewFunctionCall(
						ast.NewMemberAccessExpression(
							ast.NewMemberAccessExpression(ast.ID("util"), ast.ID("io")),
							ast.ID("log"),
						),
						[]ast.Expression{
							ast.ID("msg"),
							ast.Int(42),
						},
						nil,
						false,
					),
				}),
				nil,
				nil,
				nil,
				false,
				false,
			),
			ast.NewFunctionDefinition(
				ast.ID("transform"),
				[]*ast.FunctionParameter{
					ast.NewFunctionParameter(ast.ID("value"), nil),
				},
				ast.NewBlockExpression([]ast.Statement{
					ast.NewFunctionCall(
						ast.ID("identity"),
						[]ast.Expression{
							ast.ID("value"),
						},
						[]ast.TypeExpression{
							ast.Ty("String"),
						},
						false,
					),
				}),
				nil,
				nil,
				nil,
				false,
				false,
			),
			ast.NewFunctionDefinition(
				ast.ID("aggregate"),
				[]*ast.FunctionParameter{
					ast.NewFunctionParameter(ast.ID("items"), nil),
					ast.NewFunctionParameter(ast.ID("seed"), nil),
				},
				ast.NewBlockExpression([]ast.Statement{
					ast.NewFunctionCall(
						ast.NewMemberAccessExpression(
							ast.ID("items"),
							ast.ID("reduce"),
						),
						[]ast.Expression{
							ast.ID("seed"),
							ast.NewLambdaExpression(
								[]*ast.FunctionParameter{
									ast.NewFunctionParameter(ast.ID("acc"), nil),
									ast.NewFunctionParameter(ast.ID("item"), nil),
								},
								ast.NewBinaryExpression("+", ast.ID("acc"), ast.ID("item")),
								nil,
								nil,
								nil,
								false,
							),
						},
						nil,
						true,
					),
				}),
				nil,
				nil,
				nil,
				false,
				false,
			),
			ast.NewFunctionDefinition(
				ast.ID("map_items"),
				[]*ast.FunctionParameter{
					ast.NewFunctionParameter(ast.ID("items"), nil),
				},
				ast.NewBlockExpression([]ast.Statement{
					ast.NewFunctionCall(
						ast.NewMemberAccessExpression(
							ast.ID("items"),
							ast.ID("map"),
						),
						[]ast.Expression{
							ast.NewLambdaExpression(
								[]*ast.FunctionParameter{
									ast.NewFunctionParameter(ast.ID("item"), nil),
								},
								ast.ID("item"),
								nil,
								nil,
								nil,
								false,
							),
						},
						nil,
						true,
					),
				}),
				nil,
				nil,
				nil,
				false,
				false,
			),
		},
		[]*ast.ImportStatement{
			ast.NewImportStatement(
				[]*ast.Identifier{ast.ID("alpha"), ast.ID("beta")},
				false,
				[]*ast.ImportSelector{
					ast.NewImportSelector(ast.ID("Foo"), nil),
					ast.NewImportSelector(ast.ID("Bar"), ast.ID("B")),
				},
				nil,
			),
			ast.NewImportStatement(
				[]*ast.Identifier{ast.ID("gamma"), ast.ID("delta")},
				true,
				nil,
				nil,
			),
			ast.NewImportStatement(
				[]*ast.Identifier{ast.ID("util"), ast.ID("io")},
				false,
				nil,
				nil,
			),
		},
		ast.NewPackageStatement([]*ast.Identifier{ast.ID("sample"), ast.ID("core")}, false),
	)

	assertModulesEqual(t, expected, mod)
}

func TestParsePreludeAndExtern(t *testing.T) {
	source := `prelude go {
  package main
}

extern go fn host_fn(x: i32) -> i32 {
  return x
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	module, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	fn := ast.NewFunctionDefinition(
		ast.ID("host_fn"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("x"), ast.Ty("i32")),
		},
		nil,
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewPreludeStatement(ast.HostTargetGo, "package main"),
			ast.NewExternFunctionBody(ast.HostTargetGo, fn, "return x"),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, module)
}

func TestParseErrorHandlingExpressions(t *testing.T) {
	source := `value := (maybe() else { |err|
  err
} rescue {
  case err => err
} ensure {
  finalize()
})
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	module, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	maybeCall := ast.NewFunctionCall(ast.ID("maybe"), []ast.Expression{}, nil, false)
	handlerBlock := ast.NewBlockExpression([]ast.Statement{
		ast.ID("err"),
	})
	orElse := ast.NewOrElseExpression(maybeCall, handlerBlock, ast.ID("err"))

	rescueClause := ast.NewMatchClause(ast.ID("err"), ast.ID("err"), nil)
	rescueExpr := ast.NewRescueExpression(orElse, []*ast.MatchClause{rescueClause})

	ensureBlock := ast.NewBlockExpression([]ast.Statement{
		ast.NewFunctionCall(ast.ID("finalize"), []ast.Expression{}, nil, false),
	})
	ensureExpr := ast.NewEnsureExpression(rescueExpr, ensureBlock)

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewAssignmentExpression(
				ast.AssignmentDeclare,
				ast.ID("value"),
				ensureExpr,
			),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, module)
}

func TestParseUnionAndInterface(t *testing.T) {
	source := `union Number = i32, f64

interface Display {
  fn to_string() -> string;
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	module, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	unionDef := ast.NewUnionDefinition(
		ast.ID("Number"),
		[]ast.TypeExpression{
			ast.Ty("i32"),
			ast.Ty("f64"),
		},
		nil,
		nil,
		false,
	)

	signature := ast.NewFunctionSignature(
		ast.ID("to_string"),
		nil,
		ast.Ty("string"),
		nil,
		nil,
		nil,
	)

	interfaceDef := ast.NewInterfaceDefinition(
		ast.ID("Display"),
		[]*ast.FunctionSignature{signature},
		nil,
		nil,
		nil,
		nil,
		false,
	)

	expected := ast.NewModule(
		[]ast.Statement{
			unionDef,
			interfaceDef,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, module)
}

func TestParsePropagationExpression(t *testing.T) {
	source := `value := task!`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewAssignmentExpression(
				ast.AssignmentDeclare,
				ast.ID("value"),
				ast.NewPropagationExpression(ast.ID("task")),
			),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseBitwiseXorExpression(t *testing.T) {
	source := `value := left \xor right`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewAssignmentExpression(
				ast.AssignmentDeclare,
				ast.ID("value"),
				ast.NewBinaryExpression(`\xor`, ast.ID("left"), ast.ID("right")),
			),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseBitwiseXorAssignment(t *testing.T) {
	source := `value := 0
value \xor= mask
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewAssignmentExpression(ast.AssignmentDeclare, ast.ID("value"), ast.Int(0)),
			ast.NewAssignmentExpression(ast.AssignmentBitXor, ast.ID("value"), ast.ID("mask")),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseNumberLiteralPrefixes(t *testing.T) {
	source := `hex := 0xff
bin := 0b1010
oct := 0o77
typed := 1_u16
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	u16 := ast.IntegerTypeU16
	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewAssignmentExpression(ast.AssignmentDeclare, ast.ID("hex"), ast.Int(255)),
			ast.NewAssignmentExpression(ast.AssignmentDeclare, ast.ID("bin"), ast.Int(10)),
			ast.NewAssignmentExpression(ast.AssignmentDeclare, ast.ID("oct"), ast.Int(63)),
			ast.NewAssignmentExpression(ast.AssignmentDeclare, ast.ID("typed"), ast.IntTyped(1, &u16)),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParsePipeChainExpression(t *testing.T) {
	source := `value |> inc |> double`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	first := ast.NewBinaryExpression("|>", ast.ID("value"), ast.ID("inc"))
	chain := ast.NewBinaryExpression("|>", first, ast.ID("double"))
	expected := ast.NewModule([]ast.Statement{chain}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseFunctionTypeMultiParam(t *testing.T) {
	source := `fn register(handler: (i32, string) -> void) {}`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	paramType := ast.NewFunctionTypeExpression(
		[]ast.TypeExpression{ast.Ty("i32"), ast.Ty("string")},
		ast.Ty("void"),
	)
	fn := ast.NewFunctionDefinition(
		ast.ID("register"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("handler"), paramType),
		},
		ast.NewBlockExpression([]ast.Statement{}),
		nil,
		nil,
		nil,
		false,
		false,
	)
	expected := ast.NewModule(
		[]ast.Statement{fn},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseIteratorLiteral(t *testing.T) {
	source := `items := Iterator {
 	next := 1
 	next
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	iteratorBody := []ast.Statement{
		ast.NewAssignmentExpression(
			ast.AssignmentDeclare,
			ast.ID("next"),
			ast.Int(1),
		),
		ast.ID("next"),
	}

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewAssignmentExpression(
				ast.AssignmentDeclare,
				ast.ID("items"),
				ast.NewIteratorLiteral(iteratorBody),
			),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParsePreludeExternMultiTarget(t *testing.T) {
	source := `prelude typescript {
  export const value = 1;
}

extern python fn run() {
  pass
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	fn := ast.NewFunctionDefinition(
		ast.ID("run"),
		nil,
		nil,
		nil,
		nil,
		nil,
		false,
		false,
	)

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewPreludeStatement(ast.HostTargetTypeScript, "export const value = 1;"),
			ast.NewExternFunctionBody(ast.HostTargetPython, fn, "pass"),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseInterfaceCompositeGenerics(t *testing.T) {
	source := `interface Repository<T> where T: Display = Iterable string + Sized
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	genericParam := ast.NewGenericParameter(ast.ID("T"), nil)
	whereConstraint := ast.NewWhereClauseConstraint(
		ast.ID("T"),
		[]*ast.InterfaceConstraint{
			ast.NewInterfaceConstraint(ast.TyID(ast.ID("Display"))),
		},
	)

	iterBase := ast.NewGenericTypeExpression(ast.Ty("Iterable"), []ast.TypeExpression{ast.Ty("string")})

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewInterfaceDefinition(
				ast.ID("Repository"),
				[]*ast.FunctionSignature{},
				[]*ast.GenericParameter{genericParam},
				nil,
				[]*ast.WhereClauseConstraint{whereConstraint},
				[]ast.TypeExpression{
					iterBase,
					ast.Ty("Sized"),
				},
				false,
			),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseInterfaceCompositeNestedGenerics(t *testing.T) {
	source := `interface Feed<T> = Iterable (Option string) + Display + Sized
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	optionString := ast.NewGenericTypeExpression(ast.Ty("Option"), []ast.TypeExpression{ast.Ty("string")})
	iterOptionString := ast.NewGenericTypeExpression(ast.Ty("Iterable"), []ast.TypeExpression{optionString})

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewInterfaceDefinition(
				ast.ID("Feed"),
				[]*ast.FunctionSignature{},
				[]*ast.GenericParameter{
					ast.NewGenericParameter(ast.ID("T"), nil),
				},
				nil,
				nil,
				[]ast.TypeExpression{
					iterOptionString,
					ast.Ty("Display"),
					ast.Ty("Sized"),
				},
				false,
			),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseLambdaExpressionLiteral(t *testing.T) {
	source := `fn make(offset) {
  handler := { value => value + offset }
  handler(3)
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	handlerLambda := ast.Lam(
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("value"), nil),
		},
		ast.Bin("+", ast.ID("value"), ast.ID("offset")),
	)

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewFunctionDefinition(
				ast.ID("make"),
				[]*ast.FunctionParameter{
					ast.NewFunctionParameter(ast.ID("offset"), nil),
				},
				ast.Block(
					ast.Assign(ast.ID("handler"), handlerLambda),
					ast.CallExpr(ast.ID("handler"), ast.Int(3)),
				),
				nil,
				nil,
				nil,
				false,
				false,
			),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseTrailingLambdaCallSimple(t *testing.T) {
	source := `fn apply(items) {
  items.map { item => item + 1 }
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	lambda := ast.Lam(
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("item"), nil),
		},
		ast.Bin("+", ast.ID("item"), ast.Int(1)),
	)

	call := ast.NewFunctionCall(
		ast.NewMemberAccessExpression(ast.ID("items"), ast.ID("map")),
		[]ast.Expression{
			lambda,
		},
		nil,
		true,
	)

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewFunctionDefinition(
				ast.ID("apply"),
				[]*ast.FunctionParameter{
					ast.NewFunctionParameter(ast.ID("items"), nil),
				},
				ast.Block(
					call,
				),
				nil,
				nil,
				nil,
				false,
				false,
			),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseWhileLoopWithBreakAndContinue(t *testing.T) {
	source := `value := 0
while value < 10 {
  value += 1
  if value % 2 == 0 {
    continue
  }
  if value == 5 {
    break value
  }
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	assignment := ast.Assign(ast.ID("value"), ast.Int(0))
	condition := ast.Bin("<", ast.ID("value"), ast.Int(10))
	increment := ast.AssignOp(ast.AssignmentAdd, ast.ID("value"), ast.Int(1))

	modCondition := ast.Bin("==",
		ast.Bin("%", ast.ID("value"), ast.Int(2)),
		ast.Int(0),
	)

	continueBlock := ast.Block(ast.NewContinueStatement(nil))
	continueIf := ast.IfExpr(modCondition, continueBlock)
	continueIf.OrClauses = []*ast.OrClause{}

	breakCondition := ast.Bin("==", ast.ID("value"), ast.Int(5))
	breakBlock := ast.Block(ast.NewBreakStatement(nil, ast.ID("value")))
	breakIf := ast.IfExpr(breakCondition, breakBlock)
	breakIf.OrClauses = []*ast.OrClause{}

	whileBody := ast.Block(
		increment,
		continueIf,
		breakIf,
	)

	whileLoop := ast.NewWhileLoop(condition, whileBody)

	expected := ast.NewModule(
		[]ast.Statement{
			assignment,
			whileLoop,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseForLoopWithAssignment(t *testing.T) {
	source := `items := [1, 2, 3]
sum := 0
for item in items {
  sum = sum + item
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	itemsAssign := ast.Assign(ast.ID("items"), ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3)))
	sumAssign := ast.Assign(ast.ID("sum"), ast.Int(0))

	body := ast.Block(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("sum"),
			ast.Bin("+", ast.ID("sum"), ast.ID("item")),
		),
	)

	forLoop := ast.ForIn("item", ast.ID("items"), body.Body...)

	expected := ast.NewModule(
		[]ast.Statement{
			itemsAssign,
			sumAssign,
			forLoop,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseReturnStatements(t *testing.T) {
	source := `fn check(flag) {
  if flag {
    return
  }
  return 42
}
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	ifBody := ast.Block(ast.NewReturnStatement(nil))
	ifExpr := ast.IfExpr(ast.ID("flag"), ifBody)
	ifExpr.OrClauses = []*ast.OrClause{}

	fnBody := ast.Block(
		ifExpr,
		ast.NewReturnStatement(ast.Int(42)),
	)

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewFunctionDefinition(
				ast.ID("check"),
				[]*ast.FunctionParameter{
					ast.NewFunctionParameter(ast.ID("flag"), nil),
				},
				fnBody,
				nil,
				nil,
				nil,
				false,
				false,
			),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseStructDefinitions(t *testing.T) {
	source := `struct Point {
  x: i32,
  y: i32,
}

struct Vec3(i32, i32, i32)
`

	p, err := parser.NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	pointStruct := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
			ast.FieldDef(ast.Ty("i32"), "y"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	vec3Struct := ast.StructDef(
		"Vec3",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), nil),
			ast.FieldDef(ast.Ty("i32"), nil),
			ast.FieldDef(ast.Ty("i32"), nil),
		},
		ast.StructKindPositional,
		nil,
		nil,
		false,
	)

	expected := ast.NewModule(
		[]ast.Statement{
			pointStruct,
			vec3Struct,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

var (
	concurrencyFixtures = []fixtureCase{
		{
			name: "concurrency/proc_cancel_value",
			source: `handle := proc do {
	  0
	}

	_cancelResult := handle.cancel()
	result := handle.value()
	result
	`,
		},
		{
			name: "concurrency/future_memoization",
			source: `count := 0
	future := spawn do {
	  count += 1
	  1
	}

	first := future.value()
	second := future.value()
	count
	`,
		},
		{
			name: "concurrency/proc_cancelled_helper",
			source: `trace := ""
	handle := proc do {
	  trace = trace + "A"
	  handle.cancel()
	  if proc_cancelled() {
	    trace = trace + "C"
	  }
	  0
	}

	_result := handle.value()
	trace
	`,
		},
	}

	controlFlowFixtures = []fixtureCase{
		{
			name: "control/if_else_branch",
			source: `if false {
	  print("true")
	} else {
	  print("false")
	}

	"after"
	`,
		},
		{
			name: "control/for_range_break",
			source: `sum := 0
	for n in 0..5 {
	  do {
	    sum = sum + n
	    if n >= 2 {
	      break sum
	    }
	  }
	}
	`,
		},
		{
			name: "control/for_continue",
			source: `sum := 0
	items := [1, 2, 3]
	for n in items {
	  if n == 2 {
	    continue
	  }
	  sum = sum + n
	}

	sum
	`,
		},
		{
			name: "control/while_sum",
			source: `sum := 0
	i := 0
	limit := 3
	while i < limit {
	  sum = sum + i
	  i = i + 1
	}

	sum
	`,
		},
	}

	typeFixturePaths = []string{
		"types/generic_type_expression",
		"types/function_type_expression",
		"types/nullable_type_expression",
		"types/result_type_expression",
		"types/union_type_expression",
		"types/generic_where_constraint",
	}
)

func TestParseConcurrencyFixtures(t *testing.T) {
	runFixtureCases(t, concurrencyFixtures)
}

func TestParseControlFlowFixtures(t *testing.T) {
	runFixtureCases(t, controlFlowFixtures)
}

func TestParseTypeExpressionFixtures(t *testing.T) {
	t.Helper()
	t.Logf("TODO: enable once type expression parser is in place for fixtures: %v", typeFixturePaths)
	t.Skip("type expression mapping not implemented yet")
}
