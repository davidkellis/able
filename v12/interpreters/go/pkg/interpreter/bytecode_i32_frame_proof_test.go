package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
	"able/interpreter-go/pkg/typechecker"
)

func bytecodeI32FrameProofProgram(t *testing.T) (*driver.Program, *driver.Module) {
	t.Helper()
	module := mustParseModuleSource(t, `
package proof

fn typed(x: i32) -> i32 {
  y: i32 := x
  if true {
    y = y + 1
  } else {
    y = y + 2
  }
  y
}

fn inferred(x: i32) -> i32 {
  y := x
  if true {
    y = y + 1
  } else {
    y = y + 2
  }
  y
}
`)
	loaded := &driver.Module{Package: "proof", AST: module}
	return &driver.Program{Entry: loaded, Modules: []*driver.Module{loaded}}, loaded
}

func bytecodeProgramForNamedFunction(t *testing.T, env *runtime.Environment, name string) *bytecodeProgram {
	t.Helper()
	value, ok := env.Lookup(name)
	if !ok {
		t.Fatalf("missing %s function", name)
	}
	fn, ok := value.(*runtime.FunctionValue)
	if !ok || fn == nil {
		t.Fatalf("%s value = %#v, want function", name, value)
	}
	program, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || program == nil || program.frameLayout == nil {
		t.Fatalf("%s bytecode = %#v, want slot program", name, fn.Bytecode)
	}
	return program
}

func bytecodeSlotForName(t *testing.T, program *bytecodeProgram, name string) int {
	t.Helper()
	for _, instr := range program.instructions {
		if instr.name == name && bytecodeInstructionWritesI32FrameSlot(instr) {
			return instr.target
		}
	}
	t.Fatalf("missing slot write for %s", name)
	return -1
}

func TestBytecodeI32FrameProofUsesProgramTypecheckFacts(t *testing.T) {
	program, _ := bytecodeI32FrameProofProgram(t)
	interp := NewBytecode()
	_, env, check, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{})
	if err != nil {
		t.Fatalf("EvaluateProgram: %v", err)
	}
	if len(check.Diagnostics) != 0 {
		t.Fatalf("typecheck diagnostics = %v, want none", check.Diagnostics)
	}

	typed := bytecodeProgramForNamedFunction(t, env, "typed")
	if typed.frameLayout.i32FrameProof == nil {
		t.Fatal("typed function missing typechecker-backed i32 frame proof")
	}
	typedLocal := bytecodeSlotForName(t, typed, "y")
	if !typed.frameLayout.i32FrameProof.hasSlot(0) || !typed.frameLayout.i32FrameProof.hasSlot(typedLocal) {
		t.Fatalf("typed proof slots = %#v, want proven parameter and local", typed.frameLayout.i32FrameProof.slots)
	}

	inferred := bytecodeProgramForNamedFunction(t, env, "inferred")
	if inferred.frameLayout.i32FrameProof == nil {
		t.Fatal("inferred function missing frame proof metadata")
	}
	if !inferred.frameLayout.i32FrameProof.hasSlot(0) {
		t.Fatalf("inferred parameter proof slots = %#v, want slot 0", inferred.frameLayout.i32FrameProof.slots)
	}
	inferredLocal := bytecodeSlotForName(t, inferred, "y")
	if inferredLocal >= len(inferred.frameLayout.slotKinds) || inferred.frameLayout.slotKinds[inferredLocal] != bytecodeCellKindValue {
		t.Fatalf("unannotated local kind = %#v, want value slot", inferred.frameLayout.slotKinds)
	}
	if inferred.frameLayout.i32FrameProof.hasSlot(inferredLocal) {
		t.Fatalf("unannotated local unexpectedly gained i32 proof: %#v", inferred.frameLayout.i32FrameProof.slots)
	}
}

func TestBytecodeI32FrameProofSkippedTypecheckFallsBack(t *testing.T) {
	program, _ := bytecodeI32FrameProofProgram(t)
	interp := NewBytecode()
	_, env, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{SkipTypecheck: true})
	if err != nil {
		t.Fatalf("EvaluateProgram with skipped typecheck: %v", err)
	}
	if got := bytecodeProgramForNamedFunction(t, env, "typed").frameLayout.i32FrameProof; got != nil {
		t.Fatalf("skipped typecheck attached proof %#v", got)
	}
}

func TestBytecodeI32FrameProofUsesStandaloneModuleTypecheck(t *testing.T) {
	_, loaded := bytecodeI32FrameProofProgram(t)
	interp := NewBytecode()
	interp.EnableTypechecker(TypecheckConfig{FailFast: true})
	_, env, err := interp.EvaluateModule(loaded.AST)
	if err != nil {
		t.Fatalf("EvaluateModule: %v", err)
	}
	program := bytecodeProgramForNamedFunction(t, env, "typed")
	if program.frameLayout.i32FrameProof == nil || !program.frameLayout.i32FrameProof.hasSlot(0) {
		t.Fatalf("standalone typechecked proof = %#v, want proven i32 parameter", program.frameLayout.i32FrameProof)
	}
}

func TestBytecodeI32FrameProofRejectsAliasAndGenericFacts(t *testing.T) {
	if bytecodeInferenceIsConcreteI32(typechecker.AliasType{
		AliasName: "Count",
		Target:    typechecker.IntegerType{Suffix: "i32"},
	}) {
		t.Fatal("alias fact must not authorize concrete i32 frame storage")
	}
	if bytecodeInferenceIsConcreteI32(typechecker.TypeParameterType{ParameterName: "T"}) {
		t.Fatal("generic fact must not authorize concrete i32 frame storage")
	}
}

func TestBytecodeInferenceFactsExcludeDiagnosticModule(t *testing.T) {
	cleanNode := ast.ID("clean")
	badNode := ast.ID("bad")
	clean := &driver.Module{Package: "clean", AST: ast.Mod(nil, nil, ast.Pkg([]interface{}{"clean"}, false))}
	bad := &driver.Module{Package: "bad", AST: ast.Mod(nil, nil, ast.Pkg([]interface{}{"bad"}, false))}
	check := ProgramCheckResult{
		Diagnostics: []typechecker.ModuleDiagnostic{{Package: "bad"}},
		Inferred: map[string]typechecker.InferenceMap{
			"clean": {cleanNode: typechecker.IntegerType{Suffix: "i32"}},
			"bad":   {badNode: typechecker.IntegerType{Suffix: "i32"}},
		},
	}
	facts := bytecodeInferenceFactsForCheckedProgram(&driver.Program{
		Entry:   clean,
		Modules: []*driver.Module{clean, bad},
	}, check)
	if !bytecodeInferenceIsConcreteI32(facts[cleanNode]) {
		t.Fatalf("clean module fact = %#v, want i32", facts[cleanNode])
	}
	if facts[badNode] != nil {
		t.Fatalf("diagnostic module fact = %#v, want absent", facts[badNode])
	}
}
