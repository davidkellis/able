package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeInstruction struct {
	op                  bytecodeOp
	name                string
	nameSimple          bool
	storeTyped          bool
	operator            string
	value               runtime.Value
	genericStructMatch  *bytecodeGenericStructPatternPlan
	intImmediate        runtime.IntegerValue
	intImmediate2       runtime.IntegerValue
	intImmediateRaw     int64
	intImmediate2Raw    int64
	floatImmediateRaw   float64
	floatImmediateKind  runtime.FloatType
	typeExpr            ast.TypeExpression
	target              int
	argCount            int
	loopBreak           int
	loopContinue        int
	node                ast.Node
	program             *bytecodeProgram
	hasIntImmediate     bool
	hasIntImmediate2    bool
	hasIntRaw           bool
	hasIntRaw2          bool
	hasFloatImmediate   bool
	typeSimpleCheck     bytecodeSimpleTypeCheck
	memberFastPath      bytecodeMemberMethodFastPathKind
	storeRawI32Sidecar  bool
	bitwiseRawCandidate bool
	slotArgs            bool
	discardResult       bool
	safe                bool
	preferMethods       bool
	transientScope      bool
}

type bytecodeProgram struct {
	instructions                    []bytecodeInstruction
	reach                           *bytecodeProgramReach
	frameLayout                     *bytecodeFrameLayout // non-nil when slot-indexed locals are used
	followedByPropagation           []bool               // optional: true when the next instruction is bytecodeOpPropagation
	integerConstValidationKnown     bool
	hasIntegerConstValidation       bool
	integerConstInstructionCount    int
	slotConstIntImmTable            *bytecodeSlotConstIntImmediateTable
	slotConstIntImmTableKnown       bool
	slotConstIntImmInstructionCount int
	returnGenericNames              map[string]struct{}
	returnGenericNamesCached        bool
	returnType                      ast.TypeExpression
	returnSimpleType                string
	returnSimpleCheck               bytecodeSimpleTypeCheck
	returnNullableSimple            string
	returnTypeUsesGenerics          bool
	returnTypeMetadataCached        bool
	i32RecurrenceKernel             *bytecodeI32RecurrenceKernel
	namedStructLiterals             map[int]bytecodeNamedStructLiteralPlan
	namedStructMembers              map[int]bytecodeNamedStructMemberPlan
	f64DotLoops                     map[int]bytecodeF64DotLoopPlan
	f64MatrixRowLoops               map[int]bytecodeF64MatrixRowLoopPlan
	f64AffineRowLoops               map[int]bytecodeF64AffineRowLoopPlan
	f64TransposeRowLoops            map[int]bytecodeF64TransposeRowLoopPlan
	f64AffinePushes                 map[int]bytecodeF64AffineProductPushPlan
	f64NestedGetPushes              map[int]bytecodeF64NestedArrayGetPushPlan
	floatMulAddMulJumps             map[int]bytecodeFloatMulAddMulCompareConstJumpPlan
	floatAddCompareConstJumps       map[int]bytecodeFloatAddCompareConstJumpPlan
	floatAffineStores               map[int]bytecodeStoreSlotFloatAffinePlan
	floatUpdatePairs                map[int]bytecodeFloatUpdatePairPlan
	floatRegions                    []bytecodeFloatRegionPlan
	arrayOwnershipMetadata          bytecodeArrayOwnershipProgramMetadata
}

type bytecodeNamedStructLiteralPlan struct {
	definition *runtime.StructDefinitionValue
	fieldOrder []int
}

type bytecodeNamedStructMemberPlan struct {
	definition *runtime.StructDefinitionValue
	fieldIndex int
}

type bytecodeI32RegisterFrame struct {
	values []int32
	valid  []bool
}

type bytecodeValueSlotI32Frame struct {
	values []int32
	valid  []bool
}

type bytecodeValueSlotFloatFrame struct {
	values []float64
	kinds  []runtime.FloatType
	valid  []bool
}

const bytecodeProgramMetadataDirectCacheSize = 8

type bytecodeValidatedIntConstDirectCacheEntry struct {
	program *bytecodeProgram
	values  []bool
}

type bytecodeSlotConstIntImmediateDirectCacheEntry struct {
	program *bytecodeProgram
	table   *bytecodeSlotConstIntImmediateTable
}

type bytecodeCallFrame struct {
	returnIP             int
	stackBase            int
	program              *bytecodeProgram
	slots                []runtime.Value
	env                  *runtime.Environment
	activeLookup         bytecodeActiveLookupProgramState
	transientScopeBase   int
	returnGenericNames   map[string]struct{}
	returnCoercionFn     *runtime.FunctionValue
	i32RegisterProgram   *bytecodeProgram
	i32Registers         []int32
	i32RegisterValid     []bool
	implicitSlotActive   []bool
	slotI32Values        []int32
	slotI32Valid         []bool
	slotFloatValues      []float64
	slotFloatKinds       []runtime.FloatType
	slotFloatValid       []bool
	iterBase             int
	loopBase             int
	hasImplicitReceiver  bool
	selfFast             bool
	arrayOwnershipParent *bytecodeArrayOwnershipFrame
}

type bytecodeActiveLookupProgramState struct {
	program               *bytecodeProgram
	globalLookupEntries   []bytecodeGlobalLookupCacheEntry
	scopeLookupEntries    []bytecodeScopeLookupCacheEntry
	callNameEntries       []*bytecodeCallNameCacheEntry
	indexMethodTable      *bytecodeIndexMethodCacheTable
	indexMethodGetEntries []bytecodeIndexMethodCacheEntry
	indexMethodSetEntries []bytecodeIndexMethodCacheEntry
}

type bytecodeCallFrameKind uint8

const (
	bytecodeCallFrameKindFull bytecodeCallFrameKind = iota
	bytecodeCallFrameKindSelfFast
	bytecodeCallFrameKindSelfFastMinimal
)

type bytecodeSelfFastCallFrame struct {
	returnIP             int
	stackBase            int
	slots                []runtime.Value
	env                  *runtime.Environment
	transientScopeBase   int
	returnGenericNames   map[string]struct{}
	returnCoercionFn     *runtime.FunctionValue
	i32RegisterProgram   *bytecodeProgram
	i32Registers         []int32
	i32RegisterValid     []bool
	implicitSlotActive   []bool
	slotI32Values        []int32
	slotI32Valid         []bool
	slotFloatValues      []float64
	slotFloatKinds       []runtime.FloatType
	slotFloatValid       []bool
	iterBase             int
	loopBase             int
	hasImplicitReceiver  bool
	arrayOwnershipParent *bytecodeArrayOwnershipFrame
}

type bytecodeSelfFastMinimalCallFrame struct {
	returnIP             int
	stackBase            int
	slots                []runtime.Value
	slot0                runtime.Value
	env                  *runtime.Environment
	transientScopeBase   int
	i32RegisterProgram   *bytecodeProgram
	i32Registers         []int32
	i32RegisterValid     []bool
	implicitSlotActive   []bool
	slotI32Values        []int32
	slotI32Valid         []bool
	slotFloatValues      []float64
	slotFloatKinds       []runtime.FloatType
	slotFloatValid       []bool
	iterBase             int
	loopBase             int
	slot0I32Raw          int32
	slot0I32Valid        bool
	slot0FloatRaw        float64
	slot0FloatKind       runtime.FloatType
	slot0FloatValid      bool
	reusesSlots          bool
	arrayOwnershipParent *bytecodeArrayOwnershipFrame
}

type bytecodeCallOperandRegion struct {
	site            bytecodeStackPeakSiteKey
	base            int
	operandValues   int
	expectedResults int
	valid           bool
}

// bytecodeOperandStack is intentionally boxed during the first stack
// representation migration. Naming the representation lets direct slice
// access move behind bytecodeVM helpers before a later scalar arm changes the
// element layout.
type bytecodeOperandStack []runtime.Value

type bytecodeVM struct {
	interp                                   *Interpreter
	stack                                    bytecodeOperandStack
	i32Stack                                 []int32
	i32UnboxFallbackValue                    runtime.Value
	i32UnboxFallbackSet                      bool
	i32RegisterProgram                       *bytecodeProgram
	i32Registers                             []int32
	i32RegisterValid                         []bool
	implicitSlotActive                       []bool
	i32RegisterFramePool                     map[int][]bytecodeI32RegisterFrame
	selfFastSlot0I32Raw                      int32
	selfFastSlot0I32Valid                    bool
	env                                      *runtime.Environment
	runtimeDataCacheEnv                      *runtime.Environment
	runtimeDataCacheState                    uint64
	runtimeDataCacheValue                    any
	runtimeDataCacheRev                      uint64
	runtimeDataCacheEnvRev                   uint64
	runtimeDataCacheKnown                    bool
	ip                                       int
	iterStack                                []forLoopIterator
	loopStack                                []bytecodeLoopFrame
	ensureStack                              []bytecodeEnsureFrame
	slots                                    []runtime.Value
	slotI32Owner                             *runtime.Value
	slotI32Values                            []int32
	slotI32Valid                             []bool
	slotFloatOwner                           *runtime.Value
	slotFloatValues                          []float64
	slotFloatKinds                           []runtime.FloatType
	slotFloatValid                           []bool
	slotFramePool                            map[int][][]runtime.Value
	slotFrameSmallHotPools                   [bytecodeSlotFrameSmallHotMaxSlots + 1][][]runtime.Value
	slotFrameHotSize                         int
	slotFrameHotPool                         [][]runtime.Value
	slotI32FramePool                         map[int][]bytecodeValueSlotI32Frame
	slotFloatFramePool                       map[int][]bytecodeValueSlotFloatFrame
	ownedI32Slots                            map[*runtime.Value]*runtime.IntegerValue
	ownedFloatSlots                          map[*runtime.Value]*runtime.FloatValue
	stackI32Cells                            []*bytecodeRawI32StackCell
	stackI64Cells                            []*bytecodeRawI64SlotCell
	stackIntegerCells                        []*bytecodeRawIntegerSlotCell
	rawI64SlotCellPool                       []*bytecodeRawI64SlotCell
	rawIntegerSlotCellPool                   []*bytecodeRawIntegerSlotCell
	nativeRawArgsInline                      [4]runtime.RawValue
	nativeRawArgs                            []runtime.RawValue
	nativeRawArgsBusy                        bool
	rawIntegerReturnScratch                  bytecodeRawIntegerReturnScratch
	callFrameKinds                           []bytecodeCallFrameKind
	callFrames                               []bytecodeCallFrame
	selfFastCallFrames                       []bytecodeSelfFastCallFrame
	selfFastMinimal                          []bytecodeSelfFastMinimalCallFrame
	selfFastMinimalSuffix                    int
	arrayOwnershipObserver                   *bytecodeArrayOwnershipObserver
	bytecodeStatsStackCapacity               int
	bytecodeStatsStackCapacityObserved       bool
	bytecodeStatsPendingOp                   bytecodeOp
	bytecodeStatsPendingIP                   int
	bytecodeStatsPendingProgram              *bytecodeProgram
	bytecodeStatsPendingStackDepth           int
	bytecodeStatsPendingOpObserved           bool
	bytecodeStatsPendingPeakSite             bytecodeStackPeakSiteKey
	bytecodeStatsPendingCallOperand          bytecodeCallOperandRegion
	bytecodeStatsStackPeakDepth              int
	bytecodeStatsInlineCallOperands          []bytecodeCallOperandRegion
	bytecodePrimitiveMaterializationCounters map[bytecodePrimitiveMaterializationKey]*uint64
	currentProgram                           *bytecodeProgram // tracks the active program for resume after yield
	bytecodeProgramEntryPending              bool
	activeLookup                             bytecodeActiveLookupProgramState
	globalLookupCache                        map[*bytecodeProgram][]bytecodeGlobalLookupCacheEntry
	globalLookupHotProgram                   *bytecodeProgram
	globalLookupHotEntries                   []bytecodeGlobalLookupCacheEntry
	scopeLookupCache                         map[*bytecodeProgram][]bytecodeScopeLookupCacheEntry
	scopeLookupHotProgram                    *bytecodeProgram
	scopeLookupHotEntries                    []bytecodeScopeLookupCacheEntry
	nameLookupHot                            bytecodeInlineNameLookupCacheEntry
	nameLookupHotEntry                       bytecodeScopeLookupCacheEntry
	callNameCache                            map[*bytecodeProgram][]*bytecodeCallNameCacheEntry
	callNameHotProgram                       *bytecodeProgram
	callNameHotEntries                       []*bytecodeCallNameCacheEntry
	callNameHot                              bytecodeInlineCallNameCacheEntry
	resolvedCallArgsSpill                    []runtime.Value
	memberMethodCache                        map[bytecodeMemberMethodCacheKey]bytecodeMemberMethodCacheEntry
	memberMethodHot                          bytecodeInlineMemberMethodCacheEntry
	memberMethodDirect                       [bytecodeMemberMethodDirectEntries]bytecodeInlineMemberMethodCacheEntry
	memberMethodNameRevisions                [bytecodeMemberMethodNameRevisionCacheSize]bytecodeMemberMethodNameRevisionCacheEntry
	memberMethodNameRevisionNext             uint8
	memberMethodFastPaths                    map[bytecodeMemberMethodFastPathCacheKey]bytecodeMemberMethodFastPathKind
	genericUnionCallDirect                   [bytecodeGenericUnionCallDirectSets][bytecodeGenericUnionCallDirectWays]bytecodeGenericUnionCallCacheEntry
	genericUnionCallDirectNext               [bytecodeGenericUnionCallDirectSets]uint8
	staticMemberCallCache                    map[bytecodeStaticMemberCallCacheKey]bytecodeStaticMemberCallCacheEntry
	staticMemberCallHot                      bytecodeInlineStaticMemberCallCacheEntry
	arrayGetOverloadHot                      *runtime.FunctionOverloadValue
	arrayGetOverloadHotVersion               uint64
	arrayGetOverloadHotOK                    bool
	arrayGetOverloadPairNullable             *runtime.FunctionValue
	arrayGetOverloadPairResult               *runtime.FunctionValue
	arrayGetOverloadPairVersion              uint64
	arrayGetOverloadPairOK                   bool
	arrayGetCallCache                        map[bytecodeGlobalLookupCacheKey]bytecodeArrayGetCallCacheEntry
	arrayGetCallDirect                       [bytecodeArrayGetCallDirectEntries]bytecodeInlineArrayGetCallCacheEntry
	arrayGetCallHot                          [bytecodeArrayGetCallHotEntries]bytecodeInlineArrayGetCallCacheEntry
	arrayGetF32NoErrorVersion                uint64
	arrayGetF32NoErrorKnown                  bool
	arrayGetF32NoError                       bool
	arrayGetF64NoErrorVersion                uint64
	arrayGetF64NoErrorKnown                  bool
	arrayGetF64NoError                       bool
	arrayGetPrimitiveNoErrorTokenHotToken    uint16
	arrayGetPrimitiveNoErrorTokenHotVersion  uint64
	arrayGetPrimitiveNoErrorTokenHotKnown    bool
	arrayGetPrimitiveNoErrorTokenHotNoError  bool
	arrayValueNoErrorVersion                 uint64
	arrayValueNoErrorKnown                   bool
	arrayValueNoError                        bool
	f64ArrayCache                            map[*runtime.ArrayState]bytecodeF64ArrayCacheEntry
	f64MatrixRowsCache                       map[*runtime.ArrayState]bytecodeF64MatrixRowsCacheEntry
	arrayNewCallCache                        map[bytecodeGlobalLookupCacheKey]bytecodeArrayNewCallCacheEntry
	arrayNewCallHot                          [bytecodeArrayNewCallHotEntries]bytecodeInlineArrayNewCallCacheEntry
	arraySlotCallCache                       map[bytecodeGlobalLookupCacheKey]bytecodeArraySlotCallCacheEntry
	arraySlotCallDirect                      [bytecodeArraySlotCallDirectEntries]bytecodeInlineArraySlotCallCacheEntry
	arraySlotCallHot                         [bytecodeArraySlotCallHotEntries]bytecodeInlineArraySlotCallCacheEntry
	stringDef                                *runtime.StructDefinitionValue
	stringDefSet                             bool
	stringBuilderDef                         *runtime.StructDefinitionValue
	stringBuilderDefSet                      bool
	stringBytesIterDef                       *runtime.StructDefinitionValue
	stringBytesIterDefSet                    bool
	stringCharsIterDef                       *runtime.StructDefinitionValue
	stringCharsIterDefSet                    bool
	stringBytesIteratorInterfaceDef          *runtime.InterfaceDefinitionValue
	stringBytesIteratorInterfaceDefSet       bool
	stringBytesIteratorNextMethod            runtime.Value
	stringBytesIteratorNextVersion           uint64
	stringBytesIteratorNextGlobalRev         uint64
	stringBytesIteratorNextSet               bool
	stringCharsIteratorNextMethod            runtime.Value
	stringCharsIteratorNextVersion           uint64
	stringCharsIteratorNextGlobalRev         uint64
	stringCharsIteratorNextSet               bool
	indexMethodCache                         map[*bytecodeProgram]*bytecodeIndexMethodCacheTable
	indexMethodDirect                        [bytecodeIndexMethodDirectEntries]bytecodeInlineIndexMethodCacheEntry
	indexMethodHot                           bytecodeInlineIndexMethodCacheEntry
	indexMethodHotAlt                        bytecodeInlineIndexMethodCacheEntry
	arrayIndexReceiverIdentityCache          map[int64]bytecodeArrayIndexReceiverIdentityCacheEntry
	arrayIndexReceiverIdentityHotHandle      int64
	arrayIndexReceiverIdentityHot            bytecodeArrayIndexReceiverIdentityCacheEntry
	arrayIndexReceiverMonoTokenHotHandle     int64
	arrayIndexReceiverMonoTokenHotRevision   uint64
	arrayIndexReceiverMonoTokenHot           uint16
	arrayIndexReceiverMonoTokenHotOK         bool
	validatedIntConsts                       map[*bytecodeProgram][]bool
	validatedIntConstsHotProgram             *bytecodeProgram
	validatedIntConstsHotValues              []bool
	validatedIntConstsHotAltProgram          *bytecodeProgram
	validatedIntConstsHotAltValues           []bool
	slotConstIntImm                          map[*bytecodeProgram]*bytecodeSlotConstIntImmediateTable
	validatedIntConstsDirect                 [bytecodeProgramMetadataDirectCacheSize]bytecodeValidatedIntConstDirectCacheEntry
	validatedIntConstsDirectNext             uint8
	slotConstIntImmDirect                    [bytecodeProgramMetadataDirectCacheSize]bytecodeSlotConstIntImmediateDirectCacheEntry
	slotConstIntImmDirectNext                uint8
	stringInterpParts                        []runtime.Value
	resolvedCallArgsInline                   [bytecodeInlinePreparedCallArgStorage]runtime.Value
	activeTransientScopeEnvs                 []*runtime.Environment
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
		interp:                   interp,
		env:                      env,
		stack:                    make([]runtime.Value, 0, 8),
		i32Stack:                 make([]int32, 0, 8),
		iterStack:                make([]forLoopIterator, 0, 2),
		loopStack:                make([]bytecodeLoopFrame, 0, 4),
		ensureStack:              make([]bytecodeEnsureFrame, 0, 2),
		callFrameKinds:           make([]bytecodeCallFrameKind, 0),
		callFrames:               make([]bytecodeCallFrame, 0),
		selfFastCallFrames:       make([]bytecodeSelfFastCallFrame, 0),
		selfFastMinimal:          make([]bytecodeSelfFastMinimalCallFrame, 0),
		activeTransientScopeEnvs: make([]*runtime.Environment, 0, 4),
	}
}
