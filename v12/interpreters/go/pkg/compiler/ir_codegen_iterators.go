package compiler

import "fmt"

func (e *irGoEmitter) scanFunctions(funcs []*IRFunction) {
	for _, fn := range funcs {
		if fn == nil {
			continue
		}
		for _, block := range fn.Blocks {
			if block == nil {
				continue
			}
			for _, instr := range block.Instructions {
				switch instr.(type) {
				case *IRIteratorLiteral:
					e.needsIterator = true
				}
			}
		}
	}
}

func (e *irGoEmitter) emitIteratorLiteral(node *IRIteratorLiteral) error {
	if node == nil || node.Dest == nil || node.Body == nil {
		return fmt.Errorf("compiler: IR iterator literal missing body or destination")
	}
	dest := e.nameForValue(node.Dest)
	args := make([]string, 0, len(node.Captures))
	for _, slot := range node.Captures {
		if slot == nil {
			continue
		}
		args = append(args, e.nameForSlot(slot))
	}
	bodyName := e.functionName(node.Body)
	runVar := e.mangler.unique("iter_run")
	fmt.Fprintf(&e.buf, "\t%s := func(gen *__able_generator) error {\n", runVar)
	fmt.Fprintf(&e.buf, "\t\t_, err := %s(rt, gen", bodyName)
	for _, arg := range args {
		fmt.Fprintf(&e.buf, ", %s", arg)
	}
	fmt.Fprintf(&e.buf, ")\n")
	fmt.Fprintf(&e.buf, "\t\treturn err\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\t%s = __able_new_iterator(rt, %s)\n", dest, runVar)
	fmt.Fprintf(&e.buf, "\tif %s == nil {\n\t\t%s = runtime.NilValue{}\n\t}\n", dest, dest)
	return nil
}

func (e *irGoEmitter) emitIteratorHelpers() {
	fmt.Fprintf(&e.buf, "type __able_generator_stop struct{}\n\n")
	fmt.Fprintf(&e.buf, "func (__able_generator_stop) Error() string { return \"generator stopped\" }\n\n")
	fmt.Fprintf(&e.buf, "func __able_is_generator_stop(err error) bool {\n")
	fmt.Fprintf(&e.buf, "\tif err == nil {\n\t\treturn false\n\t}\n")
	fmt.Fprintf(&e.buf, "\tvar stop __able_generator_stop\n")
	fmt.Fprintf(&e.buf, "\treturn errors.As(err, &stop)\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "type __able_generator_result struct {\n")
	fmt.Fprintf(&e.buf, "\tvalue runtime.Value\n")
	fmt.Fprintf(&e.buf, "\tdone bool\n")
	fmt.Fprintf(&e.buf, "\terr error\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "type __able_generator struct {\n")
	fmt.Fprintf(&e.buf, "\trt *bridge.Runtime\n")
	fmt.Fprintf(&e.buf, "\trequests chan struct{}\n")
	fmt.Fprintf(&e.buf, "\tresults chan __able_generator_result\n")
	fmt.Fprintf(&e.buf, "\tmu sync.Mutex\n")
	fmt.Fprintf(&e.buf, "\tclosed bool\n")
	fmt.Fprintf(&e.buf, "\tbusy bool\n")
	fmt.Fprintf(&e.buf, "\tdone bool\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func __able_new_iterator(rt *bridge.Runtime, run func(gen *__able_generator) error) *runtime.IteratorValue {\n")
	fmt.Fprintf(&e.buf, "\tgen := &__able_generator{rt: rt, requests: make(chan struct{}), results: make(chan __able_generator_result)}\n")
	fmt.Fprintf(&e.buf, "\tgo gen.run(run)\n")
	fmt.Fprintf(&e.buf, "\treturn runtime.NewIteratorValue(gen.next, gen.close)\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func (g *__able_generator) run(run func(gen *__able_generator) error) {\n")
	fmt.Fprintf(&e.buf, "\tdefer close(g.results)\n")
	fmt.Fprintf(&e.buf, "\tif !g.awaitRequest() {\n\t\treturn\n\t}\n")
	fmt.Fprintf(&e.buf, "\tvar runErr error\n")
	fmt.Fprintf(&e.buf, "\tfunc() {\n")
	fmt.Fprintf(&e.buf, "\t\tdefer func() {\n")
	fmt.Fprintf(&e.buf, "\t\t\tif r := recover(); r != nil {\n")
	fmt.Fprintf(&e.buf, "\t\t\t\tif err, ok := r.(error); ok {\n")
	fmt.Fprintf(&e.buf, "\t\t\t\t\tif __able_is_generator_stop(err) {\n")
	fmt.Fprintf(&e.buf, "\t\t\t\t\t\trunErr = __able_generator_stop{}\n")
	fmt.Fprintf(&e.buf, "\t\t\t\t\t\treturn\n")
	fmt.Fprintf(&e.buf, "\t\t\t\t\t}\n")
	fmt.Fprintf(&e.buf, "\t\t\t\t\trunErr = err\n")
	fmt.Fprintf(&e.buf, "\t\t\t\t\treturn\n")
	fmt.Fprintf(&e.buf, "\t\t\t\t}\n")
	fmt.Fprintf(&e.buf, "\t\t\t\trunErr = bridge.Recover(g.rt, nil, r)\n")
	fmt.Fprintf(&e.buf, "\t\t\t}\n")
	fmt.Fprintf(&e.buf, "\t\t}()\n")
	fmt.Fprintf(&e.buf, "\t\trunErr = run(g)\n")
	fmt.Fprintf(&e.buf, "\t}()\n")
	fmt.Fprintf(&e.buf, "\tif runErr != nil {\n")
	fmt.Fprintf(&e.buf, "\t\tif __able_is_generator_stop(runErr) {\n")
	fmt.Fprintf(&e.buf, "\t\t\tg.results <- __able_generator_result{done: true}\n")
	fmt.Fprintf(&e.buf, "\t\t\treturn\n")
	fmt.Fprintf(&e.buf, "\t\t}\n")
	fmt.Fprintf(&e.buf, "\t\tg.results <- __able_generator_result{err: runErr, done: true}\n")
	fmt.Fprintf(&e.buf, "\t\treturn\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\tg.results <- __able_generator_result{done: true}\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func (g *__able_generator) next() (runtime.Value, bool, error) {\n")
	fmt.Fprintf(&e.buf, "\tg.mu.Lock()\n")
	fmt.Fprintf(&e.buf, "\tif g.busy {\n")
	fmt.Fprintf(&e.buf, "\t\tg.mu.Unlock()\n")
	fmt.Fprintf(&e.buf, "\t\treturn nil, true, fmt.Errorf(\"iterator.next re-entered while suspended at yield\")\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\tif g.closed || g.done {\n")
	fmt.Fprintf(&e.buf, "\t\tg.mu.Unlock()\n")
	fmt.Fprintf(&e.buf, "\t\treturn runtime.IteratorEnd, true, nil\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\tg.busy = true\n")
	fmt.Fprintf(&e.buf, "\tg.mu.Unlock()\n")
	fmt.Fprintf(&e.buf, "\tg.requests <- struct{}{}\n")
	fmt.Fprintf(&e.buf, "\tres, ok := <-g.results\n")
	fmt.Fprintf(&e.buf, "\tg.mu.Lock()\n")
	fmt.Fprintf(&e.buf, "\tg.busy = false\n")
	fmt.Fprintf(&e.buf, "\tif !ok {\n")
	fmt.Fprintf(&e.buf, "\t\tg.done = true\n")
	fmt.Fprintf(&e.buf, "\t\tg.mu.Unlock()\n")
	fmt.Fprintf(&e.buf, "\t\treturn runtime.IteratorEnd, true, nil\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\tif res.err != nil {\n")
	fmt.Fprintf(&e.buf, "\t\tg.done = true\n")
	fmt.Fprintf(&e.buf, "\t\tg.mu.Unlock()\n")
	fmt.Fprintf(&e.buf, "\t\treturn nil, true, res.err\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\tif res.done {\n")
	fmt.Fprintf(&e.buf, "\t\tg.done = true\n")
	fmt.Fprintf(&e.buf, "\t\tg.mu.Unlock()\n")
	fmt.Fprintf(&e.buf, "\t\treturn runtime.IteratorEnd, true, nil\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\tg.mu.Unlock()\n")
	fmt.Fprintf(&e.buf, "\tif res.value == nil {\n")
	fmt.Fprintf(&e.buf, "\t\treturn runtime.NilValue{}, false, nil\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\treturn res.value, false, nil\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func (g *__able_generator) close() {\n")
	fmt.Fprintf(&e.buf, "\tg.mu.Lock()\n")
	fmt.Fprintf(&e.buf, "\tif g.closed {\n")
	fmt.Fprintf(&e.buf, "\t\tg.mu.Unlock()\n")
	fmt.Fprintf(&e.buf, "\t\treturn\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\tg.closed = true\n")
	fmt.Fprintf(&e.buf, "\tclose(g.requests)\n")
	fmt.Fprintf(&e.buf, "\tg.mu.Unlock()\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func (g *__able_generator) awaitRequest() bool {\n")
	fmt.Fprintf(&e.buf, "\t_, ok := <-g.requests\n")
	fmt.Fprintf(&e.buf, "\treturn ok\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func (g *__able_generator) emit(value runtime.Value) error {\n")
	fmt.Fprintf(&e.buf, "\tg.results <- __able_generator_result{value: value}\n")
	fmt.Fprintf(&e.buf, "\tif !g.awaitRequest() {\n")
	fmt.Fprintf(&e.buf, "\t\treturn __able_generator_stop{}\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\treturn nil\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func (g *__able_generator) stop() error {\n")
	fmt.Fprintf(&e.buf, "\tg.close()\n")
	fmt.Fprintf(&e.buf, "\treturn __able_generator_stop{}\n")
	fmt.Fprintf(&e.buf, "}\n\n")
}
