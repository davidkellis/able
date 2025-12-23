package interpreter

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) getTypeInfoForValue(value runtime.Value) (typeInfo, bool) {
	switch v := value.(type) {
	case *runtime.StructDefinitionValue:
		if v == nil || v.Node == nil || v.Node.ID == nil || !isSingletonStructDef(v.Node) {
			return typeInfo{}, false
		}
		return typeInfo{name: v.Node.ID.Name}, true
	case runtime.StructDefinitionValue:
		return i.getTypeInfoForValue(&v)
	case *runtime.StructInstanceValue:
		return i.typeInfoFromStructInstance(v)
	case *runtime.InterfaceValue:
		return i.getTypeInfoForValue(v.Underlying)
	case runtime.StringValue, runtime.BoolValue, runtime.CharValue, runtime.NilValue, runtime.VoidValue,
		runtime.IntegerValue, *runtime.IntegerValue,
		runtime.FloatValue, *runtime.FloatValue,
		*runtime.ArrayValue:
		typeExpr := i.typeExpressionFromValue(value)
		if info, ok := parseTypeExpression(typeExpr); ok {
			return info, true
		}
		return typeInfo{}, false
	case runtime.ErrorValue:
		if v.TypeName != nil && v.TypeName.Name != "" {
			return typeInfo{name: v.TypeName.Name}, true
		}
		return typeInfo{name: "Error"}, true
	case *runtime.ErrorValue:
		if v == nil {
			return typeInfo{}, false
		}
		if v.TypeName != nil && v.TypeName.Name != "" {
			return typeInfo{name: v.TypeName.Name}, true
		}
		return typeInfo{name: "Error"}, true
	default:
		return typeInfo{}, false
	}
}

func (i *Interpreter) typeExpressionFromValue(value runtime.Value) ast.TypeExpression {
	switch v := value.(type) {
	case *runtime.StructDefinitionValue:
		if v == nil || v.Node == nil || v.Node.ID == nil || !isSingletonStructDef(v.Node) {
			return nil
		}
		return ast.Ty(v.Node.ID.Name)
	case runtime.StructDefinitionValue:
		return i.typeExpressionFromValue(&v)
	case runtime.StringValue:
		return ast.Ty("String")
	case runtime.BoolValue:
		return ast.Ty("bool")
	case runtime.CharValue:
		return ast.Ty("char")
	case runtime.NilValue:
		return ast.Ty("nil")
	case runtime.VoidValue:
		return ast.Ty("void")
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
		generics := v.Definition.Node.GenericParams
		if len(generics) > 0 {
			typeArgs := v.TypeArguments
			needsInference := len(typeArgs) != len(generics)
			if !needsInference {
				genericNames := genericNameSet(generics)
				for _, arg := range typeArgs {
					if arg == nil {
						needsInference = true
						break
					}
					if _, ok := arg.(*ast.WildcardTypeExpression); ok {
						needsInference = true
						break
					}
					if simple, ok := arg.(*ast.SimpleTypeExpression); ok && simple.Name != nil {
						if _, ok := genericNames[simple.Name.Name]; ok {
							needsInference = true
							break
						}
					}
				}
			}
			if needsInference {
				typeArgs = i.inferStructTypeArguments(v.Definition.Node, v.Fields, v.Positional)
			}
			if len(typeArgs) > 0 {
				return ast.Gen(base, typeArgs...)
			}
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

func (i *Interpreter) canonicalTypeNames(name string) []string {
	seen := make(map[string]struct{})
	add := func(n string) {
		if n == "" {
			return
		}
		if _, ok := seen[n]; ok {
			return
		}
		seen[n] = struct{}{}
	}
	add(name)
	if name != "" {
		if expanded := expandTypeAliases(ast.Ty(name), i.typeAliases, nil); expanded != nil {
			if info, ok := parseTypeExpression(expanded); ok && info.name != "" {
				add(info.name)
			}
		}
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	return names
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
		return nil, fmt.Errorf("ambiguous implementations of %s for %s: %s", interfaceName, typeDesc, strings.Join(detail, ", "))
	}
	if best == nil {
		return nil, nil
	}
	return best, nil
}

func (i *Interpreter) findMethod(info typeInfo, methodName string, interfaceFilter string) (runtime.Value, error) {
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
					method = &runtime.FunctionValue{Declaration: defaultDef, Closure: ifaceDef.Env, MethodPriority: -1}
					if cand.entry.methods == nil {
						cand.entry.methods = make(map[string]runtime.Value)
					}
					mergeFunctionLike(cand.entry.methods, methodName, method)
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
		ifaceName := methodName
		if len(ambiguous) > 0 && ambiguous[0].candidate.entry != nil && ambiguous[0].candidate.entry.interfaceName != "" {
			ifaceName = ambiguous[0].candidate.entry.interfaceName
		}
		if len(detail) == 0 {
			detail = []string{"<unknown>"}
		}
		return nil, fmt.Errorf("ambiguous implementations of %s for %s: %s", ifaceName, typeDesc, strings.Join(detail, ", "))
	}
	if best == nil {
		return nil, nil
	}
	if fnVal := firstFunction(best.method); fnVal != nil {
		if fnDef, ok := fnVal.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
			return nil, fmt.Errorf("Method '%s' on %s is private", methodName, info.name)
		}
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

func (i *Interpreter) selectStructMethod(inst *runtime.StructInstanceValue, methodName string) (runtime.Value, error) {
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
	if typeExpr != nil {
		if expanded := expandTypeAliases(typeExpr, i.typeAliases, nil); expanded != nil {
			typeExpr = expanded
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
		switch name {
		case "String":
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
			if iv.TypeSuffix == targetKind || integerRangeWithinKinds(iv.TypeSuffix, targetKind) {
				return true
			}
			if iv.Val != nil && integerValueWithinRange(iv.Val, targetKind) {
				return true
			}
			return false
		case "f32", "f64":
			fv, ok := value.(runtime.FloatValue)
			if !ok {
				return false
			}
			return string(fv.TypeSuffix) == name
		default:
			if unionDef, ok := i.unionDefinitions[name]; ok && unionDef != nil && unionDef.Node != nil {
				for _, variant := range unionDef.Node.Variants {
					if variant != nil && i.matchesType(variant, value) {
						return true
					}
				}
				return false
			}
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
			if defVal, ok := value.(*runtime.StructDefinitionValue); ok && isSingletonStructDef(defVal.Node) {
				if defVal.Node != nil && defVal.Node.ID != nil {
					return defVal.Node.ID.Name == name
				}
			}
			if defVal, ok := value.(runtime.StructDefinitionValue); ok && isSingletonStructDef(defVal.Node) {
				if defVal.Node != nil && defVal.Node.ID != nil {
					return defVal.Node.ID.Name == name
				}
			}
			if structVal, ok := value.(*runtime.StructInstanceValue); ok {
				if structVal.Definition != nil && structVal.Definition.Node != nil && structVal.Definition.Node.ID != nil {
					return structVal.Definition.Node.ID.Name == name
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
					if variant != nil && i.matchesType(variant, value) {
						return true
					}
				}
				return false
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
		return runtime.IsFunctionLike(value)
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
	switch ifaceName {
	case "Hash":
		return methodName == "hash"
	case "Eq":
		return methodName == "eq" || methodName == "ne"
	default:
		return false
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
			if name == "Error" {
				if _, ok := value.(runtime.ErrorValue); ok {
					return value, nil
				}
				if errVal, ok := value.(*runtime.ErrorValue); ok && errVal != nil {
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
	methods := make(map[string]runtime.Value, len(candidate.entry.methods))
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
