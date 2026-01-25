package typechecker

import "able/interpreter-go/pkg/ast"

type unionMember struct {
	typ  Type
	node ast.Node
}

type unionNormalizationOptions struct {
	warnRedundant func(t Type, node ast.Node)
}

func normalizeUnionTypes(members []Type) Type {
	if len(members) == 0 {
		return UnknownType{}
	}
	entries := make([]unionMember, 0, len(members))
	for _, member := range members {
		entries = append(entries, unionMember{typ: member})
	}
	return normalizeUnionMembers(entries, unionNormalizationOptions{})
}

func normalizeUnionMembers(members []unionMember, opts unionNormalizationOptions) Type {
	var normalized []unionMember
	nilType := PrimitiveType{Kind: PrimitiveNil}
	var addMember func(entry unionMember)
	addMember = func(entry unionMember) {
		entry.typ = expandAliasForUnion(entry.typ)
		if entry.typ == nil || isUnknownType(entry.typ) {
			return
		}
		switch v := entry.typ.(type) {
		case UnionLiteralType:
			for _, member := range v.Members {
				addMember(unionMember{typ: member, node: entry.node})
			}
			return
		case NullableType:
			addMember(unionMember{typ: nilType, node: entry.node})
			addMember(unionMember{typ: v.Inner, node: entry.node})
			return
		}
		for _, existing := range normalized {
			if unionMembersEquivalent(existing.typ, entry.typ) {
				if opts.warnRedundant != nil && entry.node != nil {
					opts.warnRedundant(entry.typ, entry.node)
				}
				return
			}
		}
		normalized = append(normalized, entry)
	}
	for _, entry := range members {
		addMember(entry)
	}
	if len(normalized) == 0 {
		return UnknownType{}
	}
	if len(normalized) == 1 {
		return normalized[0].typ
	}
	if len(normalized) == 2 {
		nilIndex := -1
		for idx, entry := range normalized {
			if isNilType(expandAliasForUnion(entry.typ)) {
				nilIndex = idx
				break
			}
		}
		if nilIndex != -1 {
			other := normalized[1-nilIndex].typ
			return NullableType{Inner: other}
		}
	}
	out := make([]Type, len(normalized))
	for i, entry := range normalized {
		out[i] = entry.typ
	}
	return UnionLiteralType{Members: out}
}

func unionVariantsFromType(t Type) []Type {
	switch v := t.(type) {
	case UnionLiteralType:
		return append([]Type(nil), v.Members...)
	case NullableType:
		return []Type{PrimitiveType{Kind: PrimitiveNil}, v.Inner}
	default:
		return []Type{t}
	}
}

func unionMembersEquivalent(a, b Type) bool {
	return sameType(expandAliasForUnion(a), expandAliasForUnion(b))
}

func expandAliasForUnion(t Type) Type {
	if t == nil {
		return nil
	}
	switch v := t.(type) {
	case AliasType:
		target, _ := instantiateAlias(v, nil)
		return expandAliasForUnion(target)
	case AppliedType:
		if alias, ok := v.Base.(AliasType); ok {
			target, _ := instantiateAlias(alias, v.Arguments)
			return expandAliasForUnion(target)
		}
	}
	return t
}

func isNilType(t Type) bool {
	prim, ok := t.(PrimitiveType)
	return ok && prim.Kind == PrimitiveNil
}
