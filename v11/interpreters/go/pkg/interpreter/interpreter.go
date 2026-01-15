package interpreter

import (
	"fmt"
	"strings"
	"sync"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
	"able/interpreter-go/pkg/typechecker"
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
}

func newEvalState() *evalState {
	return &evalState{
		raiseStack:        make([]runtime.Value, 0),
		breakpoints:       make([]string, 0),
		implicitReceivers: make([]runtime.Value, 0),
		placeholderStack:  make([]placeholderFrame, 0),
		blockFrames:       make(map[*ast.BlockExpression]*blockFrame),
		callStack:         make([]runtimeCallFrame, 0),
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
	s.callStack = append(s.callStack, runtimeCallFrame{node: node})
}

func (s *evalState) popCallFrame() {
	if s == nil || len(s.callStack) == 0 {
		return
	}
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

// Interpreter drives evaluation of Able v11 AST nodes.
type Interpreter struct {
	global                *runtime.Environment
	inherentMethods       map[string]map[string]runtime.Value
	interfaces            map[string]*runtime.InterfaceDefinitionValue
	unionDefinitions      map[string]*runtime.UnionDefinitionValue
	typeAliases           map[string]*ast.TypeAliasDefinition
	implMethods           map[string][]implEntry
	genericImpls          []implEntry
	rangeImplementations  []rangeImplementation
	unnamedImpls          map[string]map[string]map[string]bool
	packageRegistry       map[string]map[string]runtime.Value
	packageMetadata       map[string]packageMeta
	externHostPackages    map[string]*externHostPackage
	currentPackage        string
	dynamicDefinitionMode bool
	dynPackageDefMethod   runtime.NativeFunctionValue
	dynPackageEvalMethod  runtime.NativeFunctionValue
	dynamicPackageEnvs    map[string]*runtime.Environment
	executor              Executor
	rootState             *evalState
	nodeOrigins           map[ast.Node]string

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
	standardErrorStructs    map[string]*runtime.StructDefinitionValue

	stringHostReady bool
	osReady         bool
	osArgs          []string

	hasherReady      bool
	hasherMu         sync.Mutex
	hashers          map[int64]*hasherState
	nextHasherHandle int64
	ratioReady       bool

	orderingStructs map[string]*runtime.StructDefinitionValue
	divModStruct    *runtime.StructDefinitionValue
	ratioStruct     *runtime.StructDefinitionValue

	arrayReady        bool
	arrayStates       map[int64]*arrayState
	arraysByHandle    map[int64]map[*runtime.ArrayValue]struct{}
	nextArrayHandle   int64
	hashMapReady      bool
	hashMapStates     map[int64]*runtime.HashMapValue
	nextHashMapHandle int64

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
	if !i.dynamicDefinitionMode {
		if existing, ok := bucket[name]; ok {
			if merged, ok := runtime.MergeFunctionValues(existing, value); ok {
				value = merged
			}
		}
	}
	bucket[name] = value
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
		unionDefinitions:     make(map[string]*runtime.UnionDefinitionValue),
		typeAliases:          make(map[string]*ast.TypeAliasDefinition),
		implMethods:          make(map[string][]implEntry),
		genericImpls:         make([]implEntry, 0),
		rangeImplementations: make([]rangeImplementation, 0),
		unnamedImpls:         make(map[string]map[string]map[string]bool),
		packageRegistry:      make(map[string]map[string]runtime.Value),
		packageMetadata:      make(map[string]packageMeta),
		externHostPackages:   make(map[string]*externHostPackage),
		dynamicPackageEnvs:   make(map[string]*runtime.Environment),
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
		standardErrorStructs:    make(map[string]*runtime.StructDefinitionValue),
		hashers:                 make(map[int64]*hasherState),
		orderingStructs:         make(map[string]*runtime.StructDefinitionValue),
		arrayStates:             make(map[int64]*arrayState),
		arraysByHandle:          make(map[int64]map[*runtime.ArrayValue]struct{}),
		nextArrayHandle:         1,
		hashMapStates:           make(map[int64]*runtime.HashMapValue),
		nextHashMapHandle:       1,
		errorNativeMethods:      make(map[string]runtime.NativeFunctionValue),
	}
	i.initConcurrencyBuiltins()
	i.initChannelMutexBuiltins()
	i.initArrayBuiltins()
	i.initHashMapBuiltins()
	i.initStringHostBuiltins()
	i.initOsBuiltins()
	i.initErrorBuiltins()
	i.initHasherBuiltins()
	i.initRatioBuiltins()
	i.initInterfaceBuiltins()
	i.initDynamicBuiltins()
	return i
}

// GlobalEnvironment returns the interpreter’s global environment.
func (i *Interpreter) GlobalEnvironment() *runtime.Environment {
	return i.global
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
	i.registerExternStatements(module)

	state := i.stateFromEnv(moduleEnv)
	for _, imp := range module.Imports {
		if _, err := i.evaluateImportStatement(imp, moduleEnv); err != nil {
			return nil, nil, i.attachRuntimeContext(err, imp, state)
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
