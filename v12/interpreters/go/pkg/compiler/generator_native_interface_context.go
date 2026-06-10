package compiler

import (
	"bytes"
	"fmt"
	"strings"
)

func nativeInterfaceContextMethodName(method *nativeInterfaceMethod) string {
	if method == nil || method.GoName == "" {
		return ""
	}
	return "__able_ctx_" + method.GoName
}

func (g *generator) renderNativeInterfaceMethodSignature(
	buf *bytes.Buffer,
	receiverType string,
	method *nativeInterfaceMethod,
	contextAware bool,
) {
	methodName := method.GoName
	if contextAware {
		methodName = nativeInterfaceContextMethodName(method)
	}
	fmt.Fprintf(buf, "func (w %s) %s(", receiverType, methodName)
	for idx, paramType := range method.ParamGoTypes {
		if idx > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "arg%d %s", idx, paramType)
	}
	if contextAware {
		if len(method.ParamGoTypes) > 0 {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprintf(buf, "__able_exec_ctx %s", executionContextType)
	}
	fmt.Fprintf(buf, ") (%s, *__ableControl) {\n", method.ReturnGoType)
}

func nativeInterfaceMethodArgNames(method *nativeInterfaceMethod) []string {
	if method == nil {
		return nil
	}
	args := make([]string, 0, len(method.ParamGoTypes))
	for idx := range method.ParamGoTypes {
		args = append(args, fmt.Sprintf("arg%d", idx))
	}
	return args
}

func (g *generator) renderNativeInterfaceCompatibilityDelegator(
	buf *bytes.Buffer,
	receiverType string,
	method *nativeInterfaceMethod,
) {
	if !g.executionContextsEnabled() {
		return
	}
	g.renderNativeInterfaceMethodSignature(buf, receiverType, method, false)
	args := nativeInterfaceMethodArgNames(method)
	args = append(args, "__able_context_from_args()")
	fmt.Fprintf(buf, "\treturn w.%s(%s)\n", nativeInterfaceContextMethodName(method), strings.Join(args, ", "))
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderNativeInterfaceCompiledContextCall(
	buf *bytes.Buffer,
	indent string,
	resultName string,
	controlName string,
	info *functionInfo,
	args []string,
) {
	if !g.executionContextsEnabled() {
		fmt.Fprintf(
			buf,
			"%s%s, %s := %s(%s)\n",
			indent,
			resultName,
			controlName,
			g.compiledCallTargetName("", info),
			strings.Join(args, ", "),
		)
		return
	}

	fmt.Fprintf(buf, "%svar %s %s\n", indent, resultName, info.ReturnType)
	fmt.Fprintf(buf, "%svar %s *__ableControl\n", indent, controlName)
	independent := g.environmentIndependent[info] || g.compiledFunctionEnvironmentIndependentByGoName(info.GoName)
	contextArgs := append(append([]string{}, args...), "__able_exec_ctx")
	bodyCall := fmt.Sprintf("%s(%s)", g.compiledContextBodyName(info), strings.Join(contextArgs, ", "))
	entryCall := fmt.Sprintf("%s(%s)", g.compiledContextEntryName(info), strings.Join(contextArgs, ", "))
	if independent {
		fmt.Fprintf(buf, "%s%s, %s = %s\n", indent, resultName, controlName, bodyCall)
		return
	}
	if envVar, ok := g.packageEnvVar(info.Package); ok && envVar != "" {
		fmt.Fprintf(buf, "%sif __able_exec_ctx != nil && __able_exec_ctx.packageEnv == %s {\n", indent, envVar)
		fmt.Fprintf(buf, "%s\t%s, %s = %s\n", indent, resultName, controlName, bodyCall)
		fmt.Fprintf(buf, "%s} else {\n", indent)
		fmt.Fprintf(buf, "%s\t%s, %s = %s\n", indent, resultName, controlName, entryCall)
		fmt.Fprintf(buf, "%s}\n", indent)
		return
	}
	fmt.Fprintf(buf, "%s%s, %s = %s\n", indent, resultName, controlName, bodyCall)
}
