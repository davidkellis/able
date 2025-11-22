package typechecker

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
	case UnionType:
		params := make([]GenericParamSpec, len(v.TypeParams))
		for i, param := range v.TypeParams {
			constraints := make([]Type, len(param.Constraints))
			for j, constraint := range param.Constraints {
				constraints[j] = substituteType(constraint, subst)
			}
			params[i] = GenericParamSpec{
				Name:        param.Name,
				Constraints: constraints,
			}
		}
		where := substituteWhereSpecs(v.Where, subst)
		variants := make([]Type, len(v.Variants))
		for i, variant := range v.Variants {
			variants[i] = substituteType(variant, subst)
		}
		return UnionType{
			UnionName:  v.UnionName,
			TypeParams: params,
			Where:      where,
			Variants:   variants,
		}
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

func instantiateAlias(alias AliasType, args []Type) (Type, map[string]Type) {
	subst := make(map[string]Type, len(alias.TypeParams))
	for idx, param := range alias.TypeParams {
		if param.Name == "" {
			continue
		}
		if idx < len(args) && args[idx] != nil {
			subst[param.Name] = args[idx]
			continue
		}
		if _, exists := subst[param.Name]; !exists {
			subst[param.Name] = UnknownType{}
		}
	}
	target := alias.Target
	if len(subst) > 0 {
		target = substituteType(target, subst)
	}
	if target == nil {
		target = UnknownType{}
	}
	return target, subst
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
