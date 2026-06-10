package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) matchesType(typeExpr ast.TypeExpression, value runtime.Value) bool {
	if typeExpr != nil {
		if expanded := i.expandTypeAliasesCached(typeExpr); expanded != nil {
			typeExpr = expanded
		}
	}
	if matched, ok := i.matchesTypeWithoutRuntimeValue(typeExpr); ok {
		return matched
	}
	if matched, ok := i.matchesTypeWithRawPrimitiveValue(typeExpr, value); ok {
		return matched
	}
	value = bytecodeMaterializeRawValue(bytecodeSlotReadValue(value))
	if matched, ok := fastExactNamedStructTypeMatch(i, typeExpr, value); ok {
		return matched
	}
	switch t := typeExpr.(type) {
	case *ast.WildcardTypeExpression:
		return true
	case *ast.SimpleTypeExpression:
		name := normalizeKernelAliasName(t.Name.Name)
		if name == "Error" {
			return i.matchesErrorValue(value)
		}
		if name == "Self" {
			return true
		}
		if errVal, ok := asErrorValue(value); ok {
			if errVal.TypeName != nil && errVal.TypeName.Name == name {
				return true
			}
			if errVal.Payload != nil {
				if payload, ok := errVal.Payload["value"]; ok && payload != nil {
					if structVal, ok := payload.(*runtime.StructInstanceValue); ok {
						if structVal.Definition != nil && structVal.Definition.Node != nil && structVal.Definition.Node.ID != nil {
							if structVal.Definition.Node.ID.Name == name {
								return true
							}
						}
					}
					if defVal, ok := payload.(*runtime.StructDefinitionValue); ok && isSingletonStructDef(defVal.Node) {
						if defVal.Node != nil && defVal.Node.ID != nil && defVal.Node.ID.Name == name {
							return true
						}
					}
				}
			}
		}
		switch name {
		case "IoHandle", "ProcHandle":
			switch hv := value.(type) {
			case *runtime.HostHandleValue:
				if hv == nil {
					return false
				}
				return hv.HandleType == name
			default:
				return false
			}
		case "String":
			if _, ok := value.(runtime.StringValue); ok {
				return true
			}
			if structVal, ok := value.(*runtime.StructInstanceValue); ok {
				if structVal.Definition != nil && structVal.Definition.Node != nil && structVal.Definition.Node.ID != nil {
					return structVal.Definition.Node.ID.Name == "String"
				}
			}
			return false
		case "bool":
			_, ok := value.(runtime.BoolValue)
			return ok
		case "char":
			_, ok := value.(runtime.CharValue)
			return ok
		case "nil":
			_, ok := value.(runtime.NilValue)
			return ok
		case "void":
			switch value.(type) {
			case runtime.VoidValue, *runtime.VoidValue:
				return true
			default:
				return false
			}
		case "IteratorEnd":
			return i.isIteratorEnd(value)
		case "Iterator", "Future":
			switch v := value.(type) {
			case *runtime.IteratorValue, *runtime.FutureValue:
				return true
			case *runtime.InterfaceValue:
				return i.interfaceMatches(v, name, nil)
			case runtime.InterfaceValue:
				return i.interfaceMatches(&v, name, nil)
			default:
				info, ok := i.getTypeInfoForValue(value)
				if !ok {
					return false
				}
				okImpl, err := i.typeImplementsInterface(info, name, nil, make(map[interfaceImplCacheKey]struct{}))
				return err == nil && okImpl
			}
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
			if iv.TypeSuffix == targetKind || integerRangeWithinKinds(iv.TypeSuffix, targetKind) {
				return true
			}
			if iv.IsSmall() {
				return smallIntWithinRange(iv.Int64Fast(), targetKind)
			}
			if iv.Val != nil && integerValueWithinRange(iv.Val, targetKind) {
				return true
			}
			return false
		case "f32", "f64":
			switch val := value.(type) {
			case runtime.FloatValue:
				return true
			case *runtime.FloatValue:
				return val != nil
			case runtime.IntegerValue:
				return true
			case *runtime.IntegerValue:
				return val != nil
			default:
				return false
			}
		default:
			if unionDef, ok := i.unionDefinitions[name]; ok && unionDef != nil && unionDef.Node != nil {
				for _, variant := range unionDef.Node.Variants {
					if variant != nil && i.matchesType(variant, value) {
						return true
					}
				}
				return false
			}
			if defVal, ok := value.(*runtime.StructDefinitionValue); ok && isSingletonStructDef(defVal.Node) {
				if defVal.Node != nil && defVal.Node.ID != nil {
					if defVal.Node.ID.Name == name {
						return true
					}
				}
			}
			if defVal, ok := value.(runtime.StructDefinitionValue); ok && isSingletonStructDef(defVal.Node) {
				if defVal.Node != nil && defVal.Node.ID != nil {
					if defVal.Node.ID.Name == name {
						return true
					}
				}
			}
			if structVal, ok := value.(*runtime.StructInstanceValue); ok {
				if structVal.Definition != nil && structVal.Definition.Node != nil && structVal.Definition.Node.ID != nil {
					if structVal.Definition.Node.ID.Name == name {
						return true
					}
				}
			}
			if _, ok := i.interfaces[name]; ok {
				switch v := value.(type) {
				case *runtime.InterfaceValue:
					return i.interfaceMatches(v, name, nil)
				case runtime.InterfaceValue:
					return i.interfaceMatches(&v, name, nil)
				default:
					if inst, ok := value.(*runtime.StructInstanceValue); ok && inst != nil && (inst.Definition == nil || inst.Definition.Node == nil) {
						if ifaceDef, ok := i.interfaces[name]; ok && ifaceDef != nil {
							if i.structImplementsInterfaceByFields(inst, ifaceDef) {
								return true
							}
						}
					}
					info, ok := i.getTypeInfoForValue(value)
					if !ok {
						if i.compiledImplChecker != nil {
							return false
						}
						return false
					}
					okImpl, err := i.typeImplementsInterface(info, name, nil, make(map[interfaceImplCacheKey]struct{}))
					if err == nil && okImpl {
						return true
					}
					if i.compiledImplChecker != nil {
						return i.compiledImplChecker(info.name, name)
					}
					return false
				}
			}
			if i.isKnownTypeName(name) {
				return false
			}
			return true
		}
	case *ast.GenericTypeExpression:
		var baseName string
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base.Name != nil {
			baseName = normalizeKernelAliasName(base.Name.Name)
		}
		if baseName == "Self" || (len(baseName) == 1 && baseName[0] >= 'A' && baseName[0] <= 'Z') {
			return true
		}
		if baseName == "Iterator" || baseName == "Future" {
			switch v := value.(type) {
			case *runtime.IteratorValue, *runtime.FutureValue:
				return true
			case *runtime.InterfaceValue:
				return i.interfaceMatches(v, baseName, t.Arguments)
			case runtime.InterfaceValue:
				return i.interfaceMatches(&v, baseName, t.Arguments)
			default:
				info, ok := i.getTypeInfoForValue(value)
				if !ok {
					return false
				}
				okImpl, err := i.typeImplementsInterface(info, baseName, t.Arguments, make(map[interfaceImplCacheKey]struct{}))
				return err == nil && okImpl
			}
		}
		if baseName == "Array" {
			var arr *runtime.ArrayValue
			switch v := value.(type) {
			case *runtime.ArrayValue:
				arr = v
			case *runtime.StructInstanceValue:
				if v == nil || v.Definition == nil || v.Definition.Node == nil || v.Definition.Node.ID == nil || v.Definition.Node.ID.Name != "Array" {
					return false
				}
				converted, err := i.toArrayValue(v)
				if err != nil {
					return false
				}
				arr = converted
			default:
				return false
			}
			if len(t.Arguments) == 0 {
				return true
			}
			elemType := t.Arguments[0]
			if matched, decided := i.matchesMonoArrayElementTypeWithoutMaterializing(arr, elemType); decided {
				return matched
			}
			values := arr.Elements
			if len(values) == 0 {
				values = i.arrayValuesForTypeInspection(arr)
			}
			for _, el := range values {
				if !i.matchesType(elemType, el) {
					return false
				}
			}
			return true
		}
		info, ok := i.getTypeInfoForValue(value)
		if !ok {
			return true
		}
		if baseName == "" {
			if simple, ok := t.Base.(*ast.SimpleTypeExpression); ok && simple.Name != nil {
				baseName = normalizeKernelAliasName(simple.Name.Name)
			}
		}
		if baseName != "" {
			if unionDef, ok := i.unionDefinitions[baseName]; ok && unionDef != nil && unionDef.Node != nil {
				for _, variant := range unionDef.Node.Variants {
					if variant == nil {
						continue
					}
					if i.matchesType(variant, value) {
						return true
					}
				}
				return false
			}
		}
		if baseName == "Result" && len(t.Arguments) > 0 {
			if i.matchesType(t.Arguments[0], value) {
				return true
			}
			return i.matchesErrorValue(value)
		}
		if baseName == "Option" && len(t.Arguments) > 0 {
			if _, ok := value.(runtime.NilValue); ok {
				return true
			}
			return i.matchesType(t.Arguments[0], value)
		}
		if baseName != "" {
			if _, ok := i.interfaces[baseName]; ok {
				switch v := value.(type) {
				case *runtime.InterfaceValue:
					return i.interfaceMatches(v, baseName, t.Arguments)
				case runtime.InterfaceValue:
					return i.interfaceMatches(&v, baseName, t.Arguments)
				default:
					if inst, ok := value.(*runtime.StructInstanceValue); ok && inst != nil && (inst.Definition == nil || inst.Definition.Node == nil) {
						if ifaceDef, ok := i.interfaces[baseName]; ok && ifaceDef != nil {
							if i.structImplementsInterfaceByFields(inst, ifaceDef) {
								return true
							}
						}
					}
					info, ok := i.getTypeInfoForValue(value)
					if !ok {
						return false
					}
					okImpl, err := i.typeImplementsInterface(info, baseName, t.Arguments, make(map[interfaceImplCacheKey]struct{}))
					if err == nil && okImpl {
						return true
					}
					if i.compiledImplChecker != nil {
						return i.compiledImplChecker(info.name, baseName)
					}
					return false
				}
			}
			// Compiled no-bootstrap fallback for generic interfaces not in i.interfaces
			if i.compiledImplChecker != nil {
				if valInfo, ok := i.getTypeInfoForValue(value); ok {
					if i.compiledImplChecker(valInfo.name, baseName) {
						return true
					}
				}
			}
		}
		if baseName != "" && info.name != "" && baseName != info.name {
			return false
		}
		if len(t.Arguments) > 0 {
			if len(info.typeArgs) == 0 {
				return true
			}
			if len(t.Arguments) != len(info.typeArgs) {
				return false
			}
			for idx, arg := range t.Arguments {
				actual := info.typeArgs[idx]
				if arg == nil || actual == nil {
					continue
				}
				if _, ok := actual.(*ast.WildcardTypeExpression); ok {
					continue
				}
				if simple, ok := actual.(*ast.SimpleTypeExpression); ok && simple.Name != nil {
					name := simple.Name.Name
					if !i.isKnownTypeName(name) && !isPrimitiveName(name) {
						continue
					}
				}
				if simple, ok := arg.(*ast.SimpleTypeExpression); ok && simple.Name != nil {
					name := simple.Name.Name
					if !i.isKnownTypeName(name) && !isPrimitiveName(name) {
						continue
					}
				}
				if _, ok := arg.(*ast.WildcardTypeExpression); ok {
					continue
				}
				if !typeExpressionsEqual(arg, actual) {
					return false
				}
			}
		}
		return true
	case *ast.FunctionTypeExpression:
		return isCallableValue(value)
	case *ast.NullableTypeExpression:
		if _, ok := value.(runtime.NilValue); ok {
			return true
		}
		return i.matchesType(t.InnerType, value)
	case *ast.ResultTypeExpression:
		if i.matchesType(t.InnerType, value) {
			return true
		}
		return i.matchesErrorValue(value)
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

func (i *Interpreter) matchesTypeWithoutRuntimeValue(typeExpr ast.TypeExpression) (bool, bool) {
	switch t := typeExpr.(type) {
	case *ast.WildcardTypeExpression:
		return true, true
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return false, false
		}
		name := normalizeKernelAliasName(t.Name.Name)
		if name == "Self" {
			return true, true
		}
		if isPrimitiveName(name) {
			return false, false
		}
		switch name {
		case "Error", "IoHandle", "ProcHandle", "String", "bool", "char", "nil", "void",
			"IteratorEnd", "Iterator", "Future":
			return false, false
		}
		if _, ok := i.unionDefinitions[name]; ok {
			return false, false
		}
		if _, ok := i.interfaces[name]; ok {
			return false, false
		}
		if i.isKnownTypeName(name) {
			return false, false
		}
		return true, true
	case *ast.GenericTypeExpression:
		if t == nil {
			return false, false
		}
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil {
			baseName := normalizeKernelAliasName(base.Name.Name)
			if baseName == "Self" || (len(baseName) == 1 && baseName[0] >= 'A' && baseName[0] <= 'Z') {
				return true, true
			}
		}
	}
	return false, false
}

func (i *Interpreter) matchesTypeWithRawPrimitiveValue(typeExpr ast.TypeExpression, value runtime.Value) (bool, bool) {
	if bytecodeIsRawIntegerCarrier(value) {
		if kind, raw, ok := bytecodeRawIntegerValueInfo(value); ok {
			return i.matchesRawIntegerType(typeExpr, kind, raw)
		}
	}
	if _, _, ok := bytecodeDirectRawFloatValue(value); ok {
		return matchesRawFloatType(typeExpr)
	}
	return false, false
}

func (i *Interpreter) matchesRawIntegerType(typeExpr ast.TypeExpression, sourceKind runtime.IntegerType, raw int64) (bool, bool) {
	simple, ok := typeExpr.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return false, false
	}
	name := normalizeKernelAliasName(simple.Name.Name)
	targetKind := runtime.IntegerType(name)
	if _, ok := lookupIntegerInfo(targetKind); ok {
		if sourceKind == targetKind || integerRangeWithinKinds(sourceKind, targetKind) {
			return true, true
		}
		sourceInfo, sourceKnown := lookupIntegerInfo(sourceKind)
		if sourceKnown && !sourceInfo.signed && raw < 0 {
			return false, true
		}
		return smallIntWithinRange(raw, targetKind), true
	}
	switch name {
	case "f32", "f64":
		return true, true
	case "String", "bool", "char", "nil", "void", "IteratorEnd", "Iterator", "Future":
		return false, true
	}
	if i != nil {
		if _, ok := i.unionDefinitions[name]; ok {
			return false, false
		}
		if _, ok := i.interfaces[name]; ok {
			return false, false
		}
		if i.isKnownTypeName(name) {
			return false, true
		}
	}
	return false, false
}

func matchesRawFloatType(typeExpr ast.TypeExpression) (bool, bool) {
	simple, ok := typeExpr.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return false, false
	}
	switch normalizeKernelAliasName(simple.Name.Name) {
	case "f32", "f64":
		return true, true
	case "i8", "i16", "i32", "i64", "i128", "u8", "u16", "u32", "u64", "u128",
		"String", "bool", "char", "nil", "void", "IteratorEnd", "Iterator", "Future":
		return false, true
	default:
		return false, false
	}
}

// MatchesType exposes runtime type matching for compiler helpers.
func (i *Interpreter) MatchesType(typeExpr ast.TypeExpression, value runtime.Value) bool {
	return i.matchesType(typeExpr, value)
}

func (i *Interpreter) isKnownTypeName(name string) bool {
	if name == "" {
		return false
	}
	if i != nil {
		if known, ok := i.lookupKnownTypeNameCache(name); ok {
			return known
		}
	}
	known := i.scanKnownTypeName(name)
	if i != nil {
		i.storeKnownTypeNameCache(name, known)
	}
	return known
}

func (i *Interpreter) scanKnownTypeName(name string) bool {
	if i == nil {
		return false
	}
	for _, pkg := range i.packageRegistry {
		if val, ok := pkg[name]; ok {
			if packageRegistrySymbolIsKnownType(val) {
				return true
			}
		}
	}
	return false
}

func (i *Interpreter) lookupKnownTypeNameCache(name string) (bool, bool) {
	if i == nil || name == "" {
		return false, false
	}
	if i.envSingleThread {
		known, ok := i.knownTypeNameCache[name]
		return known, ok
	}
	i.typeInfoCacheMu.RLock()
	defer i.typeInfoCacheMu.RUnlock()
	known, ok := i.knownTypeNameCache[name]
	return known, ok
}

func (i *Interpreter) storeKnownTypeNameCache(name string, known bool) {
	if i == nil || name == "" {
		return
	}
	if i.envSingleThread {
		if i.knownTypeNameCache == nil {
			i.knownTypeNameCache = make(map[string]bool)
		}
		i.knownTypeNameCache[name] = known
		return
	}
	i.typeInfoCacheMu.Lock()
	defer i.typeInfoCacheMu.Unlock()
	if i.knownTypeNameCache == nil {
		i.knownTypeNameCache = make(map[string]bool)
	}
	i.knownTypeNameCache[name] = known
}

func (i *Interpreter) updateKnownTypeNameCacheForPackageSymbol(name string, val runtime.Value) {
	if packageRegistrySymbolIsKnownType(val) {
		i.storeKnownTypeNameCache(name, true)
		return
	}
	if _, ok := i.lookupKnownTypeNameCache(name); ok {
		i.storeKnownTypeNameCache(name, i.scanKnownTypeName(name))
	}
}

func packageRegistrySymbolIsKnownType(val runtime.Value) bool {
	switch val.(type) {
	case *runtime.StructDefinitionValue,
		runtime.StructDefinitionValue,
		runtime.UnionDefinitionValue,
		*runtime.UnionDefinitionValue:
		return true
	default:
		return false
	}
}

func (i *Interpreter) matchesMonoArrayElementTypeWithoutMaterializing(arr *runtime.ArrayValue, elemType ast.TypeExpression) (bool, bool) {
	if arr == nil || elemType == nil {
		return false, false
	}
	handle, sourceTypeName, ok := monoArrayHandleAndElementTypeName(arr)
	if !ok {
		return false, false
	}
	size, err := runtime.ArrayStoreSize(handle)
	if err != nil {
		return false, false
	}
	if size == 0 {
		return true, true
	}
	if expanded := i.expandTypeAliasesCached(elemType); expanded != nil {
		elemType = expanded
	}
	switch t := elemType.(type) {
	case *ast.WildcardTypeExpression:
		return true, true
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return false, false
		}
		return i.matchesMonoArraySimpleElementType(sourceTypeName, normalizeKernelAliasName(t.Name.Name))
	default:
		return false, false
	}
}

func monoArrayHandleAndElementTypeName(arr *runtime.ArrayValue) (int64, string, bool) {
	if arr == nil {
		return 0, "", false
	}
	for _, handle := range []int64{arr.Handle, arr.TrackedHandle} {
		if handle == 0 {
			continue
		}
		typeName, ok, err := runtime.ArrayStoreMonoElementTypeNameIfKnown(handle)
		if err == nil && ok {
			return handle, typeName, true
		}
	}
	return 0, "", false
}

func (i *Interpreter) matchesMonoArraySimpleElementType(sourceTypeName string, targetName string) (bool, bool) {
	if targetName == "" {
		return false, false
	}
	if _, ok := i.interfaces[targetName]; ok {
		okImpl, err := i.typeImplementsInterface(typeInfo{name: sourceTypeName}, targetName, nil, make(map[interfaceImplCacheKey]struct{}))
		return err == nil && okImpl, true
	}

	sourceIntegerKind := runtime.IntegerType(sourceTypeName)
	_, sourceIsInteger := lookupIntegerInfo(sourceIntegerKind)
	targetIntegerKind := runtime.IntegerType(targetName)
	_, targetIsInteger := lookupIntegerInfo(targetIntegerKind)
	sourceIsFloat := sourceTypeName == "f32" || sourceTypeName == "f64"

	switch targetName {
	case "bool", "char", "String":
		return sourceTypeName == targetName, true
	case "nil", "void", "IteratorEnd", "Iterator", "Future":
		return false, true
	case "f32", "f64":
		return sourceIsInteger || sourceIsFloat, true
	}

	if targetIsInteger {
		if !sourceIsInteger {
			return false, true
		}
		if sourceIntegerKind == targetIntegerKind || integerRangeWithinKinds(sourceIntegerKind, targetIntegerKind) {
			return true, true
		}
		return false, false
	}

	if i.isKnownTypeName(targetName) {
		return false, true
	}
	return false, false
}

func isCallableValue(value runtime.Value) bool {
	switch value.(type) {
	case *runtime.FunctionValue,
		*runtime.FunctionOverloadValue,
		runtime.NativeFunctionValue,
		*runtime.NativeFunctionValue,
		runtime.BoundMethodValue,
		*runtime.BoundMethodValue,
		runtime.NativeBoundMethodValue,
		*runtime.NativeBoundMethodValue,
		runtime.PartialFunctionValue,
		*runtime.PartialFunctionValue:
		return true
	default:
		return false
	}
}

func isPrimitiveName(name string) bool {
	switch name {
	case "bool", "String", "char", "nil", "void":
		return true
	case "f32", "f64":
		return true
	}
	if _, ok := lookupIntegerInfo(runtime.IntegerType(name)); ok {
		return true
	}
	return false
}

func normalizeKernelAliasName(name string) string {
	switch name {
	case "able.core.interfaces.Error":
		return "Error"
	case "able.core.iteration.Iterator":
		return "Iterator"
	case "able.concurrency.future.Future":
		return "Future"
	case "KernelArray":
		return "Array"
	case "KernelChannel":
		return "Channel"
	case "KernelHashMap":
		return "HashMap"
	case "KernelMutex":
		return "Mutex"
	case "KernelRange":
		return "Range"
	case "KernelRangeFactory":
		return "RangeFactory"
	case "KernelRatio":
		return "Ratio"
	case "KernelAwaitable":
		return "Awaitable"
	case "KernelAwaitWaker":
		return "AwaitWaker"
	case "KernelAwaitRegistration":
		return "AwaitRegistration"
	case "KernelLess":
		return "Less"
	case "KernelGreater":
		return "Greater"
	case "KernelEqual":
		return "Equal"
	case "KernelOrdering":
		return "Ordering"
	default:
		return name
	}
}

func isSingletonStructDef(def *ast.StructDefinition) bool {
	if def == nil || len(def.GenericParams) > 0 {
		return false
	}
	if def.Kind == ast.StructKindSingleton {
		return true
	}
	return def.Kind == ast.StructKindNamed && len(def.Fields) == 0
}

func primitiveImplementsInterfaceMethod(typeName, ifaceName, methodName string) bool {
	if typeName == "" || typeName == "void" {
		return false
	}
	if typeName == "nil" {
		return ifaceName == "Clone" && methodName == "clone"
	}
	if !isPrimitiveName(typeName) {
		return false
	}
	isInteger := false
	if _, err := getIntegerInfo(runtime.IntegerType(typeName)); err == nil {
		isInteger = true
	}
	isFloat := typeName == "f32" || typeName == "f64"
	isComparable := typeName == "String" || typeName == "bool" || typeName == "char" || isInteger
	switch ifaceName {
	case "Hash":
		return methodName == "hash" && isComparable
	case "Eq":
		return (methodName == "eq" || methodName == "ne") && isComparable
	case "PartialEq":
		return (methodName == "eq" || methodName == "ne") && (isComparable || isFloat)
	case "Clone":
		return methodName == "clone"
	case "Ord":
		return methodName == "cmp" && isComparable
	case "PartialOrd":
		return methodName == "partial_cmp" && (isComparable || isFloat)
	default:
		return false
	}
}
