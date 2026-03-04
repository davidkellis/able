package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (e *irGoEmitter) emitStringInterpolation(node *IRStringInterpolation) error {
	if node == nil || node.Dest == nil {
		return fmt.Errorf("compiler: IR string interpolation missing destination")
	}
	dest := e.nameForValue(node.Dest)
	builder := e.mangler.unique("str_builder")
	fmt.Fprintf(&e.buf, "\tvar %s strings.Builder\n", builder)
	for _, part := range node.Parts {
		expr, err := e.emitValueRef(part)
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\t%s.WriteString(__able_stringify(rt, %s))\n", builder, expr)
	}
	fmt.Fprintf(&e.buf, "\t%s = runtime.StringValue{Val: %s.String()}\n", dest, builder)
	return nil
}

func (e *irGoEmitter) emitDestructure(node *IRDestructure) error {
	if node == nil || node.Error == nil {
		return fmt.Errorf("compiler: IR destructure missing error destination")
	}
	origPatternSlots := e.patternSlots
	if len(node.Bindings) > 0 {
		merged := make(map[ast.Node]*IRSlot, len(origPatternSlots)+len(node.Bindings))
		for key, slot := range origPatternSlots {
			merged[key] = slot
		}
		for key, slot := range node.Bindings {
			if key == nil || slot == nil {
				continue
			}
			merged[key] = slot
		}
		e.patternSlots = merged
		defer func() { e.patternSlots = origPatternSlots }()
	}
	valueExpr, err := e.emitValueRef(node.Value)
	if err != nil {
		return err
	}
	errDest := e.nameForValue(node.Error)
	valueTemp := e.mangler.unique("pattern_value")
	fmt.Fprintf(&e.buf, "\t%s := %s\n", valueTemp, valueExpr)

	slotTemps := make(map[*IRSlot]string)
	slots, err := e.collectPatternSlots(node.Pattern)
	if err != nil {
		return err
	}
	for _, slot := range slots {
		if slot == nil {
			continue
		}
		temp := e.mangler.unique("bind_" + sanitizeIdent(slot.Name))
		slotTemps[slot] = temp
		fmt.Fprintf(&e.buf, "\tvar %s runtime.Value\n", temp)
	}

	okVar := e.mangler.unique("pattern_ok")
	errVar := e.mangler.unique("pattern_err")
	fmt.Fprintf(&e.buf, "\t%s := true\n", okVar)
	fmt.Fprintf(&e.buf, "\t%s := runtime.NilValue{}\n", errVar)

	if err := e.emitPattern(node.Pattern, valueTemp, okVar, errVar, slotTemps); err != nil {
		return err
	}

	fmt.Fprintf(&e.buf, "\tif %s {\n", okVar)
	for _, slot := range slots {
		if slot == nil {
			continue
		}
		temp := slotTemps[slot]
		fmt.Fprintf(&e.buf, "\t\t*%s = %s\n", e.nameForSlot(slot), temp)
	}
	fmt.Fprintf(&e.buf, "\t\t%s = runtime.NilValue{}\n", errDest)
	fmt.Fprintf(&e.buf, "\t} else {\n")
	fmt.Fprintf(&e.buf, "\t\t%s = %s\n", errDest, errVar)
	fmt.Fprintf(&e.buf, "\t}\n")
	return nil
}

func (e *irGoEmitter) emitTerminator(term IRTerminator) error {
	switch node := term.(type) {
	case *IRReturn:
		value, err := e.emitValueRef(node.Value)
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\treturn %s, nil\n", value)
		return nil
	case *IRJump:
		fmt.Fprintf(&e.buf, "\tgoto %s\n", e.nameForLabel(node.Target))
		return nil
	case *IRBranch:
		cond, err := e.emitValueRef(node.Condition)
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\tif __able_truthy(rt, %s) {\n", cond)
		fmt.Fprintf(&e.buf, "\t\tgoto %s\n", e.nameForLabel(node.TrueLabel))
		fmt.Fprintf(&e.buf, "\t}\n")
		fmt.Fprintf(&e.buf, "\tgoto %s\n", e.nameForLabel(node.FalseLabel))
		return nil
	case *IRCheck:
		errValue, err := e.emitValueRef(node.Error)
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\tif __able_is_nil(%s) {\n", errValue)
		fmt.Fprintf(&e.buf, "\t\tgoto %s\n", e.nameForLabel(node.OkLabel))
		fmt.Fprintf(&e.buf, "\t}\n")
		fmt.Fprintf(&e.buf, "\tgoto %s\n", e.nameForLabel(node.ErrLabel))
		return nil
	case *IRUnreachable:
		fmt.Fprintf(&e.buf, "\treturn runtime.NilValue{}, fmt.Errorf(\"unreachable\")\n")
		return nil
	default:
		return fmt.Errorf("compiler: unsupported IR terminator %T", term)
	}
}
