package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const schemaVersion = 4

type report struct {
	Kind               string                         `json:"kind"`
	SchemaVersion      int                            `json:"schema_version"`
	Directory          string                         `json:"directory"`
	GoFiles            int                            `json:"go_files"`
	FunctionCounts     map[string]int                 `json:"function_counts"`
	DeclaredNominals   int                            `json:"declared_nominal_structs"`
	Scopes             map[string]*scopeReport        `json:"scopes"`
	NominalProofs      map[string]*nominalProof       `json:"main_direct_reachable_nominal_proofs"`
	NominalEffectLinks map[string][]nominalEffectLink `json:"nominal_effect_links,omitempty"`
	ParseErrors        []string                       `json:"parse_errors,omitempty"`
}

type scopeReport struct {
	Functions                int                                  `json:"functions"`
	RuntimeValueTypes        int                                  `json:"runtime_value_type_sites"`
	BoundaryCategories       map[string]int                       `json:"boundary_categories"`
	BoundaryCallees          map[string]int                       `json:"boundary_callees"`
	BoundaryCallers          map[string]map[string]int            `json:"boundary_callers"`
	SemanticParentBoundaries map[string]map[string]map[string]int `json:"semantic_parent_boundaries"`
	HeapNominalLiterals      map[string]int                       `json:"heap_nominal_literals"`
}

func newScopeReport() *scopeReport {
	return &scopeReport{
		BoundaryCategories:       make(map[string]int),
		BoundaryCallees:          make(map[string]int),
		BoundaryCallers:          make(map[string]map[string]int),
		SemanticParentBoundaries: make(map[string]map[string]map[string]int),
		HeapNominalLiterals:      make(map[string]int),
	}
}

func main() {
	nominalEffectsJSON := flag.String(
		"nominal-effects-json",
		"",
		"join compiler nominal-effect summaries to generated unknown call sites",
	)
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: able-generated-boundary-census GENERATED_GO_DIRECTORY")
		os.Exit(2)
	}
	dir, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		exitErr(err)
	}
	result, err := analyze(dir)
	if err != nil {
		exitErr(err)
	}
	if *nominalEffectsJSON != "" {
		result.NominalEffectLinks, err = joinNominalEffects(
			*nominalEffectsJSON,
			result.NominalProofs,
		)
		if err != nil {
			exitErr(err)
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		exitErr(err)
	}
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "able-generated-boundary-census: %v\n", err)
	os.Exit(1)
}

func analyze(dir string) (*report, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("generated Go path is not a directory")
	}

	fset := token.NewFileSet()
	files := make(map[string]*ast.File)
	err = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != dir {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return fmt.Errorf("%s: %w", path, parseErr)
		}
		files[path] = file
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, errors.New("generated Go directory contains no Go files")
	}

	nominals := collectNominalStructs(files)
	result := &report{
		Kind:             "able-generated-static-boundary-census",
		SchemaVersion:    schemaVersion,
		Directory:        dir,
		GoFiles:          len(files),
		FunctionCounts:   make(map[string]int),
		DeclaredNominals: len(nominals),
		Scopes: map[string]*scopeReport{
			"main_direct_reachable": newScopeReport(),
			"compiled_body":         newScopeReport(),
			"entry_wrapper":         newScopeReport(),
			"runtime_wrapper":       newScopeReport(),
			"support":               newScopeReport(),
		},
	}
	functions := make(map[string]*ast.FuncDecl)
	for _, file := range files {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			functions[function.Name.Name] = function
			scope := functionScope(function.Name.Name)
			result.FunctionCounts[scope]++
			scopeResult := result.Scopes[scope]
			scopeResult.Functions++
			analyzeFunction(function, scopeResult, nominals)
		}
	}
	reachable := directlyReachableCompiledFunctions(functions, "__able_compiled_fn_main")
	result.FunctionCounts["main_direct_reachable"] = len(reachable)
	for _, function := range reachable {
		result.Scopes["main_direct_reachable"].Functions++
		analyzeFunction(function, result.Scopes["main_direct_reachable"], nominals)
	}
	result.NominalProofs = analyzeNominalProofs(fset, files, functions, nominals)
	return result, nil
}

func collectNominalStructs(files map[string]*ast.File) map[string]struct{} {
	result := make(map[string]struct{})
	for _, file := range files {
		for _, declaration := range file.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}
			for _, spec := range general.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, ok := typeSpec.Type.(*ast.StructType); !ok {
					continue
				}
				name := typeSpec.Name.Name
				if isNominalStructName(name) {
					result[name] = struct{}{}
				}
			}
		}
	}
	return result
}

func isNominalStructName(name string) bool {
	if name == "" || strings.HasPrefix(name, "__able") {
		return false
	}
	switch name {
	case "Array", "String", "Runtime":
		return false
	default:
		return true
	}
}

func functionScope(name string) string {
	switch {
	case strings.HasPrefix(name, "__able_compiled_entry_"):
		return "entry_wrapper"
	case strings.HasPrefix(name, "__able_compiled_"):
		return "compiled_body"
	case strings.HasPrefix(name, "__able_wrap_"),
		strings.HasPrefix(name, "__able_function_thunk_"),
		strings.HasPrefix(name, "__able_method_thunk_"),
		strings.HasPrefix(name, "__able_public_package_"):
		return "runtime_wrapper"
	default:
		return "support"
	}
}

func analyzeFunction(function *ast.FuncDecl, result *scopeReport, nominals map[string]struct{}) {
	caller := function.Name.Name
	ast.Inspect(function.Type, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if ok && selectorName(selector) == "runtime.Value" {
			result.RuntimeValueTypes++
			recordSemanticParentBoundary(
				result,
				"runtime_value_type",
				"runtime.Value",
				caller,
			)
		}
		return true
	})
	ast.Inspect(function.Body, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.SelectorExpr:
			if selectorName(typed) == "runtime.Value" {
				result.RuntimeValueTypes++
				recordSemanticParentBoundary(
					result,
					"runtime_value_type",
					"runtime.Value",
					caller,
				)
			}
		case *ast.CallExpr:
			callee := expressionName(typed.Fun)
			category := boundaryCategory(callee)
			if category == "" {
				break
			}
			result.BoundaryCategories[category]++
			result.BoundaryCallees[callee]++
			if result.BoundaryCallers[caller] == nil {
				result.BoundaryCallers[caller] = make(map[string]int)
			}
			result.BoundaryCallers[caller][category]++
			recordSemanticParentBoundary(result, category, callee, caller)
		case *ast.UnaryExpr:
			if typed.Op != token.AND {
				break
			}
			literal, ok := typed.X.(*ast.CompositeLit)
			if !ok {
				break
			}
			name := typeName(literal.Type)
			if _, ok := nominals[name]; ok {
				result.HeapNominalLiterals[name]++
				recordSemanticParentBoundary(
					result,
					"heap_nominal_literal",
					"&"+name,
					caller,
				)
			}
		}
		return true
	})
}

func recordSemanticParentBoundary(
	result *scopeReport,
	category string,
	callee string,
	parent string,
) {
	if result.SemanticParentBoundaries[category] == nil {
		result.SemanticParentBoundaries[category] =
			make(map[string]map[string]int)
	}
	if result.SemanticParentBoundaries[category][callee] == nil {
		result.SemanticParentBoundaries[category][callee] =
			make(map[string]int)
	}
	result.SemanticParentBoundaries[category][callee][parent]++
}

func directlyReachableCompiledFunctions(
	functions map[string]*ast.FuncDecl,
	root string,
) []*ast.FuncDecl {
	pending := []string{root}
	seen := make(map[string]bool)
	var result []*ast.FuncDecl
	for len(pending) > 0 {
		name := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if seen[name] {
			continue
		}
		seen[name] = true
		function := functions[name]
		if function == nil || functionScope(name) != "compiled_body" {
			continue
		}
		result = append(result, function)
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			callee := expressionName(call.Fun)
			target := functions[callee]
			if target != nil &&
				functionScope(target.Name.Name) == "compiled_body" &&
				!seen[callee] {
				pending = append(pending, callee)
			}
			return true
		})
	}
	return result
}

func boundaryCategory(callee string) string {
	switch {
	case callee == "":
		return ""
	case callee == "__able_call_named",
		strings.HasPrefix(callee, "__able_call_value"),
		strings.HasPrefix(callee, "__able_dynamic"),
		callee == "__able_method_call_node",
		callee == "__able_member_node",
		callee == "__able_index_node":
		return "erased_or_dynamic_call"
	case strings.HasPrefix(callee, "__able_struct_") &&
		(strings.Contains(callee, "_from") || strings.Contains(callee, "_to")):
		return "struct_runtime_conversion"
	case strings.HasPrefix(callee, "__able_union_") &&
		(strings.Contains(callee, "_from_value") || strings.Contains(callee, "_to_runtime")):
		return "union_runtime_conversion"
	case strings.HasPrefix(callee, "__able_union_") &&
		(strings.Contains(callee, "_wrap_") || strings.Contains(callee, "_as_")):
		return "native_union_wrap_or_projection"
	case strings.HasPrefix(callee, "__able_iface_") &&
		(strings.Contains(callee, "_from_value") || strings.Contains(callee, "_to_runtime") ||
			strings.Contains(callee, "_apply_runtime")):
		return "interface_runtime_conversion"
	case strings.HasPrefix(callee, "__able_iface_") && strings.Contains(callee, "_wrap_"):
		return "native_interface_adapter"
	case strings.HasPrefix(callee, "__able_fn_") &&
		(strings.Contains(callee, "_from_runtime") || strings.Contains(callee, "_to_runtime")):
		return "callable_runtime_conversion"
	case strings.HasPrefix(callee, "__able_array_") &&
		(strings.Contains(callee, "_from") || strings.Contains(callee, "_to_runtime")):
		return "array_runtime_conversion"
	case callee == "__able_control_from_error",
		callee == "__able_control_from_error_with_node",
		callee == "__able_control_to_error":
		return "control_error_conversion"
	case callee == "__able_any_to_value",
		strings.HasPrefix(callee, "__able_integer_from_value"),
		strings.HasPrefix(callee, "__able_int64_from_value"),
		strings.HasPrefix(callee, "__able_nullable_error_to_value"):
		return "primitive_or_any_runtime_conversion"
	case strings.HasPrefix(callee, "bridge.To"):
		return "bridge_encode"
	case strings.HasPrefix(callee, "bridge.As"):
		return "bridge_decode"
	case strings.HasPrefix(callee, "bridge.Raise"):
		return "bridge_error"
	case strings.HasPrefix(callee, "bridge."):
		return "bridge_other"
	case strings.HasPrefix(callee, "runtime.") &&
		(strings.Contains(callee, "Value") || strings.Contains(callee, "Instance")):
		return "runtime_value_constructor"
	default:
		return ""
	}
}

func expressionName(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		left := expressionName(typed.X)
		if left == "" {
			return typed.Sel.Name
		}
		return left + "." + typed.Sel.Name
	case *ast.IndexExpr:
		return expressionName(typed.X)
	case *ast.IndexListExpr:
		return expressionName(typed.X)
	case *ast.ParenExpr:
		return expressionName(typed.X)
	default:
		return ""
	}
}

func selectorName(selector *ast.SelectorExpr) string {
	return expressionName(selector)
}

func typeName(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return typed.Sel.Name
	case *ast.IndexExpr:
		return typeName(typed.X)
	case *ast.IndexListExpr:
		return typeName(typed.X)
	case *ast.StarExpr:
		return typeName(typed.X)
	default:
		return ""
	}
}
