package driver

import (
	"time"

	"able/interpreter-go/pkg/parser"
)

// LoaderPhase identifies an opt-in source-loader subphase.
type LoaderPhase string

const (
	LoaderPhaseNativeParse      LoaderPhase = "native_parse"
	LoaderPhaseASTMapping       LoaderPhase = "ast_mapping"
	LoaderPhaseOriginAnnotation LoaderPhase = "origin_annotation"
)

// LoaderPhaseSample reports one source-loader subphase sample.
type LoaderPhaseSample struct {
	Phase       LoaderPhase
	Duration    time.Duration
	SourceBytes int
}

// LoaderPhaseObserver receives opt-in source-loader subphase samples.
type LoaderPhaseObserver func(LoaderPhaseSample)

// SetPhaseObserver installs source-loader timing diagnostics. Normal loader
// users leave this unset, which also disables the parser's timing calls.
func (l *Loader) SetPhaseObserver(observer LoaderPhaseObserver) {
	if l == nil {
		return
	}
	l.phaseObserver = observer
	if l.parser == nil {
		return
	}
	if observer == nil {
		l.parser.SetPhaseObserver(nil)
		return
	}
	l.parser.SetPhaseObserver(func(sample parser.ModuleParsePhaseSample) {
		observer(LoaderPhaseSample{
			Phase:       LoaderPhaseNativeParse,
			Duration:    sample.NativeParse,
			SourceBytes: sample.SourceBytes,
		})
		observer(LoaderPhaseSample{
			Phase:       LoaderPhaseASTMapping,
			Duration:    sample.ASTMapping,
			SourceBytes: sample.SourceBytes,
		})
	})
}
