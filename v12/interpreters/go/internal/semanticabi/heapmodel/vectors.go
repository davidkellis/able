package heapmodel

import (
	"encoding/hex"
	"fmt"
	"strings"

	"able/interpreter-go/internal/semanticabi"
)

type VectorResult struct {
	Name   string `json:"name"`
	Detail string `json:"detail"`
}

type Report struct {
	Schema                   string         `json:"schema"`
	Decision                 string         `json:"decision"`
	ManifestIdentity         string         `json:"manifest_identity"`
	RuntimeKinds             int            `json:"runtime_kinds"`
	SharedHeapLayouts        int            `json:"shared_heap_layouts"`
	InternalHeapLayouts      int            `json:"internal_heap_layouts"`
	HostLayouts              int            `json:"host_layouts"`
	MutableHeapLayouts       int            `json:"mutable_heap_layouts"`
	ImmutableHeapLayouts     int            `json:"immutable_heap_layouts"`
	DeterministicVectorCount int            `json:"deterministic_vector_count"`
	Vectors                  []VectorResult `json:"vectors"`
	Gates                    ContractGates  `json:"gates"`
	NextLane                 string         `json:"next_lane"`
	Exclusions               []string       `json:"exclusions"`
}

type ContractGates struct {
	ExhaustiveSharedLayouts bool `json:"exhaustive_shared_layouts"`
	ExhaustiveHostLayouts   bool `json:"exhaustive_host_layouts"`
	DescriptorDriven        bool `json:"descriptor_driven"`
	GenerationChecked       bool `json:"generation_checked"`
	CycleSafeTracing        bool `json:"cycle_safe_tracing"`
	HostHeldRoots           bool `json:"host_held_roots"`
	ProductionChange        bool `json:"production_change"`
}

func RunConformanceVectors() (Report, error) {
	vectors := []struct {
		name string
		run  func() (string, error)
	}{
		{"alias-mutation", vectorAliasMutation},
		{"immutable-rejection", vectorImmutableRejection},
		{"unrooted-cycle", vectorUnrootedCycle},
		{"closure-environment-cycle", vectorClosureCycle},
		{"stale-generation", vectorStaleGeneration},
		{"host-held-root", vectorHostHeldRoot},
		{"interface-error-iterator-chain", vectorSemanticChain},
		{"root-frame-lifetime", vectorRootFrameLifetime},
	}
	report := Report{
		Schema:           "able.semanticabi.heap-conformance.v1",
		Decision:         "retain-shared-value-heap-conformance-contract",
		ManifestIdentity: hex.EncodeToString(semanticabi.ManifestIdentity[:]),
		RuntimeKinds:     len(semanticabi.KindManifest),
		HostLayouts:      len(semanticabi.HostLayoutManifest),
		Gates: ContractGates{
			ExhaustiveSharedLayouts: true, ExhaustiveHostLayouts: true,
			DescriptorDriven: true, GenerationChecked: true, CycleSafeTracing: true,
			HostHeldRoots: true, ProductionChange: false,
		},
		NextLane: "shared-value-heap-go-binding-conformance",
		Exclusions: []string{
			"production runtime migration", "foreign heap", "cgo runtime", "JIT or backend",
			"executable memory", "benchmark branch", "named-container or non-primitive nominal special case", "WASM",
		},
	}
	for _, layout := range semanticabi.ObjectLayoutManifest {
		if layout.RuntimeTag == 0 {
			report.InternalHeapLayouts++
		} else {
			report.SharedHeapLayouts++
		}
		if layout.Mutability == semanticabi.LayoutMutable {
			report.MutableHeapLayouts++
		} else {
			report.ImmutableHeapLayouts++
		}
	}
	for _, vector := range vectors {
		detail, err := vector.run()
		if err != nil {
			return Report{}, fmt.Errorf("heap conformance vector %s: %w", vector.name, err)
		}
		report.Vectors = append(report.Vectors, VectorResult{Name: vector.name, Detail: detail})
	}
	report.DeterministicVectorCount = len(report.Vectors)
	return report, nil
}

func vectorAliasMutation() (string, error) {
	h := New()
	definition, err := allocate(h, semanticabi.LayoutStructDefinition, func(fields []FieldValue) error {
		return SetBytes(semanticabi.LayoutStructDefinition, fields, "field_names", []byte("value"))
	})
	if err != nil {
		return "", err
	}
	one := immediateInteger(1)
	instance, err := allocate(h, semanticabi.LayoutStructInstance, func(fields []FieldValue) error {
		if err := SetObjects(semanticabi.LayoutStructInstance, fields, "definition", definition); err != nil {
			return err
		}
		return SetCells(semanticabi.LayoutStructInstance, fields, "fields", one)
	})
	if err != nil {
		return "", err
	}
	cell, err := h.Cell(semanticabi.TagKindStructInstance, instance)
	if err != nil {
		return "", err
	}
	frame := h.OpenRootFrame(cell, cell)
	object, _ := h.Resolve(instance)
	fields := object.Fields
	if err := SetCells(semanticabi.LayoutStructInstance, fields, "fields", immediateInteger(2)); err != nil {
		return "", err
	}
	if err := h.Mutate(instance, fields); err != nil {
		return "", err
	}
	resolved, err := h.Resolve(semanticabi.ObjectID(cell.Payload))
	if err != nil || resolved.Revision != 1 {
		return "", fmt.Errorf("alias did not observe revision 1")
	}
	collection, err := h.Collect()
	if err != nil || collection.Reachable != 2 {
		return "", fmt.Errorf("alias graph reachability = %+v: %v", collection, err)
	}
	_ = h.CloseRootFrame(frame)
	return "two cells preserve one identity; revision=1; reachable=2", nil
}

func vectorImmutableRejection() (string, error) {
	h := New()
	id, err := allocate(h, semanticabi.LayoutString, func(fields []FieldValue) error {
		return SetBytes(semanticabi.LayoutString, fields, "utf8", []byte("able"))
	})
	if err != nil {
		return "", err
	}
	object, _ := h.Resolve(id)
	err = h.Mutate(id, object.Fields)
	if err == nil || !strings.Contains(err.Error(), "immutable") {
		return "", fmt.Errorf("immutable mutation was not rejected: %v", err)
	}
	return "String mutation rejected by layout policy", nil
}

func vectorUnrootedCycle() (string, error) {
	h := New()
	environment, err := allocate(h, semanticabi.LayoutEnvironment, nil)
	if err != nil {
		return "", err
	}
	object, _ := h.Resolve(environment)
	if err := SetObjects(semanticabi.LayoutEnvironment, object.Fields, "parent", environment); err != nil {
		return "", err
	}
	if err := h.Mutate(environment, object.Fields); err != nil {
		return "", err
	}
	collection, err := h.Collect()
	if err != nil || collection.Collected != 1 || h.Stats().LiveObjects != 0 {
		return "", fmt.Errorf("cycle collection = %+v, stats=%+v: %v", collection, h.Stats(), err)
	}
	return "self-cycle collected with no roots", nil
}

func vectorClosureCycle() (string, error) {
	h := New()
	environment, err := allocate(h, semanticabi.LayoutEnvironment, nil)
	if err != nil {
		return "", err
	}
	function, err := allocate(h, semanticabi.LayoutFunction, func(fields []FieldValue) error {
		return SetObjects(semanticabi.LayoutFunction, fields, "environment", environment)
	})
	if err != nil {
		return "", err
	}
	functionCell, _ := h.Cell(semanticabi.TagKindFunction, function)
	binding, err := allocate(h, semanticabi.LayoutBindingCell, func(fields []FieldValue) error {
		if err := SetBytes(semanticabi.LayoutBindingCell, fields, "name", []byte("recursive")); err != nil {
			return err
		}
		return SetCells(semanticabi.LayoutBindingCell, fields, "value", functionCell)
	})
	if err != nil {
		return "", err
	}
	envObject, _ := h.Resolve(environment)
	if err := SetObjects(semanticabi.LayoutEnvironment, envObject.Fields, "bindings", binding); err != nil {
		return "", err
	}
	if err := h.Mutate(environment, envObject.Fields); err != nil {
		return "", err
	}
	frame := h.OpenRootFrame(functionCell)
	first, err := h.Collect()
	if err != nil || first.Reachable != 3 {
		return "", fmt.Errorf("rooted closure reachability = %+v: %v", first, err)
	}
	_ = h.CloseRootFrame(frame)
	second, err := h.Collect()
	if err != nil || second.Collected != 3 {
		return "", fmt.Errorf("released closure collection = %+v: %v", second, err)
	}
	return "function/environment/binding cycle survives root then all 3 collect", nil
}

func vectorStaleGeneration() (string, error) {
	h := New()
	oldID, err := allocate(h, semanticabi.LayoutString, nil)
	if err != nil {
		return "", err
	}
	if _, err := h.Collect(); err != nil {
		return "", err
	}
	newID, err := allocate(h, semanticabi.LayoutString, nil)
	if err != nil {
		return "", err
	}
	if oldID.Index() != newID.Index() || newID.Generation() != oldID.Generation()+1 {
		return "", fmt.Errorf("slot reuse old=%d:%d new=%d:%d", oldID.Index(), oldID.Generation(), newID.Index(), newID.Generation())
	}
	if _, err := h.Resolve(oldID); err == nil || !strings.Contains(err.Error(), "stale") {
		return "", fmt.Errorf("old generation did not fail stale: %v", err)
	}
	return fmt.Sprintf("slot %d reused at generation %d; generation %d rejected", newID.Index(), newID.Generation(), oldID.Generation()), nil
}

func vectorHostHeldRoot() (string, error) {
	h := New()
	array, err := allocate(h, semanticabi.LayoutArray, nil)
	if err != nil {
		return "", err
	}
	arrayCell, _ := h.Cell(semanticabi.TagKindArray, array)
	future, err := h.Hosts().Register(semanticabi.TagKindFuture, arrayCell)
	if err != nil {
		return "", err
	}
	if err := h.Hosts().Cancel(future); err != nil {
		return "", err
	}
	first, err := h.Collect()
	if err != nil || first.Reachable != 1 {
		return "", fmt.Errorf("host-held root collection = %+v: %v", first, err)
	}
	if err := h.Hosts().Release(future); err != nil {
		return "", err
	}
	second, err := h.Collect()
	if err != nil || second.Collected != 1 {
		return "", fmt.Errorf("released host root collection = %+v: %v", second, err)
	}
	return "canceled Future retains result until registry release", nil
}

func vectorSemanticChain() (string, error) {
	h := New()
	array, err := allocate(h, semanticabi.LayoutArray, nil)
	if err != nil {
		return "", err
	}
	arrayCell, _ := h.Cell(semanticabi.TagKindArray, array)
	interfaceDefinition, err := allocate(h, semanticabi.LayoutInterfaceDefinition, nil)
	if err != nil {
		return "", err
	}
	interfaceValue, err := allocate(h, semanticabi.LayoutInterfaceValue, func(fields []FieldValue) error {
		if err := SetObjects(semanticabi.LayoutInterfaceValue, fields, "interface", interfaceDefinition); err != nil {
			return err
		}
		return SetCells(semanticabi.LayoutInterfaceValue, fields, "underlying", arrayCell)
	})
	if err != nil {
		return "", err
	}
	interfaceCell, _ := h.Cell(semanticabi.TagKindInterfaceValue, interfaceValue)
	context, err := allocate(h, semanticabi.LayoutErrorContext, func(fields []FieldValue) error {
		return SetCells(semanticabi.LayoutErrorContext, fields, "values", interfaceCell)
	})
	if err != nil {
		return "", err
	}
	errorID, err := allocate(h, semanticabi.LayoutError, func(fields []FieldValue) error {
		return SetObjects(semanticabi.LayoutError, fields, "context", context)
	})
	if err != nil {
		return "", err
	}
	errorCell, _ := h.Cell(semanticabi.TagKindError, errorID)
	state, err := allocate(h, semanticabi.LayoutIteratorState, func(fields []FieldValue) error {
		return SetCells(semanticabi.LayoutIteratorState, fields, "values", errorCell)
	})
	if err != nil {
		return "", err
	}
	iterator, err := allocate(h, semanticabi.LayoutIterator, func(fields []FieldValue) error {
		return SetObjects(semanticabi.LayoutIterator, fields, "state", state)
	})
	if err != nil {
		return "", err
	}
	iteratorCell, _ := h.Cell(semanticabi.TagKindIterator, iterator)
	frame := h.OpenRootFrame(iteratorCell)
	collection, err := h.Collect()
	if err != nil || collection.Reachable != 7 {
		return "", fmt.Errorf("semantic chain reachability = %+v: %v", collection, err)
	}
	_ = h.CloseRootFrame(frame)
	return "iterator -> error -> interface -> array graph retains 7 objects", nil
}

func vectorRootFrameLifetime() (string, error) {
	h := New()
	first, _ := allocate(h, semanticabi.LayoutString, nil)
	second, _ := allocate(h, semanticabi.LayoutString, nil)
	firstCell, _ := h.Cell(semanticabi.TagKindString, first)
	secondCell, _ := h.Cell(semanticabi.TagKindString, second)
	frame := h.OpenRootFrame(firstCell)
	if err := h.ReplaceRootFrame(frame, secondCell); err != nil {
		return "", err
	}
	collection, err := h.Collect()
	if err != nil || collection.Reachable != 1 || collection.Collected != 1 {
		return "", fmt.Errorf("root replacement collection = %+v: %v", collection, err)
	}
	if _, err := h.Resolve(first); err == nil {
		return "", fmt.Errorf("replaced root remained live")
	}
	if _, err := h.Resolve(second); err != nil {
		return "", fmt.Errorf("replacement root was lost: %v", err)
	}
	return "root replacement drops old identity and retains new identity", nil
}

func allocate(h *Heap, layoutID uint16, configure func([]FieldValue) error) (semanticabi.ObjectID, error) {
	fields, err := NewFields(layoutID)
	if err != nil {
		return 0, err
	}
	if configure != nil {
		if err := configure(fields); err != nil {
			return 0, err
		}
	}
	return h.AllocateLayout(layoutID, fields)
}

func immediateInteger(value uint64) semanticabi.Cell {
	cell, err := semanticabi.ImmediateCell(semanticabi.TagKindInteger, 0, value)
	if err != nil {
		panic(err)
	}
	return cell
}
