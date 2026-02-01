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
				var ifaceArgs []ast.TypeExpression
				if info, ok := parseTypeExpression(typeExpr); ok && len(info.typeArgs) > 0 {
					ifaceArgs = info.typeArgs
				}
				return i.coerceToInterfaceValue(name, value, ifaceArgs)
			}
		}
	case *ast.GenericTypeExpression:
		if info, ok := parseTypeExpression(typeExpr); ok && info.name != "" {
			name := normalizeKernelAliasName(info.name)
			if _, ok := i.interfaces[name]; ok {
				return i.coerceToInterfaceValue(name, value, info.typeArgs)
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
				wrapped := patternToInteger(bitPattern(val.Val, info), info)
				return runtime.IntegerValue{Val: new(big.Int).Set(wrapped), TypeSuffix: targetKind}, nil
			case *runtime.IntegerValue:
				if val == nil {
					return nil, fmt.Errorf("cannot cast <nil> to %s", targetKind)
				}
				wrapped := patternToInteger(bitPattern(val.Val, info), info)
				return runtime.IntegerValue{Val: new(big.Int).Set(wrapped), TypeSuffix: targetKind}, nil
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
			var ifaceArgs []ast.TypeExpression
			if info, ok := parseTypeExpression(typeExpr); ok && len(info.typeArgs) > 0 {
				ifaceArgs = info.typeArgs
			}
			return i.coerceToInterfaceValue(name, value, ifaceArgs)
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

func (i *Interpreter) CastValueToType(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, error) {
	return i.castValueToType(typeExpr, value)
}

func (i *Interpreter) interfaceDispatchSets(interfaceName string) (map[string]struct{}, map[string]struct{}) {
	base := make(map[string]struct{})
	impls := make(map[string]struct{})
	addExpanded := func(name string, target map[string]struct{}) {
		if name == "" {
			return
		}
		for _, ifaceName := range i.interfaceSearchNames(name, make(map[string]struct{})) {
			if ifaceName != "" {
				target[ifaceName] = struct{}{}
			}
		}
	}
	addExpanded(interfaceName, base)
	if entries, ok := i.implMethods[interfaceName]; ok {
		for _, entry := range entries {
			addExpanded(entry.interfaceName, impls)
		}
	}
	for name := range base {
		delete(impls, name)
	}
	return base, impls
}

func (i *Interpreter) buildInterfaceMethodDictionary(interfaceName string, ifaceArgs []ast.TypeExpression, info typeInfo) (map[string]runtime.Value, error) {
	methods := map[string]runtime.Value{}
	typeArgs := ifaceArgs
	if len(typeArgs) == 0 && len(info.typeArgs) > 0 {
		typeArgs = info.typeArgs
	}
	baseInterfaceArgs := make(map[string][]ast.TypeExpression)
	if ifaceDef, ok := i.interfaces[interfaceName]; ok && ifaceDef != nil && ifaceDef.Node != nil {
		for _, base := range ifaceDef.Node.BaseInterfaces {
			baseInfo, ok := parseTypeExpression(base)
			if !ok || baseInfo.name == "" {
				continue
			}
			if len(baseInfo.typeArgs) > 0 {
				baseInterfaceArgs[baseInfo.name] = baseInfo.typeArgs
			}
		}
	}
	base, impls := i.interfaceDispatchSets(interfaceName)
	collect := func(ifaceName string, targetInfo typeInfo, ifaceArgs []ast.TypeExpression) error {
		ifaceDef, ok := i.interfaces[ifaceName]
		if !ok || ifaceDef == nil || ifaceDef.Node == nil {
			return nil
		}
		for _, sig := range ifaceDef.Node.Signatures {
			if sig == nil || sig.Name == nil {
				continue
			}
			methodName := sig.Name.Name
			if methodName == "" || methods[methodName] != nil {
				continue
			}
			method, err := i.findMethod(targetInfo, methodName, ifaceName, ifaceArgs)
			if err != nil {
				return err
			}
			if method == nil {
				if ifaceName == interfaceName {
					return fmt.Errorf("No method '%s' for interface %s", methodName, interfaceName)
				}
				continue
			}
			methods[methodName] = method
		}
		return nil
	}
	for ifaceName := range base {
		args := ifaceArgs
		if ifaceName != interfaceName {
			args = baseInterfaceArgs[ifaceName]
		}
		if err := collect(ifaceName, info, args); err != nil {
			return nil, err
		}
	}
	for ifaceName := range impls {
		targetInfo := typeInfo{name: interfaceName, typeArgs: typeArgs}
		if err := collect(ifaceName, targetInfo, nil); err != nil {
			return nil, err
		}
	}
	if len(methods) == 0 {
		return nil, nil
	}
	return methods, nil
}

func (i *Interpreter) coerceToInterfaceValue(interfaceName string, value runtime.Value, ifaceArgs []ast.TypeExpression) (runtime.Value, error) {
	if ifaceVal, ok := value.(*runtime.InterfaceValue); ok {
		if i.interfaceMatches(ifaceVal, interfaceName, ifaceArgs) {
			return value, nil
		}
	}
	if ifaceVal, ok := value.(runtime.InterfaceValue); ok {
		if i.interfaceMatches(&ifaceVal, interfaceName, ifaceArgs) {
			return value, nil
		}
	}
	ifaceDef, ok := i.interfaces[interfaceName]
	if !ok {
		return nil, fmt.Errorf("Interface '%s' is not defined", interfaceName)
	}
	if interfaceName == "Iterator" {
		switch value.(type) {
		case *runtime.IteratorValue:
			methods, err := i.iteratorInterfaceMethodDictionary(ifaceDef)
			if err != nil {
				return nil, err
			}
			return &runtime.InterfaceValue{
				Interface:     ifaceDef,
				Underlying:    value,
				Methods:       methods,
				InterfaceArgs: ifaceArgs,
			}, nil
		}
	}
	info, ok := i.getTypeInfoForValue(value)
	if !ok {
		return nil, fmt.Errorf("Value does not implement interface %s", interfaceName)
	}
	okImpl, err := i.typeImplementsInterface(info, interfaceName, ifaceArgs, make(map[string]struct{}))
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
	if _, err := i.lookupImplEntry(info, interfaceName, ifaceArgs); err != nil {
		return nil, err
	}
	methods, err := i.buildInterfaceMethodDictionary(interfaceName, ifaceArgs, info)
	if err != nil {
		return nil, err
	}
	return &runtime.InterfaceValue{
		Interface:     ifaceDef,
		Underlying:    value,
		Methods:       methods,
		InterfaceArgs: ifaceArgs,
	}, nil
}

func (i *Interpreter) iteratorInterfaceMethodDictionary(ifaceDef *runtime.InterfaceDefinitionValue) (map[string]runtime.Value, error) {
	if ifaceDef == nil || ifaceDef.Node == nil || ifaceDef.Node.ID == nil {
		return nil, fmt.Errorf("Iterator interface is not defined")
	}
	methods := make(map[string]runtime.Value)
	for _, sig := range ifaceDef.Node.Signatures {
		if sig == nil || sig.Name == nil {
			continue
		}
		name := sig.Name.Name
		if name == "" || methods[name] != nil {
			continue
		}
		if name == "next" {
			methods[name] = iteratorNextNativeMethod()
			continue
		}
		if sig.DefaultImpl != nil {
			defaultDef := ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
			defaultVal := &runtime.FunctionValue{Declaration: defaultDef, Closure: ifaceDef.Env, MethodPriority: -1}
			if program, err := i.lowerFunctionDefinitionBytecode(defaultDef); err != nil {
				if i.execMode == execModeBytecode {
					return nil, err
				}
			} else {
				defaultVal.Bytecode = program
			}
			methods[name] = defaultVal
			continue
		}
		return nil, fmt.Errorf("No method '%s' for interface %s", name, ifaceDef.Node.ID.Name)
	}
	if len(methods) == 0 {
		return nil, nil
	}
	return methods, nil
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
