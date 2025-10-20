package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
	"able/interpreter10-go/pkg/typechecker"
)

type packageMeta struct {
	namePath  []string
	isPrivate bool
}

type evalState struct {
	raiseStack  []runtime.Value
	breakpoints []string
}

func newEvalState() *evalState {
	return &evalState{
		raiseStack:  make([]runtime.Value, 0),
		breakpoints: make([]string, 0),
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

// Interpreter drives evaluation of Able v10 AST nodes.
type Interpreter struct {
	global          *runtime.Environment
	inherentMethods map[string]map[string]*runtime.FunctionValue
	interfaces      map[string]*runtime.InterfaceDefinitionValue
	implMethods     map[string][]implEntry
	unnamedImpls    map[string]map[string]map[string]struct{}
	packageRegistry map[string]map[string]runtime.Value
	packageMetadata map[string]packageMeta
	currentPackage  string
	executor        Executor
	rootState       *evalState

	concurrencyReady    bool
	procErrorStruct     *runtime.StructDefinitionValue
	procStatusStructs   map[string]*runtime.StructDefinitionValue
	procStatusPending   runtime.Value
	procStatusResolved  runtime.Value
	procStatusCancelled runtime.Value

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
		global:          runtime.NewEnvironment(nil),
		inherentMethods: make(map[string]map[string]*runtime.FunctionValue),
		interfaces:      make(map[string]*runtime.InterfaceDefinitionValue),
		implMethods:     make(map[string][]implEntry),
		unnamedImpls:    make(map[string]map[string]map[string]struct{}),
		packageRegistry: make(map[string]map[string]runtime.Value),
		packageMetadata: make(map[string]packageMeta),
		executor:        exec,
		rootState:       newEvalState(),
		procStatusStructs: map[string]*runtime.StructDefinitionValue{
			"Pending":   nil,
			"Resolved":  nil,
			"Cancelled": nil,
			"Failed":    nil,
		},
	}
	i.initConcurrencyBuiltins()
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
