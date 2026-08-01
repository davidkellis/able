package typechecker

import (
	"able/interpreter-go/pkg/ast"
)

// PatternCoverageFact records positive control-flow facts proven by the
// checker. Absence from the map means lowering must preserve its dynamic
// fallback.
type PatternCoverageFact struct {
	Exhaustive bool
}

// PatternCoverageMap keeps semantic facts outside the canonical AST.
type PatternCoverageMap map[ast.Node]PatternCoverageFact

// Clone returns an independent fact map.
func (m PatternCoverageMap) Clone() PatternCoverageMap {
	if m == nil {
		return nil
	}
	out := make(PatternCoverageMap, len(m))
	for node, fact := range m {
		out[node] = fact
	}
	return out
}

func (c *Checker) recordPatternCoverage(env *Environment, node ast.Node, subject Type, clauses []*ast.MatchClause) {
	if c == nil || node == nil {
		return
	}
	for _, clause := range clauses {
		if clause != nil && clause.Guard == nil && c.patternUniversallyIrrefutable(env, clause.Pattern) {
			c.patternCoverage[node] = PatternCoverageFact{Exhaustive: true}
			return
		}
	}

	components, known := c.patternCoverageComponents(subject)
	if !known || len(components) == 0 {
		return
	}
	covered := make([]bool, len(components))
	for _, clause := range clauses {
		if clause == nil || clause.Guard != nil || clause.Pattern == nil {
			continue
		}
		for idx, component := range components {
			if covered[idx] {
				continue
			}
			if c.patternIrrefutableForType(env, clause.Pattern, component) {
				covered[idx] = true
			}
		}
	}
	for _, componentCovered := range covered {
		if !componentCovered {
			return
		}
	}
	c.patternCoverage[node] = PatternCoverageFact{Exhaustive: true}
}

func (c *Checker) patternUniversallyIrrefutable(env *Environment, pattern ast.Pattern) bool {
	switch pat := pattern.(type) {
	case *ast.WildcardPattern:
		return pat != nil
	case *ast.Identifier:
		return pat != nil && !c.identifierPatternIsSingleton(env, pat.Name)
	default:
		return false
	}
}

func (c *Checker) patternIrrefutableForType(env *Environment, pattern ast.Pattern, subject Type) bool {
	if pattern == nil || subject == nil || isUnknownType(subject) {
		return false
	}
	switch pat := pattern.(type) {
	case *ast.WildcardPattern:
		return pat != nil
	case *ast.Identifier:
		if pat == nil {
			return false
		}
		if !c.identifierPatternIsSingleton(env, pat.Name) {
			return true
		}
		name, ok := structName(unwrapPatternCoverageAlias(subject))
		return ok && name == pat.Name
	case *ast.TypedPattern:
		if pat == nil || pat.Pattern == nil || pat.TypeAnnotation == nil {
			return false
		}
		annotation := c.resolveTypeReference(pat.TypeAnnotation)
		if annotation == nil || isUnknownType(annotation) || !typeAssignable(subject, annotation) {
			return false
		}
		return c.patternIrrefutableForType(env, pat.Pattern, subject)
	case *ast.StructPattern:
		return c.structPatternIrrefutableForType(env, pat, subject)
	case *ast.LiteralPattern:
		if pat == nil {
			return false
		}
		_, nilPattern := pat.Literal.(*ast.NilLiteral)
		primitive, nilSubject := normalizeSpecialType(unwrapPatternCoverageAlias(subject)).(PrimitiveType)
		return nilPattern && nilSubject && primitive.Kind == PrimitiveNil
	default:
		return false
	}
}

func (c *Checker) identifierPatternIsSingleton(env *Environment, name string) bool {
	if c == nil || name == "" || name == "_" {
		return false
	}
	for scope := env; scope != nil && scope != c.global; scope = scope.parent {
		if _, ok := scope.symbols[name]; ok {
			return false
		}
	}
	if c.global == nil {
		return false
	}
	typ, ok := c.global.Lookup(name)
	if !ok || typ == nil {
		return false
	}
	nominal, ok := structName(unwrapPatternCoverageAlias(typ))
	return ok && nominal == name
}

func (c *Checker) structPatternIrrefutableForType(env *Environment, pattern *ast.StructPattern, subject Type) bool {
	if pattern == nil {
		return false
	}
	info, ok := patternCoverageStruct(subject)
	if !ok {
		return false
	}
	if pattern.StructType != nil && pattern.StructType.Name != "" && pattern.StructType.Name != info.StructName {
		return false
	}

	positional := pattern.IsPositional || len(info.Positional) > 0
	if positional {
		if len(pattern.Fields) != len(info.Positional) {
			return false
		}
		for idx, field := range pattern.Fields {
			if !c.structFieldPatternIrrefutable(env, field, info.Positional[idx]) {
				return false
			}
		}
		return true
	}

	for _, field := range pattern.Fields {
		if field == nil || field.FieldName == nil || field.FieldName.Name == "" {
			return false
		}
		fieldType, ok := info.Fields[field.FieldName.Name]
		if !ok || !c.structFieldPatternIrrefutable(env, field, fieldType) {
			return false
		}
	}
	return true
}

func (c *Checker) structFieldPatternIrrefutable(env *Environment, field *ast.StructPatternField, fieldType Type) bool {
	if field == nil || fieldType == nil || isUnknownType(fieldType) {
		return false
	}
	if field.TypeAnnotation != nil {
		annotation := c.resolveTypeReference(field.TypeAnnotation)
		if annotation == nil || isUnknownType(annotation) || !typeAssignable(fieldType, annotation) {
			return false
		}
	}
	if field.Pattern != nil && !c.patternIrrefutableForType(env, field.Pattern, fieldType) {
		return false
	}
	return true
}

func patternCoverageStruct(typ Type) (StructType, bool) {
	switch value := normalizeSpecialType(unwrapPatternCoverageAlias(typ)).(type) {
	case StructType:
		return value, true
	case StructInstanceType:
		return StructType{
			StructName: value.StructName,
			Fields:     value.Fields,
			Positional: value.Positional,
		}, true
	case AppliedType:
		if base, ok := unwrapPatternCoverageAlias(value.Base).(StructType); ok {
			return base, true
		}
	}
	return StructType{}, false
}

func (c *Checker) patternCoverageComponents(typ Type) ([]Type, bool) {
	if typ == nil || isUnknownType(typ) {
		return nil, false
	}
	typ = normalizeSpecialType(unwrapPatternCoverageAlias(typ))
	var components []Type
	switch value := typ.(type) {
	case UnionLiteralType:
		for _, member := range value.Members {
			nested, ok := c.patternCoverageComponents(member)
			if !ok {
				return nil, false
			}
			components = appendPatternCoverageComponents(components, nested...)
		}
	case UnionType:
		for _, member := range value.Variants {
			nested, ok := c.patternCoverageComponents(member)
			if !ok {
				return nil, false
			}
			components = appendPatternCoverageComponents(components, nested...)
		}
	case NullableType:
		components = append(components, PrimitiveType{Kind: PrimitiveNil})
		nested, ok := c.patternCoverageComponents(value.Inner)
		if !ok {
			return nil, false
		}
		components = appendPatternCoverageComponents(components, nested...)
	case AppliedType:
		// `!T` is a language-level Result boundary even when a loader has not
		// imported the canonical Result union declaration into this module.
		// Preserve that syntax contract without general named-nominal rules.
		if name, ok := structName(value.Base); ok && name == "Result" && len(value.Arguments) == 1 {
			success, successKnown := c.patternCoverageComponents(value.Arguments[0])
			if !successKnown {
				return nil, false
			}
			failure, failureKnown := c.patternCoverageComponents(c.lookupErrorType())
			if !failureKnown {
				return nil, false
			}
			components = appendPatternCoverageComponents(components, success...)
			components = appendPatternCoverageComponents(components, failure...)
		} else {
			components = append(components, typ)
		}
	default:
		components = append(components, typ)
	}
	return components, len(components) > 0
}

func appendPatternCoverageComponents(existing []Type, additions ...Type) []Type {
	for _, addition := range additions {
		if addition == nil || isUnknownType(addition) {
			continue
		}
		duplicate := false
		for _, current := range existing {
			if sameType(current, addition) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			existing = append(existing, addition)
		}
	}
	return existing
}

func unwrapPatternCoverageAlias(typ Type) Type {
	for depth := 0; depth < 32; depth++ {
		alias, ok := typ.(AliasType)
		if !ok || alias.Target == nil {
			return typ
		}
		typ = alias.Target
	}
	return UnknownType{}
}
