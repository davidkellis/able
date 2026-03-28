package compiler

import "fmt"

func (g *generator) inlineRuntimeEnvSwapLinesForPackage(pkgName string) []string {
	if g == nil {
		return nil
	}
	envVar, ok := g.packageEnvVar(pkgName)
	if !ok || envVar == "" {
		return nil
	}
	return []string{
		fmt.Sprintf("if __able_runtime != nil && %s != nil {", envVar),
		fmt.Sprintf("if __able_prev_env, __able_swapped_env := bridge.SwapEnvIfNeeded(__able_runtime, %s); __able_swapped_env {", envVar),
		"defer bridge.RestoreEnvIfNeeded(__able_runtime, __able_prev_env, __able_swapped_env)",
		"}",
		"}",
	}
}
