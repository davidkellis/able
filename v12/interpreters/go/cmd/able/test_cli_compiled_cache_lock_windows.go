//go:build windows

package main

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

type compiledTestCachePlatformLock struct {
	overlapped windows.Overlapped
}

func lockCompiledTestCacheFile(
	file *os.File,
	exclusive bool,
	nonblocking bool,
) (compiledTestCachePlatformLock, bool, error) {
	state := compiledTestCachePlatformLock{}
	var flags uint32
	if exclusive {
		flags |= windows.LOCKFILE_EXCLUSIVE_LOCK
	}
	if nonblocking {
		flags |= windows.LOCKFILE_FAIL_IMMEDIATELY
	}
	err := windows.LockFileEx(
		windows.Handle(file.Fd()),
		flags,
		0,
		1,
		0,
		&state.overlapped,
	)
	if nonblocking && errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return compiledTestCachePlatformLock{}, false, nil
	}
	if err != nil {
		return compiledTestCachePlatformLock{}, false, err
	}
	return state, true, nil
}

func unlockCompiledTestCacheFile(
	file *os.File,
	state compiledTestCachePlatformLock,
) error {
	return windows.UnlockFileEx(
		windows.Handle(file.Fd()),
		0,
		1,
		0,
		&state.overlapped,
	)
}
