package interpreter

import "able/interpreter-go/pkg/runtime"

type nativeBorrowCallArgScratch struct {
	inline [bytecodeInlinePreparedCallArgStorage]runtime.Value
	spill  []runtime.Value
}

func (s *nativeBorrowCallArgScratch) buffer(size int) []runtime.Value {
	if s == nil || size <= 0 {
		return nil
	}
	if size <= len(s.inline) {
		return s.inline[:size]
	}
	if cap(s.spill) < size {
		s.spill = make([]runtime.Value, size)
	} else {
		s.spill = s.spill[:size]
	}
	return s.spill
}

func (s *nativeBorrowCallArgScratch) reset() {
	if s == nil {
		return
	}
	clear(s.inline[:])
	if cap(s.spill) > bytecodeResolvedCallArgRetainLimit {
		s.spill = nil
		return
	}
	if len(s.spill) > 0 {
		clear(s.spill)
		s.spill = s.spill[:0]
	}
}
