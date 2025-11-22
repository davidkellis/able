package interpreter

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) getTypeInfoForValue(value runtime.Value) (typeInfo, bool) {
	switch v := value.(type) {
	case *runtime.StructInstanceValue:
		return i.typeInfoFromStructInstance(v)
	case *runtime.InterfaceValue:
		return i.getTypeInfoForValue(v.Underlying)
	default:
		return typeInfo{}, false
	}
}

func (i *Interpreter) typeExpressionFromValue(value runtime.Value) ast.TypeExpression {
	switch v := value.(type) {
	case runtime.StringValue:
		return ast.Ty("string")
	case runtime.BoolValue:
		return ast.Ty("bool")
	case runtime.CharValue:
		return ast.Ty("char")
	case runtime.NilValue:
		return ast.Ty("nil")
	case runtime.IntegerValue:
		return ast.Ty(string(v.TypeSuffix))
	case *runtime.IntegerValue:
		if v == nil {
			return nil
		}
		return ast.Ty(string(v.TypeSuffix))
	case runtime.FloatValue:
		return ast.Ty(string(v.TypeSuffix))
	case *runtime.FloatValue:
		if v == nil {
			return nil
		}
		return ast.Ty(string(v.TypeSuffix))
	case *runtime.StructInstanceValue:
		if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil {
			return nil
		}
		base := ast.Ty(v.Definition.Node.ID.Name)
		if len(v.TypeArguments) > 0 {
			return ast.Gen(base, v.TypeArguments...)
		}
		return base
	case *runtime.InterfaceValue:
		if v == nil || v.Interface == nil || v.Interface.Node == nil || v.Interface.Node.ID == nil {
			return nil
		}
		return ast.Ty(v.Interface.Node.ID.Name)
	case runtime.InterfaceValue:
		return i.typeExpressionFromValue(&v)
	case *runtime.ArrayValue:
		if v == nil {
			return nil
		}
		var elemType ast.TypeExpression
		for _, el := range v.Elements {
			inferred := i.typeExpressionFromValue(el)
			if inferred == nil {
				continue
			}
			if elemType == nil {
				elemType = inferred
				continue
			}
			if !typeExpressionsEqual(elemType, inferred) {
				elemType = ast.NewWildcardTypeExpression()
				break
			}
		}
		if elemType == nil {
			elemType = ast.NewWildcardTypeExpression()
		}
		return ast.Gen(ast.Ty("Array"), elemType)
	case runtime.ErrorValue:
		if v.TypeName != nil {
			return ast.Ty(v.TypeName.Name)
		}
		return ast.Ty("Error")
	default:
		return nil
	}
}

func (i *Interpreter) lookupImplEntry(info typeInfo, interfaceName string) (*implCandidate, error) {
	matches, err := i.collectImplCandidates(info, interfaceName, "")
	if len(matches) == 0 {
		return nil, err
	}
	best, ambiguous := i.selectBestCandidate(matches)
	if ambiguous != nil {
		detail := descriptionsFromCandidates(ambiguous)
		typeDesc := typeInfoToString(info)
		if typeDesc == "<unknown>" {
			typeDesc = info.name
		}
		return nil, fmt.Errorf("Ambiguous impl for interface '%s' on type '%s' (candidates: %s)", interfaceName, typeDesc, strings.Join(detail, ", "))
	}
	if best == nil {
		return nil, nil
	}
	return best, nil
}

func (i *Interpreter) findMethod(info typeInfo, methodName string, interfaceFilter string) (*runtime.FunctionValue, error) {
	matches, err := i.collectImplCandidates(info, interfaceFilter, methodName)
	if len(matches) == 0 {
		return nil, err
	}
	methodMatches := make([]methodMatch, 0, len(matches))
	for _, cand := range matches {
		method := cand.entry.methods[methodName]
		if method == nil {
			if ifaceDef, ok := i.interfaces[cand.entry.interfaceName]; ok && ifaceDef.Node != nil {
				for _, sig := range ifaceDef.Node.Signatures {
					if sig == nil || sig.Name == nil || sig.Name.Name != methodName || sig.DefaultImpl == nil {
						continue
					}
					defaultDef := ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
					method = &runtime.FunctionValue{Declaration: defaultDef, Closure: ifaceDef.Env}
					if cand.entry.methods == nil {
						cand.entry.methods = make(map[string]*runtime.FunctionValue)
					}
					cand.entry.methods[methodName] = method
					break
				}
			}
		}
		if method == nil {
			continue
		}
		methodMatches = append(methodMatches, methodMatch{candidate: cand, method: method})
	}
	if len(methodMatches) == 0 {
		return nil, err
	}
	best, ambiguous := i.selectBestMethodCandidate(methodMatches)
	if ambiguous != nil {
		detail := descriptionsFromMethodMatches(ambiguous)
		typeDesc := typeInfoToString(info)
		if typeDesc == "<unknown>" {
			typeDesc = info.name
		}
		return nil, fmt.Errorf("Ambiguous method '%s' for type '%s' (candidates: %s)", methodName, typeDesc, strings.Join(detail, ", "))
	}
	if best == nil {
		return nil, nil
	}
	if fnDef, ok := best.method.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
		return nil, fmt.Errorf("Method '%s' on %s is private", methodName, info.name)
	}
	return best.method, nil
}

func (i *Interpreter) interfaceMatches(val *runtime.InterfaceValue, interfaceName string) bool {
	if val == nil {
		return false
	}
	if val.Interface != nil && val.Interface.Node != nil && val.Interface.Node.ID != nil {
		if val.Interface.Node.ID.Name == interfaceName {
			return true
		}
	}
	info, ok := i.getTypeInfoForValue(val.Underlying)
	if !ok {
		return false
	}
	entry, err := i.lookupImplEntry(info, interfaceName)
	return err == nil && entry != nil
}

func (i *Interpreter) selectStructMethod(inst *runtime.StructInstanceValue, methodName string) (*runtime.FunctionValue, error) {
	if inst == nil {
		return nil, nil
	}
	info, ok := i.typeInfoFromStructInstance(inst)
	if !ok {
		return nil, nil
	}
	return i.findMethod(info, methodName, "")
}

func (i *Interpreter) matchesType(typeExpr ast.TypeExpression, value runtime.Value) bool {
	switch t := typeExpr.(type) {
	case *ast.WildcardTypeExpression:
		return true
	case *ast.SimpleTypeExpression:
		name := t.Name.Name
		switch name {
		case "string":
			_, ok := value.(runtime.StringValue)
			return ok
		case "bool":
			_, ok := value.(runtime.BoolValue)
			return ok
		case "char":
			_, ok := value.(runtime.CharValue)
			return ok
		case "nil":
			_, ok := value.(runtime.NilValue)
			return ok
		case "i8", "i16", "i32", "i64", "i128", "u8", "u16", "u32", "u64", "u128":
			var iv runtime.IntegerValue
			switch val := value.(type) {
			case runtime.IntegerValue:
				iv = val
			case *runtime.IntegerValue:
				if val == nil {
					return false
				}
				iv = *val
			default:
				return false
			}
			targetKind := runtime.IntegerType(name)
			return integerRangeWithinKinds(iv.TypeSuffix, targetKind)
		case "f32", "f64":
			fv, ok := value.(runtime.FloatValue)
			if !ok {
				return false
			}
			return string(fv.TypeSuffix) == name
		case "Error":
			_, ok := value.(runtime.ErrorValue)
			return ok
		default:
			if _, ok := i.interfaces[name]; ok {
				switch v := value.(type) {
				case *runtime.InterfaceValue:
					return i.interfaceMatches(v, name)
				case runtime.InterfaceValue:
					return i.interfaceMatches(&v, name)
				default:
					info, ok := i.getTypeInfoForValue(value)
					if !ok {
						return false
					}
					if candidate, err := i.lookupImplEntry(info, name); err == nil && candidate != nil {
						return true
					}
					return false
				}
			}
			if structVal, ok := value.(*runtime.StructInstanceValue); ok {
				if structVal.Definition != nil && structVal.Definition.Node != nil && structVal.Definition.Node.ID != nil {
					return structVal.Definition.Node.ID.Name == name
				}
			}
			return false
		}
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base.Name.Name == "Array" {
			arr, ok := value.(*runtime.ArrayValue)
			if !ok {
				return false
			}
			if len(t.Arguments) == 0 {
				return true
			}
			elemType := t.Arguments[0]
			for _, el := range arr.Elements {
				if !i.matchesType(elemType, el) {
					return false
				}
			}
			return true
		}
		return true
	case *ast.FunctionTypeExpression:
		_, ok := value.(*runtime.FunctionValue)
		return ok
	case *ast.NullableTypeExpression:
		if _, ok := value.(runtime.NilValue); ok {
			return true
		}
		return i.matchesType(t.InnerType, value)
	case *ast.ResultTypeExpression:
		return i.matchesType(t.InnerType, value)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if i.matchesType(member, value) {
				return true
			}
		}
		return false
	default:
		return true
	}
}

func (i *Interpreter) coerceValueToType(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, error) {
	switch t := typeExpr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			name := t.Name.Name
			targetKind := runtime.IntegerType(name)
			if _, err := getIntegerInfo(targetKind); err == nil {
				switch val := value.(type) {
				case runtime.IntegerValue:
					if val.TypeSuffix != targetKind && integerRangeWithinKinds(val.TypeSuffix, targetKind) {
						return runtime.IntegerValue{Val: new(big.Int).Set(val.Val), TypeSuffix: targetKind}, nil
					}
				case *runtime.IntegerValue:
					if val != nil && val.TypeSuffix != targetKind && integerRangeWithinKinds(val.TypeSuffix, targetKind) {
						return runtime.IntegerValue{Val: new(big.Int).Set(val.Val), TypeSuffix: targetKind}, nil
					}
				}
			}
			if _, ok := i.interfaces[name]; ok {
				return i.coerceToInterfaceValue(name, value)
			}
		}
	}
	return value, nil
}

func (i *Interpreter) coerceToInterfaceValue(interfaceName string, value runtime.Value) (runtime.Value, error) {
	if ifaceVal, ok := value.(*runtime.InterfaceValue); ok {
		if i.interfaceMatches(ifaceVal, interfaceName) {
			return value, nil
		}
	}
	if ifaceVal, ok := value.(runtime.InterfaceValue); ok {
		if i.interfaceMatches(&ifaceVal, interfaceName) {
			return value, nil
		}
	}
	ifaceDef, ok := i.interfaces[interfaceName]
	if !ok {
		return nil, fmt.Errorf("Interface '%s' is not defined", interfaceName)
	}
	info, ok := i.getTypeInfoForValue(value)
	if !ok {
		return nil, fmt.Errorf("Value does not implement interface %s", interfaceName)
	}
	candidate, err := i.lookupImplEntry(info, interfaceName)
	if err != nil {
		return nil, err
	}
	if candidate == nil || candidate.entry == nil {
		typeDesc := typeInfoToString(info)
		if typeDesc == "<unknown>" {
			typeDesc = info.name
		}
		return nil, fmt.Errorf("Type '%s' does not implement interface %s", typeDesc, interfaceName)
	}
	methods := make(map[string]*runtime.FunctionValue, len(candidate.entry.methods))
	for name, fn := range candidate.entry.methods {
		methods[name] = fn
	}
	return &runtime.InterfaceValue{Interface: ifaceDef, Underlying: value, Methods: methods}, nil
}

func rangeEndpoint(val runtime.Value) (int, error) {
	switch v := val.(type) {
	case runtime.IntegerValue:
		return int(v.Val.Int64()), nil
	case runtime.FloatValue:
		if math.IsNaN(v.Val) || math.IsInf(v.Val, 0) {
			return 0, fmt.Errorf("Range endpoint must be finite")
		}
		return int(math.Trunc(v.Val)), nil
	default:
		return 0, fmt.Errorf("range endpoint must be numeric, got %s", val.Kind())
	}
}
