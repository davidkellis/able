package runtime

var (
	uint128Maximum = Uint128{High: ^uint64(0), Low: ^uint64(0)}
	int128Minimum  = Int128{High: uint64(1) << 63}
	int128Maximum  = Int128{High: ^(uint64(1) << 63), Low: ^uint64(0)}
)

func (value Uint128) WrappingAdd(other Uint128) Uint128 {
	result, _ := value.AddChecked(other)
	return result
}

func (value Uint128) WrappingSub(other Uint128) Uint128 {
	result, _ := value.SubChecked(other)
	return result
}

func (value Uint128) WrappingMul(other Uint128) Uint128 {
	result, _ := value.MulChecked(other)
	return result
}

func (value Uint128) SaturatingAdd(other Uint128) Uint128 {
	if result, ok := value.AddChecked(other); ok {
		return result
	}
	return uint128Maximum
}

func (value Uint128) SaturatingSub(other Uint128) Uint128 {
	if result, ok := value.SubChecked(other); ok {
		return result
	}
	return Uint128{}
}

func (value Uint128) SaturatingMul(other Uint128) Uint128 {
	if result, ok := value.MulChecked(other); ok {
		return result
	}
	return uint128Maximum
}

func (value Int128) WrappingAdd(other Int128) Int128 {
	result, _ := value.AddChecked(other)
	return result
}

func (value Int128) WrappingSub(other Int128) Int128 {
	result, _ := value.SubChecked(other)
	return result
}

func (value Int128) WrappingMul(other Int128) Int128 {
	product, _ := Uint128FromBits(value).MulChecked(Uint128FromBits(other))
	return Int128FromBits(product)
}

func (value Int128) SaturatingAdd(other Int128) Int128 {
	if result, ok := value.AddChecked(other); ok {
		return result
	}
	if value.IsNegative() {
		return int128Minimum
	}
	return int128Maximum
}

func (value Int128) SaturatingSub(other Int128) Int128 {
	if result, ok := value.SubChecked(other); ok {
		return result
	}
	if value.IsNegative() {
		return int128Minimum
	}
	return int128Maximum
}

func (value Int128) SaturatingMul(other Int128) Int128 {
	if result, ok := value.MulChecked(other); ok {
		return result
	}
	if value.IsNegative() != other.IsNegative() {
		return int128Minimum
	}
	return int128Maximum
}
