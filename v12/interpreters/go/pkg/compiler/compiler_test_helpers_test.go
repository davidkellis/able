package compiler

import (
	"os"
	"path/filepath"
	"strings"
)

const compilerExecGocacheEnv = "ABLE_COMPILER_EXEC_GOCACHE"

func compilerExecGocache(moduleRoot string) string {
	if value := strings.TrimSpace(os.Getenv(compilerExecGocacheEnv)); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("GOCACHE")); value != "" {
		return value
	}
	if moduleRoot == "" {
		return ""
	}
	return filepath.Join(moduleRoot, ".gocache")
}

func withEnv(env []string, key, value string) []string {
	if value == "" {
		return env
	}
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			continue
		}
		out = append(out, entry)
	}
	out = append(out, prefix+value)
	return out
}
