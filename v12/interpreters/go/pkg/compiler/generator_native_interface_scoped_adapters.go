package compiler

import (
	"sort"
	"strings"
)

func (g *generator) recordNativeInterfaceExplicitAdapter(info *nativeInterfaceInfo, adapter *nativeInterfaceAdapter) {
	if g == nil || info == nil || adapter == nil || adapter.GoType == "" {
		return
	}
	if g.nativeInterfaceExplicitAdapters == nil {
		g.nativeInterfaceExplicitAdapters = make(map[string]map[string]*nativeInterfaceAdapter)
	}
	adapters := g.nativeInterfaceExplicitAdapters[info.Key]
	if adapters == nil {
		adapters = make(map[string]*nativeInterfaceAdapter)
		g.nativeInterfaceExplicitAdapters[info.Key] = adapters
	}
	key := nativeInterfaceAdapterIdentity(adapter)
	adapters[key] = adapter
	for _, existing := range info.Adapters {
		if existing != nil && nativeInterfaceAdapterIdentity(existing) == key {
			return
		}
	}
	info.Adapters = append(info.Adapters, adapter)
}

func nativeInterfaceAdapterIdentity(adapter *nativeInterfaceAdapter) string {
	if adapter == nil {
		return ""
	}
	if adapter.AdapterType != "" {
		return adapter.AdapterType
	}
	return adapter.GoType
}

func (g *generator) nativeInterfaceKnownAdapters(info *nativeInterfaceInfo) []*nativeInterfaceAdapter {
	if g == nil || info == nil {
		return nil
	}
	if g.nativeInterfaceRefreshAllowed() && (info.AdapterVersion != g.nativeInterfaceAdapterVersion ||
		(len(info.Adapters) == 0 && len(g.nativeInterfaceExplicitAdapters[info.Key]) == 0)) {
		g.refreshNativeInterfaceAdapters(info)
	}
	adapterMap := make(map[string]*nativeInterfaceAdapter)
	for _, adapter := range info.Adapters {
		if adapter == nil || adapter.GoType == "" {
			continue
		}
		adapterMap[nativeInterfaceAdapterIdentity(adapter)] = adapter
	}
	if extra := g.nativeInterfaceExplicitAdapters[info.Key]; extra != nil {
		for key, adapter := range extra {
			if adapter == nil || key == "" {
				continue
			}
			if _, exists := adapterMap[key]; exists {
				continue
			}
			adapterMap[key] = adapter
		}
	}
	if len(adapterMap) == 0 {
		return nil
	}
	adapters := make([]*nativeInterfaceAdapter, 0, len(adapterMap))
	for _, adapter := range adapterMap {
		adapters = append(adapters, adapter)
	}
	sort.Slice(adapters, func(i, j int) bool {
		return nativeInterfaceAdapterIdentity(adapters[i]) < nativeInterfaceAdapterIdentity(adapters[j])
	})
	return adapters
}

func (g *generator) nativeInterfaceAdapterVisibleInPackage(adapter *nativeInterfaceAdapter, pkgName string) bool {
	if adapter == nil || adapter.ImplDefinition == nil {
		return true
	}
	pkgName = strings.TrimSpace(pkgName)
	implPkg := strings.TrimSpace(adapter.ImplPackage)
	if pkgName == implPkg {
		return true
	}
	if adapter.ImplDefinition.IsPrivate {
		return false
	}
	for _, binding := range g.staticImports[pkgName] {
		if strings.TrimSpace(binding.SourcePackage) == implPkg {
			return true
		}
	}
	return false
}

func (g *generator) nativeInterfaceAdapterForActualInPackage(
	info *nativeInterfaceInfo,
	actual string,
	pkgName string,
) (*nativeInterfaceAdapter, bool) {
	if g == nil || info == nil || actual == "" {
		return nil, false
	}
	var found *nativeInterfaceAdapter
	for _, adapter := range g.nativeInterfaceKnownAdapters(info) {
		if adapter == nil || adapter.GoType != actual || !g.nativeInterfaceAdapterVisibleInPackage(adapter, pkgName) {
			continue
		}
		if found != nil && found != adapter {
			return nil, false
		}
		found = adapter
	}
	return found, found != nil
}
