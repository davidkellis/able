package interpreter

import (
	"fmt"
	"plugin"
	"reflect"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type externTargetState struct {
	preludes   []string
	externs    []*ast.ExternFunctionBody
	externByID map[string]int
}

type externHostPackage struct {
	targets map[ast.HostTarget]*externTargetState
	modules map[ast.HostTarget]*externHostModule
}

type externHostModule struct {
	hash    string
	plugin  *plugin.Plugin
	symbols map[string]reflect.Value
}

const externCacheVersion = "v2"

func (i *Interpreter) isKernelExtern(name string) bool {
	return strings.HasPrefix(name, "__able_")
}

func (i *Interpreter) registerExternStatements(module *ast.Module) {
	if module == nil {
		return
	}
	i.externHostMu.Lock()
	defer i.externHostMu.Unlock()
	pkgName := i.currentPackage
	if pkgName == "" {
		pkgName = "<root>"
	}
	if i.externHostPackages == nil {
		i.externHostPackages = make(map[string]*externHostPackage)
	}
	pkg := i.externHostPackages[pkgName]
	if pkg == nil {
		pkg = &externHostPackage{
			targets: make(map[ast.HostTarget]*externTargetState),
			modules: make(map[ast.HostTarget]*externHostModule),
		}
		i.externHostPackages[pkgName] = pkg
	}
	for _, stmt := range module.Body {
		switch s := stmt.(type) {
		case *ast.PreludeStatement:
			if s == nil {
				continue
			}
			state := ensureExternTarget(pkg, s.Target)
			state.preludes = append(state.preludes, s.Code)
		case *ast.ExternFunctionBody:
			if s == nil || s.Signature == nil || s.Signature.ID == nil {
				continue
			}
			name := s.Signature.ID.Name
			if name == "" {
				continue
			}
			state := ensureExternTarget(pkg, s.Target)
			if idx, ok := state.externByID[name]; ok {
				state.externs[idx] = s
			} else {
				state.externByID[name] = len(state.externs)
				state.externs = append(state.externs, s)
			}
		}
	}
}

func ensureExternTarget(pkg *externHostPackage, target ast.HostTarget) *externTargetState {
	if pkg.targets == nil {
		pkg.targets = make(map[ast.HostTarget]*externTargetState)
	}
	state := pkg.targets[target]
	if state == nil {
		state = &externTargetState{externByID: make(map[string]int)}
		pkg.targets[target] = state
	}
	return state
}

func (i *Interpreter) invokeExternHostFunction(pkgName string, def *ast.ExternFunctionBody, args []runtime.Value) (runtime.Value, error) {
	if def == nil || def.Signature == nil || def.Signature.ID == nil {
		return runtime.NilValue{}, nil
	}
	if pkgName == "" {
		pkgName = "<root>"
	}
	i.externHostMu.Lock()
	pkg := i.externHostPackages[pkgName]
	if pkg == nil {
		i.externHostMu.Unlock()
		return nil, fmt.Errorf("extern package %s is not registered", pkgName)
	}
	targetState := pkg.targets[def.Target]
	if targetState == nil {
		i.externHostMu.Unlock()
		return nil, fmt.Errorf("extern target %s is not registered", def.Target)
	}
	module, err := i.ensureExternHostModule(pkgName, def.Target, targetState, pkg)
	i.externHostMu.Unlock()
	if err != nil {
		return nil, err
	}
	fn, err := module.lookup(def.Signature.ID.Name)
	if err != nil {
		return nil, err
	}
	fnType := fn.Type()
	paramCount := fnType.NumIn()
	if len(args) != paramCount {
		return nil, fmt.Errorf("extern function %s expects %d args, got %d", def.Signature.ID.Name, paramCount, len(args))
	}
	callArgs := make([]reflect.Value, paramCount)
	for idx := 0; idx < paramCount; idx++ {
		hostVal, convErr := i.toHostValue(def.Signature.Params[idx].ParamType, args[idx], fnType.In(idx))
		if convErr != nil {
			return nil, convErr
		}
		callArgs[idx] = hostVal
	}

	var results []reflect.Value
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("extern panic: %v", r)
			}
		}()
		if err == nil {
			results = fn.Call(callArgs)
		}
	}()
	if err != nil {
		return nil, err
	}
	return i.fromHostResults(def, results)
}

func (i *Interpreter) ensureExternHostModule(pkgName string, target ast.HostTarget, state *externTargetState, pkg *externHostPackage) (*externHostModule, error) {
	hash := hashExternState(target, state)
	if existing := pkg.modules[target]; existing != nil && existing.hash == hash && existing.plugin != nil {
		return existing, nil
	}
	module, err := buildExternModule(pkgName, target, state, hash)
	if err != nil {
		return nil, err
	}
	pkg.modules[target] = module
	return module, nil
}
