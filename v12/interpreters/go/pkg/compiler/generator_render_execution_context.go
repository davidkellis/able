package compiler

import "bytes"

func (g *generator) renderExecutionContextRuntime(buf *bytes.Buffer) {
	if !g.executionContextsEnabled() {
		return
	}
	nativeField := ""
	nativePreparation := ""
	nativeReuse := ""
	payloadAttachment := ""
	prepare := func(expr string) string {
		return expr
	}
	if g.callableExecutionContextsEnabled() {
		nativeField = "\tnative     runtime.NativeCallContext\n"
		nativePreparation = `
func __able_prepare_execution_context(ctx *__able_execution_context) *__able_execution_context {
	if ctx == nil {
		return nil
	}
	var state any
	if ctx.payload != nil {
		state = ctx.payload
	} else if ctx.env != nil {
		state = ctx.env.RuntimeData()
	}
ctx.native = runtime.NativeCallContext{Env: ctx.env, State: state}
	return ctx
}
`
		nativeReuse = `	if payload, ok := native.State.(*__able_async_payload); ok && payload != nil {
		if ctx := payload.executionContext.Load(); ctx != nil && ctx.env == native.Env {
			return ctx
		}
	}
`
		payloadAttachment = `		if child.payload != nil {
			child.payload.executionContext.Store(child)
		}
`
		prepare = func(expr string) string {
			return "__able_prepare_execution_context(" + expr + ")"
		}
	}
	buf.WriteString(`
// __able_execution_context carries task-local state through generated calls.
// The bridge only remains for dynamically unresolved language boundaries.
type __able_execution_context struct {
	env        *runtime.Environment
	packageEnv *runtime.Environment
	payload    *__able_async_payload
` + nativeField + `}
` + nativePreparation + `

func __able_context_from_args(contexts ...*__able_execution_context) *__able_execution_context {
	for _, ctx := range contexts {
		if ctx != nil {
			return ctx
		}
	}
	env := (*runtime.Environment)(nil)
	if __able_runtime != nil {
		env = __able_runtime.Env()
	}
	return ` + prepare("&__able_execution_context{env: env, packageEnv: env}") + `
}

func __able_context_from_environment(env *runtime.Environment) *__able_execution_context {
	ctx := &__able_execution_context{env: env, packageEnv: env}
	if env != nil {
		ctx.payload, _ = env.RuntimeData().(*__able_async_payload)
	}
	return ` + prepare("ctx") + `
}

func __able_context_from_native(native *runtime.NativeCallContext) *__able_execution_context {
	if native == nil {
		return __able_context_from_args()
	}
` + nativeReuse + `	ctx := __able_context_from_environment(native.Env)
	if payload, ok := native.State.(*__able_async_payload); ok && payload != nil {
		ctx.payload = payload
	}
	return ` + prepare("ctx") + `
}

func __able_context_with_environment(ctx *__able_execution_context, env *runtime.Environment, local *__able_execution_context) *__able_execution_context {
	ctx = __able_context_from_args(ctx)
	if env == nil || (ctx.env == env && ctx.packageEnv == env) {
		return ctx
	}
	*local = __able_execution_context{env: env, packageEnv: env, payload: ctx.payload}
	return ` + prepare("local") + `
}

func __able_context_payload(contexts ...*__able_execution_context) *__able_async_payload {
	for _, ctx := range contexts {
		if ctx != nil && ctx.payload != nil {
			return ctx.payload
		}
	}
	return __able_current_payload()
}

func __able_mark_context_task_blocked(contexts ...*__able_execution_context) {
	payload := __able_context_payload(contexts...)
	if payload == nil || payload.handle == nil {
		return
	}
	if exec := __able_future_executor(); exec != nil {
		exec.MarkBlocked(payload.handle)
	}
}

func __able_mark_context_task_unblocked(contexts ...*__able_execution_context) {
	payload := __able_context_payload(contexts...)
	if payload == nil || payload.handle == nil {
		return
	}
	if exec := __able_future_executor(); exec != nil {
		exec.MarkUnblocked(payload.handle)
	}
}

func __able_spawn_context(parent *__able_execution_context, task func(*__able_execution_context) (runtime.Value, error)) runtime.Value {
	if __able_runtime == nil {
		panic(fmt.Errorf("compiler: missing runtime"))
	}
	if task == nil {
		return runtime.NilValue{}
	}
	__able_runtime.MarkConcurrent()
	parent = __able_context_from_args(parent)
	env := parent.env
	if env != nil {
		env = runtime.NewEnvironment(env)
	}
	exec := __able_future_executor()
	if exec == nil {
		return runtime.NilValue{}
	}
	future := exec.RunFuture(env, func(taskEnv *runtime.Environment) (runtime.Value, error) {
		child := __able_context_from_environment(taskEnv)
		child.packageEnv = parent.packageEnv
` + payloadAttachment + `		return task(child)
	})
	if future == nil {
		return runtime.NilValue{}
	}
	return future
}
`)
}
