package interpreter

import "able/interpreter-go/pkg/runtime"

func bytecodeDirectMemberCallHasDistinctInjectedReceiver(receiver runtime.Value, injectedReceiver runtime.Value, hasInjectedReceiver bool) bool {
	if !hasInjectedReceiver {
		return false
	}
	return !bytecodeCanReuseResolvedMemberStackReceiver(receiver, injectedReceiver)
}

func bytecodeCanReuseResolvedMemberStackReceiver(stackReceiver runtime.Value, injectedReceiver runtime.Value) bool {
	switch injected := injectedReceiver.(type) {
	case runtime.StringValue:
		other, ok := stackReceiver.(runtime.StringValue)
		return ok && other == injected
	case *runtime.StringValue:
		other, ok := stackReceiver.(*runtime.StringValue)
		return ok && other == injected
	case runtime.BoolValue:
		other, ok := stackReceiver.(runtime.BoolValue)
		return ok && other == injected
	case *runtime.BoolValue:
		other, ok := stackReceiver.(*runtime.BoolValue)
		return ok && other == injected
	case runtime.CharValue:
		other, ok := stackReceiver.(runtime.CharValue)
		return ok && other == injected
	case *runtime.CharValue:
		other, ok := stackReceiver.(*runtime.CharValue)
		return ok && other == injected
	case runtime.NilValue:
		_, ok := stackReceiver.(runtime.NilValue)
		return ok
	case *runtime.NilValue:
		other, ok := stackReceiver.(*runtime.NilValue)
		return ok && other == injected
	case runtime.IntegerValue:
		other, ok := stackReceiver.(runtime.IntegerValue)
		return ok && other == injected
	case *runtime.IntegerValue:
		other, ok := stackReceiver.(*runtime.IntegerValue)
		return ok && other == injected
	case runtime.FloatValue:
		other, ok := stackReceiver.(runtime.FloatValue)
		return ok && other == injected
	case *runtime.FloatValue:
		other, ok := stackReceiver.(*runtime.FloatValue)
		return ok && other == injected
	case *runtime.ArrayValue:
		other, ok := stackReceiver.(*runtime.ArrayValue)
		return ok && other == injected
	case *runtime.StructInstanceValue:
		other, ok := stackReceiver.(*runtime.StructInstanceValue)
		return ok && other == injected
	case *runtime.InterfaceValue:
		other, ok := stackReceiver.(*runtime.InterfaceValue)
		return ok && other == injected
	case *runtime.IteratorValue:
		other, ok := stackReceiver.(*runtime.IteratorValue)
		return ok && other == injected
	case *runtime.FutureValue:
		other, ok := stackReceiver.(*runtime.FutureValue)
		return ok && other == injected
	case *runtime.HasherValue:
		other, ok := stackReceiver.(*runtime.HasherValue)
		return ok && other == injected
	case *runtime.HostHandleValue:
		other, ok := stackReceiver.(*runtime.HostHandleValue)
		return ok && other == injected
	default:
		return false
	}
}
