import * as AST from "../ast";
import type { BytecodeInstruction, BytecodeProgram } from "./bytecode";
import { makeIntegerValue, makeFloatValue } from "../interpreter/numeric";
import type { RuntimeValue } from "../interpreter/values";

type LoweringContext = {
  instructions: BytecodeInstruction[];
  locals: Map<string, number>;
};

const NIL_VALUE: RuntimeValue = { kind: "nil", value: null };
const VOID_VALUE: RuntimeValue = { kind: "void" };

export function lowerExpression(expr: AST.Expression): BytecodeProgram {
  const ctx: LoweringContext = { instructions: [], locals: new Map() };
  emitExpression(ctx, expr);
  ctx.instructions.push({ op: "return" });
  return { instructions: ctx.instructions, locals: ctx.locals.size };
}

export function lowerModule(module: AST.Module): BytecodeProgram {
  const ctx: LoweringContext = { instructions: [], locals: new Map() };
  for (let i = 0; i < module.body.length; i += 1) {
    const stmt = module.body[i]!;
    const isLast = i === module.body.length - 1;
    emitStatement(ctx, stmt, isLast);
  }
  if (ctx.instructions.length === 0) {
    ctx.instructions.push({ op: "const", value: NIL_VALUE });
    ctx.instructions.push({ op: "return" });
  } else if (ctx.instructions[ctx.instructions.length - 1]?.op !== "return") {
    ctx.instructions.push({ op: "return" });
  }
  return { instructions: ctx.instructions, locals: ctx.locals.size };
}

function emitStatement(ctx: LoweringContext, stmt: AST.Statement, isLast: boolean): void {
  switch (stmt.type) {
    case "ReturnStatement": {
      if (stmt.argument) {
        emitExpression(ctx, stmt.argument);
      } else {
        ctx.instructions.push({ op: "const", value: VOID_VALUE });
      }
      ctx.instructions.push({ op: "return" });
      return;
    }
    default:
      if (isExpression(stmt)) {
        emitExpression(ctx, stmt);
        if (!isLast) ctx.instructions.push({ op: "pop" });
        return;
      }
      throw new Error(`bytecode lowering unsupported statement: ${stmt.type}`);
  }
}

function emitExpression(ctx: LoweringContext, expr: AST.Expression): void {
  switch (expr.type) {
    case "IntegerLiteral": {
      const kind = expr.integerType ?? "i32";
      const raw = typeof expr.value === "bigint" ? expr.value : BigInt(expr.value ?? 0);
      ctx.instructions.push({ op: "const", value: makeIntegerValue(kind, raw) });
      return;
    }
    case "FloatLiteral": {
      const kind = expr.floatType ?? "f64";
      ctx.instructions.push({ op: "const", value: makeFloatValue(kind, expr.value) });
      return;
    }
    case "BooleanLiteral":
      ctx.instructions.push({ op: "const", value: { kind: "bool", value: expr.value } });
      return;
    case "NilLiteral":
      ctx.instructions.push({ op: "const", value: NIL_VALUE });
      return;
    case "Identifier": {
      const slot = lookupLocal(ctx, expr.name);
      ctx.instructions.push({ op: "load", slot });
      return;
    }
    case "BinaryExpression": {
      if (expr.operator !== "+") {
        throw new Error(`bytecode lowering unsupported binary operator: ${expr.operator}`);
      }
      emitExpression(ctx, expr.left);
      emitExpression(ctx, expr.right);
      ctx.instructions.push({ op: "add" });
      return;
    }
    case "AssignmentExpression": {
      if (expr.operator !== ":=" && expr.operator !== "=") {
        throw new Error(`bytecode lowering unsupported assignment operator: ${expr.operator}`);
      }
      const name = resolveIdentifierPattern(expr.left);
      if (!name) {
        throw new Error("bytecode lowering only supports identifier assignments");
      }
      emitExpression(ctx, expr.right);
      const slot = expr.operator === ":=" ? declareLocal(ctx, name) : lookupLocal(ctx, name);
      ctx.instructions.push({ op: "store", slot });
      return;
    }
    case "BlockExpression": {
      if (expr.body.length === 0) {
        ctx.instructions.push({ op: "const", value: NIL_VALUE });
        return;
      }
      for (let i = 0; i < expr.body.length; i += 1) {
        const stmt = expr.body[i]!;
        const isLast = i === expr.body.length - 1;
        emitStatement(ctx, stmt, isLast);
      }
      return;
    }
    default:
      throw new Error(`bytecode lowering unsupported expression: ${expr.type}`);
  }
}

function isExpression(stmt: AST.Statement): stmt is AST.Expression {
  return ![
    "FunctionDefinition",
    "StructDefinition",
    "UnionDefinition",
    "TypeAliasDefinition",
    "InterfaceDefinition",
    "ImplementationDefinition",
    "MethodsDefinition",
    "ImportStatement",
    "PackageStatement",
    "ReturnStatement",
    "RaiseStatement",
    "RethrowStatement",
    "BreakStatement",
    "ContinueStatement",
    "WhileLoop",
    "ForLoop",
    "LoopExpression",
    "YieldStatement",
    "PreludeStatement",
    "ExternFunctionBody",
    "DynImportStatement",
  ].includes(stmt.type);
}

function resolveIdentifierPattern(pattern: AST.Pattern | AST.MemberAccessExpression | AST.IndexExpression | AST.ImplicitMemberExpression): string | null {
  if (pattern.type === "Identifier") return pattern.name;
  if (pattern.type === "TypedPattern") {
    return pattern.pattern.type === "Identifier" ? pattern.pattern.name : null;
  }
  return null;
}

function declareLocal(ctx: LoweringContext, name: string): number {
  const existing = ctx.locals.get(name);
  if (existing !== undefined) return existing;
  const slot = ctx.locals.size;
  ctx.locals.set(name, slot);
  return slot;
}

function lookupLocal(ctx: LoweringContext, name: string): number {
  const slot = ctx.locals.get(name);
  if (slot === undefined) {
    throw new Error(`bytecode lowering missing local '${name}'`);
  }
  return slot;
}
