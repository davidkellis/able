package compiler

import "strings"

func (g *generator) compiledBodyName(info *functionInfo) string {
	if info == nil {
		return ""
	}
	return "__able_compiled_" + info.GoName
}

func (g *generator) compiledEntryName(info *functionInfo) string {
	if info == nil {
		return ""
	}
	return "__able_compiled_entry_" + info.GoName
}

func (g *generator) compiledContextBodyName(info *functionInfo) string {
	return compiledContextName(g.compiledBodyName(info))
}

func (g *generator) compiledContextEntryName(info *functionInfo) string {
	return compiledContextName(g.compiledEntryName(info))
}

func (g *generator) compiledCallTargetNameForPackage(callerPackage string, calleePackage string, goName string) string {
	goName = strings.TrimSpace(goName)
	if goName == "" {
		return ""
	}
	callerPackage = strings.TrimSpace(callerPackage)
	calleePackage = strings.TrimSpace(calleePackage)
	if (callerPackage != "" && callerPackage == calleePackage) || g.compiledFunctionEnvironmentIndependentByGoName(goName) {
		return "__able_compiled_" + goName
	}
	return "__able_compiled_entry_" + goName
}

// compiledCallTargetName returns the function name that a compiled caller
// should target directly. Same-package calls and imported functions proven not
// to consult package environment state can jump straight to the raw body.
func (g *generator) compiledCallTargetName(callerPackage string, info *functionInfo) string {
	if info == nil {
		return ""
	}
	return g.compiledCallTargetNameForPackage(callerPackage, info.Package, info.GoName)
}
