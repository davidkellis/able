package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

type patternIntent struct {
	declarationNames map[string]struct{}
	allowFallback    bool
}

func (c *Checker) bindPattern(env *Environment, target ast.AssignmentTarget, valueType Type, allowDefine bool, intent *patternIntent) []Diagnostic {
	if target == nil {
		return nil
	}

	switch pat := target.(type) {
	case *ast.Identifier:
		if pat == nil {
			return nil
		}
		if allowDefine {
			if intent == nil || intent.declarationNames == nil {
				env.Define(pat.Name, valueType)
			} else {
				if _, ok := intent.declarationNames[pat.Name]; ok {
					env.Define(pat.Name, valueType)
				} else {
					env.Assign(pat.Name, valueType)
				}
			}
		} else {
			assigned := env.Assign(pat.Name, valueType)
			if !assigned && intent != nil && intent.allowFallback {
				env.Define(pat.Name, valueType)
			}
		}
		c.infer.set(pat, valueType)
		return nil
	case *ast.WildcardPattern:
		c.infer.set(pat, valueType)
		return nil
	case *ast.StructPattern:
		return c.bindStructPattern(env, pat, valueType, allowDefine, intent)
	case *ast.ArrayPattern:
		return c.bindArrayPattern(env, pat, valueType, allowDefine, intent)
	case *ast.LiteralPattern:
		literalType := c.literalPatternType(pat.Literal)
		expected := literalType
		if valueType == nil || isUnknownType(valueType) {
			valueType = expected
		}
		finalType := expected
		if (finalType == nil || isUnknownType(finalType)) && valueType != nil {
			finalType = valueType
		}
		c.infer.set(pat, finalType)
		return nil
	case *ast.TypedPattern:
		var diags []Diagnostic
		var expected Type = UnknownType{}
		if pat.TypeAnnotation != nil {
			expected = c.resolveTypeReference(pat.TypeAnnotation)
		}
		innerType := valueType
		if expected != nil && !isUnknownType(expected) {
			innerType = expected
		}
		if inner, ok := pat.Pattern.(ast.AssignmentTarget); ok {
			diags = append(diags, c.bindPattern(env, inner, innerType, allowDefine, intent)...)
		} else if pat.Pattern != nil {
			if node, ok := pat.Pattern.(ast.Node); ok {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: unsupported nested pattern %T", pat.Pattern),
					Node:    node,
				})
			}
		}
		finalType := valueType
		if expected != nil && !isUnknownType(expected) {
			finalType = expected
		}
		if expected != nil && !isUnknownType(expected) && valueType != nil && !isUnknownType(valueType) {
			if msg, ok := literalMismatchMessage(valueType, expected); ok {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: %s", msg),
					Node:    pat,
				})
			}
		}
		c.infer.set(pat, finalType)
		return diags
	default:
		if node, ok := target.(ast.Node); ok {
			return []Diagnostic{{
				Message: fmt.Sprintf("typechecker: pattern %T not supported yet", pat),
				Node:    node,
			}}
		}
		return []Diagnostic{{
			Message: fmt.Sprintf("typechecker: pattern %T not supported yet", pat),
		}}
	}
}

func isErrorTypeAnnotation(t Type) bool {
	if t == nil {
		return false
	}
	if st, ok := t.(StructType); ok {
		return st.StructName == "Error"
	}
	return false
}

func (c *Checker) resolveTypeReference(expr ast.TypeExpression) Type {
	if expr == nil {
		return UnknownType{}
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return UnknownType{}
		}
		name := t.Name.Name
		if c.typeParamInScope(name) {
			return TypeParameterType{ParameterName: name}
		}
		switch name {
		case "bool":
			return PrimitiveType{Kind: PrimitiveBool}
		case "string":
			return PrimitiveType{Kind: PrimitiveString}
		case "String":
			return PrimitiveType{Kind: PrimitiveString}
		case "char":
			return PrimitiveType{Kind: PrimitiveChar}
		case "nil":
			return PrimitiveType{Kind: PrimitiveNil}
		case "_":
			return UnknownType{}
		case "i8", "i16", "i32", "i64", "i128", "isize":
			return IntegerType{Suffix: name}
		case "u8", "u16", "u32", "u64", "u128", "usize":
			return IntegerType{Suffix: name}
		case "f32", "f64":
			return FloatType{Suffix: name}
		default:
			if typ, ok := c.global.Lookup(name); ok {
				if alias, ok := typ.(AliasType); ok {
					inst, subst := instantiateAlias(alias, nil)
					c.verifyAliasConstraints(alias, subst, t)
					return inst
				}
				return typ
			}
			return StructType{StructName: name}
		}
	case *ast.GenericTypeExpression:
		if simple, ok := t.Base.(*ast.SimpleTypeExpression); ok && simple.Name != nil && !c.typeParamInScope(simple.Name.Name) {
			if typ, ok := c.global.Lookup(simple.Name.Name); ok {
				if alias, ok := typ.(AliasType); ok {
					args := make([]Type, len(t.Arguments))
					for i, arg := range t.Arguments {
						args[i] = c.resolveTypeReference(arg)
					}
					inst, subst := instantiateAlias(alias, args)
					c.verifyAliasConstraints(alias, subst, t)
					return inst
				}
			}
		}
		base := c.resolveTypeReference(t.Base)
		args := make([]Type, len(t.Arguments))
		for i, arg := range t.Arguments {
			args[i] = c.resolveTypeReference(arg)
		}
		if base == nil {
			return UnknownType{}
		}
		return AppliedType{
			Base:      base,
			Arguments: args,
		}
	case *ast.NullableTypeExpression:
		inner := c.resolveTypeReference(t.InnerType)
		return NullableType{Inner: inner}
	default:
		return UnknownType{}
	}
}

func (c *Checker) bindStructPattern(env *Environment, pat *ast.StructPattern, valueType Type, allowDefine bool, intent *patternIntent) []Diagnostic {
	var diags []Diagnostic
	var structInfo StructType
	hasInfo := false

	if isUnknownType(valueType) {
		for _, field := range pat.Fields {
			if field == nil {
				continue
			}
			bind := func(target ast.AssignmentTarget) {
				if target != nil {
					diags = append(diags, c.bindPattern(env, target, UnknownType{}, allowDefine, intent)...)
				}
			}
			bind(field.Binding)
			if inner, ok := field.Pattern.(ast.AssignmentTarget); ok {
				bind(inner)
			}
		}
		c.infer.set(pat, valueType)
		return diags
	}

	if inst, ok := valueType.(StructInstanceType); ok {
		structInfo = StructType{StructName: inst.StructName, Fields: inst.Fields, Positional: inst.Positional}
		hasInfo = true
	} else if st, ok := valueType.(StructType); ok {
		structInfo = st
		hasInfo = true
	} else if candidate, ok := c.structFromUnionVariant(valueType, pat); ok {
		structInfo = candidate
		hasInfo = true
	} else if nullable, ok := valueType.(NullableType); ok {
		inner := nullable.Inner
		if inst, ok := inner.(StructInstanceType); ok {
			structInfo = StructType{StructName: inst.StructName, Fields: inst.Fields, Positional: inst.Positional}
			hasInfo = true
		} else if st, ok := inner.(StructType); ok {
			structInfo = st
			hasInfo = true
		} else if candidate, ok := c.structFromUnionVariant(inner, pat); ok {
			structInfo = candidate
			hasInfo = true
		} else if isUnionLike(inner) {
			// allow unknown field types for union members
		} else if !isUnknownType(inner) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: struct pattern cannot match type %s", typeName(inner)),
				Node:    pat,
			})
		}
		valueType = inner
	} else if isUnionLike(valueType) {
		// Union with no struct match; permit pattern binding with unknown field types.
	} else if !isUnknownType(valueType) {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: struct pattern cannot match type %s", typeName(valueType)),
			Node:    pat,
		})
	}

	if !hasInfo && pat.StructType != nil && pat.StructType.Name != "" {
		if typ, ok := c.global.Lookup(pat.StructType.Name); ok {
			if st, ok := typ.(StructType); ok {
				structInfo = st
				hasInfo = true
			}
		}
	}

	bindField := func(expected Type, field *ast.StructPatternField) {
		if field == nil {
			return
		}
		effective := expected
		if field.TypeAnnotation != nil {
			if resolved := c.resolveTypeReference(field.TypeAnnotation); resolved != nil {
				effective = resolved
			}
		}
		if field.Binding != nil {
			diags = append(diags, c.bindPattern(env, field.Binding, effective, allowDefine, intent)...)
		}
		if inner, ok := field.Pattern.(ast.AssignmentTarget); ok {
			diags = append(diags, c.bindPattern(env, inner, effective, allowDefine, intent)...)
		}
	}

	if pat.IsPositional {
		for idx, field := range pat.Fields {
			expected := Type(UnknownType{})
			if hasInfo && idx < len(structInfo.Positional) {
				expected = structInfo.Positional[idx]
			}
			bindField(expected, field)
		}
		c.infer.set(pat, valueType)
		return diags
	}

	for _, field := range pat.Fields {
		if field == nil {
			continue
		}
		name := ""
		if field.FieldName != nil {
			name = field.FieldName.Name
		}
		fieldType := Type(UnknownType{})
		if hasInfo && name != "" {
			if t, ok := structInfo.Fields[name]; ok {
				fieldType = t
			} else {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: struct pattern field '%s' not found", name),
					Node:    field,
				})
			}
		}
		bindField(fieldType, field)
	}

	c.infer.set(pat, valueType)
	return diags
}

func (c *Checker) structFromUnionVariant(valueType Type, pat *ast.StructPattern) (StructType, bool) {
	matchStruct := func(candidate Type) (StructType, bool) {
		switch v := candidate.(type) {
		case StructType:
			if pat.StructType == nil || pat.StructType.Name == "" || v.StructName == pat.StructType.Name {
				return v, true
			}
		case StructInstanceType:
			if pat.StructType == nil || pat.StructType.Name == "" || v.StructName == pat.StructType.Name {
				return StructType{StructName: v.StructName, Fields: v.Fields, Positional: v.Positional}, true
			}
		case AliasType:
			if st, ok := v.Target.(StructType); ok {
				if pat.StructType == nil || pat.StructType.Name == "" || st.StructName == pat.StructType.Name {
					return st, true
				}
			}
		}
		return StructType{}, false
	}

	switch v := valueType.(type) {
	case UnionType:
		for _, variant := range v.Variants {
			if st, ok := matchStruct(variant); ok {
				return st, true
			}
		}
	case UnionLiteralType:
		for _, variant := range v.Members {
			if st, ok := matchStruct(variant); ok {
				return st, true
			}
		}
	case AliasType:
		if st, ok := matchStruct(v.Target); ok {
			return st, true
		}
		if ut, ok := v.Target.(UnionType); ok {
			return c.structFromUnionVariant(ut, pat)
		}
		if ul, ok := v.Target.(UnionLiteralType); ok {
			return c.structFromUnionVariant(ul, pat)
		}
	}
	return StructType{}, false
}

func isUnionLike(t Type) bool {
	switch v := t.(type) {
	case UnionType, UnionLiteralType:
		return true
	case AliasType:
		return isUnionLike(v.Target)
	default:
		return false
	}
}

func (c *Checker) bindArrayPattern(env *Environment, pat *ast.ArrayPattern, valueType Type, allowDefine bool, intent *patternIntent) []Diagnostic {
	var diags []Diagnostic
	elemType := Type(UnknownType{})
	if arr, ok := valueType.(ArrayType); ok {
		if arr.Element != nil {
			elemType = arr.Element
		}
	}
	for _, elem := range pat.Elements {
		if elem == nil {
			continue
		}
		switch node := elem.(type) {
		case ast.Pattern:
			if target, ok := node.(ast.AssignmentTarget); ok {
				diags = append(diags, c.bindPattern(env, target, elemType, allowDefine, intent)...)
			}
		default:
			if target, ok := elem.(ast.AssignmentTarget); ok {
				diags = append(diags, c.bindPattern(env, target, elemType, allowDefine, intent)...)
			}
		}
	}
	if pat.RestPattern != nil {
		if target, ok := pat.RestPattern.(ast.AssignmentTarget); ok {
			diags = append(diags, c.bindPattern(env, target, ArrayType{Element: elemType}, allowDefine, intent)...)
		}
	}
	c.infer.set(pat, valueType)
	return diags
}

func (c *Checker) literalPatternType(lit ast.Literal) Type {
	switch v := lit.(type) {
	case *ast.IntegerLiteral:
		suffix := "i32"
		if v.IntegerType != nil {
			suffix = string(*v.IntegerType)
		}
		return IntegerType{Suffix: suffix}
	case *ast.FloatLiteral:
		suffix := "f64"
		if v.FloatType != nil {
			suffix = string(*v.FloatType)
		}
		return FloatType{Suffix: suffix}
	case *ast.StringLiteral:
		return PrimitiveType{Kind: PrimitiveString}
	case *ast.BooleanLiteral:
		return PrimitiveType{Kind: PrimitiveBool}
	case *ast.NilLiteral:
		return PrimitiveType{Kind: PrimitiveNil}
	case *ast.CharLiteral:
		return PrimitiveType{Kind: PrimitiveChar}
	default:
		return UnknownType{}
	}
}
