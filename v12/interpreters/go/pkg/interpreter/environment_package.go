package interpreter

import "able/interpreter-go/pkg/runtime"

func (i *Interpreter) packageNameForEnvironment(env *runtime.Environment) string {
	if i == nil {
		return ""
	}
	if env == nil {
		return i.currentPackage
	}
	for current := env; current != nil; current = current.Parent() {
		if packageName, ok := i.packageNamesByEnv[current]; ok {
			return packageName
		}
	}
	return i.currentPackage
}
