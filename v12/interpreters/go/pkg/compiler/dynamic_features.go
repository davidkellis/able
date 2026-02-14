package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

// DynamicFeatureUsage summarizes dynamic usage within a module.
type DynamicFeatureUsage struct {
	HasDynImports        bool
	HasDynImportWildcard bool
	HasDynamicCalls      bool
}

func (u DynamicFeatureUsage) UsesDynamic() bool {
	return u.HasDynImports || u.HasDynImportWildcard || u.HasDynamicCalls
}

// DynamicFunctionUsage summarizes dynamic usage within a function definition.
type DynamicFunctionUsage struct {
	Package     string
	Name        string
	Definition  *ast.FunctionDefinition
	UsesDynamic bool
}

// DynamicFeatureReport aggregates dynamic usage for a program.
type DynamicFeatureReport struct {
	Modules     map[string]DynamicFeatureUsage
	Functions   []DynamicFunctionUsage
	DynBindings map[string]map[string]struct{}
}

func (r *DynamicFeatureReport) UsesDynamic() bool {
	if r == nil {
		return false
	}
	for _, usage := range r.Modules {
		if usage.UsesDynamic() {
			return true
		}
	}
	return false
}

// DetectDynamicFeatures scans a program for dynamic feature usage.
func DetectDynamicFeatures(program *driver.Program) (*DynamicFeatureReport, error) {
	if program == nil {
		return nil, fmt.Errorf("compiler: program is nil")
	}
	if program.Entry == nil || program.Entry.AST == nil {
		return nil, fmt.Errorf("compiler: program missing entry module")
	}
	report := &DynamicFeatureReport{
		Modules:     make(map[string]DynamicFeatureUsage),
		Functions:   make([]DynamicFunctionUsage, 0),
		DynBindings: make(map[string]map[string]struct{}),
	}

	for _, mod := range program.Modules {
		if mod == nil || mod.AST == nil {
			continue
		}
		bindings, hasWildcard := collectDynBindings(mod.AST)
		report.DynBindings[mod.Package] = bindings

		usage := DynamicFeatureUsage{
			HasDynImports:        len(mod.DynImports) > 0,
			HasDynImportWildcard: hasWildcard,
		}
		moduleDynamic := false
		for _, stmt := range mod.AST.Body {
			if fn, ok := stmt.(*ast.FunctionDefinition); ok {
				fnUsesDynamic := hasWildcard || functionUsesDynamic(fn, bindings)
				if fnUsesDynamic {
					moduleDynamic = true
				}
				report.Functions = append(report.Functions, DynamicFunctionUsage{
					Package:     mod.Package,
					Name:        identifierName(fn.ID),
					Definition:  fn,
					UsesDynamic: fnUsesDynamic,
				})
				continue
			}
			if statementUsesDynamic(stmt, bindings) {
				moduleDynamic = true
			}
		}
		usage.HasDynamicCalls = moduleDynamic
		report.Modules[mod.Package] = usage
	}

	return report, nil
}

func appendDynamicFeatureWarnings(gen *generator, report *DynamicFeatureReport) {
	if gen == nil || report == nil {
		return
	}
	modules := make([]string, 0)
	for name, usage := range report.Modules {
		if usage.UsesDynamic() {
			modules = append(modules, name)
		}
	}
	if len(modules) == 0 {
		return
	}
	sort.Strings(modules)
	for _, name := range modules {
		usage := report.Modules[name]
		parts := make([]string, 0, 3)
		if usage.HasDynImports {
			parts = append(parts, "dynimport")
		}
		if usage.HasDynImportWildcard {
			parts = append(parts, "dynimport *")
		}
		if usage.HasDynamicCalls {
			parts = append(parts, "dynamic calls")
		}
		details := strings.Join(parts, ", ")
		if details == "" {
			details = "dynamic features"
		}
		gen.warnings = append(gen.warnings, fmt.Sprintf("compiler: module %s uses %s; compiled output may rely on interpreter execution", name, details))
	}
}

func collectDynBindings(module *ast.Module) (map[string]struct{}, bool) {
	bindings := make(map[string]struct{})
	hasWildcard := false
	if module == nil {
		return bindings, false
	}
	for _, stmt := range module.Body {
		if stmt == nil {
			continue
		}
		if dyn, ok := stmt.(*ast.DynImportStatement); ok {
			if dyn.IsWildcard {
				hasWildcard = true
			}
			if dyn.Alias != nil && dyn.Alias.Name != "" {
				bindings[dyn.Alias.Name] = struct{}{}
			}
			for _, sel := range dyn.Selectors {
				if sel == nil {
					continue
				}
				if sel.Alias != nil && sel.Alias.Name != "" {
					bindings[sel.Alias.Name] = struct{}{}
					continue
				}
				if sel.Name != nil && sel.Name.Name != "" {
					bindings[sel.Name.Name] = struct{}{}
				}
			}
		}
	}
	return bindings, hasWildcard
}

func functionUsesDynamic(def *ast.FunctionDefinition, bindings map[string]struct{}) bool {
	if def == nil || def.Body == nil {
		return false
	}
	for _, stmt := range def.Body.Body {
		if statementUsesDynamic(stmt, bindings) {
			return true
		}
	}
	return false
}

func statementUsesDynamic(stmt ast.Statement, bindings map[string]struct{}) bool {
	if stmt == nil {
		return false
	}
	switch s := stmt.(type) {
	case *ast.DynImportStatement:
		return true
	case *ast.ReturnStatement:
		return exprUsesDynamic(s.Argument, bindings)
	case *ast.FunctionDefinition:
		return functionUsesDynamic(s, bindings)
	case *ast.ImplementationDefinition:
		return definitionsUseDynamic(s.Definitions, bindings)
	case *ast.MethodsDefinition:
		return definitionsUseDynamic(s.Definitions, bindings)
	case *ast.InterfaceDefinition:
		for _, sig := range s.Signatures {
			if sig == nil || sig.DefaultImpl == nil {
				continue
			}
			if blockUsesDynamic(sig.DefaultImpl, bindings) {
				return true
			}
		}
		return false
	case *ast.WhileLoop:
		return exprUsesDynamic(s.Condition, bindings) || blockUsesDynamic(s.Body, bindings)
	case *ast.ForLoop:
		return exprUsesDynamic(s.Iterable, bindings) || blockUsesDynamic(s.Body, bindings)
	case *ast.BreakStatement:
		return exprUsesDynamic(s.Value, bindings)
	case *ast.RaiseStatement:
		return exprUsesDynamic(s.Expression, bindings)
	case *ast.YieldStatement:
		return exprUsesDynamic(s.Expression, bindings)
	default:
		if expr, ok := stmt.(ast.Expression); ok {
			return exprUsesDynamic(expr, bindings)
		}
	}
	return false
}

func blockUsesDynamic(block *ast.BlockExpression, bindings map[string]struct{}) bool {
	if block == nil {
		return false
	}
	for _, stmt := range block.Body {
		if statementUsesDynamic(stmt, bindings) {
			return true
		}
	}
	return false
}

func exprUsesDynamic(expr ast.Expression, bindings map[string]struct{}) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		if e != nil && e.Name != "" {
			if _, ok := bindings[e.Name]; ok {
				return true
			}
		}
		return false
	case *ast.FunctionCall:
		if calleeIsDynamic(e.Callee, bindings) {
			return true
		}
		if exprUsesDynamic(e.Callee, bindings) {
			return true
		}
		for _, arg := range e.Arguments {
			if exprUsesDynamic(arg, bindings) {
				return true
			}
		}
		return false
	case *ast.MemberAccessExpression:
		if memberIsDynamic(e, bindings) {
			return true
		}
		if exprUsesDynamic(e.Object, bindings) {
			return true
		}
		if memberExpr, ok := e.Member.(ast.Expression); ok && exprUsesDynamic(memberExpr, bindings) {
			return true
		}
		return false
	case *ast.BlockExpression:
		return blockUsesDynamic(e, bindings)
	case *ast.ArrayLiteral:
		for _, element := range e.Elements {
			if exprUsesDynamic(element, bindings) {
				return true
			}
		}
		return false
	case *ast.StructLiteral:
		for _, field := range e.Fields {
			if field != nil && exprUsesDynamic(field.Value, bindings) {
				return true
			}
		}
		for _, src := range e.FunctionalUpdateSources {
			if exprUsesDynamic(src, bindings) {
				return true
			}
		}
		return false
	case *ast.MapLiteral:
		for _, element := range e.Elements {
			switch elem := element.(type) {
			case *ast.MapLiteralEntry:
				if exprUsesDynamic(elem.Key, bindings) || exprUsesDynamic(elem.Value, bindings) {
					return true
				}
			case *ast.MapLiteralSpread:
				if exprUsesDynamic(elem.Expression, bindings) {
					return true
				}
			}
		}
		return false
	case *ast.IndexExpression:
		return exprUsesDynamic(e.Object, bindings) || exprUsesDynamic(e.Index, bindings)
	case *ast.UnaryExpression:
		return exprUsesDynamic(e.Operand, bindings)
	case *ast.BinaryExpression:
		return exprUsesDynamic(e.Left, bindings) || exprUsesDynamic(e.Right, bindings)
	case *ast.RangeExpression:
		return exprUsesDynamic(e.Start, bindings) || exprUsesDynamic(e.End, bindings)
	case *ast.TypeCastExpression:
		return exprUsesDynamic(e.Expression, bindings)
	case *ast.AssignmentExpression:
		return exprUsesDynamic(e.Right, bindings) || assignmentTargetUsesDynamic(e.Left, bindings)
	case *ast.IfExpression:
		if exprUsesDynamic(e.IfCondition, bindings) || exprUsesDynamic(e.IfBody, bindings) {
			return true
		}
		for _, clause := range e.ElseIfClauses {
			if clause == nil {
				continue
			}
			if exprUsesDynamic(clause.Condition, bindings) || exprUsesDynamic(clause.Body, bindings) {
				return true
			}
		}
		return e.ElseBody != nil && exprUsesDynamic(e.ElseBody, bindings)
	case *ast.MatchExpression:
		if exprUsesDynamic(e.Subject, bindings) {
			return true
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil && exprUsesDynamic(clause.Guard, bindings) {
				return true
			}
			if clause.Body != nil && exprUsesDynamic(clause.Body, bindings) {
				return true
			}
		}
		return false
	case *ast.RescueExpression:
		if exprUsesDynamic(e.MonitoredExpression, bindings) {
			return true
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil && exprUsesDynamic(clause.Guard, bindings) {
				return true
			}
			if clause.Body != nil && exprUsesDynamic(clause.Body, bindings) {
				return true
			}
		}
		return false
	case *ast.EnsureExpression:
		return exprUsesDynamic(e.TryExpression, bindings) || blockUsesDynamic(e.EnsureBlock, bindings)
	case *ast.LambdaExpression:
		return exprUsesDynamic(e.Body, bindings)
	case *ast.SpawnExpression:
		return exprUsesDynamic(e.Expression, bindings)
	case *ast.AwaitExpression:
		return exprUsesDynamic(e.Expression, bindings)
	case *ast.PropagationExpression:
		return exprUsesDynamic(e.Expression, bindings)
	case *ast.OrElseExpression:
		return exprUsesDynamic(e.Expression, bindings) || blockUsesDynamic(e.Handler, bindings)
	case *ast.BreakpointExpression:
		return blockUsesDynamic(e.Body, bindings)
	case *ast.LoopExpression:
		return blockUsesDynamic(e.Body, bindings)
	case *ast.StringInterpolation:
		for _, part := range e.Parts {
			if exprUsesDynamic(part, bindings) {
				return true
			}
		}
		return false
	case *ast.IteratorLiteral:
		for _, stmt := range e.Body {
			if statementUsesDynamic(stmt, bindings) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func assignmentTargetUsesDynamic(target ast.AssignmentTarget, bindings map[string]struct{}) bool {
	if target == nil {
		return false
	}
	switch t := target.(type) {
	case *ast.MemberAccessExpression:
		return exprUsesDynamic(t.Object, bindings) || memberIsDynamic(t, bindings)
	case *ast.IndexExpression:
		return exprUsesDynamic(t.Object, bindings) || exprUsesDynamic(t.Index, bindings)
	case *ast.TypedPattern:
		return patternUsesDynamic(t.Pattern, bindings)
	case *ast.Identifier:
		return false
	case *ast.WildcardPattern:
		return false
	case *ast.LiteralPattern:
		return false
	case *ast.StructPattern:
		for _, field := range t.Fields {
			if field == nil {
				continue
			}
			if patternUsesDynamic(field.Pattern, bindings) {
				return true
			}
		}
		return false
	case *ast.ArrayPattern:
		for _, elem := range t.Elements {
			if patternUsesDynamic(elem, bindings) {
				return true
			}
		}
		if t.RestPattern != nil {
			return patternUsesDynamic(t.RestPattern, bindings)
		}
		return false
	default:
		if expr, ok := target.(ast.Expression); ok {
			return exprUsesDynamic(expr, bindings)
		}
	}
	return false
}

func patternUsesDynamic(pattern ast.Pattern, bindings map[string]struct{}) bool {
	if pattern == nil {
		return false
	}
	if target, ok := pattern.(ast.AssignmentTarget); ok {
		return assignmentTargetUsesDynamic(target, bindings)
	}
	return false
}

func definitionsUseDynamic(definitions []*ast.FunctionDefinition, bindings map[string]struct{}) bool {
	for _, fn := range definitions {
		if fn == nil {
			continue
		}
		if functionUsesDynamic(fn, bindings) {
			return true
		}
	}
	return false
}

func calleeIsDynamic(expr ast.Expression, bindings map[string]struct{}) bool {
	switch e := expr.(type) {
	case *ast.Identifier:
		if e == nil {
			return false
		}
		_, ok := bindings[e.Name]
		return ok
	case *ast.MemberAccessExpression:
		return memberIsDynamic(e, bindings)
	default:
		return false
	}
}

func memberIsDynamic(expr *ast.MemberAccessExpression, bindings map[string]struct{}) bool {
	if expr == nil {
		return false
	}
	if ident, ok := expr.Object.(*ast.Identifier); ok && ident != nil && ident.Name == "dyn" {
		return true
	}
	if memberIdent, ok := expr.Member.(*ast.Identifier); ok && memberIdent != nil {
		switch memberIdent.Name {
		case "def_package", "package", "eval", "def", "as_interface", "call", "construct":
			return true
		}
	}
	return false
}

func identifierName(id *ast.Identifier) string {
	if id == nil {
		return ""
	}
	return id.Name
}
