package bridge

import "able/interpreter-go/pkg/runtime"

// SwapEnvIfNeeded avoids redundant environment swaps when compiled code is
// already executing under the target package environment.
func SwapEnvIfNeeded(r *Runtime, env *runtime.Environment) (*runtime.Environment, bool) {
	if r == nil || env == nil {
		return nil, false
	}
	if !r.isConcurrent() {
		prev := r.env.Load()
		if prev == env {
			return nil, false
		}
		r.env.Store(env)
		return prev, true
	}
	gid := currentGID()
	prev := r.goroutineEnv(gid)
	if prev == nil {
		prev = r.env.Load()
	}
	if prev == env {
		return nil, false
	}
	r.envByGID.Store(gid, env)
	return prev, true
}

func RestoreEnvIfNeeded(r *Runtime, prev *runtime.Environment, swapped bool) {
	if r == nil || !swapped {
		return
	}
	r.SwapEnv(prev)
}
