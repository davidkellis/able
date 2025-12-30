package interpreter

import (
	"errors"
	"fmt"
)

type exitSignal struct {
	code int
}

func (e exitSignal) Error() string {
	return fmt.Sprintf("exit %d", e.code)
}

// ExitCodeFromError returns the exit code if err is an exit signal.
func ExitCodeFromError(err error) (int, bool) {
	var sig exitSignal
	if errors.As(err, &sig) {
		return sig.code, true
	}
	return 0, false
}
