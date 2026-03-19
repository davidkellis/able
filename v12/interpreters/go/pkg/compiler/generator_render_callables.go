package compiler

import (
	"bytes"
	"fmt"
)

func (g *generator) renderNativeCallables(buf *bytes.Buffer) {
	if g == nil || buf == nil || len(g.nativeCallables) == 0 {
		return
	}
	for _, key := range g.sortedNativeCallableKeys() {
		info := g.nativeCallables[key]
		if info == nil {
			continue
		}
		g.renderNativeCallableType(buf, info)
		g.renderNativeCallableBoundaryHelpers(buf, info)
	}
}

func (g *generator) renderNativeCallableType(buf *bytes.Buffer, info *nativeCallableInfo) {
	fmt.Fprintf(buf, "type %s func(", info.GoType)
	for idx, paramType := range info.ParamGoTypes {
		if idx > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "arg%d %s", idx, paramType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl)\n\n", info.ReturnGoType)
}

func (g *generator) renderNativeCallableBoundaryHelpers(buf *bytes.Buffer, info *nativeCallableInfo) {
	fmt.Fprintf(buf, "func %s(rt *bridge.Runtime, value runtime.Value) (%s, error) {\n", info.FromRuntimeHelper, info.GoType)
	fmt.Fprintf(buf, "\tif rt == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing runtime bridge\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif value == nil || __able_is_nil(value) {\n")
	fmt.Fprintf(buf, "\t\treturn nil, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn func(")
	for idx, paramType := range info.ParamGoTypes {
		if idx > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "arg%d %s", idx, paramType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", info.ReturnGoType)
	zeroExpr, zeroOK := g.zeroValueExpr(info.ReturnGoType)
	if !zeroOK {
		fmt.Fprintf(buf, "\t\tvar zero %s\n", info.ReturnGoType)
		zeroExpr = "zero"
	}
	if len(info.ParamGoTypes) > 0 {
		fmt.Fprintf(buf, "\t\targs := make([]runtime.Value, 0, %d)\n", len(info.ParamGoTypes))
		for idx, paramType := range info.ParamGoTypes {
			target := fmt.Sprintf("arg%dValue", idx)
			g.renderNativeInterfaceGoToRuntimeValueControl(buf, target, fmt.Sprintf("arg%d", idx), paramType, info.ReturnGoType)
			fmt.Fprintf(buf, "\t\targs = append(args, %s)\n", target)
		}
	} else {
		fmt.Fprintf(buf, "\t\tvar args []runtime.Value\n")
	}
	fmt.Fprintf(buf, "\t\tresult, control := __able_call_value(value, args, nil)\n")
	fmt.Fprintf(buf, "\t\tif control != nil {\n")
	fmt.Fprintf(buf, "\t\t\treturn %s, control\n", zeroExpr)
	fmt.Fprintf(buf, "\t\t}\n")
	if g.renderNativeInterfaceRuntimeToGoValueControl(buf, "converted", "result", info.ReturnGoType, info.ReturnGoType) {
		fmt.Fprintf(buf, "\t\treturn converted, nil\n")
	}
	fmt.Fprintf(buf, "\t}, nil\n")
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
	fmt.Fprintf(buf, "\treturn runtime.NativeFunctionValue{Name: %q, Arity: %d, Impl: func(callCtx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n", info.TypeString, len(info.ParamGoTypes))
	fmt.Fprintf(buf, "\t\tif __able_runtime != nil && callCtx != nil && callCtx.Env != nil { prevEnv := __able_runtime.SwapEnv(callCtx.Env); defer __able_runtime.SwapEnv(prevEnv) }\n")
	fmt.Fprintf(buf, "\t\tif len(args) != %d {\n", len(info.ParamGoTypes))
	fmt.Fprintf(buf, "\t\t\treturn nil, fmt.Errorf(\"callable expects %d arguments, got %%d\", len(args))\n", len(info.ParamGoTypes))
	fmt.Fprintf(buf, "\t\t}\n")
	for idx, paramType := range info.ParamGoTypes {
		valueVar := fmt.Sprintf("args[%d]", idx)
		target := fmt.Sprintf("arg%d", idx)
		g.renderNativeInterfaceRuntimeToGoValueError(buf, target, valueVar, paramType, "\t\t")
	}
	argList := ""
	for idx := range info.ParamGoTypes {
		if idx > 0 {
			argList += ", "
		}
		argList += fmt.Sprintf("arg%d", idx)
	}
	fmt.Fprintf(buf, "\t\tresult, control := value(%s)\n", argList)
	fmt.Fprintf(buf, "\t\tif control != nil {\n")
	fmt.Fprintf(buf, "\t\t\treturn nil, __able_control_to_error(__able_runtime, callCtx, control)\n")
	fmt.Fprintf(buf, "\t\t}\n")
	g.renderNativeInterfaceGoToRuntimeValueError(buf, "converted", "result", info.ReturnGoType)
	fmt.Fprintf(buf, "\t\treturn converted, nil\n")
	fmt.Fprintf(buf, "\t}}, nil\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func %s(value %s) runtime.Value {\n", info.ToRuntimePanic, info.GoType)
	fmt.Fprintf(buf, "\tconverted, err := %s(__able_runtime, value)\n", info.ToRuntimeHelper)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(err)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn converted\n")
	fmt.Fprintf(buf, "}\n\n")
}
