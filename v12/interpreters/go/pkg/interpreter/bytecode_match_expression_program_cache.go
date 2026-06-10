package interpreter

import "able/interpreter-go/pkg/ast"

type bytecodeMatchClauseProgramKind uint8

const (
	bytecodeMatchClauseProgramGuard bytecodeMatchClauseProgramKind = iota
	bytecodeMatchClauseProgramBody
)

type bytecodeMatchClausePrograms struct {
	guard    *bytecodeProgram
	body     *bytecodeProgram
	guardSet bool
	bodySet  bool
}

func (i *Interpreter) matchExpressionBytecodePrograms(expr *ast.MatchExpression) []bytecodeMatchClausePrograms {
	if i == nil || expr == nil {
		return nil
	}
	if i.envSingleThread {
		if programs, ok := i.matchExpressionBytecodeProgramsCache[expr]; ok && len(programs) == len(expr.Clauses) {
			return programs
		}
		programs := make([]bytecodeMatchClausePrograms, len(expr.Clauses))
		i.matchExpressionBytecodeProgramsCache[expr] = programs
		return programs
	}
	i.matchExpressionBytecodeProgramsCacheMu.RLock()
	programs, ok := i.matchExpressionBytecodeProgramsCache[expr]
	i.matchExpressionBytecodeProgramsCacheMu.RUnlock()
	if ok && len(programs) == len(expr.Clauses) {
		return programs
	}
	programs = make([]bytecodeMatchClausePrograms, len(expr.Clauses))
	i.matchExpressionBytecodeProgramsCacheMu.Lock()
	if cached, ok := i.matchExpressionBytecodeProgramsCache[expr]; ok && len(cached) == len(expr.Clauses) {
		i.matchExpressionBytecodeProgramsCacheMu.Unlock()
		return cached
	}
	i.matchExpressionBytecodeProgramsCache[expr] = programs
	i.matchExpressionBytecodeProgramsCacheMu.Unlock()
	return programs
}

func (i *Interpreter) matchExpressionClauseBytecodeProgram(programs []bytecodeMatchClausePrograms, idx int, expr ast.Expression, kind bytecodeMatchClauseProgramKind) (*bytecodeProgram, error) {
	if i == nil || expr == nil {
		return nil, nil
	}
	if idx < 0 || idx >= len(programs) {
		return i.lowerExpressionToBytecode(expr)
	}
	if i.envSingleThread {
		entry := &programs[idx]
		switch kind {
		case bytecodeMatchClauseProgramGuard:
			if entry.guardSet {
				return entry.guard, nil
			}
			program, err := i.lowerExpressionToBytecode(expr)
			if err != nil {
				return nil, err
			}
			entry.guard = program
			entry.guardSet = true
			return program, nil
		case bytecodeMatchClauseProgramBody:
			if entry.bodySet {
				return entry.body, nil
			}
			program, err := i.lowerExpressionToBytecode(expr)
			if err != nil {
				return nil, err
			}
			entry.body = program
			entry.bodySet = true
			return program, nil
		default:
			return i.lowerExpressionToBytecode(expr)
		}
	}
	if program, ok := i.lookupMatchExpressionClauseBytecodeProgram(programs, idx, kind); ok {
		return program, nil
	}
	program, err := i.lowerExpressionToBytecode(expr)
	if err != nil {
		return nil, err
	}
	return i.cacheMatchExpressionClauseBytecodeProgram(programs, idx, kind, program), nil
}

func (i *Interpreter) lookupMatchExpressionClauseBytecodeProgram(programs []bytecodeMatchClausePrograms, idx int, kind bytecodeMatchClauseProgramKind) (*bytecodeProgram, bool) {
	if i == nil || idx < 0 || idx >= len(programs) {
		return nil, false
	}
	i.matchExpressionBytecodeProgramsCacheMu.RLock()
	entry := programs[idx]
	i.matchExpressionBytecodeProgramsCacheMu.RUnlock()
	switch kind {
	case bytecodeMatchClauseProgramGuard:
		return entry.guard, entry.guardSet
	case bytecodeMatchClauseProgramBody:
		return entry.body, entry.bodySet
	default:
		return nil, false
	}
}

func (i *Interpreter) cacheMatchExpressionClauseBytecodeProgram(programs []bytecodeMatchClausePrograms, idx int, kind bytecodeMatchClauseProgramKind, program *bytecodeProgram) *bytecodeProgram {
	if i == nil || idx < 0 || idx >= len(programs) || program == nil {
		return program
	}
	i.matchExpressionBytecodeProgramsCacheMu.Lock()
	entry := &programs[idx]
	switch kind {
	case bytecodeMatchClauseProgramGuard:
		if entry.guardSet {
			program = entry.guard
		} else {
			entry.guard = program
			entry.guardSet = true
		}
	case bytecodeMatchClauseProgramBody:
		if entry.bodySet {
			program = entry.body
		} else {
			entry.body = program
			entry.bodySet = true
		}
	}
	i.matchExpressionBytecodeProgramsCacheMu.Unlock()
	return program
}
