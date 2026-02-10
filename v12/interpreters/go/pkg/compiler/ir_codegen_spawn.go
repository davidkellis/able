package compiler

import (
	"fmt"
	"sort"

	"able/interpreter-go/pkg/ast"
)

func (e *irGoEmitter) resetFunctionState() {
	e.paramValues = make(map[*IRValue]struct{})
	e.callCallees = make(map[*IRValue]struct{})
	e.declaredVals = e.declaredVals[:0]
	e.patternSlots = make(map[ast.Node]*IRSlot)
}

func collectIRFunctions(root *IRFunction) []*IRFunction {
	seen := make(map[*IRFunction]struct{})
	ordered := make([]*IRFunction, 0)
	var walk func(fn *IRFunction)
	walk = func(fn *IRFunction) {
		if fn == nil {
			return
		}
		if _, ok := seen[fn]; ok {
			return
		}
		seen[fn] = struct{}{}
		ordered = append(ordered, fn)

		labels := make([]string, 0, len(fn.Blocks))
		for label := range fn.Blocks {
			labels = append(labels, label)
		}
		sort.Strings(labels)
		for _, label := range labels {
			block := fn.Blocks[label]
			if block == nil {
				continue
			}
			for _, instr := range block.Instructions {
				spawn, ok := instr.(*IRSpawn)
				if !ok || spawn == nil || spawn.Body == nil {
					if iter, ok := instr.(*IRIteratorLiteral); ok && iter != nil && iter.Body != nil {
						walk(iter.Body)
					}
					continue
				}
				walk(spawn.Body)
			}
		}
	}
	walk(root)
	return ordered
}

func (e *irGoEmitter) emitSpawn(node *IRSpawn) error {
	if node == nil || node.Value == nil || node.Error == nil || node.Body == nil {
		return fmt.Errorf("compiler: IR spawn missing body or outputs")
	}
	dest := e.nameForValue(node.Value)
	errDest := e.nameForValue(node.Error)
	args := make([]string, 0, len(node.Captures))
	for _, slot := range node.Captures {
		if slot == nil {
			continue
		}
		args = append(args, e.nameForSlot(slot))
	}
	bodyName := e.functionName(node.Body)
	taskVar := e.mangler.unique("spawn_task")
	fmt.Fprintf(&e.buf, "\t%s := func(_ *runtime.Environment) (runtime.Value, error) {\n", taskVar)
	fmt.Fprintf(&e.buf, "\t\treturn %s(rt", bodyName)
	for _, arg := range args {
		fmt.Fprintf(&e.buf, ", %s", arg)
	}
	fmt.Fprintf(&e.buf, ")\n\t}\n")
	fmt.Fprintf(&e.buf, "\t%s = __able_spawn(%s)\n", dest, taskVar)
	fmt.Fprintf(&e.buf, "\t%s = runtime.NilValue{}\n", errDest)
	return nil
}
