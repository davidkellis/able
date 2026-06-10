package interpreter

import (
	"fmt"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_CallNameCacheRecordsDirectInlineShape(t *testing.T) {
	env := runtime.NewEnvironment(nil)
	layout := &bytecodeFrameLayout{
		slotCount:          3,
		paramSlots:         3,
		slotKinds:          []bytecodeCellKind{bytecodeCellKindValue, bytecodeCellKindI32, bytecodeCellKindI32},
		paramSimpleChecks:  []bytecodeSimpleTypeCheck{bytecodeSimpleTypeCheckUnknown, bytecodeSimpleTypeCheckI32, bytecodeSimpleTypeCheckI32},
		paramNeedsCoercion: []bool{false, true, true},
	}
	program := &bytecodeProgram{
		frameLayout:              layout,
		returnGenericNamesCached: true,
	}
	fn := &runtime.FunctionValue{Closure: env, Bytecode: program}
	lookup := bytecodeResolvedIdentifierLookup{
		value: fn,
		env:   env,
		owner: env,
	}

	callNode := ast.NewFunctionCall(ast.ID("swap"), nil, nil, false)
	entry := bytecodeBuildCallNameCacheEntry("swap", lookup, fn, 3, callNode)

	if entry.dispatch != bytecodeCallNameDispatchInline {
		t.Fatalf("expected inline dispatch, got %v", entry.dispatch)
	}
	if !entry.inlineDirect {
		t.Fatalf("expected cache entry to record direct inline shape")
	}
	if entry.inlineProgram != program || entry.inlineLayout != layout {
		t.Fatalf("unexpected direct inline metadata: program=%p layout=%p", entry.inlineProgram, entry.inlineLayout)
	}
	if entry.inlineI32ParamMask != 0b110 {
		t.Fatalf("inline i32 mask = %03b, want 110", entry.inlineI32ParamMask)
	}
	if entry.inlineKeepNilI32Mask != 0b110 {
		t.Fatalf("inline keep-nil i32 mask = %03b, want 110", entry.inlineKeepNilI32Mask)
	}
	if entry.inlineCoercionMask != 0b110 {
		t.Fatalf("inline coercion mask = %03b, want 110", entry.inlineCoercionMask)
	}
}

func TestBytecodeVM_CallNameDirectSlotInlineCopiesRawI64Cell(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironment(nil)
	callerProgram := &bytecodeProgram{instructions: make([]bytecodeInstruction, 1)}
	calleeLayout := &bytecodeFrameLayout{
		slotCount:          1,
		paramSlots:         1,
		slotKinds:          []bytecodeCellKind{bytecodeCellKindValue},
		paramNeedsCoercion: []bool{false},
		selfCallSlot:       -1,
	}
	calleeProgram := &bytecodeProgram{frameLayout: calleeLayout}
	fn := &runtime.FunctionValue{Closure: env, Bytecode: calleeProgram}
	entry := &bytecodeCallNameCacheEntry{
		inlineFn:      fn,
		inlineProgram: calleeProgram,
		inlineLayout:  calleeLayout,
		inlineDirect:  true,
	}
	callerCell := &bytecodeRawI64SlotCell{Val: 77}
	vm := newBytecodeVM(interp, env)
	vm.slots = []runtime.Value{callerCell}

	newProg, handled, err := vm.tryInlineCachedCallNameDirectFromSlots(
		entry,
		bytecodeInstruction{slotArgs: true, argCount: 1, target: 0},
		nil,
		callerProgram,
	)
	if err != nil {
		t.Fatalf("tryInlineCachedCallNameDirectFromSlots: %v", err)
	}
	if !handled || newProg != calleeProgram {
		t.Fatalf("direct slot inline = (%p, %t), want callee program handled", newProg, handled)
	}
	calleeCell, ok := vm.slots[0].(*bytecodeRawI64SlotCell)
	if !ok || calleeCell == nil {
		t.Fatalf("callee slot = %#v, want raw i64 cell", vm.slots[0])
	}
	if calleeCell == callerCell {
		t.Fatalf("callee raw i64 cell aliased caller cell")
	}
	if calleeCell.Val != 77 {
		t.Fatalf("callee raw i64 value = %d, want 77", calleeCell.Val)
	}
	calleeCell.Val = 99
	if callerCell.Val != 77 {
		t.Fatalf("caller raw i64 value changed through callee alias: got %d", callerCell.Val)
	}

	vm.releaseSlotFrame(vm.slots)
	reused := vm.acquireRawI64SlotCell(123)
	if reused != calleeCell {
		t.Fatalf("expected released callee raw i64 cell to be reused")
	}
	if reused.Val != 123 {
		t.Fatalf("reused raw i64 value = %d, want 123", reused.Val)
	}
}

func TestBytecodeVM_CallNameDispatchStatsSplitExactNativeAndDirectInline(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	interp.GlobalEnvironment().Define("native_id", runtime.NativeFunctionValue{
		Name:       "native_id",
		Arity:      1,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				t.Fatalf("native_id got %d args, want 1", len(args))
			}
			return args[0], nil
		},
	})
	module := mustParseModuleSource(t, `
fn inc(value: i32) -> i32 {
  value + 1
}

fn main() -> i32 {
  native_id(inc(1)) + inc(2)
}

main()
`)

	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := runtime.NewSmallInt(5, runtime.IntegerI32)
	if !valuesEqual(got, want) {
		t.Fatalf("unexpected result: got=%#v want=%#v", got, want)
	}
	stats := interp.BytecodeStats()
	if stats.CallNameExactNativeHits == 0 {
		t.Fatalf("expected exact-native call-name dispatch hits, got %#v", stats)
	}
	directInlineHits := stats.CallNameInlineDirectSlotHits + stats.CallNameInlineDirectStackHits
	if directInlineHits == 0 {
		t.Fatalf("expected direct-inline call-name dispatch hits, got %#v", stats)
	}
}

func TestBytecodeVM_CallNameCacheReusesGlobalOwnerAcrossSameShapeEnvs(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	fn := &runtime.FunctionValue{Closure: global}
	global.Define("f", fn)

	envA := runtime.NewEnvironment(global)
	envB := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, envA)
	lookup := bytecodeResolvedIdentifierLookup{
		value:        fn,
		env:          envA,
		envVersion:   vm.bytecodeEnvRevision(envA),
		owner:        global,
		ownerVersion: vm.bytecodeEnvRevision(global),
	}
	entry := bytecodeBuildCallNameCacheEntry("f", lookup, fn, 0, ast.NewFunctionCall(ast.ID("f"), nil, nil, false))
	if cached := vm.storeCachedCallName(program, 1, entry); cached == nil {
		t.Fatalf("expected call-name cache store")
	}

	vm.env = envB
	got, ok := vm.lookupCachedCallName(program, 1, "f")
	if !ok || got == nil || got.callee != fn {
		t.Fatalf("lookupCachedCallName() across same-shape envs = (%#v, %t), want global function", got, ok)
	}
}

func TestBytecodeVM_CallNameCacheRejectsGlobalOwnerAfterLocalShadow(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	fn := &runtime.FunctionValue{Closure: global}
	global.Define("f", fn)

	envA := runtime.NewEnvironment(global)
	envB := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, envA)
	lookup := bytecodeResolvedIdentifierLookup{
		value:        fn,
		env:          envA,
		envVersion:   vm.bytecodeEnvRevision(envA),
		owner:        global,
		ownerVersion: vm.bytecodeEnvRevision(global),
	}
	entry := bytecodeBuildCallNameCacheEntry("f", lookup, fn, 0, ast.NewFunctionCall(ast.ID("f"), nil, nil, false))
	if cached := vm.storeCachedCallName(program, 1, entry); cached == nil {
		t.Fatalf("expected call-name cache store")
	}

	envB.Define("f", runtime.NewSmallInt(1, runtime.IntegerI32))
	vm.env = envB
	if got, ok := vm.lookupCachedCallName(program, 1, "f"); ok || got != nil {
		t.Fatalf("lookupCachedCallName() after local shadow = (%#v, %t), want miss", got, ok)
	}
}

func TestBytecodeVM_CallNameCacheSkipsDirectInlineForTypeArguments(t *testing.T) {
	env := runtime.NewEnvironment(nil)
	program := &bytecodeProgram{
		frameLayout:              &bytecodeFrameLayout{slotCount: 1, paramSlots: 1},
		returnGenericNamesCached: true,
	}
	fn := &runtime.FunctionValue{Closure: env, Bytecode: program}
	lookup := bytecodeResolvedIdentifierLookup{
		value: fn,
		env:   env,
		owner: env,
	}

	callNode := ast.NewFunctionCall(ast.ID("id"), nil, []ast.TypeExpression{ast.Ty("i32")}, false)
	entry := bytecodeBuildCallNameCacheEntry("id", lookup, fn, 1, callNode)

	if entry.dispatch != bytecodeCallNameDispatchInline {
		t.Fatalf("expected generic inline dispatch to remain available, got %v", entry.dispatch)
	}
	if entry.inlineDirect {
		t.Fatalf("did not expect direct inline metadata for explicit type-argument call")
	}
}

func TestBytecodeVM_CallNameCacheSkipsDirectInlineForGenericLambda(t *testing.T) {
	env := runtime.NewEnvironment(nil)
	program := &bytecodeProgram{
		frameLayout:              &bytecodeFrameLayout{slotCount: 1, paramSlots: 1},
		returnGenericNames:       map[string]struct{}{"T": {}},
		returnGenericNamesCached: true,
	}
	fn := &runtime.FunctionValue{
		Declaration: ast.NewLambdaExpression(
			[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
			ast.ID("x"),
			ast.Ty("T"),
			[]*ast.GenericParameter{ast.GenericParam("T")},
			nil,
			false,
		),
		Closure:  env,
		Bytecode: program,
	}
	lookup := bytecodeResolvedIdentifierLookup{
		value: fn,
		env:   env,
		owner: env,
	}

	entry := bytecodeBuildCallNameCacheEntry("id", lookup, fn, 1, ast.NewFunctionCall(ast.ID("id"), nil, nil, false))

	if entry.dispatch != bytecodeCallNameDispatchInline {
		t.Fatalf("expected inline dispatch to remain available, got %v", entry.dispatch)
	}
	if entry.inlineDirect {
		t.Fatalf("did not expect direct inline metadata for generic lambda")
	}
}

func TestBytecodeVM_CallNameCacheBuildsArityOnlyRuntimeMatchForGenericFunction(t *testing.T) {
	env := runtime.NewEnvironment(nil)
	fnDef := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	fn := &runtime.FunctionValue{Declaration: fnDef, Closure: env}
	lookup := bytecodeResolvedIdentifierLookup{
		value: fn,
		env:   env,
		owner: env,
	}

	entry := bytecodeBuildCallNameCacheEntry("id", lookup, fn, 1, ast.NewFunctionCall(ast.ID("id"), nil, nil, false))

	if !entry.inlineRuntimeValid {
		t.Fatalf("expected cached runtime match plan for generic function")
	}
	if !entry.inlineRuntimeMatch.arityOnly {
		t.Fatalf("expected generic function runtime match plan to be arity-only")
	}
	if matched, decided := entry.inlineRuntimeMatch.matches([]runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI32)}); !decided || !matched {
		t.Fatalf("runtime match = (%t, %t), want decided match", matched, decided)
	}
	if matched, decided := entry.inlineRuntimeMatch.matches(nil); !decided || matched {
		t.Fatalf("runtime match with missing arg = (%t, %t), want decided miss", matched, decided)
	}
}

func TestBytecodeVM_CallNameRuntimeMatchSimpleCheckFallsBackForCoercibleInteger(t *testing.T) {
	env := runtime.NewEnvironment(nil)
	fnDef := ast.Fn(
		"widen",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("i64"))},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("i64"),
		nil,
		nil,
		false,
		false,
	)
	fn := &runtime.FunctionValue{Declaration: fnDef, Closure: env}
	plan, ok := bytecodeBuildCallNameRuntimeMatchPlan(fn)
	if !ok {
		t.Fatalf("expected runtime match plan")
	}

	if matched, decided := plan.matches([]runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI64)}); !decided || !matched {
		t.Fatalf("i64 exact runtime match = (%t, %t), want decided match", matched, decided)
	}
	if matched, decided := plan.matches([]runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI32)}); decided || matched {
		t.Fatalf("i32-to-i64 runtime match = (%t, %t), want fallback to full matcher", matched, decided)
	}
	if matched, decided := plan.matches([]runtime.Value{
		runtime.NewSmallInt(7, runtime.IntegerI64),
		runtime.NewSmallInt(8, runtime.IntegerI64),
	}); !decided || matched {
		t.Fatalf("extra-arg runtime match = (%t, %t), want decided miss", matched, decided)
	}
}

func TestBytecodeVM_CallNameRuntimeMatchDoesNotOveracceptPointerString(t *testing.T) {
	env := runtime.NewEnvironment(nil)
	fnDef := ast.Fn(
		"echo",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("String"))},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	fn := &runtime.FunctionValue{Declaration: fnDef, Closure: env}
	plan, ok := bytecodeBuildCallNameRuntimeMatchPlan(fn)
	if !ok {
		t.Fatalf("expected runtime match plan")
	}

	if matched, decided := plan.matches([]runtime.Value{runtime.StringValue{Val: "ok"}}); !decided || !matched {
		t.Fatalf("string value runtime match = (%t, %t), want decided match", matched, decided)
	}
	if matched, decided := plan.matches([]runtime.Value{&runtime.StringValue{Val: "fallback"}}); decided || matched {
		t.Fatalf("pointer string runtime match = (%t, %t), want fallback to full matcher", matched, decided)
	}
}

func TestBytecodeVM_CallNameCacheRecordsSingleOverloadInlineShape(t *testing.T) {
	env := runtime.NewEnvironment(nil)
	layout := &bytecodeFrameLayout{
		slotCount:  1,
		paramSlots: 1,
	}
	program := &bytecodeProgram{
		frameLayout:              layout,
		returnGenericNamesCached: true,
	}
	fn := &runtime.FunctionValue{Closure: env, Bytecode: program}
	overload := &runtime.FunctionOverloadValue{Overloads: []*runtime.FunctionValue{fn}}
	lookup := bytecodeResolvedIdentifierLookup{
		value: overload,
		env:   env,
		owner: env,
	}

	callNode := ast.NewFunctionCall(ast.ID("id"), nil, nil, false)
	entry := bytecodeBuildCallNameCacheEntry("id", lookup, overload, 1, callNode)

	if entry.dispatch != bytecodeCallNameDispatchInline {
		t.Fatalf("expected inline dispatch for single-overload wrapper, got %v", entry.dispatch)
	}
	if entry.inlineFn != fn {
		t.Fatalf("cached inline function = %p, want %p", entry.inlineFn, fn)
	}
	if !entry.inlineDirect {
		t.Fatalf("expected single-overload wrapper to keep direct inline metadata")
	}
	if entry.inlineProgram != program || entry.inlineLayout != layout {
		t.Fatalf("unexpected direct inline metadata: program=%p layout=%p", entry.inlineProgram, entry.inlineLayout)
	}
}

func TestBytecodeVM_CallNameCacheRecordsBoundSingleOverloadReceiver(t *testing.T) {
	env := runtime.NewEnvironment(nil)
	fn := &runtime.FunctionValue{Closure: env}
	receiver := runtime.StringValue{Val: "receiver"}
	callee := runtime.BoundMethodValue{
		Receiver: receiver,
		Method: &runtime.FunctionOverloadValue{
			Overloads: []*runtime.FunctionValue{fn},
		},
	}
	lookup := bytecodeResolvedIdentifierLookup{
		value: callee,
		env:   env,
		owner: env,
	}

	entry := bytecodeBuildCallNameCacheEntry("bound", lookup, callee, 0, ast.NewFunctionCall(ast.ID("bound"), nil, nil, false))

	if entry.dispatch != bytecodeCallNameDispatchInline {
		t.Fatalf("expected inline dispatch for bound single-overload wrapper, got %v", entry.dispatch)
	}
	if entry.inlineFn != fn {
		t.Fatalf("cached inline function = %p, want %p", entry.inlineFn, fn)
	}
	if !entry.hasInjectedReceiver {
		t.Fatalf("expected cached injected receiver for bound single-overload wrapper")
	}
	if entry.injectedReceiver != receiver {
		t.Fatalf("cached receiver = %#v, want %#v", entry.injectedReceiver, receiver)
	}
}

func TestBytecodeVM_PrepareRunProgramSeedsActiveCallNameEntriesFromExistingCache(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 3)}
	callee := &runtime.FunctionValue{Closure: env}

	vm.storeCachedCallName(program, 2, bytecodeCallNameCacheEntry{
		name:         "f",
		env:          env,
		envVersion:   vm.bytecodeEnvRevision(env),
		owner:        env,
		ownerVersion: vm.bytecodeEnvRevision(env),
		callee:       callee,
	})

	vm.callNameHotProgram = nil
	vm.callNameHotEntries = nil
	vm.callNameHot = bytecodeInlineCallNameCacheEntry{}

	vm.prepareRunProgram(program, false)

	if vm.activeLookup.program != program {
		t.Fatalf("active lookup program = %p, want %p", vm.activeLookup.program, program)
	}
	if vm.activeLookup.callNameEntries != nil {
		t.Fatalf("active call-name entries should stay lazy until first call-name lookup")
	}

	entry, ok := vm.lookupCachedCallName(program, 2, "f")
	if !ok || entry == nil || entry.callee != callee {
		t.Fatalf("lookupCachedCallName() = (%#v, %t), want callee %#v", entry, ok, callee)
	}
	if len(vm.activeLookup.callNameEntries) != len(program.instructions) {
		t.Fatalf("active call-name entries length = %d, want %d", len(vm.activeLookup.callNameEntries), len(program.instructions))
	}

	vm.callNameCache = nil
	vm.callNameHotProgram = nil
	vm.callNameHotEntries = nil
	vm.callNameHot = bytecodeInlineCallNameCacheEntry{}

	entry, ok = vm.lookupCachedCallName(program, 2, "f")
	if !ok || entry == nil || entry.callee != callee {
		t.Fatalf("lookupCachedCallName() after backing cache clear = (%#v, %t), want callee %#v", entry, ok, callee)
	}
}

func TestBytecodeVM_StoreCachedCallNameReusesEntryObject(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	first := &runtime.FunctionValue{Closure: env}
	second := &runtime.FunctionValue{Closure: env}

	vm.storeCachedCallName(program, 1, bytecodeCallNameCacheEntry{
		name:         "f",
		env:          env,
		envVersion:   vm.bytecodeEnvRevision(env),
		owner:        env,
		ownerVersion: vm.bytecodeEnvRevision(env),
		callee:       first,
	})
	entry := vm.callNameCache[program][1]
	if entry == nil {
		t.Fatalf("expected call-name cache entry after first store")
	}

	vm.storeCachedCallName(program, 1, bytecodeCallNameCacheEntry{
		name:         "f",
		env:          env,
		envVersion:   vm.bytecodeEnvRevision(env),
		owner:        env,
		ownerVersion: vm.bytecodeEnvRevision(env),
		callee:       second,
	})
	updated := vm.callNameCache[program][1]
	if updated == nil {
		t.Fatalf("expected call-name cache entry after second store")
	}
	if updated != entry {
		t.Fatalf("expected call-name cache entry reuse, got %p then %p", entry, updated)
	}
	if updated.callee != second {
		t.Fatalf("updated call-name callee = %#v, want %#v", updated.callee, second)
	}
	if vm.callNameHot.entry != updated {
		t.Fatalf("expected hot call-name cache to point at reused entry")
	}
	if vm.callNameHotProgram != program || len(vm.callNameHotEntries) != len(program.instructions) {
		t.Fatalf("expected call-name cache hot entries to cache program slice")
	}
}

func TestBytecodeVM_CallNameCacheSkipsShapeMetadataForSameEnvOwner(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	first := &runtime.FunctionValue{Closure: env}
	second := &runtime.FunctionValue{Closure: env}
	env.Define("f", first)
	lookup := bytecodeResolvedIdentifierLookup{
		value:        first,
		env:          env,
		envVersion:   vm.bytecodeEnvRevision(env),
		owner:        env,
		ownerVersion: vm.bytecodeEnvRevision(env),
	}
	entry := bytecodeBuildCallNameCacheEntry("f", lookup, first, 0, ast.NewFunctionCall(ast.ID("f"), nil, nil, false))
	if entry.nameShapeStateID != 0 || entry.bindingShapeVersion != 0 || entry.nameShapeVersion != 0 {
		t.Fatalf("same-env owner should skip shape metadata, got state=%d shape=%d name=%d", entry.nameShapeStateID, entry.bindingShapeVersion, entry.nameShapeVersion)
	}
	cached := vm.storeCachedCallName(program, 1, entry)
	if cached == nil {
		t.Fatalf("expected call-name cache store")
	}
	if cached.nameShapeStateID != 0 || cached.bindingShapeVersion != 0 || cached.nameShapeVersion != 0 {
		t.Fatalf("same-env owner store should keep shape metadata empty, got state=%d shape=%d name=%d", cached.nameShapeStateID, cached.bindingShapeVersion, cached.nameShapeVersion)
	}
	if got, ok := vm.lookupCachedCallName(program, 1, "f"); !ok || got == nil || got.callee != first {
		t.Fatalf("lookupCachedCallName() = (%#v, %t), want first callee", got, ok)
	}

	if !env.AssignExisting("f", second) {
		t.Fatalf("expected assignment to existing function binding to succeed")
	}
	if got, ok := vm.lookupCachedCallName(program, 1, "f"); ok || got != nil {
		t.Fatalf("lookupCachedCallName() after same-env owner mutation = (%#v, %t), want miss", got, ok)
	}
}

func TestBytecodeVM_LookupCachedCallNameInvalidatesOuterOwnerRevision(t *testing.T) {
	interp := NewBytecode()
	global := interp.GlobalEnvironment()
	env := runtime.NewEnvironment(global)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	callee := &runtime.FunctionValue{Closure: env}

	global.Define("f", callee)
	vm.storeCachedCallName(program, 1, bytecodeCallNameCacheEntry{
		name:         "f",
		env:          env,
		envVersion:   vm.bytecodeEnvRevision(env),
		owner:        global,
		ownerVersion: vm.bytecodeEnvRevision(global),
		callee:       callee,
	})

	if entry, ok := vm.lookupCachedCallName(program, 1, "f"); !ok || entry == nil || entry.callee != callee {
		t.Fatalf("lookupCachedCallName() = (%#v, %t), want cached callee %#v", entry, ok, callee)
	}

	global.Define("g", runtime.NewSmallInt(9, runtime.IntegerI32))

	if entry, ok := vm.lookupCachedCallName(program, 1, "f"); ok || entry != nil {
		t.Fatalf("lookupCachedCallName() after outer owner mutation = (%#v, %t), want miss", entry, ok)
	}
}

func TestBytecodeVM_LookupCachedCallNameIgnoresUnrelatedCurrentEnvShapeChange(t *testing.T) {
	interp := NewBytecode()
	parent := runtime.NewEnvironment(interp.GlobalEnvironment())
	env := runtime.NewEnvironment(parent)
	vm := newBytecodeVM(interp, env)
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}
	callee := &runtime.FunctionValue{Closure: parent}
	shadow := &runtime.FunctionValue{Closure: env}

	parent.Define("f", callee)
	cached := vm.storeCachedCallName(program, 1, bytecodeCallNameCacheEntry{
		name:         "f",
		env:          env,
		envVersion:   vm.bytecodeEnvRevision(env),
		owner:        parent,
		ownerVersion: vm.bytecodeEnvRevision(parent),
		callee:       callee,
	})
	if cached == nil {
		t.Fatalf("storeCachedCallName returned nil")
	}
	initialShapeVersion := cached.bindingShapeVersion

	env.Define("other", runtime.NewSmallInt(9, runtime.IntegerI32))
	afterOtherShapeVersion := env.BindingShapeRevision()
	if afterOtherShapeVersion == initialShapeVersion {
		t.Fatalf("expected unrelated binding to advance binding shape version")
	}
	if entry, ok := vm.lookupCachedCallName(program, 1, "f"); !ok || entry == nil || entry.callee != callee {
		t.Fatalf("lookupCachedCallName() after unrelated current binding = (%#v, %t), want cached callee %#v", entry, ok, callee)
	}
	if cached.bindingShapeVersion != afterOtherShapeVersion {
		t.Fatalf("cached binding shape version = %d, want refreshed %d", cached.bindingShapeVersion, afterOtherShapeVersion)
	}

	env.Define("f", shadow)
	if entry, ok := vm.lookupCachedCallName(program, 1, "f"); ok || entry != nil {
		t.Fatalf("lookupCachedCallName() after same-name shadow = (%#v, %t), want miss", entry, ok)
	}
}

func TestBytecodeVM_ExecCachedCallNameUsesStackResolvedFunctionAfterInlineMiss(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	env := runtime.NewEnvironment(nil)
	fnDef := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("i32"))},
		[]ast.Statement{ast.Int(0)},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	fn := &runtime.FunctionValue{
		Declaration: fnDef,
		Closure:     env,
		Bytecode: CompiledThunk(func(_ *runtime.Environment, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("compiled id args = %d, want 1", len(args))
			}
			return args[0], nil
		}),
	}
	entry := &bytecodeCallNameCacheEntry{
		name:         "id",
		env:          env,
		envVersion:   1,
		owner:        env,
		ownerVersion: 1,
		callee:       runtime.BoolValue{Val: true},
		dispatch:     bytecodeCallNameDispatchInline,
		inlineFn:     fn,
	}
	vm := newBytecodeVM(interp, env)
	vm.ip = 4
	vm.stack = append(vm.stack, runtime.NewSmallInt(41, runtime.IntegerI32))

	newProg, err := vm.execCachedCallName(entry, 0, 1, ast.NewFunctionCall(ast.ID("id"), nil, nil, false), nil)
	if err != nil {
		t.Fatalf("cached call-name execution failed: %v", err)
	}
	if newProg != nil {
		t.Fatalf("expected resolved-function fallback to complete without switching programs")
	}
	if vm.ip != 5 {
		t.Fatalf("vm ip = %d, want 5", vm.ip)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack size = %d, want 1", len(vm.stack))
	}
	got, ok := vm.stack[0].(runtime.IntegerValue)
	gotVal, gotOK := got.ToInt64()
	if !ok || !gotOK || got.TypeSuffix != runtime.IntegerI32 || gotVal != 41 {
		t.Fatalf("stack result = %#v, want i32 41", vm.stack[0])
	}
	stats := interp.BytecodeStats()
	if stats.DirectFunctionStackHits != 1 {
		t.Fatalf("DirectFunctionStackHits = %d, want 1", stats.DirectFunctionStackHits)
	}
	if stats.CallNameResolvedFunctionHits != 1 {
		t.Fatalf("CallNameResolvedFunctionHits = %d, want 1", stats.CallNameResolvedFunctionHits)
	}
	if stats.CallNameGenericFallbacks != 0 {
		t.Fatalf("CallNameGenericFallbacks = %d, want 0", stats.CallNameGenericFallbacks)
	}
}

func TestBytecodeVM_ExecCachedCallNameInlinesResolvedFunctionWithTypeArguments(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	env := runtime.NewEnvironment(nil)
	fnDef := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("T"))},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	fnProgram, err := interp.lowerFunctionDefinitionBytecode(fnDef)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	fn := &runtime.FunctionValue{
		Declaration: fnDef,
		Closure:     env,
	}
	setFunctionBytecodeProgram(fn, fnProgram)
	entry := &bytecodeCallNameCacheEntry{
		name:         "id",
		env:          env,
		envVersion:   1,
		owner:        env,
		ownerVersion: 1,
		callee:       runtime.BoolValue{Val: true},
		dispatch:     bytecodeCallNameDispatchInline,
		inlineFn:     fn,
	}
	vm := newBytecodeVM(interp, env)
	vm.ip = 4
	vm.stack = append(vm.stack, runtime.NewSmallInt(41, runtime.IntegerI32))

	newProg, err := vm.execCachedCallName(entry, 0, 1, ast.NewFunctionCall(ast.ID("id"), nil, []ast.TypeExpression{ast.Ty("i32")}, false), nil)
	if err != nil {
		t.Fatalf("cached call-name execution failed: %v", err)
	}
	if newProg != fnProgram {
		t.Fatalf("expected inline resolved function program %p, got %p", fnProgram, newProg)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("stack size = %d, want 0 after inline frame setup", len(vm.stack))
	}
	stats := interp.BytecodeStats()
	if stats.CallNameInlineResolvedHits != 1 {
		t.Fatalf("CallNameInlineResolvedHits = %d, want 1", stats.CallNameInlineResolvedHits)
	}
	if stats.CallNameResolvedFunctionHits != 0 {
		t.Fatalf("CallNameResolvedFunctionHits = %d, want 0", stats.CallNameResolvedFunctionHits)
	}
}

func TestBytecodeVM_LoweringEmitsCallNameSlotArgsForIdentifierArgs(t *testing.T) {
	def := ast.Fn(
		"caller",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("i", ast.Ty("i32")),
			ast.Param("j", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Call("swap", ast.ID("arr"), ast.ID("i"), ast.ID("j")),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	var sawSlotArgs bool
	for _, instr := range program.instructions {
		if instr.op != bytecodeOpCallName || instr.name != "swap" {
			continue
		}
		sawSlotArgs = true
		if !instr.slotArgs {
			t.Fatalf("expected call-name instruction to use slot args")
		}
		if instr.argCount != 3 || instr.target != 0 || instr.loopBreak != 1 || instr.loopContinue != 2 {
			t.Fatalf("call-name slot args = count %d slots %d/%d/%d, want 3 slots 0/1/2", instr.argCount, instr.target, instr.loopBreak, instr.loopContinue)
		}
	}
	if !sawSlotArgs {
		t.Fatalf("expected lowering to emit call-name slot args")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpLoadSlot) {
		t.Fatalf("expected slot-arg call lowering to skip standalone argument LoadSlot opcodes")
	}
}
