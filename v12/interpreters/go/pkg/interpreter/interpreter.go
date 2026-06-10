package interpreter

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"weak"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type packageMeta struct {
	namePath  []string
	isPrivate bool
}

type evalState struct {
	raiseStack        []runtime.Value
	breakpoints       []string
	implicitReceivers []runtime.Value
	placeholderStack  []placeholderFrame
	blockFrames       map[*ast.BlockExpression]*blockFrame
	callStack         []runtimeCallFrame
	pendingDiagCtxs   []*runtimeDiagnosticContext
}

func newEvalState() *evalState {
	return &evalState{
		raiseStack:        make([]runtime.Value, 0),
		breakpoints:       make([]string, 0),
		implicitReceivers: make([]runtime.Value, 0),
		placeholderStack:  make([]placeholderFrame, 0),
		blockFrames:       make(map[*ast.BlockExpression]*blockFrame),
		callStack:         make([]runtimeCallFrame, 0),
		pendingDiagCtxs:   make([]*runtimeDiagnosticContext, 0),
	}
}

type blockFrame struct {
	env    *runtime.Environment
	index  int
	result runtime.Value
}

func (s *evalState) pushBreakpoint(label string) {
	if s == nil {
		return
	}
	s.breakpoints = append(s.breakpoints, label)
}

func (s *evalState) popBreakpoint() {
	if s == nil || len(s.breakpoints) == 0 {
		return
	}
	s.breakpoints = s.breakpoints[:len(s.breakpoints)-1]
}

func (s *evalState) hasBreakpoint(label string) bool {
	if s == nil {
		return false
	}
	for idx := len(s.breakpoints) - 1; idx >= 0; idx-- {
		if s.breakpoints[idx] == label {
			return true
		}
	}
	return false
}

func (s *evalState) pushRaise(val runtime.Value) {
	if s == nil {
		return
	}
	s.raiseStack = append(s.raiseStack, val)
}

func (s *evalState) popRaise() (runtime.Value, bool) {
	if s == nil || len(s.raiseStack) == 0 {
		return nil, false
	}
	last := s.raiseStack[len(s.raiseStack)-1]
	s.raiseStack = s.raiseStack[:len(s.raiseStack)-1]
	return last, true
}

func (s *evalState) peekRaise() (runtime.Value, bool) {
	if s == nil || len(s.raiseStack) == 0 {
		return nil, false
	}
	return s.raiseStack[len(s.raiseStack)-1], true
}

func (s *evalState) pushImplicitReceiver(val runtime.Value) {
	if s == nil {
		return
	}
	s.implicitReceivers = append(s.implicitReceivers, val)
}

func (s *evalState) popImplicitReceiver() {
	if s == nil || len(s.implicitReceivers) == 0 {
		return
	}
	s.implicitReceivers = s.implicitReceivers[:len(s.implicitReceivers)-1]
}

func (s *evalState) currentImplicitReceiver() (runtime.Value, bool) {
	if s == nil || len(s.implicitReceivers) == 0 {
		return nil, false
	}
	return s.implicitReceivers[len(s.implicitReceivers)-1], true
}

func (s *evalState) pushPlaceholderFrame(paramCount int, args []runtime.Value) {
	if s == nil {
		return
	}
	frame := placeholderFrame{
		args:       args,
		paramCount: paramCount,
	}
	s.placeholderStack = append(s.placeholderStack, frame)
}

func (s *evalState) popPlaceholderFrame() {
	if s == nil || len(s.placeholderStack) == 0 {
		return
	}
	s.placeholderStack = s.placeholderStack[:len(s.placeholderStack)-1]
}

func (s *evalState) currentPlaceholderFrame() (*placeholderFrame, bool) {
	if s == nil || len(s.placeholderStack) == 0 {
		return nil, false
	}
	return &s.placeholderStack[len(s.placeholderStack)-1], true
}

func (s *evalState) hasPlaceholderFrame() bool {
	return s != nil && len(s.placeholderStack) > 0
}

func (s *evalState) pushCallFrame(node *ast.FunctionCall) {
	if s == nil || node == nil {
		return
	}
	s.freezePendingDiagnosticContexts()
	s.callStack = append(s.callStack, runtimeCallFrame{node: node})
}

func (s *evalState) popCallFrame() {
	if s == nil || len(s.callStack) == 0 {
		return
	}
	s.freezePendingDiagnosticContexts()
	s.callStack = s.callStack[:len(s.callStack)-1]
}

func (s *evalState) snapshotCallStack() []runtimeCallFrame {
	if s == nil || len(s.callStack) == 0 {
		return nil
	}
	out := make([]runtimeCallFrame, len(s.callStack))
	copy(out, s.callStack)
	return out
}

func (s *evalState) snapshotCallStackPrefix(depth int) []runtimeCallFrame {
	if s == nil || depth <= 0 || len(s.callStack) == 0 {
		return nil
	}
	if depth > len(s.callStack) {
		depth = len(s.callStack)
	}
	out := make([]runtimeCallFrame, depth)
	copy(out, s.callStack[:depth])
	return out
}

func (s *evalState) registerPendingDiagnosticContext(ctx *runtimeDiagnosticContext) {
	if s == nil || ctx == nil {
		return
	}
	s.pendingDiagCtxs = append(s.pendingDiagCtxs, ctx)
}

func (s *evalState) freezePendingDiagnosticContexts() {
	if s == nil || len(s.pendingDiagCtxs) == 0 {
		return
	}
	for _, ctx := range s.pendingDiagCtxs {
		if ctx != nil {
			ctx.freezeCallStack()
		}
	}
	clear(s.pendingDiagCtxs)
	s.pendingDiagCtxs = s.pendingDiagCtxs[:0]
}

type placeholderFrame struct {
	args       []runtime.Value
	paramCount int
}

func (f *placeholderFrame) valueAt(index int) (runtime.Value, error) {
	if index <= 0 || index > len(f.args) {
		return nil, fmt.Errorf("Placeholder index @%d is out of range", index)
	}
	return f.args[index-1], nil
}

type execMode int

const (
	execModeTreewalker execMode = iota
	execModeBytecode
)

// Interpreter drives evaluation of Able v12 AST nodes.
type Interpreter struct {
	global                 *runtime.Environment
	inherentMethods        map[string]map[string]runtime.Value
	interfaces             map[string]*runtime.InterfaceDefinitionValue
	unionDefinitions       map[string]*runtime.UnionDefinitionValue
	typeAliases            map[string]*ast.TypeAliasDefinition
	implMethods            map[string][]implEntry
	genericImpls           []implEntry
	arrayIndexImpls        bool
	arrayIndexMutImpls     bool
	rangeImplementations   []rangeImplementation
	unnamedImpls           map[string]map[string]map[string]bool
	packageRegistry        map[string]map[string]runtime.Value
	packageMetadata        map[string]packageMeta
	packageEnvs            map[string]*runtime.Environment
	externHostPackages     map[string]*externHostPackage
	externHostMu           sync.Mutex
	currentPackage         string
	dynamicDefinitionMode  bool
	dynPackageDefMethod    runtime.NativeFunctionValue
	dynPackageEvalMethod   runtime.NativeFunctionValue
	dynamicPackageEnvs     map[string]*runtime.Environment
	executor               Executor
	execMode               execMode
	rootState              *evalState
	runtimeDataCacheEnv    *runtime.Environment
	runtimeDataCacheState  uint64
	runtimeDataCacheValue  any
	runtimeDataCacheRev    uint64
	runtimeDataCacheEnvRev uint64
	runtimeDataCacheKnown  bool
	nodeOrigins            map[ast.Node]string

	concurrencyReady      bool
	futureErrorStruct     *runtime.StructDefinitionValue
	futureStatusStructs   map[string]*runtime.StructDefinitionValue
	futureStatusPending   runtime.Value
	futureStatusResolved  runtime.Value
	futureStatusCancelled runtime.Value
	awaitWakerStruct      *runtime.StructDefinitionValue
	awaitRoundRobinIndex  int

	channelMutexReady       bool
	channelMu               sync.Mutex
	channels                map[int64]*channelState
	nextChannelHandle       int64
	pendingChannelSends     map[*runtime.FutureValue]*channelSendWaiter
	pendingChannelReceives  map[*runtime.FutureValue]*channelReceiveWaiter
	mutexMu                 sync.Mutex
	mutexes                 map[int64]*mutexState
	nextMutexHandle         int64
	concurrencyErrorStructs map[string]*runtime.StructDefinitionValue
	standardErrorStructs    map[string]*runtime.StructDefinitionValue

	stringHostReady bool
	osReady         bool
	osArgs          []string
	ratioReady      bool

	orderingStructs map[string]*runtime.StructDefinitionValue
	orderingValues  map[string]*runtime.StructInstanceValue
	divModStruct    *runtime.StructDefinitionValue
	ratioStruct     *runtime.StructDefinitionValue

	arrayReady     bool
	arrayMu        sync.Mutex
	arraysByHandle map[int64]arrayHandleTracking
	hashMapReady   bool

	errorNativeMethods map[string]runtime.NativeFunctionValue

	generatorStack []*generatorInstance

	interfaceBuiltinsReady bool
	envSingleThread        bool

	methodCache                                map[methodCacheKey]methodCacheEntry
	equalityDispatchCache                      map[equalityDispatchCacheKey]equalityDispatchCacheEntry
	interfaceImplCache                         map[interfaceImplCacheKey]interfaceImplCacheEntry
	selectedInterfaceImplCache                 map[interfaceImplCacheKey]*selectedInterfaceImplCacheEntry
	interfaceMethodDictionaryCache             map[interfaceMethodDictionaryCacheKey]interfaceMethodDictionaryCacheEntry
	iteratorInterfaceMethodDictionaryCache     map[*runtime.InterfaceDefinitionValue]iteratorInterfaceMethodDictionaryCacheEntry
	interfaceDefaultMethodCache                map[interfaceDefaultMethodCacheKey]interfaceDefaultMethodCacheEntry
	implTargetMatchCache                       map[implTargetMatchCacheKey]bool
	boundMethodCache                           map[boundMethodCacheKey]runtime.Value
	methodScopeCallableCache                   map[methodScopeCallableCacheKey]methodScopeCallableCacheEntry
	methodScopeHasCache                        map[methodScopeHasCacheKey]methodScopeHasCacheEntry
	propagationErrorCache                      map[string]bool
	methodCacheMu                              sync.RWMutex
	methodCacheVersion                         uint64
	overloadCache                              map[overloadCacheKey]*runtime.FunctionValue
	typeAliasCacheMu                           sync.RWMutex
	typeAliasBaseCache                         map[string][]string
	typeAliasReferenceCache                    map[ast.TypeExpression]bool
	typeAliasExpansionCache                    map[ast.TypeExpression]ast.TypeExpression
	functionCallGenericPlanCacheMu             sync.RWMutex
	functionCallGenericPlanCache               map[ast.Node]*functionCallGenericPlan
	callableExplicitRuntimeBindingUsageCacheMu sync.RWMutex
	callableExplicitRuntimeBindingUsageCache   map[ast.Node]bool
	functionRuntimeGenericBindingPlanCacheMu   sync.RWMutex
	functionRuntimeGenericBindingPlanCache     map[*runtime.FunctionValue]*functionRuntimeGenericBindingPlan
	methodSetConstraintPlanCacheMu             sync.RWMutex
	methodSetConstraintPlanCache               map[*runtime.MethodSet]*methodSetConstraintPlan
	functionCallConstraintResultCache          map[functionCallConstraintResultCacheKey]functionCallConstraintResultCacheEntry
	methodSetConstraintResultCache             map[methodSetConstraintResultCacheKey]methodSetConstraintResultCacheEntry
	namedStructLiteralPlanCacheMu              sync.RWMutex
	namedStructLiteralPlanCache                map[*ast.StructLiteral]namedStructLiteralPlan
	namedStructPatternPlanCacheMu              sync.RWMutex
	namedStructPatternPlanCache                map[namedStructPatternPlanCacheKey]namedStructPatternPlan
	structGenericInferencePlanCacheMu          sync.RWMutex
	structGenericInferencePlanCache            map[*ast.StructDefinition]*structGenericInferencePlan
	blockTransientRuntimeScopeCacheMu          sync.RWMutex
	blockTransientRuntimeScopeCache            map[*ast.BlockExpression]bool
	matchClauseScopePlanCacheMu                sync.RWMutex
	matchClauseScopePlanCache                  map[*ast.MatchClause]clauseScopePlan
	matchExpressionClausePlansCacheMu          sync.RWMutex
	matchExpressionClausePlansCache            map[*ast.MatchExpression][]clauseScopePlan
	matchExpressionBytecodeProgramsCacheMu     sync.RWMutex
	matchExpressionBytecodeProgramsCache       map[*ast.MatchExpression][]bytecodeMatchClausePrograms
	rescueExpressionClausePlansCacheMu         sync.RWMutex
	rescueExpressionClausePlansCache           map[*ast.RescueExpression][]clauseScopePlan
	callTypeArgumentStateMu                    sync.RWMutex
	callTypeArgumentState                      map[*ast.FunctionCall]callTypeArgumentState
	inferredCallTypeArgumentCacheMu            sync.RWMutex
	inferredCallTypeArgumentCache              map[inferredCallTypeArgumentCacheKey][]ast.TypeExpression
	inferredCallTypeArgumentRuntimeCache1      map[inferredCallTypeArgumentRuntimeCacheKey1][]ast.TypeExpression
	inferredCallTypeArgumentRuntimeCache2      map[inferredCallTypeArgumentRuntimeCacheKey2][]ast.TypeExpression
	inferredCallTypeArgumentRuntimeCache       map[inferredCallTypeArgumentRuntimeCacheKey][]ast.TypeExpression
	explicitCallTypeBindingCacheMu             sync.RWMutex
	explicitCallTypeBindingCache               map[explicitCallTypeBindingCacheKey][]runtime.EnvironmentBinding
	callLocalTypeBindingCacheMu                sync.RWMutex
	callLocalTypeBindingCache                  map[callLocalTypeBindingCacheKey][]runtime.EnvironmentBinding
	reusableBytecodeCallEnvCacheMu             sync.RWMutex
	reusableBytecodeCallEnvCache               map[reusableBytecodeCallEnvCacheKey]*runtime.Environment
	callableTransientCallEnvReuseCacheMu       sync.RWMutex
	callableTransientCallEnvReuseCache         map[ast.Node]bool
	transientClauseEnvPool                     sync.Pool
	transientClauseBindingPool                 sync.Pool
	transientCallEnvPool                       sync.Pool
	transientRuntimeScopeEnvPool               sync.Pool
	typeInfoCacheMu                            sync.RWMutex
	knownTypeNameCache                         map[string]bool
	typeInfoNameCache                          map[typeInfoCacheKey]string
	typeInfoExpressionCache                    map[typeExpressionCacheKey]ast.TypeExpression
	typeExpressionTupleCache                   map[typeExpressionSliceKey][]ast.TypeExpression
	bytecodeVMPool                             sync.Pool
	bytecodeArrayOwnershipProfile              *bytecodeArrayOwnershipProfile
	bytecodeStringStats                        *bytecodeStringStats
	nativeCallContextPool                      sync.Pool
	nativeBorrowCallArgScratchPool             sync.Pool
	fixedCallArg2Pool                          sync.Pool
	bytecodeExprCacheMu                        sync.RWMutex
	bytecodeExprCache                          map[bytecodeExpressionProgramCacheKey]*bytecodeProgram
	bytecodeLambdaCacheMu                      sync.RWMutex
	bytecodeLambdaCache                        map[bytecodeLambdaProgramCacheKey]*bytecodeProgram
	bytecodeLambdaDependencyNames              map[*ast.LambdaExpression][]string
	bytecodeInferenceFactsMu                   sync.RWMutex
	bytecodeInferenceFacts                     bytecodeInferenceFacts
	bytecodeMethodSelections                   bytecodeMethodSelections
	runtimeInferenceFacts                      bytecodeInferenceFacts
	runtimeStaticCallReceiverTypes             map[*ast.FunctionCall]ast.TypeExpression

	bytecodeStatsEnabled                         bool
	bytecodePrimitiveMaterializationStatsEnabled bool
	bytecodePrimitiveMaterializationsMu          sync.Mutex
	bytecodePrimitiveMaterializations            map[bytecodePrimitiveMaterializationKey]*uint64
	bytecodePrimitiveMaterializationsDropped     uint64
	bytecodeProgramReachMu                       sync.Mutex
	bytecodeProgramReach                         map[*bytecodeProgram]struct{}
	bytecodeProgramReachDropped                  uint64
	bytecodeOpCounts                             [bytecodeOpCount]uint64
	bytecodeValueStackDeltas                     [bytecodeOpCount]int64
	bytecodeStackPeakSitesMu                     sync.Mutex
	bytecodeStackPeakSites                       map[bytecodeStackPeakSiteKey]uint64
	bytecodeStackDeltaSitesMu                    sync.Mutex
	bytecodeStackDeltaSites                      map[bytecodeStackPeakSiteKey]int64
	bytecodeCallOperandBalancesMu                sync.Mutex
	bytecodeCallOperandBalances                  map[bytecodeStackPeakSiteKey]bytecodeCallOperandBalance
	bytecodeLoopBackedgeBalancesMu               sync.Mutex
	bytecodeLoopBackedgeBalances                 map[bytecodeLoopBackedgeBalanceKey]bytecodeLoopBackedgeBalance
	bytecodeInlineFrameBalancesMu                sync.Mutex
	bytecodeInlineFrameBalances                  map[bytecodeInlineFrameBalanceKey]bytecodeInlineFrameBalance
	bytecodeValueStackMaxDepth                   uint64
	bytecodeValueStackMaxCapacity                uint64
	bytecodeValueStackCapacityGrowths            uint64
	bytecodeCallFrameMaxDepth                    uint64
	bytecodeLoadNameLookups                      uint64
	bytecodeLoadNameCountsMu                     sync.Mutex
	bytecodeLoadNameCounts                       map[string]uint64
	bytecodeLoadNameHotHits                      uint64
	bytecodeLoadNameScopeCacheHits               uint64
	bytecodeLoadNameGlobalCacheHits              uint64
	bytecodeLoadNameDirectCurrent                uint64
	bytecodeLoadNameDirectOuter                  uint64
	bytecodeLoadNameScopeStores                  uint64
	bytecodeLoadNameGlobalStores                 uint64
	bytecodeCallNameLookups                      uint64
	bytecodeCallNameDottedFallbacks              uint64
	bytecodeCallNameExactNativeHits              uint64
	bytecodeCallNameInlineDirectSlotHits         uint64
	bytecodeCallNameInlineDirectStackHits        uint64
	bytecodeCallNameInlineResolvedHits           uint64
	bytecodeCallNameInlineGenericHits            uint64
	bytecodeCallNameResolvedFunctionHits         uint64
	bytecodeCallNameGenericFallbacks             uint64
	bytecodeGenericUnionCallCacheHits            uint64
	bytecodeGenericUnionCallCacheMisses          uint64
	bytecodeInlineCallHits                       uint64
	bytecodeInlineCallMisses                     uint64
	bytecodeDirectFunctionStackHits              uint64
	bytecodeInlineResolvedMissNoBytecode         uint64
	bytecodeInlineResolvedMissArity              uint64
	bytecodeInlineResolvedMissTypeArgs           uint64
	bytecodeInlineResolvedMissGenericLambda      uint64
	bytecodeMemberMethodCacheHits                uint64
	bytecodeMemberMethodCacheMisses              uint64
	bytecodeCallMemberResolvedExactNative        uint64
	bytecodeCallMemberResolvedInline             uint64
	bytecodeCallMemberResolvedGeneric            uint64
	bytecodeCallMemberResolvedFallback           uint64
	bytecodeCallMemberStaticCacheHits            uint64
	bytecodeCallMemberStaticCacheMisses          uint64
	bytecodeCallMemberStaticExactNative          uint64
	bytecodeCallMemberStaticInline               uint64
	bytecodeCallMemberStaticGeneric              uint64
	bytecodeExprCacheHits                        uint64
	bytecodeExprCacheMisses                      uint64
	bytecodeArrayIndexSlotLookups                uint64
	bytecodeArrayIndexSlotTrackedHits            uint64
	bytecodeArrayIndexSlotMonoUnsignedHits       uint64
	bytecodeArrayIndexSlotDirectHits             uint64
	bytecodeArrayIndexSlotFallbacks              uint64
	bytecodeArrayIndexSlotFastDisabledMiss       uint64
	bytecodeArrayIndexSlotReceiverMiss           uint64
	bytecodeArrayIndexSlotIndexMiss              uint64
	bytecodeArrayIndexSlotHandleMiss             uint64
	bytecodeArrayIndexSlotDirectMiss             uint64
	bytecodeArrayMemberSlotLookups               uint64
	bytecodeArrayMemberSlotLenLookups            uint64
	bytecodeArrayMemberSlotReadLookups           uint64
	bytecodeArrayMemberSlotWriteLookups          uint64
	bytecodeArrayMemberSlotPushLookups           uint64
	bytecodeArrayMemberSlotCacheHits             uint64
	bytecodeArrayMemberSlotFastHits              uint64
	bytecodeArrayMemberSlotFallbacks             uint64
	bytecodeArrayMemberSlotReceiverMiss          uint64
	bytecodeArrayMemberSlotCacheMiss             uint64
	bytecodeArrayMemberSlotFastPathMiss          uint64
	bytecodeTraceEnabled                         bool
	bytecodeTraceMu                              sync.Mutex
	bytecodeTraceCounts                          map[bytecodeTraceKey]uint64

	typecheckerEnabled   bool
	typecheckerStrict    bool
	typechecker          interpreterTypechecker
	typecheckDiagnostics []interpreterTypecheckDiagnostic

	interfaceMethodResolver  func(receiver runtime.Value, interfaceName string, methodName string) (runtime.Value, bool)
	compiledImplChecker      func(typeName string, interfaceName string) bool
	compiledInstanceMethodFn func(typeName string, methodName string) (runtime.Value, bool)
	// compiledInterfaceMemberFn searches all compiled interface dispatch tables
	// for a method by name, regardless of interface. Maps to __able_interface_dispatch_member.
	compiledInterfaceMemberFn func(receiver runtime.Value, methodName string) (runtime.Value, bool)
}

type arrayHandleTracking struct {
	single weak.Pointer[runtime.ArrayValue]
	many   map[weak.Pointer[runtime.ArrayValue]]struct{}
}

func identifiersToStrings(ids []*ast.Identifier) []string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == nil {
			continue
		}
		parts = append(parts, id.Name)
	}
	return parts
}

func joinIdentifierNames(ids []*ast.Identifier) string {
	parts := identifiersToStrings(ids)
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ".")
}

func (i *Interpreter) qualifiedName(name string) string {
	if i.currentPackage == "" {
		return ""
	}
	return i.currentPackage + "." + name
}

func (i *Interpreter) stateFromEnv(env *runtime.Environment) *evalState {
	if env != nil {
		if data := i.runtimeDataFromEnv(env); data != nil {
			if payload, ok := data.(*asyncContextPayload); ok {
				if payload.state == nil {
					payload.state = newEvalState()
				}
				return payload.state
			}
		}
	}
	if i.rootState == nil {
		i.rootState = newEvalState()
	}
	return i.rootState
}

func (i *Interpreter) registerSymbol(name string, value runtime.Value) {
	if i.currentPackage == "" {
		return
	}
	bucket, ok := i.packageRegistry[i.currentPackage]
	if !ok {
		bucket = make(map[string]runtime.Value)
		i.packageRegistry[i.currentPackage] = bucket
	}
	if !i.dynamicDefinitionMode {
		if existing, ok := bucket[name]; ok {
			if merged, ok := runtime.MergeFunctionValues(existing, value); ok {
				value = merged
			}
		}
	}
	bucket[name] = value
	i.updateKnownTypeNameCacheForPackageSymbol(name, value)
	if qn := i.qualifiedName(name); qn != "" {
		i.global.Define(qn, value)
	}
}

func (i *Interpreter) defineInEnv(env *runtime.Environment, name string, value runtime.Value) {
	if env == nil || name == "" {
		return
	}
	if i.dynamicDefinitionMode && env.HasInCurrentScope(name) {
		_ = env.Assign(name, value)
		return
	}
	env.Define(name, value)
}

func newInterpreter(exec Executor, mode execMode) *Interpreter {
	if exec == nil {
		exec = NewSerialExecutor(nil)
	}
	i := &Interpreter{
		global:               runtime.NewEnvironment(nil),
		inherentMethods:      make(map[string]map[string]runtime.Value),
		interfaces:           make(map[string]*runtime.InterfaceDefinitionValue),
		unionDefinitions:     make(map[string]*runtime.UnionDefinitionValue),
		typeAliases:          make(map[string]*ast.TypeAliasDefinition),
		implMethods:          make(map[string][]implEntry),
		genericImpls:         make([]implEntry, 0),
		rangeImplementations: make([]rangeImplementation, 0),
		unnamedImpls:         make(map[string]map[string]map[string]bool),
		packageRegistry:      make(map[string]map[string]runtime.Value),
		packageMetadata:      make(map[string]packageMeta),
		packageEnvs:          make(map[string]*runtime.Environment),
		externHostPackages:   make(map[string]*externHostPackage),
		dynamicPackageEnvs:   make(map[string]*runtime.Environment),
		executor:             exec,
		execMode:             mode,
		rootState:            newEvalState(),
		futureStatusStructs: map[string]*runtime.StructDefinitionValue{
			"Pending":   nil,
			"Resolved":  nil,
			"Cancelled": nil,
			"Failed":    nil,
		},
		channels:                                 make(map[int64]*channelState),
		pendingChannelSends:                      make(map[*runtime.FutureValue]*channelSendWaiter),
		pendingChannelReceives:                   make(map[*runtime.FutureValue]*channelReceiveWaiter),
		mutexes:                                  make(map[int64]*mutexState),
		concurrencyErrorStructs:                  make(map[string]*runtime.StructDefinitionValue),
		standardErrorStructs:                     make(map[string]*runtime.StructDefinitionValue),
		orderingStructs:                          make(map[string]*runtime.StructDefinitionValue),
		orderingValues:                           make(map[string]*runtime.StructInstanceValue),
		arraysByHandle:                           make(map[int64]arrayHandleTracking),
		errorNativeMethods:                       make(map[string]runtime.NativeFunctionValue),
		methodCache:                              make(map[methodCacheKey]methodCacheEntry),
		equalityDispatchCache:                    make(map[equalityDispatchCacheKey]equalityDispatchCacheEntry, equalityDispatchCacheInitialEntries),
		interfaceImplCache:                       make(map[interfaceImplCacheKey]interfaceImplCacheEntry),
		selectedInterfaceImplCache:               make(map[interfaceImplCacheKey]*selectedInterfaceImplCacheEntry),
		interfaceMethodDictionaryCache:           make(map[interfaceMethodDictionaryCacheKey]interfaceMethodDictionaryCacheEntry),
		implTargetMatchCache:                     make(map[implTargetMatchCacheKey]bool),
		boundMethodCache:                         make(map[boundMethodCacheKey]runtime.Value, boundMethodCacheInitialEntries),
		methodScopeCallableCache:                 make(map[methodScopeCallableCacheKey]methodScopeCallableCacheEntry, methodScopeLookupCacheInitialEntries),
		methodScopeHasCache:                      make(map[methodScopeHasCacheKey]methodScopeHasCacheEntry, methodScopeLookupCacheInitialEntries),
		propagationErrorCache:                    make(map[string]bool),
		overloadCache:                            make(map[overloadCacheKey]*runtime.FunctionValue),
		typeAliasBaseCache:                       make(map[string][]string),
		typeAliasReferenceCache:                  make(map[ast.TypeExpression]bool),
		typeAliasExpansionCache:                  make(map[ast.TypeExpression]ast.TypeExpression),
		functionCallGenericPlanCache:             make(map[ast.Node]*functionCallGenericPlan),
		callableExplicitRuntimeBindingUsageCache: make(map[ast.Node]bool),
		functionRuntimeGenericBindingPlanCache:   make(map[*runtime.FunctionValue]*functionRuntimeGenericBindingPlan),
		methodSetConstraintPlanCache:             make(map[*runtime.MethodSet]*methodSetConstraintPlan),
		functionCallConstraintResultCache:        make(map[functionCallConstraintResultCacheKey]functionCallConstraintResultCacheEntry),
		methodSetConstraintResultCache:           make(map[methodSetConstraintResultCacheKey]methodSetConstraintResultCacheEntry),
		namedStructLiteralPlanCache:              make(map[*ast.StructLiteral]namedStructLiteralPlan),
		namedStructPatternPlanCache:              make(map[namedStructPatternPlanCacheKey]namedStructPatternPlan),
		structGenericInferencePlanCache:          make(map[*ast.StructDefinition]*structGenericInferencePlan),
		blockTransientRuntimeScopeCache:          make(map[*ast.BlockExpression]bool),
		matchClauseScopePlanCache:                make(map[*ast.MatchClause]clauseScopePlan),
		matchExpressionClausePlansCache:          make(map[*ast.MatchExpression][]clauseScopePlan),
		matchExpressionBytecodeProgramsCache:     make(map[*ast.MatchExpression][]bytecodeMatchClausePrograms),
		rescueExpressionClausePlansCache:         make(map[*ast.RescueExpression][]clauseScopePlan),
		callTypeArgumentState:                    make(map[*ast.FunctionCall]callTypeArgumentState),
		inferredCallTypeArgumentCache:            make(map[inferredCallTypeArgumentCacheKey][]ast.TypeExpression),
		inferredCallTypeArgumentRuntimeCache1:    make(map[inferredCallTypeArgumentRuntimeCacheKey1][]ast.TypeExpression),
		inferredCallTypeArgumentRuntimeCache2:    make(map[inferredCallTypeArgumentRuntimeCacheKey2][]ast.TypeExpression),
		inferredCallTypeArgumentRuntimeCache:     make(map[inferredCallTypeArgumentRuntimeCacheKey][]ast.TypeExpression),
		explicitCallTypeBindingCache:             make(map[explicitCallTypeBindingCacheKey][]runtime.EnvironmentBinding),
		callLocalTypeBindingCache:                make(map[callLocalTypeBindingCacheKey][]runtime.EnvironmentBinding),
		reusableBytecodeCallEnvCache:             make(map[reusableBytecodeCallEnvCacheKey]*runtime.Environment),
		callableTransientCallEnvReuseCache:       make(map[ast.Node]bool),
		knownTypeNameCache:                       make(map[string]bool),
		typeInfoNameCache:                        make(map[typeInfoCacheKey]string),
		typeInfoExpressionCache:                  make(map[typeExpressionCacheKey]ast.TypeExpression),
		bytecodeExprCache:                        make(map[bytecodeExpressionProgramCacheKey]*bytecodeProgram),
		bytecodeLambdaCache:                      make(map[bytecodeLambdaProgramCacheKey]*bytecodeProgram),
		bytecodeLambdaDependencyNames:            make(map[*ast.LambdaExpression][]string),
		bytecodeStatsEnabled:                     os.Getenv("ABLE_BYTECODE_STATS") != "",
		bytecodePrimitiveMaterializationStatsEnabled: os.Getenv("ABLE_BYTECODE_STATS") != "" ||
			os.Getenv(bytecodePrimitiveMaterializationStatsEnv) != "",
		bytecodeTraceEnabled: os.Getenv("ABLE_BYTECODE_TRACE") != "",
	}
	if os.Getenv("ABLE_BYTECODE_STRING_STATS") != "" {
		i.bytecodeStringStats = &bytecodeStringStats{}
	}
	i.nativeCallContextPool.New = func() any {
		return &runtime.NativeCallContext{}
	}
	i.nativeBorrowCallArgScratchPool.New = func() any {
		return &nativeBorrowCallArgScratch{}
	}
	i.fixedCallArg2Pool.New = func() any {
		return &fixedCallArg2{}
	}
	i.transientClauseBindingPool.New = func() any {
		return &transientClauseBindingBuffer{}
	}
	i.initConcurrencyBuiltins()
	i.initChannelMutexBuiltins()
	i.initArrayBuiltins()
	i.initHashMapBuiltins()
	i.initStringHostBuiltins()
	i.initOsBuiltins()
	i.initErrorBuiltins()
	i.initRatioBuiltins()
	i.initInterfaceBuiltins()
	i.initDynamicBuiltins()
	i.global.SetSingleThread()
	i.envSingleThread = true
	return i
}

// ensureMultiThread switches the environment to multi-thread mode before the
// first concurrent spawn. This is a no-op if already in multi-thread mode.
func (i *Interpreter) ensureMultiThread() {
	if i.envSingleThread {
		i.global.SetMultiThread()
		i.envSingleThread = false
	}
}

// New returns a tree-walker interpreter with an empty global environment.
func New() *Interpreter {
	return newInterpreter(NewSerialExecutor(nil), execModeTreewalker)
}

// NewWithExecutor allows configuring the executor used for asynchronous tasks.
func NewWithExecutor(exec Executor) *Interpreter {
	return newInterpreter(exec, execModeTreewalker)
}

// NewBytecode returns a bytecode-backed interpreter.
func NewBytecode() *Interpreter {
	return newInterpreter(NewSerialExecutor(nil), execModeBytecode)
}

// NewBytecodeWithExecutor allows configuring the executor for bytecode runs.
func NewBytecodeWithExecutor(exec Executor) *Interpreter {
	return newInterpreter(exec, execModeBytecode)
}

// GlobalEnvironment returns the interpreter’s global environment.
func (i *Interpreter) GlobalEnvironment() *runtime.Environment {
	return i.global
}

// PackageEnvironment returns the environment for a named package if known.
// The empty package name resolves to the global environment.
func (i *Interpreter) PackageEnvironment(name string) *runtime.Environment {
	if i == nil {
		return nil
	}
	if name == "" {
		return i.global
	}
	if env, ok := i.packageEnvs[name]; ok {
		return env
	}
	if env, ok := i.dynamicPackageEnvs[name]; ok {
		return env
	}
	return nil
}

// SetArgs seeds os.args() for this interpreter run.
func (i *Interpreter) SetArgs(args []string) {
	if args == nil {
		i.osArgs = nil
		return
	}
	i.osArgs = append([]string{}, args...)
}

// SetNodeOrigins seeds per-node origin paths for diagnostic reporting.
func (i *Interpreter) SetNodeOrigins(origins map[ast.Node]string) {
	if origins == nil {
		i.nodeOrigins = nil
		return
	}
	copied := make(map[ast.Node]string, len(origins))
	for node, origin := range origins {
		copied[node] = origin
	}
	i.nodeOrigins = copied
}

// SetInterfaceMethodResolver registers a fallback resolver for interface methods.
// This is used by the AOT compiler to provide compiled interface dispatch when
// the interpreter's own impl registry has not been populated (no-bootstrap mode).
func (i *Interpreter) SetInterfaceMethodResolver(resolver func(receiver runtime.Value, interfaceName string, methodName string) (runtime.Value, bool)) {
	if i == nil {
		return
	}
	i.interfaceMethodResolver = resolver
	i.clearEqualityDispatchCache()
}

// SetCompiledImplChecker registers a callback that checks whether a given type
// implements a given interface via compiled dispatch tables. Used by the AOT
// compiler in no-bootstrap mode for truthiness checks (Error interface).
func (i *Interpreter) SetCompiledImplChecker(checker func(typeName string, interfaceName string) bool) {
	if i == nil {
		return
	}
	i.compiledImplChecker = checker
	i.invalidateMethodCache()
}

// SetCompiledInstanceMethodResolver registers a callback that resolves compiled
// instance methods by type and method name. Used for to_string and other inherent
// methods that are compiled but not in the interpreter's inherentMethods registry.
func (i *Interpreter) SetCompiledInstanceMethodResolver(resolver func(typeName string, methodName string) (runtime.Value, bool)) {
	if i == nil {
		return
	}
	i.compiledInstanceMethodFn = resolver
}

// SetCompiledInterfaceMemberResolver registers a callback that searches all
// compiled interface dispatch tables for a method by name, regardless of interface.
func (i *Interpreter) SetCompiledInterfaceMemberResolver(resolver func(receiver runtime.Value, methodName string) (runtime.Value, bool)) {
	if i == nil {
		return
	}
	i.compiledInterfaceMemberFn = resolver
}

// RegisterTypeAlias registers a type alias definition. Used by the AOT compiler
// in no-bootstrap mode to provide type alias expansion for bridge.MatchType.
func (i *Interpreter) RegisterTypeAlias(name string, alias *ast.TypeAliasDefinition) {
	i.setTypeAlias(name, alias)
}

func (i *Interpreter) setTypeAlias(name string, alias *ast.TypeAliasDefinition) {
	if i == nil || name == "" || alias == nil {
		return
	}
	if i.typeAliases == nil {
		i.typeAliases = make(map[string]*ast.TypeAliasDefinition)
	}
	i.typeAliases[name] = alias
	i.typeAliasCacheMu.Lock()
	clear(i.typeAliasBaseCache)
	clear(i.typeAliasReferenceCache)
	clear(i.typeAliasExpansionCache)
	i.typeAliasCacheMu.Unlock()
	i.explicitCallTypeBindingCacheMu.Lock()
	clear(i.explicitCallTypeBindingCache)
	i.explicitCallTypeBindingCacheMu.Unlock()
	i.callLocalTypeBindingCacheMu.Lock()
	clear(i.callLocalTypeBindingCache)
	i.callLocalTypeBindingCacheMu.Unlock()
	i.reusableBytecodeCallEnvCacheMu.Lock()
	clear(i.reusableBytecodeCallEnvCache)
	i.reusableBytecodeCallEnvCacheMu.Unlock()
	i.callableExplicitRuntimeBindingUsageCacheMu.Lock()
	clear(i.callableExplicitRuntimeBindingUsageCache)
	i.callableExplicitRuntimeBindingUsageCacheMu.Unlock()
	i.functionRuntimeGenericBindingPlanCacheMu.Lock()
	clear(i.functionRuntimeGenericBindingPlanCache)
	i.functionRuntimeGenericBindingPlanCacheMu.Unlock()
	i.methodCacheMu.Lock()
	clear(i.functionCallConstraintResultCache)
	clear(i.methodSetConstraintResultCache)
	i.methodCacheMu.Unlock()
}

// AddNodeOrigin registers a single node origin path for runtime diagnostics.
func (i *Interpreter) AddNodeOrigin(node ast.Node, origin string) {
	if i == nil || node == nil || origin == "" {
		return
	}
	if i.nodeOrigins == nil {
		i.nodeOrigins = make(map[ast.Node]string)
	}
	i.nodeOrigins[node] = origin
}

// ReserveNodeOrigins pre-sizes runtime diagnostic origin storage when a
// compiled launcher knows its complete generated node count.
func (i *Interpreter) ReserveNodeOrigins(capacity int) {
	if i == nil || capacity <= 0 || i.nodeOrigins != nil {
		return
	}
	i.nodeOrigins = make(map[ast.Node]string, capacity)
}

// EvaluateModule executes a module node and returns the last evaluated value and environment.
func (i *Interpreter) EvaluateModule(module *ast.Module) (runtime.Value, *runtime.Environment, error) {
	return i.evaluateModuleWithProgram(module, nil)
}

func (i *Interpreter) evaluateModuleWithProgram(module *ast.Module, program *bytecodeProgram) (runtime.Value, *runtime.Environment, error) {
	moduleEnv := i.global
	prevPackage := i.currentPackage
	defer func() { i.currentPackage = prevPackage }()

	restoreBytecodeInferenceFacts, typecheckErr := i.prepareModuleTypechecking(module)
	if typecheckErr != nil {
		return nil, nil, typecheckErr
	}
	if restoreBytecodeInferenceFacts != nil {
		defer restoreBytecodeInferenceFacts()
	}

	if module.Package != nil {
		pkgParts := identifiersToStrings(module.Package.NamePath)
		pkgName := strings.Join(pkgParts, ".")
		if i.dynamicDefinitionMode {
			if existing, ok := i.dynamicPackageEnvs[pkgName]; ok {
				moduleEnv = existing
			} else {
				moduleEnv = runtime.NewEnvironment(i.global)
				i.dynamicPackageEnvs[pkgName] = moduleEnv
			}
		} else {
			moduleEnv = runtime.NewEnvironment(i.global)
		}
		i.packageEnvs[pkgName] = moduleEnv
		i.currentPackage = pkgName
		if _, ok := i.packageRegistry[pkgName]; !ok {
			i.packageRegistry[pkgName] = make(map[string]runtime.Value)
		}
		i.packageMetadata[pkgName] = packageMeta{
			namePath:  pkgParts,
			isPrivate: module.Package.IsPrivate,
		}
	} else {
		i.currentPackage = ""
		i.packageEnvs[""] = moduleEnv
	}
	i.registerExternStatements(module)

	state := i.stateFromEnv(moduleEnv)
	for _, imp := range module.Imports {
		if _, err := i.evaluateImportStatement(imp, moduleEnv); err != nil {
			return nil, nil, i.attachRuntimeContext(err, imp, state)
		}
	}

	var (
		last runtime.Value
		err  error
	)
	if i.execMode == execModeBytecode {
		last, err = i.evaluateModuleBodyBytecodeWithProgram(module, moduleEnv, program)
	} else {
		last, err = i.evaluateModuleBodyTreewalker(module, moduleEnv)
	}
	if err != nil {
		if rs, ok := err.(raiseSignal); ok {
			return nil, moduleEnv, rs
		}
		if _, ok := err.(returnSignal); ok {
			return nil, nil, fmt.Errorf("return outside function")
		}
		return nil, nil, err
	}
	if err := i.evaluateModuleExports(module.Exports, moduleEnv); err != nil {
		return nil, nil, err
	}
	return last, moduleEnv, nil
}

func (i *Interpreter) evaluateModuleBodyTreewalker(module *ast.Module, env *runtime.Environment) (runtime.Value, error) {
	var last runtime.Value = runtime.NilValue{}
	for _, stmt := range module.Body {
		val, err := i.evaluateStatement(stmt, env)
		if err != nil {
			return nil, err
		}
		last = val
	}
	return last, nil
}

func (i *Interpreter) evaluateModuleBodyBytecode(module *ast.Module, env *runtime.Environment) (runtime.Value, error) {
	return i.evaluateModuleBodyBytecodeWithProgram(module, env, nil)
}

func (i *Interpreter) evaluateModuleBodyBytecodeWithProgram(module *ast.Module, env *runtime.Environment, program *bytecodeProgram) (runtime.Value, error) {
	var err error
	if program == nil {
		program, err = i.lowerModuleToBytecode(module)
		if err != nil {
			return nil, err
		}
	}
	vm := i.acquireBytecodeVM(env)
	defer i.releaseBytecodeVM(vm)
	return vm.run(program)
}

func (i *Interpreter) getPackageMeta(pkgName string, namePath []string) packageMeta {
	if meta, ok := i.packageMetadata[pkgName]; ok {
		return meta
	}
	dup := make([]string, len(namePath))
	copy(dup, namePath)
	return packageMeta{namePath: dup, isPrivate: false}
}
