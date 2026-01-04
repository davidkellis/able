package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/parser"
	"able/interpreter-go/pkg/runtime"
)

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
			code, ok := asStringValue(args[1])
			if !ok {
				return runtime.ErrorValue{Message: "dyn.Package.def expects String"}, nil
			}
			return i.evaluateDynamicDefinition(pkgName, code), nil
		},
	}
	i.dynPackageDefMethod = defMethod

	packageFn := runtime.NativeFunctionValue{
		Name:  "dyn.package",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			name, ok := asStringValue(firstArg(args))
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
			name, ok := asStringValue(firstArg(args))
			if !ok {
				return runtime.ErrorValue{Message: "dyn.def_package expects String"}, nil
			}
			i.ensureDynamicPackage(name)
			return runtime.DynPackageValue{Name: name}, nil
		},
	}

	dynPkg := runtime.PackageValue{
		Name:      "dyn",
		NamePath:  []string{"dyn"},
		IsPrivate: false,
		Public: map[string]runtime.Value{
			"package":     packageFn,
			"def_package": defPackageFn,
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

func asStringValue(val runtime.Value) (string, bool) {
	switch v := val.(type) {
	case runtime.StringValue:
		return v.Val, true
	case *runtime.StringValue:
		if v == nil {
			return "", false
		}
		return v.Val, true
	default:
		return "", false
	}
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

func (i *Interpreter) evaluateDynamicDefinition(pkgName, source string) runtime.Value {
	if pkgName == "" {
		return runtime.ErrorValue{Message: "dyn.def requires package name"}
	}
	i.ensureDynamicPackage(pkgName)
	mod, err := parseDynamicModule(source)
	if err != nil {
		return runtime.ErrorValue{Message: fmt.Sprintf("dyn.def parse error: %s", err.Error())}
	}
	baseParts := strings.Split(pkgName, ".")
	var targetParts []string
	if mod.Package != nil {
		pkgParts := identifiersToStrings(mod.Package.NamePath)
		targetParts = resolveDynamicPackage(baseParts, pkgParts)
	} else {
		targetParts = baseParts
	}
	mod.Package = ast.NewPackageStatement(stringsToIdentifiers(targetParts), false)

	prevDynamic := i.dynamicDefinitionMode
	prevTypecheck := i.typecheckerEnabled
	prevStrict := i.typecheckerStrict
	i.dynamicDefinitionMode = true
	i.typecheckerEnabled = false
	i.typecheckerStrict = false
	_, _, evalErr := i.EvaluateModule(mod)
	i.dynamicDefinitionMode = prevDynamic
	i.typecheckerEnabled = prevTypecheck
	i.typecheckerStrict = prevStrict

	if evalErr != nil {
		switch v := evalErr.(type) {
		case raiseSignal:
			return i.makeErrorValue(v.value, i.global)
		default:
			return runtime.ErrorValue{Message: fmt.Sprintf("dyn.def error: %s", evalErr.Error())}
		}
	}
	return runtime.NilValue{}
}

func parseDynamicModule(source string) (*ast.Module, error) {
	p, err := parser.NewModuleParser()
	if err != nil {
		return nil, err
	}
	defer p.Close()
	return p.ParseModule([]byte(source))
}

func resolveDynamicPackage(base, target []string) []string {
	if len(target) == 0 {
		return base
	}
	if len(target) > 0 && target[0] == "root" {
		return target
	}
	if len(base) > 0 && len(target) >= len(base) {
		matches := true
		for i := range base {
			if target[i] != base[i] {
				matches = false
				break
			}
		}
		if matches {
			return target
		}
	}
	combined := make([]string, 0, len(base)+len(target))
	combined = append(combined, base...)
	combined = append(combined, target...)
	return combined
}

func stringsToIdentifiers(parts []string) []*ast.Identifier {
	idents := make([]*ast.Identifier, 0, len(parts))
	for _, part := range parts {
		idents = append(idents, ast.NewIdentifier(part))
	}
	return idents
}
