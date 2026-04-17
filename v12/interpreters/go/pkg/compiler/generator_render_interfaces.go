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
		if g.renderNativeInterfaceScaffoldingForInfo(buf, info) {
			progress = true
		}
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
		_ = g.nativeInterfaceBoundarySiblingInfos(info)
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
	for _, key := range g.sortedNativeInterfaceKeys() {
		info := g.nativeInterfaces[key]
		if info == nil {
			continue
		}
		if info.AdapterVersion != g.nativeInterfaceAdapterVersion {
			g.refreshNativeInterfaceAdapters(info)
		}
		g.renderNativeInterfaceBoundaryDirectAdapters(buf, info)
		g.renderNativeInterfaceBoundaryHelpers(buf, info)
	}
}

func (g *generator) renderNativeInterfaceBoundaryDirectAdapters(buf *bytes.Buffer, info *nativeInterfaceInfo) {
	if g == nil || buf == nil || info == nil {
		return
	}
	for _, actualKey := range g.sortedNativeInterfaceKeys() {
		actualInfo := g.nativeInterfaces[actualKey]
		if actualInfo == nil || actualInfo.Key == info.Key || actualInfo.GoType == "" {
			continue
		}
		if !g.nativeInterfaceDirectAdapterPossible(actualInfo, info) {
			continue
		}
		adapter, ok := g.nativeInterfaceAdapterForActual(info, actualInfo.GoType)
		if !ok || adapter == nil || adapter.GoType == "" {
			continue
		}
		renderKey := info.Key + "::" + adapter.GoType
		if _, rendered := g.nativeInterfaceRenderedAdapters[renderKey]; rendered {
			continue
		}
		g.renderNativeInterfaceConcreteAdapter(buf, info, adapter)
		g.nativeInterfaceRenderedAdapters[renderKey] = struct{}{}
	}
}

func (g *generator) renderPendingNativeInterfaceConcreteAdapters(buf *bytes.Buffer) bool {
	if g == nil || buf == nil {
		return false
	}
	progress := false
	for _, key := range g.sortedNativeInterfaceKeys() {
		info := g.nativeInterfaces[key]
		if info == nil {
			continue
		}
		if g.renderNativeInterfaceConcreteAdaptersForInfo(buf, info) {
			progress = true
		}
	}
	return progress
}

func (g *generator) renderNativeInterfaceScaffoldingForInfo(buf *bytes.Buffer, info *nativeInterfaceInfo) bool {
	if g == nil || buf == nil || info == nil {
		return false
	}
	if _, ok := g.nativeInterfaceRenderedInfos[info.Key]; ok {
		return false
	}
	g.renderNativeInterfaceType(buf, info)
	g.renderNativeInterfaceRuntimeIteratorAdapter(buf, info)
	g.renderNativeInterfaceRuntimeAdapter(buf, info)
	g.nativeInterfaceRenderedInfos[info.Key] = struct{}{}
	return true
}

func (g *generator) renderNativeInterfaceConcreteAdaptersForInfo(buf *bytes.Buffer, info *nativeInterfaceInfo) bool {
	if g == nil || buf == nil || info == nil {
		return false
	}
	progress := g.renderNativeInterfaceScaffoldingForInfo(buf, info)
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
	if actualInfo := g.nativeInterfaceInfoForGoType(adapter.GoType); actualInfo != nil && actualInfo.GoType == adapter.GoType && actualInfo.GoType != info.GoType {
		for _, method := range info.Methods {
			if method == nil {
				continue
			}
			actualMethod := nativeInterfaceMethodNamed(actualInfo, method.Name)
			if actualMethod == nil {
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
			fmt.Fprintf(buf, "\tif w.Value == nil {\n")
			fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(fmt.Errorf(\"missing interface value\"))\n", zeroExpr)
			fmt.Fprintf(buf, "\t}\n")
			args := make([]string, 0, len(method.ParamGoTypes))
			for idx, paramType := range method.ParamGoTypes {
				argExpr := fmt.Sprintf("arg%d", idx)
				actualParamType := actualMethod.ParamGoTypes[idx]
				if actualParamType != paramType {
					runtimeArg := fmt.Sprintf("__able_arg%d_value", idx)
					g.renderNativeInterfaceGoToRuntimeValueControl(buf, runtimeArg, argExpr, paramType, method.ReturnGoType)
					if actualParamType != "runtime.Value" {
						convertedArg := fmt.Sprintf("__able_arg%d_converted", idx)
						g.renderNativeInterfaceRuntimeToGoValueControl(buf, convertedArg, runtimeArg, actualParamType, method.ReturnGoType)
						argExpr = convertedArg
					} else {
						argExpr = runtimeArg
					}
				}
				args = append(args, argExpr)
			}
			fmt.Fprintf(buf, "\tresult, control := w.Value.%s(%s)\n", actualMethod.GoName, strings.Join(args, ", "))
			fmt.Fprintf(buf, "\tif control != nil {\n")
			fmt.Fprintf(buf, "\t\treturn %s, control\n", zeroExpr)
			fmt.Fprintf(buf, "\t}\n")
			if actualMethod.ReturnGoType == method.ReturnGoType {
				fmt.Fprintf(buf, "\treturn result, nil\n")
				fmt.Fprintf(buf, "}\n\n")
				continue
			}
			if g.renderNativeInterfaceDirectCoercionControl(buf, "converted", "result", actualMethod.ReturnGoType, method.ReturnGoType, method.ReturnGoType) {
				fmt.Fprintf(buf, "\treturn converted, nil\n")
				fmt.Fprintf(buf, "}\n\n")
				continue
			}
			runtimeResult := "__able_result_value"
			g.renderNativeInterfaceGoToRuntimeValueControl(buf, runtimeResult, "result", actualMethod.ReturnGoType, method.ReturnGoType)
			if g.renderNativeInterfaceRuntimeToGoValueControl(buf, "converted", runtimeResult, method.ReturnGoType, method.ReturnGoType) {
				fmt.Fprintf(buf, "\treturn converted, nil\n")
				fmt.Fprintf(buf, "}\n\n")
				continue
			}
			fmt.Fprintf(buf, "\treturn result, nil\n")
			fmt.Fprintf(buf, "}\n\n")
		}
		return
	}
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
