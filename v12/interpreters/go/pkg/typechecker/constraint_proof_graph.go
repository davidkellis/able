package typechecker

import "fmt"

const maxConstraintProofDepth = 128

type constraintProofStatus uint8

const (
	constraintProofActive constraintProofStatus = iota + 1
	constraintProofSucceeded
	constraintProofFailed
)

type constraintProof struct {
	status   constraintProofStatus
	anchored bool
	detail   string
}

func (c *Checker) typeImplementsInterface(subject Type, iface InterfaceType, args []Type) (ok bool, detail string) {
	key := canonicalConstraintProofKey(subject, iface, args)
	if c.constraintProofs == nil {
		c.constraintProofs = make(map[string]*constraintProof)
	}
	if proof, exists := c.constraintProofs[key]; exists {
		switch proof.status {
		case constraintProofSucceeded:
			return true, ""
		case constraintProofFailed:
			return false, proof.detail
		case constraintProofActive:
			if proof.anchored {
				return true, ""
			}
			return false, fmt.Sprintf("recursive constraint proof for %s is not anchored", key)
		}
	}
	if len(c.constraintProofStack) >= maxConstraintProofDepth {
		return false, fmt.Sprintf(
			"recursive constraint proof exceeds %d distinct obligations and is not well-founded",
			maxConstraintProofDepth,
		)
	}

	proof := &constraintProof{status: constraintProofActive}
	c.constraintProofs[key] = proof
	c.constraintProofStack = append(c.constraintProofStack, key)
	defer func() {
		c.constraintProofStack = c.constraintProofStack[:len(c.constraintProofStack)-1]
		if ok {
			proof.status = constraintProofSucceeded
			proof.detail = ""
			return
		}
		proof.status = constraintProofFailed
		proof.detail = detail
	}()

	return c.typeImplementsInterfaceUncached(subject, iface, args)
}

func canonicalConstraintProofKey(subject Type, iface InterfaceType, args []Type) string {
	interfaceLabel := iface.InterfaceName
	if len(args) > 0 {
		interfaceLabel = formatInterfaceApplication(iface, args)
	}
	return fmt.Sprintf("%T:%s => %s", subject, formatType(subject), interfaceLabel)
}

func (c *Checker) anchorCurrentConstraintProof() {
	if len(c.constraintProofStack) == 0 || c.constraintProofs == nil {
		return
	}
	key := c.constraintProofStack[len(c.constraintProofStack)-1]
	if proof := c.constraintProofs[key]; proof != nil && proof.status == constraintProofActive {
		proof.anchored = true
	}
}
