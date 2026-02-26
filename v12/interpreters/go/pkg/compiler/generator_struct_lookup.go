package compiler

import (
	"sort"
	"strings"
)

func (g *generator) structInfoByNameUnique(name string) (*structInfo, bool) {
	if g == nil || strings.TrimSpace(name) == "" {
		return nil, false
	}
	var found *structInfo
	for _, info := range g.structs {
		if info == nil || strings.TrimSpace(info.Name) != strings.TrimSpace(name) {
			continue
		}
		if found != nil && found != info {
			return nil, false
		}
		found = info
	}
	return found, found != nil
}

func (g *generator) structInfoForTypeName(pkgName string, typeName string) (*structInfo, bool) {
	if g == nil {
		return nil, false
	}
	pkgName = strings.TrimSpace(pkgName)
	typeName = strings.TrimSpace(typeName)
	if typeName == "" {
		return nil, false
	}
	if pkgName != "" {
		if info := g.structs[qualifiedName(pkgName, typeName)]; info != nil {
			return info, true
		}
		selectorMatches := make([]*structInfo, 0, 1)
		for _, binding := range g.staticImports[pkgName] {
			if binding.Kind != staticImportBindingSelector {
				continue
			}
			if strings.TrimSpace(binding.LocalName) != typeName {
				continue
			}
			sourcePkg := strings.TrimSpace(binding.SourcePackage)
			sourceName := strings.TrimSpace(binding.SourceName)
			if sourcePkg == "" || sourceName == "" {
				continue
			}
			if info := g.structs[qualifiedName(sourcePkg, sourceName)]; info != nil {
				selectorMatches = append(selectorMatches, info)
			}
		}
		if len(selectorMatches) == 1 {
			return selectorMatches[0], true
		}
		if len(selectorMatches) > 1 {
			return nil, false
		}
		var wildcardMatch *structInfo
		for _, binding := range g.staticImports[pkgName] {
			if binding.Kind != staticImportBindingWildcard {
				continue
			}
			sourcePkg := strings.TrimSpace(binding.SourcePackage)
			if sourcePkg == "" {
				continue
			}
			info := g.structs[qualifiedName(sourcePkg, typeName)]
			if info == nil {
				continue
			}
			if wildcardMatch != nil && wildcardMatch != info {
				return nil, false
			}
			wildcardMatch = info
		}
		if wildcardMatch != nil {
			return wildcardMatch, true
		}
	}
	return g.structInfoByNameUnique(typeName)
}

func (g *generator) sortedStructKeys() []string {
	if g == nil || len(g.structs) == 0 {
		return nil
	}
	keys := make([]string, 0, len(g.structs))
	for key := range g.structs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (g *generator) sortedUniqueStructNames() []string {
	if g == nil || len(g.structs) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(g.structs))
	names := make([]string, 0, len(g.structs))
	for _, info := range g.structs {
		if info == nil || strings.TrimSpace(info.Name) == "" {
			continue
		}
		if _, ok := seen[info.Name]; ok {
			continue
		}
		seen[info.Name] = struct{}{}
		names = append(names, info.Name)
	}
	sort.Strings(names)
	return names
}
