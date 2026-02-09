package compiler

import (
	"fmt"
	"sort"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/typechecker"
)

// ProgramAnalysis captures compiler-visible metadata about a loaded program.
type ProgramAnalysis struct {
	Program   *driver.Program
	Graph     ModuleGraph
	Typecheck typechecker.CheckResult
}

// ModuleGraph describes static and dynamic dependencies between packages.
type ModuleGraph struct {
	Modules       map[string]*driver.Module
	StaticEdges   map[string][]string
	DynamicEdges  map[string][]string
	MissingStatic map[string][]string
	Order         []string
}

// AnalyzeProgram builds a dependency graph and preserves typechecker outputs.
func AnalyzeProgram(program *driver.Program) (*ProgramAnalysis, error) {
	if program == nil {
		return nil, fmt.Errorf("compiler: program is nil")
	}
	if program.Entry == nil || program.Entry.AST == nil {
		return nil, fmt.Errorf("compiler: program missing entry module")
	}
	graph, err := buildModuleGraph(program)
	if err != nil {
		return nil, err
	}
	check, err := typechecker.NewProgramChecker().Check(program)
	if err != nil {
		return nil, err
	}
	return &ProgramAnalysis{
		Program:   program,
		Graph:     graph,
		Typecheck: check,
	}, nil
}

func buildModuleGraph(program *driver.Program) (ModuleGraph, error) {
	modules := make(map[string]*driver.Module)
	for _, mod := range program.Modules {
		if mod == nil {
			continue
		}
		if mod.Package == "" {
			return ModuleGraph{}, fmt.Errorf("compiler: module missing package name")
		}
		if _, exists := modules[mod.Package]; exists {
			return ModuleGraph{}, fmt.Errorf("compiler: duplicate module package %q", mod.Package)
		}
		modules[mod.Package] = mod
	}

	staticEdges := make(map[string][]string, len(modules))
	dynamicEdges := make(map[string][]string, len(modules))
	missingStatic := make(map[string][]string)
	order := make([]string, 0, len(program.Modules))

	for _, mod := range program.Modules {
		if mod == nil {
			continue
		}
		order = append(order, mod.Package)
		staticEdges[mod.Package] = uniqueStrings(mod.Imports)
		dynamicEdges[mod.Package] = uniqueStrings(mod.DynImports)
		for _, dep := range mod.Imports {
			if dep == "" {
				continue
			}
			if _, ok := modules[dep]; !ok {
				missingStatic[mod.Package] = append(missingStatic[mod.Package], dep)
			}
		}
	}

	for key, deps := range staticEdges {
		staticEdges[key] = uniqueStrings(deps)
	}
	for key, deps := range dynamicEdges {
		dynamicEdges[key] = uniqueStrings(deps)
	}
	for key, deps := range missingStatic {
		missingStatic[key] = uniqueStrings(deps)
		if len(missingStatic[key]) == 0 {
			delete(missingStatic, key)
		}
	}

	return ModuleGraph{
		Modules:       modules,
		StaticEdges:   staticEdges,
		DynamicEdges:  dynamicEdges,
		MissingStatic: missingStatic,
		Order:         order,
	}, nil
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
