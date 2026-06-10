package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type stringSet map[string]struct{}

type functionEffect struct {
	Name                    string              `json:"name"`
	File                    string              `json:"file"`
	Line                    int                 `json:"line"`
	Direct                  bool                `json:"direct_compiled"`
	ControlFree             bool                `json:"control_free"`
	UniversalRangeFree      bool                `json:"universal_range_control_free"`
	ClosedDirectReachable   bool                `json:"closed_direct_reachable"`
	ClosedDirectRangeFree   bool                `json:"closed_direct_range_control_free"`
	RangeClass              string              `json:"range_class,omitempty"`
	Dependencies            []string            `json:"dependencies,omitempty"`
	Hazards                 []string            `json:"hazards,omitempty"`
	CallSites               []callSite          `json:"call_sites,omitempty"`
	PrimitiveRangeBlockers  []primitiveBlocker  `json:"primitive_range_blockers,omitempty"`
	RelationalBlockers      []relationalBlocker `json:"relational_blockers,omitempty"`
	ClosedDirectParamRanges []parameterRange    `json:"closed_direct_parameter_ranges,omitempty"`
	ClosedAggregateFacts    []aggregateRange    `json:"closed_direct_aggregate_facts,omitempty"`
	ClosedReturnFacts       []aggregateRange    `json:"closed_direct_return_facts,omitempty"`

	decl              *ast.FuncDecl
	closedReturnRange intRange
	closedReturnSet   bool
	closedReturnFacts map[string]intRange
	closedParamFacts  []map[string]intRange
}

type callSite struct {
	Callee string `json:"callee"`
	File   string `json:"file"`
	Line   int    `json:"line"`
}

type reportSummary struct {
	ControlFunctions           int `json:"control_functions"`
	DirectCompiledFunctions    int `json:"direct_compiled_functions"`
	ControlFreeDirect          int `json:"control_free_direct"`
	PotentiallyFallibleDirect  int `json:"potentially_fallible_direct"`
	DirectCallSites            int `json:"direct_call_sites"`
	ControlFreeDirectCallSites int `json:"control_free_direct_call_sites"`
	UniversalRangeFreeDirect   int `json:"universal_range_control_free_direct"`
	ClosedDirectReachable      int `json:"closed_direct_reachable"`
	ClosedDirectRangeFree      int `json:"closed_direct_range_control_free"`
	CallSiteSpecializable      int `json:"call_site_specializable_direct"`
	RelationalBoundsChecks     int `json:"relational_bounds_checks"`
	ClosedRelationalBoundsSafe int `json:"closed_relational_bounds_safe"`
	ReachableRelationalChecks  int `json:"closed_reachable_relational_bounds_checks"`
	ReachableRelationalSafe    int `json:"closed_reachable_relational_bounds_safe"`
}

type censusReport struct {
	Version   int               `json:"version"`
	Directory string            `json:"directory"`
	Summary   reportSummary     `json:"summary"`
	Functions []*functionEffect `json:"functions"`
}

func main() {
	var output string
	flag.StringVar(&output, "output", "", "write JSON to this path instead of stdout")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: compiled-control-census [-output path] GENERATED_DIR")
		os.Exit(2)
	}
	report, err := analyzeDirectory(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "compiled-control-census: %v\n", err)
		os.Exit(1)
	}
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "compiled-control-census: encode report: %v\n", err)
		os.Exit(1)
	}
	payload = append(payload, '\n')
	if output == "" {
		_, err = os.Stdout.Write(payload)
	} else {
		err = os.WriteFile(output, payload, 0o644)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "compiled-control-census: write report: %v\n", err)
		os.Exit(1)
	}
}

func analyzeDirectory(directory string) (*censusReport, error) {
	abs, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("resolve generated directory: %w", err)
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, fmt.Errorf("read generated directory: %w", err)
	}
	fset := token.NewFileSet()
	functions := make(map[string]*functionEffect)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		path := filepath.Join(abs, entry.Name())
		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return nil, fmt.Errorf("parse %s: %w", path, parseErr)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Body == nil || !returnsControl(fn.Type) {
				continue
			}
			position := fset.Position(fn.Pos())
			functions[fn.Name.Name] = &functionEffect{
				Name:   fn.Name.Name,
				File:   filepath.Base(position.Filename),
				Line:   position.Line,
				Direct: isDirectCompiledName(fn.Name.Name),
				decl:   fn,
			}
		}
	}
	for _, effect := range functions {
		analyzeFunction(fset, effect, functions)
	}
	resolveFixedPoint(functions)
	analyzeRangeEffects(fset, functions)

	ordered := make([]*functionEffect, 0, len(functions))
	var summary reportSummary
	summary.ControlFunctions = len(functions)
	for _, effect := range functions {
		sort.Strings(effect.Dependencies)
		sort.Strings(effect.Hazards)
		sort.Slice(effect.CallSites, func(i, j int) bool {
			if effect.CallSites[i].File != effect.CallSites[j].File {
				return effect.CallSites[i].File < effect.CallSites[j].File
			}
			if effect.CallSites[i].Line != effect.CallSites[j].Line {
				return effect.CallSites[i].Line < effect.CallSites[j].Line
			}
			return effect.CallSites[i].Callee < effect.CallSites[j].Callee
		})
		if effect.Direct {
			summary.DirectCompiledFunctions++
			if effect.ControlFree {
				summary.ControlFreeDirect++
			} else {
				summary.PotentiallyFallibleDirect++
			}
			if effect.UniversalRangeFree {
				summary.UniversalRangeFreeDirect++
			}
			if effect.ClosedDirectReachable {
				summary.ClosedDirectReachable++
				if effect.ClosedDirectRangeFree {
					summary.ClosedDirectRangeFree++
				}
			}
			if effect.RangeClass == "call-site-specializable" {
				summary.CallSiteSpecializable++
			}
		}
		summary.RelationalBoundsChecks += len(effect.RelationalBlockers)
		for _, blocker := range effect.RelationalBlockers {
			if blocker.ClosedDirectSafe {
				summary.ClosedRelationalBoundsSafe++
			}
		}
		if effect.ClosedDirectReachable {
			summary.ReachableRelationalChecks += len(effect.RelationalBlockers)
			for _, blocker := range effect.RelationalBlockers {
				if blocker.ClosedDirectSafe {
					summary.ReachableRelationalSafe++
				}
			}
		}
		for _, site := range effect.CallSites {
			if callee := functions[site.Callee]; callee != nil && callee.Direct {
				summary.DirectCallSites++
				if callee.ControlFree {
					summary.ControlFreeDirectCallSites++
				}
			}
		}
		effect.decl = nil
		ordered = append(ordered, effect)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Name < ordered[j].Name })
	return &censusReport{Version: 2, Directory: abs, Summary: summary, Functions: ordered}, nil
}

func returnsControl(fn *ast.FuncType) bool {
	if fn == nil || fn.Results == nil || len(fn.Results.List) == 0 {
		return false
	}
	last := fn.Results.List[len(fn.Results.List)-1].Type
	star, ok := last.(*ast.StarExpr)
	if !ok {
		return false
	}
	ident, ok := star.X.(*ast.Ident)
	return ok && ident.Name == "__ableControl"
}

func isDirectCompiledName(name string) bool {
	return strings.HasPrefix(name, "__able_compiled_fn_") ||
		strings.HasPrefix(name, "__able_compiled_method_")
}

func analyzeFunction(fset *token.FileSet, effect *functionEffect, functions map[string]*functionEffect) {
	producers := make(map[string]stringSet)
	dependencies := make(stringSet)
	hazards := make(stringSet)
	ast.Inspect(effect.decl.Body, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.AssignStmt:
			observeAssignment(typed.Lhs, typed.Rhs, producers, functions)
		case *ast.DeclStmt:
			if gen, ok := typed.Decl.(*ast.GenDecl); ok {
				for _, spec := range gen.Specs {
					if values, ok := spec.(*ast.ValueSpec); ok {
						lhs := make([]ast.Expr, 0, len(values.Names))
						for _, name := range values.Names {
							lhs = append(lhs, name)
						}
						observeAssignment(lhs, values.Values, producers, functions)
					}
				}
			}
		case *ast.ReturnStmt:
			observeReturn(typed, producers, functions, dependencies, hazards)
		case *ast.CallExpr:
			name := calledName(typed.Fun)
			if functions[name] != nil {
				position := fset.Position(typed.Pos())
				effect.CallSites = append(effect.CallSites, callSite{
					Callee: name,
					File:   filepath.Base(position.Filename),
					Line:   position.Line,
				})
			}
		}
		return true
	})
	effect.Dependencies = setValues(dependencies)
	effect.Hazards = setValues(hazards)
}

func observeAssignment(lhs []ast.Expr, rhs []ast.Expr, producers map[string]stringSet, functions map[string]*functionEffect) {
	if len(lhs) == 0 || len(rhs) == 0 {
		return
	}
	if len(rhs) == 1 {
		target := identName(lhs[len(lhs)-1])
		if target == "" {
			return
		}
		call, ok := rhs[0].(*ast.CallExpr)
		if !ok {
			if source := identName(rhs[0]); source != "" {
				if source != "nil" {
					mergeProducer(producers, target, producers[source])
				}
			} else {
				mergeProducer(producers, target, stringSet{"unknown:assigned-control-expression": {}})
			}
			return
		}
		name := calledName(call.Fun)
		if name == "__able_append_control_call_frame" && len(call.Args) > 0 {
			if source := identName(call.Args[0]); source != "" {
				mergeProducer(producers, target, producers[source])
				return
			}
		}
		if functions[name] != nil {
			mergeProducer(producers, target, stringSet{name: {}})
		} else {
			mergeProducer(producers, target, stringSet{"unknown:" + name: {}})
		}
		return
	}
	if len(lhs) != len(rhs) {
		return
	}
	for idx := range lhs {
		target := identName(lhs[idx])
		source := identName(rhs[idx])
		if target != "" && source != "" {
			mergeProducer(producers, target, producers[source])
		}
	}
}

func observeReturn(stmt *ast.ReturnStmt, producers map[string]stringSet, functions map[string]*functionEffect, dependencies stringSet, hazards stringSet) {
	if len(stmt.Results) == 0 {
		hazards["named-or-empty-control-return"] = struct{}{}
		return
	}
	if len(stmt.Results) == 1 {
		observeControlExpr(stmt.Results[0], producers, functions, dependencies, hazards)
		return
	}
	observeControlExpr(stmt.Results[len(stmt.Results)-1], producers, functions, dependencies, hazards)
}

func observeControlExpr(expr ast.Expr, producers map[string]stringSet, functions map[string]*functionEffect, dependencies stringSet, hazards stringSet) {
	switch typed := expr.(type) {
	case *ast.Ident:
		if typed.Name == "nil" {
			return
		}
		sources := producers[typed.Name]
		if len(sources) == 0 {
			hazards["unresolved-control:"+typed.Name] = struct{}{}
			return
		}
		for source := range sources {
			if strings.HasPrefix(source, "unknown:") {
				hazards[source] = struct{}{}
			} else {
				dependencies[source] = struct{}{}
			}
		}
	case *ast.CallExpr:
		name := calledName(typed.Fun)
		if functions[name] != nil {
			dependencies[name] = struct{}{}
		} else {
			hazards["direct-control-call:"+name] = struct{}{}
		}
	default:
		hazards[fmt.Sprintf("control-expression:%T", expr)] = struct{}{}
	}
}

func resolveFixedPoint(functions map[string]*functionEffect) {
	free := make(map[string]bool, len(functions))
	for name, effect := range functions {
		free[name] = len(effect.Hazards) == 0
	}
	changed := true
	for changed {
		changed = false
		for name, effect := range functions {
			if !free[name] {
				continue
			}
			for _, dependency := range effect.Dependencies {
				if !free[dependency] {
					free[name] = false
					changed = true
					break
				}
			}
		}
	}
	for name, effect := range functions {
		effect.ControlFree = free[name]
	}
}

func calledName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.IndexExpr:
		return calledName(typed.X)
	case *ast.IndexListExpr:
		return calledName(typed.X)
	case *ast.SelectorExpr:
		return calledName(typed.X) + "." + typed.Sel.Name
	default:
		return ""
	}
}

func identName(expr ast.Expr) string {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name
}

func mergeProducer(all map[string]stringSet, target string, values stringSet) {
	if len(values) == 0 {
		return
	}
	if all[target] == nil {
		all[target] = make(stringSet)
	}
	for value := range values {
		all[target][value] = struct{}{}
	}
}

func setValues(values stringSet) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	return out
}
