package typechecker

import (
	"fmt"
	"strings"
)

type interfaceResolution struct {
	iface InterfaceType
	args  []Type
	name  string
	err   string
}

func (c *Checker) resolveObligations() []Diagnostic {
	return c.evaluateObligations(c.obligations)
}

func (c *Checker) evaluateObligations(obligations []ConstraintObligation) []Diagnostic {
	if len(obligations) == 0 {
		return nil
	}
	methodObligation := map[string]bool{}
	for _, ob := range obligations {
		if strings.HasPrefix(ob.Owner, "methods for ") && strings.Contains(ob.Owner, "::") {
			label := strings.TrimSpace(strings.TrimPrefix(ob.Owner, "methods for "))
			if parts := strings.Split(label, "::"); len(parts) > 0 {
				methodObligation[parts[0]] = true
			}
		}
	}
	filtered := make([]ConstraintObligation, 0, len(obligations))
	for _, ob := range obligations {
		if strings.HasPrefix(ob.Owner, "methods for ") && !strings.Contains(ob.Owner, "::") && ob.Context == "via method set" {
			label := strings.TrimSpace(strings.TrimPrefix(ob.Owner, "methods for "))
			if methodObligation[label] {
				continue
			}
		}
		filtered = append(filtered, ob)
	}
	var diags []Diagnostic
	for _, ob := range filtered {
		diags = append(diags, c.evaluateObligation(ob)...)
	}
	return diags
}

func (c *Checker) evaluateObligation(ob ConstraintObligation) []Diagnostic {
	res := c.resolveConstraintInterfaceType(ob.Constraint)
	contextLabel := ""
	if ob.Context != "" {
		contextLabel = " (" + ob.Context + ")"
	}
	if ob.Context == "via method set" && strings.HasPrefix(ob.Owner, "methods for ") {
		contextLabel = ""
	}
	if res.err != "" {
		diags := []Diagnostic{{
			Message: fmt.Sprintf("typechecker: %s constraint on %s%s %s", ob.Owner, ob.TypeParam, contextLabel, res.err),
			Node:    ob.Node,
		}}
		if ob.Subject != nil && !isUnknownType(ob.Subject) && !isTypeParameter(ob.Subject) {
			subject := formatType(ob.Subject)
			interfaceLabel := res.name
			if interfaceLabel == "" && res.iface.InterfaceName != "" {
				interfaceLabel = res.iface.InterfaceName
			}
			if len(res.args) > 0 {
				interfaceLabel = formatInterfaceApplication(res.iface, res.args)
			}
			if interfaceLabel == "" {
				interfaceLabel = "<unknown>"
			}
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: %s constraint on %s%s is not satisfied: %s does not implement %s", ob.Owner, ob.TypeParam, contextLabel, subject, interfaceLabel),
				Node:    ob.Node,
			})
		}
		return diags
	}
	expectedParams := len(res.iface.TypeParams)
	explicitParams := explicitInterfaceParamCount(res.iface)
	providedArgs := len(res.args)
	if explicitParams > 0 && providedArgs == 0 {
		return []Diagnostic{{
			Message: fmt.Sprintf("typechecker: %s constraint on %s%s requires %d type argument(s) for interface '%s'", ob.Owner, ob.TypeParam, contextLabel, explicitParams, res.iface.InterfaceName),
			Node:    ob.Node,
		}}
	}
	if explicitParams == 0 && expectedParams == 1 && providedArgs == 0 {
		arg := ob.Subject
		if arg == nil || isUnknownType(arg) {
			if ob.TypeParam != "" {
				arg = TypeParameterType{ParameterName: ob.TypeParam}
			} else {
				arg = UnknownType{}
			}
		}
		res.args = []Type{arg}
		providedArgs = 1
	}
	if expectedParams != providedArgs && providedArgs != 0 {
		return []Diagnostic{{
			Message: fmt.Sprintf("typechecker: %s constraint on %s%s expected %d type argument(s) for interface '%s', got %d", ob.Owner, ob.TypeParam, contextLabel, expectedParams, res.iface.InterfaceName, providedArgs),
			Node:    ob.Node,
		}}
	}
	if ok, detail := c.obligationSatisfied(ob, res); !ok {
		subject := formatType(ob.Subject)
		interfaceLabel := formatInterfaceApplication(res.iface, res.args)
		reason := ""
		if detail != "" {
			reason = ": " + detail
		}
		return []Diagnostic{{
			Message: fmt.Sprintf("typechecker: %s constraint on %s%s is not satisfied: %s does not implement %s%s", ob.Owner, ob.TypeParam, contextLabel, subject, interfaceLabel, reason),
			Node:    ob.Node,
		}}
	}
	return nil
}

func explicitInterfaceParamCount(iface InterfaceType) int {
	if len(iface.TypeParams) == 0 {
		return 0
	}
	count := 0
	seenInferred := false
	for _, param := range iface.TypeParams {
		if param.IsInferred {
			seenInferred = true
		}
		if param.Name == "" || param.IsInferred {
			continue
		}
		count++
	}
	if seenInferred {
		return count
	}
	selfNames := collectSelfPatternNames(iface.SelfTypePattern)
	if len(selfNames) == 0 {
		return len(iface.TypeParams)
	}
	count = 0
	for _, param := range iface.TypeParams {
		if param.Name == "" {
			continue
		}
		if _, ok := selfNames[param.Name]; ok {
			continue
		}
		count++
	}
	return count
}

func (c *Checker) resolveConstraintInterfaceType(t Type) interfaceResolution {
	switch val := t.(type) {
	case InterfaceType:
		return interfaceResolution{iface: val, name: val.InterfaceName}
	case AppliedType:
		base := c.resolveConstraintInterfaceType(val.Base)
		if base.err != "" {
			return base
		}
		args := append([]Type{}, val.Arguments...)
		return interfaceResolution{iface: base.iface, args: args, name: base.iface.InterfaceName}
	case StructType:
		return c.interfaceFromName(val.StructName)
	case StructInstanceType:
		return c.interfaceFromName(val.StructName)
	default:
		return interfaceResolution{err: fmt.Sprintf("must reference an interface (got %s)", typeName(t)), name: typeName(t)}
	}
}

func (c *Checker) interfaceFromName(name string) interfaceResolution {
	if name == "" {
		return interfaceResolution{err: "must reference an interface (got <unknown>)", name: "<unknown>"}
	}
	if c.global == nil {
		return interfaceResolution{err: fmt.Sprintf("references unknown interface '%s'", name), name: name}
	}
	decl, ok := c.global.Lookup(name)
	if !ok {
		return interfaceResolution{err: fmt.Sprintf("references unknown interface '%s'", name), name: name}
	}
	iface, ok := decl.(InterfaceType)
	if !ok {
		return interfaceResolution{err: fmt.Sprintf("references '%s' which is not an interface", name), name: name}
	}
	return interfaceResolution{iface: iface, name: iface.InterfaceName}
}

func (c *Checker) obligationSatisfied(ob ConstraintObligation, res interfaceResolution) (bool, string) {
	subject := ob.Subject
	if subject == nil || isUnknownType(subject) || isTypeParameter(subject) {
		return true, ""
	}
	ok, detail := c.typeImplementsInterface(subject, res.iface, res.args)
	if !ok {
		return false, detail
	}
	return true, ""
}

func (c *Checker) typeImplementsInterface(subject Type, iface InterfaceType, args []Type) (bool, string) {
	if implementsIntrinsicInterface(subject, iface.InterfaceName) {
		return true, ""
	}
	var implDetail string
	switch val := subject.(type) {
	case NullableType:
		if ok, detail := c.implementationProvidesInterface(subject, iface, args); ok {
			return true, ""
		} else if detail != "" {
			implDetail = detail
		}
		ok, detail := c.typeImplementsInterface(val.Inner, iface, args)
		if !ok {
			if detail != "" {
				return false, detail
			}
			if implDetail != "" {
				return false, implDetail
			}
			return false, ""
		}
		return true, ""
	case UnionLiteralType:
		if ok, detail := c.implementationProvidesInterface(subject, iface, args); ok {
			return true, ""
		} else if detail != "" {
			implDetail = detail
		}
		for _, member := range val.Members {
			ok, detail := c.typeImplementsInterface(member, iface, args)
			if !ok {
				if detail != "" {
					return false, detail
				}
				if implDetail != "" {
					return false, implDetail
				}
				return false, detail
			}
		}
		return true, ""
	}
	if subjectMatchesInterface(subject, iface, args) {
		return true, ""
	}
	if ok, detail := c.implementationProvidesInterface(subject, iface, args); ok {
		return true, ""
	} else if detail != "" {
		implDetail = detail
	}
	if ok, detail := c.methodSetProvidesInterface(subject, iface, args); ok {
		return true, ""
	} else if detail != "" {
		return false, detail
	}
	if implDetail != "" {
		return false, implDetail
	}
	return false, ""
}

func implementsIntrinsicInterface(subject Type, interfaceName string) bool {
	switch interfaceName {
	case "Hash", "Eq":
		switch val := subject.(type) {
		case PrimitiveType:
			return val.Kind == PrimitiveString || val.Kind == PrimitiveBool || val.Kind == PrimitiveChar
		case IntegerType:
			return val.Suffix != ""
		case FloatType:
			return val.Suffix != ""
		}
	}
	return false
}

func subjectMatchesInterface(subject Type, iface InterfaceType, args []Type) bool {
	switch val := subject.(type) {
	case InterfaceType:
		return val.InterfaceName == iface.InterfaceName
	case AppliedType:
		baseIface, ok := val.Base.(InterfaceType)
		if !ok || baseIface.InterfaceName != iface.InterfaceName {
			return false
		}
		if len(args) == 0 {
			return true
		}
		return interfaceArgsCompatible(val.Arguments, args)
	default:
		return false
	}
}

func (c *Checker) obligationSetSatisfied(obligations []ConstraintObligation) (bool, string, ConstraintObligation) {
	if len(obligations) == 0 {
		return true, "", ConstraintObligation{}
	}
	appendContext := func(detail string, context string) string {
		if context == "" {
			return detail
		}
		if detail == "" {
			return context
		}
		return detail + " (" + context + ")"
	}
	for _, ob := range obligations {
		if ob.Constraint == nil {
			continue
		}
		res := c.resolveConstraintInterfaceType(ob.Constraint)
		if res.err != "" {
			return false, appendContext(res.err, ob.Context), ob
		}
		if ok, detail := c.obligationSatisfied(ob, res); !ok {
			if detail != "" {
				return false, appendContext(detail, ob.Context), ob
			}
			return false, appendContext(detail, ob.Context), ob
		}
	}
	return true, "", ConstraintObligation{}
}
