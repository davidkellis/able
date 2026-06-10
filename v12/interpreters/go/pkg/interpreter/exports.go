package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// evaluateModuleExports publishes explicit source re-exports only after the
// module's declarations have run. This keeps `export Name` order-independent
// while preserving the original runtime value and nominal identity.
func (i *Interpreter) evaluateModuleExports(exports []*ast.ExportStatement, env *runtime.Environment) error {
	if i == nil || len(exports) == 0 {
		return nil
	}
	for _, export := range exports {
		if export == nil {
			continue
		}
		if !export.IsWildcard {
			if export.Name == nil || strings.TrimSpace(export.Name.Name) == "" {
				return fmt.Errorf("Export error: named export requires a binding")
			}
			value, err := env.Get(export.Name.Name)
			if err != nil || value == nil {
				return fmt.Errorf("Export error: symbol '%s' is not defined", export.Name.Name)
			}
			if isPrivateSymbol(value) {
				return fmt.Errorf("Export error: symbol '%s' is private", export.Name.Name)
			}
			i.registerSymbol(export.Name.Name, value)
			continue
		}

		pkgName := strings.Join(identifiersToStrings(export.PackagePath), ".")
		if pkgName == "" {
			return fmt.Errorf("Export error: wildcard export requires a package")
		}
		bucket, ok := i.packageRegistry[pkgName]
		if !ok {
			return fmt.Errorf("Export error: package '%s' not found", pkgName)
		}
		for name, value := range bucket {
			if name == "" || value == nil || isPrivateSymbol(value) {
				continue
			}
			i.registerSymbol(name, value)
		}
	}
	return nil
}
