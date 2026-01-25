package parser

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestParseImplicitMethods(t *testing.T) {
	source := "struct Counter {\n  value: i32,\n}\n\nmethods Counter {\n  fn #increment() {\n    #value = #value + 1\n  }\n\n  fn #add(amount: i32) {\n    #value = #value + amount\n  }\n}\n\nimpl Display for Counter {\n  fn #to_string() -> String {\n    `Counter(${#value})`\n  }\n}\n"

	p, err := NewModuleParser()
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
				ast.Ty("String"),
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

func TestParseBreakpointExpression(t *testing.T) {
	source := `fn debug(value: i32) -> i32 {
  breakpoint 'trace {
    log(value)
    value + 1
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

func TestParseImplInterfaceArgsParentheses(t *testing.T) {
	source := "impl Foo Array i64 for Bar {}"

	p, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}
	if len(mod.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(mod.Body))
	}
	implDef, ok := mod.Body[0].(*ast.ImplementationDefinition)
	if !ok {
		t.Fatalf("expected implementation definition, got %T", mod.Body[0])
	}
	if len(implDef.InterfaceArgs) != 2 {
		t.Fatalf("expected 2 interface args, got %d", len(implDef.InterfaceArgs))
	}

	valid := "impl Foo (Array i64) for Bar {}"
	validMod, err := p.ParseModule([]byte(valid))
	if err != nil {
		t.Fatalf("expected parenthesized interface arg to parse, got: %v", err)
	}
	if len(validMod.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(validMod.Body))
	}
	validImpl, ok := validMod.Body[0].(*ast.ImplementationDefinition)
	if !ok {
		t.Fatalf("expected implementation definition, got %T", validMod.Body[0])
	}
	if len(validImpl.InterfaceArgs) != 1 {
		t.Fatalf("expected 1 interface arg, got %d", len(validImpl.InterfaceArgs))
	}
}

func TestParseBreakpointWithLabel(t *testing.T) {
	source := `fn debug() {
  breakpoint 'trace {
    1
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

	breakpointExpr := ast.NewBreakpointExpression(
		ast.ID("trace"),
		ast.Block(ast.Int(1)),
	)

	fn := ast.NewFunctionDefinition(
		ast.ID("debug"),
		nil,
		ast.Block(breakpointExpr),
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

func TestParseFunctionCallWithTypeArguments(t *testing.T) {
	source := `result := identity<String, Option<i32>>(value)
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

	optionI32 := ast.NewGenericTypeExpression(ast.Ty("Option"), []ast.TypeExpression{ast.Ty("i32")})
	call := ast.NewFunctionCall(
		ast.ID("identity"),
		[]ast.Expression{ast.ID("value")},
		[]ast.TypeExpression{
			ast.Ty("String"),
			optionI32,
		},
		false,
	)

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewAssignmentExpression(
				ast.AssignmentDeclare,
				ast.ID("result"),
				call,
			),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseFunctionTypeMultiParam(t *testing.T) {
	source := `fn register(handler: (i32, String) -> void) {}`

	p, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	paramType := ast.NewFunctionTypeExpression(
		[]ast.TypeExpression{ast.Ty("i32"), ast.Ty("String")},
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

	p, err := NewModuleParser()
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

func TestParseIteratorLiteralWithAnnotationAndBinding(t *testing.T) {
	source := `iter := Iterator i32 {
  driver =>
    driver.yield(1)
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

	if len(mod.Body) != 1 {
		t.Fatalf("expected module body len 1, got %d", len(mod.Body))
	}
	assign, ok := mod.Body[0].(*ast.AssignmentExpression)
	if !ok {
		t.Fatalf("expected first statement to be assignment, got %T", mod.Body[0])
	}
	literal, ok := assign.Right.(*ast.IteratorLiteral)
	if !ok {
		t.Fatalf("expected assignment RHS to be iterator literal, got %T", assign.Right)
	}
	if literal.Binding == nil || literal.Binding.Name != "driver" {
		t.Fatalf("expected iterator binding 'driver', got %#v", literal.Binding)
	}
	simple, ok := literal.ElementType.(*ast.SimpleTypeExpression)
	if !ok {
		t.Fatalf("expected simple element type, got %T", literal.ElementType)
	}
	if simple.Name == nil || simple.Name.Name != "i32" {
		t.Fatalf("expected element type i32, got %#v", simple.Name)
	}
}

func TestParseLambdaExpressionLiteral(t *testing.T) {
	source := `fn make(offset) {
  handler := { value => value + offset }
  handler(3)
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

	p, err := NewModuleParser()
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

	p, err := NewModuleParser()
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
	continueIf.ElseIfClauses = []*ast.ElseIfClause{}

	breakCondition := ast.Bin("==", ast.ID("value"), ast.Int(5))
	breakBlock := ast.Block(ast.NewBreakStatement(nil, ast.ID("value")))
	breakIf := ast.IfExpr(breakCondition, breakBlock)
	breakIf.ElseIfClauses = []*ast.ElseIfClause{}

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

func TestParseLoopExpression(t *testing.T) {
	source := `result := loop {
  if counter >= 2 {
    break counter
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

	breakBlock := ast.Block(ast.NewBreakStatement(nil, ast.ID("counter")))
	breakIf := ast.IfExpr(ast.Bin(">=", ast.ID("counter"), ast.Int(2)), breakBlock)
	breakIf.ElseIfClauses = []*ast.ElseIfClause{}

	loopExpr := ast.Loop(breakIf)
	expected := ast.NewModule(
		[]ast.Statement{
			ast.Assign(ast.ID("result"), loopExpr),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseLoopExpressionStatement(t *testing.T) {
	source := `counter := 3
loop {
  counter = (counter - 1)
  if counter < 0 {
    break
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

	assign := ast.Assign(ast.ID("counter"), ast.Int(3))

	decrement := ast.AssignOp(
		ast.AssignmentAssign,
		ast.ID("counter"),
		ast.Bin("-", ast.ID("counter"), ast.Int(1)),
	)

	breakBlock := ast.Block(ast.NewBreakStatement(nil, nil))
	breakIf := ast.IfExpr(ast.Bin("<", ast.ID("counter"), ast.Int(0)), breakBlock)
	breakIf.ElseIfClauses = []*ast.ElseIfClause{}

	loopExpr := ast.Loop(decrement, breakIf)

	expected := ast.NewModule(
		[]ast.Statement{
			assign,
			loopExpr,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseLoopExpressionStatementWithoutParens(t *testing.T) {
	source := `counter := 3
loop {
  counter = counter - 1
  if counter < 0 {
    break
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

	assign := ast.Assign(ast.ID("counter"), ast.Int(3))

	decrement := ast.AssignOp(
		ast.AssignmentAssign,
		ast.ID("counter"),
		ast.Bin("-", ast.ID("counter"), ast.Int(1)),
	)

	breakBlock := ast.Block(ast.NewBreakStatement(nil, nil))
	breakIf := ast.IfExpr(ast.Bin("<", ast.ID("counter"), ast.Int(0)), breakBlock)
	breakIf.ElseIfClauses = []*ast.ElseIfClause{}

	loopExpr := ast.Loop(decrement, breakIf)

	expected := ast.NewModule(
		[]ast.Statement{
			assign,
			loopExpr,
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

	p, err := NewModuleParser()
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

	p, err := NewModuleParser()
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
	ifExpr.ElseIfClauses = []*ast.ElseIfClause{}

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
