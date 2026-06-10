//go:build !(js && wasm)

package interpreter

import (
	"fmt"
	"sync"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/typechecker"
)

type cachedProgramEvaluationState struct {
	nodeOriginsOnce sync.Once
	nodeOrigins     map[ast.Node]string

	prepareOnce sync.Once
	prepareErr  error
}

type cachedLoadedModuleBytecodeState struct {
	once    sync.Once
	program *bytecodeProgram
	err     error
}

var programEvaluationStateCache sync.Map // map[*driver.Program]*cachedProgramEvaluationState
var loadedModuleBytecodeCache sync.Map   // map[*driver.Module]*cachedLoadedModuleBytecodeState

func cachedProgramEvaluationEntry(program *driver.Program) *cachedProgramEvaluationState {
	if program == nil {
		return nil
	}
	if cached, ok := programEvaluationStateCache.Load(program); ok {
		return cached.(*cachedProgramEvaluationState)
	}
	entry := &cachedProgramEvaluationState{}
	actual, _ := programEvaluationStateCache.LoadOrStore(program, entry)
	return actual.(*cachedProgramEvaluationState)
}

func cachedProgramNodeOrigins(program *driver.Program) map[ast.Node]string {
	entry := cachedProgramEvaluationEntry(program)
	if entry == nil {
		return nil
	}
	entry.nodeOriginsOnce.Do(func() {
		entry.nodeOrigins = mergeNodeOrigins(program.Modules)
	})
	return entry.nodeOrigins
}

func cachedPrepareProgramForEvaluation(program *driver.Program) error {
	entry := cachedProgramEvaluationEntry(program)
	if entry == nil {
		return fmt.Errorf("typechecker: program is nil")
	}
	entry.prepareOnce.Do(func() {
		entry.prepareErr = typechecker.PrepareProgramForEvaluation(program)
	})
	return entry.prepareErr
}

func cachedLoadedModuleBytecodeProgram(i *Interpreter, module *driver.Module) (*bytecodeProgram, error) {
	if module == nil || module.AST == nil {
		return nil, fmt.Errorf("bytecode lowering module is nil")
	}
	entry := &cachedLoadedModuleBytecodeState{}
	actual, _ := loadedModuleBytecodeCache.LoadOrStore(module, entry)
	entry = actual.(*cachedLoadedModuleBytecodeState)
	entry.once.Do(func() {
		entry.program, entry.err = i.lowerModuleToBytecode(module.AST)
	})
	return entry.program, entry.err
}
