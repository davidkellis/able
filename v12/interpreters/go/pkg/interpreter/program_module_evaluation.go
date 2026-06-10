//go:build !(js && wasm)

package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) evaluateLoadedProgramModule(module *driver.Module) (runtime.Value, *runtime.Environment, error) {
	if module == nil || module.AST == nil {
		return nil, nil, fmt.Errorf("interpreter: module is nil")
	}
	var program *bytecodeProgram
	if i.execMode == execModeBytecode {
		cached, err := cachedLoadedModuleBytecodeProgram(i, module)
		if err != nil {
			return nil, nil, err
		}
		program = cached
	}
	return i.evaluateModuleWithProgram(module.AST, program)
}
