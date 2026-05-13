package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_CallNameCacheRecordsDirectInlineShape(t *testing.T) {
	env := runtime.NewEnvironment(nil)
	layout := &bytecodeFrameLayout{slotCount: 3, paramSlots: 3}
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
