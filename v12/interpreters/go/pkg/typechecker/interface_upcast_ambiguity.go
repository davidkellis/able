package typechecker

import "strings"

func isAmbiguousImplementationDetail(detail string) bool {
	return strings.HasPrefix(strings.TrimSpace(detail), "ambiguous implementations of ")
}

func (c *Checker) staticInterfaceUpcastAmbiguity(actual, expected Type) string {
	if c == nil || actual == nil || expected == nil || isUnknownType(actual) || isUnknownType(expected) {
		return ""
	}
	iface, args, ok := interfaceFromType(expected)
	if !ok {
		return ""
	}
	return c.interfaceUpcastAmbiguityForSubject(actual, iface, args)
}

func (c *Checker) interfaceUpcastAmbiguityForSubject(subject Type, iface InterfaceType, args []Type) string {
	if subject == nil || isUnknownType(subject) || isTypeParameter(subject) {
		return ""
	}
	if _, detail := c.implementationProvidesInterface(subject, iface, args); isAmbiguousImplementationDetail(detail) {
		return detail
	}
	switch value := subject.(type) {
	case AliasType:
		return c.interfaceUpcastAmbiguityForSubject(value.Target, iface, args)
	case NullableType:
		return c.interfaceUpcastAmbiguityForSubject(value.Inner, iface, args)
	case UnionLiteralType:
		for _, member := range value.Members {
			if detail := c.interfaceUpcastAmbiguityForSubject(member, iface, args); detail != "" {
				return detail
			}
		}
	case UnionType:
		for _, member := range value.Variants {
			if detail := c.interfaceUpcastAmbiguityForSubject(member, iface, args); detail != "" {
				return detail
			}
		}
	}
	return ""
}
