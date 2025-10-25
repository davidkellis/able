package interpreter

import (
	"fmt"
	"sync"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

type generatorResult struct {
	value runtime.Value
	done  bool
	err   error
}

type generatorInstance struct {
	interpreter *Interpreter
	env         *runtime.Environment
	body        []ast.Statement

	requests chan struct{}
	results  chan generatorResult

	mu      sync.Mutex
	started bool
	busy    bool
	done    bool
	err     error
	closed  bool
	control runtime.Value
}

func newGeneratorInstance(i *Interpreter, env *runtime.Environment, body []ast.Statement) *generatorInstance {
	return &generatorInstance{
		interpreter: i,
		env:         env,
		body:        body,
		requests:    make(chan struct{}),
		results:     make(chan generatorResult),
	}
}

func (g *generatorInstance) next() (runtime.Value, bool, error) {
	g.mu.Lock()
	if g.busy {
		g.mu.Unlock()
		return nil, true, fmt.Errorf("iterator.next re-entered while suspended at yield")
	}
	if g.closed {
		g.mu.Unlock()
		return runtime.IteratorEnd, true, nil
	}
	if g.done {
		err := g.err
		g.mu.Unlock()
		if err != nil {
			return nil, true, err
		}
		return runtime.IteratorEnd, true, nil
	}
	g.busy = true
	if !g.started {
		g.started = true
		go g.run()
	}
	requestCh := g.requests
	g.mu.Unlock()

	requestCh <- struct{}{}
	res, ok := <-g.results

	g.mu.Lock()
	g.busy = false
	if !ok {
		g.done = true
		err := g.err
		g.mu.Unlock()
		if err != nil {
			return nil, true, err
		}
		return runtime.IteratorEnd, true, nil
	}
	if res.err != nil {
		g.done = true
		g.err = res.err
		g.mu.Unlock()
		return nil, true, res.err
	}
	if res.done {
		g.done = true
		g.mu.Unlock()
		return runtime.IteratorEnd, true, nil
	}
	g.mu.Unlock()
	if res.value == nil {
		return runtime.NilValue{}, false, nil
	}
	return res.value, false, nil
}

func (g *generatorInstance) run() {
	defer close(g.results)

	g.interpreter.pushGenerator(g)
	defer g.interpreter.popGenerator()

	if !g.awaitRequest() {
		g.mu.Lock()
		g.done = true
		g.mu.Unlock()
		return
	}

	if err := g.execute(); err != nil {
		switch sig := err.(type) {
		case returnSignal:
			g.results <- generatorResult{done: true}
		case raiseSignal:
			g.mu.Lock()
			g.err = sig
			g.mu.Unlock()
			g.results <- generatorResult{err: sig}
		default:
			g.mu.Lock()
			g.err = err
			g.mu.Unlock()
			g.results <- generatorResult{err: err}
		}
		return
	}

	g.results <- generatorResult{done: true}
}

func (g *generatorInstance) execute() error {
	for _, stmt := range g.body {
		if _, err := g.interpreter.evaluateStatement(stmt, g.env); err != nil {
			return err
		}
	}
	return nil
}

func (g *generatorInstance) emit(value runtime.Value) error {
	g.results <- generatorResult{value: value}
	if !g.awaitRequest() {
		g.mu.Lock()
		g.done = true
		g.mu.Unlock()
		return nil
	}
	return nil
}

func (g *generatorInstance) awaitRequest() bool {
	_, ok := <-g.requests
	return ok
}

func (g *generatorInstance) close() {
	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		return
	}
	g.closed = true
	close(g.requests)
	g.mu.Unlock()
}

func (i *Interpreter) pushGenerator(g *generatorInstance) {
	i.generatorStack = append(i.generatorStack, g)
}

func (i *Interpreter) popGenerator() {
	if len(i.generatorStack) == 0 {
		return
	}
	i.generatorStack = i.generatorStack[:len(i.generatorStack)-1]
}

func (i *Interpreter) currentGenerator() *generatorInstance {
	if len(i.generatorStack) == 0 {
		return nil
	}
	return i.generatorStack[len(i.generatorStack)-1]
}

func (g *generatorInstance) controllerValue() runtime.Value {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.control != nil {
		return g.control
	}

	yieldFn := runtime.NativeFunctionValue{
		Name:  "__iterator_controller_yield",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			var value runtime.Value = runtime.NilValue{}
			if len(args) > 0 {
				if len(args) > 1 {
					return nil, fmt.Errorf("gen.yield expects at most one argument")
				}
				value = args[0]
			}
			if err := g.emit(value); err != nil {
				return nil, err
			}
			return runtime.NilValue{}, nil
		},
	}

	closeFn := runtime.NativeFunctionValue{
		Name:  "__iterator_controller_close",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			g.close()
			return runtime.NilValue{}, nil
		},
	}

	g.control = &runtime.StructInstanceValue{
		Fields: map[string]runtime.Value{
			"yield": yieldFn,
			"close": closeFn,
		},
	}
	return g.control
}
