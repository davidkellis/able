package typechecker

func structName(t Type) (string, bool) {
	switch s := t.(type) {
	case StructType:
		return s.StructName, true
	case StructInstanceType:
		return s.StructName, true
	case AppliedType:
		if base, ok := s.Base.(StructType); ok {
			return base.StructName, true
		}
	}
	return "", false
}

func unionName(t Type) (string, bool) {
	switch u := t.(type) {
	case UnionType:
		return u.UnionName, u.UnionName != ""
	case AppliedType:
		return unionName(u.Base)
	}
	return "", false
}

func arrayElementType(t Type) (Type, bool) {
	switch arr := t.(type) {
	case ArrayType:
		return arr.Element, true
	case StructType:
		if arr.StructName == "Array" {
			return argumentOrUnknown(arr.Positional, 0), true
		}
	case StructInstanceType:
		if arr.StructName == "Array" {
			args := arr.TypeArgs
			if len(args) == 0 {
				args = arr.Positional
			}
			return argumentOrUnknown(args, 0), true
		}
	case AppliedType:
		if name, ok := structName(arr.Base); ok && name == "Array" {
			return argumentOrUnknown(arr.Arguments, 0), true
		}
	}
	return nil, false
}

func iterableElementType(t Type) (Type, bool) {
	if t == nil {
		return UnknownType{}, true
	}
	if _, ok := t.(UnknownType); ok {
		return UnknownType{}, true
	}
	if elem, ok := arrayElementType(t); ok {
		if elem == nil || isUnknownType(elem) {
			return UnknownType{}, true
		}
		return elem, true
	}
	if rng, ok := t.(RangeType); ok {
		if rng.Element == nil || isUnknownType(rng.Element) {
			return UnknownType{}, true
		}
		return rng.Element, true
	}
	if iter, ok := t.(IteratorType); ok {
		if iter.Element == nil || isUnknownType(iter.Element) {
			return UnknownType{}, true
		}
		return iter.Element, true
	}
	if name, ok := structName(t); ok {
		switch name {
		case "Array", "Iterator", "List", "LinkedList", "LazySeq", "Vector", "HashSet", "Deque", "Queue":
			elem := typeArgumentOrUnknown(t, 0)
			if elem == nil || isUnknownType(elem) {
				return UnknownType{}, true
			}
			return elem, true
		case "BitSet":
			return IntegerType{Suffix: "i32"}, true
		case "Range":
			return UnknownType{}, true
		}
	}
	if applied, ok := t.(AppliedType); ok {
		switch base := applied.Base.(type) {
		case InterfaceType:
			if base.InterfaceName == "Iterable" {
				if len(applied.Arguments) == 0 {
					return UnknownType{}, true
				}
				elem := applied.Arguments[0]
				if elem == nil || isUnknownType(elem) {
					return UnknownType{}, true
				}
				return elem, true
			}
		}
	}
	return UnknownType{}, false
}

func typeArgumentOrUnknown(t Type, idx int) Type {
	switch v := t.(type) {
	case StructType:
		return argumentOrUnknown(v.Positional, idx)
	case StructInstanceType:
		args := v.TypeArgs
		if len(args) == 0 {
			args = v.Positional
		}
		return argumentOrUnknown(args, idx)
	case AppliedType:
		return argumentOrUnknown(v.Arguments, idx)
	default:
		return UnknownType{}
	}
}
