package typechecker

import "fmt"

func summarizeStructType(src StructType) ExportedStructSummary {
	summary := ExportedStructSummary{
		TypeParams: summarizeGenericParams(src.TypeParams),
		Fields:     summarizeTypeMap(src.Fields),
		Positional: summarizeTypeSlice(src.Positional),
		Where:      summarizeWhereConstraints(src.Where),
	}
	if summary.Fields == nil {
		summary.Fields = map[string]string{}
	}
	if summary.Positional == nil {
		summary.Positional = []string{}
	}
	if summary.TypeParams == nil {
		summary.TypeParams = []ExportedGenericParamSummary{}
	}
	if summary.Where == nil {
		summary.Where = []ExportedWhereConstraintSummary{}
	}
	return summary
}

func summarizeInterfaceType(src InterfaceType) ExportedInterfaceSummary {
	methods := make(map[string]ExportedFunctionSummary, len(src.Methods))
	for name, fn := range src.Methods {
		methods[name] = summarizeFunctionType(fn)
	}
	if methods == nil {
		methods = map[string]ExportedFunctionSummary{}
	}
	return ExportedInterfaceSummary{
		TypeParams: summarizeGenericParams(src.TypeParams),
		Methods:    methods,
		Where:      summarizeWhereConstraints(src.Where),
	}
}

func summarizeFunctionType(src FunctionType) ExportedFunctionSummary {
	return ExportedFunctionSummary{
		Parameters:  summarizeTypeSlice(src.Params),
		ReturnType:  formatType(src.Return),
		TypeParams:  summarizeGenericParams(src.TypeParams),
		Where:       summarizeWhereConstraints(src.Where),
		Obligations: summarizeObligations(src.Obligations),
	}
}

func summarizeImplementation(src ImplementationSpec) ExportedImplementationSummary {
	return ExportedImplementationSummary{
		ImplName:      src.ImplName,
		InterfaceName: src.InterfaceName,
		Target:        formatType(src.Target),
		InterfaceArgs: summarizeTypeSlice(src.InterfaceArgs),
		TypeParams:    summarizeGenericParams(src.TypeParams),
		Methods:       summarizeFunctionMap(src.Methods),
		Where:         summarizeWhereConstraints(src.Where),
		Obligations:   summarizeObligations(src.Obligations),
	}
}

func summarizeMethodSet(src MethodSetSpec) ExportedMethodSetSummary {
	qualifier := typeName(src.Target)
	methods := src.Methods
	if len(src.TypeQualified) > 0 && qualifier != "" {
		remapped := make(map[string]FunctionType, len(src.Methods))
		for name, fn := range src.Methods {
			key := name
			if src.TypeQualified != nil && src.TypeQualified[name] && qualifier != "" {
				key = fmt.Sprintf("%s.%s", qualifier, name)
			}
			remapped[key] = fn
		}
		methods = remapped
	}
	return ExportedMethodSetSummary{
		TypeParams:  summarizeGenericParams(src.TypeParams),
		Target:      formatType(src.Target),
		Methods:     summarizeFunctionMap(methods),
		Where:       summarizeWhereConstraints(src.Where),
		Obligations: summarizeObligations(src.Obligations),
	}
}

func summarizeGenericParams(params []GenericParamSpec) []ExportedGenericParamSummary {
	if len(params) == 0 {
		return nil
	}
	out := make([]ExportedGenericParamSummary, len(params))
	for i, param := range params {
		out[i] = ExportedGenericParamSummary{
			Name:        param.Name,
			Constraints: summarizeTypeSlice(param.Constraints),
		}
		if len(out[i].Constraints) == 0 {
			out[i].Constraints = nil
		}
	}
	return out
}

func summarizeWhereConstraints(constraints []WhereConstraintSpec) []ExportedWhereConstraintSummary {
	if len(constraints) == 0 {
		return nil
	}
	out := make([]ExportedWhereConstraintSummary, len(constraints))
	for i, constraint := range constraints {
		out[i] = ExportedWhereConstraintSummary{
			TypeParam:   constraint.TypeParam,
			Constraints: summarizeTypeSlice(constraint.Constraints),
		}
		if len(out[i].Constraints) == 0 {
			out[i].Constraints = nil
		}
	}
	return out
}

func summarizeObligations(obligations []ConstraintObligation) []ExportedObligationSummary {
	if len(obligations) == 0 {
		return nil
	}
	out := make([]ExportedObligationSummary, len(obligations))
	for i, ob := range obligations {
		out[i] = ExportedObligationSummary{
			Owner:      ob.Owner,
			TypeParam:  ob.TypeParam,
			Constraint: formatType(ob.Constraint),
			Subject:    formatType(ob.Subject),
			Context:    ob.Context,
		}
	}
	return out
}

func summarizeTypeMap(src map[string]Type) map[string]string {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]string, len(src))
	for name, typ := range src {
		out[name] = formatType(typ)
	}
	return out
}

func summarizeFunctionMap(src map[string]FunctionType) map[string]ExportedFunctionSummary {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]ExportedFunctionSummary, len(src))
	for name, fn := range src {
		out[name] = summarizeFunctionType(fn)
	}
	return out
}

func summarizeTypeSlice(src []Type) []string {
	if len(src) == 0 {
		return nil
	}
	out := make([]string, len(src))
	for i, typ := range src {
		out[i] = formatType(typ)
	}
	return out
}
