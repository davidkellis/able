package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const compiledTestCacheLockName = ".compiled-test-cache.lock"

type compiledTestCacheFileLock struct {
	file  *os.File
	state compiledTestCachePlatformLock
}

func acquireCompiledTestCacheFileLock(
	root string,
	exclusive bool,
	nonblocking bool,
) (*compiledTestCacheFileLock, bool, error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, false, fmt.Errorf("compiled-test cache: create lock root: %w", err)
	}
	lockPath := filepath.Join(root, compiledTestCacheLockName)
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, false, fmt.Errorf("compiled-test cache: open lifecycle lock: %w", err)
	}
	state, acquired, err := lockCompiledTestCacheFile(file, exclusive, nonblocking)
	if err != nil {
		_ = file.Close()
		return nil, false, fmt.Errorf("compiled-test cache: acquire lifecycle lock: %w", err)
	}
	if !acquired {
		_ = file.Close()
		return nil, false, nil
	}
	return &compiledTestCacheFileLock{file: file, state: state}, true, nil
}

func (lock *compiledTestCacheFileLock) release() error {
	if lock == nil || lock.file == nil {
		return nil
	}
	unlockErr := unlockCompiledTestCacheFile(lock.file, lock.state)
	closeErr := lock.file.Close()
	lock.file = nil
	if unlockErr != nil {
		return fmt.Errorf("compiled-test cache: release lifecycle lock: %w", unlockErr)
	}
	if closeErr != nil {
		return fmt.Errorf("compiled-test cache: close lifecycle lock: %w", closeErr)
	}
	return nil
}
