package parser

import "time"

// ModuleParsePhaseSample reports opt-in front-end phase timing for one source
// module. Normal parser users leave the observer unset.
type ModuleParsePhaseSample struct {
	SourceBytes int
	NativeParse time.Duration
	ASTMapping  time.Duration
}

// ModuleParsePhaseObserver receives one completed module parse sample.
type ModuleParsePhaseObserver func(ModuleParsePhaseSample)

// SetPhaseObserver installs an optional module parse observer.
func (p *ModuleParser) SetPhaseObserver(observer ModuleParsePhaseObserver) {
	if p == nil {
		return
	}
	p.phaseObserver = observer
}
