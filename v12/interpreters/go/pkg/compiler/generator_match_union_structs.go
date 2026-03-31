package compiler

import "able/interpreter-go/pkg/ast"

func effectiveStructPatternPositional(pattern *ast.StructPattern, info *structInfo) bool {
	if pattern == nil {
		return false
	}
	if pattern.StructType != nil && pattern.StructType.Name != "" {
		return pattern.IsPositional
	}
	if info != nil {
		switch info.Kind {
		case ast.StructKindPositional:
			return true
		case ast.StructKindNamed:
			return false
		}
	}
	return pattern.IsPositional
}

func (g *generator) structPatternMatchesInfo(pattern *ast.StructPattern, info *structInfo) bool {
	if g == nil || pattern == nil || info == nil {
		return false
	}
	if pattern.StructType != nil && pattern.StructType.Name != "" && info.Name != pattern.StructType.Name {
		return false
	}
	if effectiveStructPatternPositional(pattern, info) {
		if info.Kind != ast.StructKindPositional {
			return false
		}
		if len(pattern.Fields) != len(info.Fields) {
			return false
		}
		for _, field := range pattern.Fields {
			if _, ok := positionalStructFieldPattern(field); !ok {
				return false
			}
		}
		return true
	}
	if info.Kind == ast.StructKindPositional {
		return false
	}
	for _, field := range pattern.Fields {
		if _, ok := positionalStructFieldPattern(field); !ok {
			return false
		}
		if field.FieldName == nil || field.FieldName.Name == "" {
			return false
		}
		if g.fieldInfo(info, field.FieldName.Name) == nil {
			return false
		}
	}
	return true
}

func (g *generator) nativeUnionStructPatternMember(union *nativeUnionInfo, pattern *ast.StructPattern) (*nativeUnionMember, string, bool) {
	if g == nil || union == nil || pattern == nil {
		return nil, "", false
	}
	if pattern.StructType != nil && pattern.StructType.Name != "" {
		return nil, "", false
	}
	var matched *nativeUnionMember
	for _, member := range union.Members {
		if member == nil {
			continue
		}
		info := g.structInfoByGoName(member.GoType)
		if !g.structPatternMatchesInfo(pattern, info) {
			continue
		}
		if matched != nil {
			return nil, "", false
		}
		matched = member
	}
	if matched == nil {
		return nil, "", false
	}
	return matched, matched.GoType, true
}
