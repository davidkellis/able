import path from "node:path";
import { fileURLToPath } from "node:url";

import { Language, Parser } from "web-tree-sitter";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const DEFAULT_LANGUAGE_WASM_PATH = path.resolve(
  __dirname,
  "../parser/tree-sitter-able/tree-sitter-able.wasm",
);

const WRAPPER_TYPES = new Set([
  "expression_statement",
  "low_precedence_pipe_expression",
  "pipe_expression",
  "matchable_expression",
  "pipe_operand_base",
  "range_expression",
  "logical_or_expression",
  "logical_and_expression",
  "bitwise_or_expression",
  "bitwise_xor_expression",
  "bitwise_and_expression",
  "equality_expression",
  "comparison_expression",
  "shift_expression",
  "additive_expression",
  "multiplicative_expression",
  "unary_expression",
  "exponent_expression",
  "primary_expression",
  "assignment_target",
  "pattern_base",
]);

const BINARY_TYPES = new Set([
  "logical_or_expression",
  "logical_and_expression",
  "bitwise_or_expression",
  "bitwise_xor_expression",
  "bitwise_and_expression",
  "equality_expression",
  "comparison_expression",
  "shift_expression",
  "additive_expression",
  "multiplicative_expression",
]);

const ASSIGNMENT_OPERATORS = new Set([
  ":=",
  "=",
  "+=",
  "-=",
  "*=",
  "/=",
  "%=",
  ".&=",
  ".|=",
  ".^=",
  ".<<=",
  ".>>=",
]);

export async function createAbleParser(languageWasmPath = DEFAULT_LANGUAGE_WASM_PATH) {
  await Parser.init();
  const language = await Language.load(languageWasmPath);
  const parser = new Parser();
  parser.setLanguage(language);
  return parser;
}

export function parseSourceToAstModule(parser, source) {
  const tree = parser.parse(source);
  try {
    const root = tree.rootNode;
    if (root.type !== "source_file") {
      throw new Error(`expected source_file root, got ${root.type}`);
    }
    if (root.hasError) {
      throw new Error("tree-sitter reported parse errors");
    }

    const moduleAst = {
      type: "Module",
      imports: [],
      body: [],
    };

    const packageNode = root.childForFieldName("package");
    if (packageNode) {
      moduleAst.package = parsePackageStatement(packageNode);
    }

    for (const stmt of root.namedChildren) {
      if (stmt.type === "package_statement") {
        continue;
      }
      if (stmt.type === "import_statement") {
        const importNode = parseImportStatement(stmt);
        if (importNode.type === "ImportStatement") {
          moduleAst.imports.push(importNode);
        } else {
          moduleAst.body.push(importNode);
        }
        continue;
      }
      moduleAst.body.push(parseStatement(stmt));
    }

    return moduleAst;
  } finally {
    tree.delete();
  }
}

function parseStatement(node) {
  if (node.type === "expression_statement") {
    const expr = node.childForFieldName("expression") ?? node.namedChild(0);
    return parseExpression(expr);
  }
  if (node.type === "import_statement") {
    return parseImportStatement(node);
  }
  throw unsupported(node, "statement");
}

function parseExpression(node) {
  if (!node) {
    throw new Error("missing expression node");
  }

  if (node.type === "postfix_expression") {
    return parsePostfixExpression(node);
  }

  if (node.type === "assignment_expression") {
    const left = node.childForFieldName("left");
    const opNode = node.childForFieldName("operator");
    const right = node.childForFieldName("right");
    if (left && opNode && right) {
      const operator = opNode.text.trim();
      if (!ASSIGNMENT_OPERATORS.has(operator)) {
        throw new Error(`unsupported assignment operator ${JSON.stringify(operator)}`);
      }
      return {
        type: "AssignmentExpression",
        operator,
        left: parseAssignmentTarget(left),
        right: parseExpression(right),
      };
    }
    if (node.namedChildCount === 1) {
      return parseExpression(node.namedChild(0));
    }
  }

  if (BINARY_TYPES.has(node.type)) {
    if (node.namedChildCount === 1) {
      return parseExpression(node.namedChild(0));
    }
    if (node.namedChildCount === 2) {
      return {
        type: "BinaryExpression",
        operator: extractOperator(node),
        left: parseExpression(node.namedChild(0)),
        right: parseExpression(node.namedChild(1)),
      };
    }
  }

  if (WRAPPER_TYPES.has(node.type) && node.namedChildCount === 1) {
    return parseExpression(node.namedChild(0));
  }

  switch (node.type) {
    case "identifier":
      return parseIdentifier(node);
    case "number_literal":
      return parseNumberLiteral(node.text);
    case "boolean_literal":
      return {
        type: "BooleanLiteral",
        value: node.text.trim() === "true",
      };
    case "string_literal":
      return {
        type: "StringLiteral",
        value: JSON.parse(node.text),
      };
    case "nil_literal":
      return {
        type: "NilLiteral",
        value: null,
      };
    default:
      throw unsupported(node, "expression");
  }
}

function parsePostfixExpression(node) {
  if (node.namedChildCount === 0) {
    throw unsupported(node, "postfix expression");
  }
  let current = parseExpression(node.namedChild(0));
  for (let i = 1; i < node.childCount; i += 1) {
    const child = node.child(i);
    if (!child || !child.isNamed) {
      continue;
    }
    if (child.type === "call_suffix") {
      const args = child.namedChildren.map(parseExpression);
      current = {
        type: "FunctionCall",
        callee: current,
        arguments: args,
        isTrailingLambda: false,
      };
      continue;
    }
    if (child.type === "member_access") {
      const member = child.namedChild(0);
      if (!member || member.type !== "identifier") {
        throw unsupported(child, "member access");
      }
      current = {
        type: "MemberAccessExpression",
        object: current,
        member: parseIdentifier(member),
      };
      continue;
    }
    throw unsupported(child, "postfix operation");
  }
  return current;
}

function parseAssignmentTarget(node) {
  const unwrapped = unwrap(node);
  if (unwrapped.type === "identifier") {
    return parseIdentifier(unwrapped);
  }
  if (unwrapped.type === "postfix_expression") {
    return parsePostfixExpression(unwrapped);
  }
  throw unsupported(unwrapped, "assignment target");
}

function parsePackageStatement(node) {
  const idents = node.namedChildren.filter((child) => child.type === "identifier");
  if (idents.length === 0) {
    throw unsupported(node, "package statement");
  }
  return {
    type: "PackageStatement",
    namePath: idents.map(parseIdentifier),
  };
}

function parseImportStatement(node) {
  const kindNode = node.childForFieldName("kind");
  const pathNode = node.childForFieldName("path");
  const aliasNode = node.childForFieldName("alias");
  const clauseNode = node.childForFieldName("clause");

  if (!kindNode || !pathNode) {
    throw unsupported(node, "import statement");
  }

  const packagePath = pathNode.namedChildren.map(parseIdentifier);
  const kind = kindNode.text.trim();
  const isDynamic = kind === "dynimport";

  let isWildcard = false;
  let selectors = undefined;
  if (clauseNode) {
    const wildcard = clauseNode.namedChildren.find((child) => child.type === "import_wildcard_clause");
    if (wildcard) {
      isWildcard = true;
    } else {
      selectors = clauseNode.namedChildren
        .filter((child) => child.type === "import_selector")
        .map(parseImportSelector);
      if (selectors.length === 0) {
        selectors = undefined;
      }
    }
  }

  const out = {
    type: isDynamic ? "DynImportStatement" : "ImportStatement",
    packagePath,
    isWildcard,
  };
  if (aliasNode) {
    out.alias = parseIdentifier(aliasNode);
  }
  if (selectors) {
    out.selectors = selectors;
  }
  return out;
}

function parseImportSelector(node) {
  const idents = node.namedChildren.filter((child) => child.type === "identifier");
  if (idents.length === 0) {
    throw unsupported(node, "import selector");
  }
  const out = {
    type: "ImportSelector",
    name: parseIdentifier(idents[0]),
  };
  if (idents.length > 1) {
    out.alias = parseIdentifier(idents[1]);
  }
  return out;
}

function parseIdentifier(node) {
  if (node.type !== "identifier") {
    throw unsupported(node, "identifier");
  }
  return {
    type: "Identifier",
    name: node.text,
  };
}

function parseNumberLiteral(raw) {
  const match = raw.match(/^(.*?)(?:_((?:i|u)(?:8|16|32|64|128)|f32|f64))?$/);
  if (!match) {
    throw new Error(`invalid numeric literal ${JSON.stringify(raw)}`);
  }
  const [, bodyRaw, suffix] = match;
  const body = bodyRaw.replaceAll("_", "");

  if (suffix?.startsWith("f") || body.includes(".") || /[eE]/.test(body)) {
    const value = Number.parseFloat(body);
    if (!Number.isFinite(value)) {
      throw new Error(`unsupported float literal ${JSON.stringify(raw)}`);
    }
    const out = {
      type: "FloatLiteral",
      value,
    };
    if (suffix) {
      out.floatType = suffix;
    }
    return out;
  }

  const value = parseIntegerBody(body);
  const out = {
    type: "IntegerLiteral",
    value,
  };
  if (suffix) {
    out.integerType = suffix;
  }
  return out;
}

function parseIntegerBody(body) {
  if (body.length === 0) {
    throw new Error("empty integer literal");
  }

  let sign = 1;
  let digits = body;
  if (digits.startsWith("-")) {
    sign = -1;
    digits = digits.slice(1);
  } else if (digits.startsWith("+")) {
    digits = digits.slice(1);
  }

  let parsed;
  if (/^0[bB][01]+$/.test(digits)) {
    parsed = Number.parseInt(digits.slice(2), 2);
  } else if (/^0[oO][0-7]+$/.test(digits)) {
    parsed = Number.parseInt(digits.slice(2), 8);
  } else if (/^0[xX][0-9a-fA-F]+$/.test(digits)) {
    parsed = Number.parseInt(digits.slice(2), 16);
  } else {
    parsed = Number.parseInt(digits, 10);
  }

  if (!Number.isFinite(parsed)) {
    throw new Error(`unsupported integer literal ${JSON.stringify(body)}`);
  }
  return sign * parsed;
}

function unwrap(node) {
  let current = node;
  while (WRAPPER_TYPES.has(current.type) && current.namedChildCount === 1) {
    current = current.namedChild(0);
  }
  return current;
}

function extractOperator(node) {
  for (let i = 0; i < node.childCount; i += 1) {
    const child = node.child(i);
    if (!child || child.isNamed) {
      continue;
    }
    const text = child.text.trim();
    if (text.length > 0) {
      return text;
    }
  }
  throw unsupported(node, "binary operator");
}

function unsupported(node, context) {
  const snippet = node?.text ? ` (${JSON.stringify(node.text.slice(0, 80))})` : "";
  return new Error(`unsupported ${context}: ${node?.type ?? "<nil>"}${snippet}`);
}
