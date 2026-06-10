package interpreter

import (
	"fmt"
	"sync"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

var generatorControllerStructDefinition = &runtime.StructDefinitionValue{
	Node: ast.StructDef("__able_iterator_controller", []*ast.StructFieldDefinition{
		ast.FieldDef(nil, "yield"),
		ast.FieldDef(nil, "close"),
		ast.FieldDef(nil, "stop"),
	}, ast.StructKindNamed, nil, nil, true),
}

type generatorResult struct {
	value runtime.RawValue
	done  bool
	err   error
}

type generatorInstance struct {
	interpreter          *Interpreter
	env                  *runtime.Environment
	body                 []ast.Statement
	bytecode             *bytecodeProgram
	bytecodeSlotArg      runtime.Value
	bytecodeSlotArgCount int

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

type generatorStopSignal struct{}

func (generatorStopSignal) Error() string {
	return "generator stopped"
}

func newGeneratorInstance(i *Interpreter, env *runtime.Environment, body []ast.Statement) *generatorInstance {
	return newGeneratorInstanceWithBytecode(i, env, body, nil)
}

func newGeneratorInstanceWithBytecode(i *Interpreter, env *runtime.Environment, body []ast.Statement, program *bytecodeProgram) *generatorInstance {
	return &generatorInstance{
		interpreter: i,
		env:         env,
		body:        body,
		bytecode:    program,
		requests:    make(chan struct{}),
		results:     make(chan generatorResult, 1),
	}
}

func (g *generatorInstance) next() (runtime.Value, bool, error) {
	value, done, err := g.nextRaw()
	if err != nil {
		return nil, done, err
	}
	return value.Materialize(), done, nil
}

func (g *generatorInstance) nextRaw() (runtime.RawValue, bool, error) {
	g.mu.Lock()
	if g.busy {
		g.mu.Unlock()
		return runtime.RawValue{}, true, fmt.Errorf("iterator.next re-entered while suspended at yield")
	}
	if g.closed {
		g.mu.Unlock()
		return runtime.NewRawValue(runtime.IteratorEnd), true, nil
	}
	if g.done {
		err := g.err
		g.mu.Unlock()
		if err != nil {
			return runtime.RawValue{}, true, err
		}
		return runtime.NewRawValue(runtime.IteratorEnd), true, nil
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
			return runtime.RawValue{}, true, err
		}
		return runtime.NewRawValue(runtime.IteratorEnd), true, nil
	}
	if res.err != nil {
		g.done = true
		g.err = res.err
		g.mu.Unlock()
		return runtime.RawValue{}, true, res.err
	}
	if res.done {
		g.done = true
		g.mu.Unlock()
		return runtime.NewRawValue(runtime.IteratorEnd), true, nil
	}
	g.mu.Unlock()
	if res.value.Kind() == runtime.RawValueMaterialized && res.value.Value() == nil {
		return runtime.NewRawValue(runtime.NilValue{}), false, nil
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
		case generatorStopSignal:
			g.mu.Lock()
			g.done = true
			g.mu.Unlock()
			g.results <- generatorResult{done: true}
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
	if g.bytecode != nil {
		vm := g.interpreter.acquireBytecodeVM(g.env)
		defer g.interpreter.releaseBytecodeVM(vm)
		g.prepareBytecodeSlotFrame(vm)
		_, err := vm.run(g.bytecode)
		return err
	}
	for _, stmt := range g.body {
		if _, err := g.interpreter.evaluateStatement(stmt, g.env); err != nil {
			return err
		}
	}
	return nil
}

func (g *generatorInstance) prepareBytecodeSlotFrame(vm *bytecodeVM) {
	if g == nil || vm == nil || g.bytecode == nil || g.bytecode.frameLayout == nil {
		return
	}
	layout := g.bytecode.frameLayout
	if layout.slotCount <= 0 {
		return
	}
	slots := vm.acquireSlotFrame(layout.slotCount)
	for idx := 0; idx < layout.paramSlots && idx < g.bytecodeSlotArgCount; idx++ {
		slots[idx] = g.bytecodeSlotArg
	}
	vm.slots = slots
}

func (g *generatorInstance) emit(value runtime.Value) error {
	return g.emitRaw(runtime.NewRawValue(value))
}

func (g *generatorInstance) emitRaw(value runtime.RawValue) error {
	g.results <- generatorResult{value: value}
	if !g.awaitRequest() {
		return generatorStopSignal{}
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

var (
	generatorControllerYieldNative = runtime.NativeFunctionValue{
		Name:        "__iterator_controller_yield",
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			_, generator, err := generatorControllerFromArgs(args, "gen.yield")
			if err != nil {
				return nil, err
			}
			if len(args) > 2 {
				return nil, fmt.Errorf("gen.yield expects at most one argument")
			}
			value := runtime.Value(runtime.NilValue{})
			if len(args) == 2 {
				value = args[1]
			}
			if err := generator.emit(value); err != nil {
				return nil, err
			}
			return runtime.NilValue{}, nil
		},
		RawImpl: func(_ *runtime.NativeCallContext, args []runtime.RawValue) (runtime.RawValue, error) {
			_, generator, err := generatorControllerFromRawArgs(args, "gen.yield")
			if err != nil {
				return runtime.RawValue{}, err
			}
			if len(args) > 2 {
				return runtime.RawValue{}, fmt.Errorf("gen.yield expects at most one argument")
			}
			value := runtime.NewRawValue(runtime.NilValue{})
			if len(args) == 2 {
				value = args[1]
			}
			if err := generator.emitRaw(value); err != nil {
				return runtime.RawValue{}, err
			}
			return runtime.NewRawValue(runtime.NilValue{}), nil
		},
	}
	generatorControllerCloseNative = runtime.NativeFunctionValue{
		Name:        "__iterator_controller_close",
		Arity:       0,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			_, generator, err := generatorControllerFromArgs(args, "gen.close")
			if err != nil {
				return nil, err
			}
			if len(args) != 1 {
				return nil, fmt.Errorf("gen.close expects no arguments")
			}
			generator.close()
			return runtime.NilValue{}, nil
		},
	}
	generatorControllerStopNative = runtime.NativeFunctionValue{
		Name:        "__iterator_controller_stop",
		Arity:       0,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			_, generator, err := generatorControllerFromArgs(args, "gen.stop")
			if err != nil {
				return nil, err
			}
			if len(args) != 1 {
				return nil, fmt.Errorf("gen.stop expects no arguments")
			}
			generator.close()
			return runtime.NilValue{}, generatorStopSignal{}
		},
	}
)

func generatorControllerFromArgs(args []runtime.Value, name string) (*runtime.StructInstanceValue, *generatorInstance, error) {
	if len(args) == 0 {
		return nil, nil, fmt.Errorf("%s missing generator controller receiver", name)
	}
	controller, ok := args[0].(*runtime.StructInstanceValue)
	if !ok || controller == nil {
		return nil, nil, fmt.Errorf("%s receiver is not a generator controller", name)
	}
	generator, ok := controller.Native.(*generatorInstance)
	if !ok || generator == nil {
		return nil, nil, fmt.Errorf("%s receiver is not a live generator controller", name)
	}
	return controller, generator, nil
}

func generatorControllerFromRawArgs(args []runtime.RawValue, name string) (*runtime.StructInstanceValue, *generatorInstance, error) {
	if len(args) == 0 {
		return nil, nil, fmt.Errorf("%s missing generator controller receiver", name)
	}
	controller, ok := args[0].Materialize().(*runtime.StructInstanceValue)
	if !ok || controller == nil {
		return nil, nil, fmt.Errorf("%s receiver is not a generator controller", name)
	}
	generator, ok := controller.Native.(*generatorInstance)
	if !ok || generator == nil {
		return nil, nil, fmt.Errorf("%s receiver is not a live generator controller", name)
	}
	return controller, generator, nil
}

func (g *generatorInstance) controllerValue() runtime.Value {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.control != nil {
		return g.control
	}

	controller, fields := runtime.NewStructInstancePositionalSized(generatorControllerStructDefinition, 3, nil)
	controller.Native = g
	fields[0] = runtime.NativeBoundMethodValue{Receiver: controller, Method: generatorControllerYieldNative}
	fields[1] = runtime.NativeBoundMethodValue{Receiver: controller, Method: generatorControllerCloseNative}
	fields[2] = runtime.NativeBoundMethodValue{Receiver: controller, Method: generatorControllerStopNative}
	g.control = controller
	return controller
}
