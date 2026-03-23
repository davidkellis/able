package compiler

import (
	"bytes"
	"fmt"
	"sort"
)

import "able/interpreter-go/pkg/ast"

type iteratorCollectMonoArrayInfo struct {
	Key               string
	GoName            string
	Package           string
	ReceiverType      string
	ReturnType        string
	ElemGoType        string
	IteratorEndUnwrap string
	ValueUnwrap       string
}

func (g *generator) compileStaticIteratorCollectMonoArrayCall(ctx *compileContext, call *ast.FunctionCall, expected string, receiverExpr string, receiverType string, method *nativeInterfaceGenericMethod, returnGoType string, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Name != "collect" || receiverExpr == "" || receiverType == "" {
		return nil, "", "", false
	}
	info, ok := g.ensureIteratorCollectMonoArrayInfo(method, receiverType, returnGoType)
	if !ok || info == nil {
		return nil, "", "", false
	}
	methodInfo := &methodInfo{
		MethodName:  method.Name,
		ExpectsSelf: true,
		Info: &functionInfo{
			GoName:      info.GoName,
			Package:     info.Package,
			Params:      []paramInfo{{GoName: "self", GoType: info.ReceiverType}},
			ReturnType:  info.ReturnType,
			Compileable: true,
		},
	}
	return g.lowerResolvedMethodDispatch(ctx, call, expected, methodInfo, receiverExpr, receiverType, callNode)
}

func (g *generator) ensureIteratorCollectMonoArrayInfo(method *nativeInterfaceGenericMethod, receiverType string, returnGoType string) (*iteratorCollectMonoArrayInfo, bool) {
	if g == nil || method == nil || method.Name != "collect" || receiverType == "" || returnGoType == "" {
		return nil, false
	}
	spec, ok := g.monoArraySpecForGoType(returnGoType)
	if !ok || spec == nil {
		return nil, false
	}
	nextMethod, ok := g.nativeInterfaceMethodForGoType(receiverType, "next")
	if !ok || nextMethod == nil || nextMethod.ReturnGoType == "" {
		return nil, false
	}
	nextUnion := g.nativeUnionInfoForGoType(nextMethod.ReturnGoType)
	if nextUnion == nil {
		return nil, false
	}
	iterEndGoType, ok := g.lowerCarrierTypeInPackage(method.InterfacePackage, ast.Ty("IteratorEnd"))
	if !ok || iterEndGoType == "" {
		return nil, false
	}
	iterEndMember, ok := g.nativeUnionMember(nextUnion, iterEndGoType)
	if !ok || iterEndMember == nil {
		return nil, false
	}
	valueMember, ok := g.nativeUnionMember(nextUnion, spec.ElemGoType)
	if !ok || valueMember == nil {
		return nil, false
	}
	key := fmt.Sprintf("%s::%s::%s", method.InterfacePackage, receiverType, returnGoType)
	if existing, ok := g.iteratorCollectMonoArrays[key]; ok && existing != nil {
		return existing, true
	}
	info := &iteratorCollectMonoArrayInfo{
		Key:               key,
		GoName:            g.mangler.unique(fmt.Sprintf("iface_%s_collect_%s", sanitizeIdent(method.InterfaceName), sanitizeIdent(returnGoType))),
		Package:           method.InterfacePackage,
		ReceiverType:      receiverType,
		ReturnType:        returnGoType,
		ElemGoType:        spec.ElemGoType,
		IteratorEndUnwrap: iterEndMember.UnwrapHelper,
		ValueUnwrap:       valueMember.UnwrapHelper,
	}
	g.iteratorCollectMonoArrays[key] = info
	return info, true
}

func (g *generator) renderIteratorCollectMonoArrayHelpers(buf *bytes.Buffer) {
	if g == nil || buf == nil || len(g.iteratorCollectMonoArrays) == 0 {
		return
	}
	keys := make([]string, 0, len(g.iteratorCollectMonoArrays))
	for key := range g.iteratorCollectMonoArrays {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		info := g.iteratorCollectMonoArrays[key]
		if info == nil {
			continue
		}
		g.renderIteratorCollectMonoArrayHelper(buf, info)
	}
}

func (g *generator) renderIteratorCollectMonoArrayHelper(buf *bytes.Buffer, info *iteratorCollectMonoArrayInfo) {
	if g == nil || buf == nil || info == nil {
		return
	}
	spec, ok := g.monoArraySpecForGoType(info.ReturnType)
	if !ok || spec == nil {
		return
	}
	zeroExpr, ok := g.zeroValueExpr(info.ReturnType)
	if !ok {
		return
	}
	if envVar, ok := g.packageEnvVar(info.Package); ok {
		fmt.Fprintf(buf, "func __able_compiled_%s(self %s) (%s, *__ableControl) {\n", info.GoName, info.ReceiverType, info.ReturnType)
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "__able_runtime", envVar, "")
	} else {
		fmt.Fprintf(buf, "func __able_compiled_%s(self %s) (%s, *__ableControl) {\n", info.GoName, info.ReceiverType, info.ReturnType)
	}
	fmt.Fprintf(buf, "\tacc := &%s{Elements: make([]%s, 0)}\n", spec.GoName, spec.ElemGoType)
	fmt.Fprintf(buf, "\titer := self\n")
	fmt.Fprintf(buf, "\tfor {\n")
	fmt.Fprintf(buf, "\t\t__able_push_call_frame(nil)\n")
	fmt.Fprintf(buf, "\t\tnext, control := iter.next()\n")
	fmt.Fprintf(buf, "\t\t__able_pop_call_frame()\n")
	fmt.Fprintf(buf, "\t\tif control != nil {\n")
	fmt.Fprintf(buf, "\t\t\treturn %s, control\n", zeroExpr)
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tendValue, endOK := %s(next)\n", info.IteratorEndUnwrap)
	fmt.Fprintf(buf, "\t\tendMatch := false\n")
	fmt.Fprintf(buf, "\t\tif endOK {\n")
	fmt.Fprintf(buf, "\t\t\tendMatch = (endValue != nil)\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif endMatch {\n")
	fmt.Fprintf(buf, "\t\t\tbreak\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tvalue, valueOK := %s(next)\n", info.ValueUnwrap)
	fmt.Fprintf(buf, "\t\tif valueOK {\n")
	fmt.Fprintf(buf, "\t\t\tacc.Elements = append(acc.Elements, value)\n")
	fmt.Fprintf(buf, "\t\t\tcontinue\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn %s, __able_runtime_error_control(nil, fmt.Errorf(\"Non-exhaustive match\"))\n", zeroExpr)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\t%s\n", g.staticArraySyncCall(info.ReturnType, "acc"))
	fmt.Fprintf(buf, "\treturn acc, nil\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) finishNativeInterfaceGenericCallReturn(ctx *compileContext, lines []string, resultExpr string, resultType string, expected string) ([]string, string, string, bool) {
	if g == nil || ctx == nil {
		return nil, "", "", false
	}
	if expected == "" || g.typeMatches(expected, resultType) {
		return lines, resultExpr, resultType, true
	}
	if expected != "runtime.Value" && resultType == "runtime.Value" {
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, resultExpr, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	if expected == "runtime.Value" && resultType != "runtime.Value" {
		convLines, converted, ok := g.lowerRuntimeValue(ctx, resultExpr, resultType)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, "runtime.Value", true
	}
	if expected != "" && expected != "any" && resultType == "any" {
		anyTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", anyTemp, resultExpr))
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, anyTemp, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	if expected != "" && expected != "runtime.Value" && expected != "any" && g.canCoerceStaticExpr(expected, resultType) {
		return g.lowerCoerceExpectedStaticExpr(ctx, lines, resultExpr, resultType, expected)
	}
	ctx.setReason("call return type mismatch")
	return nil, "", "", false
}
