package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	goParser "able/interpreter-go/pkg/parser"
	"able/interpreter-go/pkg/runtime"
)

func mustParseModuleSource(t *testing.T, source string) *ast.Module {
	t.Helper()
	parser, err := goParser.NewModuleParser()
	if err != nil {
		t.Fatalf("module parser init failed: %v", err)
	}
	defer parser.Close()

	module, err := parser.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("module parse failed: %v", err)
	}
	return module
}

func TestBytecodeVM_ImplGenericMethodReturnUsesMethodSetGenerics(t *testing.T) {
	module := mustParseModuleSource(t, `
interface Iterator T {
  fn next(self: Self) -> (?T)
}

struct OneIter T {
  value: T
}

impl Iterator T for OneIter {
  fn next(self: Self) -> (?T) {
    self.value
  }
}

interface Enumerable T for C _ {
  fn iterator(self: C T) -> (Iterator T)
}

struct Bag T {
  value: T
}

impl Enumerable T for Bag {
  fn iterator(self: Self) -> (Iterator T) {
    OneIter { value: self.value }
  }
}

Bag { value: 7 }.iterator().next() or { 0 }
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode impl-generic interface return mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeProgramReturnGenericNamesCacheIncludesMethodSetGenerics(t *testing.T) {
	def := ast.Fn(
		"next",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.ID("self"),
		},
		ast.Ty("T"),
		nil,
		nil,
		false,
		false,
	)
	methodSet := &runtime.MethodSet{
		TargetType: ast.Ty("Box"),
		GenericParams: []*ast.GenericParameter{
			ast.GenericParam("T", nil),
		},
	}
	fn := &runtime.FunctionValue{
		Declaration: def,
		MethodSet:   methodSet,
	}
	program := &bytecodeProgram{}

	setFunctionBytecodeProgram(fn, program)

	if !program.returnGenericNamesCached {
		t.Fatalf("expected bytecode program generic names to be cached")
	}
	if _, ok := program.returnGenericNames["T"]; !ok {
		t.Fatalf("expected cached generic names to include method-set generic T, got %#v", program.returnGenericNames)
	}
	if got := bytecodeProgramReturnGenericNames(fn, program); got == nil {
		t.Fatalf("expected cached generic names from program")
	}
}
