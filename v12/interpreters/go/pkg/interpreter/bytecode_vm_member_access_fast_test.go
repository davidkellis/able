package interpreter

import (
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ExecMemberAccessNamedFieldFastPath(t *testing.T) {
	def := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Node",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
	}
	inst, values := runtime.NewStructInstancePositionalSized(def, 1, nil)
	values[0] = runtime.NewSmallInt(7, runtime.IntegerI32)

	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = append(vm.stack, inst)

	if err := vm.execMemberAccess(bytecodeInstruction{name: "value"}); err != nil {
		t.Fatalf("execMemberAccess failed: %v", err)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("expected single stack result, got %d entries", len(vm.stack))
	}
	got, ok := vm.stack[0].(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%#v)", vm.stack[0], vm.stack[0])
	}
	if raw, ok := got.ToInt64(); !ok || raw != 7 {
		t.Fatalf("expected field value 7, got %#v", vm.stack[0])
	}
}

func TestBytecodeVM_ExecMemberAccessPreferMethodsCallableFieldFastPathSkipsMethodCache(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	callable := runtime.NativeFunctionValue{
		Name:  "call",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			return runtime.NewSmallInt(1, runtime.IntegerI32), nil
		},
	}
	def := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Wrapper",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(nil, "call"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
	}
	inst, values := runtime.NewStructInstancePositionalSized(def, 1, nil)
	values[0] = callable

	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = append(vm.stack, inst)

	if err := vm.execMemberAccess(bytecodeInstruction{name: "call", preferMethods: true}); err != nil {
		t.Fatalf("execMemberAccess failed: %v", err)
	}
	got, ok := vm.stack[0].(runtime.NativeFunctionValue)
	if !ok {
		t.Fatalf("expected callable field result, got %T (%#v)", vm.stack[0], vm.stack[0])
	}
	if got.Name != "call" {
		t.Fatalf("expected callable field to be returned, got %#v", got)
	}

	stats := interp.BytecodeStats()
	if stats.MemberMethodCacheHits != 0 || stats.MemberMethodCacheMiss != 0 {
		t.Fatalf("expected callable field fast path to skip member-method cache counters, got hits=%d misses=%d", stats.MemberMethodCacheHits, stats.MemberMethodCacheMiss)
	}
}

func TestBytecodeVM_ExecMemberAccessNamedFieldPlanFastPath(t *testing.T) {
	def := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Node",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
	}
	inst, values := runtime.NewStructInstancePositionalSized(def, 1, nil)
	values[0] = runtime.NewSmallInt(9, runtime.IntegerI32)

	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.currentProgram = &bytecodeProgram{
		namedStructMembers: map[int]bytecodeNamedStructMemberPlan{
			0: {definition: def, fieldIndex: 0},
		},
	}
	vm.stack = append(vm.stack, inst)

	if err := vm.execMemberAccess(bytecodeInstruction{name: "value"}); err != nil {
		t.Fatalf("execMemberAccess failed: %v", err)
	}
	got, ok := vm.stack[0].(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%#v)", vm.stack[0], vm.stack[0])
	}
	if raw, ok := got.ToInt64(); !ok || raw != 9 {
		t.Fatalf("expected field value 9, got %#v", vm.stack[0])
	}
}

func TestBytecodeVM_ExecMemberSetNamedFieldPlanFastPath(t *testing.T) {
	def := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Node",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
	}
	inst, values := runtime.NewStructInstancePositionalSized(def, 1, nil)
	values[0] = runtime.NewSmallInt(3, runtime.IntegerI32)

	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.currentProgram = &bytecodeProgram{
		namedStructMembers: map[int]bytecodeNamedStructMemberPlan{
			0: {definition: def, fieldIndex: 0},
		},
	}
	replacement := runtime.NewSmallInt(11, runtime.IntegerI32)
	vm.stack = append(vm.stack, replacement, inst)

	instr := bytecodeInstruction{
		op:       bytecodeOpMemberSet,
		name:     "value",
		operator: string(ast.AssignmentAssign),
		node:     ast.Member(ast.ID("node"), "value"),
	}
	if err := vm.execMemberSet(instr); err != nil {
		t.Fatalf("execMemberSet failed: %v", err)
	}
	if len(vm.stack) != 1 || vm.stack[0] != replacement {
		t.Fatalf("expected replacement result on stack, got %#v", vm.stack)
	}
	if values[0] != replacement {
		t.Fatalf("expected planned member set to update positional field, got %#v", values[0])
	}
}

func TestBytecodeVM_NamedStructMemberPlansKeepDistinctFields(t *testing.T) {
	module := mustParseModuleSource(t, `
struct Ledger {
  total: i64,
  commits: i64,
  checksum: i64
}

fn update(ledger: Ledger) -> i64 {
  ledger.total = ledger.total + 37_i64
  ledger.commits = ledger.commits + 1_i64
  ledger.checksum = ledger.checksum + 481_i64
  (ledger.commits * 1_000_000_i64) + (ledger.total * 1_000_i64) + ledger.checksum
}

update(Ledger { total: 0_i64, commits: 0_i64, checksum: 0_i64 })
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("named struct field plan mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_NamedStructMemberPlansKeepDistinctLocalFields(t *testing.T) {
	module := mustParseModuleSource(t, `
struct Ledger {
  total: i64,
  commits: i64,
  checksum: i64
}

fn main() -> i64 {
  ledger := Ledger { total: 0_i64, commits: 0_i64, checksum: 0_i64 }
  ledger.total = ledger.total + 37_i64
  ledger.commits = ledger.commits + 1_i64
  ledger.checksum = ledger.checksum + 481_i64
  (ledger.commits * 1_000_000_i64) + (ledger.total * 1_000_i64) + ledger.checksum
}

main()
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("named struct local field plan mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_MemberAccessMaterializesRawFloatCarrier(t *testing.T) {
	loader, err := driver.NewLoader([]driver.SearchPath{
		{Path: stdlibRoot, Kind: driver.RootStdlib},
		{Path: kernelRoot, Kind: driver.RootStdlib},
	})
	if err != nil {
		t.Fatalf("loader init failed: %v", err)
	}
	stdlibProgram, err := loader.Load(filepath.Join(stdlibRoot, "numbers", "primitives.able"))
	if err != nil {
		t.Fatalf("load stdlib numbers/primitives failed: %v", err)
	}
	module := mustParseModuleSource(t, `
import able.numbers.primitives

fn main() -> bool {
  ((2.75.fract() - 0.75).abs()) < 0.000000000001
}

main()
`)

	treeInterp := New()
	if _, _, _, err := treeInterp.EvaluateProgram(stdlibProgram, ProgramEvaluationOptions{}); err != nil {
		t.Fatalf("tree stdlib preload failed: %v", err)
	}
	want, _, err := treeInterp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("module evaluation failed: %v", err)
	}

	byteInterp := NewBytecode()
	if _, _, _, err := byteInterp.EvaluateProgram(stdlibProgram, ProgramEvaluationOptions{}); err != nil {
		t.Fatalf("bytecode stdlib preload failed: %v", err)
	}
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode raw-float member access mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ExecMemberAccessDirectArrayLengthKeepsMonoHandle(t *testing.T) {
	handle := runtime.ArrayStoreMonoNewWithCapacityI32(8)
	if err := runtime.ArrayStoreMonoWriteI32(handle, 0, 5); err != nil {
		t.Fatalf("seed mono i32 array: %v", err)
	}
	arr, _, err := runtime.ArrayStoreValueViewFromHandle(handle, 1, 8)
	if err != nil {
		t.Fatalf("create handle-backed array view: %v", err)
	}
	if arr.State != nil {
		t.Fatalf("expected mono handle-backed array to start without tracked state")
	}

	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = append(vm.stack, arr)

	if err := vm.execMemberAccess(bytecodeInstruction{name: "length"}); err != nil {
		t.Fatalf("execMemberAccess failed: %v", err)
	}
	got, ok := vm.stack[0].(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%#v)", vm.stack[0], vm.stack[0])
	}
	if raw, ok := got.ToInt64(); !ok || raw != 1 {
		t.Fatalf("array length = %#v, want 1", vm.stack[0])
	}
	if arr.State != nil {
		t.Fatalf("direct bytecode array length should not materialize dynamic array state")
	}
	if _, err := runtime.ArrayStoreMonoReadI32(handle, 0); err != nil {
		t.Fatalf("mono i32 array should remain typed after bytecode member access: %v", err)
	}
}

func TestBytecodeVM_LoweringEmitsNamedStructMemberPlanForParamReceiver(t *testing.T) {
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("?Node"), "left"),
			ast.FieldDef(ast.Ty("?Node"), "right"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	readLeft := ast.Fn(
		"read_left",
		[]*ast.FunctionParameter{ast.Param("node", ast.Ty("Node"))},
		[]ast.Statement{
			ast.Ret(ast.Member(ast.ID("node"), "left")),
		},
		ast.Ty("?Node"),
		nil,
		nil,
		false,
		false,
	)

	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	if _, err := interp.evaluateStructDefinition(nodeDef, env); err != nil {
		t.Fatalf("evaluateStructDefinition failed: %v", err)
	}
	program, err := interp.lowerFunctionDefinitionBytecodeWithEnv(readLeft, env)
	if err != nil {
		t.Fatalf("lowerFunctionDefinitionBytecodeWithEnv failed: %v", err)
	}
	if program == nil {
		t.Fatalf("expected program")
	}
	found := false
	for ip, instr := range program.instructions {
		if instr.op != bytecodeOpLoadSlotStructField {
			continue
		}
		plan, ok := program.namedStructMembers[ip]
		if !ok {
			t.Fatalf("slot struct field plan missing at ip=%d", ip)
		}
		if plan.definition == nil || plan.definition.Node != nodeDef {
			t.Fatalf("slot struct field definition plan = %#v, want Node definition", plan.definition)
		}
		if plan.fieldIndex != 0 {
			t.Fatalf("slot struct field index = %d, want 0", plan.fieldIndex)
		}
		found = true
	}
	if !found {
		t.Fatalf("expected slot struct field load instruction")
	}
}

func TestBytecodeVM_LoweringEmitsNamedStructMemberPlanForGenericParamReceiver(t *testing.T) {
	module := mustParseModuleSource(t, `
struct Box T {
  value: T
}

fn read_box<T>(box: Box T) -> T {
  box.value
}

read_box(Box { value: 5 })
`)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("generic planned read mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "read_box")
	found := false
	for ip, instr := range program.instructions {
		if instr.op != bytecodeOpLoadSlotStructField {
			continue
		}
		plan, ok := program.namedStructMembers[ip]
		if !ok {
			t.Fatalf("generic slot struct field plan missing at ip=%d", ip)
		}
		if plan.definition == nil || plan.definition.Node == nil || plan.definition.Node.ID == nil || plan.definition.Node.ID.Name != "Box" {
			t.Fatalf("generic slot struct field definition plan = %#v, want Box definition", plan.definition)
		}
		if plan.fieldIndex != 0 {
			t.Fatalf("generic slot struct field index = %d, want 0", plan.fieldIndex)
		}
		found = true
	}
	if !found {
		t.Fatalf("expected generic slot struct field load instruction")
	}
}

func TestBytecodeVM_LoweringEmitsNamedStructMemberPlanForGenericMemberSet(t *testing.T) {
	module := mustParseModuleSource(t, `
struct Box T {
  value: T
}

fn write_box<T>(box: Box T, value: T) -> T {
  box.value = value
  box.value
}

write_box(Box { value: 5 }, 8)
`)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("generic planned write mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "write_box")
	foundSet := false
	for ip, instr := range program.instructions {
		if instr.op != bytecodeOpMemberSet {
			continue
		}
		plan, ok := program.namedStructMembers[ip]
		if !ok {
			t.Fatalf("generic member set plan missing at ip=%d", ip)
		}
		if plan.definition == nil || plan.definition.Node == nil || plan.definition.Node.ID == nil || plan.definition.Node.ID.Name != "Box" {
			t.Fatalf("generic member set definition plan = %#v, want Box definition", plan.definition)
		}
		if plan.fieldIndex != 0 {
			t.Fatalf("generic member set field index = %d, want 0", plan.fieldIndex)
		}
		foundSet = true
	}
	if !foundSet {
		t.Fatalf("expected generic member set instruction")
	}
}

func TestBytecodeVM_LoweringEmitsNamedStructMemberPlanForGenericSelfMemberSet(t *testing.T) {
	module := mustParseModuleSource(t, `
struct Box T {
  value: T
}

methods Box T {
  fn write(self: Self, value: T) -> T {
    self.value = value
    self.value
  }
}

box := Box { value: 5 }
box.write(8)
`)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("generic self planned write mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "write")
	foundSet := false
	foundLoad := false
	for ip, instr := range program.instructions {
		switch instr.op {
		case bytecodeOpMemberSet:
			plan, ok := program.namedStructMembers[ip]
			if !ok {
				t.Fatalf("generic self member set plan missing at ip=%d", ip)
			}
			if plan.definition == nil || plan.definition.Node == nil || plan.definition.Node.ID == nil || plan.definition.Node.ID.Name != "Box" {
				t.Fatalf("generic self member set definition plan = %#v, want Box definition", plan.definition)
			}
			if plan.fieldIndex != 0 {
				t.Fatalf("generic self member set field index = %d, want 0", plan.fieldIndex)
			}
			foundSet = true
		case bytecodeOpLoadSlotStructField:
			plan, ok := program.namedStructMembers[ip]
			if !ok {
				t.Fatalf("generic self slot struct field plan missing at ip=%d", ip)
			}
			if plan.definition == nil || plan.definition.Node == nil || plan.definition.Node.ID == nil || plan.definition.Node.ID.Name != "Box" {
				t.Fatalf("generic self slot struct field definition plan = %#v, want Box definition", plan.definition)
			}
			if plan.fieldIndex != 0 {
				t.Fatalf("generic self slot struct field index = %d, want 0", plan.fieldIndex)
			}
			foundLoad = true
		default:
			continue
		}
	}
	if !foundSet {
		t.Fatalf("expected generic self member set instruction")
	}
	if !foundLoad {
		t.Fatalf("expected generic self slot struct field load instruction")
	}
}

func TestBytecodeVM_ImplicitMemberSetMaterializesReusableRawCells(t *testing.T) {
	module := mustParseModuleSource(t, `
struct Ledger {
  total: i64,
  commits: i64,
  checksum: i64
}

methods Ledger {
  fn update(self: Self) -> i64 {
    #total = #total + 37_i64
    #commits = #commits + 1_i64
    #checksum = #checksum + 481_i64
    (#total * 1_000_000_i64) + (#commits * 1_000_i64) + #checksum
  }
}

ledger := Ledger { total: 0_i64, commits: 0_i64, checksum: 0_i64 }
ledger.update()
`)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("implicit member raw-cell materialization mismatch: got=%#v want=%#v", got, want)
	}
	if !valuesEqual(got, runtime.NewSmallInt(37_001_481, runtime.IntegerI64)) {
		t.Fatalf("implicit member result = %#v, want 37001481", got)
	}
}

func TestBytecodeVM_LoweringEmitsNamedStructMemberPlanForTypedPatternBinding(t *testing.T) {
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("?Node"), "left"),
			ast.FieldDef(ast.Ty("?Node"), "right"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	firstLeft := ast.Fn(
		"first_left",
		[]*ast.FunctionParameter{ast.Param("node", ast.Ty("Node"))},
		[]ast.Statement{
			ast.Ret(ast.Match(
				ast.Member(ast.ID("node"), "left"),
				ast.Mc(ast.LitP(ast.Nil()), ast.Nil()),
				ast.Mc(ast.TypedP(ast.ID("left"), ast.Ty("Node")), ast.Member(ast.ID("left"), "left")),
			)),
		},
		ast.Ty("?Node"),
		nil,
		nil,
		false,
		false,
	)

	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	if _, err := interp.evaluateStructDefinition(nodeDef, env); err != nil {
		t.Fatalf("evaluateStructDefinition failed: %v", err)
	}
	program, err := interp.lowerFunctionDefinitionBytecodeWithEnv(firstLeft, env)
	if err != nil {
		t.Fatalf("lowerFunctionDefinitionBytecodeWithEnv failed: %v", err)
	}
	if program == nil {
		t.Fatalf("expected program")
	}
	plannedMembers := 0
	for ip, instr := range program.instructions {
		if instr.op != bytecodeOpLoadSlotStructField {
			continue
		}
		plan, ok := program.namedStructMembers[ip]
		if !ok || plan.definition == nil || plan.definition.Node != nodeDef {
			t.Fatalf("slot struct field plan missing or wrong at ip=%d: %#v", ip, plan)
		}
		plannedMembers++
	}
	if plannedMembers < 2 {
		t.Fatalf("expected at least 2 planned slot struct field loads, got %d", plannedMembers)
	}
}
