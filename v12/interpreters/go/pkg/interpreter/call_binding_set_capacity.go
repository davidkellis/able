package interpreter

import "able/interpreter-go/pkg/runtime"

const linearCallBindingDedupLimit = 8

func (bindings functionCallTypeBindingSet) distinctLen() int {
	total := len(bindings.explicit) + len(bindings.callLocal)
	if total == 0 {
		return 0
	}
	if total <= linearCallBindingDedupLimit {
		var seen [linearCallBindingDedupLimit]string
		count := 0
		count = appendDistinctCallBindingNamesLinear(seen[:], count, bindings.explicit)
		return appendDistinctCallBindingNamesLinear(seen[:], count, bindings.callLocal)
	}
	seen := make(map[string]struct{}, total)
	appendDistinctCallBindingNamesMap(seen, bindings.explicit)
	appendDistinctCallBindingNamesMap(seen, bindings.callLocal)
	return len(seen)
}

func (bindings functionCallTypeBindingSet) envValueCapacity(base int) int {
	if len(bindings.callLocal) == 0 {
		return base
	}
	if len(bindings.explicit) == 0 {
		return base + bindings.distinctLen()
	}
	total := base
	if len(bindings.callLocal)+len(bindings.explicit) <= linearCallBindingDedupLimit {
		var seen [linearCallBindingDedupLimit]string
		count := appendDistinctCallBindingNamesLinear(seen[:], 0, bindings.explicit)
		return total + countAdditionalCallBindingNamesLinear(seen[:], count, bindings.callLocal)
	}
	seen := make(map[string]struct{}, len(bindings.explicit)+len(bindings.callLocal))
	appendDistinctCallBindingNamesMap(seen, bindings.explicit)
	return total + countAdditionalCallBindingNamesMap(seen, bindings.callLocal)
}

func appendDistinctCallBindingNamesLinear(seen []string, count int, bindings []runtime.EnvironmentBinding) int {
	for _, binding := range bindings {
		if binding.Name == "" {
			continue
		}
		duplicate := false
		for idx := 0; idx < count; idx++ {
			if seen[idx] == binding.Name {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		seen[count] = binding.Name
		count++
	}
	return count
}

func countAdditionalCallBindingNamesLinear(seen []string, count int, bindings []runtime.EnvironmentBinding) int {
	additional := 0
	for _, binding := range bindings {
		if binding.Name == "" {
			continue
		}
		duplicate := false
		for idx := 0; idx < count; idx++ {
			if seen[idx] == binding.Name {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		seen[count] = binding.Name
		count++
		additional++
	}
	return additional
}

func appendDistinctCallBindingNamesMap(seen map[string]struct{}, bindings []runtime.EnvironmentBinding) {
	for _, binding := range bindings {
		if binding.Name == "" {
			continue
		}
		seen[binding.Name] = struct{}{}
	}
}

func countAdditionalCallBindingNamesMap(seen map[string]struct{}, bindings []runtime.EnvironmentBinding) int {
	additional := 0
	for _, binding := range bindings {
		if binding.Name == "" {
			continue
		}
		if _, ok := seen[binding.Name]; ok {
			continue
		}
		seen[binding.Name] = struct{}{}
		additional++
	}
	return additional
}
