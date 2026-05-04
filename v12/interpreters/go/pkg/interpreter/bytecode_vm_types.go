package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeInstruction struct {
	op              bytecodeOp
	name            string
	nameSimple      bool
	storeTyped      bool
	operator        string
	value           runtime.Value
	intImmediate    runtime.IntegerValue
	intImmediateRaw int64
	typeExpr        ast.TypeExpression
	target          int
	argCount        int
	loopBreak       int
	loopContinue    int
	node            ast.Node
	program         *bytecodeProgram
	hasIntImmediate bool
	hasIntRaw       bool
	safe            bool
	preferMethods   bool
}

type bytecodeProgram struct {
	instructions             []bytecodeInstruction
	frameLayout              *bytecodeFrameLayout // non-nil when slot-indexed locals are used
	returnGenericNames       map[string]struct{}
	returnGenericNamesCached bool
}

type bytecodeCallFrame struct {
	returnIP            int
	program             *bytecodeProgram
	slots               []runtime.Value
	env                 *runtime.Environment
	returnGenericNames  map[string]struct{}
	iterBase            int
	loopBase            int
	hasImplicitReceiver bool
	selfFast            bool
}

type bytecodeCallFrameKind uint8

const (
	bytecodeCallFrameKindFull bytecodeCallFrameKind = iota
	bytecodeCallFrameKindSelfFast
	bytecodeCallFrameKindSelfFastMinimal
)

type bytecodeSelfFastCallFrame struct {
	returnIP            int
	slots               []runtime.Value
	returnGenericNames  map[string]struct{}
	iterBase            int
	loopBase            int
	hasImplicitReceiver bool
}

type bytecodeSelfFastMinimalCallFrame struct {
	returnIP      int
	slots         []runtime.Value
	slot0         runtime.Value
	slot0I32Raw   int32
	slot0I32Valid bool
	reusesSlots   bool
}

type bytecodeVM struct {
	interp                             *Interpreter
	stack                              []runtime.Value
	i32Stack                           []int32
	selfFastSlot0I32Raw                int32
	selfFastSlot0I32Valid              bool
	env                                *runtime.Environment
	ip                                 int
	iterStack                          []forLoopIterator
	loopStack                          []bytecodeLoopFrame
	ensureStack                        []bytecodeEnsureFrame
	slots                              []runtime.Value
	slotFramePool                      map[int][][]runtime.Value
	slotFrameHotSize                   int
	slotFrameHotPool                   [][]runtime.Value
	callFrameKinds                     []bytecodeCallFrameKind
	callFrames                         []bytecodeCallFrame
	selfFastCallFrames                 []bytecodeSelfFastCallFrame
	selfFastMinimal                    []bytecodeSelfFastMinimalCallFrame
	selfFastMinimalSuffix              int
	currentProgram                     *bytecodeProgram // tracks the active program for resume after yield
	globalLookupCache                  map[bytecodeGlobalLookupCacheKey]bytecodeGlobalLookupCacheEntry
	scopeLookupCache                   map[bytecodeGlobalLookupCacheKey]bytecodeScopeLookupCacheEntry
	nameLookupHot                      bytecodeInlineNameLookupCacheEntry
	callNameCache                      map[bytecodeGlobalLookupCacheKey]*bytecodeCallNameCacheEntry
	callNameHot                        bytecodeInlineCallNameCacheEntry
	memberMethodCache                  map[bytecodeMemberMethodCacheKey]bytecodeMemberMethodCacheEntry
	memberMethodHot                    bytecodeInlineMemberMethodCacheEntry
	memberMethodFastPaths              map[bytecodeMemberMethodFastPathCacheKey]bytecodeMemberMethodFastPathKind
	arrayGetOverloadHot                *runtime.FunctionOverloadValue
	arrayGetOverloadHotVersion         uint64
	arrayGetOverloadHotOK              bool
	arrayGetOverloadPairNullable       *runtime.FunctionValue
	arrayGetOverloadPairResult         *runtime.FunctionValue
	arrayGetOverloadPairVersion        uint64
	arrayGetOverloadPairOK             bool
	arrayGetCallCache                  map[bytecodeGlobalLookupCacheKey]bytecodeArrayGetCallCacheEntry
	arrayGetCallHot                    [bytecodeArrayGetCallHotEntries]bytecodeInlineArrayGetCallCacheEntry
	stringBytesIterDef                 *runtime.StructDefinitionValue
	stringBytesIterDefSet              bool
	stringBytesIteratorInterfaceDef    *runtime.InterfaceDefinitionValue
	stringBytesIteratorInterfaceDefSet bool
	stringBytesIteratorNextMethod      runtime.Value
	stringBytesIteratorNextVersion     uint64
	stringBytesIteratorNextGlobalRev   uint64
	stringBytesIteratorNextSet         bool
	indexMethodCache                   map[*bytecodeProgram]*bytecodeIndexMethodCacheTable
	indexMethodHot                     bytecodeInlineIndexMethodCacheEntry
	validatedIntConsts                 map[*bytecodeProgram][]bool
	slotConstIntImm                    map[*bytecodeProgram]*bytecodeSlotConstIntImmediateTable
	stringInterpParts                  []runtime.Value
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
		interp:             interp,
		env:                env,
		stack:              make([]runtime.Value, 0, 8),
		i32Stack:           make([]int32, 0, 8),
		iterStack:          make([]forLoopIterator, 0, 2),
		loopStack:          make([]bytecodeLoopFrame, 0, 4),
		ensureStack:        make([]bytecodeEnsureFrame, 0, 2),
		callFrameKinds:     make([]bytecodeCallFrameKind, 0),
		callFrames:         make([]bytecodeCallFrame, 0),
		selfFastCallFrames: make([]bytecodeSelfFastCallFrame, 0),
		selfFastMinimal:    make([]bytecodeSelfFastMinimalCallFrame, 0),
	}
}
