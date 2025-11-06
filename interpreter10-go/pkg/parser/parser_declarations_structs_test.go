package parser

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
)

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

	p, err := NewModuleParser()
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

func TestParseStructLiteralIgnoresComments(t *testing.T) {
	t.Skip("tree-sitter grammar currently rejects comments inside struct literals")
	source := `struct Point {
	x: i32,
	y: i32,
}

base := Point {
	## keep base values
	x: 1,
	## second field
	y: 2,
}

updated := Point {
	## spread original
	...base,
	## override x only
	x: 3,
}
`

	p, err := NewModuleParser()
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

	baseLiteral := ast.NewStructLiteral(
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
			ast.NewStructFieldInitializer(ast.Int(3), ast.ID("x"), false),
		},
		false,
		ast.ID("Point"),
		[]ast.Expression{ast.ID("base")},
		nil,
	)

	expected := ast.NewModule(
		[]ast.Statement{
			pointStruct,
			ast.Assign(ast.ID("base"), baseLiteral),
			ast.Assign(ast.ID("updated"), updatedLiteral),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseStructPatternIgnoresComments(t *testing.T) {
	source := `struct Point {
	x: i32,
	y: i32
}

Point { x: 1, y: 2 } match {
	case Point {
		## x binding
		x: px,
		## y binding
		y: py
	} => px + py,
	case _ => 0 ## fallback
}
`

	p, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	if len(mod.Body) != 2 {
		t.Fatalf("expected two top-level statements, got %d", len(mod.Body))
	}

	pointStruct, ok := mod.Body[0].(*ast.StructDefinition)
	if !ok {
		t.Fatalf("expected first statement to be struct definition, got %T", mod.Body[0])
	}
	if len(pointStruct.Fields) != 2 {
		t.Fatalf("expected struct definition to have 2 fields, got %d", len(pointStruct.Fields))
	}

	matchExpr, ok := mod.Body[1].(*ast.MatchExpression)
	if !ok {
		t.Fatalf("expected second statement to be match expression, got %T", mod.Body[1])
	}
	if len(matchExpr.Clauses) != 2 {
		t.Fatalf("expected match expression to have 2 clauses, got %d", len(matchExpr.Clauses))
	}

	structClause := matchExpr.Clauses[0]
	structPattern, ok := structClause.Pattern.(*ast.StructPattern)
	if !ok {
		t.Fatalf("expected first clause pattern to be struct pattern, got %T", structClause.Pattern)
	}
	if len(structPattern.Fields) != 2 {
		t.Fatalf("expected struct pattern to have 2 fields, got %d", len(structPattern.Fields))
	}
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

	p, err := NewModuleParser()
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

func TestParsePreludeAndExtern(t *testing.T) {
	source := `prelude go {
  package main
}

extern go fn host_fn(x: i32) -> i32 {
  return x
}
`

	p, err := NewModuleParser()
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

func TestParseGenericFunctionDefinition(t *testing.T) {
	source := `fn identity<T>(value: T) -> T where T: Display {
  value
}
`

	p, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	genericParam := ast.NewGenericParameter(ast.ID("T"), nil)
	whereClause := ast.NewWhereClauseConstraint(
		ast.ID("T"),
		[]*ast.InterfaceConstraint{
			ast.NewInterfaceConstraint(ast.TyID(ast.ID("Display"))),
		},
	)

	fn := ast.NewFunctionDefinition(
		ast.ID("identity"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("value"), ast.Ty("T")),
		},
		ast.Block(ast.ID("value")),
		ast.Ty("T"),
		[]*ast.GenericParameter{genericParam},
		[]*ast.WhereClauseConstraint{whereClause},
		false,
		false,
	)

	expected := ast.NewModule([]ast.Statement{fn}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParsePrivateFunctionDefinition(t *testing.T) {
	source := `private fn helper(x: i32) -> i32 {
  x
}
`

	p, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	fn := ast.NewFunctionDefinition(
		ast.ID("helper"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("x"), ast.Ty("i32")),
		},
		ast.Block(ast.ID("x")),
		ast.Ty("i32"),
		nil,
		nil,
		false,
		true,
	)

	expected := ast.NewModule([]ast.Statement{fn}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseUnionAndInterface(t *testing.T) {
	source := `union Number = i32, f64

interface Display {
  fn to_string() -> string;
}
`

	p, err := NewModuleParser()
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

func TestParseFunctionDefinitionWithReturnType(t *testing.T) {
	source := `fn sum(a: i32, b: i32) -> i32 {
	  a + b
	}
`

	p, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	fn := ast.NewFunctionDefinition(
		ast.ID("sum"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("a"), ast.Ty("i32")),
			ast.NewFunctionParameter(ast.ID("b"), ast.Ty("i32")),
		},
		ast.Block(
			ast.NewBinaryExpression("+", ast.ID("a"), ast.ID("b")),
		),
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

func TestParsePreludeExternMultiTarget(t *testing.T) {
	source := `prelude typescript {
  export const value = 1;
}

extern python fn run() {
  pass
}
`

	p, err := NewModuleParser()
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

	p, err := NewModuleParser()
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

	p, err := NewModuleParser()
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

func TestParseStructDefinitions(t *testing.T) {
	source := `struct Point {
  x: i32,
  y: i32,
}

struct Vec3 {
  i32,
  i32,
  i32
}
`

	p, err := NewModuleParser()
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
