//go:build js && wasm

package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/runtime"
)

const wasmDynamicSourceUnsupported = "dynamic source parsing is unavailable on js/wasm; provide pre-parsed AST modules"

func (i *Interpreter) initDynamicBuiltins() {
	defMethod := runtime.NativeFunctionValue{
		Name:  "dyn.Package.def",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) < 2 {
				return runtime.ErrorValue{Message: "dyn.Package.def expects code"}, nil
			}
			pkgName, ok := dynPackageName(args[0])
			if !ok {
				return runtime.ErrorValue{Message: "dyn.Package.def called on non-dyn package"}, nil
			}
			code, ok := asStringValue(i, args[1])
			if !ok {
				return runtime.ErrorValue{Message: "dyn.Package.def expects String"}, nil
			}
			return i.evaluateDynamicDefinition(pkgName, code), nil
		},
	}
	i.dynPackageDefMethod = defMethod

	evalMethod := runtime.NativeFunctionValue{
		Name:  "dyn.Package.eval",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) < 2 {
				return runtime.ErrorValue{Message: "dyn.Package.eval expects code"}, nil
			}
			pkgName, ok := dynPackageName(args[0])
			if !ok {
				return runtime.ErrorValue{Message: "dyn.Package.eval called on non-dyn package"}, nil
			}
			code, ok := asStringValue(i, args[1])
			if !ok {
				return runtime.ErrorValue{Message: "dyn.Package.eval expects String"}, nil
			}
			return i.evaluateDynamicEval(pkgName, code), nil
		},
	}
	i.dynPackageEvalMethod = evalMethod

	packageFn := runtime.NativeFunctionValue{
		Name:  "dyn.package",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			name, ok := asStringValue(i, firstArg(args))
			if !ok {
				return runtime.ErrorValue{Message: "dyn.package expects String"}, nil
			}
			if _, ok := i.packageRegistry[name]; !ok {
				return runtime.ErrorValue{Message: fmt.Sprintf("dyn.package: package '%s' not found", name)}, nil
			}
			return runtime.DynPackageValue{Name: name}, nil
		},
	}

	defPackageFn := runtime.NativeFunctionValue{
		Name:  "dyn.def_package",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			name, ok := asStringValue(i, firstArg(args))
			if !ok {
				return runtime.ErrorValue{Message: "dyn.def_package expects String"}, nil
			}
			i.ensureDynamicPackage(name)
			return runtime.DynPackageValue{Name: name}, nil
		},
	}

	evalFn := runtime.NativeFunctionValue{
		Name:  "dyn.eval",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			code, ok := asStringValue(i, firstArg(args))
			if !ok {
				return runtime.ErrorValue{Message: "dyn.eval expects String"}, nil
			}
			return i.evaluateDynamicEval("dyn.eval", code), nil
		},
	}

	dynPkg := runtime.PackageValue{
		Name:      "dyn",
		NamePath:  []string{"dyn"},
		IsPrivate: false,
		Public: map[string]runtime.Value{
			"package":     packageFn,
			"def_package": defPackageFn,
			"eval":        evalFn,
		},
	}
	i.global.Define("dyn", dynPkg)
}

func firstArg(args []runtime.Value) runtime.Value {
	if len(args) == 0 {
		return nil
	}
	return args[0]
}

func asStringValue(i *Interpreter, val runtime.Value) (string, bool) {
	for {
		if iface, ok := val.(*runtime.InterfaceValue); ok && iface != nil {
			val = iface.Underlying
			continue
		}
		if iface, ok := val.(runtime.InterfaceValue); ok {
			val = iface.Underlying
			continue
		}
		break
	}
	str, err := i.coerceStringValue(val)
	if err != nil {
		return "", false
	}
	return str, true
}

func dynPackageName(val runtime.Value) (string, bool) {
	switch v := val.(type) {
	case runtime.DynPackageValue:
		if v.Name != "" {
			return v.Name, true
		}
		if len(v.NamePath) > 0 {
			return strings.Join(v.NamePath, "."), true
		}
	case *runtime.DynPackageValue:
		if v == nil {
			return "", false
		}
		if v.Name != "" {
			return v.Name, true
		}
		if len(v.NamePath) > 0 {
			return strings.Join(v.NamePath, "."), true
		}
	}
	return "", false
}

func (i *Interpreter) ensureDynamicPackage(name string) {
	if name == "" {
		return
	}
	if _, ok := i.packageRegistry[name]; !ok {
		i.packageRegistry[name] = make(map[string]runtime.Value)
	}
	if _, ok := i.packageMetadata[name]; !ok {
		parts := strings.Split(name, ".")
		i.packageMetadata[name] = packageMeta{namePath: parts, isPrivate: false}
	}
}

func (i *Interpreter) evaluateDynamicDefinition(pkgName, _ string) runtime.Value {
	if pkgName == "" {
		return runtime.ErrorValue{Message: "dyn.def requires package name"}
	}
	i.ensureDynamicPackage(pkgName)
	return runtime.ErrorValue{Message: "dyn.def unsupported in wasm runtime: " + wasmDynamicSourceUnsupported}
}

func (i *Interpreter) evaluateDynamicEval(pkgName, _ string) runtime.Value {
	if pkgName == "" {
		return runtime.ErrorValue{Message: "dyn.eval requires package name"}
	}
	i.ensureDynamicPackage(pkgName)
	return runtime.ErrorValue{Message: "dyn.eval unsupported in wasm runtime: " + wasmDynamicSourceUnsupported}
}
