package compiler

import "strings"

type importedSelectorTypeAliasCacheEntry struct {
	SourcePackage string
	SourceName    string
}

type sourceReexportResolutionCacheEntry struct {
	SourcePackage string
	SourceName    string
	Resolved      bool
}

type importResolutionCacheKey struct {
	PackageName string
	Name        string
}

func newImportResolutionCacheKey(pkgName string, name string) importResolutionCacheKey {
	return importResolutionCacheKey{
		PackageName: strings.TrimSpace(pkgName),
		Name:        strings.TrimSpace(name),
	}
}

func (g *generator) invalidateImportResolutionCaches() {
	if g == nil {
		return
	}
	g.importedSelectorTypeAliasCache = make(map[importResolutionCacheKey]importedSelectorTypeAliasCacheEntry)
	g.typeExprPackageCache = make(map[typeExprPackageCacheKey]typeExprPackageCacheEntry)
	g.sourceReexportResolutionCache = make(map[importResolutionCacheKey]sourceReexportResolutionCacheEntry)
	g.importableNameSetCache = make(map[string]map[string]struct{})
}
