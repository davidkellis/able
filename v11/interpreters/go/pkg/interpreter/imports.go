package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func isPrivateSymbol(val runtime.Value) bool {
	switch v := val.(type) {
	case *runtime.FunctionValue, *runtime.FunctionOverloadValue:
		if fn := firstFunction(v); fn != nil {
			if def, ok := fn.Declaration.(*ast.FunctionDefinition); ok {
				return def.IsPrivate
			}
		}
	case *runtime.StructDefinitionValue:
		return v.Node != nil && v.Node.IsPrivate
	case runtime.StructDefinitionValue:
		return v.Node != nil && v.Node.IsPrivate
	case *runtime.InterfaceDefinitionValue:
		return v.Node != nil && v.Node.IsPrivate
	case runtime.InterfaceDefinitionValue:
		return v.Node != nil && v.Node.IsPrivate
	case *runtime.UnionDefinitionValue:
		return v.Node != nil && v.Node.IsPrivate
	case runtime.UnionDefinitionValue:
		return v.Node != nil && v.Node.IsPrivate
	}
	return false
}

func importPrivacyError(name string, val runtime.Value) error {
	switch v := val.(type) {
	case *runtime.FunctionValue, *runtime.FunctionOverloadValue:
		if fn := firstFunction(v); fn != nil {
			if def, ok := fn.Declaration.(*ast.FunctionDefinition); ok && def.IsPrivate {
				return fmt.Errorf("Import error: function '%s' is private", name)
			}
		}
	case *runtime.StructDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("Import error: struct '%s' is private", name)
		}
	case runtime.StructDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("Import error: struct '%s' is private", name)
		}
	case *runtime.InterfaceDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("Import error: interface '%s' is private", name)
		}
	case runtime.InterfaceDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("Import error: interface '%s' is private", name)
		}
	case *runtime.UnionDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("Import error: union '%s' is private", name)
		}
	case runtime.UnionDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("Import error: union '%s' is private", name)
		}
	}
	return fmt.Errorf("Import error: symbol '%s' is private", name)
}

func dynImportPrivacyError(name string, val runtime.Value) error {
	switch v := val.(type) {
	case *runtime.FunctionValue, *runtime.FunctionOverloadValue:
		if fn := firstFunction(v); fn != nil {
			if def, ok := fn.Declaration.(*ast.FunctionDefinition); ok && def.IsPrivate {
				return fmt.Errorf("dynimport error: function '%s' is private", name)
			}
		}
	case *runtime.StructDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("dynimport error: struct '%s' is private", name)
		}
	case runtime.StructDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("dynimport error: struct '%s' is private", name)
		}
	case *runtime.InterfaceDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("dynimport error: interface '%s' is private", name)
		}
	case runtime.InterfaceDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("dynimport error: interface '%s' is private", name)
		}
	case *runtime.UnionDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("dynimport error: union '%s' is private", name)
		}
	case runtime.UnionDefinitionValue:
		if v.Node != nil && v.Node.IsPrivate {
			return fmt.Errorf("dynimport error: union '%s' is private", name)
		}
	}
	return fmt.Errorf("dynimport error: symbol '%s' is private", name)
}

func copyPublicSymbols(bucket map[string]runtime.Value) map[string]runtime.Value {
	public := make(map[string]runtime.Value)
	for name, val := range bucket {
		if isPrivateSymbol(val) {
			continue
		}
		public[name] = val
	}
	return public
}

func (i *Interpreter) evaluateImportStatement(imp *ast.ImportStatement, env *runtime.Environment) (runtime.Value, error) {
	return i.processImport(imp.PackagePath, imp.IsWildcard, imp.Selectors, imp.Alias, env, false)
}

func (i *Interpreter) evaluateDynImportStatement(imp *ast.DynImportStatement, env *runtime.Environment) (runtime.Value, error) {
	return i.processImport(imp.PackagePath, imp.IsWildcard, imp.Selectors, imp.Alias, env, true)
}

func (i *Interpreter) processImport(packagePath []*ast.Identifier, isWildcard bool, selectors []*ast.ImportSelector, alias *ast.Identifier, env *runtime.Environment, dynamic bool) (runtime.Value, error) {
	pkgParts := identifiersToStrings(packagePath)
	pkgName := strings.Join(pkgParts, ".")

	if dynamic {
		return i.processDynImport(pkgName, pkgParts, isWildcard, selectors, alias, env)
	}

	if alias != nil && !isWildcard && len(selectors) == 0 {
		bucket, ok := i.packageRegistry[pkgName]
		if !ok {
			return nil, fmt.Errorf("Import error: package '%s' not found", pkgName)
		}
		public := copyPublicSymbols(bucket)
		meta := i.getPackageMeta(pkgName, pkgParts)
		env.Define(alias.Name, runtime.PackageValue{
			Name:      pkgName,
			NamePath:  meta.namePath,
			IsPrivate: meta.isPrivate,
			Public:    public,
		})
		return runtime.NilValue{}, nil
	}

	if isWildcard {
		bucket, ok := i.packageRegistry[pkgName]
		if !ok {
			return nil, fmt.Errorf("Import error: package '%s' not found", pkgName)
		}
		for name, val := range bucket {
			if isPrivateSymbol(val) {
				continue
			}
			env.Define(name, val)
		}
		return runtime.NilValue{}, nil
	}

	if len(selectors) > 0 {
		for _, sel := range selectors {
			if sel == nil || sel.Name == nil {
				return nil, fmt.Errorf("Import selector missing name")
			}
			original := sel.Name.Name
			aliasName := original
			if sel.Alias != nil {
				aliasName = sel.Alias.Name
			}
			val, err := i.lookupImportSymbol(pkgName, original)
			if err != nil {
				return nil, err
			}
			if isPrivateSymbol(val) {
				return nil, importPrivacyError(original, val)
			}
			env.Define(aliasName, val)
		}
		return runtime.NilValue{}, nil
	}

	if pkgName != "" && alias == nil {
		bucket, ok := i.packageRegistry[pkgName]
		if !ok {
			return nil, fmt.Errorf("Import error: package '%s' not found", pkgName)
		}
		public := copyPublicSymbols(bucket)
		meta := i.getPackageMeta(pkgName, pkgParts)
		env.Define(pkgName, runtime.PackageValue{
			Name:      pkgName,
			NamePath:  meta.namePath,
			IsPrivate: meta.isPrivate,
			Public:    public,
		})
	}

	return runtime.NilValue{}, nil
}

func (i *Interpreter) processDynImport(pkgName string, pkgParts []string, isWildcard bool, selectors []*ast.ImportSelector, alias *ast.Identifier, env *runtime.Environment) (runtime.Value, error) {
	bucket, ok := i.packageRegistry[pkgName]
	if !ok {
		return nil, fmt.Errorf("dynimport error: package '%s' not found", pkgName)
	}

	if alias != nil && !isWildcard && len(selectors) == 0 {
		meta := i.getPackageMeta(pkgName, pkgParts)
		env.Define(alias.Name, runtime.DynPackageValue{
			Name:      pkgName,
			NamePath:  meta.namePath,
			IsPrivate: meta.isPrivate,
		})
		return runtime.NilValue{}, nil
	}

	if isWildcard {
		for name, val := range bucket {
			if isPrivateSymbol(val) {
				continue
			}
			env.Define(name, runtime.DynRefValue{Package: pkgName, Name: name})
		}
		return runtime.NilValue{}, nil
	}

	if len(selectors) > 0 {
		for _, sel := range selectors {
			if sel == nil || sel.Name == nil {
				return nil, fmt.Errorf("dynimport selector missing name")
			}
			original := sel.Name.Name
			aliasName := original
			if sel.Alias != nil {
				aliasName = sel.Alias.Name
			}
			sym, ok := bucket[original]
			if !ok {
				return nil, fmt.Errorf("dynimport error: '%s' not found in '%s'", original, pkgName)
			}
			if isPrivateSymbol(sym) {
				return nil, dynImportPrivacyError(original, sym)
			}
			env.Define(aliasName, runtime.DynRefValue{Package: pkgName, Name: original})
		}
		return runtime.NilValue{}, nil
	}

	if pkgName != "" && alias == nil {
		meta := i.getPackageMeta(pkgName, pkgParts)
		env.Define(pkgName, runtime.DynPackageValue{
			Name:      pkgName,
			NamePath:  meta.namePath,
			IsPrivate: meta.isPrivate,
		})
	}

	return runtime.NilValue{}, nil
}

func (i *Interpreter) lookupImportSymbol(pkgName, symbol string) (runtime.Value, error) {
	if pkgName != "" {
		if bucket, ok := i.packageRegistry[pkgName]; ok {
			if val, ok := bucket[symbol]; ok {
				return val, nil
			}
		}
		if val, err := i.global.Get(pkgName + "." + symbol); err == nil {
			return val, nil
		}
	}
	if val, err := i.global.Get(symbol); err == nil {
		return val, nil
	}
	if pkgName != "" {
		return nil, fmt.Errorf("Import error: symbol '%s' from '%s' not found in globals", symbol, pkgName)
	}
	return nil, fmt.Errorf("Import error: symbol '%s' not found in globals", symbol)
}

func (i *Interpreter) evaluateRethrowStatement(_ *ast.RethrowStatement, env *runtime.Environment) (runtime.Value, error) {
	state := i.stateFromEnv(env)
	if val, ok := state.peekRaise(); ok {
		return nil, raiseSignal{value: val}
	}
	return nil, raiseSignal{value: runtime.ErrorValue{Message: "Unknown rethrow"}}
}
