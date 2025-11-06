package parser

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
)

func TestParsePackageStatement(t *testing.T) {
	source := `package sample.core;
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

func TestParseModuleImports(t *testing.T) {
	p, err := NewModuleParser()
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
