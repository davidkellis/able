package main

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
)

type nominalProof struct {
	HeapLiterals                   int                      `json:"reachable_heap_literals"`
	FieldWrites                    int                      `json:"reachable_field_writes"`
	UnresolvedFieldWrites          int                      `json:"unresolved_field_writes"`
	UnknownMutationCapableCalls    int                      `json:"unknown_mutation_capable_calls"`
	RuntimeConversions             int                      `json:"runtime_or_host_conversions"`
	OpaqueFieldBoundarySites       int                      `json:"opaque_field_boundary_sites"`
	NonPrimitiveFieldSites         int                      `json:"non_primitive_field_sites"`
	NativeUnionSites               int                      `json:"native_union_sites"`
	NativeInterfaceSites           int                      `json:"native_interface_sites"`
	SpecializedPointerStorageSites int                      `json:"specialized_pointer_storage_sites"`
	PointerIdentityComparisons     int                      `json:"pointer_identity_comparisons"`
	UnknownCallSites               []nominalUnknownCallSite `json:"unknown_mutation_capable_call_sites,omitempty"`
	Eligible                       bool                     `json:"eligible"`
	Blockers                       []string                 `json:"blockers"`
}

type nominalUnknownCallSite struct {
	Caller          string `json:"caller"`
	Callee          string `json:"callee"`
	File            string `json:"file"`
	Line            int    `json:"line"`
	Column          int    `json:"column"`
	ArgumentIndexes []int  `json:"argument_indexes"`
}

type nominalFacts struct {
	fields map[string]struct{}
}

type nominalCarrier struct {
	name    string
	pointer bool
}

func analyzeNominalProofs(
	fset *token.FileSet,
	files map[string]*ast.File,
	functions map[string]*ast.FuncDecl,
	nominals map[string]struct{},
) map[string]*nominalProof {
	facts := collectNominalFacts(files, nominals)
	proofs := make(map[string]*nominalProof, len(facts))
	for name := range facts {
		proofs[name] = &nominalProof{}
	}
	collectSpecializedPointerStorage(files, proofs)
	collectNonPrimitiveFields(files, proofs)

	reachable := directlyReachableFunctions(functions, "__able_compiled_fn_main")
	returns := collectFunctionNominalReturns(functions, nominals)
	for _, function := range reachable {
		analyzeNominalFunction(fset, function, functions, returns, facts, proofs)
	}
	for _, proof := range proofs {
		if proof.HeapLiterals == 0 {
			proof.Blockers = append(proof.Blockers, "no-reachable-construction")
		}
		if proof.FieldWrites > 0 {
			proof.Blockers = append(proof.Blockers, "reachable-field-mutation")
		}
		if proof.UnresolvedFieldWrites > 0 {
			proof.Blockers = append(proof.Blockers, "unresolved-field-mutation")
		}
		if proof.PointerIdentityComparisons > 0 {
			proof.Blockers = append(proof.Blockers, "pointer-identity-observation")
		}
		if proof.RuntimeConversions > 0 {
			proof.Blockers = append(proof.Blockers, "runtime-or-host-identity-exposure")
		}
		if proof.OpaqueFieldBoundarySites > 0 {
			proof.Blockers = append(proof.Blockers, "opaque-field-boundary")
		}
		if proof.NonPrimitiveFieldSites > 0 {
			proof.Blockers = append(proof.Blockers, "non-primitive-field-carrier")
		}
		if proof.NativeInterfaceSites > 0 {
			proof.Blockers = append(proof.Blockers, "native-interface-identity-exposure")
		}
		if proof.UnknownMutationCapableCalls > 0 {
			proof.Blockers = append(proof.Blockers, "unknown-mutation-capable-call")
		}
		if len(proof.Blockers) == 0 {
			proof.Eligible = true
		}
	}
	return proofs
}

func collectNominalFacts(
	files map[string]*ast.File,
	nominals map[string]struct{},
) map[string]nominalFacts {
	result := make(map[string]nominalFacts, len(nominals))
	for name := range nominals {
		result[name] = nominalFacts{fields: make(map[string]struct{})}
	}
	for _, file := range files {
		for _, declaration := range file.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}
			for _, spec := range general.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				fact, ok := result[typeSpec.Name.Name]
				if !ok {
					continue
				}
				structure, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, field := range structure.Fields.List {
					for _, name := range field.Names {
						fact.fields[name.Name] = struct{}{}
					}
				}
				result[typeSpec.Name.Name] = fact
			}
		}
	}
	return result
}

func collectSpecializedPointerStorage(
	files map[string]*ast.File,
	proofs map[string]*nominalProof,
) {
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			arrayType, ok := node.(*ast.ArrayType)
			if !ok {
				return true
			}
			star, ok := arrayType.Elt.(*ast.StarExpr)
			if !ok {
				return true
			}
			name := typeName(star.X)
			if proof := proofs[name]; proof != nil {
				proof.SpecializedPointerStorageSites++
			}
			return true
		})
	}
}

func collectNonPrimitiveFields(
	files map[string]*ast.File,
	proofs map[string]*nominalProof,
) {
	for _, file := range files {
		for _, declaration := range file.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}
			for _, spec := range general.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || proofs[typeSpec.Name.Name] == nil {
					continue
				}
				structure, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, field := range structure.Fields.List {
					if !isPrimitiveCarrierType(field.Type) {
						proofs[typeSpec.Name.Name].NonPrimitiveFieldSites++
					}
				}
			}
		}
	}
}

func isPrimitiveCarrierType(expression ast.Expr) bool {
	switch typed := expression.(type) {
	case *ast.Ident:
		switch typed.Name {
		case "bool", "byte", "rune", "string",
			"int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
			"float32", "float64":
			return true
		default:
			return false
		}
	case *ast.StructType:
		return typed.Fields == nil || len(typed.Fields.List) == 0
	case *ast.IndexExpr:
		return expressionName(typed.X) == "__able_nullable" &&
			isPrimitiveCarrierType(typed.Index)
	default:
		return false
	}
}

func collectFunctionNominalReturns(
	functions map[string]*ast.FuncDecl,
	nominals map[string]struct{},
) map[string]nominalCarrier {
	result := make(map[string]nominalCarrier)
	for name, function := range functions {
		if function.Type.Results == nil || len(function.Type.Results.List) == 0 {
			continue
		}
		if carrier, ok := nominalCarrierForType(
			function.Type.Results.List[0].Type,
			nominals,
		); ok {
			result[name] = carrier
		}
	}
	return result
}

func directlyReachableFunctions(
	functions map[string]*ast.FuncDecl,
	root string,
) []*ast.FuncDecl {
	pending := []string{root}
	seen := make(map[string]bool)
	var result []*ast.FuncDecl
	for len(pending) > 0 {
		name := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if seen[name] {
			continue
		}
		seen[name] = true
		function := functions[name]
		if function == nil || function.Body == nil {
			continue
		}
		result = append(result, function)
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			callee := expressionName(call.Fun)
			if functions[callee] != nil && !seen[callee] {
				pending = append(pending, callee)
			}
			return true
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name.Name < result[j].Name.Name
	})
	return result
}

func analyzeNominalFunction(
	fset *token.FileSet,
	function *ast.FuncDecl,
	functions map[string]*ast.FuncDecl,
	returns map[string]nominalCarrier,
	facts map[string]nominalFacts,
	proofs map[string]*nominalProof,
) {
	environment := make(map[string]nominalCarrier)
	seedNominalFields(function.Type.Params, environment, proofs)
	if function.Recv != nil {
		seedNominalFields(function.Recv, environment, proofs)
	}

	ast.Inspect(function.Body, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.DeclStmt:
			seedNominalDeclaration(typed.Decl, environment, proofs, returns)
		case *ast.AssignStmt:
			recordNominalAssignment(
				typed,
				function.Name.Name,
				environment,
				returns,
				facts,
				proofs,
			)
		case *ast.UnaryExpr:
			if typed.Op != token.AND {
				break
			}
			literal, ok := typed.X.(*ast.CompositeLit)
			if !ok {
				break
			}
			if proof := proofs[typeName(literal.Type)]; proof != nil {
				proof.HeapLiterals++
			}
		case *ast.BinaryExpr:
			if typed.Op != token.EQL && typed.Op != token.NEQ {
				break
			}
			left, leftOK := expressionNominalCarrier(
				typed.X,
				environment,
				returns,
				proofs,
			)
			right, rightOK := expressionNominalCarrier(
				typed.Y,
				environment,
				returns,
				proofs,
			)
			if leftOK && rightOK && left.pointer && right.pointer &&
				left.name == right.name {
				proofs[left.name].PointerIdentityComparisons++
			}
		case *ast.CallExpr:
			recordNominalCall(
				fset,
				typed,
				function.Name.Name,
				environment,
				functions,
				returns,
				proofs,
			)
		}
		return true
	})
}

func seedNominalFields(
	fields *ast.FieldList,
	environment map[string]nominalCarrier,
	proofs map[string]*nominalProof,
) {
	if fields == nil {
		return
	}
	nominals := proofNameSet(proofs)
	for _, field := range fields.List {
		carrier, ok := nominalCarrierForType(field.Type, nominals)
		if !ok {
			continue
		}
		for _, name := range field.Names {
			environment[name.Name] = carrier
		}
	}
}

func seedNominalDeclaration(
	declaration ast.Decl,
	environment map[string]nominalCarrier,
	proofs map[string]*nominalProof,
	returns map[string]nominalCarrier,
) {
	general, ok := declaration.(*ast.GenDecl)
	if !ok || general.Tok != token.VAR {
		return
	}
	nominals := proofNameSet(proofs)
	for _, rawSpec := range general.Specs {
		spec, ok := rawSpec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		carrier, typed := nominalCarrierForType(spec.Type, nominals)
		for index, name := range spec.Names {
			current, ok := carrier, typed
			if !ok && index < len(spec.Values) {
				current, ok = expressionNominalCarrier(
					spec.Values[index],
					environment,
					returns,
					proofs,
				)
			}
			if ok {
				environment[name.Name] = current
			}
		}
	}
}

func recordNominalAssignment(
	assignment *ast.AssignStmt,
	functionName string,
	environment map[string]nominalCarrier,
	returns map[string]nominalCarrier,
	facts map[string]nominalFacts,
	proofs map[string]*nominalProof,
) {
	for _, left := range assignment.Lhs {
		selector, ok := left.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if carrier, ok := expressionNominalCarrier(
			selector.X,
			environment,
			returns,
			proofs,
		); ok {
			if _, exists := facts[carrier.name].fields[selector.Sel.Name]; exists {
				if !strings.HasPrefix(
					functionName,
					"__able_struct_"+carrier.name+"_from",
				) {
					proofs[carrier.name].FieldWrites++
				}
			}
			continue
		}
		for name, fact := range facts {
			if _, exists := fact.fields[selector.Sel.Name]; exists {
				proofs[name].UnresolvedFieldWrites++
			}
		}
	}
	for index, left := range assignment.Lhs {
		identifier, ok := left.(*ast.Ident)
		if !ok || index >= len(assignment.Rhs) {
			continue
		}
		if carrier, ok := expressionNominalCarrier(
			assignment.Rhs[index],
			environment,
			returns,
			proofs,
		); ok {
			environment[identifier.Name] = carrier
		}
	}
}

func recordNominalCall(
	fset *token.FileSet,
	call *ast.CallExpr,
	caller string,
	environment map[string]nominalCarrier,
	functions map[string]*ast.FuncDecl,
	returns map[string]nominalCarrier,
	proofs map[string]*nominalProof,
) {
	callee := expressionName(call.Fun)
	arguments := make(map[string]struct{})
	argumentIndexes := make(map[string][]int)
	fieldReferences := make(map[string]struct{})
	for index, argument := range call.Args {
		if carrier, ok := expressionNominalCarrier(
			argument,
			environment,
			returns,
			proofs,
		); ok {
			arguments[carrier.name] = struct{}{}
			argumentIndexes[carrier.name] = append(
				argumentIndexes[carrier.name],
				index,
			)
		}
		for name := range nominalFieldReferences(argument, environment, proofs) {
			fieldReferences[name] = struct{}{}
		}
	}
	if isOpaqueBoundaryCall(callee, functions) {
		for name := range fieldReferences {
			proofs[name].OpaqueFieldBoundarySites++
		}
	}
	for name := range arguments {
		proof := proofs[name]
		switch {
		case strings.HasPrefix(callee, "__able_struct_"+name+"_") &&
			(strings.Contains(callee, "_from") ||
				strings.Contains(callee, "_to") ||
				strings.Contains(callee, "_apply")):
			// Counted once below, including conversions that return a nominal.
		case strings.HasPrefix(callee, "__able_union_") &&
			strings.Contains(callee, "_ptr_"+name):
			// Counted once below, including projections with no nominal argument.
		case strings.HasPrefix(callee, "__able_iface_") &&
			strings.Contains(callee, "_ptr_"+name):
			// Counted once below, including projections with no nominal argument.
		case functions[callee] != nil:
			// The whole generated call graph is inspected separately.
		case callee == "append", callee == "copy", callee == "len", callee == "cap":
			// Built-in storage operations do not mutate the nominal payload.
		default:
			proof.UnknownMutationCapableCalls++
			position := fset.Position(call.Pos())
			proof.UnknownCallSites = append(
				proof.UnknownCallSites,
				nominalUnknownCallSite{
					Caller:          caller,
					Callee:          callee,
					File:            filepath.Base(position.Filename),
					Line:            position.Line,
					Column:          position.Column,
					ArgumentIndexes: argumentIndexes[name],
				},
			)
		}
	}

	// Conversions can return a nominal without accepting one as an argument.
	for name, proof := range proofs {
		if strings.HasPrefix(callee, "__able_struct_"+name+"_") &&
			(strings.Contains(callee, "_from") ||
				strings.Contains(callee, "_to") ||
				strings.Contains(callee, "_apply")) {
			proof.RuntimeConversions++
		}
		if strings.HasPrefix(callee, "__able_union_") &&
			strings.Contains(callee, "_ptr_"+name) {
			proof.NativeUnionSites++
		}
		if strings.HasPrefix(callee, "__able_iface_") &&
			strings.Contains(callee, "_ptr_"+name) {
			proof.NativeInterfaceSites++
		}
	}
}

func nominalFieldReferences(
	expression ast.Expr,
	environment map[string]nominalCarrier,
	proofs map[string]*nominalProof,
) map[string]struct{} {
	result := make(map[string]struct{})
	ast.Inspect(expression, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		carrier, ok := expressionNominalCarrier(
			selector.X,
			environment,
			nil,
			proofs,
		)
		if ok {
			result[carrier.name] = struct{}{}
		}
		return true
	})
	return result
}

func isOpaqueBoundaryCall(
	callee string,
	functions map[string]*ast.FuncDecl,
) bool {
	if callee == "" {
		return true
	}
	if functions[callee] != nil {
		return false
	}
	switch callee {
	case "append", "copy", "len", "cap", "int", "int8", "int16", "int32",
		"int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64", "string", "rune", "byte", "bool":
		return false
	default:
		return true
	}
}

func expressionNominalCarrier(
	expression ast.Expr,
	environment map[string]nominalCarrier,
	returns map[string]nominalCarrier,
	proofs map[string]*nominalProof,
) (nominalCarrier, bool) {
	switch typed := expression.(type) {
	case *ast.Ident:
		carrier, ok := environment[typed.Name]
		return carrier, ok
	case *ast.ParenExpr:
		return expressionNominalCarrier(
			typed.X,
			environment,
			returns,
			proofs,
		)
	case *ast.UnaryExpr:
		if typed.Op != token.AND {
			return nominalCarrier{}, false
		}
		literal, ok := typed.X.(*ast.CompositeLit)
		if !ok || proofs[typeName(literal.Type)] == nil {
			return nominalCarrier{}, false
		}
		return nominalCarrier{name: typeName(literal.Type), pointer: true}, true
	case *ast.CompositeLit:
		name := typeName(typed.Type)
		if proofs[name] == nil {
			return nominalCarrier{}, false
		}
		return nominalCarrier{name: name}, true
	case *ast.CallExpr:
		carrier, ok := returns[expressionName(typed.Fun)]
		return carrier, ok
	default:
		return nominalCarrier{}, false
	}
}

func nominalCarrierForType(
	expression ast.Expr,
	nominals map[string]struct{},
) (nominalCarrier, bool) {
	if expression == nil {
		return nominalCarrier{}, false
	}
	pointer := false
	if star, ok := expression.(*ast.StarExpr); ok {
		pointer = true
		expression = star.X
	}
	name := typeName(expression)
	if _, ok := nominals[name]; !ok {
		return nominalCarrier{}, false
	}
	return nominalCarrier{name: name, pointer: pointer}, true
}

func proofNameSet(proofs map[string]*nominalProof) map[string]struct{} {
	result := make(map[string]struct{}, len(proofs))
	for name := range proofs {
		result[name] = struct{}{}
	}
	return result
}
