package typechecker

func extendImplementationSubstitution(subst map[string]Type, iface InterfaceType, args []Type) {
	if subst == nil {
		return
	}
	if len(iface.TypeParams) == 0 {
		return
	}
	for idx, param := range iface.TypeParams {
		if param.Name == "" {
			continue
		}
		var arg Type = UnknownType{}
		if idx < len(args) && args[idx] != nil {
			arg = substituteType(args[idx], subst)
		}
		subst[param.Name] = arg
	}
}

func matchMethodTarget(object Type, target Type, params []GenericParamSpec) (map[string]Type, int, bool) {
	if isStringStructType(object) && isPrimitiveStringType(target) {
		return nil, 0, true
	}
	if isStringStructType(target) && isPrimitiveStringType(object) {
		return nil, 0, true
	}
	if objPrim, ok := object.(PrimitiveType); ok {
		if targetPrim, ok := target.(PrimitiveType); ok && targetPrim.Kind == objPrim.Kind {
			return nil, 0, true
		}
	}
	if objInt, ok := object.(IntegerType); ok {
		if targetInt, ok := target.(IntegerType); ok && targetInt.Suffix == objInt.Suffix {
			return nil, 0, true
		}
	}
	if objFloat, ok := object.(FloatType); ok {
		if targetFloat, ok := target.(FloatType); ok && targetFloat.Suffix == objFloat.Suffix {
			return nil, 0, true
		}
	}
	if _, ok := target.(TypeParameterType); ok {
		subst := make(map[string]Type)
		ok, score := matchTypeArgument(object, target, subst)
		if !ok {
			return nil, 0, false
		}
		return finalizeMatchResult(subst, params, score)
	}
	if targetUnion, ok := target.(UnionLiteralType); ok {
		return matchUnionLiteralTarget(object, targetUnion, params)
	}
	objInfo, ok := structInfoFromType(object)
	if !ok {
		return nil, 0, false
	}
	targetInfo, ok := structInfoFromType(target)
	if !ok {
		return nil, 0, false
	}
	if targetInfo.name == "" || objInfo.name == "" || targetInfo.name != objInfo.name || targetInfo.isUnion != objInfo.isUnion || targetInfo.isNullable != objInfo.isNullable {
		return nil, 0, false
	}
	subst := make(map[string]Type)
	score := 0
	for idx, targetArg := range targetInfo.args {
		var objArg Type
		if idx < len(objInfo.args) {
			objArg = objInfo.args[idx]
		}
		ok, argScore := matchTypeArgument(objArg, targetArg, subst)
		if !ok {
			return nil, 0, false
		}
		score += argScore
	}
	return finalizeMatchResult(subst, params, score)
}

func isPrimitiveStringType(t Type) bool {
	prim, ok := t.(PrimitiveType)
	return ok && prim.Kind == PrimitiveString
}

func isStringStructType(t Type) bool {
	name, ok := structName(t)
	return ok && name == "String"
}

func matchTypeArgument(actual Type, pattern Type, subst map[string]Type) (bool, int) {
	if pattern == nil || isUnknownType(pattern) {
		return true, 0
	}
	if actual == nil || isUnknownType(actual) {
		return true, 0
	}
	switch p := pattern.(type) {
	case TypeParameterType:
		if p.ParameterName == "" {
			return true, 0
		}
		if existing, ok := subst[p.ParameterName]; ok {
			if !typesEquivalentForSignature(existing, actual) {
				return false, 0
			}
			return true, 0
		}
		subst[p.ParameterName] = actual
		if actual == nil || isUnknownType(actual) {
			return true, 0
		}
		return true, 1
	case NullableType:
		av, ok := actual.(NullableType)
		if !ok {
			return false, 0
		}
		return matchTypeArgument(av.Inner, p.Inner, subst)
	case AppliedType:
		av, ok := actual.(AppliedType)
		if !ok {
			return false, 0
		}
		if !nominalBasesCompatible(av.Base, p.Base) {
			return false, 0
		}
		if len(p.Arguments) != len(av.Arguments) {
			return false, 0
		}
		score := 0
		for i := range p.Arguments {
			ok, s := matchTypeArgument(av.Arguments[i], p.Arguments[i], subst)
			if !ok {
				return false, 0
			}
			score += s
		}
		return true, score
	case UnionType:
		name, ok := unionName(actual)
		if !ok || name != p.UnionName {
			return false, 0
		}
		return true, 0
	case StructType:
		name, ok := structName(actual)
		if !ok || name != p.StructName {
			return false, 0
		}
		return true, 0
	case StructInstanceType:
		name, ok := structName(actual)
		if !ok || name != p.StructName {
			return false, 0
		}
		return true, 0
	case UnionLiteralType:
		if actualUnion, ok := actual.(UnionLiteralType); ok {
			if len(actualUnion.Members) != len(p.Members) {
				return false, 0
			}
			score := 0
			for i := range p.Members {
				ok, s := matchTypeArgument(actualUnion.Members[i], p.Members[i], subst)
				if !ok {
					return false, 0
				}
				score += s
			}
			return true, score
		}
	}
	if typesEquivalentForSignature(actual, pattern) {
		return true, 0
	}
	return false, 0
}

func matchUnionLiteralTarget(object Type, target UnionLiteralType, params []GenericParamSpec) (map[string]Type, int, bool) {
	if objUnion, ok := object.(UnionLiteralType); ok {
		if len(objUnion.Members) != len(target.Members) {
			return nil, 0, false
		}
		subst := make(map[string]Type)
		score := 0
		for i := range target.Members {
			ok, s := matchTypeArgument(objUnion.Members[i], target.Members[i], subst)
			if !ok {
				return nil, 0, false
			}
			score += s
		}
		return finalizeMatchResult(subst, params, score)
	}
	for _, member := range target.Members {
		if member == nil {
			continue
		}
		subst := make(map[string]Type)
		ok, s := matchTypeArgument(object, member, subst)
		if !ok {
			continue
		}
		return finalizeMatchResult(subst, params, s)
	}
	return nil, 0, false
}

func finalizeMatchResult(subst map[string]Type, params []GenericParamSpec, score int) (map[string]Type, int, bool) {
	if len(subst) == 0 {
		subst = nil
	}
	subst = ensureTypeParams(subst, params)
	if subst == nil || len(subst) == 0 {
		return nil, score, true
	}
	return subst, score, true
}

func ensureTypeParams(subst map[string]Type, params []GenericParamSpec) map[string]Type {
	if len(params) == 0 {
		return subst
	}
	if subst == nil {
		subst = make(map[string]Type)
	}
	for _, param := range params {
		if param.Name == "" {
			continue
		}
		if _, ok := subst[param.Name]; !ok {
			subst[param.Name] = UnknownType{}
		}
	}
	if len(subst) == 0 {
		return nil
	}
	return subst
}

func nominalBasesCompatible(actual Type, pattern Type) bool {
	if pattern == nil {
		return true
	}
	if actual == nil {
		return false
	}
	switch pb := pattern.(type) {
	case StructType:
		name, ok := structName(actual)
		return ok && name == pb.StructName
	case StructInstanceType:
		name, ok := structName(actual)
		return ok && name == pb.StructName
	case UnionType:
		name, ok := unionName(actual)
		return ok && name == pb.UnionName
	case InterfaceType:
		if iface, ok := actual.(InterfaceType); ok {
			return iface.InterfaceName == pb.InterfaceName
		}
	default:
		return typesEquivalentForSignature(actual, pattern)
	}
	return false
}

type structInfo struct {
	name       string
	args       []Type
	isUnion    bool
	isNullable bool
}

func structInfoFromType(t Type) (structInfo, bool) {
	switch v := t.(type) {
	case StructType:
		return structInfo{name: v.StructName}, v.StructName != ""
	case StructInstanceType:
		return structInfo{name: v.StructName, args: v.TypeArgs}, v.StructName != ""
	case UnionType:
		return structInfo{name: v.UnionName, isUnion: true}, v.UnionName != ""
	case NullableType:
		return structInfo{name: nullableTypeLabel, args: []Type{v.Inner}, isNullable: true}, true
	case AppliedType:
		if baseName, ok := structName(v.Base); ok {
			return structInfo{name: baseName, args: v.Arguments}, true
		}
		if unionName, ok := unionName(v.Base); ok {
			return structInfo{name: unionName, args: v.Arguments, isUnion: true}, true
		}
	case ArrayType:
		return structInfo{name: "Array", args: []Type{v.Element}}, true
	case AliasType:
		return structInfoFromType(v.Target)
	}
	return structInfo{}, false
}

func cloneTypeMap(src map[string]Type) map[string]Type {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]Type, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
