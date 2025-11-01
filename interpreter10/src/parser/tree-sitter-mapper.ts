import type { SyntaxNode } from "web-tree-sitter";

import * as AST from "../ast";
import type {
  ArrayPattern,
  AssignmentExpression,
  BinaryExpression,
  BlockExpression,
  BreakStatement,
  ContinueStatement,
  DynImportStatement,
  ExternFunctionBody,
  Expression,
  FunctionCall,
  FunctionDefinition,
  FunctionParameter,
  FunctionSignature,
  GenericParameter,
  HostTarget,
  Identifier,
  ImportSelector,
  ImportStatement,
  InterfaceConstraint,
  InterfaceDefinition,
  IteratorLiteral,
  LambdaExpression,
  Literal,
  MatchClause,
  MethodsDefinition,
  Module,
  OrClause,
  Pattern,
  PreludeStatement,
  RaiseStatement,
  ReturnStatement,
  Statement,
  StructDefinition,
  StructFieldDefinition,
  StructFieldInitializer,
  StructPattern,
  StructPatternField,
  TypeExpression,
  UnionDefinition,
  WhereClauseConstraint,
} from "../ast";

class MapperError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "MapperError";
  }
}
type Node = SyntaxNode;

function sliceText(node: Node | null | undefined, source: string): string {
  if (!node) return "";
  const start = node.startIndex;
  const end = node.endIndex;
  if (start < 0 || end < start || end > source.length) {
    return "";
  }
  return source.slice(start, end);
}

function firstNamedChild(node: Node | null | undefined): Node | null {
  if (!node) return null;
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    return child;
  }
  return null;
}

function nextNamedSibling(parent: Node | null | undefined, currentIndex: number): Node | null {
  if (!parent) return null;
  for (let i = currentIndex + 1; i < parent.namedChildCount; i++) {
    const sibling = parent.namedChild(i);
    if (!sibling || !sibling.isNamed || isIgnorableNode(sibling)) continue;
    return sibling;
  }
  return null;
}

function hasLeadingPrivate(node: Node | null | undefined): boolean {
  if (!node) return false;
  for (let i = 0; i < node.childCount; i++) {
    const child = node.child(i);
    if (!child || isIgnorableNode(child)) continue;
    if (child.type === "private") {
      return true;
    }
    if (child.isNamed) break;
  }
  return false;
}

function sameNode(a: Node | null | undefined, b: Node | null | undefined): boolean {
  if (!a || !b) return false;
  return a.type === b.type && a.startIndex === b.startIndex && a.endIndex === b.endIndex;
}

function isIgnorableNode(node: Node | null | undefined): boolean {
  if (!node) return false;
  switch (node.type) {
    case "comment":
    case "line_comment":
    case "block_comment":
      return true;
    default:
      return false;
  }
}

function findIdentifier(node: Node | null | undefined, source: string): Identifier | null {
  if (!node) return null;
  if (isIgnorableNode(node)) return null;
  if (node.type === "identifier") {
    return AST.identifier(sliceText(node, source));
  }
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    const result = findIdentifier(child, source);
    if (result) return result;
  }
  return null;
}

function parseIdentifier(node: Node | null | undefined, source: string): Identifier {
  if (!node || node.type !== "identifier") {
    throw new MapperError("parser: expected identifier");
  }
  return AST.identifier(sliceText(node, source));
}

function identifiersToStrings(ids: Identifier[]): string[] {
  return ids.map(id => id.name);
}

function pruneUndefined<T>(value: T): T {
  if (Array.isArray(value)) {
    for (let i = 0; i < value.length; i++) {
      value[i] = pruneUndefined(value[i]);
    }
    return value;
  }
  if (value && typeof value === "object") {
    const record = value as Record<string, unknown>;
    for (const key of Object.keys(record)) {
      const entry = record[key];
      if (entry === undefined) {
        delete record[key];
      } else {
        record[key] = pruneUndefined(entry);
      }
    }
    return value;
  }
  return value;
}

const INFIX_OPERATOR_SETS = new Map<string, string[]>([
  ["logical_or_expression", ["||"]],
  ["logical_and_expression", ["&&"]],
  ["bitwise_or_expression", ["|"]],
  ["bitwise_xor_expression", ["\\xor"]],
  ["bitwise_and_expression", ["&"]],
  ["equality_expression", ["==", "!="]],
  ["comparison_expression", [">", "<", ">=", "<="]],
  ["shift_expression", ["<<", ">>"]],
  ["additive_expression", ["+", "-"]],
  ["multiplicative_expression", ["*", "/", "%"]],
  ["exponent_expression", ["**"]],
]);

export function mapSourceFile(root: Node, source: string): Module {
  if (!root) {
    throw new MapperError("parser: missing root node");
  }
  if (root.type !== "source_file") {
    throw new MapperError(`parser: unexpected root node ${root.type}`);
  }
  if ((root as unknown as { hasError?: boolean }).hasError) {
    throw new MapperError("parser: syntax errors present");
  }

  let packageStmt: AST.PackageStatement | undefined;
  const imports: ImportStatement[] = [];
  const body: Statement[] = [];

  for (let i = 0; i < root.namedChildCount; i++) {
    const node = root.namedChild(i);
    if (!node || isIgnorableNode(node)) continue;
    switch (node.type) {
      case "package_statement":
        packageStmt = parsePackageStatement(node, source);
        break;
      case "import_statement": {
        const kindNode = node.childForFieldName("kind");
        if (!kindNode) {
          throw new MapperError("parser: import missing kind");
        }
        const pathNode = node.childForFieldName("path");
        const path = parseQualifiedIdentifier(pathNode, source);
        const clauseNode = node.childForFieldName("clause");
        const clause = parseImportClause(clauseNode, source);
        if (kindNode.type === "import") {
          imports.push(AST.importStatement(path, clause.isWildcard, clause.selectors, clause.alias));
        } else if (kindNode.type === "dynimport") {
          body.push(AST.dynImportStatement(path, clause.isWildcard, clause.selectors, clause.alias));
        } else {
          throw new MapperError(`parser: unsupported import kind ${kindNode.type}`);
        }
        break;
      }
      default: {
        if (!node.isNamed) continue;
        const stmt = parseStatement(node, source);
        if (!stmt) {
          throw new MapperError(`parser: unsupported top-level node ${node.type}`);
        }
        if (stmt.type === "LambdaExpression" && body.length > 0) {
          const prev = body[body.length - 1];
          if (prev.type === "FunctionCall") {
            const call = prev as FunctionCall;
            if (call.arguments.length === 0 || call.arguments[call.arguments.length - 1] !== stmt) {
              call.arguments.push(stmt);
            }
            call.isTrailingLambda = true;
            continue;
          }
          if ((prev as Expression).type) {
            const call = AST.functionCall(prev as Expression, [], undefined, true);
            call.arguments.push(stmt);
            body[body.length - 1] = call;
            continue;
          }
        }
        body.push(stmt);
        break;
      }
    }
  }

  const module = AST.module(body, imports, packageStmt);
  return pruneUndefined(module);
}

function parsePackageStatement(node: Node, source: string): AST.PackageStatement {
  const parts: Identifier[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    parts.push(parseIdentifier(child, source));
  }
  if (parts.length === 0) {
    throw new MapperError("parser: empty package statement");
  }
  return AST.packageStatement(parts, false);
}

function parseQualifiedIdentifier(node: Node | null | undefined, source: string): Identifier[] {
  if (!node || node.type !== "qualified_identifier") {
    throw new MapperError("parser: expected qualified identifier");
  }
  const parts: Identifier[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    parts.push(parseIdentifier(child, source));
  }
  if (parts.length === 0) {
    throw new MapperError("parser: empty qualified identifier");
  }
  return parts;
}

function parseImportClause(node: Node | null | undefined, source: string): {
  isWildcard: boolean;
  selectors?: ImportSelector[];
  alias?: Identifier;
} {
  if (!node) {
    return { isWildcard: false };
  }

  let isWildcard = false;
  const selectors: ImportSelector[] = [];
  let alias: Identifier | undefined;

  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    switch (child.type) {
      case "import_selector":
        selectors.push(parseImportSelector(child, source));
        break;
      case "import_wildcard_clause":
        isWildcard = true;
        break;
      case "identifier":
        if (alias) {
          throw new MapperError("parser: multiple aliases in import clause");
        }
        alias = parseIdentifier(child, source);
        break;
      default:
        throw new MapperError(`parser: unsupported import clause node ${child.type}`);
    }
  }

  if (isWildcard && selectors.length > 0) {
    throw new MapperError("parser: wildcard import cannot include selectors");
  }
  if (alias && selectors.length > 0) {
    throw new MapperError("parser: alias cannot be combined with selectors");
  }

  return { isWildcard, selectors: selectors.length ? selectors : undefined, alias };
}

function parseImportSelector(node: Node, source: string): ImportSelector {
  if (node.type !== "import_selector") {
    throw new MapperError("parser: expected import_selector node");
  }
  if (node.namedChildCount === 0) {
    throw new MapperError("parser: empty import selector");
  }
  const name = parseIdentifier(node.namedChild(0), source);
  let alias: Identifier | undefined;
  if (node.namedChildCount > 1) {
    alias = parseIdentifier(node.namedChild(1), source);
  }
  return AST.importSelector(name, alias);
}

function parseStatement(node: Node, source: string): Statement | null {
  switch (node.type) {
    case "expression_statement": {
      const exprNode = firstNamedChild(node);
      if (!exprNode) {
        throw new MapperError("parser: expression statement missing expression");
      }
      return parseExpression(exprNode, source);
    }
    case "return_statement": {
      const valueNode = firstNamedChild(node);
      if (!valueNode) {
        return AST.returnStatement();
      }
      const expr = parseExpression(valueNode, source);
      return AST.returnStatement(expr);
    }
    case "while_statement": {
      if (node.namedChildCount < 2) {
        throw new MapperError("parser: malformed while statement");
      }
      const condition = parseExpression(node.namedChild(0), source);
      const body = parseBlock(node.namedChild(1), source);
      return AST.whileLoop(condition, body);
    }
    case "for_statement": {
      if (node.namedChildCount < 3) {
        throw new MapperError("parser: malformed for statement");
      }
      const pattern = parsePattern(node.namedChild(0), source);
      const iterable = parseExpression(node.namedChild(1), source);
      const body = parseBlock(node.namedChild(2), source);
      return AST.forLoop(pattern, iterable, body);
    }
    case "break_statement": {
      const labelNode = node.childForFieldName("label");
      const valueNode = node.childForFieldName("value");
      const label = labelNode ? parseLabel(labelNode, source) : undefined;
      const value = valueNode ? parseExpression(valueNode, source) : undefined;
      return AST.breakStatement(label, value);
    }
    case "continue_statement":
      return AST.continueStatement();
    case "raise_statement": {
      const valueNode = firstNamedChild(node);
      if (!valueNode) {
        throw new MapperError("parser: raise statement missing expression");
      }
      return AST.raiseStatement(parseExpression(valueNode, source));
    }
    case "rethrow_statement":
      return AST.rethrowStatement();
    case "struct_definition":
      return parseStructDefinition(node, source);
    case "methods_definition":
      return parseMethodsDefinition(node, source);
    case "implementation_definition":
      return parseImplementationDefinition(node, source);
    case "named_implementation_definition":
      return parseNamedImplementationDefinition(node, source);
    case "union_definition":
      return parseUnionDefinition(node, source);
    case "interface_definition":
      return parseInterfaceDefinition(node, source);
    case "prelude_statement":
      return parsePreludeStatement(node, source);
    case "extern_function":
      return parseExternFunction(node, source);
    case "function_definition":
      return parseFunctionDefinition(node, source);
    default:
      return null;
  }
}

function parseBlock(node: Node | null | undefined, source: string): BlockExpression {
  if (!node) {
    return AST.blockExpression([]);
  }

  const statements: Statement[] = [];

  for (let i = 0; i < node.namedChildCount; ) {
    const child = node.namedChild(i);
    i++;
    if (!child || !child.isNamed || isIgnorableNode(child)) continue;
    if (node.fieldNameForChild(i - 1) === "binding") continue;

    let stmt: Statement | null = null;
    if (child.type === "break_statement") {
      stmt = parseStatement(child, source);
      const breakStmt = stmt as AST.BreakStatement | undefined;
      if (breakStmt && breakStmt.type === "BreakStatement" && !breakStmt.value) {
        const next = nextNamedSibling(node, i - 1);
        if (next && next.type === "expression_statement") {
          const exprNode = firstNamedChild(next);
          if (exprNode) {
            breakStmt.value = parseExpression(exprNode, source);
            i++;
          }
        }
      }
    } else {
      stmt = parseStatement(child, source);
    }

    if (!stmt) continue;

    if (stmt.type === "LambdaExpression" && statements.length > 0) {
      const prev = statements[statements.length - 1];
      if (prev.type === "FunctionCall") {
        const call = prev as FunctionCall;
        if (call.arguments.length === 0 || call.arguments[call.arguments.length - 1] !== stmt) {
          call.arguments.push(stmt);
        }
        call.isTrailingLambda = true;
        continue;
      }
      if ((prev as Expression).type) {
        const call = AST.functionCall(prev as Expression, [], undefined, true);
        call.arguments.push(stmt);
        statements[statements.length - 1] = call;
        continue;
      }
    }

    statements.push(stmt);
  }

  return AST.blockExpression(statements);
}

function parseExpression(node: Node | null | undefined, source: string): Expression {
  if (!node) {
    throw new MapperError("parser: nil expression node");
  }

  switch (node.type) {
    case "identifier":
      return parseIdentifier(node, source);
    case "number_literal":
      return parseNumberLiteral(node, source);
    case "boolean_literal":
      return parseBooleanLiteral(node, source);
    case "nil_literal":
      return parseNilLiteral(node, source);
    case "string_literal":
      return parseStringLiteral(node, source);
    case "character_literal":
      return parseCharLiteral(node, source);
    case "array_literal":
      return parseArrayLiteral(node, source);
    case "struct_literal":
      return parseStructLiteral(node, source);
    case "block":
      return parseBlock(node, source);
    case "do_expression":
      return parseDoExpression(node, source);
    case "lambda_expression":
      return parseLambdaExpression(node, source);
    case "postfix_expression":
    case "call_target":
      return parsePostfixExpression(node, source);
    case "member_access": {
      if (node.namedChildCount < 2) {
        throw new MapperError("parser: malformed member access");
      }
      const objectExpr = parseExpression(node.namedChild(0), source);
      const memberExpr = parseExpression(node.namedChild(1), source);
      return AST.memberAccessExpression(objectExpr, memberExpr);
    }
    case "proc_expression":
      return parseProcExpression(node, source);
    case "spawn_expression":
      return parseSpawnExpression(node, source);
    case "breakpoint_expression":
      return parseBreakpointExpression(node, source);
    case "handling_expression":
      return parseHandlingExpression(node, source);
    case "rescue_expression":
      return parseRescueExpression(node, source);
    case "ensure_expression":
      return parseEnsureExpression(node, source);
    case "if_expression":
      return parseIfExpression(node, source);
    case "match_expression":
      return parseMatchExpression(node, source);
    case "range_expression":
      return parseRangeExpression(node, source);
    case "assignment_expression":
      return parseAssignmentExpression(node, source);
    case "unary_expression":
      return parseUnaryExpression(node, source);
    case "implicit_member_expression":
      return parseImplicitMemberExpression(node, source);
    case "placeholder_expression":
      return parsePlaceholderExpression(node, source);
    case "topic_reference":
      return AST.topicReferenceExpression();
    case "interpolated_string":
      return parseInterpolatedString(node, source);
    case "iterator_literal":
      return parseIteratorLiteral(node, source);
    case "parenthesized_expression": {
      const child = firstNamedChild(node);
      if (child) {
        return parseExpression(child, source);
      }
      throw new MapperError("parser: empty parenthesized expression");
    }
    case "pipe_expression":
      return parsePipeExpression(node, source);
    case "matchable_expression": {
      const child = firstNamedChild(node);
      if (child) return parseExpression(child, source);
      break;
    }
  }

  if (INFIX_OPERATOR_SETS.has(node.type)) {
    const operators = INFIX_OPERATOR_SETS.get(node.type)!;
    return parseInfixExpression(node, source, operators);
  }

  const child = firstNamedChild(node);
  if (child && child !== node) {
    return parseExpression(child, source);
  }

  const fallbackId = findIdentifier(node, source);
  if (fallbackId) {
    return fallbackId;
  }

  throw new MapperError(`parser: unsupported expression kind ${node.type}`);
}

function parseNumberLiteral(node: Node, source: string): Expression {
  const content = sliceText(node, source);
  if (!content) {
    throw new MapperError("parser: empty number literal");
  }

  let base = content;
  let integerType: AST.IntegerLiteral["integerType"] | undefined;
  let floatType: AST.FloatLiteral["floatType"] | undefined;

  const underscoreIdx = content.lastIndexOf("_");
  if (underscoreIdx > 0) {
    const suffix = content.slice(underscoreIdx + 1);
    if (isNumericSuffix(suffix)) {
      base = content.slice(0, underscoreIdx);
      if (suffix === "f32" || suffix === "f64") {
        floatType = suffix as AST.FloatLiteral["floatType"];
      } else {
        integerType = suffix as AST.IntegerLiteral["integerType"];
      }
    }
  }

  const sanitized = base.replace(/_/g, "");
  if (base.includes(".") || /[eE]/.test(base) || floatType) {
    const value = Number(sanitized);
    if (!Number.isFinite(value)) {
      throw new MapperError(`parser: invalid number literal ${content}`);
    }
    return AST.floatLiteral(value, floatType);
  }

  let numberValue: number | bigint;
  try {
    numberValue = BigInt(sanitized);
  } catch {
    const value = Number(sanitized);
    if (!Number.isFinite(value)) {
      throw new MapperError(`parser: invalid number literal ${content}`);
    }
    numberValue = value;
  }

  if (typeof numberValue === "bigint") {
    if (numberValue <= BigInt(Number.MAX_SAFE_INTEGER) && numberValue >= BigInt(Number.MIN_SAFE_INTEGER)) {
      numberValue = Number(numberValue);
    }
  }

  return AST.integerLiteral(numberValue, integerType);
}

function parseBooleanLiteral(node: Node, source: string): Expression {
  const value = sliceText(node, source).trim();
  if (value === "true") return AST.booleanLiteral(true);
  if (value === "false") return AST.booleanLiteral(false);
  throw new MapperError(`parser: invalid boolean literal ${value}`);
}

function parseNilLiteral(node: Node, source: string): Expression {
  const value = sliceText(node, source).trim();
  if (value !== "nil") {
    throw new MapperError(`parser: invalid nil literal ${value}`);
  }
  return AST.nilLiteral();
}

function parseStringLiteral(node: Node, source: string): Expression {
  const raw = sliceText(node, source);
  try {
    return AST.stringLiteral(JSON.parse(raw));
  } catch (error) {
    throw new MapperError(`parser: invalid string literal ${raw}: ${error}`);
  }
}

function parseCharLiteral(node: Node, source: string): Expression {
  const raw = sliceText(node, source);
  let value: string;
  try {
    value = JSON.parse(raw.replace(/^'|'+$/g, match => (match === "'" ? '"' : match)));
  } catch (error) {
    throw new MapperError(`parser: invalid character literal ${raw}: ${error}`);
  }
  if (Array.from(value).length !== 1) {
    throw new MapperError(`parser: character literal ${raw} must resolve to a single rune`);
  }
  return AST.charLiteral(value);
}

function parseArrayLiteral(node: Node, source: string): Expression {
  const elements: Expression[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || !child.isNamed || isIgnorableNode(child)) continue;
    elements.push(parseExpression(child, source));
  }
  return AST.arrayLiteral(elements);
}

function parseStructLiteral(node: Node, source: string): Expression {
  const typeNode = node.childForFieldName("type");
  if (!typeNode) {
    throw new MapperError("parser: struct literal missing type");
  }
  const parts = parseQualifiedIdentifier(typeNode, source);
  if (parts.length === 0) {
    throw new MapperError("parser: invalid struct literal type");
  }
  let structType = parts[parts.length - 1];
  if (parts.length > 1) {
    structType = AST.identifier(identifiersToStrings(parts).join("."));
  }

  const typeArgsNode = node.childForFieldName("type_arguments");
  const typeArguments = typeArgsNode ? parseTypeArgumentList(typeArgsNode, source) ?? undefined : undefined;

  const fields: StructFieldInitializer[] = [];
  const functionalUpdates: Expression[] = [];

  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    const fieldName = node.fieldNameForChild(i);
    if (
      fieldName === "type" ||
      fieldName === "type_arguments" ||
      sameNode(child, typeNode) ||
      (typeArgsNode && sameNode(child, typeArgsNode))
    ) {
      continue;
    }

    let elem: Node | null = child;
    if (child.type === "struct_literal_element") {
      elem = firstNamedChild(child);
      if (!elem) continue;
    }

    switch (elem.type) {
      case "struct_literal_field": {
        const nameNode = elem.childForFieldName("name");
        if (!nameNode) {
          throw new MapperError("parser: struct literal field missing name");
        }
        const name = parseIdentifier(nameNode, source);
        const valueNode = elem.childForFieldName("value");
        if (!valueNode) {
          throw new MapperError("parser: struct literal field missing value");
        }
        const value = parseExpression(valueNode, source);
        fields.push(AST.structFieldInitializer(value, name, false));
        break;
      }
      case "struct_literal_shorthand_field": {
        let nameNode = elem.childForFieldName("name");
        if (!nameNode) {
          nameNode = firstNamedChild(elem);
        }
        if (!nameNode) {
          throw new MapperError("parser: struct literal shorthand missing name");
        }
        const name = parseIdentifier(nameNode, source);
        fields.push(AST.structFieldInitializer(AST.identifier(name.name), name, true));
        break;
      }
      case "struct_literal_spread": {
        const exprNode = firstNamedChild(elem);
        if (!exprNode) {
          throw new MapperError("parser: struct spread missing expression");
        }
        functionalUpdates.push(parseExpression(exprNode, source));
        break;
      }
      default: {
        fields.push(AST.structFieldInitializer(parseExpression(elem, source), undefined, false));
        break;
      }
    }
  }

  const positional = fields.some(field => !field.name);

  return AST.structLiteral(fields, positional, structType, functionalUpdates.length ? functionalUpdates : undefined, typeArguments ?? undefined);
}

function applyGenericType(base: TypeExpression | null, args: TypeExpression[]): TypeExpression | null {
  if (!base) return null;
  if (base.type === "NullableTypeExpression") {
    const inner = applyGenericType(base.innerType, args);
    return AST.nullableTypeExpression(inner ?? base.innerType);
  }
  if (base.type === "ResultTypeExpression") {
    const inner = applyGenericType(base.innerType, args);
    return AST.resultTypeExpression(inner ?? base.innerType);
  }
  return AST.genericTypeExpression(base, args);
}

function parseReturnType(node: Node | null | undefined, source: string): TypeExpression | undefined {
  const expr = parseTypeExpression(node, source);
  return expr ?? undefined;
}

function parseTypeExpression(node: Node | null | undefined, source: string): TypeExpression | null {
  if (!node) return null;
  switch (node.type) {
    case "return_type":
    case "type_expression":
    case "type_prefix":
    case "type_atom": {
      if (node.namedChildCount === 0) break;
      const child = firstNamedChild(node);
      if (child && child !== node) {
        const expr = parseTypeExpression(child, source);
        if (node.type === "type_prefix" && expr) {
          let text = sliceText(node, source).trim();
          let result: TypeExpression | null = expr;
          while (text.startsWith("?") || text.startsWith("!")) {
            if (text.startsWith("?")) {
              text = text.slice(1);
              result = result ? AST.nullableTypeExpression(result) : result;
            } else if (text.startsWith("!")) {
              text = text.slice(1);
              result = result ? AST.resultTypeExpression(result) : result;
            } else {
              break;
            }
          }
          return result;
        }
        return expr;
      }
      break;
    }
    case "type_suffix": {
      if (node.namedChildCount > 1) {
        const base = parseTypeExpression(node.namedChild(0), source);
        const args: TypeExpression[] = [];
        for (let i = 1; i < node.namedChildCount; i++) {
          const child = node.namedChild(i);
          if (!child || !child.isNamed || isIgnorableNode(child)) continue;
          if (child.type === "type_arguments") {
            const typeArgs = parseTypeArgumentList(child, source);
            if (typeArgs) args.push(...typeArgs);
            continue;
          }
          const arg = parseTypeExpression(child, source);
          if (arg) args.push(arg);
        }
        if (base && args.length > 0) {
          return applyGenericType(base, args);
        }
      }
      const child = firstNamedChild(node);
      if (child && child !== node) {
        return parseTypeExpression(child, source);
      }
      break;
    }
    case "type_arrow": {
      if (node.namedChildCount >= 2) {
        const [paramTypes, ok] = parseFunctionParameterTypes(node.namedChild(0), source);
        if (ok && paramTypes) {
          const returnExpr = parseTypeExpression(node.namedChild(1), source);
          if (returnExpr) {
            return AST.functionTypeExpression(paramTypes, returnExpr);
          }
        }
      }
      const child = firstNamedChild(node);
      if (child && child !== node) {
        return parseTypeExpression(child, source);
      }
      break;
    }
    case "type_generic_application": {
      if (node.namedChildCount === 0) break;
      const base = parseTypeExpression(node.namedChild(0), source);
      if (!base) return null;
      const args: TypeExpression[] = [];
      for (let i = 1; i < node.namedChildCount; i++) {
        const arg = parseTypeExpression(node.namedChild(i), source);
        if (arg) args.push(arg);
      }
      if (args.length === 0) return base;
      return applyGenericType(base, args) ?? base;
    }
    case "type_union": {
      const members: TypeExpression[] = [];
      for (let i = 0; i < node.namedChildCount; i++) {
        const child = node.namedChild(i);
        const member = parseTypeExpression(child, source);
        if (member) members.push(member);
      }
      if (members.length === 1) return members[0];
      if (members.length > 1) {
        return AST.unionTypeExpression(members);
      }
      break;
    }
    case "type_identifier": {
      const child = firstNamedChild(node);
      if (child && child !== node) {
        return parseTypeExpression(child, source);
      }
      break;
    }
    case "identifier":
      return AST.simpleTypeExpression(parseIdentifier(node, source));
    case "qualified_identifier": {
      const parts = parseQualifiedIdentifier(node, source);
      if (parts.length === 0) return null;
      if (parts.length === 1) return AST.simpleTypeExpression(parts[0]);
      const name = AST.identifier(identifiersToStrings(parts).join("."));
      return AST.simpleTypeExpression(name);
    }
    default: {
      const child = firstNamedChild(node);
      if (child && child !== node) {
        const expr = parseTypeExpression(child, source);
        if (expr) return expr;
      }
      break;
    }
  }
  const text = sliceText(node, source).trim();
  if (text === "") {
    return null;
  }
  return AST.simpleTypeExpression(AST.identifier(text.replace(/\s+/g, "")));
}

function parseTypeArgumentList(node: Node | null | undefined, source: string): TypeExpression[] | null {
  if (!node) return null;
  const args: TypeExpression[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || !child.isNamed || isIgnorableNode(child)) continue;
    const typeExpr = parseTypeExpression(child, source);
    if (!typeExpr) {
      throw new MapperError(`parser: unsupported type argument kind ${child.type}`);
    }
    args.push(typeExpr);
  }
  return args;
}

function parseFunctionParameterTypes(node: Node | null | undefined, source: string): [TypeExpression[] | null, boolean] {
  if (!node) {
    return [null, false];
  }

  let current: Node | null = node;
  while (current) {
    if (current.type === "parenthesized_type") {
      if (current.namedChildCount === 0) {
        return [null, false];
      }
      const params: TypeExpression[] = [];
      for (let i = 0; i < current.namedChildCount; i++) {
        const child = current.namedChild(i);
        if (!child || isIgnorableNode(child)) continue;
        const param = parseTypeExpression(child, source);
        if (!param) return [null, false];
        params.push(param);
      }
      return [params, true];
    }
    if (
      current.type !== "type_suffix" &&
      current.type !== "type_prefix" &&
      current.type !== "type_atom"
    ) {
      break;
    }
    const child = firstNamedChild(current);
    if (child && child !== current) {
      current = child;
      continue;
    }
    break;
  }

  const param = parseTypeExpression(node, source);
  if (!param) {
    return [null, false];
  }
  return [[param], true];
}

function parseTypeParameters(node: Node | null | undefined, source: string): GenericParameter[] | undefined {
  if (!node) return undefined;
  switch (node.type) {
    case "declaration_type_parameters":
      if (node.namedChildCount === 0) return undefined;
      return parseTypeParameters(node.namedChild(0), source);
    case "type_parameter_list": {
      const params: GenericParameter[] = [];
      for (let i = 0; i < node.namedChildCount; i++) {
        const child = node.namedChild(i);
        if (!child || isIgnorableNode(child) || child.type !== "type_parameter") continue;
        params.push(parseTypeParameter(child, source));
      }
      return params.length ? params : undefined;
    }
    case "generic_parameter_list": {
      const params: GenericParameter[] = [];
      for (let i = 0; i < node.namedChildCount; i++) {
        const child = node.namedChild(i);
        if (!child || isIgnorableNode(child) || child.type !== "generic_parameter") continue;
        params.push(parseGenericParameter(child, source));
      }
      return params.length ? params : undefined;
    }
    default:
      throw new MapperError(`parser: unsupported type parameter node ${node.type}`);
  }
}

function parseTypeParameter(node: Node, source: string): GenericParameter {
  if (node.type !== "type_parameter") {
    throw new MapperError("parser: expected type_parameter node");
  }
  return buildGenericParameter(node, source);
}

function parseGenericParameter(node: Node, source: string): GenericParameter {
  if (node.type !== "generic_parameter") {
    throw new MapperError("parser: expected generic_parameter node");
  }
  return buildGenericParameter(node, source);
}

function buildGenericParameter(node: Node, source: string): GenericParameter {
  let nameNode: Node | null = firstNamedChild(node);
  if (!nameNode || nameNode.type !== "identifier") {
    throw new MapperError("parser: generic parameter missing identifier");
  }
  const name = parseIdentifier(nameNode, source);

  let constraints: InterfaceConstraint[] | undefined;
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child) || sameNode(child, nameNode)) continue;
    const typeExprs = parseTypeBoundList(child, source);
    constraints = typeExprs?.map(expr => AST.interfaceConstraint(expr));
    if (constraints && constraints.length > 0) break;
  }

  return AST.genericParameter(name, constraints);
}

function parseTypeBoundList(node: Node | null | undefined, source: string): TypeExpression[] | undefined {
  if (!node) return undefined;
  const bounds: TypeExpression[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    if (child.type === "type_bound_list") {
      const nested = parseTypeBoundList(child, source);
      if (nested) bounds.push(...nested);
      continue;
    }
    const expr = parseReturnType(child, source);
    if (expr) bounds.push(expr);
  }
  if (bounds.length === 0) {
    throw new MapperError("parser: empty type bound list");
  }
  return bounds;
}

function parseWhereClause(node: Node | null | undefined, source: string): WhereClauseConstraint[] | undefined {
  if (!node) return undefined;
  if (node.type !== "where_clause") {
    throw new MapperError(`parser: expected where clause, found ${node.type}`);
  }
  const constraints: WhereClauseConstraint[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child) || child.type !== "where_constraint") continue;
    constraints.push(parseWhereConstraint(child, source));
  }
  return constraints.length ? constraints : undefined;
}

function parseWhereConstraint(node: Node, source: string): WhereClauseConstraint {
  if (node.type !== "where_constraint") {
    throw new MapperError("parser: expected where_constraint node");
  }
  if (node.namedChildCount === 0) {
    throw new MapperError("parser: empty where constraint");
  }
  const nameNode = firstNamedChild(node);
  if (!nameNode || nameNode.type !== "identifier") {
    throw new MapperError("parser: where constraint missing identifier");
  }
  const name = parseIdentifier(nameNode, source);
  let constraintNode: Node | null = null;
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child) || sameNode(child, nameNode)) continue;
    constraintNode = child;
    break;
  }
  const typeExprs = parseTypeBoundList(constraintNode, source);
  if (!typeExprs || typeExprs.length === 0) {
    throw new MapperError("parser: where constraint missing bounds");
  }
  const interfaceConstraints = typeExprs.map(expr => AST.interfaceConstraint(expr));
  return AST.whereClauseConstraint(name, interfaceConstraints);
}

function isLiteralExpression(expr: Expression): expr is Literal {
  switch (expr.type) {
    case "StringLiteral":
    case "IntegerLiteral":
    case "FloatLiteral":
    case "BooleanLiteral":
    case "NilLiteral":
    case "CharLiteral":
    case "ArrayLiteral":
      return true;
    default:
      return false;
  }
}

function parsePattern(node: Node | null | undefined, source: string): Pattern {
  if (!node) {
    throw new MapperError("parser: nil pattern");
  }

  if (node.type === "pattern" || node.type === "pattern_base") {
    if (node.namedChildCount === 0) {
      const text = sliceText(node, source).trim();
      if (text === "_") {
        return AST.wildcardPattern();
      }
      for (let i = 0; i < node.childCount; i++) {
        const child = node.child(i);
        if (!child || isIgnorableNode(child)) continue;
        if (child.isNamed) {
          return parsePattern(child, source);
        }
        if (sliceText(child, source).trim() === "_") {
          return AST.wildcardPattern();
        }
      }
      throw new MapperError(`parser: empty ${node.type}`);
    }
    return parsePattern(node.namedChild(0), source);
  }

  switch (node.type) {
    case "identifier":
      return parseIdentifier(node, source);
    case "_":
      return AST.wildcardPattern();
    case "literal_pattern":
      return parseLiteralPattern(node, source);
    case "struct_pattern":
      return parseStructPattern(node, source);
    case "array_pattern":
      return parseArrayPattern(node, source);
    case "parenthesized_pattern": {
      const inner = firstNamedChild(node);
      if (inner) {
        return parsePattern(inner, source);
      }
      throw new MapperError("parser: empty parenthesized pattern");
    }
    case "typed_pattern":
      if (node.namedChildCount < 2) {
        throw new MapperError("parser: malformed typed pattern");
      }
      const innerPattern = parsePattern(node.namedChild(0), source);
      const typeExpr = parseTypeExpression(node.namedChild(1), source);
      if (!typeExpr) {
        throw new MapperError("parser: typed pattern missing type expression");
      }
      return AST.typedPattern(innerPattern, typeExpr);
    case "pattern":
    case "pattern_base":
      return parsePattern(node.namedChild(0), source);
    default:
      throw new MapperError(`parser: unsupported pattern kind ${node.type}`);
  }
}

function parseLiteralPattern(node: Node, source: string): Pattern {
  const literalNode = firstNamedChild(node);
  if (!literalNode) {
    throw new MapperError("parser: literal pattern missing literal");
  }
  const literalExpr = parseExpression(literalNode, source);
  if (!isLiteralExpression(literalExpr)) {
    throw new MapperError(`parser: literal pattern must contain literal, found ${literalExpr.type}`);
  }
  return AST.literalPattern(literalExpr);
}

function parseStructPattern(node: Node, source: string): Pattern {
  let structType: Identifier | undefined;
  const typeNode = node.childForFieldName("type");
  if (typeNode) {
    const parts = parseQualifiedIdentifier(typeNode, source);
    if (parts.length === 0) {
      throw new MapperError("parser: struct pattern type missing identifier");
    }
    structType = parts[parts.length - 1];
    if (parts.length > 1) {
      structType = AST.identifier(identifiersToStrings(parts).join("."));
    }
  }

  const fields: StructPatternField[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    const fieldName = node.fieldNameForChild(i);
    if (fieldName === "type" || (typeNode && sameNode(child, typeNode))) {
      continue;
    }
    let elem: Node | null = child;
    if (child.type === "struct_pattern_element") {
      elem = firstNamedChild(child);
      if (!elem) continue;
    }

    if (elem.type === "struct_pattern_field") {
      if (!elem.childForFieldName("binding") && !elem.childForFieldName("value")) {
        const fieldNode = elem.childForFieldName("field");
        if (!fieldNode) {
          throw new MapperError("parser: struct pattern field missing identifier");
        }
        const pat = parseIdentifier(fieldNode, source);
        fields.push(AST.structPatternField(pat, undefined, undefined));
        continue;
      }
      fields.push(parseStructPatternField(elem, source));
      continue;
    }

    const pattern = parsePattern(elem, source);
    fields.push(AST.structPatternField(pattern, undefined, undefined));
  }

  const isPositional = fields.some(field => !field.fieldName);

  return AST.structPattern(fields, isPositional, structType);
}

function parseStructPatternField(node: Node, source: string): StructPatternField {
  if (node.type !== "struct_pattern_field") {
    throw new MapperError("parser: expected struct_pattern_field node");
  }

  let fieldName: Identifier | undefined;
  const nameNode = node.childForFieldName("field");
  if (nameNode) {
    fieldName = parseIdentifier(nameNode, source);
  }

  let binding: Identifier | undefined;
  const bindingNode = node.childForFieldName("binding");
  if (bindingNode) {
    binding = parseIdentifier(bindingNode, source);
  }

  let pattern: Pattern;
  const valueNode = node.childForFieldName("value");
  if (valueNode) {
    pattern = parsePattern(valueNode, source);
  } else if (binding) {
    pattern = binding;
  } else if (fieldName) {
    pattern = fieldName;
  } else {
    pattern = AST.wildcardPattern();
  }

  return AST.structPatternField(pattern, fieldName, binding);
}

function parseArrayPattern(node: Node, source: string): Pattern {
  const elements: Pattern[] = [];
  let rest: Pattern | undefined;

  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || isIgnorableNode(child)) continue;
    if (child.type === "array_pattern_rest") {
      if (rest) {
        throw new MapperError("parser: multiple array rest patterns");
      }
      rest = parseArrayPatternRest(child, source);
      continue;
    }
    elements.push(parsePattern(child, source));
  }

  return AST.arrayPattern(elements, rest);
}

function parseArrayPatternRest(node: Node, source: string): Pattern {
  if (node.namedChildCount === 0) {
    return AST.wildcardPattern();
  }
  return parsePattern(node.namedChild(0), source);
}

function parseImplicitMemberExpression(node: Node, source: string): Expression {
  const memberNode = node.childForFieldName("member") ?? firstNamedChild(node);
  if (!memberNode) {
    throw new MapperError("parser: implicit member missing identifier");
  }
  const member = parseIdentifier(memberNode, source);
  return AST.implicitMemberExpression(member);
}

function parsePlaceholderExpression(node: Node, source: string): Expression {
  const raw = sliceText(node, source).trim();
  if (raw === "@" || raw === "@0") {
    return AST.placeholderExpression();
  }
  if (raw.startsWith("@")) {
    const value = raw.slice(1);
    if (value === "") {
      return AST.placeholderExpression();
    }
    const index = Number.parseInt(value, 10);
    if (!Number.isInteger(index) || index <= 0) {
      throw new MapperError(`parser: invalid placeholder index ${raw}`);
    }
    return AST.placeholderExpression(index);
  }
  throw new MapperError(`parser: unsupported placeholder token ${raw}`);
}

function parseInterpolatedString(node: Node, source: string): Expression {
  const parts: (AST.StringLiteral | Expression)[] = [];
  for (let i = 0; i < node.childCount; i++) {
    const child = node.child(i);
    if (!child || isIgnorableNode(child)) continue;
    switch (child.type) {
      case "interpolation_text": {
        const text = sliceText(child, source);
        if (text !== "") {
          parts.push(AST.stringLiteral(text));
        }
        break;
      }
      case "string_interpolation": {
        const exprNode = child.childForFieldName("expression");
        if (!exprNode) {
          throw new MapperError("parser: interpolation missing expression");
        }
        parts.push(parseExpression(exprNode, source));
        break;
      }
      default:
        break;
    }
  }
  return AST.stringInterpolation(parts);
}

function parseIteratorLiteral(node: Node, source: string): IteratorLiteral {
  const bodyNode = node.childForFieldName("body");
  if (!bodyNode) {
    throw new MapperError("parser: iterator literal missing body");
  }
  const block = parseBlock(bodyNode, source);
  return AST.iteratorLiteral(block.body);
}

function parsePostfixExpression(node: Node, source: string): Expression {
  if (node.namedChildCount === 0) {
    throw new MapperError("parser: empty postfix expression");
  }

  let result = parseExpression(node.namedChild(0), source);
  let pendingTypeArgs: TypeExpression[] | null = null;
  let lastCall: FunctionCall | undefined;

  for (let i = 1; i < node.namedChildCount; i++) {
    const suffix = node.namedChild(i);
    switch (suffix.type) {
      case "member_access": {
        let memberNode = suffix.childForFieldName("member") ?? firstNamedChild(suffix);
        if (!memberNode) {
          throw new MapperError("parser: member access missing member");
        }

        let memberExpr: Expression;
        if (memberNode.type === "numeric_member") {
          const valueText = sliceText(memberNode, source);
          if (valueText === "") {
            throw new MapperError("parser: empty numeric member access");
          }
          const intValue = Number.parseInt(valueText, 10);
          if (!Number.isInteger(intValue)) {
            throw new MapperError(`parser: invalid numeric member ${valueText}`);
          }
          memberExpr = AST.integerLiteral(intValue);
        } else {
          memberExpr = parseExpression(memberNode, source);
        }
        result = AST.memberAccessExpression(result, memberExpr);
        lastCall = undefined;
        break;
      }
      case "type_arguments": {
        pendingTypeArgs = parseTypeArgumentList(suffix, source);
        lastCall = undefined;
        break;
      }
      case "index_suffix": {
        if (suffix.namedChildCount === 0) {
          throw new MapperError("parser: index expression missing index value");
        }
        if (suffix.namedChildCount > 1) {
          throw new MapperError("parser: slice expressions are not supported yet");
        }
        const indexExpr = parseExpression(suffix.namedChild(0), source);
        result = AST.indexExpression(result, indexExpr);
        lastCall = undefined;
        break;
      }
      case "call_suffix": {
        const args = parseCallArguments(suffix, source);
        const typeArgs = pendingTypeArgs ?? undefined;
        pendingTypeArgs = null;
        const callExpr = AST.functionCall(result, args, typeArgs, false);
        result = callExpr;
        lastCall = callExpr;
        break;
      }
      case "lambda_expression": {
        const lambdaExpr = parseLambdaExpression(suffix, source);
        const typeArgs = pendingTypeArgs ?? undefined;
        pendingTypeArgs = null;
        if (lastCall && !lastCall.isTrailingLambda) {
          lastCall.arguments.push(lambdaExpr);
          lastCall.isTrailingLambda = true;
          result = lastCall;
        } else {
          const callExpr = AST.functionCall(result, [], typeArgs, true);
          callExpr.arguments.push(lambdaExpr);
          result = callExpr;
          lastCall = callExpr;
        }
        break;
      }
      case "propagate_suffix": {
        if (pendingTypeArgs && pendingTypeArgs.length > 0) {
          throw new MapperError("parser: dangling type arguments before propagation");
        }
        result = AST.propagationExpression(result);
        lastCall = undefined;
        break;
      }
      default:
        throw new MapperError(`parser: unsupported postfix suffix ${suffix.type}`);
    }
  }

  if (pendingTypeArgs && pendingTypeArgs.length > 0) {
    throw new MapperError("parser: dangling type arguments in expression");
  }

  return result;
}

function parseCallArguments(node: Node, source: string): Expression[] {
  const args: Expression[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || !child.isNamed || isIgnorableNode(child)) continue;
    args.push(parseExpression(child, source));
  }
  return args;
}

function parsePipeExpression(node: Node, source: string): Expression {
  if (node.namedChildCount === 0) {
    throw new MapperError("parser: empty pipe expression");
  }
  let result = parseExpression(node.namedChild(0), source);
  for (let i = 1; i < node.namedChildCount; i++) {
    const stepNode = node.namedChild(i);
    const stepExpr = parseExpression(stepNode, source);
    result = AST.binaryExpression("|>", result, stepExpr);
  }
  return result;
}

function parseInfixExpression(node: Node, source: string, operators: string[]): Expression {
  const count = node.namedChildCount;
  if (count === 0) {
    throw new MapperError(`parser: empty ${node.type}`);
  }
  if (count === 1) {
    return parseExpression(node.namedChild(0), source);
  }
  let result = parseExpression(node.namedChild(0), source);
  let previous = node.namedChild(0);
  for (let i = 1; i < count; i++) {
    const rightNode = node.namedChild(i);
    const rightExpr = parseExpression(rightNode, source);
    const operator = extractOperatorBetween(previous, rightNode, source, operators);
    if (!operator) {
      throw new MapperError(`parser: could not determine operator between operands in ${node.type}`);
    }
    result = AST.binaryExpression(operator, result, rightExpr);
    previous = rightNode;
  }
  return result;
}

function extractOperatorBetween(left: Node | null, right: Node | null, source: string, allowed: string[]): string {
  if (!left || !right) return "";
  const start = left.endIndex;
  const end = right.startIndex;
  if (start < 0 || end < start || end > source.length) {
    return "";
  }
  const segment = source.slice(start, end).trim();
  if (segment === "") {
    return "";
  }
  for (const op of allowed) {
    if (segment === op) return op;
  }
  for (const op of allowed) {
    if (segment.includes(op)) return op;
  }
  return "";
}

const ASSIGNMENT_OPERATORS = new Set([
  ":=",
  "=",
  "+=",
  "-=",
  "*=",
  "/=",
  "%=",
  "&=",
  "|=",
  "\\xor=",
  "<<=",
  ">>=",
]);

function parseAssignmentExpression(node: Node, source: string): Expression {
  const operatorNode = node.childForFieldName("operator");
  if (!operatorNode) {
    const child = firstNamedChild(node);
    if (!child) {
      throw new MapperError("parser: empty assignment expression");
    }
    return parseExpression(child, source);
  }
  const leftNode = node.childForFieldName("left");
  const rightNode = node.childForFieldName("right");
  if (!leftNode || !rightNode) {
    throw new MapperError("parser: malformed assignment expression");
  }
  const left = parseAssignmentTarget(leftNode, source);
  const right = parseExpression(rightNode, source);
  const operatorText = sliceText(operatorNode, source).trim();
  if (!ASSIGNMENT_OPERATORS.has(operatorText)) {
    throw new MapperError(`parser: unsupported assignment operator ${operatorText}`);
  }
  return AST.assignmentExpression(operatorText as AssignmentExpression["operator"], left, right);
}

function parseAssignmentTarget(node: Node, source: string): AssignmentExpression["left"] {
  switch (node.type) {
    case "assignment_target": {
      const child = firstNamedChild(node);
      if (!child) {
        throw new MapperError("parser: empty assignment target");
      }
      return parseAssignmentTarget(child, source);
    }
    case "pattern":
    case "pattern_base":
    case "typed_pattern":
    case "struct_pattern":
    case "array_pattern":
      return parsePattern(node, source);
    default: {
      const expr = parseExpression(node, source);
      if (expr.type === "MemberAccessExpression" || expr.type === "IndexExpression") {
        return expr;
      }
      if (
        expr.type === "Identifier" ||
        expr.type === "StructPattern" ||
        expr.type === "ArrayPattern" ||
        expr.type === "TypedPattern" ||
        expr.type === "WildcardPattern" ||
        expr.type === "LiteralPattern"
      ) {
        return expr as Pattern;
      }
      throw new MapperError(`parser: expression cannot be used as assignment target: ${expr.type}`);
    }
  }
}

function parseUnaryExpression(node: Node, source: string): Expression {
  const operandNode = firstNamedChild(node);
  if (!operandNode) {
    throw new MapperError("parser: unary expression missing operand");
  }
  if (node.startIndex === operandNode.startIndex) {
    return parseExpression(operandNode, source);
  }
  const operatorText = source.slice(node.startIndex, operandNode.startIndex).trim();
  if (operatorText === "") {
    return parseExpression(operandNode, source);
  }
  const operand = parseExpression(operandNode, source);
  if (operatorText === "-" || operatorText === "!" || operatorText === "~") {
    return AST.unaryExpression(operatorText as "-" | "!" | "~", operand);
  }
  throw new MapperError(`parser: unsupported unary operator ${operatorText}`);
}

function parseRangeExpression(node: Node, source: string): Expression {
  const operatorNode = node.childForFieldName("operator");
  if (!operatorNode || node.namedChildCount < 2) {
    const child = firstNamedChild(node);
    if (child) {
      return parseExpression(child, source);
    }
    throw new MapperError("parser: malformed range expression");
  }
  const startExpr = parseExpression(node.namedChild(0), source);
  const endExpr = parseExpression(node.namedChild(1), source);
  const operatorText = sliceText(operatorNode, source).trim();
  if (operatorText !== ".." && operatorText !== "...") {
    throw new MapperError(`parser: unsupported range operator ${operatorText}`);
  }
  return AST.rangeExpression(startExpr, endExpr, operatorText === "...");
}

function parseLambdaExpression(node: Node, source: string): LambdaExpression {
  if (node.type !== "lambda_expression") {
    throw new MapperError("parser: expected lambda expression");
  }

  const params: FunctionParameter[] = [];
  const paramsNode = node.childForFieldName("parameters");
  if (paramsNode) {
    for (let i = 0; i < paramsNode.namedChildCount; i++) {
      const paramNode = paramsNode.namedChild(i);
      if (!paramNode || paramNode.type !== "lambda_parameter") continue;
      params.push(parseLambdaParameter(paramNode, source));
    }
  }

  const returnNode = node.childForFieldName("return_type");
  const returnType = returnNode ? parseReturnType(returnNode, source) : undefined;

  const bodyNode = node.childForFieldName("body");
  if (!bodyNode) {
    throw new MapperError("parser: lambda missing body");
  }
  const bodyExpr = parseExpression(bodyNode, source);

  return AST.lambdaExpression(params, bodyExpr, returnType, undefined, undefined, false);
}

function parseLambdaParameter(node: Node, source: string): FunctionParameter {
  const nameNode = node.childForFieldName("name");
  if (!nameNode) {
    throw new MapperError("parser: lambda parameter missing name");
  }
  const id = parseIdentifier(nameNode, source);
  return AST.functionParameter(id);
}

function parseIfExpression(node: Node, source: string): Expression {
  const conditionNode = node.childForFieldName("condition");
  if (!conditionNode) {
    throw new MapperError("parser: if expression missing condition");
  }
  const condition = parseExpression(conditionNode, source);

  const bodyNode = node.childForFieldName("consequence");
  if (!bodyNode) {
    throw new MapperError("parser: if expression missing body");
  }
  const body = parseBlock(bodyNode, source);

  const clauses: OrClause[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (child.type === "or_clause") {
      clauses.push(parseOrClause(child, source));
    }
  }

  const elseNode = findElseBlock(node, bodyNode);
  if (elseNode) {
    const elseBody = parseBlock(elseNode, source);
    clauses.push(AST.orClause(elseBody, undefined));
  }

  return AST.ifExpression(condition, body, clauses);
}

function parseOrClause(node: Node, source: string): OrClause {
  const bodyNode = node.childForFieldName("consequence");
  if (!bodyNode) {
    throw new MapperError("parser: or clause missing body");
  }
  const body = parseBlock(bodyNode, source);

  const conditionNode = node.childForFieldName("condition");
  let condition: Expression | undefined;
  if (conditionNode) {
    condition = parseExpression(conditionNode, source);
  }

  return AST.orClause(body, condition);
}

function findElseBlock(ifNode: Node, consequence: Node): Node | null {
  const consequenceStart = consequence.startIndex;
  const consequenceEnd = consequence.endIndex;
  for (let i = 0; i < ifNode.namedChildCount; i++) {
    const child = ifNode.namedChild(i);
    if (child.type !== "block") continue;
    if (child.startIndex === consequenceStart && child.endIndex === consequenceEnd) {
      continue;
    }
    return child;
  }
  return null;
}

function parseMatchExpression(node: Node, source: string): Expression {
  const subjectNode = node.childForFieldName("subject");
  if (!subjectNode) {
    throw new MapperError("parser: match expression missing subject");
  }
  const subject = parseExpression(subjectNode, source);

  const clauses: MatchClause[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (child.type === "match_clause") {
      clauses.push(parseMatchClause(child, source));
    }
  }

  if (clauses.length === 0) {
    throw new MapperError("parser: match expression requires at least one clause");
  }

  return AST.matchExpression(subject, clauses);
}

function parseMatchClause(node: Node, source: string): MatchClause {
  const patternNode = node.childForFieldName("pattern");
  if (!patternNode) {
    throw new MapperError("parser: match clause missing pattern");
  }
  const pattern = parsePattern(patternNode, source);

  let guardExpr: Expression | undefined;
  const guardNode = node.childForFieldName("guard");
  if (guardNode) {
    const guardChild = firstNamedChild(guardNode);
    if (!guardChild) {
      throw new MapperError("parser: match guard missing expression");
    }
    guardExpr = parseExpression(guardChild, source);
  }

  const bodyNode = node.childForFieldName("body");
  if (!bodyNode) {
    throw new MapperError("parser: match clause missing body");
  }

  let body: Expression;
  if (bodyNode.type === "block") {
    body = parseBlock(bodyNode, source);
  } else {
    body = parseExpression(bodyNode, source);
  }

  return AST.matchClause(pattern, body, guardExpr);
}

function parseHandlingExpression(node: Node, source: string): Expression {
  if (node.type !== "handling_expression") {
    throw new MapperError("parser: expected handling_expression node");
  }
  if (node.namedChildCount === 0) {
    throw new MapperError("parser: handling expression missing base expression");
  }
  const baseExpr = parseExpression(node.namedChild(0), source);

  let current: Expression | undefined = baseExpr;
  let assignment: AssignmentExpression | undefined;
  if (baseExpr.type === "AssignmentExpression") {
    assignment = baseExpr;
    current = baseExpr.right;
  }

  for (let i = 1; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || child.type !== "else_clause") continue;
    const handlerNode = child.childForFieldName("handler");
    if (!handlerNode) {
      throw new MapperError("parser: else clause missing handler block");
    }
    const { block, binding } = parseHandlingBlock(handlerNode, source);
    current = AST.orElseExpression(current, block, binding);
  }

  if (assignment) {
    if (!current) {
      throw new MapperError("parser: or-else assignment missing right-hand expression");
    }
    assignment.right = current;
    return assignment;
  }

  if (!current) {
    throw new MapperError("parser: handling expression missing result");
  }

  return current;
}

function parseHandlingBlock(node: Node, source: string): { block: BlockExpression; binding?: Identifier } {
  if (node.type !== "handling_block") {
    throw new MapperError("parser: expected handling_block node");
  }

  let binding: Identifier | undefined;
  const bindingNode = node.childForFieldName("binding");
  if (bindingNode) {
    binding = parseIdentifier(bindingNode, source);
  }

  const statements: Statement[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || !child.isNamed) continue;
    if (node.fieldNameForChild(i) === "binding") continue;
    const stmt = parseStatement(child, source);
    if (stmt) {
      statements.push(stmt);
    }
  }

  return { block: AST.blockExpression(statements), binding };
}

function parseRescueExpression(node: Node, source: string): Expression {
  if (node.type !== "rescue_expression") {
    throw new MapperError("parser: expected rescue_expression node");
  }
  if (node.namedChildCount === 0) {
    throw new MapperError("parser: rescue expression missing monitored expression");
  }

  let monitoredNode: Node | null = null;
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || child.type === "rescue_block") continue;
    monitoredNode = child;
    break;
  }

  if (!monitoredNode) {
    throw new MapperError("parser: rescue expression missing monitored expression");
  }

  const expr = parseExpression(monitoredNode, source);
  const rescueNode = node.childForFieldName("rescue");
  if (!rescueNode) {
    throw new MapperError("parser: rescue expression missing rescue block");
  }

  const clauses = parseRescueBlock(rescueNode, source);

  if (expr.type === "AssignmentExpression") {
    if (!expr.right) {
      throw new MapperError("parser: rescue assignment missing right-hand expression");
    }
    const rescueExpr = AST.rescueExpression(expr.right, clauses);
    expr.right = rescueExpr;
    return expr;
  }

  return AST.rescueExpression(expr, clauses);
}

function parseRescueBlock(node: Node, source: string): MatchClause[] {
  if (node.type !== "rescue_block") {
    throw new MapperError("parser: expected rescue_block node");
  }
  const clauses: MatchClause[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || child.type !== "match_clause") continue;
    clauses.push(parseMatchClause(child, source));
  }
  if (clauses.length === 0) {
    throw new MapperError("parser: rescue block requires at least one clause");
  }
  return clauses;
}

function parseEnsureExpression(node: Node, source: string): Expression {
  if (node.type !== "ensure_expression") {
    throw new MapperError("parser: expected ensure_expression node");
  }

  let tryNode: Node | null = null;
  const ensureNode = node.childForFieldName("ensure");
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child || child === ensureNode) continue;
    tryNode = child;
    break;
  }

  if (!tryNode) {
    throw new MapperError("parser: ensure expression missing try expression");
  }

  const tryExpr = parseExpression(tryNode, source);
  if (!ensureNode) {
    throw new MapperError("parser: ensure expression missing ensure block");
  }
  const ensureBlock = parseBlock(ensureNode, source);

  if (tryExpr.type === "AssignmentExpression") {
    if (!tryExpr.right) {
      throw new MapperError("parser: ensure assignment missing right-hand expression");
    }
    const ensureExpr = AST.ensureExpression(tryExpr.right, ensureBlock);
    tryExpr.right = ensureExpr;
    return tryExpr;
  }

  return AST.ensureExpression(tryExpr, ensureBlock);
}

function parseBreakpointExpression(node: Node, source: string): Expression {
  if (node.type !== "breakpoint_expression") {
    throw new MapperError("parser: expected breakpoint_expression node");
  }

  let labelNode = node.childForFieldName("label");
  if (!labelNode) {
    labelNode = fallbackBreakpointLabel(node);
  }
  if (!labelNode) {
    throw new MapperError("parser: breakpoint expression missing label");
  }
  let label: Identifier;
  if (labelNode.type === "label") {
    label = parseLabel(labelNode, source);
  } else {
    label = parseIdentifier(labelNode, source);
  }

  let bodyNode: Node | null = null;
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (child && child.type === "block") {
      bodyNode = child;
      break;
    }
  }
  if (!bodyNode) {
    throw new MapperError("parser: breakpoint expression missing body");
  }

  const body = parseBlock(bodyNode, source);
  return AST.breakpointExpression(label, body);
}

function fallbackBreakpointLabel(node: Node): Node | null {
  for (let i = 0; i < node.childCount; i++) {
    const child = node.child(i);
    if (!child || isIgnorableNode(child)) continue;
    if (child.type === "identifier" || child.type === "label") {
      return child;
    }
    if (child.type === "ERROR" && child.childCount === 1) {
      const grand = child.child(0);
      if (grand && (grand.type === "identifier" || grand.type === "label")) {
        return grand;
      }
    }
  }
  return null;
}

function parseFunctionDefinition(node: Node, source: string): FunctionDefinition {
  if (node.type !== "function_definition") {
    throw new MapperError("parser: expected function_definition node");
  }
  const core = parseFunctionCore(node, source);
  return AST.functionDefinition(
    core.name,
    core.params,
    core.body,
    core.returnType,
    core.generics,
    core.whereClause,
    core.isMethodShorthand,
    core.isPrivate,
  );
}

function parseFunctionCore(node: Node, source: string): {
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
  const params = parseParameterList(node.childForFieldName("parameters"), source);
  const bodyNode = node.childForFieldName("body");
  const body = parseBlock(bodyNode, source);

  let isPrivate = false;
  for (let i = 0; i < node.childCount; i++) {
    const child = node.child(i);
    if (!child || isIgnorableNode(child)) continue;
    if (child.type === "private") {
      isPrivate = true;
      break;
    }
  }

  const returnType = parseReturnType(node.childForFieldName("return_type"), source);
  const generics = parseTypeParameters(node.childForFieldName("type_parameters"), source);
  const whereClause = parseWhereClause(node.childForFieldName("where_clause"), source);
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

function parseParameterList(node: Node | null | undefined, source: string): FunctionParameter[] {
  if (!node || node.namedChildCount === 0) {
    return [];
  }
  const params: FunctionParameter[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const paramNode = node.namedChild(i);
    if (!paramNode || isIgnorableNode(paramNode)) continue;
    params.push(parseParameter(paramNode, source));
  }
  return params;
}

function parseParameter(node: Node, source: string): FunctionParameter {
  if (node.type !== "parameter") {
    throw new MapperError("parser: expected parameter node");
  }
  const patternNode = node.childForFieldName("pattern");
  const pattern = parsePattern(patternNode, source);
  let namePattern: Pattern = pattern;
  let paramType: TypeExpression | undefined;
  if (pattern.type === "TypedPattern") {
    paramType = pattern.typeAnnotation;
    namePattern = pattern.pattern;
  }
  const typeNode = node.childForFieldName("type");
  if (!paramType && typeNode) {
    paramType = parseTypeExpression(typeNode, source) ?? undefined;
  }
  return AST.functionParameter(namePattern, paramType);
}

function parseStructDefinition(node: Node, source: string): StructDefinition {
  if (node.type !== "struct_definition") {
    throw new MapperError("parser: expected struct_definition node");
  }

  const nameNode = node.childForFieldName("name");
  const id = parseIdentifier(nameNode, source);

  const generics = parseTypeParameters(node.childForFieldName("type_parameters"), source);
  const whereClause = parseWhereClause(node.childForFieldName("where_clause"), source);
  const isPrivate = hasLeadingPrivate(node);

  let kind: StructDefinition["kind"] = "singleton";
  const fields: StructFieldDefinition[] = [];

  const recordNode = node.childForFieldName("record");
  const tupleNode = node.childForFieldName("tuple");

  if (recordNode) {
    kind = "named";
    for (let i = 0; i < recordNode.namedChildCount; i++) {
      const fieldNode = recordNode.namedChild(i);
      if (!fieldNode || isIgnorableNode(fieldNode) || fieldNode.type !== "struct_field") continue;
      fields.push(parseStructFieldDefinition(fieldNode, source));
    }
  } else if (tupleNode) {
    kind = "positional";
    for (let i = 0; i < tupleNode.namedChildCount; i++) {
      const child = tupleNode.namedChild(i);
      if (!child || !child.isNamed || isIgnorableNode(child)) continue;
      const fieldType = parseTypeExpression(child, source);
      if (!fieldType) {
        throw new MapperError("parser: unsupported tuple field type");
      }
      fields.push(AST.structFieldDefinition(fieldType));
    }
  }

  return AST.structDefinition(id, fields, kind, generics, whereClause, isPrivate ? true : undefined);
}

function parseStructFieldDefinition(node: Node, source: string): StructFieldDefinition {
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
      fieldType = parseTypeExpression(child, source);
    }
  }

  if (!fieldType) {
    throw new MapperError("parser: struct field missing type");
  }

  return AST.structFieldDefinition(fieldType, name);
}

function parseMethodsDefinition(node: Node, source: string): MethodsDefinition {
  if (node.type !== "methods_definition") {
    throw new MapperError("parser: expected methods_definition node");
  }

  const generics = parseTypeParameters(node.childForFieldName("type_parameters"), source);
  const whereClause = parseWhereClause(node.childForFieldName("where_clause"), source);
  const targetType = parseTypeExpression(node.childForFieldName("target"), source);
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
    if (
      sameNode(child, targetNode) ||
      sameNode(child, typeParamsNode) ||
      sameNode(child, whereNode)
    ) {
      continue;
    }
    if (child.type === "function_definition") {
      const fn = parseFunctionDefinition(child, source);
      definitions.push(fn);
    } else if (child.type === "method_member") {
      for (let j = 0; j < child.namedChildCount; j++) {
        const member = child.namedChild(j);
        if (!member || member.type !== "function_definition") continue;
        const fn = parseFunctionDefinition(member, source);
        definitions.push(fn);
      }
    }
  }

  return AST.methodsDefinition(targetType, definitions, generics, whereClause);
}

function parseImplementationDefinitionNode(node: Node, source: string): ImplementationDefinition {
  if (node.type !== "implementation_definition") {
    throw new MapperError("parser: expected implementation_definition node");
  }

  const interfaceNode = node.childForFieldName("interface");
  if (!interfaceNode) {
    throw new MapperError("parser: implementation missing interface");
  }

  const parts = parseQualifiedIdentifier(interfaceNode, source);
  let interfaceName = parts[parts.length - 1];
  if (parts.length > 1) {
    interfaceName = AST.identifier(identifiersToStrings(parts).join("."));
  }

  const interfaceArgs = parseInterfaceArguments(node.childForFieldName("interface_args"), source);
  const targetType = parseTypeExpression(node.childForFieldName("target"), source);
  if (!targetType) {
    throw new MapperError("parser: implementation missing target type");
  }
  const generics = parseTypeParameters(node.childForFieldName("type_parameters"), source);
  const whereClause = parseWhereClause(node.childForFieldName("where_clause"), source);
  const isPrivate = hasLeadingPrivate(node);

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
      (fieldName === "interface" || fieldName === "interface_args" || fieldName === "target" || fieldName === "type_parameters" || fieldName === "where_clause") &&
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
      const fn = parseFunctionDefinition(child, source);
      definitions.push(fn);
    } else if (child.type === "method_member") {
      for (let j = 0; j < child.namedChildCount; j++) {
        const member = child.namedChild(j);
        if (!member || member.type !== "function_definition") continue;
        const fn = parseFunctionDefinition(member, source);
        definitions.push(fn);
      }
    }
  }

  return AST.implementationDefinition(
    interfaceName,
    targetType,
    definitions,
    undefined,
    generics,
    interfaceArgs ?? undefined,
    whereClause,
  );
}

function parseImplementationDefinition(node: Node, source: string): ImplementationDefinition {
  return parseImplementationDefinitionNode(node, source);
}

function parseNamedImplementationDefinition(node: Node, source: string): ImplementationDefinition {
  if (node.type !== "named_implementation_definition") {
    throw new MapperError("parser: expected named implementation node");
  }
  const nameNode = node.childForFieldName("name");
  const implNode = node.childForFieldName("implementation");
  if (!implNode) {
    throw new MapperError("parser: named implementation missing implementation body");
  }
  const impl = parseImplementationDefinitionNode(implNode, source);
  if (nameNode) {
    impl.implName = parseIdentifier(nameNode, source);
  }
  return impl;
}

function parseInterfaceArguments(node: Node | null | undefined, source: string): TypeExpression[] | undefined {
  if (!node) return undefined;
  const args: TypeExpression[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child) continue;
    const expr = parseTypeExpression(child, source);
    if (!expr) {
      throw new MapperError(`parser: unsupported interface argument kind ${child.type}`);
    }
    args.push(expr);
  }
  return args.length ? args : undefined;
}

function parseUnionDefinition(node: Node, source: string): UnionDefinition {
  if (node.type !== "union_definition") {
    throw new MapperError("parser: expected union_definition node");
  }

  const nameNode = node.childForFieldName("name");
  const name = parseIdentifier(nameNode, source);
  const typeParamsNode = node.childForFieldName("type_parameters");
  const typeParams = parseTypeParameters(typeParamsNode, source);

  const variants: TypeExpression[] = [];
  for (let i = 0; i < node.namedChildCount; i++) {
    const child = node.namedChild(i);
    if (!child) continue;
    if (sameNode(child, nameNode) || sameNode(child, typeParamsNode)) continue;
    const variant = parseTypeExpression(child, source);
    if (!variant) {
      throw new MapperError("parser: invalid union variant");
    }
    variants.push(variant);
  }

  if (variants.length === 0) {
    throw new MapperError("parser: union definition requires variants");
  }

  return AST.unionDefinition(name, variants, typeParams, undefined, hasLeadingPrivate(node));
}

function parseInterfaceDefinition(node: Node, source: string): InterfaceDefinition {
  if (node.type !== "interface_definition") {
    throw new MapperError("parser: expected interface_definition node");
  }

  const nameNode = node.childForFieldName("name");
  const name = parseIdentifier(nameNode, source);
  const typeParamsNode = node.childForFieldName("type_parameters");
  const typeParams = parseTypeParameters(typeParamsNode, source);

  let selfType: TypeExpression | undefined;
  const selfNode = node.childForFieldName("self_type");
  if (selfNode) {
    selfType = parseTypeExpression(selfNode, source) ?? undefined;
  }

  const whereNode = node.childForFieldName("where_clause");
  const whereClause = parseWhereClause(whereNode, source);

  const compositeNode = node.childForFieldName("composite");

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
    const signature = parseFunctionSignature(sigNode, source);
    const defaultBody = child.childForFieldName("default_body");
    if (defaultBody) {
      signature.defaultImpl = parseBlock(defaultBody, source);
    }
    signatures.push(signature);
  }

  let baseInterfaces: TypeExpression[] | undefined;
  if (compositeNode) {
    const bounds = parseTypeBoundList(compositeNode, source);
    if (bounds && bounds.length > 0) {
      baseInterfaces = bounds;
    }
  }

  return AST.interfaceDefinition(name, signatures, typeParams, selfType, whereClause, baseInterfaces, hasLeadingPrivate(node));
}

function parseFunctionSignature(node: Node, source: string): FunctionSignature {
  if (node.type !== "function_signature") {
    throw new MapperError("parser: expected function_signature node");
  }

  const name = parseIdentifier(node.childForFieldName("name"), source);
  const params = parseParameterList(node.childForFieldName("parameters"), source);
  const returnType = parseReturnType(node.childForFieldName("return_type"), source);
  const generics = parseTypeParameters(node.childForFieldName("type_parameters"), source);
  const whereClause = parseWhereClause(node.childForFieldName("where_clause"), source);

  return AST.functionSignature(name, params, returnType, generics, whereClause, undefined);
}

function parsePreludeStatement(node: Node, source: string): PreludeStatement {
  if (node.type !== "prelude_statement") {
    throw new MapperError("parser: expected prelude_statement node");
  }
  const target = parseHostTarget(node.childForFieldName("target"), source);
  const code = parseHostCodeBlock(node.childForFieldName("body"), source);
  return AST.preludeStatement(target, code);
}

function parseExternFunction(node: Node, source: string): ExternFunctionBody {
  if (node.type !== "extern_function") {
    throw new MapperError("parser: expected extern_function node");
  }
  const target = parseHostTarget(node.childForFieldName("target"), source);
  const signatureNode = node.childForFieldName("signature");
  if (!signatureNode) {
    throw new MapperError("parser: extern function missing signature");
  }
  const signature = parseFunctionSignature(signatureNode, source);
  const body = parseHostCodeBlock(node.childForFieldName("body"), source);

  const fn = AST.functionDefinition(
    signature.name,
    signature.params,
    AST.blockExpression([]),
    signature.returnType,
    signature.genericParams,
    signature.whereClause,
    false,
    false,
  );

  return AST.externFunctionBody(target, fn, body);
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
  // Remove enclosing braces `{` `}`
  const inner = source.slice(start + 1, end - 1).trim();
  return inner;
}




// Placeholder stubs for the remaining parse helpers. These will be fleshed out as the mapper is completed.
function parseDoExpression(node: Node, source: string): Expression {
  const bodyNode = firstNamedChild(node);
  if (!bodyNode) {
    throw new MapperError("parser: do expression missing body");
  }
  return parseBlock(bodyNode, source);
}

function parseProcExpression(node: Node, source: string): Expression {
  const bodyNode = firstNamedChild(node);
  if (!bodyNode) {
    throw new MapperError("parser: proc expression missing body");
  }
  const expr = parseExpression(bodyNode, source);
  if (expr.type !== "FunctionCall" && expr.type !== "BlockExpression") {
    throw new MapperError("parser: proc expression requires function call or block");
  }
  return AST.procExpression(expr as FunctionCall | BlockExpression);
}

function parseSpawnExpression(node: Node, source: string): Expression {
  const bodyNode = firstNamedChild(node);
  if (!bodyNode) {
    throw new MapperError("parser: spawn expression missing body");
  }
  const expr = parseExpression(bodyNode, source);
  if (expr.type !== "FunctionCall" && expr.type !== "BlockExpression") {
    throw new MapperError("parser: spawn expression requires function call or block");
  }
  return AST.spawnExpression(expr as FunctionCall | BlockExpression);
}

// --- Numerous helpers omitted for brevity; the completed mapper will port every parser helper from Go. ---

function isNumericSuffix(value: string): boolean {
  switch (value) {
    case "i8":
    case "i16":
    case "i32":
    case "i64":
    case "i128":
    case "u8":
    case "u16":
    case "u32":
    case "u64":
    case "u128":
    case "f32":
    case "f64":
      return true;
    default:
      return false;
  }
}

function parseLabel(node: Node, source: string): Identifier {
  if (node.type !== "label") {
    throw new MapperError("parser: expected label");
  }
  let content = sliceText(node, source).trim();
  if (content.startsWith("'")) {
    content = content.slice(1);
  }
  if (!content) {
    throw new MapperError("parser: empty label");
  }
  return AST.identifier(content);
}
