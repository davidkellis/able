package parser

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestParseModuleIgnoresComments(t *testing.T) {
	mp, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser: %v", err)
	}
	t.Cleanup(func() { mp.Close() })

	source := []byte(`
## Leading comment
package sample

## function comment
fn main() -> void { }
`)

	mod, err := mp.ParseModule(source)
	if err != nil {
		t.Fatalf("ParseModule returned error: %v", err)
	}
	if mod == nil {
		t.Fatalf("ParseModule returned nil module")
	}
	if mod.Package == nil || len(mod.Package.NamePath) == 0 || mod.Package.NamePath[0] == nil || mod.Package.NamePath[0].Name != "sample" {
		t.Fatalf("expected package sample, got %#v", mod.Package)
	}
	if len(mod.Body) != 1 {
		t.Fatalf("expected single statement in module body, got %d", len(mod.Body))
	}
	if _, ok := mod.Body[0].(*ast.FunctionDefinition); !ok {
		t.Fatalf("expected first body element to be FunctionDefinition, got %T", mod.Body[0])
	}
}

func TestParseModuleAnnotatesSpans(t *testing.T) {
	mp, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser: %v", err)
	}
	t.Cleanup(func() { mp.Close() })

	source := []byte("package demo\n\nfn add(x: i32, y: i32) -> i32 {\n  return x + y\n}\n")

	mod, err := mp.ParseModule(source)
	if err != nil {
		t.Fatalf("ParseModule returned error: %v", err)
	}
	if mod == nil {
		t.Fatalf("ParseModule returned nil module")
	}

	checkSpan(t, "module", mod.Span(), 1, 1, 6, 1)

	if len(mod.Body) != 1 {
		t.Fatalf("expected single statement in module body, got %d", len(mod.Body))
	}

	fn, ok := mod.Body[0].(*ast.FunctionDefinition)
	if !ok {
		t.Fatalf("expected first body element to be FunctionDefinition, got %T", mod.Body[0])
	}
	checkSpan(t, "function", fn.Span(), 3, 1, 5, 2)

	if len(fn.Params) != 2 {
		t.Fatalf("expected two parameters, got %d", len(fn.Params))
	}
	checkSpan(t, "param x", fn.Params[0].Span(), 3, 8, 3, 14)
	checkSpan(t, "param y", fn.Params[1].Span(), 3, 16, 3, 22)

	if fn.Body == nil || len(fn.Body.Body) != 1 {
		t.Fatalf("expected function body to contain a single statement")
	}
	ret, ok := fn.Body.Body[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected return statement, got %T", fn.Body.Body[0])
	}
	checkSpan(t, "return", ret.Span(), 4, 3, 4, 15)

	binary, ok := ret.Argument.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected return expression to be binary expression, got %T", ret.Argument)
	}
	checkSpan(t, "binary expression", binary.Span(), 4, 10, 4, 15)
	checkSpan(t, "binary lhs", binary.Left.Span(), 4, 10, 4, 11)
	checkSpan(t, "binary rhs", binary.Right.Span(), 4, 14, 4, 15)
}

func TestParseAnnotatesPatternAndTypeSpans(t *testing.T) {
	mp, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser: %v", err)
	}
	t.Cleanup(func() { mp.Close() })

	source := []byte(`package sample

struct Point {
  x: i32,
  y: i32,
}

fn extract(point: Point) -> i32 {
  point match {
    case Point { x: a, y: b } => a + b,
    case _ => 0
  }
}
`)

	mod, err := mp.ParseModule(source)
	if err != nil {
		t.Fatalf("ParseModule returned error: %v", err)
	}
	if mod == nil || len(mod.Body) < 2 {
		t.Fatalf("expected struct and function definitions in module body")
	}

	fn, ok := mod.Body[1].(*ast.FunctionDefinition)
	if !ok {
		t.Fatalf("expected second body element to be FunctionDefinition, got %T", mod.Body[1])
	}
	if len(fn.Params) != 1 {
		t.Fatalf("expected single function parameter, got %d", len(fn.Params))
	}
	param := fn.Params[0]
	checkSpan(t, "function parameter", param.Span(), 8, 12, 8, 24)
	if param.ParamType == nil {
		t.Fatalf("expected parameter to carry type annotation")
	}
	checkSpan(t, "parameter type", param.ParamType.Span(), 8, 19, 8, 24)

	if fn.Body == nil || len(fn.Body.Body) == 0 {
		t.Fatalf("expected function body to contain statements")
	}
	matchExpr, ok := fn.Body.Body[0].(*ast.MatchExpression)
	if !ok {
		t.Fatalf("expected first body statement to be MatchExpression, got %T", fn.Body.Body[0])
	}
	checkSpan(t, "match expression", matchExpr.Span(), 9, 3, 12, 4)
	if len(matchExpr.Clauses) == 0 {
		t.Fatalf("expected match expression to contain clauses")
	}
	structPattern, ok := matchExpr.Clauses[0].Pattern.(*ast.StructPattern)
	if !ok {
		t.Fatalf("expected first match clause to use struct pattern, got %T", matchExpr.Clauses[0].Pattern)
	}
	checkSpan(t, "struct pattern", structPattern.Span(), 10, 10, 10, 30)
	if len(structPattern.Fields) == 0 {
		t.Fatalf("expected struct pattern to include fields")
	}
	checkSpan(t, "struct pattern field", structPattern.Fields[0].Span(), 10, 18, 10, 22)

	if len(matchExpr.Clauses) < 2 {
		t.Fatalf("expected match expression to contain wildcard clause")
	}
	wildcardPattern, ok := matchExpr.Clauses[1].Pattern.(*ast.WildcardPattern)
	if !ok {
		t.Fatalf("expected second match clause to use wildcard pattern, got %T", matchExpr.Clauses[1].Pattern)
	}
	checkSpan(t, "wildcard pattern", wildcardPattern.Span(), 11, 10, 11, 11)
}

func TestParseRejectsPrefixMatchExpression(t *testing.T) {
	mp, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser: %v", err)
	}
	t.Cleanup(func() { mp.Close() })

	source := []byte(`package sample

fn main() {
  match 1 {
    case _ => 2
  }
}
`)

	_, err = mp.ParseModule(source)
	if err == nil {
		t.Fatalf("expected parse error for prefix match expression")
	}
	if !strings.Contains(err.Error(), "syntax errors") {
		t.Fatalf("unexpected error for prefix match expression: %v", err)
	}
}

func TestParsePlaceholderExpressions(t *testing.T) {
	source := "fn partials(data, factor) {\n  add(@, 10)\n  merge(@, @2, @1)\n  5.add\n}\n"

	p, err := NewModuleParser()
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

func placeholder(index int) ast.Expression {
	if index <= 0 {
		return ast.NewPlaceholderExpression(nil)
	}
	i := index
	return ast.NewPlaceholderExpression(&i)
}

func TestParseIgnoresCommentsInCompositeLists(t *testing.T) {
	source := `struct Pair {
	first: i32,
	## keep documenting the second slot
	second: i32,
}

values := [
	1,
	## keep trailing entry
	2,
]

result := make_pair(
	## left operand
	values[0],
	## right operand
	values[1],
)

fn make_pair(lhs: i32, rhs: i32) -> Pair {
	Pair { first: lhs, second: rhs }
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

	if len(mod.Body) < 4 {
		t.Fatalf("expected at least 4 statements, got %d", len(mod.Body))
	}

	structDef, ok := mod.Body[0].(*ast.StructDefinition)
	if !ok {
		t.Fatalf("expected first statement to be struct definition, got %T", mod.Body[0])
	}
	if len(structDef.Fields) != 2 {
		t.Fatalf("expected struct definition to have 2 fields, got %d", len(structDef.Fields))
	}

	assignValues, ok := mod.Body[1].(*ast.AssignmentExpression)
	if !ok {
		t.Fatalf("expected second statement to be assignment, got %T", mod.Body[1])
	}
	arrayLiteral, ok := assignValues.Right.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("expected values assignment to produce array literal, got %T", assignValues.Right)
	}
	if len(arrayLiteral.Elements) != 2 {
		t.Fatalf("expected array literal to have 2 elements, got %d", len(arrayLiteral.Elements))
	}

	assignResult, ok := mod.Body[2].(*ast.AssignmentExpression)
	if !ok {
		t.Fatalf("expected third statement to be assignment, got %T", mod.Body[2])
	}
	callExpr, ok := assignResult.Right.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected result assignment to be function call, got %T", assignResult.Right)
	}
	if len(callExpr.Arguments) != 2 {
		t.Fatalf("expected function call to have 2 arguments, got %d", len(callExpr.Arguments))
	}
}

func TestParseRaiseAndRethrowStatements(t *testing.T) {
	source := `fn fail() {
  raise Error("boom")
}

fn retry() {
  rethrow
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

	raiseFn := ast.NewFunctionDefinition(
		ast.ID("fail"),
		nil,
		ast.Block(
			ast.NewRaiseStatement(
				ast.NewFunctionCall(
					ast.ID("Error"),
					[]ast.Expression{ast.Str("boom")},
					nil,
					false,
				),
			),
		),
		nil,
		nil,
		nil,
		false,
		false,
	)

	rethrowFn := ast.NewFunctionDefinition(
		ast.ID("retry"),
		nil,
		ast.Block(
			ast.NewRethrowStatement(),
		),
		nil,
		nil,
		nil,
		false,
		false,
	)

	expected := ast.NewModule(
		[]ast.Statement{
			raiseFn,
			rethrowFn,
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseErrorHandlingExpressions(t *testing.T) {
	source := `value := (maybe() or { err =>
  err
} rescue {
  case err => err
} ensure {
  finalize()
})
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

func TestParseBitwiseXorExpression(t *testing.T) {
	source := `value := left .^ right`

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
				ast.NewBinaryExpression(".^", ast.ID("left"), ast.ID("right")),
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
value .^= mask
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

func TestParseArithmeticPrecedence(t *testing.T) {
	source := `result := 1 + 2 * 3
total := 4 + 6 / 3
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

	product := ast.NewBinaryExpression("*", ast.Int(2), ast.Int(3))
	sum := ast.NewBinaryExpression("+", ast.Int(1), product)

	division := ast.NewBinaryExpression("/", ast.Int(6), ast.Int(3))
	total := ast.NewBinaryExpression("+", ast.Int(4), division)

	expected := ast.NewModule(
		[]ast.Statement{
			ast.NewAssignmentExpression(
				ast.AssignmentDeclare,
				ast.ID("result"),
				sum,
			),
			ast.NewAssignmentExpression(
				ast.AssignmentDeclare,
				ast.ID("total"),
				total,
			),
		},
		nil,
		nil,
	)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}

func TestParseSimpleTypeExpressions(t *testing.T) {
	source := `fn format(flag: bool) -> String {
  if flag { "yes" } else { "no" }
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

	ifExpr := ast.NewIfExpression(
		ast.ID("flag"),
		ast.Block(ast.Str("yes")),
		[]*ast.ElseIfClause{},
		ast.Block(ast.Str("no")),
	)

	fn := ast.NewFunctionDefinition(
		ast.ID("format"),
		[]*ast.FunctionParameter{
			ast.NewFunctionParameter(ast.ID("flag"), ast.Ty("bool")),
		},
		ast.Block(ifExpr),
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
