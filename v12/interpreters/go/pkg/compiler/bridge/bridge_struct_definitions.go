package bridge

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/runtime"
)

type structDefinitionCacheKey struct {
	env  *runtime.Environment
	name string
}

func (r *Runtime) StructDefinition(name string) (*runtime.StructDefinitionValue, error) {
	if r == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	r.mu.RLock()
	if def, ok := r.qualifiedStructs[name]; ok {
		r.mu.RUnlock()
		return def, nil
	}
	r.mu.RUnlock()
	env := r.currentEnv()
	if env == nil {
		return nil, fmt.Errorf("compiler bridge: missing global environment")
	}
	cacheKey := structCacheKey(env, name)
	r.mu.RLock()
	if def, ok := r.qualifiedStructs[name]; ok {
		r.mu.RUnlock()
		return def, nil
	}
	if def, ok := r.structs[cacheKey]; ok {
		r.mu.RUnlock()
		return def, nil
	}
	r.mu.RUnlock()
	aliases := []string{name}
	if idx := strings.LastIndex(strings.TrimSpace(name), "."); idx >= 0 && idx+1 < len(name) {
		if leaf := strings.TrimSpace(name[idx+1:]); leaf != "" && leaf != name {
			aliases = append(aliases, leaf)
		}
	}
	var aliasUsed string
	def, ok := env.StructDefinition(name)
	if !ok || def == nil {
		for _, alias := range aliases[1:] {
			if seeded, found := env.StructDefinition(alias); found && seeded != nil {
				def, ok = seeded, true
				aliasUsed = alias
				break
			}
		}
	}
	if (!ok || def == nil) && r.interp != nil {
		for _, alias := range aliases {
			if seeded, found := r.interp.LookupStructDefinition(alias); found && seeded != nil {
				def, ok = seeded, true
				aliasUsed = alias
				env.DefineStruct(name, seeded)
				if alias != "" && alias != name {
					env.DefineStruct(alias, seeded)
				}
				if seeded.Node != nil && seeded.Node.ID != nil {
					if canonical := strings.TrimSpace(seeded.Node.ID.Name); canonical != "" {
						if canonical != name {
							env.DefineStruct(canonical, seeded)
						}
						if alias != "" && alias != canonical {
							env.DefineStruct(alias, seeded)
						}
					}
				}
				break
			}
		}
	}
	if (!ok || def == nil) && r.interp != nil && env != r.interp.GlobalEnvironment() && r.globalLookupFallback() {
		if fallback := r.interp.GlobalEnvironment(); fallback != nil {
			for _, alias := range aliases {
				if alt, found := fallback.StructDefinition(alias); found && alt != nil {
					recordGlobalLookupFallback("struct_global", alias)
					def, ok = alt, true
					aliasUsed = alias
					break
				}
			}
		}
	}
	if (!ok || def == nil) && r.interp != nil && r.globalLookupFallback() {
		for _, alias := range aliases {
			if alt, found := r.interp.LookupStructDefinition(alias); found && alt != nil {
				recordGlobalLookupFallback("struct_registry", alias)
				def, ok = alt, true
				aliasUsed = alias
				break
			}
		}
	}
	if !ok || def == nil {
		return nil, fmt.Errorf("compiler bridge: struct %s not found", name)
	}
	r.mu.Lock()
	r.structs[cacheKey] = def
	if aliasUsed != "" && aliasUsed != name {
		r.structs[structCacheKey(env, aliasUsed)] = def
	}
	r.mu.Unlock()
	return def, nil
}

// StructDefinitionIn resolves a definition against an already-known
// environment. Generated static paths use this to avoid rediscovering the
// current goroutine solely to perform the same environment-scoped lookup.
func (r *Runtime) StructDefinitionIn(env *runtime.Environment, name string) (*runtime.StructDefinitionValue, error) {
	if r == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	r.mu.RLock()
	if def, ok := r.qualifiedStructs[name]; ok {
		r.mu.RUnlock()
		return def, nil
	}
	r.mu.RUnlock()
	if env == nil {
		return nil, fmt.Errorf("compiler bridge: missing global environment")
	}
	cacheKey := structCacheKey(env, name)
	r.mu.RLock()
	if def, ok := r.qualifiedStructs[name]; ok {
		r.mu.RUnlock()
		return def, nil
	}
	if def, ok := r.structs[cacheKey]; ok {
		r.mu.RUnlock()
		return def, nil
	}
	r.mu.RUnlock()
	aliases := []string{name}
	if idx := strings.LastIndex(strings.TrimSpace(name), "."); idx >= 0 && idx+1 < len(name) {
		if leaf := strings.TrimSpace(name[idx+1:]); leaf != "" && leaf != name {
			aliases = append(aliases, leaf)
		}
	}
	var aliasUsed string
	def, ok := env.StructDefinition(name)
	if !ok || def == nil {
		for _, alias := range aliases[1:] {
			if seeded, found := env.StructDefinition(alias); found && seeded != nil {
				def, ok = seeded, true
				aliasUsed = alias
				break
			}
		}
	}
	if (!ok || def == nil) && r.interp != nil {
		for _, alias := range aliases {
			if seeded, found := r.interp.LookupStructDefinition(alias); found && seeded != nil {
				def, ok = seeded, true
				aliasUsed = alias
				env.DefineStruct(name, seeded)
				if alias != "" && alias != name {
					env.DefineStruct(alias, seeded)
				}
				if seeded.Node != nil && seeded.Node.ID != nil {
					if canonical := strings.TrimSpace(seeded.Node.ID.Name); canonical != "" {
						if canonical != name {
							env.DefineStruct(canonical, seeded)
						}
						if alias != "" && alias != canonical {
							env.DefineStruct(alias, seeded)
						}
					}
				}
				break
			}
		}
	}
	if (!ok || def == nil) && r.interp != nil && env != r.interp.GlobalEnvironment() && r.globalLookupFallback() {
		if fallback := r.interp.GlobalEnvironment(); fallback != nil {
			for _, alias := range aliases {
				if alt, found := fallback.StructDefinition(alias); found && alt != nil {
					recordGlobalLookupFallback("struct_global", alias)
					def, ok = alt, true
					aliasUsed = alias
					break
				}
			}
		}
	}
	if (!ok || def == nil) && r.interp != nil && r.globalLookupFallback() {
		for _, alias := range aliases {
			if alt, found := r.interp.LookupStructDefinition(alias); found && alt != nil {
				recordGlobalLookupFallback("struct_registry", alias)
				def, ok = alt, true
				aliasUsed = alias
				break
			}
		}
	}
	if !ok || def == nil {
		return nil, fmt.Errorf("compiler bridge: struct %s not found", name)
	}
	r.mu.Lock()
	r.structs[cacheKey] = def
	if aliasUsed != "" && aliasUsed != name {
		r.structs[structCacheKey(env, aliasUsed)] = def
	}
	r.mu.Unlock()
	return def, nil
}

func (r *Runtime) UnionDefinition(name string) (*runtime.UnionDefinitionValue, error) {
	if r == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := r.currentEnv()
	if env == nil {
		return nil, fmt.Errorf("compiler bridge: missing global environment")
	}
	if val, err := env.Get(name); err == nil {
		if def, conv := toUnionDefinitionValue(val, name); conv == nil && def != nil {
			return def, nil
		}
	}
	if r.interp != nil && env != r.interp.GlobalEnvironment() && r.globalLookupFallback() {
		if fallback := r.interp.GlobalEnvironment(); fallback != nil {
			if val, err := fallback.Get(name); err == nil {
				if def, conv := toUnionDefinitionValue(val, name); conv == nil && def != nil {
					recordGlobalLookupFallback("union_global", name)
					return def, nil
				}
			}
		}
	}
	if r.interp != nil {
		def, ok := r.interp.LookupUnionDefinition(name)
		if ok && def != nil {
			return def, nil
		}
	}
	return nil, fmt.Errorf("compiler bridge: union %s not found", name)
}

func structCacheKey(env *runtime.Environment, name string) structDefinitionCacheKey {
	return structDefinitionCacheKey{env: env, name: name}
}

func toUnionDefinitionValue(val runtime.Value, name string) (*runtime.UnionDefinitionValue, error) {
	switch typed := val.(type) {
	case *runtime.UnionDefinitionValue:
		if typed == nil {
			return nil, fmt.Errorf("compiler bridge: %s is not a union definition", name)
		}
		return typed, nil
	case runtime.UnionDefinitionValue:
		copy := typed
		return &copy, nil
	default:
		return nil, fmt.Errorf("compiler bridge: %s is not a union definition", name)
	}
}
