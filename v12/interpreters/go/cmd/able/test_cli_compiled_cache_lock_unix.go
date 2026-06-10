//go:build !windows

package main

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

type compiledTestCachePlatformLock struct{}

func lockCompiledTestCacheFile(
	file *os.File,
	exclusive bool,
	nonblocking bool,
) (compiledTestCachePlatformLock, bool, error) {
	operation := unix.LOCK_SH
	if exclusive {
		operation = unix.LOCK_EX
	}
	if nonblocking {
		operation |= unix.LOCK_NB
	}
	err := unix.Flock(int(file.Fd()), operation)
	if nonblocking && (errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN)) {
		return compiledTestCachePlatformLock{}, false, nil
	}
	if err != nil {
		return compiledTestCachePlatformLock{}, false, err
	}
	return compiledTestCachePlatformLock{}, true, nil
}

func unlockCompiledTestCacheFile(
	file *os.File,
	_ compiledTestCachePlatformLock,
) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_UN)
}
