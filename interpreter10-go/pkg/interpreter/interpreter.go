package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

type packageMeta struct {
	namePath  []string
	isPrivate bool
}

// Interpreter drives evaluation of Able v10 AST nodes.
type Interpreter struct {
	global          *runtime.Environment
	inherentMethods map[string]map[string]*runtime.FunctionValue
	interfaces      map[string]*runtime.InterfaceDefinitionValue
	implMethods     map[string][]implEntry
	unnamedImpls    map[string]map[string]map[string]struct{}
	raiseStack      []runtime.Value
	packageRegistry map[string]map[string]runtime.Value
	packageMetadata map[string]packageMeta
	currentPackage  string
	breakpoints     []string
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

func (i *Interpreter) pushBreakpoint(label string) {
	i.breakpoints = append(i.breakpoints, label)
}

func (i *Interpreter) popBreakpoint() {
	if len(i.breakpoints) == 0 {
		return
	}
	i.breakpoints = i.breakpoints[:len(i.breakpoints)-1]
}

func (i *Interpreter) hasBreakpoint(label string) bool {
	for idx := len(i.breakpoints) - 1; idx >= 0; idx-- {
		if i.breakpoints[idx] == label {
			return true
		}
	}
	return false
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
	return &Interpreter{
		global:          runtime.NewEnvironment(nil),
		inherentMethods: make(map[string]map[string]*runtime.FunctionValue),
		interfaces:      make(map[string]*runtime.InterfaceDefinitionValue),
		implMethods:     make(map[string][]implEntry),
		unnamedImpls:    make(map[string]map[string]map[string]struct{}),
		raiseStack:      make([]runtime.Value, 0),
		packageRegistry: make(map[string]map[string]runtime.Value),
		packageMetadata: make(map[string]packageMeta),
		breakpoints:     make([]string, 0),
	}
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
