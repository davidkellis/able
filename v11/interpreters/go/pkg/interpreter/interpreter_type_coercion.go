package interpreter

import (
	"fmt"
	"math"
	"math/big"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) coerceValueToType(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, error) {
	switch t := typeExpr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			name := normalizeKernelAliasName(t.Name.Name)
			if errVal, ok := asErrorValue(value); ok {
				if errVal.Payload != nil {
					if payload, ok := errVal.Payload["value"]; ok && payload != nil {
						if structVal, ok := payload.(*runtime.StructInstanceValue); ok {
							if structVal.Definition != nil && structVal.Definition.Node != nil && structVal.Definition.Node.ID != nil {
								if structVal.Definition.Node.ID.Name == name {
									return payload, nil
								}
							}
						}
						if defVal, ok := payload.(*runtime.StructDefinitionValue); ok && isSingletonStructDef(defVal.Node) {
							if defVal.Node != nil && defVal.Node.ID != nil && defVal.Node.ID.Name == name {
								return payload, nil
							}
						}
					}
				}
			}
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
			if name == "f32" || name == "f64" {
				targetFloat := runtime.FloatType(name)
				switch val := value.(type) {
				case runtime.FloatValue:
					if val.TypeSuffix != targetFloat {
						return runtime.FloatValue{Val: normalizeFloat(targetFloat, val.Val), TypeSuffix: targetFloat}, nil
					}
				case *runtime.FloatValue:
					if val != nil && val.TypeSuffix != targetFloat {
						return runtime.FloatValue{Val: normalizeFloat(targetFloat, val.Val), TypeSuffix: targetFloat}, nil
					}
				case runtime.IntegerValue:
					if val.Val != nil {
						f := bigIntToFloat(val.Val)
						return runtime.FloatValue{Val: normalizeFloat(targetFloat, f), TypeSuffix: targetFloat}, nil
					}
				case *runtime.IntegerValue:
					if val != nil && val.Val != nil {
						f := bigIntToFloat(val.Val)
						return runtime.FloatValue{Val: normalizeFloat(targetFloat, f), TypeSuffix: targetFloat}, nil
					}
				}
			}
			if name == "Error" {
				if _, ok := value.(runtime.ErrorValue); ok {
					return value, nil
				}
				if errVal, ok := value.(*runtime.ErrorValue); ok && errVal != nil {
					return value, nil
				}
			}
			if structVal, ok := value.(*runtime.StructInstanceValue); ok {
				if structVal.Definition != nil && structVal.Definition.Node != nil && structVal.Definition.Node.ID != nil {
					if structVal.Definition.Node.ID.Name == name {
						return value, nil
					}
				}
			}
			if defVal, ok := value.(*runtime.StructDefinitionValue); ok && isSingletonStructDef(defVal.Node) {
				if defVal.Node != nil && defVal.Node.ID != nil && defVal.Node.ID.Name == name {
					return value, nil
				}
			}
			if _, ok := i.interfaces[name]; ok {
				return i.coerceToInterfaceValue(name, value)
			}
		}
	}
	return value, nil
}

func (i *Interpreter) castValueToType(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, error) {
	if typeExpr != nil {
		if expanded := expandTypeAliases(typeExpr, i.typeAliases, nil); expanded != nil {
			typeExpr = expanded
		}
	}
	rawValue := value
	switch v := value.(type) {
	case runtime.InterfaceValue:
		rawValue = v.Underlying
	case *runtime.InterfaceValue:
		if v != nil {
			rawValue = v.Underlying
		}
	}
	switch t := typeExpr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			break
		}
		name := normalizeKernelAliasName(t.Name.Name)
		if errVal, ok := asErrorValue(value); ok {
			if errVal.Payload != nil {
				if payload, ok := errVal.Payload["value"]; ok && payload != nil {
					if structVal, ok := payload.(*runtime.StructInstanceValue); ok {
						if structVal.Definition != nil && structVal.Definition.Node != nil && structVal.Definition.Node.ID != nil {
							if structVal.Definition.Node.ID.Name == name {
								return payload, nil
							}
						}
					}
					if defVal, ok := payload.(*runtime.StructDefinitionValue); ok && isSingletonStructDef(defVal.Node) {
						if defVal.Node != nil && defVal.Node.ID != nil && defVal.Node.ID.Name == name {
							return payload, nil
						}
					}
				}
			}
		}
		targetKind := runtime.IntegerType(name)
		if info, err := getIntegerInfo(targetKind); err == nil {
			switch val := rawValue.(type) {
			case runtime.IntegerValue:
				if err := ensureFitsInteger(info, val.Val); err != nil {
					return nil, err
				}
				return runtime.IntegerValue{Val: new(big.Int).Set(val.Val), TypeSuffix: targetKind}, nil
			case *runtime.IntegerValue:
				if val == nil {
					return nil, fmt.Errorf("cannot cast <nil> to %s", targetKind)
				}
				if err := ensureFitsInteger(info, val.Val); err != nil {
					return nil, err
				}
				return runtime.IntegerValue{Val: new(big.Int).Set(val.Val), TypeSuffix: targetKind}, nil
			case runtime.FloatValue:
				if math.IsNaN(val.Val) || math.IsInf(val.Val, 0) {
					return nil, fmt.Errorf("cannot cast non-finite float to %s", targetKind)
				}
				f := new(big.Float).SetFloat64(val.Val)
				intVal, _ := f.Int(nil)
				if err := ensureFitsInteger(info, intVal); err != nil {
					return nil, err
				}
				return runtime.IntegerValue{Val: intVal, TypeSuffix: targetKind}, nil
			case *runtime.FloatValue:
				if val == nil {
					return nil, fmt.Errorf("cannot cast <nil> to %s", targetKind)
				}
				if math.IsNaN(val.Val) || math.IsInf(val.Val, 0) {
					return nil, fmt.Errorf("cannot cast non-finite float to %s", targetKind)
				}
				f := new(big.Float).SetFloat64(val.Val)
				intVal, _ := f.Int(nil)
				if err := ensureFitsInteger(info, intVal); err != nil {
					return nil, err
				}
				return runtime.IntegerValue{Val: intVal, TypeSuffix: targetKind}, nil
			}
		}
		if name == "f32" || name == "f64" {
			targetFloat := runtime.FloatType(name)
			switch val := rawValue.(type) {
			case runtime.FloatValue:
				return runtime.FloatValue{Val: normalizeFloat(targetFloat, val.Val), TypeSuffix: targetFloat}, nil
			case *runtime.FloatValue:
				if val == nil {
					return nil, fmt.Errorf("cannot cast <nil> to %s", name)
				}
				return runtime.FloatValue{Val: normalizeFloat(targetFloat, val.Val), TypeSuffix: targetFloat}, nil
			case runtime.IntegerValue:
				if val.Val == nil {
					return nil, fmt.Errorf("cannot cast <nil> to %s", name)
				}
				f := bigIntToFloat(val.Val)
				return runtime.FloatValue{Val: normalizeFloat(targetFloat, f), TypeSuffix: targetFloat}, nil
			case *runtime.IntegerValue:
				if val == nil || val.Val == nil {
					return nil, fmt.Errorf("cannot cast <nil> to %s", name)
				}
				f := bigIntToFloat(val.Val)
				return runtime.FloatValue{Val: normalizeFloat(targetFloat, f), TypeSuffix: targetFloat}, nil
			}
		}
		if name == "Error" {
			switch rawValue.(type) {
			case runtime.ErrorValue, *runtime.ErrorValue:
				return rawValue, nil
			}
		}
		if _, ok := i.interfaces[name]; ok {
			return i.coerceToInterfaceValue(name, value)
		}
	}
	if i.matchesType(typeExpr, value) {
		return value, nil
	}
	typeDesc := "<unknown>"
	if info, ok := i.getTypeInfoForValue(rawValue); ok {
		typeDesc = typeInfoToString(info)
	}
	return nil, fmt.Errorf("cannot cast %s to %s", typeDesc, typeExpressionToString(typeExpr))
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
	okImpl, err := i.typeImplementsInterface(info, interfaceName, make(map[string]struct{}))
	if err != nil {
		return nil, err
	}
	if !okImpl {
		typeDesc := typeInfoToString(info)
		if typeDesc == "<unknown>" {
			typeDesc = info.name
		}
		return nil, fmt.Errorf("Type '%s' does not implement interface %s", typeDesc, interfaceName)
	}
	candidate, err := i.lookupImplEntry(info, interfaceName)
	if err != nil {
		return nil, err
	}
	var methods map[string]runtime.Value
	if candidate != nil && candidate.entry != nil {
		methods = make(map[string]runtime.Value, len(candidate.entry.methods))
		for name, fn := range candidate.entry.methods {
			methods[name] = fn
		}
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
