package runtime

import (
	"encoding/binary"
)

const (
	fnvOffset64 uint64 = 14695981039346656037
	fnvPrime64  uint64 = 1099511628211
)

// HashWithTag seeds a new FNV-1a stream tagged with the provided discriminator.
func HashWithTag(tag byte, data []byte) uint64 {
	hash := fnvOffset64
	hash = HashBytes(hash, []byte{tag})
	if len(data) > 0 {
		hash = HashBytes(hash, data)
	}
	return hash
}

// HashBytes feeds the FNV-1a state with additional data.
func HashBytes(hash uint64, data []byte) uint64 {
	for _, b := range data {
		hash ^= uint64(b)
		hash *= fnvPrime64
	}
	return hash
}

// NewHasherValue constructs a hasher seeded with the FNV-1a offset basis.
func NewHasherValue() *HasherValue {
	return &HasherValue{state: fnvOffset64}
}

// Reset restores the hasher to its initial state.
func (h *HasherValue) Reset() {
	if h == nil {
		return
	}
	h.state = fnvOffset64
}

// WriteBytes appends raw bytes to the hasher state.
func (h *HasherValue) WriteBytes(data []byte) {
	if h == nil {
		return
	}
	h.state = HashBytes(h.state, data)
}

// WriteString appends the UTF-8 bytes of the provided string.
func (h *HasherValue) WriteString(val string) {
	h.WriteBytes([]byte(val))
}

// WriteBool appends a single byte representing the boolean value.
func (h *HasherValue) WriteBool(val bool) {
	var b byte
	if val {
		b = 1
	}
	h.WriteBytes([]byte{b})
}

// WriteUint64 encodes the integer in big-endian order and appends it.
func (h *HasherValue) WriteUint64(val uint64) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], val)
	h.WriteBytes(buf[:])
}

// WriteInt64 encodes the signed integer using two's complement big-endian form.
func (h *HasherValue) WriteInt64(val int64) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(val))
	h.WriteBytes(buf[:])
}

// Finish returns the current FNV-1a digest.
func (h *HasherValue) Finish() uint64 {
	if h == nil {
		return 0
	}
	return h.state
}
