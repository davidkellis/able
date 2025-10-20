package typechecker

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

func (c *Checker) checkStructLiteral(env *Environment, expr *ast.StructLiteral) ([]Diagnostic, Type) {
	var diags []Diagnostic
	if expr == nil {
		return nil, UnknownType{}
	}

	var (
		structName  string
		structInfo  StructType
		hasInfo     bool
		typeArgs    []Type
		structSubst map[string]Type
	)
	if expr.StructType != nil && expr.StructType.Name != "" {
		structName = expr.StructType.Name
		if typ, ok := c.global.Lookup(structName); ok {
			if st, ok := typ.(StructType); ok {
				structInfo = st
				hasInfo = true
			}
		}
		if !hasInfo {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: unknown struct '%s'", structName),
				Node:    expr,
			})
		}
	}

	if len(expr.TypeArguments) > 0 {
		typeArgs = make([]Type, len(expr.TypeArguments))
		for i, arg := range expr.TypeArguments {
			typeArgs[i] = c.resolveTypeReference(arg)
		}
	}

	if hasInfo {
		expected := len(structInfo.TypeParams)
		if expected > 0 {
			if provided := len(typeArgs); provided > 0 && provided != expected {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: struct '%s' expects %d type argument(s), got %d", structInfo.StructName, expected, provided),
					Node:    expr,
				})
			}
		}
		if len(structInfo.TypeParams) > 0 && len(typeArgs) > 0 {
			structSubst = make(map[string]Type, len(structInfo.TypeParams))
			for i, param := range structInfo.TypeParams {
				if param.Name == "" {
					continue
				}
				if i < len(typeArgs) && typeArgs[i] != nil {
					structSubst[param.Name] = typeArgs[i]
					continue
				}
				structSubst[param.Name] = UnknownType{}
			}
		}
	}

	fields := make(map[string]Type)
	seen := make(map[string]struct{})
	var positional []Type
	if hasInfo && len(structInfo.Positional) > 0 {
		positional = make([]Type, len(structInfo.Positional))
		copy(positional, structInfo.Positional)
	} else if expr.IsPositional {
		positional = make([]Type, len(expr.Fields))
	}

	if expr.FunctionalUpdateSource != nil {
		sourceDiags, sourceType := c.checkExpression(env, expr.FunctionalUpdateSource)
		diags = append(diags, sourceDiags...)
		switch st := sourceType.(type) {
		case StructInstanceType:
			for name, typ := range st.Fields {
				fields[name] = typ
			}
			if len(st.Positional) > 0 {
				positional = append([]Type(nil), st.Positional...)
			}
		case StructType:
			if st.Fields != nil {
				for name, typ := range st.Fields {
					fields[name] = typ
				}
			}
			if len(st.Positional) > 0 {
				positional = append([]Type(nil), st.Positional...)
			}
		}
	}

	ensurePos := func(idx int) {
		if positional == nil {
			return
		}
		if idx >= len(positional) {
			extended := make([]Type, idx+1)
			copy(extended, positional)
			positional = extended
		}
	}

	for idx, field := range expr.Fields {
		if field == nil {
			continue
		}
		var name string
		if field.Name != nil {
			name = field.Name.Name
		}
		if name == "" && !expr.IsPositional {
			diags = append(diags, Diagnostic{
				Message: "typechecker: struct field requires a name",
				Node:    field,
			})
			continue
		}
		if name != "" {
			if _, exists := seen[name]; exists {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: duplicate struct field '%s'", name),
					Node:    field,
				})
				continue
			}
			seen[name] = struct{}{}
		}

		valueDiags, valueType := c.checkExpression(env, field.Value)
		diags = append(diags, valueDiags...)

		expected := Type(UnknownType{})
		if hasInfo {
			if name != "" {
				if t, ok := structInfo.Fields[name]; ok {
					expected = t
					if len(structSubst) > 0 {
						expected = substituteType(expected, structSubst)
					}
				} else {
					diags = append(diags, Diagnostic{
						Message: fmt.Sprintf("typechecker: struct '%s' has no field '%s'", structInfo.StructName, name),
						Node:    field,
					})
				}
			} else if idx < len(structInfo.Positional) {
				expected = structInfo.Positional[idx]
				if len(structSubst) > 0 {
					expected = substituteType(expected, structSubst)
				}
			} else {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: positional field %d out of range for struct '%s'", idx, structInfo.StructName),
					Node:    field,
				})
			}
		}

		chosen := expected
		if (chosen == nil || isUnknownType(chosen) || isTypeParameter(chosen)) && valueType != nil && !isUnknownType(valueType) {
			chosen = valueType
		}

		if name != "" {
			fields[name] = chosen
		} else {
			ensurePos(idx)
			if positional != nil {
				positional[idx] = chosen
			}
		}
	}

	if expr.FunctionalUpdateSource != nil {
		// already checked above to seed field information
	}

	instance := StructInstanceType{
		StructName: structName,
		Fields:     fields,
		Positional: positional,
	}
	c.infer.set(expr, instance)
	return diags, instance
}
