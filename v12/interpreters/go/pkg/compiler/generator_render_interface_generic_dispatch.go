package compiler

import (
	"bytes"
	"fmt"
	"strings"
)

func (g *generator) renderNativeInterfaceGenericDispatchHelpers(buf *bytes.Buffer) {
	if g == nil || buf == nil || len(g.nativeInterfaceGenericDispatches) == 0 {
		return
	}
	for _, key := range g.sortedNativeInterfaceGenericDispatchKeys() {
		info := g.nativeInterfaceGenericDispatches[key]
		if info == nil {
			continue
		}
		g.renderNativeInterfaceGenericDispatchHelper(buf, info)
	}
}

func (g *generator) renderNativeInterfaceGenericDispatchHelper(buf *bytes.Buffer, dispatch *nativeInterfaceGenericDispatchInfo) {
	if g == nil || buf == nil || dispatch == nil || dispatch.Method == nil {
		return
	}
	iface := g.nativeInterfaces[dispatch.InterfaceKey]
	if iface == nil {
		return
	}
	zeroExpr, ok := g.zeroValueExpr(dispatch.ReturnGoType)
	if !ok {
		return
	}
	fmt.Fprintf(buf, "func __able_compiled_%s(receiver %s", dispatch.GoName, dispatch.InterfaceGoType)
	for idx, paramGoType := range dispatch.ParamGoTypes {
		fmt.Fprintf(buf, ", arg%d %s", idx, paramGoType)
	}
	fmt.Fprintf(buf, ", call *ast.FunctionCall) (%s, *__ableControl) {\n", dispatch.ReturnGoType)
	if envVar, ok := g.packageEnvVar(dispatch.Package); ok {
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "__able_runtime", envVar, "")
	}
	fmt.Fprintf(buf, "\tif receiver == nil {\n")
	fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(fmt.Errorf(\"missing interface value\"))\n", zeroExpr)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tswitch typed := receiver.(type) {\n")
	renderedCase := false
	for _, dispatchCase := range dispatch.Cases {
		if dispatchCase.AdapterType == "" || dispatchCase.GoType == "" || dispatchCase.Impl == nil || dispatchCase.Impl.Info == nil {
			continue
		}
		impl := dispatchCase.Impl
		renderedCase = true
		fmt.Fprintf(buf, "\tcase %s:\n", dispatchCase.AdapterType)
		args := make([]string, 0, len(dispatch.ParamGoTypes)+1)
		args = append(args, "typed.Value")
		for idx, paramGoType := range dispatch.ParamGoTypes {
			argExpr := fmt.Sprintf("arg%d", idx)
			implParamType := paramGoType
			if idx < len(impl.CompiledParamGoTypes) && impl.CompiledParamGoTypes[idx] != "" {
				implParamType = impl.CompiledParamGoTypes[idx]
			}
			if implParamType != paramGoType {
				runtimeArg := fmt.Sprintf("__able_generic_arg%d_value", idx)
				g.renderNativeInterfaceGoToRuntimeValueControl(buf, runtimeArg, argExpr, paramGoType, dispatch.ReturnGoType)
				if implParamType != "runtime.Value" {
					convertedArg := fmt.Sprintf("__able_generic_arg%d_converted", idx)
					g.renderNativeInterfaceRuntimeToGoValueControl(buf, convertedArg, runtimeArg, implParamType, dispatch.ReturnGoType)
					argExpr = convertedArg
				} else {
					argExpr = runtimeArg
				}
			}
			args = append(args, argExpr)
		}
		fmt.Fprintf(buf, "\t\t__able_push_call_frame(call)\n")
		fmt.Fprintf(buf, "\t\tresult, control := __able_compiled_%s(%s)\n", impl.Info.GoName, strings.Join(args, ", "))
		fmt.Fprintf(buf, "\t\t__able_pop_call_frame()\n")
		fmt.Fprintf(buf, "\t\tif control != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn %s, control\n", zeroExpr)
		fmt.Fprintf(buf, "\t\t}\n")
		if impl.CompiledReturnGoType == dispatch.ReturnGoType {
			fmt.Fprintf(buf, "\t\treturn result, nil\n")
			continue
		}
		if g.renderNativeInterfaceDirectCoercionControl(buf, "converted", "result", impl.CompiledReturnGoType, dispatch.ReturnGoType, dispatch.ReturnGoType) {
			fmt.Fprintf(buf, "\t\treturn converted, nil\n")
			continue
		}
		runtimeResult := "__able_generic_result_value"
		g.renderNativeInterfaceGoToRuntimeValueControl(buf, runtimeResult, "result", impl.CompiledReturnGoType, dispatch.ReturnGoType)
		if g.renderNativeInterfaceRuntimeToGoValueControl(buf, "converted", runtimeResult, dispatch.ReturnGoType, dispatch.ReturnGoType) {
			fmt.Fprintf(buf, "\t\treturn converted, nil\n")
			continue
		}
		fmt.Fprintf(buf, "\t\treturn result, nil\n")
	}
	if iface.RuntimeIteratorAdapter != "" && dispatch.RuntimeDefault != nil {
		renderedCase = true
		fmt.Fprintf(buf, "\tcase %s:\n", iface.RuntimeIteratorAdapter)
		args := make([]string, 0, len(dispatch.ParamGoTypes)+1)
		args = append(args, "typed")
		for idx := range dispatch.ParamGoTypes {
			args = append(args, fmt.Sprintf("arg%d", idx))
		}
		fmt.Fprintf(buf, "\t\t__able_push_call_frame(call)\n")
		fmt.Fprintf(buf, "\t\tresult, control := __able_compiled_%s(%s)\n", dispatch.RuntimeDefault.GoName, strings.Join(args, ", "))
		fmt.Fprintf(buf, "\t\t__able_pop_call_frame()\n")
		fmt.Fprintf(buf, "\t\tif control != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn %s, control\n", zeroExpr)
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t\treturn result, nil\n")
	} else if iface.RuntimeIteratorAdapter != "" && dispatch.MonoArrayCollect != nil {
		renderedCase = true
		fmt.Fprintf(buf, "\tcase %s:\n", iface.RuntimeIteratorAdapter)
		fmt.Fprintf(buf, "\t\treturn __able_compiled_%s(typed)\n", dispatch.MonoArrayCollect.GoName)
	}
	if iface.RuntimeAdapter != "" {
		renderedCase = true
		fmt.Fprintf(buf, "\tcase %s:\n", iface.RuntimeAdapter)
		if iface.RuntimeIteratorAdapter != "" && dispatch.RuntimeDefault != nil {
			fmt.Fprintf(buf, "\t\tif iter, ok, nilPtr := __able_runtime_iterator_value(typed.Value); ok || nilPtr {\n")
			fmt.Fprintf(buf, "\t\t\tif !ok || nilPtr {\n")
			fmt.Fprintf(buf, "\t\t\t\treturn %s, __able_control_from_error(fmt.Errorf(\"missing interface value\"))\n", zeroExpr)
			fmt.Fprintf(buf, "\t\t\t}\n")
			fmt.Fprintf(buf, "\t\t\treturn __able_compiled_%s(%s(iter), ", dispatch.GoName, iface.RuntimeWrapHelper)
			args := make([]string, 0, len(dispatch.ParamGoTypes)+1)
			for idx := range dispatch.ParamGoTypes {
				args = append(args, fmt.Sprintf("arg%d", idx))
			}
			args = append(args, "call")
			fmt.Fprintf(buf, "%s)\n", strings.Join(args, ", "))
			fmt.Fprintf(buf, "\t\t}\n")
		} else if dispatch.MonoArrayCollect != nil {
			fmt.Fprintf(buf, "\t\treturn __able_compiled_%s(typed)\n", dispatch.MonoArrayCollect.GoName)
		} else {
			fmt.Fprintf(buf, "\t\treturn __able_compiled_%s_runtime_adapter(typed", dispatch.GoName)
			for idx := range dispatch.ParamGoTypes {
				fmt.Fprintf(buf, ", arg%d", idx)
			}
			fmt.Fprintf(buf, ", call)\n")
		}
	}
	fmt.Fprintf(buf, "\tdefault:\n")
	if renderedCase {
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(fmt.Errorf(\"unsupported native generic interface dispatch for %s.%s\"))\n", zeroExpr, dispatch.Method.InterfaceName, dispatch.Method.Name)
	} else {
		fmt.Fprintf(buf, "\t\t_ = typed\n")
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(fmt.Errorf(\"unsupported native generic interface dispatch for %s.%s\"))\n", zeroExpr, dispatch.Method.InterfaceName, dispatch.Method.Name)
	}
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")
	if iface.RuntimeAdapter != "" && dispatch.MonoArrayCollect == nil {
		g.renderNativeInterfaceGenericDispatchRuntimeAdapterHelper(buf, iface, dispatch, zeroExpr)
	}
}

func (g *generator) renderNativeInterfaceGenericDispatchRuntimeAdapterHelper(buf *bytes.Buffer, iface *nativeInterfaceInfo, dispatch *nativeInterfaceGenericDispatchInfo, zeroExpr string) {
	if g == nil || buf == nil || iface == nil || dispatch == nil || iface.RuntimeAdapter == "" {
		return
	}
	fmt.Fprintf(buf, "func __able_compiled_%s_runtime_adapter(typed %s", dispatch.GoName, iface.RuntimeAdapter)
	for idx, paramGoType := range dispatch.ParamGoTypes {
		fmt.Fprintf(buf, ", arg%d %s", idx, paramGoType)
	}
	fmt.Fprintf(buf, ", call *ast.FunctionCall) (%s, *__ableControl) {\n", dispatch.ReturnGoType)
	argList := "nil"
	if len(dispatch.ParamGoTypes) > 0 {
		fmt.Fprintf(buf, "\targs := make([]runtime.Value, 0, %d)\n", len(dispatch.ParamGoTypes))
		for idx, paramGoType := range dispatch.ParamGoTypes {
			target := fmt.Sprintf("arg%dValue", idx)
			g.renderNativeInterfaceGoToRuntimeValueControl(buf, target, fmt.Sprintf("arg%d", idx), paramGoType, dispatch.ReturnGoType)
			fmt.Fprintf(buf, "\targs = append(args, %s)\n", target)
		}
		argList = "args"
	}
	fmt.Fprintf(buf, "\tresult, control := __able_method_call_node(typed.Value, %q, %s, call)\n", dispatch.Method.Name, argList)
	for idx, paramGoType := range dispatch.ParamGoTypes {
		argIface := g.nativeInterfaceInfoForGoType(paramGoType)
		if argIface == nil || argIface.ApplyRuntimeHelper == "" {
			continue
		}
		fmt.Fprintf(buf, "\tif err := %s(arg%d, arg%dValue); err != nil {\n", argIface.ApplyRuntimeHelper, idx, idx)
		fmt.Fprintf(buf, "\t\treturn %s, __able_control_from_error(err)\n", zeroExpr)
		fmt.Fprintf(buf, "\t}\n")
	}
	fmt.Fprintf(buf, "\tif control != nil {\n")
	fmt.Fprintf(buf, "\t\treturn %s, control\n", zeroExpr)
	fmt.Fprintf(buf, "\t}\n")
	switch {
	case g.isVoidType(dispatch.ReturnGoType):
		fmt.Fprintf(buf, "\t_ = result\n")
		fmt.Fprintf(buf, "\treturn struct{}{}, nil\n")
	case dispatch.ReturnGoType == "runtime.Value":
		fmt.Fprintf(buf, "\treturn result, nil\n")
	case dispatch.ReturnGoType == "any":
		fmt.Fprintf(buf, "\tvar converted any = result\n")
		fmt.Fprintf(buf, "\treturn converted, nil\n")
	default:
		if g.renderNativeInterfaceRuntimeToGoValueControl(buf, "converted", "result", dispatch.ReturnGoType, dispatch.ReturnGoType) {
			fmt.Fprintf(buf, "\treturn converted, nil\n")
		} else {
			fmt.Fprintf(buf, "\treturn %s, __able_control_from_error(fmt.Errorf(\"unsupported native generic interface conversion to %s\"))\n", zeroExpr, dispatch.ReturnGoType)
		}
	}
	fmt.Fprintf(buf, "}\n\n")
}
