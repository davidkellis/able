package compiler

import (
	"fmt"
	"sort"
)

func (g *generator) setDynamicFeatureReport(report *DynamicFeatureReport) {
	if g == nil {
		return
	}
	g.hasDynamicFeature = report != nil && report.UsesDynamic()
}

func (g *generator) ensurePackageEnvVars() {
	if g.packageEnvVars != nil && g.packageBootstrappedVars != nil {
		return
	}
	names := g.collectPackageNames()
	g.packageEnvVars = make(map[string]string, len(names))
	g.packageBootstrappedVars = make(map[string]string, len(names))
	g.packageEnvOrder = names
	for idx, name := range names {
		g.packageEnvVars[name] = fmt.Sprintf("__able_pkg_env_%d", idx)
		g.packageBootstrappedVars[name] = fmt.Sprintf("__able_pkg_bootstrapped_%d", idx)
	}
}

func (g *generator) invalidatePackageEnvVars() {
	if g == nil {
		return
	}
	g.packageEnvVars = nil
	g.packageBootstrappedVars = nil
	g.packageEnvOrder = nil
}

func (g *generator) packageEnvVar(name string) (string, bool) {
	if g == nil {
		return "", false
	}
	g.ensurePackageEnvVars()
	envVar, ok := g.packageEnvVars[name]
	return envVar, ok
}

func (g *generator) packageBootstrappedVar(name string) (string, bool) {
	if g == nil {
		return "", false
	}
	g.ensurePackageEnvVars()
	bootVar, ok := g.packageBootstrappedVars[name]
	return bootVar, ok
}

func (g *generator) collectPackageNames() []string {
	seen := make(map[string]struct{})
	var names []string
	add := func(name string) {
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	for _, name := range g.packages {
		add(name)
	}
	add(g.entryPackage)
	if len(names) == 0 {
		for pkg := range g.functions {
			add(pkg)
		}
		for pkg := range g.overloads {
			add(pkg)
		}
	}
	sort.Strings(names)
	return names
}
