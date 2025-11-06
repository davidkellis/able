import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { AST } from "../index";
import { fixtures } from "./export-fixtures/fixtures";
import type { Fixture } from "./export-fixtures/types";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FIXTURE_ROOT = path.resolve(__dirname, "../../fixtures/ast");



async function main() {
  for (const fixture of fixtures) {
    await writeFixture(fixture);
  }
  console.log(`Wrote ${fixtures.length} fixture(s) to ${FIXTURE_ROOT}`);
}

async function writeFixture(fixture: Fixture) {
  const outDir = path.join(FIXTURE_ROOT, fixture.name);
  await fs.mkdir(outDir, { recursive: true });

  normalizeModule(fixture.module);

  if (fixture.setupModules) {
    for (const [fileName, module] of Object.entries(fixture.setupModules)) {
      normalizeModule(module);
      const filePath = path.join(outDir, fileName);
      await fs.writeFile(filePath, stringify(module), "utf8");
    }
  }

  const modulePath = path.join(outDir, "module.json");
  await fs.writeFile(modulePath, stringify(fixture.module), "utf8");

  const sourcePath = path.join(outDir, "source.able");
  const source = moduleToSource(fixture.module).trimEnd();
  if (!source.trim()) {
    throw new Error(`export-fixtures: generated empty source for fixture ${fixture.name}`);
  }
  await fs.writeFile(sourcePath, source.endsWith("\n") ? source : `${source}\n`, "utf8");

  if (fixture.manifest) {
    const manifestPath = path.join(outDir, "manifest.json");
    const entry = fixture.manifest.entry ?? "module.json";
    const setup = fixture.manifest.setup ?? (fixture.setupModules ? Object.keys(fixture.setupModules) : undefined);
    const manifest = { ...fixture.manifest, entry, ...(setup ? { setup } : {}) };
    await fs.writeFile(manifestPath, JSON.stringify(manifest, null, 2), "utf8");
  }
}

function stringify(value: unknown): string {
  return JSON.stringify(
    value,
    (_key, val) => (typeof val === "bigint" ? val.toString() : val),
    2,
  );
}

function normalizeModule(module: AST.Module): void {
  // no-op; method shorthand metadata must be set explicitly in fixtures.
}

const INDENT = "  ";

function moduleToSource(module: AST.Module): string {
  const lines: string[] = [];
  if (module.package) {
    lines.push(`package ${module.package.namePath.map(printIdentifier).join(".")}`);
    lines.push("");
  }
  if (module.imports && module.imports.length > 0) {
    for (const imp of module.imports) {
      lines.push(printImport(imp));
    }
    lines.push("");
  }
  for (const stmt of module.body) {
    lines.push(printStatement(stmt, 0));
  }
  return lines
    .map((line) => line.replace(/\s+$/g, ""))
    .join("\n")
    .replace(/\n{3,}/g, "\n\n")
    .trimEnd();
}

function printImport(imp: AST.ImportStatement): string {
  const path = imp.packagePath.map(printIdentifier).join(".");
  if (imp.isWildcard) {
    return `import ${path}.*`;
  }
  if (imp.selectors && imp.selectors.length > 0) {
    const selectors = imp.selectors
      .map((sel) => (sel.alias ? `${printIdentifier(sel.name)} as ${printIdentifier(sel.alias)}` : printIdentifier(sel.name)))
      .join(", ");
    return `import ${path}.{${selectors}}`;
  }
  if (imp.alias) {
    return `import ${path} as ${printIdentifier(imp.alias)}`;
  }
  return `import ${path}`;
}

function printDynImport(imp: AST.DynImportStatement, level: number): string {
  const path = imp.packagePath.map(printIdentifier).join(".");
  if (imp.isWildcard) {
    return `${indent(level)}dynimport ${path}.*`;
  }
  if (imp.selectors && imp.selectors.length > 0) {
    const selectors = imp.selectors
      .map((sel) => (sel.alias ? `${printIdentifier(sel.name)} as ${printIdentifier(sel.alias)}` : printIdentifier(sel.name)))
      .join(", ");
    return `${indent(level)}dynimport ${path}.{${selectors}}`;
  }
  if (imp.alias) {
    return `${indent(level)}dynimport ${path} as ${printIdentifier(imp.alias)}`;
  }
  return `${indent(level)}dynimport ${path}`;
}

function printStatement(stmt: AST.Statement, level: number): string {
  switch (stmt.type) {
    case "FunctionDefinition":
      return printFunctionDefinition(stmt, level);
    case "StructDefinition":
      return printStructDefinition(stmt, level);
    case "UnionDefinition":
      return printUnionDefinition(stmt, level);
    case "InterfaceDefinition":
      return printInterfaceDefinition(stmt, level);
    case "ImplementationDefinition":
      return printImplementationDefinition(stmt, level);
    case "MethodsDefinition":
      return printMethodsDefinition(stmt, level);
    case "ReturnStatement":
      return `${indent(level)}return${stmt.argument ? ` ${printExpression(stmt.argument, level)}` : ""}`;
    case "RaiseStatement":
      return `${indent(level)}raise ${printExpression(stmt.expression, level)}`;
    case "RethrowStatement":
      return `${indent(level)}rethrow`;
    case "BreakStatement": {
      const label = stmt.label ? ` '${printIdentifier(stmt.label)}` : "";
      const value = stmt.value ? ` ${printExpression(stmt.value, level)}` : "";
      return `${indent(level)}break${label}${value}`;
    }
    case "ContinueStatement":
      return `${indent(level)}continue`;
    case "WhileLoop":
      return `${indent(level)}while ${printExpression(stmt.condition, level)} ${printBlock(stmt.body, level)}`;
    case "ForLoop":
      return `${indent(level)}for ${printPattern(stmt.pattern)} in ${printExpression(stmt.iterable, level)} ${printBlock(stmt.body, level)}`;
    case "YieldStatement":
      return `${indent(level)}yield${stmt.argument ? ` ${printExpression(stmt.argument, level)}` : ""}`;
    case "PreludeStatement":
      return `${indent(level)}prelude ${stmt.target} {\n${indent(level + 1)}${stmt.code}\n${indent(level)}}`;
    case "ExternFunctionBody":
      return printExternFunction(stmt, level);
    case "DynImportStatement":
      return printDynImport(stmt, level);
    default:
      if (isExpression(stmt)) {
        return `${indent(level)}${printExpression(stmt, level)}`;
      }
      return `${indent(level)}/* unsupported ${stmt.type} */`;
  }
}

function printFunctionDefinition(fn: AST.FunctionDefinition, level: number): string {
  let header = `${indent(level)}${fn.isPrivate ? "private " : ""}fn`;
  if (fn.isMethodShorthand) {
    header += " #";
  } else {
    header += " ";
  }
  header += printIdentifier(fn.id);
  if (fn.genericParams && fn.genericParams.length > 0) {
    header += `<${fn.genericParams.map(printGenericParameter).join(", " )}>`;
  }
  header += `(${fn.params.map(printFunctionParameter).join(", ")})`;
  if (fn.returnType) {
    header += ` -> ${printTypeExpression(fn.returnType)}`;
  }
  if (fn.whereClause && fn.whereClause.length > 0) {
    header += ` where ${fn.whereClause.map(printWhereClause).join(", ")}`;
  }
  return `${header} ${printBlock(fn.body, level)}`;
}

function printStructDefinition(def: AST.StructDefinition, level: number): string {
  const header: string[] = [];
  if (def.isPrivate) {
    header.push("private");
  }
  header.push("struct");
  header.push(printIdentifier(def.id));
  if (def.genericParams && def.genericParams.length > 0) {
    header.push(`<${def.genericParams.map(printGenericParameter).join(", ")}>`);
  }
  const prefix = `${indent(level)}${header.join(" ")}`;
  const whereSuffix = def.whereClause && def.whereClause.length > 0 ? ` where ${def.whereClause.map(printWhereClause).join(", ")}` : "";
  if (def.kind === "positional") {
    const types = (def.fields ?? []).map((field) => printTypeExpression(field.fieldType)).join(", ");
    return `${prefix}${whereSuffix} { ${types} }`;
  }
  if (def.kind === "named") {
    const fieldList = def.fields ?? [];
    const fields = fieldList.map((field, index) => {
      const suffix = index === fieldList.length - 1 ? "" : ",";
      return `${indent(level + 1)}${field.isPrivate ? "private " : ""}${printIdentifier(field.name!)}: ${printTypeExpression(field.fieldType)}${suffix}`;
    });
    return `${prefix}${whereSuffix} {\n${fields.join("\n")}\n${indent(level)}}`;
  }
  return `${prefix}${whereSuffix} {}`;
}

function printUnionDefinition(def: AST.UnionDefinition, level: number): string {
  const header: string[] = [];
  if (def.isPrivate) {
    header.push("private");
  }
  header.push("union");
  header.push(printIdentifier(def.id));
  if (def.genericParams && def.genericParams.length > 0) {
    header.push(`<${def.genericParams.map(printGenericParameter).join(", ")}>`);
  }
  const suffix = def.whereClause && def.whereClause.length > 0 ? ` where ${def.whereClause.map(printWhereClause).join(", ")}` : "";
  const variants = def.variants && def.variants.length > 0 ? ` = ${def.variants.map(printTypeExpression).join(" | ")}` : "";
  return `${indent(level)}${header.join(" ")}${suffix}${variants}`;
}

function printInterfaceDefinition(def: AST.InterfaceDefinition, level: number): string {
  const header: string[] = [];
  if (def.isPrivate) {
    header.push("private");
  }
  header.push("interface");
  header.push(printIdentifier(def.id));
  if (def.genericParams && def.genericParams.length > 0) {
    header.push(`<${def.genericParams.map(printGenericParameter).join(", ")}>`);
  }
  if (def.whereClause && def.whereClause.length > 0) {
    header.push(`where ${def.whereClause.map(printWhereClause).join(", ")}`);
  }
  if (def.baseInterfaces && def.baseInterfaces.length > 0) {
    header.push(`= ${def.baseInterfaces.map(printTypeExpression).join(" + ")}`);
  }
  const lines = [`${indent(level)}${header.join(" ")}`];
  if (def.signatures && def.signatures.length > 0) {
    lines.push(`${indent(level)}{`);
    for (const sig of def.signatures) {
      lines.push(`${indent(level + 1)}${printFunctionSignature(sig)}`);
    }
    lines.push(`${indent(level)}}`);
  }
  return lines.join("\n");
}

function printImplementationDefinition(def: AST.ImplementationDefinition, level: number): string {
  const header: string[] = [];
  if (def.isPrivate) {
    header.push("private");
  }
  header.push("impl");
  if (def.genericParams && def.genericParams.length > 0) {
    header.push(`<${def.genericParams.map(printGenericParameter).join(", ")}>`);
  }
  if (def.interfaceName) {
    header.push(printIdentifier(def.interfaceName));
    if (def.interfaceArgs && def.interfaceArgs.length > 0) {
      header.push(def.interfaceArgs.map(printTypeExpression).join(" "));
    }
  }
  if (def.targetType) {
    header.push("for");
    header.push(printTypeExpression(def.targetType));
  }
  if (def.whereClause && def.whereClause.length > 0) {
    header.push(`where ${def.whereClause.map(printWhereClause).join(", ")}`);
  }
  const lines = [`${indent(level)}${header.join(" ")}`];
  if (def.definitions && def.definitions.length > 0) {
    lines.push(`${indent(level)}{`);
    for (const inner of def.definitions) {
      lines.push(printFunctionDefinition(inner, level + 1));
    }
    lines.push(`${indent(level)}}`);
  }
  return lines.join("\n");
}

function printMethodsDefinition(def: AST.MethodsDefinition, level: number): string {
  const header: string[] = [];
  header.push("methods");
  header.push(printTypeExpression(def.targetType));
  if (def.genericParams && def.genericParams.length > 0) {
    header.push(`<${def.genericParams.map(printGenericParameter).join(", ")}>`);
  }
  if (def.whereClause && def.whereClause.length > 0) {
    header.push(`where ${def.whereClause.map(printWhereClause).join(", ")}`);
  }
  const lines = [`${indent(level)}${header.join(" ")} {`];
  if (def.definitions) {
    for (const inner of def.definitions) {
      lines.push(printFunctionDefinition(inner, level + 1));
    }
  }
  lines.push(`${indent(level)}}`);
  return lines.join("\n");
}

function printExternFunction(externFn: AST.ExternFunctionBody, level: number): string {
  const signature = printFunctionDefinition(externFn.signature, level);
  const header = `${indent(level)}extern ${externFn.target} ${signature.trimStart()}`;
  const body = externFn.body.split("\n").map((line) => `${indent(level + 1)}${line}`).join("\n");
  return `${header} {\n${body}\n${indent(level)}}`;
}

function printExpression(expr: AST.Expression, level: number): string {
  switch (expr.type) {
    case "StringLiteral":
      return `"${expr.value.replace(/\\/g, "\\\\").replace(/"/g, '\\"')}"`;
    case "IntegerLiteral":
      return expr.integerType ? `${expr.value.toString()}_${expr.integerType}` : expr.value.toString();
    case "FloatLiteral":
      return expr.floatType ? `${expr.value}_${expr.floatType}` : expr.value.toString();
    case "BooleanLiteral":
      return String(expr.value);
    case "NilLiteral":
      return "nil";
    case "CharLiteral":
      return `'${expr.value}'`;
    case "Identifier":
      return printIdentifier(expr);
    case "ArrayLiteral":
      return `[${expr.elements.map((el) => printExpression(el, level)).join(", ")}]`;
    case "AssignmentExpression":
      if (expr.right.type === "MatchExpression") {
        return `${printAssignmentLeft(expr.left)} ${expr.operator} (${printMatchExpression(expr.right, level)})`;
      }
      return `${printAssignmentLeft(expr.left)} ${expr.operator} ${printExpression(expr.right, level)}`;
    case "BinaryExpression":
      return `${printBinaryOperand(expr.left, expr.operator, "left", level)} ${expr.operator} ${printBinaryOperand(expr.right, expr.operator, "right", level)}`;
    case "UnaryExpression":
      return `${expr.operator}${printExpression(expr.operand, level)}`;
    case "FunctionCall":
      return printFunctionCall(expr, level);
    case "BlockExpression":
      return `do ${printBlock(expr, level)}`;
    case "LambdaExpression":
      return printLambda(expr, level);
    case "MemberAccessExpression":
      return `${printExpression(expr.object, level)}.${printMember(expr.member)}`;
    case "ImplicitMemberExpression":
      return `.${printIdentifier(expr.member)}`;
    case "IndexExpression":
      return `${printExpression(expr.object, level)}[${printExpression(expr.index, level)}]`;
    case "RangeExpression":
      return `${printExpression(expr.start, level)} ${expr.inclusive ? "..." : ".."} ${printExpression(expr.end, level)}`;
    case "ProcExpression":
      return expr.expression.type === "BlockExpression"
        ? `proc ${printBlock(expr.expression, level)}`
        : `proc ${printExpression(expr.expression, level)}`;
    case "SpawnExpression":
      return expr.expression.type === "BlockExpression"
        ? `spawn ${printBlock(expr.expression, level)}`
        : `spawn ${printExpression(expr.expression, level)}`;
    case "StructLiteral":
      return printStructLiteral(expr, level);
    case "IfExpression":
      return printIfExpression(expr, level);
    case "MatchExpression":
      return printMatchExpression(expr, level);
    case "PropagationExpression":
      return `${printExpression(expr.expression, level)}!`;
    case "OrElseExpression":
      return `${printExpression(expr.expression, level)} else ${printHandlingBlock(expr.handler, expr.errorBinding, level)}`;
    case "EnsureExpression":
      return `${printExpression(expr.tryExpression, level)} ensure ${printBlock(expr.ensureBlock, level)}`;
    case "RescueExpression":
      return `${printExpression(expr.monitoredExpression, level)} rescue ${printRescueBlock(expr.clauses, level)}`;
    case "IteratorLiteral":
      return printIteratorLiteral(expr, level);
    case "TopicReferenceExpression":
      return "%";
    case "PlaceholderExpression":
      return expr.index ? `@${expr.index}` : "@";
    case "BreakpointExpression":
      return `breakpoint '${printIdentifier(expr.label)} ${printBlock(expr.body, level)}`;
    case "StringInterpolation":
      return printStringInterpolation(expr, level);
    default:
      return "/* expression */";
  }
}

function printStructLiteral(lit: AST.StructLiteral, level: number): string {
  const base = lit.structType ? printIdentifier(lit.structType) : "";
  if (lit.isPositional) {
    const values = lit.fields.map((field) => printExpression(field.value!, level)).join(", ");
    return `${base} { ${values} }`;
  }
  const fields = lit.fields.map((field) => {
    if (field.isShorthand && field.name) {
      return printIdentifier(field.name);
    }
    if (field.name) {
      return `${printIdentifier(field.name)}: ${printExpression(field.value!, level)}`;
    }
    return printExpression(field.value!, level);
  });
  const spreads =
    lit.functionalUpdateSources && lit.functionalUpdateSources.length > 0
      ? lit.functionalUpdateSources.map((src) => `..${printExpression(src, level)}`)
      : [];
  const items = [...spreads, ...fields].join(", ");
  return `${base} { ${items} }`;
}

function printIteratorLiteral(lit: AST.IteratorLiteral, level: number): string {
  const lines = ["{"];
  if (lit.binding) {
    lines.push(`${indent(level + 1)}${printIdentifier(lit.binding)} =>`);
  }
  for (const stmt of lit.body ?? []) {
    lines.push(printStatement(stmt, level + 1));
  }
  lines.push(`${indent(level)}}`);
  const annotation = lit.elementType ? ` ${printTypeExpression(lit.elementType)}` : "";
  return `Iterator${annotation} ${lines.join("\n")}`;
}

function printHandlingBlock(block: AST.BlockExpression, binding: AST.Identifier | undefined, level: number): string {
  const lines = ["{"];
  if (binding) {
    lines.push(`${indent(level + 1)}| ${printIdentifier(binding)} |`);
  }
  for (const stmt of block.body ?? []) {
    lines.push(printStatement(stmt, level + 1));
  }
  lines.push(`${indent(level)}}`);
  return lines.join("\n");
}

function printRescueBlock(clauses: AST.MatchClause[], level: number): string {
  const lines = ["{"];
  clauses.forEach((clause, index) => {
    const suffix = index === clauses.length - 1 ? "" : ",";
    lines.push(`${indent(level + 1)}${printMatchClause(clause, level + 1)}${suffix}`);
  });
  lines.push(`${indent(level)}}`);
  return lines.join("\n");
}

function printStringInterpolation(interp: AST.StringInterpolation, level: number): string {
  const parts = interp.parts
    .map((part) => {
      if (typeof part === "string") {
        return part.replace(/`/g, "\\`").replace(/\$/g, "\\$");
      }
      if (part.type === "StringLiteral") {
        return part.value.replace(/`/g, "\\`").replace(/\$/g, "\\$");
      }
      return `${"${"}${printExpression(part as AST.Expression, level)}${"}"}`;
    })
    .join("");
  return `\`${parts}\``;
}

function printFunctionCall(call: AST.FunctionCall, level: number): string {
  const callee = printExpression(call.callee, level);
  const typeArgs = call.typeArguments && call.typeArguments.length > 0 ? `<${call.typeArguments.map(printTypeExpression).join(", ")}>` : "";
  if (call.isTrailingLambda && call.arguments.length > 0) {
    const trailing = call.arguments[call.arguments.length - 1];
    if (trailing.type === "LambdaExpression") {
      const precedingArgs = call.arguments.slice(0, -1).map((arg) => printExpression(arg, level)).join(", ");
      const callPart = `${callee}${typeArgs}(${precedingArgs})`.replace(/\(\)/, "");
      const spacer = callPart.length > 0 ? " " : "";
      return `${callPart}${spacer}${printLambda(trailing, level)}`.trim();
    }
  }
  const args = call.arguments.map((arg) => printExpression(arg, level)).join(", ");
  return `${callee}${typeArgs}(${args})`;
}

function printLambda(lambda: AST.LambdaExpression, level: number): string {
  const params = lambda.params.map((param) => printPattern(param.name)).join(", ");
  let result = "{";
  if (params.length > 0) {
    result += ` ${params}`;
  }
  if (lambda.returnType) {
    result += ` -> ${printTypeExpression(lambda.returnType)}`;
  }
  const bodyExpr = lambda.body.type === "BlockExpression"
    ? printBlock(lambda.body, level)
    : printExpression(lambda.body, level);
  result += ` => ${bodyExpr}`;
  if (!bodyExpr.endsWith("}")) {
    result += "";
  }
  result += "}";
  return result;
}

function printBinaryOperand(expr: AST.Expression, parentOperator: string, side: "left" | "right", level: number): string {
  const rendered = printExpression(expr, level);
  if (expr.type !== "BinaryExpression") {
    return rendered;
  }
  const parentPrecedence = getBinaryPrecedence(parentOperator);
  const childPrecedence = getBinaryPrecedence(expr.operator);
  if (parentPrecedence === -1 || childPrecedence === -1) {
    return rendered;
  }
  if (side === "left") {
    if (childPrecedence < parentPrecedence || (childPrecedence === parentPrecedence && isRightAssociative(parentOperator))) {
      return `(${rendered})`;
    }
  } else {
    if (childPrecedence < parentPrecedence || (childPrecedence === parentPrecedence && !isRightAssociative(parentOperator))) {
      return `(${rendered})`;
    }
  }
  return rendered;
}

function getBinaryPrecedence(operator: string): number {
  switch (operator) {
    case "||":
      return 1;
    case "&&":
      return 2;
    case "|":
      return 3;
    case "\\xor":
      return 4;
    case "&":
      return 5;
    case "==":
    case "!=":
      return 6;
    case ">":
    case "<":
    case ">=":
    case "<=":
      return 7;
    case "<<":
    case ">>":
      return 8;
    case "+":
    case "-":
      return 9;
    case "*":
    case "/":
    case "%":
      return 10;
    case "**":
      return 11;
    default:
      return -1;
  }
}

function isRightAssociative(operator: string): boolean {
  return operator === "**";
}

function printIfExpression(expr: AST.IfExpression, level: number): string {
  const parts: string[] = [];
  parts.push(`if ${printExpression(expr.ifCondition, level)} ${printBlock(expr.ifBody, level)}`);
  for (const clause of expr.orClauses ?? []) {
    if (clause.condition) {
      parts.push(`or ${printExpression(clause.condition, level)} ${printBlock(clause.body, level)}`);
    } else {
      parts.push(`else ${printBlock(clause.body, level)}`);
    }
  }
  return parts.join("\n");
}

function printMatchExpression(expr: AST.MatchExpression, level: number): string {
  const subject = printExpression(expr.subject, level);
  const lines = [`${subject} match {`];
  const clauses = expr.clauses ?? [];
  clauses.forEach((clause, index) => {
    const suffix = index === clauses.length - 1 ? "" : ",";
    lines.push(`${indent(level + 1)}${printMatchClause(clause, level + 1)}${suffix}`);
  });
  lines.push(`${indent(level)}}`);
  return lines.join("\n");
}

function printMatchClause(clause: AST.MatchClause, level: number): string {
  const pattern = printPattern(clause.pattern);
  const guard = clause.guard ? ` if ${printExpression(clause.guard, level)}` : "";
  const body = clause.body.type === "BlockExpression" ? printBlock(clause.body, level).trim() : printExpression(clause.body, level);
  return `case ${pattern}${guard} => ${body}`;
}

function printBlock(block: AST.BlockExpression, level: number): string {
  const lines = ["{"];
  for (const stmt of block.body ?? []) {
    lines.push(printStatement(stmt, level + 1));
  }
  lines.push(`${indent(level)}}`);
  return lines.join("\n");
}

function printAssignmentLeft(left: AST.Pattern | AST.MemberAccessExpression | AST.IndexExpression): string {
  if (left.type === "MemberAccessExpression" || left.type === "IndexExpression") {
    return printExpression(left, 0);
  }
  return printPattern(left);
}

function printPattern(pattern: AST.Pattern): string {
  switch (pattern.type) {
    case "Identifier":
      return printIdentifier(pattern);
    case "WildcardPattern":
      return "_";
    case "LiteralPattern":
      return printExpression(pattern.literal, 0);
    case "StructPattern":
      if (pattern.isPositional) {
        const fields = pattern.fields.map((field) => printPattern(field.pattern)).join(", ");
        const prefix = pattern.structType ? `${printIdentifier(pattern.structType)} ` : "";
        return `${prefix}{ ${fields} }`;
      }
      const namedFields = pattern.fields.map((field) => {
        if (field.fieldName) {
          return `${printIdentifier(field.fieldName)}: ${printPattern(field.pattern)}`;
        }
        return printPattern(field.pattern);
      });
      return `${pattern.structType ? `${printIdentifier(pattern.structType)} ` : ""}{ ${namedFields.join(", ")} }`;
    case "ArrayPattern":
      const elements = pattern.elements.map(printPattern).join(", ");
      const rest = pattern.restPattern ? `, ...${printPattern(pattern.restPattern)}` : "";
      return `[${elements}${rest}]`;
    case "TypedPattern":
      return `${printPattern(pattern.pattern)}: ${printTypeExpression(pattern.typeAnnotation)}`;
    default:
      return "_";
  }
}

function printFunctionParameter(param: AST.FunctionParameter): string {
  if (param.paramType) {
    return `${printPattern(param.name)}: ${printTypeExpression(param.paramType)}`;
  }
  return printPattern(param.name);
}

function printGenericParameter(param: AST.GenericParameter): string {
  if (param.constraints && param.constraints.length > 0) {
    return `${printIdentifier(param.name)}: ${param.constraints.map((c) => printTypeExpression(c.interfaceType)).join(" + ")}`;
  }
  return printIdentifier(param.name);
}

function printWhereClause(clause: AST.WhereClauseConstraint): string {
  return `${printIdentifier(clause.typeParam)}: ${clause.constraints.map((c) => printTypeExpression(c.interfaceType)).join(" + ")}`;
}

function printTypeExpression(typeExpr: AST.TypeExpression): string {
  switch (typeExpr.type) {
    case "SimpleTypeExpression":
      return printIdentifier(typeExpr.name);
    case "GenericTypeExpression":
      return `${printTypeExpression(typeExpr.base)} ${typeExpr.arguments.map(printTypeExpression).join(" ")}`;
    case "FunctionTypeExpression":
      return `(${typeExpr.paramTypes.map(printTypeExpression).join(", ")}) -> ${printTypeExpression(typeExpr.returnType)}`;
    case "NullableTypeExpression":
      return `?${printTypeExpression(typeExpr.innerType)}`;
    case "ResultTypeExpression":
      return `!${printTypeExpression(typeExpr.innerType)}`;
    case "UnionTypeExpression":
      return typeExpr.members.map(printTypeExpression).join(" | ");
    case "WildcardTypeExpression":
      return "_";
    default:
      return "/* type */";
  }
}

function printFunctionSignature(sig: AST.FunctionSignature): string {
  const parts: string[] = [];
  parts.push("fn");
  parts.push(printIdentifier(sig.name));
  if (sig.genericParams && sig.genericParams.length > 0) {
    parts.push(`<${sig.genericParams.map(printGenericParameter).join(", ")}>`);
  }
  parts.push(`(${sig.params.map(printFunctionParameter).join(", ")})`);
  if (sig.returnType) {
    parts.push(`-> ${printTypeExpression(sig.returnType)}`);
  }
  if (sig.whereClause && sig.whereClause.length > 0) {
    parts.push(`where ${sig.whereClause.map(printWhereClause).join(", ")}`);
  }
  if (sig.defaultImpl) {
    parts.push(printBlock(sig.defaultImpl, 0));
  }
  return parts.join(" ");
}

function printIdentifier(id: AST.Identifier | string | undefined): string {
  if (!id) return "";
  if (typeof id === "string") return id;
  return id.name;
}

function printMember(member: AST.Identifier | AST.IntegerLiteral): string {
  return member.type === "Identifier" ? printIdentifier(member) : member.value.toString();
}

function indent(level: number): string {
  return INDENT.repeat(level);
}

function isExpression(node: AST.Statement): node is AST.Expression {
  return (node as AST.Expression).type !== undefined;
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
