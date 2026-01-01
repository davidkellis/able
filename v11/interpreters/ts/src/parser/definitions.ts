import * as AST from "../ast";
import type {
  BlockExpression,
  ExternFunctionBody,
  FunctionDefinition,
  FunctionParameter,
  FunctionSignature,
  GenericParameter,
  HostTarget,
  Identifier,
  ImplementationDefinition,
  InterfaceDefinition,
  MethodsDefinition,
  Pattern,
  PreludeStatement,
  StructDefinition,
  StructFieldDefinition,
  TypeAliasDefinition,
  TypeExpression,
  UnionDefinition,
  WhereClauseConstraint,
} from "../ast";
import {
  annotate,
  annotateStatement,
  getActiveParseContext,
  identifiersToStrings,
  hasLeadingPrivate,
  isIgnorableNode,
  MapperError,
  MutableParseContext,
  Node,
  ParseContext,
  parseIdentifier,
  sameNode,
  sliceText,
} from "./shared";

export function registerDefinitionParsers(ctx: MutableParseContext): void {
  ctx.parseFunctionDefinition = node => parseFunctionDefinition(node, ctx.source);
  ctx.parseStructDefinition = node => parseStructDefinition(node, ctx.source);
  ctx.parseMethodsDefinition = node => parseMethodsDefinition(node, ctx.source);
  ctx.parseImplementationDefinition = node => parseImplementationDefinition(node, ctx.source);
  ctx.parseNamedImplementationDefinition = node => parseNamedImplementationDefinition(node, ctx.source);
  ctx.parseUnionDefinition = node => parseUnionDefinition(node, ctx.source);
  ctx.parseInterfaceDefinition = node => parseInterfaceDefinition(node, ctx.source);
  ctx.parseTypeAliasDefinition = node => parseTypeAliasDefinition(node, ctx.source);
  ctx.parsePreludeStatement = node => parsePreludeStatement(node, ctx.source);
  ctx.parseExternFunction = node => parseExternFunction(node, ctx.source);
}

function parseFunctionDefinition(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): FunctionDefinition {
  if (node.type !== "function_definition") {
    throw new MapperError("parser: expected function_definition node");
  }
  const core = parseFunctionCore(node, source, ctx);
  const fn = AST.functionDefinition(
    core.name,
    core.params,
    core.body,
    core.returnType,
    core.generics,
    core.whereClause,
    core.isMethodShorthand,
    core.isPrivate,
  );
  return annotate(fn, node);
}

function parseFunctionCore(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): {
  name: Identifier;
  generics?: GenericParameter[];
  params: FunctionParameter[];
  returnType?: TypeExpression;
  whereClause?: WhereClauseConstraint[];
  body: BlockExpression;
  isMethodShorthand: boolean;
  isPrivate: boolean;
} {
  const name = parseIdentifier(node.childForFieldName("name"), source);
  const params = parseParameterList(node.childForFieldName("parameters"), source, ctx);
  const bodyNode = node.childForFieldName("body");
  const body = ctx.parseBlock(bodyNode);

  let isPrivate = false;
  for (let i = 0; i < node.childCount; i++) {
    const child = node.child(i);
    if (!child || isIgnorableNode(child)) continue;
    if (child.type === "private") {
      isPrivate = true;
      break;
    }
  }

  const returnType = ctx.parseReturnType(node.childForFieldName("return_type"));
  const generics = ctx.parseTypeParameters(node.childForFieldName("type_parameters"));
  const whereClause = ctx.parseWhereClause(node.childForFieldName("where_clause"));
  const methodShorthand = Boolean(node.childForFieldName("method_shorthand"));

  return {
    name,
    generics,
    params,
    returnType,
    whereClause,
    body,
    isMethodShorthand: methodShorthand,
    isPrivate,
  };
}

export function parseParameterList(
  node: Node | null | undefined,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): FunctionParameter[] {
  if (!node || node.namedChildCount === 0) {
    return [];
  }
  const params: FunctionParameter[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const paramNode = node.namedChild(i);
    if (!paramNode || isIgnorableNode(paramNode)) continue;
    params.push(parseParameter(paramNode, source, ctx));
  }
  return params;
}

export function parseParameter(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): FunctionParameter {
  if (node.type !== "parameter") {
    throw new MapperError("parser: expected parameter node");
  }
  const patternNode = node.childForFieldName("pattern");
  const pattern = ctx.parsePattern(patternNode);
  let namePattern: Pattern = pattern;
  let paramType: TypeExpression | undefined;
  if (pattern.type === "TypedPattern") {
    paramType = pattern.typeAnnotation;
    namePattern = pattern.pattern;
  }
  const typeNode = node.childForFieldName("type");
  if (!paramType && typeNode) {
    paramType = ctx.parseTypeExpression(typeNode) ?? undefined;
  }
  return annotate(AST.functionParameter(namePattern, paramType), node) as FunctionParameter;
}

function parseStructDefinition(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): StructDefinition {
  if (node.type !== "struct_definition") {
    throw new MapperError("parser: expected struct_definition node");
  }

  const nameNode = node.childForFieldName("name");
  const id = parseIdentifier(nameNode, source);

  const generics = ctx.parseTypeParameters(node.childForFieldName("type_parameters"));
  const whereClause = ctx.parseWhereClause(node.childForFieldName("where_clause"));
  const isPrivate = hasLeadingPrivate(node) ? true : undefined;

  let kind: StructDefinition["kind"] = "singleton";
  const fields: StructFieldDefinition[] = [];

  const recordNode = node.childForFieldName("record");
  const tupleNode = node.childForFieldName("tuple");

  if (recordNode) {
    kind = "named";
    for (let i = 0; i < recordNode.namedChildCount; i++) {
      const fieldNode = recordNode.namedChild(i);
      if (!fieldNode || isIgnorableNode(fieldNode) || fieldNode.type !== "struct_field") continue;
      fields.push(parseStructFieldDefinition(fieldNode, source, ctx));
    }
  } else if (tupleNode) {
    kind = "positional";
    for (let i = 0; i < tupleNode.namedChildCount; i++) {
      const child = tupleNode.namedChild(i);
      if (!child || !child.isNamed || isIgnorableNode(child)) continue;
      const fieldType = ctx.parseTypeExpression(child);
      if (!fieldType) {
        throw new MapperError("parser: unsupported tuple field type");
      }
      fields.push(AST.structFieldDefinition(fieldType));
    }
  }

  if (id?.name) {
    ctx.structKinds.set(id.name, kind);
  }
  const definition = AST.structDefinition(id, fields, kind, generics, whereClause, isPrivate);
  return annotateStatement(definition, node) as StructDefinition;
}

function parseTypeAliasDefinition(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): TypeAliasDefinition {
  if (node.type !== "type_alias_definition") {
    throw new MapperError("parser: expected type_alias_definition node");
  }
  const nameNode = node.childForFieldName("name");
  const id = parseIdentifier(nameNode, source);
  if (!id) {
    throw new MapperError("parser: type alias missing identifier");
  }
  const targetNode = node.childForFieldName("target");
  const targetType = ctx.parseTypeExpression(targetNode);
  if (!targetType) {
    throw new MapperError("parser: type alias missing target type");
  }
  const generics = ctx.parseTypeParameters(node.childForFieldName("type_parameters"));
  const whereClause = ctx.parseWhereClause(node.childForFieldName("where_clause"));
  const isPrivate = hasLeadingPrivate(node) ? true : undefined;
  const alias = AST.typeAliasDefinition(id, targetType, generics, whereClause, isPrivate);
  return annotate(alias, node) as TypeAliasDefinition;
}

function parseStructFieldDefinition(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): StructFieldDefinition {
  if (node.type !== "struct_field") {
    throw new MapperError("parser: expected struct_field node");
  }

  let name: Identifier | undefined;
  let fieldType: TypeExpression | null = null;

  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    if (child.type === "identifier" && !name) {
      name = parseIdentifier(child, source);
    } else if (!fieldType) {
      fieldType = ctx.parseTypeExpression(child);
    }
  }

  if (!fieldType) {
    throw new MapperError("parser: struct field missing type");
  }

  return annotate(AST.structFieldDefinition(fieldType, name), node) as StructFieldDefinition;
}

function parseMethodsDefinition(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): MethodsDefinition {
  if (node.type !== "methods_definition") {
    throw new MapperError("parser: expected methods_definition node");
  }

  const generics = ctx.parseTypeParameters(node.childForFieldName("type_parameters"));
  const whereClause = ctx.parseWhereClause(node.childForFieldName("where_clause"));
  const targetType = ctx.parseTypeExpression(node.childForFieldName("target"));
  if (!targetType) {
    throw new MapperError("parser: methods definition missing target type");
  }

  const definitions: FunctionDefinition[] = [];

  const targetNode = node.childForFieldName("target");
  const typeParamsNode = node.childForFieldName("type_parameters");
  const whereNode = node.childForFieldName("where_clause");

  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child) continue;
    const fieldName = node.fieldNameForChild(i);
    if (
      (fieldName === "target" || fieldName === "type_parameters" || fieldName === "where_clause") &&
      child.type !== "function_definition" &&
      child.type !== "method_member"
    ) {
      continue;
    }
    if (sameNode(child, targetNode) || sameNode(child, typeParamsNode) || sameNode(child, whereNode)) {
      continue;
    }
    if (child.type === "function_definition") {
      const fn = parseFunctionDefinition(child, source, ctx);
      definitions.push(fn);
    } else if (child.type === "method_member") {
      for (let j = 0; j < child.namedChildCount; j++) {
        const member = child.namedChild(j);
        if (!member || member.type !== "function_definition") continue;
        const fn = parseFunctionDefinition(member, source, ctx);
        definitions.push(fn);
      }
    }
  }

  return annotateStatement(
    AST.methodsDefinition(targetType, definitions, generics, whereClause),
    node,
  ) as MethodsDefinition;
}

function parseImplementationDefinitionNode(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): ImplementationDefinition {
  if (node.type !== "implementation_definition") {
    throw new MapperError("parser: expected implementation_definition node");
  }

  const interfaceNode = node.childForFieldName("interface");
  if (!interfaceNode) {
    throw new MapperError("parser: implementation missing interface");
  }

  const parts = ctx.parseQualifiedIdentifier(interfaceNode);
  let interfaceName = parts[parts.length - 1];
  if (parts.length > 1) {
    interfaceName = annotate(AST.identifier(identifiersToStrings(parts).join(".")), interfaceNode) as Identifier;
  }

  const interfaceArgs = parseInterfaceArguments(node.childForFieldName("interface_args"), source, ctx);
  const targetType = ctx.parseTypeExpression(node.childForFieldName("target"));
  if (!targetType) {
    throw new MapperError("parser: implementation missing target type");
  }
  const generics = ctx.parseTypeParameters(node.childForFieldName("type_parameters"));
  const whereClause = ctx.parseWhereClause(node.childForFieldName("where_clause"));
  const isPrivate = hasLeadingPrivate(node) ? true : undefined;

  const definitions: FunctionDefinition[] = [];
  const interfaceArgsNode = node.childForFieldName("interface_args");
  const targetNode = node.childForFieldName("target");
  const typeParamsNode = node.childForFieldName("type_parameters");
  const whereNode = node.childForFieldName("where_clause");

  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child) continue;
    const fieldName = node.fieldNameForChild(i);
    if (
      (fieldName === "interface" ||
        fieldName === "interface_args" ||
        fieldName === "target" ||
        fieldName === "type_parameters" ||
        fieldName === "where_clause") &&
      child.type !== "function_definition" &&
      child.type !== "method_member"
    ) {
      continue;
    }
    if (
      sameNode(child, interfaceNode) ||
      sameNode(child, interfaceArgsNode) ||
      sameNode(child, targetNode) ||
      sameNode(child, typeParamsNode) ||
      sameNode(child, whereNode)
    ) {
      continue;
    }
    if (child.type === "function_definition") {
      const fn = parseFunctionDefinition(child, source, ctx);
      definitions.push(fn);
    } else if (child.type === "method_member") {
      for (let j = 0; j < child.namedChildCount; j++) {
        const member = child.namedChild(j);
        if (!member || member.type !== "function_definition") continue;
        const fn = parseFunctionDefinition(member, source, ctx);
        definitions.push(fn);
      }
    }
  }

  return annotateStatement(
    AST.implementationDefinition(
      interfaceName,
      targetType,
      definitions,
      undefined,
      generics,
      interfaceArgs ?? undefined,
      whereClause,
      isPrivate,
    ),
    node,
  ) as ImplementationDefinition;
}

function parseImplementationDefinition(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): ImplementationDefinition {
  return parseImplementationDefinitionNode(node, source, ctx);
}

function parseNamedImplementationDefinition(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): ImplementationDefinition {
  if (node.type !== "named_implementation_definition") {
    throw new MapperError("parser: expected named implementation node");
  }
  const nameNode = node.childForFieldName("name");
  const implNode = node.childForFieldName("implementation");
  if (!implNode) {
    throw new MapperError("parser: named implementation missing implementation body");
  }
  const impl = parseImplementationDefinitionNode(implNode, source, ctx);
  if (nameNode) {
    impl.implName = parseIdentifier(nameNode, source);
  }
  return impl;
}

function parseInterfaceArguments(
  node: Node | null | undefined,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): TypeExpression[] | undefined {
  if (!node) return undefined;
  const args: TypeExpression[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child) continue;
    const genericNode = findTopLevelGenericApplication(child);
    if (genericNode) {
      const snippet = sliceText(genericNode, source).trim();
      const detail = snippet ? `; wrap "${snippet}" in parentheses` : "";
      throw new MapperError(`parser: interface arguments require parenthesized generic applications${detail}`);
    }
    const expr = ctx.parseTypeExpression(child);
    if (!expr) {
      throw new MapperError(`parser: unsupported interface argument kind ${child.type}`);
    }
    args.push(expr);
  }
  return args.length ? args : undefined;
}

function findTopLevelGenericApplication(node: Node): Node | null {
  let current: Node | null = node;
  while (current) {
    if (current.type === "type_generic_application") {
      return current;
    }
    if (current.type === "parenthesized_type") {
      return null;
    }
    if (current.namedChildCount !== 1) {
      return null;
    }
    const child = current.namedChild(0);
    if (!child || !child.isNamed) {
      return null;
    }
    current = child;
  }
  return null;
}

function parseUnionDefinition(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): UnionDefinition {
  if (node.type !== "union_definition") {
    throw new MapperError("parser: expected union_definition node");
  }

  const nameNode = node.childForFieldName("name");
  const name = parseIdentifier(nameNode, source);
  const typeParamsNode = node.childForFieldName("type_parameters");
  const typeParams = ctx.parseTypeParameters(typeParamsNode);

  const variants: TypeExpression[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child) continue;
    if (sameNode(child, nameNode) || sameNode(child, typeParamsNode)) continue;
    const variant = ctx.parseTypeExpression(child);
    if (!variant) {
      throw new MapperError("parser: invalid union variant");
    }
    variants.push(variant);
  }

  if (variants.length === 0) {
    throw new MapperError("parser: union definition requires variants");
  }

  return annotateStatement(
    AST.unionDefinition(name, variants, typeParams, undefined, hasLeadingPrivate(node)),
    node,
  ) as UnionDefinition;
}

function parseInterfaceDefinition(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): InterfaceDefinition {
  if (node.type !== "interface_definition") {
    throw new MapperError("parser: expected interface_definition node");
  }

  const nameNode = node.childForFieldName("name");
  const name = parseIdentifier(nameNode, source);
  const typeParamsNode = node.childForFieldName("type_parameters");
  const typeParams = ctx.parseTypeParameters(typeParamsNode);

  let selfType: TypeExpression | undefined;
  const selfNode = node.childForFieldName("self_type");
  if (selfNode) {
    selfType = ctx.parseTypeExpression(selfNode) ?? undefined;
  }

  const whereNode = node.childForFieldName("where_clause");
  const whereClause = ctx.parseWhereClause(whereNode);

  const compositeNode = node.childForFieldName("composite");
  const baseInterfaces = parseInterfaceBases(compositeNode, ctx);
  const isPrivate = hasLeadingPrivate(node) ? true : undefined;

  const signatures: FunctionSignature[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child) continue;
    if (
      sameNode(child, nameNode) ||
      sameNode(child, typeParamsNode) ||
      sameNode(child, selfNode) ||
      sameNode(child, whereNode) ||
      sameNode(child, compositeNode)
    ) {
      continue;
    }
    if (child.type !== "interface_member") continue;
    const sigNode = child.childForFieldName("signature");
    if (!sigNode) {
      throw new MapperError("parser: interface member missing signature");
    }
    const signature = parseFunctionSignature(sigNode, source, ctx);
    const defaultBody = child.childForFieldName("default_body");
    if (defaultBody) {
      signature.defaultImpl = ctx.parseBlock(defaultBody);
    }
    signatures.push(signature);
  }

  return annotateStatement(
    AST.interfaceDefinition(name, signatures, typeParams, selfType, whereClause, baseInterfaces, isPrivate),
    node,
  ) as InterfaceDefinition;
}

function parseFunctionSignature(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): FunctionSignature {
  const hasBody = Boolean(node.childForFieldName("body"));
  const signature = parseFunctionCore(node, source, ctx);
  return annotate(
    AST.functionSignature(
      signature.name,
      signature.params,
      signature.returnType,
      signature.generics,
      signature.whereClause,
      hasBody ? signature.body : undefined,
    ),
    node,
  ) as FunctionSignature;
}

function parseInterfaceBases(
  node: Node | null | undefined,
  ctx: ParseContext = getActiveParseContext(),
): TypeExpression[] | undefined {
  if (!node) return undefined;
  const bases: TypeExpression[] = [];
  const stack: Node[] = [];
  for (let i = node.namedChildCount - 1; i >= 0; i--) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    stack.push(child);
  }
  while (stack.length > 0) {
    const current = stack.pop();
    if (!current || isIgnorableNode(current)) continue;
    const typeExpr = current.type === "type_bound_list" ? null : ctx.parseTypeExpression(current);
    if (typeExpr) {
      bases.push(typeExpr);
      continue;
    }
    for (let i = current.namedChildCount - 1; i >= 0; i--) {
      const child = current.namedChild(i);
      if (child) {
        stack.push(child);
      }
    }
  }
  return bases.length ? bases : undefined;
}

function parsePreludeStatement(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): PreludeStatement {
  if (node.type !== "prelude_statement") {
    throw new MapperError("parser: expected prelude_statement node");
  }
  const target = parseHostTarget(node.childForFieldName("target"), source);
  const code = parseHostCodeBlock(node.childForFieldName("body"), source);
  return annotateStatement(AST.preludeStatement(target, code), node) as PreludeStatement;
}

function parseExternFunction(
  node: Node,
  source: string,
  ctx: ParseContext = getActiveParseContext(),
): ExternFunctionBody {
  if (node.type !== "extern_function") {
    throw new MapperError("parser: expected extern_function node");
  }
  const target = parseHostTarget(node.childForFieldName("target"), source);
  const signatureNode = node.childForFieldName("signature");
  if (!signatureNode) {
    throw new MapperError("parser: extern function missing signature");
  }
  const signature = parseFunctionSignature(signatureNode, source, ctx);
  const body = parseHostCodeBlock(node.childForFieldName("body"), source);

  const fn = annotateStatement(
    AST.functionDefinition(
      signature.name,
      signature.params,
      AST.blockExpression([]),
      signature.returnType,
      signature.genericParams,
      signature.whereClause,
      false,
      false,
    ),
    signatureNode,
  ) as FunctionDefinition;

  return annotate(AST.externFunctionBody(target, fn, body), node) as ExternFunctionBody;
}

function parseHostTarget(node: Node | null | undefined, source: string): HostTarget {
  if (!node) {
    throw new MapperError("parser: missing host target");
  }
  const value = sliceText(node, source).trim();
  switch (value) {
    case "go":
      return "go";
    case "crystal":
      return "crystal";
    case "typescript":
      return "typescript";
    case "python":
      return "python";
    case "ruby":
      return "ruby";
    default:
      throw new MapperError("parser: unsupported host target");
  }
}

function parseHostCodeBlock(node: Node | null | undefined, source: string): string {
  if (!node || node.type !== "host_code_block") {
    throw new MapperError("parser: expected host_code_block node");
  }
  const start = node.startIndex;
  const end = node.endIndex;
  if (start < 0 || end > source.length || start >= end) {
    throw new MapperError("parser: invalid host code block range");
  }
  const inner = source.slice(start + 1, end - 1).trim();
  return inner;
}
