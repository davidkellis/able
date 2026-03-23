package bridge

import "able/interpreter-go/pkg/runtime"

// SwapEnvIfNeeded avoids redundant environment swaps when compiled code is
// already executing under the target package environment.
func SwapEnvIfNeeded(r *Runtime, env *runtime.Environment) (*runtime.Environment, bool) {
	if r == nil || env == nil {
		return nil, false
	}
	if r.Env() == env {
		return nil, false
	}
	return r.SwapEnv(env), true
}

func RestoreEnvIfNeeded(r *Runtime, prev *runtime.Environment, swapped bool) {
	if r == nil || !swapped {
		return
	}
	r.SwapEnv(prev)
}
