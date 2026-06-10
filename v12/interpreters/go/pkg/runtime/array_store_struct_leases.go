package runtime

import "fmt"

// arrayStructInstanceLeaseSidecar is reserved for canonical Array struct
// instances. Native is otherwise implementation-private, so callers attach
// this sidecar only after establishing that the instance represents Array.
type arrayStructInstanceLeaseSidecar struct {
	lease ArrayStoreLease
}

// ArrayStoreTrackStructInstanceLease records a canonical Array struct instance
// as the owner of handle. The sidecar keeps the lease with that instance
// without adding storage to every StructInstanceValue.
func ArrayStoreTrackStructInstanceLease(inst *StructInstanceValue, handle int64) error {
	if inst == nil {
		return fmt.Errorf("array struct lease owner is nil")
	}
	if !arrayStructInstanceLeaseOwner(inst) {
		return fmt.Errorf("array struct lease owner must be an Array instance")
	}
	if handle == 0 {
		return ArrayStoreReleaseStructInstanceLease(inst)
	}
	sidecar, err := arrayStructInstanceLeaseSidecarForTrack(inst)
	if err != nil {
		return err
	}
	return ArrayStoreTrackLeaseOwner(inst, &sidecar.lease, handle)
}

// ArrayStoreReleaseStructInstanceLease removes the canonical Array struct
// instance from diagnostic ownership accounting. It intentionally leaves
// backing state registered.
func ArrayStoreReleaseStructInstanceLease(inst *StructInstanceValue) error {
	if inst == nil {
		return fmt.Errorf("array struct lease owner is nil")
	}
	if !arrayStructInstanceLeaseOwner(inst) {
		return fmt.Errorf("array struct lease owner must be an Array instance")
	}
	sidecar, ok := inst.Native.(*arrayStructInstanceLeaseSidecar)
	if !ok || sidecar == nil {
		return nil
	}
	return ArrayStoreUpdateLease(&sidecar.lease, 0)
}

func arrayStructInstanceLeaseSidecarForTrack(inst *StructInstanceValue) (*arrayStructInstanceLeaseSidecar, error) {
	if inst == nil {
		return nil, fmt.Errorf("array struct lease owner is nil")
	}
	if inst.Native == nil {
		sidecar := &arrayStructInstanceLeaseSidecar{}
		inst.Native = sidecar
		return sidecar, nil
	}
	sidecar, ok := inst.Native.(*arrayStructInstanceLeaseSidecar)
	if !ok || sidecar == nil {
		return nil, fmt.Errorf("array struct instance already has native metadata")
	}
	return sidecar, nil
}

func arrayStructInstanceLeaseOwner(inst *StructInstanceValue) bool {
	return inst != nil &&
		inst.Definition != nil &&
		inst.Definition.Node != nil &&
		inst.Definition.Node.ID != nil &&
		inst.Definition.Node.ID.Name == "Array"
}
