package compiler

import (
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) interfaceDefinitionForPackage(pkgName string, name string) (*ast.InterfaceDefinition, string, bool) {
	if g == nil || strings.TrimSpace(name) == "" {
		return nil, "", false
	}
	pkgName = strings.TrimSpace(pkgName)
	if pkgName != "" && g.interfacesByPackage != nil {
		if defs := g.interfacesByPackage[pkgName]; defs != nil {
			if def := defs[name]; def != nil {
				return def, pkgName, true
			}
		}
		var importedDef *ast.InterfaceDefinition
		importedPkg := ""
		for _, binding := range g.staticImportBindingsForName(pkgName, name) {
			sourcePkg := strings.TrimSpace(binding.SourcePackage)
			sourceName := strings.TrimSpace(binding.SourceName)
			if sourcePkg == "" || sourceName == "" {
				continue
			}
			defs := g.interfacesByPackage[sourcePkg]
			def := defs[sourceName]
			if def == nil {
				continue
			}
			if importedDef != nil && (importedDef != def || importedPkg != sourcePkg) {
				return nil, "", false
			}
			importedDef = def
			importedPkg = sourcePkg
		}
		if importedDef != nil {
			return importedDef, importedPkg, true
		}
		if sourcePkg, sourceName, ok := g.sourceReexportSourceForName(pkgName, name); ok {
			if defs := g.interfacesByPackage[sourcePkg]; defs != nil {
				if def := defs[sourceName]; def != nil {
					return def, sourcePkg, true
				}
			}
		}
	}
	if def := g.interfaces[name]; def != nil {
		resolvedPkg := strings.TrimSpace(g.interfacePackages[name])
		if resolvedPkg == "" {
			resolvedPkg = pkgName
		}
		return def, resolvedPkg, true
	}
	return nil, "", false
}

func (g *generator) interfaceDefinedInPackage(pkgName string, name string) bool {
	if g == nil || strings.TrimSpace(pkgName) == "" || strings.TrimSpace(name) == "" || g.interfacesByPackage == nil {
		return false
	}
	defs := g.interfacesByPackage[strings.TrimSpace(pkgName)]
	return defs != nil && defs[name] != nil
}

func (g *generator) interfacePackageForName(pkgName string, name string) string {
	if _, resolvedPkg, ok := g.interfaceDefinitionForPackage(pkgName, name); ok {
		return resolvedPkg
	}
	return strings.TrimSpace(pkgName)
}

func runtimeInterfaceIdentity(pkgName string, name string) string {
	pkgName = strings.TrimSpace(pkgName)
	name = strings.TrimSpace(name)
	if pkgName == "" || name == "" {
		return name
	}
	return pkgName + "." + name
}

func (g *generator) interfaceDefinitionForImpl(impl *implMethodInfo) (*ast.InterfaceDefinition, string, bool) {
	if impl == nil {
		return nil, "", false
	}
	pkgName := ""
	if impl.Info != nil {
		pkgName = impl.Info.Package
	}
	return g.interfaceDefinitionForPackage(pkgName, impl.InterfaceName)
}
