package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeInstruction struct {
	op            bytecodeOp
	name          string
	operator      string
	value         runtime.Value
	target        int
	argCount      int
	loopBreak     int
	loopContinue  int
	node          ast.Node
	program       *bytecodeProgram
	safe          bool
	preferMethods bool
}

type bytecodeProgram struct {
	instructions []bytecodeInstruction
	frameLayout  *bytecodeFrameLayout // non-nil when slot-indexed locals are used
}

type bytecodeCallFrame struct {
	returnIP            int
	program             *bytecodeProgram
	slots               []runtime.Value
	env                 *runtime.Environment
	iterBase            int
	loopBase            int
	hasImplicitReceiver bool
}

type bytecodeVM struct {
	interp         *Interpreter
	stack          []runtime.Value
	env            *runtime.Environment
	ip             int
	iterStack      []forLoopIterator
	loopStack      []bytecodeLoopFrame
	ensureStack    []bytecodeEnsureFrame
	slots          []runtime.Value
	callFrames     []bytecodeCallFrame
	currentProgram *bytecodeProgram // tracks the active program for resume after yield
}

type bytecodeLoopFrame struct {
	breakTarget    int
	continueTarget int
	env            *runtime.Environment
}

type bytecodeEnsureFrame struct {
	result runtime.Value
	err    error
}

func newBytecodeVM(interp *Interpreter, env *runtime.Environment) *bytecodeVM {
	return &bytecodeVM{
		interp:      interp,
		env:         env,
		stack:       make([]runtime.Value, 0, 8),
		iterStack:   make([]forLoopIterator, 0, 2),
		loopStack:   make([]bytecodeLoopFrame, 0, 4),
		ensureStack: make([]bytecodeEnsureFrame, 0, 2),
	}
}
