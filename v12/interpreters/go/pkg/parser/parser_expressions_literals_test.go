package parser

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

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

	p, err := NewModuleParser()
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

func TestParseMatchExpression(t *testing.T) {
	source := `struct Point {
  x: i32,
  y: i32,
}

fn classify(point: Point) -> i32 {
  point match {
    case Point { x::a, y::b } if b > a => a + b,
    case Point { x::value, y::value } => value,
    case _ => 0,
  }
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

func TestParseCastExpression(t *testing.T) {
	source := `fn casty() {
  wrap := 256_u16 as u8
  trunc := 3.9 as i32
  nested := 1_u16 as u8 as i32
  point := Point { x: 1, y: 2 }
  show := point as Display
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

	u16 := ast.IntegerTypeU16

	wrap := ast.Assign(
		ast.ID("wrap"),
		ast.NewTypeCastExpression(ast.IntTyped(256, &u16), ast.Ty("u8")),
	)
	trunc := ast.Assign(
		ast.ID("trunc"),
		ast.NewTypeCastExpression(ast.Flt(3.9), ast.Ty("i32")),
	)
	nested := ast.Assign(
		ast.ID("nested"),
		ast.NewTypeCastExpression(
			ast.NewTypeCastExpression(ast.IntTyped(1, &u16), ast.Ty("u8")),
			ast.Ty("i32"),
		),
	)
	point := ast.Assign(
		ast.ID("point"),
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(1), "x"),
				ast.FieldInit(ast.Int(2), "y"),
			},
			false,
			"Point",
			nil,
			nil,
		),
	)
	show := ast.Assign(
		ast.ID("show"),
		ast.NewTypeCastExpression(ast.ID("point"), ast.Ty("Display")),
	)

	body := ast.Block(
		wrap,
		trunc,
		nested,
		point,
		show,
	)

	expected := ast.NewModule([]ast.Statement{
		ast.NewFunctionDefinition(
			ast.ID("casty"),
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

func TestParseCastExpressionNewlineTermination(t *testing.T) {
	source := `fn casty() {
  casted := 3.7 as i32
  print(casted)
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

	casted := ast.Assign(
		ast.ID("casted"),
		ast.NewTypeCastExpression(ast.Flt(3.7), ast.Ty("i32")),
	)
	printCall := ast.NewFunctionCall(
		ast.ID("print"),
		[]ast.Expression{ast.ID("casted")},
		nil,
		false,
	)

	body := ast.Block(
		casted,
		printCall,
	)

	fn := ast.NewFunctionDefinition(
		ast.ID("casty"),
		nil,
		body,
		nil,
		nil,
		nil,
		false,
		false,
	)

	expected := ast.NewModule([]ast.Statement{
		fn,
	}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParsePropagationAndOrElse(t *testing.T) {
source := `fn handlers(opt: ?i32, res: !i32) -> !i32 {
 	value := opt! or { 0 }
 	processed := res! or { err => err }
 	processed
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
  message: String,
}

impl Error for MyError {
  fn message(self: Self) -> String { "boom" }
  fn cause(self: Self) -> ?Error { nil }
}

fn guard() -> String {
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

	p, err := NewModuleParser()
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
			ast.NewStructFieldDefinition(ast.Ty("String"), ast.ID("message")),
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
		ast.Ty("String"),
		nil,
		nil,
		false,
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
		false,
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
		ast.Ty("String"),
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

func TestParseIfElsifExpression(t *testing.T) {
	source := `fn grade(score: i32) -> String {
  if score >= 90 { "A" }
  elsif score >= 80 { "B" }
  else { "C" }
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

	firstCond := ast.NewBinaryExpression(">=", ast.ID("score"), ast.Int(90))
	firstBody := ast.Block(ast.Str("A"))
	secondCond := ast.NewBinaryExpression(">=", ast.ID("score"), ast.Int(80))
	secondBody := ast.Block(ast.Str("B"))
	defaultBody := ast.Block(ast.Str("C"))

	ifExpr := ast.NewIfExpression(
		firstCond,
		firstBody,
		[]*ast.ElseIfClause{
			ast.NewElseIfClause(secondBody, secondCond),
		},
		defaultBody,
	)

	fnBody := ast.Block(ifExpr)

	fn := ast.NewFunctionDefinition(
		ast.ID("grade"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("score"), ast.Ty("i32")),
		},
		fnBody,
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)

	expected := ast.NewModule([]ast.Statement{fn}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParsePropagationExpression(t *testing.T) {
	source := `value := task!`

	p, err := NewModuleParser()
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

func TestParseNumberLiteralPrefixes(t *testing.T) {
	source := `hex := 0xff
bin := 0b1010
oct := 0o77
typed := 1_u16
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

	p, err := NewModuleParser()
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

func TestParsePipeWithPlaceholderSteps(t *testing.T) {
	source := `result := (value |> (@ + 2) |> add(@, 3) |> double)`

	p, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	step1 := ast.NewBinaryExpression("|>", ast.ID("value"), ast.NewBinaryExpression("+", ast.Placeholder(), ast.Int(2)))
	step2 := ast.NewBinaryExpression("|>", step1, ast.Call("add", ast.Placeholder(), ast.Int(3)))
	step3 := ast.NewBinaryExpression("|>", step2, ast.ID("double"))

	expected := ast.NewModule(
		[]ast.Statement{
			ast.Assign(ast.ID("result"), step3),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseRangeExpressions(t *testing.T) {
	source := `exclusive := 0...5
inclusive := 0..5
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

	exclusiveRange := ast.NewRangeExpression(ast.Int(0), ast.Int(5), false)
	inclusiveRange := ast.NewRangeExpression(ast.Int(0), ast.Int(5), true)

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewAssignmentExpression(
				ast.AssignmentDeclare,
				ast.ID("exclusive"),
				exclusiveRange,
			),
			ast.NewAssignmentExpression(
				ast.AssignmentDeclare,
				ast.ID("inclusive"),
				inclusiveRange,
			),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseMemberAccessChaining(t *testing.T) {
	source := `fn update(player) {
  player.position.x = player.position().x + 1
  player.stats.speed().value()
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

	positionAccess := ast.NewMemberAccessExpression(ast.ID("player"), ast.ID("position"))
	left := ast.NewMemberAccessExpression(positionAccess, ast.ID("x"))

	positionCall := ast.NewFunctionCall(
		ast.NewMemberAccessExpression(ast.ID("player"), ast.ID("position")),
		[]ast.Expression{},
		nil,
		false,
	)
	right := ast.NewBinaryExpression(
		"+",
		ast.NewMemberAccessExpression(positionCall, ast.ID("x")),
		ast.Int(1),
	)

	assign := ast.NewAssignmentExpression(ast.AssignmentAssign, left, right)

	statsAccess := ast.NewMemberAccessExpression(ast.ID("player"), ast.ID("stats"))
	speedCall := ast.NewFunctionCall(
		ast.NewMemberAccessExpression(statsAccess, ast.ID("speed")),
		[]ast.Expression{},
		nil,
		false,
	)
	valueCall := ast.NewFunctionCall(
		ast.NewMemberAccessExpression(speedCall, ast.ID("value")),
		[]ast.Expression{},
		nil,
		false,
	)

	fn := ast.NewFunctionDefinition(
		ast.ID("update"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("player"), nil),
		},
		ast.Block(
			assign,
			valueCall,
		),
		nil,
		nil,
		nil,
		false,
		false,
	)

	expected := ast.NewModule([]ast.Statement{fn}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseIndexExpressions(t *testing.T) {
	source := `first := items[0]
computed := items[getIndex()]
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

	firstIndex := ast.NewIndexExpression(ast.ID("items"), ast.Int(0))
	computedIndex := ast.NewIndexExpression(
		ast.ID("items"),
		ast.NewFunctionCall(ast.ID("getIndex"), []ast.Expression{}, nil, false),
	)

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewAssignmentExpression(ast.AssignmentDeclare, ast.ID("first"), firstIndex),
			ast.NewAssignmentExpression(ast.AssignmentDeclare, ast.ID("computed"), computedIndex),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseHexLiteralWithEDigit(t *testing.T) {
	source := `fn hex() {
  value := 0xE0
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

	body := ast.Block(ast.Assign(ast.ID("value"), ast.Int(0xE0)))
	expected := ast.NewModule([]ast.Statement{
		ast.NewFunctionDefinition(
			ast.ID("hex"),
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

func TestParseBasePrefixedLiteralWithExponentMarker(t *testing.T) {
	cases := []string{
		`fn bad() { value := 0b1e0 }`,
		`fn alsoBad() { value := 0o7E1 }`,
	}

	for _, source := range cases {
		p, err := NewModuleParser()
		if err != nil {
			t.Fatalf("NewModuleParser error: %v", err)
		}
		mod, parseErr := p.ParseModule([]byte(source))
		if parseErr == nil {
			t.Fatalf("expected parse error for %q, got module %#v", source, mod)
		}
		p.Close()
	}
}
