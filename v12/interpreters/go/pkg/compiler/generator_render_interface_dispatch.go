package compiler

import (
	"bytes"
	"fmt"
	"sort"
)

func (g *generator) renderCompiledInterfaceDispatchFile() ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", g.opts.PackageName)

	imports := []string{
		"able/interpreter-go/pkg/ast",
		"able/interpreter-go/pkg/compiler/bridge",
		"able/interpreter-go/pkg/runtime",
		"fmt",
	}
	sort.Strings(imports)
	fmt.Fprintf(&buf, "import (\n")
	for _, imp := range imports {
		fmt.Fprintf(&buf, "\t%q\n", imp)
	}
	fmt.Fprintf(&buf, ")\n\n")
	fmt.Fprintf(&buf, "var (\n")
	fmt.Fprintf(&buf, "\t_ = ast.NewIdentifier\n")
	fmt.Fprintf(&buf, "\t_ = bridge.New\n")
	fmt.Fprintf(&buf, "\t_ = fmt.Errorf\n")
	fmt.Fprintf(&buf, "\t_ runtime.Value\n")
	fmt.Fprintf(&buf, ")\n\n")

	methodOverloads := g.methodOverloadGroups()
	fmt.Fprintf(&buf, "func __able_register_compiled_interface_dispatch(rt *bridge.Runtime) error {\n")
	fmt.Fprintf(&buf, "\t_ = rt\n")
	for _, group := range g.interfaceDispatchGroups() {
		if group == nil || len(group.Entries) == 0 {
			continue
		}
		entryInfo := group.Entries[0]
		if entryInfo == nil || entryInfo.Info == nil {
			continue
		}
		info := entryInfo.Info
		arity := info.Arity
		if methodDefinitionExpectsSelf(info.Definition) && arity > 0 {
			arity--
		}
		if len(group.Entries) == 1 {
			if arity < 0 {
				arity = -1
			}
			fmt.Fprintf(&buf, "\t{\n")
			fmt.Fprintf(&buf, "\t\tfn := &runtime.NativeFunctionValue{Name: %q, Arity: %d}\n", entryInfo.MethodName, arity)
			fmt.Fprintf(&buf, "\t\tfn.Impl = func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
			fmt.Fprintf(&buf, "\t\t\treturn __able_wrap_%s(__able_runtime, ctx, args)\n", info.GoName)
			fmt.Fprintf(&buf, "\t\t}\n")
			fmt.Fprintf(&buf, "\t\t__able_register_interface_dispatch(%q, %q, %s, %s, %s, %s, fn, %t, false)\n", group.InterfaceName, group.MethodName, group.TargetExpr, group.InterfaceArgs, group.GenericNames, group.Constraints, group.IsPrivate)
			fmt.Fprintf(&buf, "\t}\n")
			continue
		}
		minArgs := -1
		for _, entry := range group.Entries {
			if entry == nil || entry.Info == nil || entry.Info.Definition == nil {
				continue
			}
			def := entry.Info.Definition
			expectsSelf := methodDefinitionExpectsSelf(def)
			paramTypes := methodDefinitionParamTypes(def, entry.TargetType, expectsSelf)
			paramCount := len(paramTypes)
			entryMin := paramCount
			if len(def.Params) > 0 && isNullableParam(def.Params[len(def.Params)-1]) {
				entryMin--
			}
			if entryMin < 0 {
				entryMin = 0
			}
			if minArgs < 0 || entryMin < minArgs {
				minArgs = entryMin
			}
		}
		fmt.Fprintf(&buf, "\t{\n")
		fmt.Fprintf(&buf, "\t\tfn := &runtime.NativeFunctionValue{Name: %q, Arity: -1}\n", group.MethodName)
		fmt.Fprintf(&buf, "\t\tfn.Impl = func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
		if minArgs > 0 {
			fmt.Fprintf(&buf, "\t\t\tif len(args) < %d {\n", minArgs)
			fmt.Fprintf(&buf, "\t\t\t\targsCopy := append([]runtime.Value{}, args...)\n")
			fmt.Fprintf(&buf, "\t\t\t\treturn runtime.PartialFunctionValue{Target: fn, BoundArgs: argsCopy}, nil\n")
			fmt.Fprintf(&buf, "\t\t\t}\n")
		}
		fmt.Fprintf(&buf, "\t\t\tbestIdx := -1\n")
		fmt.Fprintf(&buf, "\t\t\tbestScore := 0.0\n")
		fmt.Fprintf(&buf, "\t\t\tties := 0\n")
		fmt.Fprintf(&buf, "\t\t\tvar bestArgs []runtime.Value\n")
		for idx, entry := range group.Entries {
			if entry == nil || entry.Info == nil || entry.Info.Definition == nil {
				continue
			}
			def := entry.Info.Definition
			expectsSelf := methodDefinitionExpectsSelf(def)
			paramTypes := methodDefinitionParamTypes(def, entry.TargetType, expectsSelf)
			paramCount := len(paramTypes)
			optionalLast := len(def.Params) > 0 && isNullableParam(def.Params[len(def.Params)-1])
			generics := combinedGenericNameSet(entry.ImplGenerics, entry.InterfaceGenerics, def.GenericParams)
			fmt.Fprintf(&buf, "\t\t\t{\n")
			fmt.Fprintf(&buf, "\t\t\t\tif len(args) == %d", paramCount)
			if optionalLast {
				fmt.Fprintf(&buf, " || len(args) == %d", paramCount-1)
			}
			fmt.Fprintf(&buf, " {\n")
			if optionalLast {
				fmt.Fprintf(&buf, "\t\t\t\t\tmissingOptional := len(args) == %d\n", paramCount-1)
			}
			fmt.Fprintf(&buf, "\t\t\t\t\tcompatible := true\n")
			fmt.Fprintf(&buf, "\t\t\t\t\tscore := 0.0\n")
			fmt.Fprintf(&buf, "\t\t\t\t\tcoercedArgs := make([]runtime.Value, len(args))\n")
			if optionalLast {
				fmt.Fprintf(&buf, "\t\t\t\t\tif missingOptional { score -= 0.5 }\n")
			}
			for pIdx, paramType := range paramTypes {
				fmt.Fprintf(&buf, "\t\t\t\t\tif len(args) > %d {\n", pIdx)
				if paramType == nil {
					fmt.Fprintf(&buf, "\t\t\t\t\t\tcompatible = false\n")
				} else if typeExprUsesGeneric(paramType, generics) {
					fmt.Fprintf(&buf, "\t\t\t\t\t\tcoercedArgs[%d] = args[%d]\n", pIdx, pIdx)
				} else if typeExpr, ok := g.renderTypeExpression(paramType); ok {
					spec := parameterSpecificity(paramType, generics)
					fmt.Fprintf(&buf, "\t\t\t\t\t\tif compatible {\n")
					fmt.Fprintf(&buf, "\t\t\t\t\t\t\tcoerced, ok, err := bridge.MatchType(__able_runtime, %s, args[%d])\n", typeExpr, pIdx)
					fmt.Fprintf(&buf, "\t\t\t\t\t\t\tif err != nil { return nil, err }\n")
					fmt.Fprintf(&buf, "\t\t\t\t\t\t\tif !ok { compatible = false } else { coercedArgs[%d] = coerced; score += %d }\n", pIdx, spec)
					fmt.Fprintf(&buf, "\t\t\t\t\t\t}\n")
				} else {
					fmt.Fprintf(&buf, "\t\t\t\t\t\tcompatible = false\n")
				}
				fmt.Fprintf(&buf, "\t\t\t\t\t}\n")
			}
			fmt.Fprintf(&buf, "\t\t\t\t\tif compatible {\n")
			if optionalLast {
				fmt.Fprintf(&buf, "\t\t\t\t\t\tif missingOptional { coercedArgs = append(coercedArgs, runtime.NilValue{}) }\n")
			}
			fmt.Fprintf(&buf, "\t\t\t\t\t\tif bestIdx < 0 || score > bestScore {\n")
			fmt.Fprintf(&buf, "\t\t\t\t\t\t\tbestIdx = %d\n", idx)
			fmt.Fprintf(&buf, "\t\t\t\t\t\t\tbestScore = score\n")
			fmt.Fprintf(&buf, "\t\t\t\t\t\t\tbestArgs = coercedArgs\n")
			fmt.Fprintf(&buf, "\t\t\t\t\t\t\tties = 1\n")
			fmt.Fprintf(&buf, "\t\t\t\t\t\t} else if score == bestScore {\n")
			fmt.Fprintf(&buf, "\t\t\t\t\t\t\tties++\n")
			fmt.Fprintf(&buf, "\t\t\t\t\t\t}\n")
			fmt.Fprintf(&buf, "\t\t\t\t\t}\n")
			fmt.Fprintf(&buf, "\t\t\t\t}\n")
			fmt.Fprintf(&buf, "\t\t\t}\n")
		}
		fmt.Fprintf(&buf, "\t\t\tif bestIdx < 0 {\n")
		fmt.Fprintf(&buf, "\t\t\t\treturn nil, fmt.Errorf(\"No overloads of %s match provided arguments\")\n", group.MethodName)
		fmt.Fprintf(&buf, "\t\t\t}\n")
		fmt.Fprintf(&buf, "\t\t\tif ties > 1 {\n")
		fmt.Fprintf(&buf, "\t\t\t\treturn nil, fmt.Errorf(\"Ambiguous overload for %s\")\n", group.MethodName)
		fmt.Fprintf(&buf, "\t\t\t}\n")
		fmt.Fprintf(&buf, "\t\t\tswitch bestIdx {\n")
		for idx, entry := range group.Entries {
			if entry == nil || entry.Info == nil {
				continue
			}
			fmt.Fprintf(&buf, "\t\t\tcase %d:\n", idx)
			fmt.Fprintf(&buf, "\t\t\t\treturn __able_wrap_%s(__able_runtime, ctx, bestArgs)\n", entry.Info.GoName)
		}
		fmt.Fprintf(&buf, "\t\t\t}\n")
		fmt.Fprintf(&buf, "\t\t\treturn nil, fmt.Errorf(\"No overloads of %s match provided arguments\")\n", group.MethodName)
		fmt.Fprintf(&buf, "\t\t}\n")
		fmt.Fprintf(&buf, "\t\t__able_register_interface_dispatch(%q, %q, %s, %s, %s, %s, fn, %t, false)\n", group.InterfaceName, group.MethodName, group.TargetExpr, group.InterfaceArgs, group.GenericNames, group.Constraints, group.IsPrivate)
		fmt.Fprintf(&buf, "\t}\n")
	}
	if g.interfaceDispatchStrict() {
		fmt.Fprintf(&buf, "\t__able_interface_dispatch_strict = true\n")
	}
	for _, group := range methodOverloads {
		if group == nil {
			continue
		}
		minArgs := group.MinArity
		if minArgs < 0 {
			minArgs = 0
		}
		if group.ExpectsSelf {
			minArgs--
		}
		if minArgs < 0 {
			minArgs = 0
		}
		wrapperName := g.methodOverloadWrapperName(group.TargetName, group.MethodName, group.ExpectsSelf)
		valueName := g.methodOverloadValueName(group.TargetName, group.MethodName, group.ExpectsSelf)
		fmt.Fprintf(&buf, "\t__able_register_compiled_method(%q, %q, %t, -1, %d, %s)\n", group.TargetName, group.MethodName, group.ExpectsSelf, minArgs, wrapperName)
		fmt.Fprintf(&buf, "\tif entry := __able_lookup_compiled_method(%q, %q, %t); entry != nil {\n", group.TargetName, group.MethodName, group.ExpectsSelf)
		fmt.Fprintf(&buf, "\t\t%s = entry.fn\n", valueName)
		fmt.Fprintf(&buf, "\t}\n")
	}
	fmt.Fprintf(&buf, "\treturn nil\n")
	fmt.Fprintf(&buf, "}\n")

	return formatSource(buf.Bytes())
}
