package interpreter

import "able/interpreter-go/pkg/runtime"

// bytecodeArrayOwnershipObserver is an opt-in, test-only model of the
// proposed frame-local Array lifetime. It intentionally never releases an
// ArrayStore lease. Keeping the observer separate from ArrayValue lets the VM
// validate provenance and escape boundaries before reclamation changes normal
// execution.
type bytecodeArrayOwnershipObserver struct {
	current *bytecodeArrayOwnershipFrame
	stats   bytecodeArrayOwnershipSnapshot
	profile *bytecodeArrayOwnershipProfile
}

// bytecodeArrayOwnershipProfile collects completed VM-observer snapshots for
// the opt-in one-process benchmark hook. The hook is deliberately used only
// with the serial, bounded diagnostic controls; ordinary interpreters leave
// this nil and allocate no observer state.
type bytecodeArrayOwnershipProfile struct {
	observers []*bytecodeArrayOwnershipObserver
	pending   map[*runtime.ArrayValue]struct{}
}

type bytecodeArrayOwnershipFrame struct {
	owned   map[*runtime.ArrayValue]struct{}
	escaped map[*runtime.ArrayValue]bytecodeArrayOwnershipEscape
}

type bytecodeArrayOwnershipEscape uint8

const (
	bytecodeArrayOwnershipEscapePublicReturn bytecodeArrayOwnershipEscape = iota + 1
	bytecodeArrayOwnershipEscapeEnvironment
	bytecodeArrayOwnershipEscapeAggregate
	bytecodeArrayOwnershipEscapeClosure
	bytecodeArrayOwnershipEscapeFuture
	bytecodeArrayOwnershipEscapeUnknownCall
	bytecodeArrayOwnershipEscapeBorrowedArrayWrite
)

func (reason bytecodeArrayOwnershipEscape) String() string {
	switch reason {
	case bytecodeArrayOwnershipEscapePublicReturn:
		return "public_return"
	case bytecodeArrayOwnershipEscapeEnvironment:
		return "environment"
	case bytecodeArrayOwnershipEscapeAggregate:
		return "aggregate"
	case bytecodeArrayOwnershipEscapeClosure:
		return "closure"
	case bytecodeArrayOwnershipEscapeFuture:
		return "future"
	case bytecodeArrayOwnershipEscapeUnknownCall:
		return "unknown_call"
	case bytecodeArrayOwnershipEscapeBorrowedArrayWrite:
		return "borrowed_array_write"
	default:
		return "unknown"
	}
}

// bytecodeArrayOwnershipSnapshot is deliberately package-private: the first
// sidecar is test instrumentation, not an interpreter-facing runtime API.
// Counts are pointer identities, not ArrayStore handles, because aliases share
// a wrapper in the bytecode interpreter.
type bytecodeArrayOwnershipSnapshot struct {
	Created        int
	Transferred    int
	PublicReturned int
	Escaped        int
	FrameLocal     int
	ErrorUnwound   int
	Escapes        map[bytecodeArrayOwnershipEscape]int
}

func newBytecodeArrayOwnershipObserver(profile *bytecodeArrayOwnershipProfile) *bytecodeArrayOwnershipObserver {
	return &bytecodeArrayOwnershipObserver{
		profile: profile,
		stats: bytecodeArrayOwnershipSnapshot{
			Escapes: make(map[bytecodeArrayOwnershipEscape]int),
		},
	}
}

func (vm *bytecodeVM) enableBytecodeArrayOwnershipObserver() *bytecodeArrayOwnershipObserver {
	if vm == nil {
		return nil
	}
	vm.arrayOwnershipObserver = newBytecodeArrayOwnershipObserver(nil)
	return vm.arrayOwnershipObserver
}

func (vm *bytecodeVM) enableBytecodeArrayOwnershipProfileObserver(profile *bytecodeArrayOwnershipProfile) *bytecodeArrayOwnershipObserver {
	if vm == nil || profile == nil {
		return nil
	}
	vm.arrayOwnershipObserver = newBytecodeArrayOwnershipObserver(profile)
	return vm.arrayOwnershipObserver
}

func (vm *bytecodeVM) enableBytecodeArrayOwnershipObserverForTest() {
	vm.enableBytecodeArrayOwnershipObserver()
}

func (vm *bytecodeVM) bytecodeArrayOwnershipSnapshot() bytecodeArrayOwnershipSnapshot {
	if vm == nil || vm.arrayOwnershipObserver == nil {
		return bytecodeArrayOwnershipSnapshot{}
	}
	return bytecodeArrayOwnershipSnapshotCopy(vm.arrayOwnershipObserver.stats)
}

func bytecodeArrayOwnershipSnapshotCopy(snapshot bytecodeArrayOwnershipSnapshot) bytecodeArrayOwnershipSnapshot {
	if len(snapshot.Escapes) != 0 {
		escapes := make(map[bytecodeArrayOwnershipEscape]int, len(snapshot.Escapes))
		for reason, count := range snapshot.Escapes {
			escapes[reason] = count
		}
		snapshot.Escapes = escapes
	}
	return snapshot
}

func (profile *bytecodeArrayOwnershipProfile) add(observer *bytecodeArrayOwnershipObserver) {
	if profile == nil || observer == nil {
		return
	}
	profile.observers = append(profile.observers, observer)
}

func (profile *bytecodeArrayOwnershipProfile) reset() {
	if profile == nil {
		return
	}
	profile.observers = profile.observers[:0]
	clear(profile.pending)
}

func (profile *bytecodeArrayOwnershipProfile) recordDetachedReturn(arr *runtime.ArrayValue) {
	if profile == nil || arr == nil {
		return
	}
	if profile.pending == nil {
		profile.pending = make(map[*runtime.ArrayValue]struct{}, 1)
	}
	profile.pending[arr] = struct{}{}
}

func (profile *bytecodeArrayOwnershipProfile) takeDetachedReturn(arr *runtime.ArrayValue) bool {
	if profile == nil || arr == nil || profile.pending == nil {
		return false
	}
	if _, ok := profile.pending[arr]; !ok {
		return false
	}
	delete(profile.pending, arr)
	return true
}

func (profile *bytecodeArrayOwnershipProfile) hasDetachedReturn(value runtime.Value) bool {
	if profile == nil || len(profile.pending) == 0 {
		return false
	}
	for arr := range bytecodeArrayOwnershipArraysInValue(value) {
		if _, ok := profile.pending[arr]; ok {
			return true
		}
	}
	return false
}

func (profile *bytecodeArrayOwnershipProfile) snapshot() bytecodeArrayOwnershipSnapshot {
	var total bytecodeArrayOwnershipSnapshot
	if profile == nil {
		return total
	}
	for _, observer := range profile.observers {
		if observer == nil {
			continue
		}
		total.Created += observer.stats.Created
		total.Transferred += observer.stats.Transferred
		total.PublicReturned += observer.stats.PublicReturned
		total.Escaped += observer.stats.Escaped
		total.FrameLocal += observer.stats.FrameLocal
		total.ErrorUnwound += observer.stats.ErrorUnwound
		if len(observer.stats.Escapes) == 0 {
			continue
		}
		if total.Escapes == nil {
			total.Escapes = make(map[bytecodeArrayOwnershipEscape]int, len(observer.stats.Escapes))
		}
		for reason, count := range observer.stats.Escapes {
			total.Escapes[reason] += count
		}
	}
	return total
}

func (i *Interpreter) enableBytecodeArrayOwnershipProfile() *bytecodeArrayOwnershipProfile {
	if i == nil {
		return nil
	}
	profile := &bytecodeArrayOwnershipProfile{}
	i.bytecodeArrayOwnershipProfile = profile
	return profile
}

func (i *Interpreter) disableBytecodeArrayOwnershipProfile() {
	if i == nil {
		return
	}
	i.bytecodeArrayOwnershipProfile = nil
}

// ensureBytecodeArrayOwnershipForProgram enables the diagnostic observer only
// after lowering has found a verified canonical Array creation boundary. An
// observer activated in an inline callee reconstructs empty parent contexts
// for the active VM frames, so provenance can still return through callers
// that have not themselves constructed an Array.
func (vm *bytecodeVM) ensureBytecodeArrayOwnershipForProgram(program *bytecodeProgram) {
	if vm == nil || vm.interp == nil || vm.interp.bytecodeArrayOwnershipProfile == nil || program == nil || !program.arrayOwnershipMetadata.observesArrays() {
		return
	}
	vm.ensureBytecodeArrayOwnershipProfileContext()
}

func (vm *bytecodeVM) ensureBytecodeArrayOwnershipProfileContext() *bytecodeArrayOwnershipObserver {
	if vm == nil || vm.interp == nil {
		return nil
	}
	profile := vm.interp.bytecodeArrayOwnershipProfile
	if profile == nil {
		return nil
	}
	observer := vm.arrayOwnershipObserver
	if observer == nil {
		observer = vm.enableBytecodeArrayOwnershipProfileObserver(profile)
		profile.add(observer)
	}
	if observer.current == nil {
		vm.installBytecodeArrayOwnershipActiveFrames(observer)
	}
	return observer
}

func (vm *bytecodeVM) installBytecodeArrayOwnershipActiveFrames(observer *bytecodeArrayOwnershipObserver) {
	if vm == nil || observer == nil || observer.current != nil {
		return
	}
	current := &bytecodeArrayOwnershipFrame{}
	fullIndex := 0
	selfFastIndex := 0
	selfFastMinimalIndex := 0
	install := func(parent **bytecodeArrayOwnershipFrame) {
		*parent = current
		current = &bytecodeArrayOwnershipFrame{}
	}
	for _, kind := range vm.callFrameKinds {
		switch kind {
		case bytecodeCallFrameKindFull:
			if fullIndex < len(vm.callFrames) {
				install(&vm.callFrames[fullIndex].arrayOwnershipParent)
			}
			fullIndex++
		case bytecodeCallFrameKindSelfFast:
			if selfFastIndex < len(vm.selfFastCallFrames) {
				install(&vm.selfFastCallFrames[selfFastIndex].arrayOwnershipParent)
			}
			selfFastIndex++
		case bytecodeCallFrameKindSelfFastMinimal:
			if selfFastMinimalIndex < len(vm.selfFastMinimal) {
				install(&vm.selfFastMinimal[selfFastMinimalIndex].arrayOwnershipParent)
			}
			selfFastMinimalIndex++
		}
	}
	firstSuffix := len(vm.selfFastMinimal) - vm.selfFastMinimalSuffix
	if firstSuffix < selfFastMinimalIndex {
		firstSuffix = selfFastMinimalIndex
	}
	for idx := firstSuffix; idx < len(vm.selfFastMinimal); idx++ {
		install(&vm.selfFastMinimal[idx].arrayOwnershipParent)
	}
	observer.current = current
}

func (vm *bytecodeVM) beginBytecodeArrayOwnershipFrame(parent **bytecodeArrayOwnershipFrame) {
	if vm == nil || vm.arrayOwnershipObserver == nil || parent == nil {
		return
	}
	observer := vm.arrayOwnershipObserver
	if observer.profile != nil && observer.current == nil {
		*parent = nil
		return
	}
	*parent = observer.current
	// A frame must exist even when the caller has not allocated an Array yet.
	// Otherwise the first child that does allocate has a nil parent and its
	// returned wrapper is indistinguishable from a public VM result. This is a
	// bookkeeping frame only; the observer is opt-in and it holds no ArrayStore
	// lease or release behavior.
	vm.arrayOwnershipObserver.current = &bytecodeArrayOwnershipFrame{}
}

func (vm *bytecodeVM) topBytecodeArrayOwnershipParent() *bytecodeArrayOwnershipFrame {
	if vm == nil || vm.arrayOwnershipObserver == nil {
		return nil
	}
	if vm.selfFastMinimalSuffix > 0 && len(vm.selfFastMinimal) > 0 {
		return vm.selfFastMinimal[len(vm.selfFastMinimal)-1].arrayOwnershipParent
	}
	if len(vm.callFrameKinds) == 0 {
		return nil
	}
	switch vm.callFrameKinds[len(vm.callFrameKinds)-1] {
	case bytecodeCallFrameKindSelfFastMinimal:
		if len(vm.selfFastMinimal) > 0 {
			return vm.selfFastMinimal[len(vm.selfFastMinimal)-1].arrayOwnershipParent
		}
	case bytecodeCallFrameKindSelfFast:
		if len(vm.selfFastCallFrames) > 0 {
			return vm.selfFastCallFrames[len(vm.selfFastCallFrames)-1].arrayOwnershipParent
		}
	case bytecodeCallFrameKindFull:
		if len(vm.callFrames) > 0 {
			return vm.callFrames[len(vm.callFrames)-1].arrayOwnershipParent
		}
	}
	return nil
}

func (vm *bytecodeVM) trackBytecodeArrayOwnershipCreation(arr *runtime.ArrayValue) {
	if vm == nil || vm.arrayOwnershipObserver == nil || arr == nil {
		return
	}
	frame := vm.arrayOwnershipObserver.current
	if frame == nil {
		frame = &bytecodeArrayOwnershipFrame{}
		vm.arrayOwnershipObserver.current = frame
	}
	if frame.owned == nil {
		frame.owned = make(map[*runtime.ArrayValue]struct{}, 1)
	}
	if _, exists := frame.owned[arr]; exists {
		return
	}
	frame.owned[arr] = struct{}{}
	vm.arrayOwnershipObserver.stats.Created++
}

func (vm *bytecodeVM) adoptBytecodeArrayOwnershipReturnedValue(value runtime.Value) {
	if vm == nil || vm.interp == nil || vm.interp.bytecodeArrayOwnershipProfile == nil {
		return
	}
	profile := vm.interp.bytecodeArrayOwnershipProfile
	if !profile.hasDetachedReturn(value) {
		return
	}
	observer := vm.ensureBytecodeArrayOwnershipProfileContext()
	if observer == nil {
		return
	}
	for arr := range bytecodeArrayOwnershipArraysInValue(value) {
		if !observer.profile.takeDetachedReturn(arr) {
			continue
		}
		frame := observer.current
		if frame == nil {
			frame = &bytecodeArrayOwnershipFrame{}
			observer.current = frame
		}
		if frame.owned == nil {
			frame.owned = make(map[*runtime.ArrayValue]struct{}, 1)
		}
		if _, alreadyOwned := frame.owned[arr]; alreadyOwned {
			continue
		}
		frame.owned[arr] = struct{}{}
		observer.stats.Transferred++
	}
}

func (vm *bytecodeVM) bytecodeArrayOwnershipOwns(arr *runtime.ArrayValue) bool {
	if vm == nil || vm.arrayOwnershipObserver == nil || vm.arrayOwnershipObserver.current == nil || arr == nil {
		return false
	}
	_, ok := vm.arrayOwnershipObserver.current.owned[arr]
	return ok
}

func (vm *bytecodeVM) markBytecodeArrayOwnershipValueEscaped(value runtime.Value, reason bytecodeArrayOwnershipEscape) {
	if vm == nil || vm.arrayOwnershipObserver == nil || vm.arrayOwnershipObserver.current == nil {
		return
	}
	for arr := range bytecodeArrayOwnershipArraysInValue(value) {
		vm.markBytecodeArrayOwnershipArrayEscaped(arr, reason)
	}
}

func (vm *bytecodeVM) markBytecodeArrayOwnershipValuesEscaped(values []runtime.Value, reason bytecodeArrayOwnershipEscape) {
	for _, value := range values {
		vm.markBytecodeArrayOwnershipValueEscaped(value, reason)
	}
}

func (vm *bytecodeVM) markAllBytecodeArrayOwnershipEscaped(reason bytecodeArrayOwnershipEscape) {
	if vm == nil || vm.arrayOwnershipObserver == nil || vm.arrayOwnershipObserver.current == nil {
		return
	}
	for arr := range vm.arrayOwnershipObserver.current.owned {
		vm.markBytecodeArrayOwnershipArrayEscaped(arr, reason)
	}
}

func (vm *bytecodeVM) markBytecodeArrayOwnershipArrayEscaped(arr *runtime.ArrayValue, reason bytecodeArrayOwnershipEscape) {
	if vm == nil || vm.arrayOwnershipObserver == nil || vm.arrayOwnershipObserver.current == nil || arr == nil {
		return
	}
	frame := vm.arrayOwnershipObserver.current
	if _, owned := frame.owned[arr]; !owned {
		return
	}
	if frame.escaped == nil {
		frame.escaped = make(map[*runtime.ArrayValue]bytecodeArrayOwnershipEscape, 1)
	}
	if _, alreadyEscaped := frame.escaped[arr]; alreadyEscaped {
		return
	}
	frame.escaped[arr] = reason
}

func (vm *bytecodeVM) observeBytecodeArrayOwnershipArrayWrite(receiver *runtime.ArrayValue, value runtime.Value) {
	if vm == nil || vm.arrayOwnershipObserver == nil {
		return
	}
	if vm.bytecodeArrayOwnershipOwns(receiver) {
		return
	}
	vm.markBytecodeArrayOwnershipValueEscaped(value, bytecodeArrayOwnershipEscapeBorrowedArrayWrite)
}

func (vm *bytecodeVM) finishBytecodeArrayOwnershipReturn(value runtime.Value, parent *bytecodeArrayOwnershipFrame) {
	vm.finishBytecodeArrayOwnershipFrame(value, parent, false)
}

func (vm *bytecodeVM) finishBytecodeArrayOwnershipError(parent *bytecodeArrayOwnershipFrame) {
	vm.finishBytecodeArrayOwnershipFrame(nil, parent, true)
}

func (vm *bytecodeVM) finishBytecodeArrayOwnershipPublicReturn(value runtime.Value, errored bool) {
	if vm == nil || vm.arrayOwnershipObserver == nil {
		return
	}
	vm.finishBytecodeArrayOwnershipFrame(value, nil, errored)
}

// finishBytecodeArrayOwnershipDetachedReturn handles a bytecode function
// invoked through the generic call bridge rather than an inline VM frame. A
// returned Array wrapper remains pending until the caller VM receives it and
// adopts it; a non-returned wrapper is still a frame-local candidate.
func (vm *bytecodeVM) finishBytecodeArrayOwnershipDetachedReturn(value runtime.Value, errored bool) {
	if vm == nil || vm.arrayOwnershipObserver == nil {
		return
	}
	observer := vm.arrayOwnershipObserver
	frame := observer.current
	observer.current = nil
	if frame == nil {
		return
	}
	returned := bytecodeArrayOwnershipArraysInValue(value)
	for arr := range frame.owned {
		if _, isReturned := returned[arr]; isReturned && !errored {
			if reason, escaped := frame.escaped[arr]; escaped {
				observer.stats.Escaped++
				observer.stats.Escapes[reason]++
				continue
			}
			if observer.profile != nil {
				observer.profile.recordDetachedReturn(arr)
			} else {
				observer.stats.PublicReturned++
			}
			continue
		}
		if reason, escaped := frame.escaped[arr]; escaped {
			observer.stats.Escaped++
			observer.stats.Escapes[reason]++
			continue
		}
		if errored {
			observer.stats.ErrorUnwound++
		} else {
			observer.stats.FrameLocal++
		}
	}
}

func (vm *bytecodeVM) finishBytecodeArrayOwnershipFrame(value runtime.Value, parent *bytecodeArrayOwnershipFrame, errored bool) {
	if vm == nil || vm.arrayOwnershipObserver == nil {
		return
	}
	observer := vm.arrayOwnershipObserver
	frame := observer.current
	observer.current = parent
	if frame == nil {
		return
	}
	returned := bytecodeArrayOwnershipArraysInValue(value)
	for arr := range frame.owned {
		if _, isReturned := returned[arr]; isReturned && !errored {
			if reason, escaped := frame.escaped[arr]; escaped {
				observer.stats.Escaped++
				observer.stats.Escapes[reason]++
				continue
			}
			if parent != nil {
				if parent.owned == nil {
					parent.owned = make(map[*runtime.ArrayValue]struct{}, 1)
				}
				parent.owned[arr] = struct{}{}
				observer.stats.Transferred++
			} else {
				observer.stats.PublicReturned++
			}
			continue
		}
		if reason, escaped := frame.escaped[arr]; escaped {
			observer.stats.Escaped++
			observer.stats.Escapes[reason]++
			continue
		}
		if errored {
			observer.stats.ErrorUnwound++
		} else {
			observer.stats.FrameLocal++
		}
	}
}

func bytecodeArrayOwnershipArraysInValue(value runtime.Value) map[*runtime.ArrayValue]struct{} {
	arrays := make(map[*runtime.ArrayValue]struct{})
	seenArrays := make(map[*runtime.ArrayValue]struct{})
	seenStructs := make(map[*runtime.StructInstanceValue]struct{})
	seenMaps := make(map[*runtime.HashMapValue]struct{})
	var visit func(runtime.Value)
	visit = func(current runtime.Value) {
		switch val := current.(type) {
		case *runtime.ArrayValue:
			if val == nil {
				return
			}
			if _, seen := seenArrays[val]; seen {
				return
			}
			seenArrays[val] = struct{}{}
			arrays[val] = struct{}{}
			for _, element := range val.Elements {
				visit(element)
			}
			if val.State != nil && val.State.Values != nil {
				for _, element := range val.State.Values {
					visit(element)
				}
			}
		case *runtime.StructInstanceValue:
			if val == nil {
				return
			}
			if _, seen := seenStructs[val]; seen {
				return
			}
			seenStructs[val] = struct{}{}
			for _, field := range val.Fields {
				visit(field)
			}
			for _, field := range val.Positional {
				visit(field)
			}
		case runtime.InterfaceValue:
			visit(val.Underlying)
		case *runtime.InterfaceValue:
			if val != nil {
				visit(val.Underlying)
			}
		case *runtime.HashMapValue:
			if val == nil {
				return
			}
			if _, seen := seenMaps[val]; seen {
				return
			}
			seenMaps[val] = struct{}{}
			for _, entry := range val.Entries {
				visit(entry.Key)
				visit(entry.Value)
			}
		case runtime.ErrorValue:
			for _, payload := range val.Payload {
				visit(payload)
			}
		}
	}
	visit(value)
	return arrays
}
