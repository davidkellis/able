package compiler

import (
	"bytes"
	"fmt"
)

func (g *generator) renderFunctionThunks(buf *bytes.Buffer) {
	for _, info := range g.sortedFunctionInfos() {
		if info == nil || !info.Compileable || info.InternalOnly {
			continue
		}
		fmt.Fprintf(buf, "func __able_function_thunk_%s(env *runtime.Environment, args []runtime.Value) (runtime.Value, error) {\n", info.GoName)
		fmt.Fprintf(buf, "\tctx := &runtime.NativeCallContext{Env: env}\n")
		fmt.Fprintf(buf, "\treturn __able_wrap_%s(__able_runtime, ctx, args)\n", info.GoName)
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderMethodThunks(buf *bytes.Buffer) {
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil {
			continue
		}
		if !g.registerableMethod(method) {
			continue
		}
		info := method.Info
		fmt.Fprintf(buf, "func __able_method_thunk_%s(env *runtime.Environment, args []runtime.Value) (runtime.Value, error) {\n", info.GoName)
		fmt.Fprintf(buf, "\tctx := &runtime.NativeCallContext{Env: env}\n")
		fmt.Fprintf(buf, "\treturn __able_wrap_%s(__able_runtime, ctx, args)\n", info.GoName)
		fmt.Fprintf(buf, "}\n\n")
	}
}
