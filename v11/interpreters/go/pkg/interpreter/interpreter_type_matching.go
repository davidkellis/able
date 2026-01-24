package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) matchesType(typeExpr ast.TypeExpression, value runtime.Value) bool {
	if typeExpr != nil {
		if expanded := expandTypeAliases(typeExpr, i.typeAliases, nil); expanded != nil {
			typeExpr = expanded
		}
	}
	if valueExpr := i.typeExpressionFromValue(value); valueExpr != nil {
		expandedValue := expandTypeAliases(valueExpr, i.typeAliases, nil)
		if expandedValue != nil && typeExpressionsEqual(typeExpr, expandedValue) {
			return true
		}
	}
	switch t := typeExpr.(type) {
	case *ast.WildcardTypeExpression:
		return true
	case *ast.SimpleTypeExpression:
		name := normalizeKernelAliasName(t.Name.Name)
		if name == "Error" {
			switch value.(type) {
			case runtime.ErrorValue, *runtime.ErrorValue:
				return true
			}
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
		case "Iterator":
			switch value.(type) {
			case *runtime.IteratorValue:
				return true
			}
			return false
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
					info, ok := i.getTypeInfoForValue(value)
					if !ok {
						return false
					}
					okImpl, err := i.typeImplementsInterface(info, name, nil, make(map[string]struct{}))
					return err == nil && okImpl
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
		if baseName == "Iterator" {
			if _, ok := value.(*runtime.IteratorValue); ok {
				return true
			}
		}
		if baseName == "Array" {
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
			switch v := value.(type) {
			case runtime.ErrorValue, *runtime.ErrorValue:
				return true
			case runtime.InterfaceValue:
				if v.Interface != nil && v.Interface.Node != nil && v.Interface.Node.ID != nil && v.Interface.Node.ID.Name == "Error" {
					return true
				}
			case *runtime.InterfaceValue:
				if v != nil && v.Interface != nil && v.Interface.Node != nil && v.Interface.Node.ID != nil && v.Interface.Node.ID.Name == "Error" {
					return true
				}
			}
			if info, ok := i.getTypeInfoForValue(value); ok {
				okImpl, err := i.typeImplementsInterface(info, "Error", nil, make(map[string]struct{}))
				return err == nil && okImpl
			}
			return false
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
					info, ok := i.getTypeInfoForValue(value)
					if !ok {
						return false
					}
					okImpl, err := i.typeImplementsInterface(info, baseName, t.Arguments, make(map[string]struct{}))
					return err == nil && okImpl
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
		switch v := value.(type) {
		case runtime.ErrorValue, *runtime.ErrorValue:
			return true
		case runtime.InterfaceValue:
			if v.Interface != nil && v.Interface.Node != nil && v.Interface.Node.ID != nil && v.Interface.Node.ID.Name == "Error" {
				return true
			}
		case *runtime.InterfaceValue:
			if v != nil && v.Interface != nil && v.Interface.Node != nil && v.Interface.Node.ID != nil && v.Interface.Node.ID.Name == "Error" {
				return true
			}
		}
		if info, ok := i.getTypeInfoForValue(value); ok {
			okImpl, err := i.typeImplementsInterface(info, "Error", nil, make(map[string]struct{}))
			return err == nil && okImpl
		}
		return false
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

func (i *Interpreter) isKnownTypeName(name string) bool {
	if name == "" {
		return false
	}
	for _, pkg := range i.packageRegistry {
		if val, ok := pkg[name]; ok {
			switch val.(type) {
			case *runtime.StructDefinitionValue, runtime.UnionDefinitionValue:
				return true
			}
		}
	}
	return false
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
	if _, err := getIntegerInfo(runtime.IntegerType(name)); err == nil {
		return true
	}
	return false
}

func normalizeKernelAliasName(name string) string {
	switch name {
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
	if typeName == "" || typeName == "nil" || typeName == "void" {
		return false
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
