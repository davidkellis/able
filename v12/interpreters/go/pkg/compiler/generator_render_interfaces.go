package compiler

import (
	"bytes"
	"fmt"
	"strings"
)

func (g *generator) renderNativeInterfaces(buf *bytes.Buffer) {
	if g == nil || buf == nil || len(g.nativeInterfaces) == 0 {
		return
	}
	if g.nativeInterfaceRenderedAdapters == nil {
		g.nativeInterfaceRenderedAdapters = make(map[string]struct{})
	}
	if g.nativeInterfaceRenderedInfos == nil {
		g.nativeInterfaceRenderedInfos = make(map[string]struct{})
	}
	if g.nativeInterfaceRenderedDispatches == nil {
		g.nativeInterfaceRenderedDispatches = make(map[string]struct{})
	}
	if g.nativeInterfaceRenderedApplyHelpers == nil {
		g.nativeInterfaceRenderedApplyHelpers = make(map[string]struct{})
	}
	for {
		progress := g.renderPendingNativeInterfaceScaffolding(buf)
		if g.renderPendingNativeInterfaceConcreteAdapters(buf) {
			progress = true
		}
		if g.renderNativeInterfaceGenericDispatchHelpers(buf) {
			progress = true
		}
		if !progress {
			break
		}
	}
}

func (g *generator) renderPendingNativeInterfaceScaffolding(buf *bytes.Buffer) bool {
	if g == nil || buf == nil {
		return false
	}
	progress := false
	for _, key := range g.sortedNativeInterfaceKeys() {
		info := g.nativeInterfaces[key]
		if info == nil {
			continue
		}
		if _, ok := g.nativeInterfaceRenderedInfos[key]; ok {
			continue
		}
		g.renderNativeInterfaceType(buf, info)
		g.renderNativeInterfaceRuntimeIteratorAdapter(buf, info)
		g.renderNativeInterfaceRuntimeAdapter(buf, info)
		g.nativeInterfaceRenderedInfos[key] = struct{}{}
		progress = true
	}
	return progress
}

func (g *generator) renderNativeInterfaceBoundaryHelpersFinal(buf *bytes.Buffer) {
	if g == nil || buf == nil {
		return
	}
	for _, key := range g.sortedNativeInterfaceKeys() {
		info := g.nativeInterfaces[key]
		if info == nil {
			continue
		}
		if info.AdapterVersion != g.nativeInterfaceAdapterVersion {
			g.refreshNativeInterfaceAdapters(info)
		}
		g.renderNativeInterfaceBoundaryHelpers(buf, info)
	}
}

func (g *generator) renderPendingNativeInterfaceConcreteAdapters(buf *bytes.Buffer) bool {
	if g == nil || buf == nil {
		return false
	}
	progress := g.renderPendingNativeInterfaceScaffolding(buf)
	for _, key := range g.sortedNativeInterfaceKeys() {
		info := g.nativeInterfaces[key]
		if info == nil {
			continue
		}
		if info.AdapterVersion != g.nativeInterfaceAdapterVersion {
			g.refreshNativeInterfaceAdapters(info)
		}
		for _, adapter := range g.nativeInterfaceKnownAdapters(info) {
			if adapter == nil || adapter.GoType == "" {
				continue
			}
			renderKey := info.Key + "::" + adapter.GoType
			if _, ok := g.nativeInterfaceRenderedAdapters[renderKey]; ok {
				continue
			}
			g.renderNativeInterfaceConcreteAdapter(buf, info, adapter)
			g.nativeInterfaceRenderedAdapters[renderKey] = struct{}{}
			progress = true
		}
	}
	for key, adapters := range g.nativeInterfaceExplicitAdapters {
		info := g.nativeInterfaces[key]
		if info == nil {
			continue
		}
		for _, adapter := range adapters {
			if adapter == nil || adapter.GoType == "" {
				continue
			}
			renderKey := info.Key + "::" + adapter.GoType
			if _, ok := g.nativeInterfaceRenderedAdapters[renderKey]; ok {
				continue
			}
			g.renderNativeInterfaceConcreteAdapter(buf, info, adapter)
			g.nativeInterfaceRenderedAdapters[renderKey] = struct{}{}
			progress = true
		}
	}
	return progress
}

func (g *generator) renderNativeInterfaceApplyRuntimeHelpers(buf *bytes.Buffer) {
	if g == nil || buf == nil {
		return
	}
	for _, key := range g.sortedNativeInterfaceKeys() {
		info := g.nativeInterfaces[key]
		if info == nil {
			continue
		}
		if _, ok := g.nativeInterfaceRenderedApplyHelpers[key]; ok {
			continue
		}
		if info.AdapterVersion != g.nativeInterfaceAdapterVersion {
			g.refreshNativeInterfaceAdapters(info)
		}
		g.renderNativeInterfaceApplyRuntimeHelper(buf, info)
		g.nativeInterfaceRenderedApplyHelpers[key] = struct{}{}
	}
}

func (g *generator) renderNativeInterfaceRuntimeIteratorAdapter(buf *bytes.Buffer, info *nativeInterfaceInfo) {
	if g == nil || buf == nil || info == nil || info.RuntimeIteratorAdapter == "" {
		return
	}
	fmt.Fprintf(buf, "type %s struct {\n", info.RuntimeIteratorAdapter)
	fmt.Fprintf(buf, "\tValue *runtime.IteratorValue\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (%s) %s() {}\n\n", info.RuntimeIteratorAdapter, info.MarkerMethod)
	fmt.Fprintf(buf, "func (w %s) %s(rt *bridge.Runtime) (runtime.Value, error) {\n", info.RuntimeIteratorAdapter, info.ToRuntimeMethod)
	fmt.Fprintf(buf, "\t_ = rt\n")
	fmt.Fprintf(buf, "\tif w.Value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn w.Value, nil\n")
	fmt.Fprintf(buf, "}\n\n")
	for _, method := range info.Methods {
		if method == nil {
			continue
		}
		if method.Name == "next" {
			fmt.Fprintf(buf, "func (w %s) %s() (%s, *__ableControl) {\n", info.RuntimeIteratorAdapter, method.GoName, method.ReturnGoType)
			zeroExpr, zeroOK := g.zeroValueExpr(method.ReturnGoType)
			if !zeroOK {
				fmt.Fprintf(buf, "\tvar zero %s\n", method.ReturnGoType)
				zeroExpr = "zero"
			}
			fmt.Fprintf(buf, "\tif w.Value == nil {\n")
			fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(fmt.Errorf(\"missing interface value\"))\n", zeroExpr)
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\tresult, done, err := w.Value.Next()\n")
			fmt.Fprintf(buf, "\tif err != nil {\n")
			fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\tif done {\n")
			fmt.Fprintf(buf, "\t\tresult = runtime.IteratorEnd\n")
			fmt.Fprintf(buf, "\t} else if result == nil {\n")
			fmt.Fprintf(buf, "\t\tresult = runtime.NilValue{}\n")
			fmt.Fprintf(buf, "\t}\n")
			if g.renderNativeInterfaceRuntimeToGoValueControl(buf, "converted", "result", method.ReturnGoType, method.ReturnGoType) {
				fmt.Fprintf(buf, "\treturn converted, nil\n")
			} else {
				fmt.Fprintf(buf, "\treturn %s, __able_control_from_error(fmt.Errorf(\"unsupported native iterator conversion to %%s\", %q))\n", zeroExpr, method.ReturnGoType)
			}
			fmt.Fprintf(buf, "}\n\n")
			continue
		}
		fmt.Fprintf(buf, "func (w %s) %s(", info.RuntimeIteratorAdapter, method.GoName)
		args := make([]string, 0, len(method.ParamGoTypes))
		for idx, paramType := range method.ParamGoTypes {
			if idx > 0 {
				fmt.Fprintf(buf, ", ")
			}
			fmt.Fprintf(buf, "arg%d %s", idx, paramType)
			args = append(args, fmt.Sprintf("arg%d", idx))
		}
		fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", method.ReturnGoType)
		fmt.Fprintf(buf, "\treturn %s{Value: w.Value}.%s(%s)\n", info.RuntimeAdapter, method.GoName, strings.Join(args, ", "))
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderNativeInterfaceType(buf *bytes.Buffer, info *nativeInterfaceInfo) {
	fmt.Fprintf(buf, "type %s interface {\n", info.GoType)
	fmt.Fprintf(buf, "\t%s()\n", info.MarkerMethod)
	fmt.Fprintf(buf, "\t%s(*bridge.Runtime) (runtime.Value, error)\n", info.ToRuntimeMethod)
	for _, method := range info.Methods {
		if method == nil {
			continue
		}
		fmt.Fprintf(buf, "\t%s(", method.GoName)
		for idx, paramType := range method.ParamGoTypes {
			if idx > 0 {
				fmt.Fprintf(buf, ", ")
			}
			fmt.Fprintf(buf, "arg%d %s", idx, paramType)
		}
		fmt.Fprintf(buf, ") (%s, *__ableControl)\n", method.ReturnGoType)
	}
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderNativeInterfaceRuntimeAdapter(buf *bytes.Buffer, info *nativeInterfaceInfo) {
	fmt.Fprintf(buf, "type %s struct {\n", info.RuntimeAdapter)
	fmt.Fprintf(buf, "\tValue runtime.Value\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (%s) %s() {}\n\n", info.RuntimeAdapter, info.MarkerMethod)
	fmt.Fprintf(buf, "func %s(value runtime.Value) %s {\n", info.RuntimeWrapHelper, info.GoType)
	if info.RuntimeIteratorAdapter != "" {
		fmt.Fprintf(buf, "\tif iter, ok, _ := __able_runtime_iterator_value(value); ok && iter != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s{Value: iter}\n", info.RuntimeIteratorAdapter)
		fmt.Fprintf(buf, "\t}\n")
	}
	fmt.Fprintf(buf, "\treturn %s{Value: value}\n", info.RuntimeAdapter)
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (w %s) %s(rt *bridge.Runtime) (runtime.Value, error) {\n", info.RuntimeAdapter, info.ToRuntimeMethod)
	fmt.Fprintf(buf, "\t_ = rt\n")
	fmt.Fprintf(buf, "\tif w.Value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn w.Value, nil\n")
	fmt.Fprintf(buf, "}\n\n")
	for _, method := range info.Methods {
		if method == nil {
			continue
		}
		g.renderNativeInterfaceRuntimeMethod(buf, info, method)
	}
}

func (g *generator) renderNativeInterfaceRuntimeMethod(buf *bytes.Buffer, info *nativeInterfaceInfo, method *nativeInterfaceMethod) {
	fmt.Fprintf(buf, "func (w %s) %s(", info.RuntimeAdapter, method.GoName)
	for idx, paramType := range method.ParamGoTypes {
		if idx > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "arg%d %s", idx, paramType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", method.ReturnGoType)
	zeroExpr, zeroOK := g.zeroValueExpr(method.ReturnGoType)
	if !zeroOK {
		fmt.Fprintf(buf, "\tvar zero %s\n", method.ReturnGoType)
	}
	fmt.Fprintf(buf, "\tif w.Value == nil {\n")
	if zeroOK {
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(fmt.Errorf(\"missing interface value\"))\n", zeroExpr)
	} else {
		fmt.Fprintf(buf, "\t\treturn zero, __able_control_from_error(fmt.Errorf(\"missing interface value\"))\n")
	}
	fmt.Fprintf(buf, "\t}\n")
	baseName, _ := typeExprBaseName(info.TypeExpr)
	if baseName == "Iterator" && method.Name == "next" && len(method.ParamGoTypes) == 0 {
		fmt.Fprintf(buf, "\tif iter, ok, nilPtr := __able_runtime_iterator_value(w.Value); ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\tif !ok || nilPtr {\n")
		if zeroOK {
			fmt.Fprintf(buf, "\t\t\treturn %s, __able_control_from_error(fmt.Errorf(\"missing interface value\"))\n", zeroExpr)
		} else {
			fmt.Fprintf(buf, "\t\t\treturn zero, __able_control_from_error(fmt.Errorf(\"missing interface value\"))\n")
		}
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tresult, done, err := iter.Next()\n")
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		if zeroOK {
			fmt.Fprintf(buf, "\t\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		} else {
			fmt.Fprintf(buf, "\t\t\treturn zero, __able_control_from_error(err)\n")
		}
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tif done {\n")
		fmt.Fprintf(buf, "\t\t\tresult = runtime.IteratorEnd\n")
		fmt.Fprintf(buf, "\t\t} else if result == nil {\n")
		fmt.Fprintf(buf, "\t\t\tresult = runtime.NilValue{}\n")
		fmt.Fprintf(buf, "\t\t}\n")
		if g.renderNativeInterfaceRuntimeToGoValueControl(buf, "converted", "result", method.ReturnGoType, method.ReturnGoType) {
			fmt.Fprintf(buf, "\t\treturn converted, nil\n")
		}
		fmt.Fprintf(buf, "\t}\n")
	}
	argList := "nil"
	if len(method.ParamGoTypes) > 0 {
		fmt.Fprintf(buf, "\targs := make([]runtime.Value, 0, %d)\n", len(method.ParamGoTypes))
		for idx, paramType := range method.ParamGoTypes {
			target := fmt.Sprintf("arg%dValue", idx)
			g.renderNativeInterfaceGoToRuntimeValueControl(buf, target, fmt.Sprintf("arg%d", idx), paramType, method.ReturnGoType)
			fmt.Fprintf(buf, "\targs = append(args, %s)\n", target)
		}
		argList = "args"
	}
	fmt.Fprintf(buf, "\tresult, control := __able_method_call(w.Value, %q, %s)\n", method.Name, argList)
	for idx, paramType := range method.ParamGoTypes {
		if iface := g.nativeInterfaceInfoForGoType(paramType); iface != nil {
			fmt.Fprintf(buf, "\tif err := %s(arg%d, arg%dValue); err != nil {\n", iface.ApplyRuntimeHelper, idx, idx)
			fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
			fmt.Fprintf(buf, "\t}\n")
		}
	}
	fmt.Fprintf(buf, "\tif control != nil {\n")
	if zeroOK {
		fmt.Fprintf(buf, "\t\treturn %s, control\n", zeroExpr)
	} else {
		fmt.Fprintf(buf, "\t\treturn zero, control\n")
	}
	fmt.Fprintf(buf, "\t}\n")
	if g.renderNativeInterfaceRuntimeToGoValueControl(buf, "converted", "result", method.ReturnGoType, method.ReturnGoType) {
		fmt.Fprintf(buf, "\treturn converted, nil\n")
	}
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderNativeInterfaceConcreteAdapter(buf *bytes.Buffer, info *nativeInterfaceInfo, adapter *nativeInterfaceAdapter) {
	fmt.Fprintf(buf, "type %s struct {\n", adapter.AdapterType)
	fmt.Fprintf(buf, "\tValue %s\n", adapter.GoType)
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (%s) %s() {}\n\n", adapter.AdapterType, info.MarkerMethod)
	fmt.Fprintf(buf, "func %s(value %s) %s {\n", adapter.WrapHelper, adapter.GoType, info.GoType)
	fmt.Fprintf(buf, "\treturn %s{Value: value}\n", adapter.AdapterType)
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (w %s) %s(rt *bridge.Runtime) (runtime.Value, error) {\n", adapter.AdapterType, info.ToRuntimeMethod)
	g.renderNativeInterfaceGoToRuntimeValueError(buf, "converted", "w.Value", adapter.GoType)
	fmt.Fprintf(buf, "\treturn converted, nil\n")
	fmt.Fprintf(buf, "}\n\n")
	for _, method := range info.Methods {
		if method == nil {
			continue
		}
		impl := adapter.Methods[method.Name]
		if impl == nil {
			continue
		}
		fmt.Fprintf(buf, "func (w %s) %s(", adapter.AdapterType, method.GoName)
		for idx, paramType := range method.ParamGoTypes {
			if idx > 0 {
				fmt.Fprintf(buf, ", ")
			}
			fmt.Fprintf(buf, "arg%d %s", idx, paramType)
		}
		fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", method.ReturnGoType)
		zeroExpr, zeroOK := g.zeroValueExpr(method.ReturnGoType)
		if !zeroOK {
			fmt.Fprintf(buf, "\tvar zero %s\n", method.ReturnGoType)
			zeroExpr = "zero"
		}
		argCap := len(method.ParamGoTypes)
		if method.ExpectsSelf {
			argCap++
		}
		args := make([]string, 0, argCap)
		if method.ExpectsSelf {
			selfExpr := "w.Value"
			selfType := adapter.GoType
			if impl.Info != nil && len(impl.Info.Params) > 0 && impl.Info.Params[0].GoType != "" {
				implSelfType := g.canonicalMethodReceiverGoType(impl.Info, selfType)
				if implSelfType != selfType {
					if g.renderNativeInterfaceReceiverCoercionControl(buf, "\t", "__able_self_converted", "w.Value", selfType, implSelfType, method.ReturnGoType) {
						selfExpr = "__able_self_converted"
					} else {
						runtimeSelf := "__able_self_value"
						g.renderNativeInterfaceGoToRuntimeValueControl(buf, runtimeSelf, "w.Value", selfType, method.ReturnGoType)
						if implSelfType != "runtime.Value" {
							if !g.renderNativeInterfaceRuntimeToGoValueControl(buf, "__able_self_converted", runtimeSelf, implSelfType, method.ReturnGoType) {
								return
							}
							selfExpr = "__able_self_converted"
						} else {
							selfExpr = runtimeSelf
						}
					}
				}
			}
			args = append(args, selfExpr)
		}
		for idx, paramType := range method.ParamGoTypes {
			argExpr := fmt.Sprintf("arg%d", idx)
			implParamType := paramType
			if idx < len(impl.CompiledParamGoTypes) && impl.CompiledParamGoTypes[idx] != "" {
				implParamType = impl.CompiledParamGoTypes[idx]
			}
			if implParamType != paramType {
				runtimeArg := fmt.Sprintf("__able_arg%d_value", idx)
				g.renderNativeInterfaceGoToRuntimeValueControl(buf, runtimeArg, argExpr, paramType, method.ReturnGoType)
				if implParamType != "runtime.Value" {
					convertedArg := fmt.Sprintf("__able_arg%d_converted", idx)
					g.renderNativeInterfaceRuntimeToGoValueControl(buf, convertedArg, runtimeArg, implParamType, method.ReturnGoType)
					argExpr = convertedArg
				} else {
					argExpr = runtimeArg
				}
			}
			args = append(args, argExpr)
		}
		fmt.Fprintf(buf, "\tresult, control := __able_compiled_%s(%s)\n", impl.Info.GoName, strings.Join(args, ", "))
		fmt.Fprintf(buf, "\tif control != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, control\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		if impl.CompiledReturnGoType == method.ReturnGoType {
			fmt.Fprintf(buf, "\treturn result, nil\n")
			fmt.Fprintf(buf, "}\n\n")
			continue
		}
		if g.renderNativeInterfaceDirectCoercionControl(buf, "converted", "result", impl.CompiledReturnGoType, method.ReturnGoType, method.ReturnGoType) {
			fmt.Fprintf(buf, "\treturn converted, nil\n")
			fmt.Fprintf(buf, "}\n\n")
			continue
		}
		runtimeResult := "__able_result_value"
		g.renderNativeInterfaceGoToRuntimeValueControl(buf, runtimeResult, "result", impl.CompiledReturnGoType, method.ReturnGoType)
		if g.renderNativeInterfaceRuntimeToGoValueControl(buf, "converted", runtimeResult, method.ReturnGoType, method.ReturnGoType) {
			fmt.Fprintf(buf, "\treturn converted, nil\n")
			fmt.Fprintf(buf, "}\n\n")
			continue
		}
		fmt.Fprintf(buf, "\treturn result, nil\n")
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderNativeInterfaceDirectCoercionControl(buf *bytes.Buffer, target string, expr string, actualGoType string, expectedGoType string, failureGoType string) bool {
	if g == nil || buf == nil || target == "" || expr == "" || actualGoType == "" || expectedGoType == "" {
		return false
	}
	if actualGoType == expectedGoType {
		fmt.Fprintf(buf, "\t%s := %s\n", target, expr)
		return true
	}
	actualInfo := g.nativeInterfaceInfoForGoType(actualGoType)
	expectedInfo := g.nativeInterfaceInfoForGoType(expectedGoType)
	if actualInfo == nil || expectedInfo == nil {
		return false
	}
	zeroExpr, ok := g.zeroValueExpr(failureGoType)
	if !ok {
		return false
	}
	fmt.Fprintf(buf, "\tvar %s %s\n", target, expectedGoType)
	defaultLine := fmt.Sprintf("\t\treturn %s, __able_control_from_error(fmt.Errorf(\"unsupported native interface conversion from %s to %s\"))", zeroExpr, actualGoType, expectedGoType)
	return g.renderNativeInterfaceDirectCoercionSwitch(buf, "\t", target, expr, actualInfo, expectedInfo, defaultLine)
}

func (g *generator) renderNativeInterfaceDirectCoercionError(buf *bytes.Buffer, indent string, target string, expr string, actualGoType string, expectedGoType string, failureExpr string) bool {
	if g == nil || buf == nil || target == "" || expr == "" || actualGoType == "" || expectedGoType == "" || failureExpr == "" {
		return false
	}
	if actualGoType == expectedGoType {
		fmt.Fprintf(buf, "%s%s := %s\n", indent, target, expr)
		return true
	}
	actualInfo := g.nativeInterfaceInfoForGoType(actualGoType)
	expectedInfo := g.nativeInterfaceInfoForGoType(expectedGoType)
	if actualInfo == nil || expectedInfo == nil {
		return false
	}
	fmt.Fprintf(buf, "%svar %s %s\n", indent, target, expectedGoType)
	fmt.Fprintf(buf, "%sif %s == nil {\n", indent, expr)
	fmt.Fprintf(buf, "%s\t%s = %s(nil)\n", indent, target, expectedGoType)
	fmt.Fprintf(buf, "%s} else {\n", indent)
	defaultLine := fmt.Sprintf("%s\t\treturn %s, fmt.Errorf(\"unsupported native interface conversion from %s to %s\")", indent, failureExpr, actualGoType, expectedGoType)
	if !g.renderNativeInterfaceDirectCoercionSwitch(buf, indent+"\t", target, expr, actualInfo, expectedInfo, defaultLine) {
		return false
	}
	fmt.Fprintf(buf, "%s}\n", indent)
	return true
}

func (g *generator) renderNativeInterfaceDirectCoercionSwitch(buf *bytes.Buffer, indent string, target string, expr string, actualInfo *nativeInterfaceInfo, expectedInfo *nativeInterfaceInfo, defaultLine string) bool {
	if g == nil || buf == nil || target == "" || expr == "" || actualInfo == nil || expectedInfo == nil || defaultLine == "" {
		return false
	}
	fmt.Fprintf(buf, "%sswitch typed := %s.(type) {\n", indent, expr)
	for _, actualAdapter := range g.nativeInterfaceKnownAdapters(actualInfo) {
		if actualAdapter == nil || actualAdapter.AdapterType == "" || actualAdapter.GoType == "" {
			continue
		}
		expectedAdapter, ok := g.nativeInterfaceAdapterForActual(expectedInfo, actualAdapter.GoType)
		if !ok || expectedAdapter == nil || expectedAdapter.WrapHelper == "" {
			continue
		}
		fmt.Fprintf(buf, "%scase %s:\n", indent, actualAdapter.AdapterType)
		fmt.Fprintf(buf, "%s\t%s = %s(typed.Value)\n", indent, target, expectedAdapter.WrapHelper)
	}
	if actualInfo.RuntimeAdapter != "" && expectedInfo.RuntimeWrapHelper != "" {
		fmt.Fprintf(buf, "%scase %s:\n", indent, actualInfo.RuntimeAdapter)
		fmt.Fprintf(buf, "%s\t%s = %s(typed.Value)\n", indent, target, expectedInfo.RuntimeWrapHelper)
	}
	fmt.Fprintf(buf, "%sdefault:\n", indent)
	fmt.Fprintf(buf, "%s\n", defaultLine)
	fmt.Fprintf(buf, "%s}\n", indent)
	return true
}

func (g *generator) renderNativeInterfaceBoundaryHelpers(buf *bytes.Buffer, info *nativeInterfaceInfo) {
	renderedType, ok := g.renderTypeExpression(info.TypeExpr)
	if !ok {
		return
	}
	baseName, _ := typeExprBaseName(info.TypeExpr)
	fmt.Fprintf(buf, "func %s(rt *bridge.Runtime, value runtime.Value) (%s, error) {\n", info.FromRuntimeHelper, info.GoType)
	fmt.Fprintf(buf, "\tif rt == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing runtime bridge\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tbase := __able_unwrap_interface(value)\n")
	fmt.Fprintf(buf, "\t_ = base\n")
	if baseName == "Iterator" {
		fmt.Fprintf(buf, "\tif iter, ok, nilPtr := __able_runtime_iterator_value(value); ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\tif !ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil, fmt.Errorf(\"type mismatch: expected %s\")\n", info.TypeString)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\treturn %s(iter), nil\n", info.RuntimeWrapHelper)
		fmt.Fprintf(buf, "\t}\n")
	}
	for _, adapter := range g.nativeInterfaceKnownAdapters(info) {
		if adapter == nil || adapter.TypeExpr == nil {
			continue
		}
		renderedAdapterType, ok := g.renderTypeExpression(adapter.TypeExpr)
		if !ok {
			continue
		}
		fmt.Fprintf(buf, "\tif coerced, ok, err := bridge.MatchType(rt, %s, base); err != nil {\n", renderedAdapterType)
		fmt.Fprintf(buf, "\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t} else if ok {\n")
		if g.renderNativeInterfaceRuntimeToGoValueError(buf, "converted", "coerced", adapter.GoType, "\t\t") {
			fmt.Fprintf(buf, "\t\treturn %s(converted), nil\n", adapter.WrapHelper)
		} else {
			fmt.Fprintf(buf, "\t\t_ = coerced\n")
		}
		fmt.Fprintf(buf, "\t}\n")
	}
	fmt.Fprintf(buf, "\tcoerced, ok, err := bridge.MatchType(rt, %s, value)\n", renderedType)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif !ok {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"type mismatch: expected %s\")\n", info.TypeString)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn %s(coerced), nil\n", info.RuntimeWrapHelper)
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func %s(value runtime.Value) %s {\n", info.FromRuntimePanic, info.GoType)
	fmt.Fprintf(buf, "\tconverted, err := %s(__able_runtime, value)\n", info.FromRuntimeHelper)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(err)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn converted\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func %s(rt *bridge.Runtime, value %s) (runtime.Value, error) {\n", info.ToRuntimeHelper, info.GoType)
	fmt.Fprintf(buf, "\tif rt == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing runtime bridge\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn value.%s(rt)\n", info.ToRuntimeMethod)
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func %s(value %s) runtime.Value {\n", info.ToRuntimePanic, info.GoType)
	fmt.Fprintf(buf, "\tconverted, err := %s(__able_runtime, value)\n", info.ToRuntimeHelper)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(err)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn converted\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderNativeInterfaceApplyRuntimeHelper(buf *bytes.Buffer, info *nativeInterfaceInfo) {
	if info == nil {
		return
	}
	fmt.Fprintf(buf, "func %s(value %s, runtimeValue runtime.Value) error {\n", info.ApplyRuntimeHelper, info.GoType)
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tswitch typed := value.(type) {\n")
	renderedCase := false
	for _, adapter := range g.nativeInterfaceKnownAdapters(info) {
		if adapter == nil || !strings.HasPrefix(adapter.GoType, "*") || g.typeCategory(adapter.GoType) != "struct" {
			continue
		}
		baseName, ok := g.structHelperName(adapter.GoType)
		if !ok {
			baseName = strings.TrimPrefix(adapter.GoType, "*")
		}
		renderedCase = true
		fmt.Fprintf(buf, "\tcase %s:\n", adapter.AdapterType)
		fmt.Fprintf(buf, "\t\tif typed.Value == nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tconverted, err := __able_struct_%s_from(runtimeValue)\n", baseName)
		fmt.Fprintf(buf, "\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn err\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\tif converted == nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn nil\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\t*typed.Value = *converted\n")
		fmt.Fprintf(buf, "\t\treturn nil\n")
	}
	fmt.Fprintf(buf, "\tdefault:\n")
	if renderedCase {
		fmt.Fprintf(buf, "\t\treturn nil\n")
	} else {
		fmt.Fprintf(buf, "\t\t_ = typed\n")
		fmt.Fprintf(buf, "\t\t_ = runtimeValue\n")
		fmt.Fprintf(buf, "\t\treturn nil\n")
	}
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderNativeInterfaceRuntimeToGoValueError(buf *bytes.Buffer, target string, expr string, goType string, indent string) bool {
	rawVar := sanitizeIdent(target) + "_raw"
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(%s)\n", indent, target, spec.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, iface.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, callable.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	if union := g.nativeUnionInfoForGoType(goType); union != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, union.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	if helper, ok := g.nativeNullableFromRuntimeHelper(goType); ok {
		fmt.Fprintf(buf, "%s%s, err := %s(%s)\n", indent, target, helper, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		return true
	}
	switch goType {
	case "runtime.Value":
		fmt.Fprintf(buf, "%s%s := %s\n", indent, target, expr)
		return true
	case "runtime.ErrorValue":
		fmt.Fprintf(buf, "%s%s := bridge.ErrorValue(__able_runtime, %s)\n", indent, target, expr)
		return true
	case "any":
		fmt.Fprintf(buf, "%svar %s any = %s\n", indent, target, expr)
		return true
	case "struct{}":
		fmt.Fprintf(buf, "%s_ = %s\n", indent, expr)
		fmt.Fprintf(buf, "%s%s := struct{}{}\n", indent, target)
		return true
	case "bool":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsBool(%s)\n", indent, target, expr)
	case "string":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsString(%s)\n", indent, target, expr)
	case "rune":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsRune(%s)\n", indent, target, expr)
	case "float32":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsFloat(%s)\n", indent, rawVar, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := float32(%s)\n", indent, target, rawVar)
		return true
	case "float64":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsFloat(%s)\n", indent, target, expr)
	case "int":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsInt(%s, bridge.NativeIntBits)\n", indent, rawVar, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := int(%s)\n", indent, target, rawVar)
		return true
	case "uint":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsUint(%s, bridge.NativeIntBits)\n", indent, rawVar, expr)
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := uint(%s)\n", indent, target, rawVar)
		return true
	case "int8", "int16", "int32", "int64":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsInt(%s, %d)\n", indent, rawVar, expr, g.intBits(goType))
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := %s(%s)\n", indent, target, goType, rawVar)
		return true
	case "uint8", "uint16", "uint32", "uint64":
		fmt.Fprintf(buf, "%s%s, err := bridge.AsUint(%s, %d)\n", indent, rawVar, expr, g.intBits(goType))
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
		fmt.Fprintf(buf, "%s}\n", indent)
		fmt.Fprintf(buf, "%s%s := %s(%s)\n", indent, target, goType, rawVar)
		return true
	default:
		if g.typeCategory(goType) == "struct" {
			baseName, ok := g.structHelperName(goType)
			if !ok {
				baseName = strings.TrimPrefix(goType, "*")
			}
			fmt.Fprintf(buf, "%s%s, err := __able_struct_%s_from(%s)\n", indent, target, baseName, expr)
			fmt.Fprintf(buf, "%sif err != nil {\n", indent)
			fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
			fmt.Fprintf(buf, "%s}\n", indent)
			return true
		}
		fmt.Fprintf(buf, "%sreturn nil, fmt.Errorf(\"unsupported native interface conversion to %s\")\n", indent, goType)
		return false
	}
	fmt.Fprintf(buf, "%sif err != nil {\n", indent)
	fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
	fmt.Fprintf(buf, "%s}\n", indent)
	return true
}

func (g *generator) renderNativeInterfaceGoToRuntimeValueControl(buf *bytes.Buffer, target string, expr string, goType string, returnType string) {
	fmt.Fprintf(buf, "\t")
	if g.renderNativeInterfaceGoToRuntimeValue(buf, target, expr, goType, "\t", false, returnType) {
		return
	}
}

func (g *generator) renderNativeInterfaceGoToRuntimeValueError(buf *bytes.Buffer, target string, expr string, goType string) {
	g.renderNativeInterfaceGoToRuntimeValue(buf, target, expr, goType, "\t", true, "")
}

func (g *generator) renderNativeInterfaceGoToRuntimeValue(buf *bytes.Buffer, target string, expr string, goType string, indent string, returnError bool, returnType string) bool {
	failErr := func(errExpr string) {
		if returnError {
			fmt.Fprintf(buf, "%sif err != nil {\n", indent)
			fmt.Fprintf(buf, "%s\treturn nil, err\n", indent)
			fmt.Fprintf(buf, "%s}\n", indent)
			return
		}
		zeroExpr, ok := g.zeroValueExpr(returnType)
		if !ok {
			fmt.Fprintf(buf, "%svar zero %s\n", indent, returnType)
			zeroExpr = "zero"
		}
		fmt.Fprintf(buf, "%sif err != nil {\n", indent)
		fmt.Fprintf(buf, "%s\treturn %s, __able_control_from_error(%s)\n", indent, zeroExpr, errExpr)
		fmt.Fprintf(buf, "%s}\n", indent)
	}
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, spec.ToRuntimeHelper, expr)
		failErr("err")
		return true
	}
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, iface.ToRuntimeHelper, expr)
		failErr("err")
		return true
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, callable.ToRuntimeHelper, expr)
		failErr("err")
		return true
	}
	if union := g.nativeUnionInfoForGoType(goType); union != nil {
		fmt.Fprintf(buf, "%s%s, err := %s(__able_runtime, %s)\n", indent, target, union.ToRuntimeHelper, expr)
		failErr("err")
		return true
	}
	if helper, ok := g.nativeNullableToRuntimeHelper(goType); ok {
		fmt.Fprintf(buf, "%s%s := %s(%s)\n", indent, target, helper, expr)
		return true
	}
	switch goType {
	case "runtime.Value":
		fmt.Fprintf(buf, "%s%s := %s\n", indent, target, expr)
		return true
	case "runtime.ErrorValue":
		fmt.Fprintf(buf, "%s%s := %s\n", indent, target, expr)
		return true
	case "any":
		fmt.Fprintf(buf, "%s%s := __able_any_to_value(%s)\n", indent, target, expr)
		return true
	case "struct{}":
		fmt.Fprintf(buf, "%s_ = %s\n", indent, expr)
		fmt.Fprintf(buf, "%s%s := runtime.VoidValue{}\n", indent, target)
		return true
	case "bool":
		fmt.Fprintf(buf, "%s%s := bridge.ToBool(%s)\n", indent, target, expr)
		return true
	case "string":
		fmt.Fprintf(buf, "%s%s := bridge.ToString(%s)\n", indent, target, expr)
		return true
	case "rune":
		fmt.Fprintf(buf, "%s%s := bridge.ToRune(%s)\n", indent, target, expr)
		return true
	case "float32":
		fmt.Fprintf(buf, "%s%s := bridge.ToFloat32(%s)\n", indent, target, expr)
		return true
	case "float64":
		fmt.Fprintf(buf, "%s%s := bridge.ToFloat64(%s)\n", indent, target, expr)
		return true
	case "int":
		fmt.Fprintf(buf, "%s%s := bridge.ToInt(int64(%s), runtime.IntegerType(\"isize\"))\n", indent, target, expr)
		return true
	case "uint":
		fmt.Fprintf(buf, "%s%s := bridge.ToUint(uint64(%s), runtime.IntegerType(\"usize\"))\n", indent, target, expr)
		return true
	case "int8", "int16", "int32", "int64":
		suffix, _ := g.integerTypeSuffix(goType)
		fmt.Fprintf(buf, "%s%s := bridge.ToInt(int64(%s), runtime.IntegerType(%q))\n", indent, target, expr, suffix)
		return true
	case "uint8", "uint16", "uint32", "uint64":
		suffix, _ := g.integerTypeSuffix(goType)
		fmt.Fprintf(buf, "%s%s := bridge.ToUint(uint64(%s), runtime.IntegerType(%q))\n", indent, target, expr, suffix)
		return true
	}
	if g.typeCategory(goType) == "struct" {
		baseName, ok := g.structHelperName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		fmt.Fprintf(buf, "%s%s, err := __able_struct_%s_to(__able_runtime, %s)\n", indent, target, baseName, expr)
		failErr("err")
		return true
	}
	fmt.Fprintf(buf, "%svar %s runtime.Value\n", indent, target)
	fmt.Fprintf(buf, "%s_ = %s\n", indent, target)
	if returnError {
		fmt.Fprintf(buf, "%sreturn nil, fmt.Errorf(\"unsupported native interface conversion from %s\")\n", indent, goType)
	} else {
		zeroExpr, ok := g.zeroValueExpr(returnType)
		if !ok {
			fmt.Fprintf(buf, "%svar zero %s\n", indent, returnType)
			zeroExpr = "zero"
		}
		fmt.Fprintf(buf, "%sreturn %s, __able_control_from_error(fmt.Errorf(\"unsupported native interface conversion from %s\"))\n", indent, zeroExpr, goType)
	}
	return false
}

func (g *generator) renderNativeInterfaceRuntimeToGoValueControl(buf *bytes.Buffer, target string, expr string, goType string, returnType string) bool {
	zeroExpr, zeroOK := g.zeroValueExpr(returnType)
	if !zeroOK {
		fmt.Fprintf(buf, "\tvar zero %s\n", returnType)
		zeroExpr = "zero"
	}
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		fmt.Fprintf(buf, "\t%s, err := %s(%s)\n", target, spec.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	}
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		fmt.Fprintf(buf, "\t%s, err := %s(__able_runtime, %s)\n", target, iface.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		fmt.Fprintf(buf, "\t%s, err := %s(__able_runtime, %s)\n", target, callable.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	}
	if union := g.nativeUnionInfoForGoType(goType); union != nil {
		fmt.Fprintf(buf, "\t%s, err := %s(__able_runtime, %s)\n", target, union.FromRuntimeHelper, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	}
	if helper, ok := g.nativeNullableFromRuntimeHelper(goType); ok {
		fmt.Fprintf(buf, "\t%s, err := %s(%s)\n", target, helper, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	}
	switch goType {
	case "runtime.Value":
		fmt.Fprintf(buf, "\t%s := %s\n", target, expr)
		return true
	case "runtime.ErrorValue":
		fmt.Fprintf(buf, "\t%s := bridge.ErrorValue(__able_runtime, %s)\n", target, expr)
		return true
	case "any":
		fmt.Fprintf(buf, "\tvar %s any = %s\n", target, expr)
		return true
	case "struct{}":
		fmt.Fprintf(buf, "\tif errVal, ok, nilPtr := __able_runtime_error_value(%s); ok || nilPtr {\n", expr)
		fmt.Fprintf(buf, "\t\tif ok {\n")
		fmt.Fprintf(buf, "\t\t\treturn %s, __able_raise_control(nil, errVal)\n", zeroExpr)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t_ = %s\n", expr)
		fmt.Fprintf(buf, "\t%s := struct{}{}\n", target)
		return true
	case "bool":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsBool(%s)\n", target, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	case "string":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsString(%s)\n", target, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	case "rune":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsRune(%s)\n", target, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	case "float32":
		fmt.Fprintf(buf, "\traw, err := bridge.AsFloat(%s)\n", expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s := float32(raw)\n", target)
		return true
	case "float64":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsFloat(%s)\n", target, expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		return true
	case "int":
		fmt.Fprintf(buf, "\traw, err := bridge.AsInt(%s, bridge.NativeIntBits)\n", expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s := int(raw)\n", target)
		return true
	case "uint":
		fmt.Fprintf(buf, "\traw, err := bridge.AsUint(%s, bridge.NativeIntBits)\n", expr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s := uint(raw)\n", target)
		return true
	case "int8", "int16", "int32", "int64":
		fmt.Fprintf(buf, "\traw, err := bridge.AsInt(%s, %d)\n", expr, g.intBits(goType))
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s := %s(raw)\n", target, goType)
		return true
	case "uint8", "uint16", "uint32", "uint64":
		fmt.Fprintf(buf, "\traw, err := bridge.AsUint(%s, %d)\n", expr, g.intBits(goType))
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s := %s(raw)\n", target, goType)
		return true
	default:
		if g.typeCategory(goType) == "struct" {
			baseName, ok := g.structHelperName(goType)
			if !ok {
				baseName = strings.TrimPrefix(goType, "*")
			}
			fmt.Fprintf(buf, "\t%s, err := __able_struct_%s_from(%s)\n", target, baseName, expr)
			fmt.Fprintf(buf, "\tif err != nil {\n")
			fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
			fmt.Fprintf(buf, "\t}\n")
			return true
		}
		fmt.Fprintf(buf, "\tvar %s %s\n", target, goType)
		fmt.Fprintf(buf, "\t_ = %s\n", target)
		fmt.Fprintf(buf, "\t_ = %s\n", expr)
		fmt.Fprintf(buf, "\treturn %s, __able_control_from_error(fmt.Errorf(\"unsupported native interface conversion to %s\"))\n", zeroExpr, goType)
		return false
	}
}
