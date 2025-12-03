package interpreter

import (
	"fmt"
	"strings"
	"sync"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
	"able/interpreter10-go/pkg/typechecker"
)

type packageMeta struct {
	namePath  []string
	isPrivate bool
}

type evalState struct {
	raiseStack        []runtime.Value
	breakpoints       []string
	implicitReceivers []runtime.Value
	topicStack        []runtime.Value
	topicUsage        []bool
	placeholderStack  []placeholderFrame
}

func newEvalState() *evalState {
	return &evalState{
		raiseStack:        make([]runtime.Value, 0),
		breakpoints:       make([]string, 0),
		implicitReceivers: make([]runtime.Value, 0),
		topicStack:        make([]runtime.Value, 0),
		topicUsage:        make([]bool, 0),
		placeholderStack:  make([]placeholderFrame, 0),
	}
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

func (s *evalState) pushTopic(val runtime.Value) {
	if s == nil {
		return
	}
	s.topicStack = append(s.topicStack, val)
	s.topicUsage = append(s.topicUsage, false)
}

func (s *evalState) popTopic() {
	if s == nil || len(s.topicStack) == 0 {
		return
	}
	s.topicStack = s.topicStack[:len(s.topicStack)-1]
	s.topicUsage = s.topicUsage[:len(s.topicUsage)-1]
}

func (s *evalState) currentTopic() (runtime.Value, bool) {
	if s == nil || len(s.topicStack) == 0 {
		return nil, false
	}
	return s.topicStack[len(s.topicStack)-1], true
}

func (s *evalState) markTopicUsed() {
	if s == nil || len(s.topicUsage) == 0 {
		return
	}
	s.topicUsage[len(s.topicUsage)-1] = true
}

func (s *evalState) topicWasUsed() bool {
	if s == nil || len(s.topicUsage) == 0 {
		return false
	}
	return s.topicUsage[len(s.topicUsage)-1]
}

func (s *evalState) pushPlaceholderFrame(explicit map[int]struct{}, paramCount int, args []runtime.Value) {
	if s == nil {
		return
	}
	frameExplicit := make(map[int]struct{}, len(explicit))
	for idx := range explicit {
		frameExplicit[idx] = struct{}{}
	}
	frame := placeholderFrame{
		args:             args,
		explicit:         frameExplicit,
		implicitAssigned: make(map[int]struct{}),
		nextImplicit:     1,
		paramCount:       paramCount,
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

type placeholderFrame struct {
	args             []runtime.Value
	explicit         map[int]struct{}
	implicitAssigned map[int]struct{}
	nextImplicit     int
	paramCount       int
}

func (f *placeholderFrame) valueAt(index int) (runtime.Value, error) {
	if index <= 0 || index > len(f.args) {
		return nil, fmt.Errorf("Placeholder index @%d is out of range", index)
	}
	return f.args[index-1], nil
}

func (f *placeholderFrame) nextImplicitIndex() (int, error) {
	if f == nil {
		return 0, fmt.Errorf("placeholder frame missing")
	}
	idx := f.nextImplicit
	if idx < 1 {
		idx = 1
	}
	for idx <= f.paramCount {
		if _, reserved := f.explicit[idx]; reserved {
			idx++
			continue
		}
		if _, used := f.implicitAssigned[idx]; used {
			idx++
			continue
		}
		f.implicitAssigned[idx] = struct{}{}
		f.nextImplicit = idx + 1
		return idx, nil
	}
	return 0, fmt.Errorf("no implicit placeholder slots available")
}

// Interpreter drives evaluation of Able v10 AST nodes.
type Interpreter struct {
	global               *runtime.Environment
	inherentMethods      map[string]map[string]runtime.Value
	interfaces           map[string]*runtime.InterfaceDefinitionValue
	implMethods          map[string][]implEntry
	rangeImplementations []rangeImplementation
	unnamedImpls         map[string]map[string]map[string]struct{}
	packageRegistry      map[string]map[string]runtime.Value
	packageMetadata      map[string]packageMeta
	currentPackage       string
	executor             Executor
	rootState            *evalState

	concurrencyReady     bool
	procErrorStruct      *runtime.StructDefinitionValue
	procStatusStructs    map[string]*runtime.StructDefinitionValue
	procStatusPending    runtime.Value
	procStatusResolved   runtime.Value
	procStatusCancelled  runtime.Value
	awaitWakerStruct     *runtime.StructDefinitionValue
	awaitRoundRobinIndex int

	channelMutexReady       bool
	channelMu               sync.Mutex
	channels                map[int64]*channelState
	nextChannelHandle       int64
	pendingChannelSends     map[*runtime.ProcHandleValue]*channelSendWaiter
	pendingChannelReceives  map[*runtime.ProcHandleValue]*channelReceiveWaiter
	mutexMu                 sync.Mutex
	mutexes                 map[int64]*mutexState
	nextMutexHandle         int64
	concurrencyErrorStructs map[string]*runtime.StructDefinitionValue

	stringHostReady bool

	hasherReady      bool
	hasherMu         sync.Mutex
	hashers          map[int64]*hasherState
	nextHasherHandle int64

	orderingStructs map[string]*runtime.StructDefinitionValue

	arrayReady      bool
	arrayStates     map[int64]*arrayState
	arraysByHandle  map[int64]map[*runtime.ArrayValue]struct{}
	nextArrayHandle int64
	hashMapReady    bool

	errorNativeMethods map[string]runtime.NativeFunctionValue

	generatorStack []*generatorInstance

	interfaceBuiltinsReady bool

	typecheckerEnabled   bool
	typecheckerStrict    bool
	typechecker          *typechecker.Checker
	typecheckDiagnostics []typechecker.Diagnostic
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
		if data := env.RuntimeData(); data != nil {
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
	if existing, ok := bucket[name]; ok {
		if merged, ok := runtime.MergeFunctionValues(existing, value); ok {
			value = merged
		}
	}
	bucket[name] = value
	if qn := i.qualifiedName(name); qn != "" {
		i.global.Define(qn, value)
	}
}

// New returns an interpreter with an empty global environment.
func New() *Interpreter {
	return NewWithExecutor(NewSerialExecutor(nil))
}

// NewWithExecutor allows configuring the executor used for asynchronous tasks.
func NewWithExecutor(exec Executor) *Interpreter {
	if exec == nil {
		exec = NewSerialExecutor(nil)
	}
	i := &Interpreter{
		global:               runtime.NewEnvironment(nil),
		inherentMethods:      make(map[string]map[string]runtime.Value),
		interfaces:           make(map[string]*runtime.InterfaceDefinitionValue),
		implMethods:          make(map[string][]implEntry),
		rangeImplementations: make([]rangeImplementation, 0),
		unnamedImpls:         make(map[string]map[string]map[string]struct{}),
		packageRegistry:      make(map[string]map[string]runtime.Value),
		packageMetadata:      make(map[string]packageMeta),
		executor:             exec,
		rootState:            newEvalState(),
		procStatusStructs: map[string]*runtime.StructDefinitionValue{
			"Pending":   nil,
			"Resolved":  nil,
			"Cancelled": nil,
			"Failed":    nil,
		},
		channels:                make(map[int64]*channelState),
		pendingChannelSends:     make(map[*runtime.ProcHandleValue]*channelSendWaiter),
		pendingChannelReceives:  make(map[*runtime.ProcHandleValue]*channelReceiveWaiter),
		mutexes:                 make(map[int64]*mutexState),
		concurrencyErrorStructs: make(map[string]*runtime.StructDefinitionValue),
		hashers:                 make(map[int64]*hasherState),
		orderingStructs:         make(map[string]*runtime.StructDefinitionValue),
		arrayStates:             make(map[int64]*arrayState),
		arraysByHandle:          make(map[int64]map[*runtime.ArrayValue]struct{}),
		nextArrayHandle:         1,
		errorNativeMethods:      make(map[string]runtime.NativeFunctionValue),
	}
	i.initConcurrencyBuiltins()
	i.initChannelMutexBuiltins()
	i.initArrayBuiltins()
	i.initHashMapBuiltins()
	i.initStringHostBuiltins()
	i.initErrorBuiltins()
	i.initHasherBuiltins()
	i.initInterfaceBuiltins()
	return i
}

// GlobalEnvironment returns the interpreter’s global environment.
func (i *Interpreter) GlobalEnvironment() *runtime.Environment {
	return i.global
}

// EvaluateModule executes a module node and returns the last evaluated value and environment.
func (i *Interpreter) EvaluateModule(module *ast.Module) (runtime.Value, *runtime.Environment, error) {
	moduleEnv := i.global
	prevPackage := i.currentPackage
	defer func() { i.currentPackage = prevPackage }()

	i.typecheckDiagnostics = nil
	if i.typecheckerEnabled {
		if module == nil {
			return nil, nil, fmt.Errorf("typechecker: module is nil")
		}
		// When evaluating standalone modules without a prelude, fall back to the
		// legacy per-module typecheck. Callers that use MultiModuleEvaluator should
		// seed the typechecker explicitly before invoking EvaluateModule.
		if i.typechecker == nil {
			i.typechecker = typechecker.New()
		}
		diags, err := i.typechecker.CheckModule(module)
		if err != nil {
			return nil, nil, err
		}
		i.typecheckDiagnostics = append(i.typecheckDiagnostics[:0], diags...)
		if i.typecheckerStrict && len(diags) > 0 {
			msg := diags[0].Message
			if !strings.HasPrefix(msg, "typechecker:") {
				msg = "typechecker: " + msg
			}
			return nil, nil, fmt.Errorf("%s", msg)
		}
	}

	if module.Package != nil {
		moduleEnv = runtime.NewEnvironment(i.global)
		pkgParts := identifiersToStrings(module.Package.NamePath)
		pkgName := strings.Join(pkgParts, ".")
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
	}

	for _, imp := range module.Imports {
		if _, err := i.evaluateImportStatement(imp, moduleEnv); err != nil {
			return nil, nil, err
		}
	}

	var last runtime.Value = runtime.NilValue{}
	for _, stmt := range module.Body {
		val, err := i.evaluateStatement(stmt, moduleEnv)
		if err != nil {
			if rs, ok := err.(raiseSignal); ok {
				return nil, moduleEnv, rs
			}
			if _, ok := err.(returnSignal); ok {
				return nil, nil, fmt.Errorf("return outside function")
			}
			return nil, nil, err
		}
		last = val
	}
	return last, moduleEnv, nil
}

func (i *Interpreter) getPackageMeta(pkgName string, namePath []string) packageMeta {
	if meta, ok := i.packageMetadata[pkgName]; ok {
		return meta
	}
	dup := make([]string, len(namePath))
	copy(dup, namePath)
	return packageMeta{namePath: dup, isPrivate: false}
}
