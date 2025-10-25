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
