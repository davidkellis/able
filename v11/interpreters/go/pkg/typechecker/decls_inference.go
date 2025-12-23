package typechecker

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

var reservedTypeNames = map[string]struct{}{
	"bool":     {},
	"String":   {},
	"char":     {},
	"nil":      {},
	"void":     {},
	"Self":     {},
	"i8":       {},
	"i16":      {},
	"i32":      {},
	"i64":      {},
	"i128":     {},
	"isize":    {},
	"u8":       {},
	"u16":      {},
	"u32":      {},
	"u64":      {},
	"u128":     {},
	"usize":    {},
	"f32":      {},
	"f64":      {},
	"Array":    {},
	"Map":      {},
	"Range":    {},
	"Iterator": {},
	"Result":   {},
	"Option":   {},
	"Proc":     {},
	"Future":   {},
	"Channel":  {},
	"Mutex":    {},
	"Error":    {},
}

type typeIdentifierOccurrence struct {
	name            string
	node            ast.Node
	fromWhereClause bool
}

type inferenceSkipReason int

const (
	skipNone inferenceSkipReason = iota
	skipAlreadyKnown
	skipReserved
	skipKnownType
)

func (c *declarationCollector) ensureFunctionGenericInference(def *ast.FunctionDefinition, scope map[string]Type) {
	if def == nil {
		return
	}
	occs := collectFunctionTypeOccurrences(def)
	inferred := c.selectInferredGenericParameters(occs, def.GenericParams, scope)
	if len(inferred) == 0 {
		return
	}
	paramMap := make(map[string]*ast.GenericParameter, len(inferred))
	for _, param := range inferred {
		if param == nil || param.Name == nil {
			continue
		}
		param.IsInferred = true
		paramMap[param.Name.Name] = param
	}
	def.WhereClause = hoistWhereConstraints(def.WhereClause, paramMap)
	def.GenericParams = append(def.GenericParams, inferred...)
	def.InferredGenericParams = append(def.InferredGenericParams, inferred...)
}

func (c *declarationCollector) ensureSignatureGenericInference(sig *ast.FunctionSignature, scope map[string]Type) {
	if sig == nil {
		return
	}
	occs := collectSignatureTypeOccurrences(sig)
	inferred := c.selectInferredGenericParameters(occs, sig.GenericParams, scope)
	if len(inferred) == 0 {
		return
	}
	paramMap := make(map[string]*ast.GenericParameter, len(inferred))
	for _, param := range inferred {
		if param == nil || param.Name == nil {
			continue
		}
		param.IsInferred = true
		paramMap[param.Name.Name] = param
	}
	sig.WhereClause = hoistWhereConstraints(sig.WhereClause, paramMap)
	sig.GenericParams = append(sig.GenericParams, inferred...)
	sig.InferredGenericParams = append(sig.InferredGenericParams, inferred...)
}

func (c *declarationCollector) selectInferredGenericParameters(
	occs []typeIdentifierOccurrence,
	existing []*ast.GenericParameter,
	scope map[string]Type,
) []*ast.GenericParameter {
	if len(occs) == 0 {
		return nil
	}
	known := make(map[string]struct{})
	for name := range scope {
		if name == "" {
			continue
		}
		known[name] = struct{}{}
	}
	for _, param := range existing {
		if param == nil || param.Name == nil {
			continue
		}
		known[param.Name.Name] = struct{}{}
	}
	var inferred []*ast.GenericParameter
	reportedKnownType := make(map[string]bool)
	for _, occ := range occs {
		infer, reason := c.shouldInferGenericParameter(occ.name, known)
		if !infer {
			if reason == skipKnownType && occ.fromWhereClause && occ.name != "" && !reportedKnownType[occ.name] {
				msg := fmt.Sprintf("typechecker: cannot infer type parameter '%s' because a type with the same name exists; declare it explicitly or qualify the type", occ.name)
				c.diags = append(c.diags, Diagnostic{Message: msg, Node: occ.node})
				reportedKnownType[occ.name] = true
			}
			continue
		}
		param := newInferredGenericParameter(occ.name, occ.node)
		if param != nil && c.origins != nil && occ.node != nil {
			if origin, ok := c.origins[occ.node]; ok && origin != "" {
				c.origins[param] = origin
				if param.Name != nil {
					c.origins[param.Name] = origin
				}
			}
		}
		inferred = append(inferred, param)
		known[occ.name] = struct{}{}
	}
	return inferred
}

func (c *declarationCollector) shouldInferGenericParameter(name string, known map[string]struct{}) (bool, inferenceSkipReason) {
	if name == "" {
		return false, skipReserved
	}
	if _, exists := known[name]; exists {
		return false, skipAlreadyKnown
	}
	if strings.ContainsRune(name, '.') {
		return false, skipReserved
	}
	if _, reserved := reservedTypeNames[name]; reserved {
		return false, skipReserved
	}
	if c.env != nil {
		if decl, exists := c.env.Lookup(name); exists {
			if _, isFn := decl.(FunctionType); !isFn {
				return false, skipKnownType
			}
		}
	}
	return true, skipNone
}

func newInferredGenericParameter(name string, node ast.Node) *ast.GenericParameter {
	id := ast.NewIdentifier(name)
	if node != nil {
		ast.SetSpan(id, node.Span())
	}
	param := ast.NewGenericParameter(id, nil)
	param.IsInferred = true
	if node != nil {
		ast.SetSpan(param, node.Span())
	}
	return param
}

func (c *declarationCollector) ensureImplementationGenericInference(def *ast.ImplementationDefinition) {
	if def == nil {
		return
	}
	occs := collectImplementationTypeOccurrences(def)
	scope := map[string]Type{
		"Self": TypeParameterType{ParameterName: "Self"},
	}
	inferred := c.selectInferredGenericParameters(occs, def.GenericParams, scope)
	if len(inferred) == 0 {
		return
	}
	paramMap := make(map[string]*ast.GenericParameter, len(inferred))
	for _, param := range inferred {
		if param == nil || param.Name == nil {
			continue
		}
		paramMap[param.Name.Name] = param
	}
	def.WhereClause = hoistWhereConstraints(def.WhereClause, paramMap)
	def.GenericParams = append(def.GenericParams, inferred...)
}

func (c *declarationCollector) ensureMethodsGenericInference(def *ast.MethodsDefinition) {
	if def == nil {
		return
	}
	occs := collectMethodsTypeOccurrences(def)
	scope := map[string]Type{
		"Self": TypeParameterType{ParameterName: "Self"},
	}
	inferred := c.selectInferredGenericParameters(occs, def.GenericParams, scope)
	if len(inferred) == 0 {
		return
	}
	paramMap := make(map[string]*ast.GenericParameter, len(inferred))
	for _, param := range inferred {
		if param == nil || param.Name == nil {
			continue
		}
		paramMap[param.Name.Name] = param
	}
	def.WhereClause = hoistWhereConstraints(def.WhereClause, paramMap)
	def.GenericParams = append(def.GenericParams, inferred...)
}

func collectFunctionTypeOccurrences(def *ast.FunctionDefinition) []typeIdentifierOccurrence {
	var occs []typeIdentifierOccurrence
	if def == nil {
		return occs
	}
	for _, param := range def.Params {
		if param == nil {
			continue
		}
		collectTypeExpressionOccurrences(param.ParamType, &occs)
	}
	collectTypeExpressionOccurrences(def.ReturnType, &occs)
	for _, clause := range def.WhereClause {
		if clause == nil || clause.TypeParam == nil {
			continue
		}
		occs = append(occs, typeIdentifierOccurrence{name: clause.TypeParam.Name, node: clause.TypeParam, fromWhereClause: true})
	}
	return occs
}

func collectSignatureTypeOccurrences(sig *ast.FunctionSignature) []typeIdentifierOccurrence {
	var occs []typeIdentifierOccurrence
	if sig == nil {
		return occs
	}
	for _, param := range sig.Params {
		if param == nil {
			continue
		}
		collectTypeExpressionOccurrences(param.ParamType, &occs)
	}
	collectTypeExpressionOccurrences(sig.ReturnType, &occs)
	for _, clause := range sig.WhereClause {
		if clause == nil || clause.TypeParam == nil {
			continue
		}
		occs = append(occs, typeIdentifierOccurrence{name: clause.TypeParam.Name, node: clause.TypeParam, fromWhereClause: true})
	}
	return occs
}

func collectTypeExpressionOccurrences(expr ast.TypeExpression, occs *[]typeIdentifierOccurrence) {
	if expr == nil {
		return
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			*occs = append(*occs, typeIdentifierOccurrence{name: t.Name.Name, node: t.Name})
		}
	case *ast.GenericTypeExpression:
		collectTypeExpressionOccurrences(t.Base, occs)
		for _, arg := range t.Arguments {
			collectTypeExpressionOccurrences(arg, occs)
		}
	case *ast.FunctionTypeExpression:
		for _, param := range t.ParamTypes {
			collectTypeExpressionOccurrences(param, occs)
		}
		collectTypeExpressionOccurrences(t.ReturnType, occs)
	case *ast.NullableTypeExpression:
		collectTypeExpressionOccurrences(t.InnerType, occs)
	case *ast.ResultTypeExpression:
		collectTypeExpressionOccurrences(t.InnerType, occs)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			collectTypeExpressionOccurrences(member, occs)
		}
	}
}

func collectImplementationTypeOccurrences(def *ast.ImplementationDefinition) []typeIdentifierOccurrence {
	var occs []typeIdentifierOccurrence
	if def == nil {
		return occs
	}
	collectTypeExpressionOccurrences(def.TargetType, &occs)
	for _, arg := range def.InterfaceArgs {
		collectTypeExpressionOccurrences(arg, &occs)
	}
	collectWhereClauseOccurrences(def.WhereClause, &occs)
	return occs
}

func collectMethodsTypeOccurrences(def *ast.MethodsDefinition) []typeIdentifierOccurrence {
	var occs []typeIdentifierOccurrence
	if def == nil {
		return occs
	}
	collectTypeExpressionOccurrences(def.TargetType, &occs)
	collectWhereClauseOccurrences(def.WhereClause, &occs)
	return occs
}

func collectWhereClauseOccurrences(where []*ast.WhereClauseConstraint, occs *[]typeIdentifierOccurrence) {
	if len(where) == 0 {
		return
	}
	for _, clause := range where {
		if clause == nil {
			continue
		}
		if clause.TypeParam != nil {
			*occs = append(*occs, typeIdentifierOccurrence{
				name:            clause.TypeParam.Name,
				node:            clause.TypeParam,
				fromWhereClause: true,
			})
		}
		for _, constraint := range clause.Constraints {
			if constraint == nil || constraint.InterfaceType == nil {
				continue
			}
			collectTypeExpressionOccurrences(constraint.InterfaceType, occs)
		}
	}
}

func hoistWhereConstraints(where []*ast.WhereClauseConstraint, inferred map[string]*ast.GenericParameter) []*ast.WhereClauseConstraint {
	if len(where) == 0 || len(inferred) == 0 {
		return where
	}
	kept := make([]*ast.WhereClauseConstraint, 0, len(where))
	for _, clause := range where {
		if clause == nil || clause.TypeParam == nil {
			kept = append(kept, clause)
			continue
		}
		name := clause.TypeParam.Name
		param := inferred[name]
		if param == nil {
			kept = append(kept, clause)
			continue
		}
		if len(clause.Constraints) > 0 {
			param.Constraints = append(param.Constraints, clause.Constraints...)
		}
	}
	return kept
}
