package interpreter

// ExecutorKind reports the configured async scheduler kind.
func (i *Interpreter) ExecutorKind() string {
	if i == nil || i.executor == nil {
		return "serial"
	}
	switch i.executor.(type) {
	case *GoroutineExecutor:
		return "goroutine"
	default:
		return "serial"
	}
}
