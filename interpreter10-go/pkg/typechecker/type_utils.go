package typechecker

// typeName returns a human-readable identifier for a type, tolerating nil.
func typeName(t Type) string {
	if t == nil {
		return "unknown"
	}
	return t.Name()
}

func isUnknownType(t Type) bool {
	if t == nil {
		return true
	}
	_, ok := t.(UnknownType)
	return ok
}

func isTypeParameter(t Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.(TypeParameterType)
	return ok
}

func isIntegerType(t Type) bool {
	if t == nil {
		return false
	}
	switch val := t.(type) {
	case IntegerType:
		return true
	case PrimitiveType:
		return val.Kind == PrimitiveInt
	default:
		return false
	}
}

func isFloatType(t Type) bool {
	if t == nil {
		return false
	}
	switch v := t.(type) {
	case FloatType:
		return true
	case PrimitiveType:
		return v.Kind == PrimitiveFloat
	}
	return false
}

func isNumericType(t Type) bool {
	return isIntegerType(t) || isFloatType(t)
}

func isBoolType(t Type) bool {
	if t == nil {
		return false
	}
	if prim, ok := t.(PrimitiveType); ok {
		return prim.Kind == PrimitiveBool
	}
	return false
}

func isStringType(t Type) bool {
	if t == nil {
		return false
	}
	if prim, ok := t.(PrimitiveType); ok {
		return prim.Kind == PrimitiveString
	}
	return false
}

func isInterfaceLikeType(t Type) bool {
	if t == nil {
		return false
	}
	switch v := t.(type) {
	case InterfaceType:
		return true
	case AppliedType:
		return isInterfaceLikeType(v.Base)
	default:
		return false
	}
}

func isPrimitiveInt(t Type) bool {
	if t == nil {
		return false
	}
	if prim, ok := t.(PrimitiveType); ok {
		return prim.Kind == PrimitiveInt
	}
	return false
}

// typeAssignable performs a shallow compatibility check between two types.
// It intentionally permits Unknown/TypeParam targets so later passes can refine them.
func typeAssignable(from, to Type) bool {
	if to != nil {
		if name, ok := structName(to); ok && name == "void" {
			return true
		}
	}
	if to == nil || isUnknownType(to) {
		return true
	}
	if isTypeParameter(to) {
		// Type parameters accept any argument for now; constraint solving happens later.
		return true
	}
	if from == nil {
		return false
	}
	if isUnknownType(from) {
		return true
	}
	if isTypeParameter(from) {
		return true
	}
	switch target := to.(type) {
	case StructType:
		if name, ok := structName(from); ok {
			return name == target.StructName
		}
		return false
	case StructInstanceType:
		if name, ok := structName(from); ok {
			return name == target.StructName
		}
		return false
	case ArrayType:
		if elem, ok := arrayElementType(from); ok {
			return typeAssignable(elem, target.Element)
		}
		return false
	case RangeType:
		if rng, ok := from.(RangeType); ok {
			return typeAssignable(rng.Element, target.Element)
		}
		return false
	case NullableType:
		if nullable, ok := from.(NullableType); ok {
			return typeAssignable(nullable.Inner, target.Inner)
		}
		return typeAssignable(from, target.Inner)
	case UnionLiteralType:
		return unionAssignable(from, target)
	case AppliedType:
		if applied, ok := from.(AppliedType); ok {
			return appliedTypesAssignable(applied, target)
		}
		if name, ok := structName(from); ok {
			if base, ok := target.Base.(StructType); ok && base.StructName == name {
				return true
			}
		}
		return false
	}

	switch source := from.(type) {
	case StructType:
		if name, ok := structName(to); ok {
			return source.StructName == name
		}
	case StructInstanceType:
		if name, ok := structName(to); ok {
			return source.StructName == name
		}
	case AppliedType:
		if targetApplied, ok := to.(AppliedType); ok {
			return appliedTypesAssignable(source, targetApplied)
		}
		if name, ok := structName(to); ok {
			if base, ok := source.Base.(StructType); ok {
				return base.StructName == name
			}
		}
	case ArrayType:
		if elem, ok := arrayElementType(to); ok {
			return typeAssignable(source.Element, elem)
		}
	case RangeType:
		if rng, ok := to.(RangeType); ok {
			return typeAssignable(source.Element, rng.Element)
		}
	case NullableType:
		return typeAssignable(source.Inner, to)
	case UnionLiteralType:
		if targetUnion, ok := to.(UnionLiteralType); ok {
			return unionAssignable(source, targetUnion)
		}
		for _, member := range source.Members {
			if !typeAssignable(member, to) {
				return false
			}
		}
		return true
	}

	return from.Name() == to.Name()
}

func mergeBranchTypes(types []Type) Type {
	var result Type = UnknownType{}
	for _, t := range types {
		if t == nil || isUnknownType(t) {
			continue
		}
		if isUnknownType(result) {
			result = t
			continue
		}
		if result.Name() != t.Name() {
			return UnknownType{}
		}
	}
	return result
}

func mergeCompatibleTypes(a, b Type) Type {
	if a == nil || isUnknownType(a) {
		if b == nil {
			return UnknownType{}
		}
		return b
	}
	if b == nil || isUnknownType(b) {
		return a
	}
	if typeAssignable(b, a) {
		return a
	}
	if typeAssignable(a, b) {
		return b
	}
	return UnknownType{}
}

func mergeCompatibleTypesSlice(types ...Type) Type {
	var result Type = UnknownType{}
	for _, t := range types {
		if t == nil {
			continue
		}
		result = mergeCompatibleTypes(result, t)
		if isUnknownType(result) && t != nil && !isUnknownType(t) {
			result = t
		}
	}
	if result == nil {
		return UnknownType{}
	}
	return result
}

func mergeTypesAllowUnion(a, b Type) Type {
	if a == nil || isUnknownType(a) {
		return b
	}
	if b == nil || isUnknownType(b) {
		return a
	}
	if typeAssignable(b, a) {
		return a
	}
	if typeAssignable(a, b) {
		return b
	}
	return buildUnionType(a, b)
}

func buildUnionType(types ...Type) Type {
	var members []Type
	for _, t := range types {
		if t == nil || isUnknownType(t) {
			continue
		}
		members = appendUnionMember(members, t)
	}
	if len(members) == 0 {
		return UnknownType{}
	}
	if len(members) == 1 {
		return members[0]
	}
	return UnionLiteralType{Members: members}
}

func appendUnionMember(existing []Type, candidate Type) []Type {
	if candidate == nil {
		return existing
	}
	switch v := candidate.(type) {
	case UnionLiteralType:
		for _, member := range v.Members {
			existing = appendUnionMember(existing, member)
		}
		return existing
	default:
		for _, member := range existing {
			if sameType(member, candidate) {
				return existing
			}
		}
		return append(existing, candidate)
	}
}

func sameType(a, b Type) bool {
	if a == nil || b == nil {
		return false
	}
	if isUnknownType(a) || isUnknownType(b) {
		return false
	}
	if a.Name() == b.Name() {
		return true
	}
	switch av := a.(type) {
	case AppliedType:
		if bv, ok := b.(AppliedType); ok {
			if !sameType(av.Base, bv.Base) {
				return false
			}
			if len(av.Arguments) != len(bv.Arguments) {
				return false
			}
			for i := range av.Arguments {
				if !sameType(av.Arguments[i], bv.Arguments[i]) {
					return false
				}
			}
			return true
		}
	case ArrayType:
		if bv, ok := b.(ArrayType); ok {
			return sameType(av.Element, bv.Element)
		}
	case RangeType:
		if bv, ok := b.(RangeType); ok {
			return sameType(av.Element, bv.Element)
		}
	case NullableType:
		if bv, ok := b.(NullableType); ok {
			return sameType(av.Inner, bv.Inner)
		}
	case UnionLiteralType:
		if bv, ok := b.(UnionLiteralType); ok {
			if len(av.Members) != len(bv.Members) {
				return false
			}
			for i := range av.Members {
				if !sameType(av.Members[i], bv.Members[i]) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func iterableElementType(t Type) (Type, bool) {
	if t == nil {
		return UnknownType{}, true
	}
	if _, ok := t.(UnknownType); ok {
		return UnknownType{}, true
	}
	if elem, ok := arrayElementType(t); ok {
		if elem == nil || isUnknownType(elem) {
			return UnknownType{}, true
		}
		return elem, true
	}
	if rng, ok := t.(RangeType); ok {
		if rng.Element == nil || isUnknownType(rng.Element) {
			return UnknownType{}, true
		}
		return rng.Element, true
	}
	return UnknownType{}, false
}

func structName(t Type) (string, bool) {
	switch s := t.(type) {
	case StructType:
		return s.StructName, true
	case StructInstanceType:
		return s.StructName, true
	case AppliedType:
		if base, ok := s.Base.(StructType); ok {
			return base.StructName, true
		}
	}
	return "", false
}

func arrayElementType(t Type) (Type, bool) {
	switch arr := t.(type) {
	case ArrayType:
		return arr.Element, true
	case StructType:
		if arr.StructName == "Array" {
			if len(arr.Positional) > 0 {
				return arr.Positional[0], true
			}
			return UnknownType{}, true
		}
	case StructInstanceType:
		if arr.StructName == "Array" {
			if len(arr.Positional) > 0 {
				return arr.Positional[0], true
			}
			return UnknownType{}, true
		}
	case AppliedType:
		if name, ok := structName(arr.Base); ok && name == "Array" {
			if len(arr.Arguments) > 0 {
				return arr.Arguments[0], true
			}
			return UnknownType{}, true
		}
	}
	return nil, false
}

func appliedTypesAssignable(from, to AppliedType) bool {
	if !typeAssignable(from.Base, to.Base) {
		return false
	}
	if len(from.Arguments) != len(to.Arguments) {
		return false
	}
	for i := range from.Arguments {
		if !typeAssignable(from.Arguments[i], to.Arguments[i]) {
			return false
		}
	}
	return true
}

func unionAssignable(from Type, to UnionLiteralType) bool {
	if union, ok := from.(UnionLiteralType); ok {
		for _, member := range union.Members {
			if !typeAssignableToAny(member, to.Members) {
				return false
			}
		}
		return true
	}
	return typeAssignableToAny(from, to.Members)
}

func typeAssignableToAny(from Type, targets []Type) bool {
	for _, target := range targets {
		if typeAssignable(from, target) {
			return true
		}
	}
	return false
}

func substituteFunctionType(fn FunctionType, subst map[string]Type) FunctionType {
	if len(subst) == 0 {
		return fn
	}
	params := make([]Type, len(fn.Params))
	for i, param := range fn.Params {
		params[i] = substituteType(param, subst)
	}
	ret := substituteType(fn.Return, subst)
	where := substituteWhereSpecs(fn.Where, subst)
	obligations := substituteObligations(fn.Obligations, subst)
	typeParams := fn.TypeParams
	if len(typeParams) > 0 {
		filtered := make([]GenericParamSpec, 0, len(typeParams))
		for _, param := range typeParams {
			if param.Name == "" {
				filtered = append(filtered, param)
				continue
			}
			if _, ok := subst[param.Name]; ok {
				continue
			}
			filtered = append(filtered, param)
		}
		typeParams = filtered
	}
	return FunctionType{
		Params:      params,
		Return:      ret,
		TypeParams:  typeParams,
		Where:       where,
		Obligations: obligations,
	}
}

func substituteType(t Type, subst map[string]Type) Type {
	if t == nil {
		return nil
	}
	switch v := t.(type) {
	case TypeParameterType:
		if replacement, ok := subst[v.ParameterName]; ok {
			return replacement
		}
		return v
	case FunctionType:
		return substituteFunctionType(v, subst)
	case ArrayType:
		return ArrayType{Element: substituteType(v.Element, subst)}
	case NullableType:
		return NullableType{Inner: substituteType(v.Inner, subst)}
	case RangeType:
		return RangeType{Element: substituteType(v.Element, subst)}
	case AppliedType:
		base := substituteType(v.Base, subst)
		args := make([]Type, len(v.Arguments))
		for i, arg := range v.Arguments {
			args[i] = substituteType(arg, subst)
		}
		return AppliedType{Base: base, Arguments: args}
	case UnionLiteralType:
		members := make([]Type, len(v.Members))
		for i, member := range v.Members {
			members[i] = substituteType(member, subst)
		}
		return UnionLiteralType{Members: members}
	case ProcType:
		return ProcType{Result: substituteType(v.Result, subst)}
	case FutureType:
		return FutureType{Result: substituteType(v.Result, subst)}
	}
	return t
}

func substituteWhereSpecs(specs []WhereConstraintSpec, subst map[string]Type) []WhereConstraintSpec {
	if len(specs) == 0 || len(subst) == 0 {
		return specs
	}
	out := make([]WhereConstraintSpec, 0, len(specs))
	for _, spec := range specs {
		if spec.TypeParam != "" {
			if _, ok := subst[spec.TypeParam]; ok {
				// This where-clause references a type parameter that has been
				// substituted with a concrete type; drop the clause because the
				// obligation is now captured via the substituted constraints.
				continue
			}
		}
		constraints := make([]Type, len(spec.Constraints))
		for j, constraint := range spec.Constraints {
			constraints[j] = substituteType(constraint, subst)
		}
		out = append(out, WhereConstraintSpec{
			TypeParam:   spec.TypeParam,
			Constraints: constraints,
		})
	}
	return out
}

func substituteObligations(obligations []ConstraintObligation, subst map[string]Type) []ConstraintObligation {
	if len(obligations) == 0 || len(subst) == 0 {
		return obligations
	}
	out := make([]ConstraintObligation, len(obligations))
	for i, ob := range obligations {
		var subject Type
		if ob.Subject != nil {
			subject = substituteType(ob.Subject, subst)
		} else if replacement, ok := subst[ob.TypeParam]; ok {
			subject = replacement
		}
		out[i] = ConstraintObligation{
			Owner:      ob.Owner,
			TypeParam:  ob.TypeParam,
			Constraint: substituteType(ob.Constraint, subst),
			Subject:    subject,
			Context:    ob.Context,
			Node:       ob.Node,
		}
	}
	return out
}

func populateObligationSubjects(obligations []ConstraintObligation, subject Type) []ConstraintObligation {
	if len(obligations) == 0 || subject == nil || isUnknownType(subject) {
		return obligations
	}
	out := make([]ConstraintObligation, len(obligations))
	for i, ob := range obligations {
		if ob.Subject != nil && !isUnknownType(ob.Subject) {
			out[i] = ob
			continue
		}
		out[i] = ob
		out[i].Subject = subject
	}
	return out
}
