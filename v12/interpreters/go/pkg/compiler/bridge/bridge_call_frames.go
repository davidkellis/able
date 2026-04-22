package bridge

import "able/interpreter-go/pkg/ast"

type callFrameStack struct {
	frames []*ast.FunctionCall
}

func (r *Runtime) pushBridgeCallFrame(call *ast.FunctionCall) {
	if r == nil || call == nil {
		return
	}
	if !r.isConcurrent() {
		r.callFrames = append(r.callFrames, call)
		return
	}
	stack := r.goroutineCallFrameStack(currentGID(), true)
	stack.frames = append(stack.frames, call)
}

func (r *Runtime) popBridgeCallFrame() {
	if r == nil {
		return
	}
	if !r.isConcurrent() {
		if len(r.callFrames) == 0 {
			return
		}
		r.callFrames = r.callFrames[:len(r.callFrames)-1]
		return
	}
	stack := r.goroutineCallFrameStack(currentGID(), false)
	if stack == nil || len(stack.frames) == 0 {
		return
	}
	stack.frames = stack.frames[:len(stack.frames)-1]
}

func (r *Runtime) snapshotBridgeCallFrames() []*ast.FunctionCall {
	if r == nil {
		return nil
	}
	if !r.isConcurrent() {
		return cloneCallFrames(r.callFrames)
	}
	stack := r.goroutineCallFrameStack(currentGID(), false)
	if stack == nil {
		return nil
	}
	return cloneCallFrames(stack.frames)
}

func (r *Runtime) goroutineCallFrameStack(gid uint64, create bool) *callFrameStack {
	if r == nil {
		return nil
	}
	if existing, ok := r.callFramesByGID.Load(gid); ok {
		if typed, ok := existing.(*callFrameStack); ok && typed != nil {
			return typed
		}
	}
	if !create {
		return nil
	}
	stack := &callFrameStack{}
	actual, _ := r.callFramesByGID.LoadOrStore(gid, stack)
	if typed, ok := actual.(*callFrameStack); ok && typed != nil {
		return typed
	}
	return stack
}

func cloneCallFrames(frames []*ast.FunctionCall) []*ast.FunctionCall {
	if len(frames) == 0 {
		return nil
	}
	out := make([]*ast.FunctionCall, len(frames))
	copy(out, frames)
	return out
}
