package compiler

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type loweringFacadeAudit struct {
	pattern string
	allowed map[string]struct{}
}

func loweringFacadeAllowed(names ...string) map[string]struct{} {
	allowed := make(map[string]struct{}, len(names))
	for _, name := range names {
		allowed[name] = struct{}{}
	}
	return allowed
}

func TestCompilerLoweringFacadeSourceAudit(t *testing.T) {
	root := filepath.Join(repositoryRoot(), "v12", "interpreters", "go", "pkg", "compiler")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read compiler dir: %v", err)
	}

	audits := []loweringFacadeAudit{
		{pattern: `\btypeExprInContext\(`, allowed: loweringFacadeAllowed("generator_context.go", "generator_lowering_types.go")},
		{pattern: `\bmapTypeExpressionInContext\(`, allowed: loweringFacadeAllowed("generator_context.go", "generator_lowering_types.go")},
		{pattern: `\bmapTypeExpressionInPackage\(`, allowed: loweringFacadeAllowed("generator_context.go", "generator_types.go", "generator_lowering_types.go")},
		{pattern: `\bjoinResultType\(`, allowed: loweringFacadeAllowed("generator_join_types.go", "generator_lowering_patterns.go")},
		{pattern: `\bjoinResultTypeFromBranches\(`, allowed: loweringFacadeAllowed("generator_join_types.go", "generator_lowering_patterns.go")},
		{pattern: `\bjoinResultTypeAllowNil\(`, allowed: loweringFacadeAllowed("generator_join_types.go", "generator_lowering_patterns.go")},
		{pattern: `\bstaticTypedPatternCompatible\(`, allowed: loweringFacadeAllowed("generator_match.go", "generator_lowering_patterns.go")},
		{pattern: `\bruntimeTypeCheckForTypeExpression\(`, allowed: loweringFacadeAllowed("generator_match_runtime_types.go", "generator_lowering_patterns.go")},
		{pattern: `\bnativeUnionPatternMemberType\(`, allowed: loweringFacadeAllowed("generator_native_unions.go", "generator_lowering_patterns.go")},
		{pattern: `\bcontrolTransferLines\(`, allowed: loweringFacadeAllowed("generator_control_results.go", "generator_lowering_control.go")},
		{pattern: `\bcontrolCheckLines\(`, allowed: loweringFacadeAllowed("generator_control_results.go", "generator_lowering_control.go")},
		{pattern: `\bruntimeValueLines\(`, allowed: loweringFacadeAllowed("generator_value_conversions.go", "generator_lowering_boundaries.go")},
		{pattern: `\bexpectRuntimeValueExprLines\(`, allowed: loweringFacadeAllowed("generator_value_conversions.go", "generator_lowering_boundaries.go")},
		{pattern: `\bnativeUnionWrapLines\(`, allowed: loweringFacadeAllowed("generator_exprs_helpers.go", "generator_lowering_boundaries.go")},
		{pattern: `\bnativeInterfaceWrapLines\(`, allowed: loweringFacadeAllowed("generator_native_interface_coercions.go", "generator_lowering_boundaries.go")},
		{pattern: `\bnativeCallableWrapLines\(`, allowed: loweringFacadeAllowed("generator_exprs_helpers.go", "generator_lowering_boundaries.go")},
		{pattern: `\bcoerceExpectedStaticExpr\(`, allowed: loweringFacadeAllowed("generator_exprs.go", "generator_lowering_boundaries.go")},
		{pattern: `\bcompileFunctionCall\(`, allowed: loweringFacadeAllowed("generator_exprs_calls_lambda.go", "generator_lowering_dispatch.go")},
		{pattern: `\bcompileMemberAccess\(`, allowed: loweringFacadeAllowed("generator_collections.go", "generator_lowering_dispatch.go")},
		{pattern: `\bcompileIndexExpression\(`, allowed: loweringFacadeAllowed("generator_collections_static_array_access.go", "generator_lowering_dispatch.go")},
		{pattern: `\bresolveStaticMethodCall\(`, allowed: loweringFacadeAllowed("generator_specialized_impl_calls.go", "generator_lowering_dispatch.go")},
		{pattern: `\bcompileResolvedMethodCall\(`, allowed: loweringFacadeAllowed("generator_specialized_impl_calls.go", "generator_lowering_dispatch.go")},
		{pattern: `\bcompileNativeInterfaceMethodCall\(`, allowed: loweringFacadeAllowed("generator_native_interface_calls.go", "generator_lowering_dispatch.go")},
		{pattern: `\bcompileNativeInterfaceGenericMethodCall\(`, allowed: loweringFacadeAllowed("generator_native_interface_generic_calls.go", "generator_lowering_dispatch.go")},
	}

	compiled := make([]*regexp.Regexp, len(audits))
	for idx, audit := range audits {
		compiled[idx] = regexp.MustCompile(audit.pattern)
	}

	var failures []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(root, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		text := string(data)
		for idx, audit := range audits {
			if !compiled[idx].MatchString(text) {
				continue
			}
			if _, ok := audit.allowed[name]; ok {
				continue
			}
			failures = append(failures, name+": "+audit.pattern)
		}
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		t.Fatalf("source audit found lowering helper bypasses:\n%s", strings.Join(failures, "\n"))
	}
}
