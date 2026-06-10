package interpreter

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestPopulateCallTypeArgumentsFromBytecodeResolvedCallArgs_InfersGenericArgument(t *testing.T) {
	interp := NewBytecode()
	decl := ast.Fn(
		"id",
		[]*ast.FunctionParameter{
			ast.Param("x", ast.Ty("T")),
		},
		nil,
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	call := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.Int(7)}, nil, false)
	if err := interp.populateCallTypeArgumentsFromBytecodeResolvedCallArgs(
		decl,
		call,
		[]runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI32)},
		0,
		1,
		nil,
		false,
	); err != nil {
		t.Fatalf("populate call type arguments: %v", err)
	}
	if len(call.TypeArguments) != 1 {
		t.Fatalf("type arg count = %d, want 1", len(call.TypeArguments))
	}
	if got := typeExpressionToString(call.TypeArguments[0]); got != "i32" {
		t.Fatalf("inferred type arg = %s, want i32", got)
	}
}

func TestPopulateCallTypeArgumentsFromBytecodeResolvedCallArgs_RecomputesGenericArgumentForPolymorphicCallSite(t *testing.T) {
	interp := NewBytecode()
	decl := ast.Fn(
		"id",
		[]*ast.FunctionParameter{
			ast.Param("x", ast.Ty("T")),
		},
		nil,
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	call := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.ID("x")}, nil, false)

	if err := interp.populateCallTypeArgumentsFromBytecodeResolvedCallArgs(
		decl,
		call,
		[]runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI32)},
		0,
		1,
		nil,
		false,
	); err != nil {
		t.Fatalf("populate i32 call type arguments: %v", err)
	}
	if got := typeExpressionToString(call.TypeArguments[0]); got != "i32" {
		t.Fatalf("first inferred type arg = %s, want i32", got)
	}

	if err := interp.populateCallTypeArgumentsFromBytecodeResolvedCallArgs(
		decl,
		call,
		[]runtime.Value{runtime.StringValue{Val: "hello"}},
		0,
		1,
		nil,
		false,
	); err != nil {
		t.Fatalf("populate String call type arguments: %v", err)
	}
	if got := typeExpressionToString(call.TypeArguments[0]); got != "String" {
		t.Fatalf("second inferred type arg = %s, want String", got)
	}
}

func TestPopulateCallTypeArgumentsFromBytecodeResolvedCallArgs_HotRuntimeKeyCacheAvoidsAllocationsForResolvedGenericStructArg(t *testing.T) {
	interp := NewBytecode()
	decl := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		nil,
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	box := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{
			Node: ast.StructDef(
				"Box",
				[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("T"), "value")},
				ast.StructKindNamed,
				[]*ast.GenericParameter{ast.GenericParam("T")},
				nil,
				false,
			),
		},
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32")},
	}
	call := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.ID("value")}, nil, false)
	stack := []runtime.Value{box}

	if err := interp.populateCallTypeArgumentsFromBytecodeResolvedCallArgs(
		decl,
		call,
		stack,
		0,
		1,
		nil,
		false,
	); err != nil {
		t.Fatalf("populate resolved generic struct call type arguments: %v", err)
	}
	if got := typeExpressionToString(call.TypeArguments[0]); got != "Box<i32>" {
		t.Fatalf("inferred type arg = %s, want Box<i32>", got)
	}
	if got := interp.inferredCallTypeArgumentRuntimeCacheEntryCount(); got != 1 {
		t.Fatalf("expected one inferred runtime type-argument cache entry, got %d", got)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if err := interp.populateCallTypeArgumentsFromBytecodeResolvedCallArgs(
			decl,
			call,
			stack,
			0,
			1,
			nil,
			false,
		); err != nil {
			panic(err)
		}
	})
	if allocs != 0 {
		t.Fatalf("expected bytecode resolved generic-struct inferred call-type hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestPopulateCallTypeArgumentsFromBytecodeResolvedCallArgs_ExactCacheHotPathAvoidsAllocationsForThreeArgGenericStruct(t *testing.T) {
	interp := NewBytecode()
	decl := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		nil,
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	tripleDef := ast.StructDef(
		"Triple",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("A"), "first")},
		ast.StructKindNamed,
		[]*ast.GenericParameter{
			ast.GenericParam("A"),
			ast.GenericParam("B"),
			ast.GenericParam("C"),
		},
		nil,
		false,
	)
	triple := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{Node: tripleDef},
		Fields: map[string]runtime.Value{
			"first": runtime.NewSmallInt(1, runtime.IntegerI32),
		},
		TypeArguments: []ast.TypeExpression{ast.Ty("i32"), ast.Ty("String"), ast.Ty("bool")},
	}
	call := ast.NewFunctionCall(ast.ID("id"), []ast.Expression{ast.ID("value")}, nil, false)
	stack := []runtime.Value{triple}

	if err := interp.populateCallTypeArgumentsFromBytecodeResolvedCallArgs(
		decl,
		call,
		stack,
		0,
		1,
		nil,
		false,
	); err != nil {
		t.Fatalf("populate three-arg generic struct bytecode call type arguments: %v", err)
	}
	if got := typeExpressionToString(call.TypeArguments[0]); got != "Triple<i32, String, bool>" {
		t.Fatalf("inferred type arg = %s, want Triple<i32, String, bool>", got)
	}
	if got := interp.inferredCallTypeArgumentRuntimeCacheEntryCount(); got != 0 {
		t.Fatalf("expected unsupported three-arg generic struct to skip runtime key cache, got %d entries", got)
	}
	if got := len(interp.inferredCallTypeArgumentCache); got != 1 {
		t.Fatalf("expected one exact inferred call type-argument cache entry, got %d", got)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if err := interp.populateCallTypeArgumentsFromBytecodeResolvedCallArgs(
			decl,
			call,
			stack,
			0,
			1,
			nil,
			false,
		); err != nil {
			panic(err)
		}
	})
	if allocs != 0 {
		t.Fatalf("expected bytecode three-arg exact inferred call-type hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestBytecodeVM_InlineGenericNamedCallSeedsTypeBindingEnv(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	module := mustParseModuleSource(t, `
fn type_name<T>(x: T) -> String {
  T_type
}

result := `+"`start`"+`
i := 0
loop {
  if i >= 2 { break }
  result = type_name<i32>(1)
  i = i + 1
}
result
`)

	want := mustEvalModule(t, New(), module)
	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic inline env mismatch: got=%#v want=%#v", got, want)
	}

	if got := len(interp.reusableBytecodeCallEnvCache); got != 1 {
		t.Fatalf("reusable env cache size = %d, want 1", got)
	}
	stats := interp.BytecodeStats()
	if stats.CallNameInlineResolvedHits == 0 {
		t.Fatalf("expected explicit generic named call to use inline-resolved dispatch, got stats=%#v", stats)
	}
	if stats.CallNameResolvedFunctionHits != 0 {
		t.Fatalf("CallNameResolvedFunctionHits = %d, want 0", stats.CallNameResolvedFunctionHits)
	}
}

func TestBytecodeVM_InlineGenericMemberCallSeedsCallLocalBindingEnv(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	module := mustParseModuleSource(t, `
struct Box T {
  value: T
}

methods Box T {
  fn type_name(self: Self) -> String {
    Self_type
  }
}

box := Box { value: 1 }
result := `+"`start`"+`
i := 0
loop {
  if i >= 2 { break }
  result = box.type_name()
  i = i + 1
}
result
`)

	want := mustEvalModule(t, New(), module)
	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic member inline env mismatch: got=%#v want=%#v", got, want)
	}

	stats := interp.BytecodeStats()
	if stats.InlineCallHits == 0 {
		t.Fatalf("expected inline call hits > 0 for repeated generic member call site")
	}
	if stats.InlineCallMisses != 0 {
		t.Fatalf("expected repeated generic member call site to avoid inline misses, got %d", stats.InlineCallMisses)
	}
	if got := len(interp.callLocalTypeBindingCache); got != 1 {
		t.Fatalf("call-local type binding cache size = %d, want 1", got)
	}
	if got := len(interp.reusableBytecodeCallEnvCache); got != 1 {
		t.Fatalf("reusable env cache size = %d, want 1", got)
	}
}

func TestBytecodeVM_InlineMethodSetConstraintRejectsUnsatisfiedReceiver(t *testing.T) {
	module := mustParseModuleSource(t, `
interface Show {
  fn show(self: Self) -> String
}

impl Show for i32 {
  fn show(self: Self) -> String { "i32" }
}

struct Box T {
  value: T
}

methods Box T where T: Show {
  fn describe(self: Self) -> String {
    self.value.show()
  }
}

good := Box i32 { value: 7 }
good.describe()
describe(good)

bad := Box bool { value: true }
describe(bad)
`)

	interp := NewBytecode()
	program, err := interp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	_, err = newBytecodeVM(interp, interp.GlobalEnvironment()).run(program)
	if err == nil {
		t.Fatalf("expected constrained method-set call to reject Box bool receiver")
	}
	message := err.Error()
	if !strings.Contains(message, "Type 'bool' does not satisfy interface 'Show'") {
		t.Fatalf("expected method-set constraint error, got %v", err)
	}
	if strings.Contains(message, "Member access only supported") {
		t.Fatalf("bytecode entered method body before checking method-set constraints: %v", err)
	}
}
