import * as AST from "../../ast";
import type { Interpreter } from "../index";

export function applyTypeFormatAugmentations(cls: typeof Interpreter): void {
  cls.prototype.expandTypeAliases = function expandTypeAliases(
    this: Interpreter,
    t: AST.TypeExpression,
    seen = new Set<string>(),
    seenNodes = new Set<AST.TypeExpression>(),
  ): AST.TypeExpression {
    if (!t) return t;
    if (seenNodes.has(t)) return t;
    seenNodes.add(t);
    let result: AST.TypeExpression = t;
    switch (t.type) {
      case "SimpleTypeExpression": {
        const name = t.name.name;
        const alias = this.typeAliases.get(name);
        if (!alias?.targetType || seen.has(name)) {
          result = t;
          break;
        }
        seen.add(name);
        const expanded = this.expandTypeAliases(alias.targetType, seen, seenNodes);
        seen.delete(name);
        result = expanded ?? t;
        break;
      }
      case "GenericTypeExpression": {
        const baseName = t.base.type === "SimpleTypeExpression" ? t.base.name.name : null;
        const expandedArgs = (t.arguments ?? []).map((arg) => (arg ? this.expandTypeAliases(arg, seen, seenNodes) : arg));
        if (!baseName) {
          const expandedBase = this.expandTypeAliases(t.base, seen, seenNodes);
          result = { ...t, base: expandedBase, arguments: expandedArgs };
          break;
        }
        const alias = this.typeAliases.get(baseName);
        if (!alias?.targetType || seen.has(baseName)) {
          const expandedBase = seen.has(baseName) ? t.base : this.expandTypeAliases(t.base, seen, seenNodes);
          result = { ...t, base: expandedBase, arguments: expandedArgs };
          break;
        }
        const substitutions = new Map<string, AST.TypeExpression>();
        (alias.genericParams ?? []).forEach((param, index) => {
          const paramName = param?.name?.name;
          if (!paramName) return;
          substitutions.set(paramName, expandedArgs[index] ?? AST.wildcardTypeExpression());
        });
        const substitute = (expr: AST.TypeExpression): AST.TypeExpression => {
          switch (expr.type) {
            case "SimpleTypeExpression": {
              const name = expr.name.name;
              if (substitutions.has(name)) {
                return this.expandTypeAliases(substitutions.get(name)!, seen, seenNodes);
              }
              return this.expandTypeAliases(expr, seen, seenNodes);
            }
            case "GenericTypeExpression":
              return {
                ...expr,
                base: substitute(expr.base),
                arguments: (expr.arguments ?? []).map((arg) => (arg ? substitute(arg) : arg)),
              };
            case "NullableTypeExpression":
              return { ...expr, innerType: substitute(expr.innerType) };
            case "ResultTypeExpression":
              return { ...expr, innerType: substitute(expr.innerType) };
            case "UnionTypeExpression":
              return { ...expr, members: (expr.members ?? []).map((member) => substitute(member)) };
            case "FunctionTypeExpression":
              return {
                ...expr,
                paramTypes: (expr.paramTypes ?? []).map((param) => substitute(param)),
                returnType: substitute(expr.returnType),
              };
            default:
              return expr;
          }
        };
        seen.add(baseName);
        const substituted = substitute(alias.targetType);
        const expanded = this.expandTypeAliases(substituted, seen, seenNodes);
        seen.delete(baseName);
        result = expanded ?? substituted;
        break;
      }
      case "NullableTypeExpression":
        result = { ...t, innerType: this.expandTypeAliases(t.innerType, seen, seenNodes) };
        break;
      case "ResultTypeExpression":
        result = { ...t, innerType: this.expandTypeAliases(t.innerType, seen, seenNodes) };
        break;
      case "UnionTypeExpression":
        result = { ...t, members: (t.members ?? []).map((member) => this.expandTypeAliases(member, seen, seenNodes)) };
        break;
      case "FunctionTypeExpression":
        result = {
          ...t,
          paramTypes: (t.paramTypes ?? []).map((param) => this.expandTypeAliases(param, seen, seenNodes)),
          returnType: this.expandTypeAliases(t.returnType, seen, seenNodes),
        };
        break;
      default:
        result = t;
    }
    seenNodes.delete(t);
    return result;
  };

  cls.prototype.typeExpressionToString = function typeExpressionToString(
    this: Interpreter,
    t: AST.TypeExpression,
    seen = new Set<string>(),
    seenNodes = new Set<AST.TypeExpression>(),
    depth = 0,
  ): string {
    if (depth > 100) {
      return "<type>";
    }
    if (seenNodes.has(t)) {
      return "<type>";
    }
    seenNodes.add(t);
    const canonical = this.expandTypeAliases(t, seen, seenNodes);
    let result: string;
    switch (canonical.type) {
      case "SimpleTypeExpression":
        result = canonical.name.name;
        break;
      case "GenericTypeExpression":
        result = `${this.typeExpressionToString(canonical.base, seen, seenNodes, depth + 1)}<${(canonical.arguments ?? []).map(arg => this.typeExpressionToString(arg, seen, seenNodes, depth + 1)).join(",")}>`;
        break;
      case "NullableTypeExpression":
        result = `${this.typeExpressionToString(canonical.innerType, seen, seenNodes, depth + 1)}?`;
        break;
      case "ResultTypeExpression":
        result = `Result<${this.typeExpressionToString(canonical.innerType, seen, seenNodes, depth + 1)}>`;
        break;
      case "FunctionTypeExpression":
        result = `(${canonical.paramTypes.map(pt => this.typeExpressionToString(pt, seen, seenNodes, depth + 1)).join(", ")}) -> ${this.typeExpressionToString(canonical.returnType, seen, seenNodes, depth + 1)}`;
        break;
      case "UnionTypeExpression":
        result = canonical.members.map(member => this.typeExpressionToString(member, seen, seenNodes, depth + 1)).join(" | ");
        break;
      case "WildcardTypeExpression":
        result = "_";
        break;
      default:
        result = "<type>";
    }
    seenNodes.delete(t);
    return result;
  };

  cls.prototype.parseTypeExpression = function parseTypeExpression(this: Interpreter, t: AST.TypeExpression): { name: string; typeArgs: AST.TypeExpression[] } | null {
    const canonical = this.expandTypeAliases(t);
    if (canonical.type === "SimpleTypeExpression") {
      return { name: canonical.name.name, typeArgs: [] };
    }
    if (canonical.type === "GenericTypeExpression" && canonical.base.type === "SimpleTypeExpression") {
      return { name: canonical.base.name.name, typeArgs: canonical.arguments ?? [] };
    }
    return null;
  };

  cls.prototype.cloneTypeExpression = function cloneTypeExpression(this: Interpreter, t: AST.TypeExpression): AST.TypeExpression {
    switch (t.type) {
      case "SimpleTypeExpression":
        return { type: "SimpleTypeExpression", name: AST.identifier(t.name.name) };
      case "GenericTypeExpression":
        return {
          type: "GenericTypeExpression",
          base: this.cloneTypeExpression(t.base),
          arguments: (t.arguments ?? []).map(arg => this.cloneTypeExpression(arg)),
        };
      case "FunctionTypeExpression":
        return {
          type: "FunctionTypeExpression",
          paramTypes: t.paramTypes.map(pt => this.cloneTypeExpression(pt)),
          returnType: this.cloneTypeExpression(t.returnType),
        };
      case "NullableTypeExpression":
        return { type: "NullableTypeExpression", innerType: this.cloneTypeExpression(t.innerType) };
      case "ResultTypeExpression":
        return { type: "ResultTypeExpression", innerType: this.cloneTypeExpression(t.innerType) };
      case "UnionTypeExpression":
        return { type: "UnionTypeExpression", members: t.members.map(member => this.cloneTypeExpression(member)) };
      case "WildcardTypeExpression":
      default:
        return { type: "WildcardTypeExpression" };
    }
  };

  cls.prototype.typeExpressionsEqual = function typeExpressionsEqual(this: Interpreter, a: AST.TypeExpression, b: AST.TypeExpression, depth = 0): boolean {
    if (depth > 100) return true;
    const left = this.expandTypeAliases(a);
    const right = this.expandTypeAliases(b);
    if (left.type !== right.type) return false;
    switch (left.type) {
      case "SimpleTypeExpression":
        return left.name.name === (right as AST.SimpleTypeExpression).name.name;
      case "GenericTypeExpression": {
        const r = right as AST.GenericTypeExpression;
        if (!this.typeExpressionsEqual(left.base, r.base, depth + 1)) return false;
        const lArgs = left.arguments ?? [];
        const rArgs = r.arguments ?? [];
        if (lArgs.length !== rArgs.length) return false;
        for (let i = 0; i < lArgs.length; i++) {
          const lArg = lArgs[i]!;
          const rArg = rArgs[i]!;
          if (!this.typeExpressionsEqual(lArg, rArg, depth + 1)) return false;
        }
        return true;
      }
      case "NullableTypeExpression":
        return this.typeExpressionsEqual(left.innerType, (right as AST.NullableTypeExpression).innerType, depth + 1);
      case "ResultTypeExpression":
        return this.typeExpressionsEqual(left.innerType, (right as AST.ResultTypeExpression).innerType, depth + 1);
      case "FunctionTypeExpression": {
        const r = right as AST.FunctionTypeExpression;
        const lParams = left.paramTypes ?? [];
        const rParams = r.paramTypes ?? [];
        if (lParams.length !== rParams.length) return false;
        for (let i = 0; i < lParams.length; i++) {
          if (!this.typeExpressionsEqual(lParams[i]!, rParams[i]!, depth + 1)) return false;
        }
        return this.typeExpressionsEqual(left.returnType, r.returnType, depth + 1);
      }
      case "UnionTypeExpression": {
        const r = right as AST.UnionTypeExpression;
        const lMembers = left.members ?? [];
        const rMembers = r.members ?? [];
        if (lMembers.length !== rMembers.length) return false;
        for (let i = 0; i < lMembers.length; i++) {
          if (!this.typeExpressionsEqual(lMembers[i]!, rMembers[i]!, depth + 1)) return false;
        }
        return true;
      }
      case "WildcardTypeExpression":
        return true;
      default:
        return true;
    }
  };
}
