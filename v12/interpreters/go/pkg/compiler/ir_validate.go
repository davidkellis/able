package compiler

import "fmt"

// ValidateIR ensures basic invariants for an IR function.
func ValidateIR(fn *IRFunction) error {
	if fn == nil {
		return fmt.Errorf("compiler: IR function is nil")
	}
	if fn.EntryLabel == "" {
		return fmt.Errorf("compiler: IR function missing entry label")
	}
	if _, ok := fn.Blocks[fn.EntryLabel]; !ok {
		return fmt.Errorf("compiler: IR entry block %q not found", fn.EntryLabel)
	}
	for label, block := range fn.Blocks {
		if block == nil {
			return fmt.Errorf("compiler: IR block %q is nil", label)
		}
		if block.Terminator == nil {
			return fmt.Errorf("compiler: IR block %q missing terminator", label)
		}
	}
	reachable, err := collectReachable(fn)
	if err != nil {
		return err
	}
	for label := range fn.Blocks {
		if !reachable[label] {
			return fmt.Errorf("compiler: IR block %q is unreachable", label)
		}
	}
	return nil
}

func collectReachable(fn *IRFunction) (map[string]bool, error) {
	reachable := make(map[string]bool, len(fn.Blocks))
	var visit func(label string) error
	visit = func(label string) error {
		if label == "" {
			return fmt.Errorf("compiler: IR terminator references empty label")
		}
		if reachable[label] {
			return nil
		}
		block, ok := fn.Blocks[label]
		if !ok || block == nil {
			return fmt.Errorf("compiler: IR references missing block %q", label)
		}
		reachable[label] = true
		switch term := block.Terminator.(type) {
		case *IRJump:
			return visit(term.Target)
		case *IRBranch:
			if err := visit(term.TrueLabel); err != nil {
				return err
			}
			return visit(term.FalseLabel)
		case *IRCheck:
			if err := visit(term.OkLabel); err != nil {
				return err
			}
			return visit(term.ErrLabel)
		case *IRReturn, *IRUnreachable:
			return nil
		default:
			return fmt.Errorf("compiler: unknown terminator type %T", term)
		}
	}
	if err := visit(fn.EntryLabel); err != nil {
		return nil, err
	}
	return reachable, nil
}
