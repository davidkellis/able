package compiler

import (
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type staticImportBindingKind string

const (
	staticImportBindingPackage  staticImportBindingKind = "package"
	staticImportBindingSelector staticImportBindingKind = "selector"
	staticImportBindingWildcard staticImportBindingKind = "wildcard"
)

type staticImportBinding struct {
	Kind          staticImportBindingKind
	SourcePackage string
	LocalName     string
	SourceName    string
}

func (g *generator) collectStaticImportsForPackage(pkgName string, imports []*ast.ImportStatement) {
	if g == nil || len(imports) == 0 {
		return
	}
	if g.staticImports == nil {
		g.staticImports = make(map[string][]staticImportBinding)
	}
	for _, imp := range imports {
		if imp == nil {
			continue
		}
		sourcePkg := importPathString(imp.PackagePath)
		if sourcePkg == "" {
			continue
		}
		if imp.IsWildcard {
			g.addStaticImportBinding(pkgName, staticImportBinding{
				Kind:          staticImportBindingWildcard,
				SourcePackage: sourcePkg,
			})
			continue
		}
		if len(imp.Selectors) > 0 {
			for _, sel := range imp.Selectors {
				if sel == nil || sel.Name == nil || strings.TrimSpace(sel.Name.Name) == "" {
					continue
				}
				local := sel.Name.Name
				if sel.Alias != nil && strings.TrimSpace(sel.Alias.Name) != "" {
					local = strings.TrimSpace(sel.Alias.Name)
				}
				g.addStaticImportBinding(pkgName, staticImportBinding{
					Kind:          staticImportBindingSelector,
					SourcePackage: sourcePkg,
					LocalName:     local,
					SourceName:    strings.TrimSpace(sel.Name.Name),
				})
			}
			continue
		}
		local := importPackageAlias(imp.PackagePath, imp.Alias)
		if local == "" {
			continue
		}
		g.addStaticImportBinding(pkgName, staticImportBinding{
			Kind:          staticImportBindingPackage,
			SourcePackage: sourcePkg,
			LocalName:     local,
		})
	}
}

// collectSourceReexports records the import binding behind each explicit
// source export. A re-export is an additional package name for the original
// definition, not a wrapper definition, so later static resolution follows
// this binding back to the original package.
func (g *generator) collectSourceReexportsForPackage(pkgName string, exports []*ast.ExportStatement) {
	if g == nil || len(exports) == 0 {
		return
	}
	if g.sourceReexports == nil {
		g.sourceReexports = make(map[string][]staticImportBinding)
	}
	for _, exp := range exports {
		if exp == nil {
			continue
		}
		if exp.IsWildcard {
			sourcePkg := importPathString(exp.PackagePath)
			if sourcePkg == "" {
				continue
			}
			g.addSourceReexportBinding(pkgName, staticImportBinding{
				Kind:          staticImportBindingWildcard,
				SourcePackage: sourcePkg,
			})
			continue
		}
		if exp.Name == nil || strings.TrimSpace(exp.Name.Name) == "" {
			continue
		}
		name := strings.TrimSpace(exp.Name.Name)
		for _, binding := range g.staticImportsForPackage(pkgName) {
			switch binding.Kind {
			case staticImportBindingSelector:
				if strings.TrimSpace(binding.LocalName) != name {
					continue
				}
				g.addSourceReexportBinding(pkgName, binding)
			case staticImportBindingWildcard:
				g.addSourceReexportBinding(pkgName, staticImportBinding{
					Kind:          staticImportBindingSelector,
					SourcePackage: binding.SourcePackage,
					LocalName:     name,
					SourceName:    name,
				})
			}
		}
	}
}

func (g *generator) addStaticImportBinding(pkgName string, binding staticImportBinding) {
	if g == nil || binding.Kind == "" || binding.SourcePackage == "" {
		return
	}
	bucket := g.staticImports[pkgName]
	for _, existing := range bucket {
		if existing == binding {
			return
		}
	}
	g.staticImports[pkgName] = append(bucket, binding)
	g.invalidateImportResolutionCaches()
}

func (g *generator) addSourceReexportBinding(pkgName string, binding staticImportBinding) {
	if g == nil || binding.Kind == "" || strings.TrimSpace(binding.SourcePackage) == "" {
		return
	}
	bucket := g.sourceReexports[pkgName]
	for _, existing := range bucket {
		if existing == binding {
			return
		}
	}
	g.sourceReexports[pkgName] = append(bucket, binding)
	g.invalidateImportResolutionCaches()
}

func importPathString(path []*ast.Identifier) string {
	if len(path) == 0 {
		return ""
	}
	parts := make([]string, 0, len(path))
	for _, part := range path {
		if part == nil {
			continue
		}
		name := strings.TrimSpace(part.Name)
		if name == "" {
			continue
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, ".")
}

func importPackageAlias(path []*ast.Identifier, alias *ast.Identifier) string {
	if alias != nil {
		if trimmed := strings.TrimSpace(alias.Name); trimmed != "" {
			return trimmed
		}
	}
	for idx := len(path) - 1; idx >= 0; idx-- {
		if path[idx] == nil {
			continue
		}
		if trimmed := strings.TrimSpace(path[idx].Name); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (g *generator) staticImportsForPackage(pkgName string) []staticImportBinding {
	if g == nil || len(g.staticImports) == 0 {
		return nil
	}
	// Bindings are append-only during collection and read-only afterward.
	// Returning the owned slice avoids copying it on every type lookup.
	return g.staticImports[pkgName]
}

func (g *generator) sourceReexportsForPackage(pkgName string) []staticImportBinding {
	if g == nil || len(g.sourceReexports) == 0 {
		return nil
	}
	// Re-export bindings follow the same append-then-read lifecycle as imports.
	return g.sourceReexports[pkgName]
}

func (g *generator) staticImportBindingsForName(pkgName string, name string) []staticImportBinding {
	if g == nil || strings.TrimSpace(pkgName) == "" || strings.TrimSpace(name) == "" {
		return nil
	}
	name = strings.TrimSpace(name)
	var out []staticImportBinding
	for _, binding := range g.staticImportsForPackage(pkgName) {
		switch binding.Kind {
		case staticImportBindingSelector:
			if strings.TrimSpace(binding.LocalName) == name {
				out = append(out, binding)
			}
		case staticImportBindingWildcard:
			if _, exported := g.importableNameSet(binding.SourcePackage)[name]; !exported {
				continue
			}
			out = append(out, staticImportBinding{
				Kind:          staticImportBindingSelector,
				SourcePackage: binding.SourcePackage,
				LocalName:     name,
				SourceName:    name,
			})
		}
	}
	return out
}

func (g *generator) sourceReexportBindingsForName(pkgName string, name string) []staticImportBinding {
	if g == nil || strings.TrimSpace(pkgName) == "" || strings.TrimSpace(name) == "" {
		return nil
	}
	name = strings.TrimSpace(name)
	var out []staticImportBinding
	for _, binding := range g.sourceReexportsForPackage(pkgName) {
		switch binding.Kind {
		case staticImportBindingSelector:
			if strings.TrimSpace(binding.LocalName) == name {
				out = append(out, binding)
			}
		case staticImportBindingWildcard:
			if _, exported := g.importableNameSet(binding.SourcePackage)[name]; !exported {
				continue
			}
			out = append(out, staticImportBinding{
				Kind:          staticImportBindingSelector,
				SourcePackage: binding.SourcePackage,
				LocalName:     name,
				SourceName:    name,
			})
		}
	}
	return out
}

func (g *generator) sourceReexportSourceForName(pkgName string, name string) (string, string, bool) {
	if g == nil {
		return "", "", false
	}
	key := newImportResolutionCacheKey(pkgName, name)
	if cached, ok := g.sourceReexportResolutionCache[key]; ok {
		return cached.SourcePackage, cached.SourceName, cached.Resolved
	}
	sourcePkg, sourceName, resolved := g.sourceReexportSourceForNameSeen(pkgName, name, make(map[string]struct{}))
	g.sourceReexportResolutionCache[key] = sourceReexportResolutionCacheEntry{
		SourcePackage: sourcePkg,
		SourceName:    sourceName,
		Resolved:      resolved,
	}
	return sourcePkg, sourceName, resolved
}

func (g *generator) sourceReexportSourceForNameSeen(pkgName string, name string, seen map[string]struct{}) (string, string, bool) {
	if g == nil {
		return "", "", false
	}
	pkgName = strings.TrimSpace(pkgName)
	name = strings.TrimSpace(name)
	if pkgName == "" || name == "" {
		return "", "", false
	}
	key := pkgName + "::" + name
	if _, exists := seen[key]; exists {
		return "", "", false
	}
	seen[key] = struct{}{}
	defer delete(seen, key)
	var sourcePkg, sourceName string
	for _, binding := range g.sourceReexportBindingsForName(pkgName, name) {
		candidatePkg := strings.TrimSpace(binding.SourcePackage)
		candidateName := strings.TrimSpace(binding.SourceName)
		if candidatePkg == "" || candidateName == "" {
			continue
		}
		if nestedPkg, nestedName, ok := g.sourceReexportSourceForNameSeen(candidatePkg, candidateName, seen); ok {
			candidatePkg, candidateName = nestedPkg, nestedName
		}
		if sourcePkg != "" && (sourcePkg != candidatePkg || sourceName != candidateName) {
			return "", "", false
		}
		sourcePkg, sourceName = candidatePkg, candidateName
	}
	return sourcePkg, sourceName, sourcePkg != "" && sourceName != ""
}

func (g *generator) noBootstrapImportsSeedable() bool {
	if g == nil {
		return false
	}
	if !g.namedImplNamespacesSeedable() {
		return false
	}
	if len(g.staticImports) == 0 {
		return true
	}
	knownPackages := g.knownPackageNames()
	exportsByPackage := make(map[string]map[string]struct{}, len(knownPackages))
	for pkgName := range knownPackages {
		exportsByPackage[pkgName] = g.importableNameSet(pkgName)
	}
	for targetPkg, bindings := range g.staticImports {
		if len(bindings) == 0 {
			continue
		}
		if _, ok := knownPackages[targetPkg]; !ok {
			return false
		}
		for _, binding := range bindings {
			if binding.SourcePackage == "" {
				return false
			}
			if _, ok := knownPackages[binding.SourcePackage]; !ok {
				return false
			}
			switch binding.Kind {
			case staticImportBindingPackage:
				if strings.TrimSpace(binding.LocalName) == "" {
					return false
				}
			case staticImportBindingSelector:
				exports := exportsByPackage[binding.SourcePackage]
				if _, ok := exports[binding.SourceName]; !ok && !g.externCallableExists(binding.SourcePackage, binding.SourceName) {
					return false
				}
			case staticImportBindingWildcard:
				continue
			default:
				return false
			}
		}
	}
	return true
}

func (g *generator) knownPackageNames() map[string]struct{} {
	packages := make(map[string]struct{})
	if g == nil {
		return packages
	}
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		packages[name] = struct{}{}
	}
	add(g.entryPackage)
	for _, pkgName := range g.packages {
		add(pkgName)
	}
	for pkgName := range g.functions {
		add(pkgName)
	}
	for pkgName := range g.overloads {
		add(pkgName)
	}
	for _, info := range g.structs {
		if info == nil {
			continue
		}
		add(info.Package)
	}
	for _, pkgName := range g.interfacePackages {
		add(pkgName)
	}
	for _, pkgName := range g.unionPackages {
		add(pkgName)
	}
	return packages
}

func (g *generator) importableNameSet(pkgName string) map[string]struct{} {
	if g == nil {
		return nil
	}
	pkgName = strings.TrimSpace(pkgName)
	if g.importableNameSetCache == nil {
		g.importableNameSetCache = make(map[string]map[string]struct{})
	}
	if cached, ok := g.importableNameSetCache[pkgName]; ok {
		return cached
	}
	set := make(map[string]struct{})
	for _, name := range g.sortedImportableNames(pkgName) {
		set[name] = struct{}{}
	}
	g.importableNameSetCache[pkgName] = set
	return set
}

func (g *generator) sortedImportableNames(pkgName string) []string {
	return g.sortedImportableNamesSeen(pkgName, make(map[string]struct{}))
}

func (g *generator) sortedImportableNamesSeen(pkgName string, seenPackages map[string]struct{}) []string {
	if g == nil {
		return nil
	}
	pkgName = strings.TrimSpace(pkgName)
	if pkgName == "" {
		return nil
	}
	if _, seen := seenPackages[pkgName]; seen {
		return nil
	}
	seenPackages[pkgName] = struct{}{}
	defer delete(seenPackages, pkgName)
	callables := g.sortedPublicCallableNames(pkgName)
	structs := g.sortedPublicStructNames(pkgName)
	interfaces := g.sortedPublicInterfaceNames(pkgName)
	unions := g.sortedPublicUnionNames(pkgName)
	aliases := g.sortedPublicTypeAliasNames(pkgName)
	implNamespaces := g.sortedPublicImplNamespaceNames(pkgName)
	seen := make(map[string]struct{}, len(callables)+len(structs)+len(interfaces)+len(unions)+len(aliases)+len(implNamespaces))
	names := make([]string, 0, len(callables)+len(structs)+len(interfaces)+len(unions)+len(aliases)+len(implNamespaces))
	add := func(name string) {
		if strings.TrimSpace(name) == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	for _, name := range callables {
		add(name)
	}
	for _, name := range structs {
		add(name)
	}
	for _, name := range interfaces {
		add(name)
	}
	for _, name := range unions {
		add(name)
	}
	for _, name := range aliases {
		add(name)
	}
	for _, name := range implNamespaces {
		add(name)
	}
	for _, binding := range g.sourceReexportsForPackage(pkgName) {
		switch binding.Kind {
		case staticImportBindingSelector:
			add(binding.LocalName)
		case staticImportBindingWildcard:
			for _, name := range g.sortedImportableNamesSeen(binding.SourcePackage, seenPackages) {
				add(name)
			}
		}
	}
	sort.Strings(names)
	return names
}

func (g *generator) sortedExportedCallableNames(pkgName string) []string {
	return g.sortedExportedCallableNamesSeen(pkgName, make(map[string]struct{}))
}

func (g *generator) sortedExportedCallableNamesSeen(pkgName string, seenPackages map[string]struct{}) []string {
	if g == nil {
		return nil
	}
	pkgName = strings.TrimSpace(pkgName)
	if pkgName == "" {
		return nil
	}
	if _, seen := seenPackages[pkgName]; seen {
		return nil
	}
	seenPackages[pkgName] = struct{}{}
	defer delete(seenPackages, pkgName)
	seenNames := make(map[string]struct{})
	var names []string
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		if _, exists := seenNames[name]; exists {
			return
		}
		seenNames[name] = struct{}{}
		names = append(names, name)
	}
	for _, name := range g.sortedPublicCallableNames(pkgName) {
		add(name)
	}
	for _, binding := range g.sourceReexportsForPackage(pkgName) {
		switch binding.Kind {
		case staticImportBindingSelector:
			if g.callableExistsSeen(binding.SourcePackage, binding.SourceName, seenPackages) {
				add(binding.LocalName)
			}
		case staticImportBindingWildcard:
			for _, name := range g.sortedExportedCallableNamesSeen(binding.SourcePackage, seenPackages) {
				add(name)
			}
		}
	}
	sort.Strings(names)
	return names
}

func (g *generator) sortedPublicCallableNames(pkgName string) []string {
	if g == nil {
		return nil
	}
	names := g.sortedCallableNames(pkgName)
	methodNames := g.sortedPublicMethodCallableNames(pkgName)
	if len(names) == 0 && len(methodNames) == 0 {
		return nil
	}
	out := make([]string, 0, len(names)+len(methodNames))
	for _, name := range names {
		if g.isCallablePublic(pkgName, name) {
			out = append(out, name)
		}
	}
	out = append(out, methodNames...)
	sort.Strings(out)
	return out
}

func (g *generator) isCallablePublic(pkgName string, name string) bool {
	if g == nil || strings.TrimSpace(name) == "" {
		return false
	}
	if pkgFuncs := g.functions[pkgName]; pkgFuncs != nil {
		if info := pkgFuncs[name]; info != nil && info.Definition != nil {
			return !info.Definition.IsPrivate
		}
	}
	if pkgOverloads := g.overloads[pkgName]; pkgOverloads != nil {
		if overload := pkgOverloads[name]; overload != nil {
			for _, entry := range overload.Entries {
				if entry == nil || entry.Definition == nil {
					continue
				}
				if !entry.Definition.IsPrivate {
					return true
				}
			}
		}
	}
	return false
}

func (g *generator) sortedPublicStructNames(pkgName string) []string {
	if g == nil || strings.TrimSpace(pkgName) == "" {
		return nil
	}
	infos := g.sortedStructInfosForPackage(pkgName)
	if len(infos) == 0 {
		return nil
	}
	names := make([]string, 0, len(infos))
	for _, info := range infos {
		if info == nil || info.Node == nil || info.Node.IsPrivate || strings.TrimSpace(info.Name) == "" {
			continue
		}
		names = append(names, info.Name)
	}
	sort.Strings(names)
	return names
}

func (g *generator) sortedPublicInterfaceNames(pkgName string) []string {
	if g == nil || strings.TrimSpace(pkgName) == "" {
		return nil
	}
	interfaces := g.sortedInterfaceDefsForPackage(pkgName)
	if len(interfaces) == 0 {
		return nil
	}
	names := make([]string, 0, len(interfaces))
	for _, def := range interfaces {
		if def == nil || def.ID == nil || def.IsPrivate || strings.TrimSpace(def.ID.Name) == "" {
			continue
		}
		names = append(names, def.ID.Name)
	}
	sort.Strings(names)
	return names
}

func (g *generator) sortedPublicUnionNames(pkgName string) []string {
	if g == nil || strings.TrimSpace(pkgName) == "" {
		return nil
	}
	unions := g.sortedUnionDefsForPackage(pkgName)
	if len(unions) == 0 {
		return nil
	}
	names := make([]string, 0, len(unions))
	for _, def := range unions {
		if def == nil || def.ID == nil || def.IsPrivate || strings.TrimSpace(def.ID.Name) == "" {
			continue
		}
		names = append(names, def.ID.Name)
	}
	sort.Strings(names)
	return names
}

func (g *generator) sortedPublicTypeAliasNames(pkgName string) []string {
	if g == nil || strings.TrimSpace(pkgName) == "" {
		return nil
	}
	aliases := g.typeAliases[pkgName]
	if len(aliases) == 0 {
		return nil
	}
	names := make([]string, 0, len(aliases))
	for name := range aliases {
		if strings.TrimSpace(name) == "" {
			continue
		}
		if g.typeAliasPrivate[pkgName] != nil && g.typeAliasPrivate[pkgName][name] {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (g *generator) sortedPublicImplNamespaceNames(pkgName string) []string {
	if g == nil || strings.TrimSpace(pkgName) == "" {
		return nil
	}
	infos := g.sortedNamedImplNamespacesForPackage(pkgName)
	if len(infos) == 0 {
		return nil
	}
	names := make([]string, 0, len(infos))
	for _, info := range infos {
		if info == nil || info.Invalid || strings.TrimSpace(info.Name) == "" {
			continue
		}
		names = append(names, info.Name)
	}
	sort.Strings(names)
	return names
}
