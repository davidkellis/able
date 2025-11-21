import * as AST from "../ast";
import { Environment } from "./environment";
import type { InterpreterV10 } from "./index";
import type { V10Value } from "./values";
import { valuesEqual } from "./value_equals";

interface PatternAssignmentOptions {
  declarationNames?: Set<string>;
  fallbackToDeclaration?: boolean;
}

declare module "./index" {
  interface InterpreterV10 {
    tryMatchPattern(pattern: AST.Pattern, value: V10Value, baseEnv: Environment): Environment | null;
    assignByPattern(
      pattern: AST.Pattern,
      value: V10Value,
      env: Environment,
      isDeclaration: boolean,
      options?: PatternAssignmentOptions,
    ): void;
  }
}

export function applyPatternAugmentations(cls: typeof InterpreterV10): void {
  cls.prototype.tryMatchPattern = function tryMatchPattern(this: InterpreterV10, pattern: AST.Pattern, value: V10Value, baseEnv: Environment): Environment | null {
    if (pattern.type === "Identifier") {
      const e = new Environment(baseEnv);
      e.define(pattern.name, value);
      return e;
    }
    if (pattern.type === "WildcardPattern") {
      return new Environment(baseEnv);
    }
    if (pattern.type === "LiteralPattern") {
      const litVal = this.evaluate(pattern.literal, baseEnv);
      return valuesEqual(litVal, value) ? new Environment(baseEnv) : null;
    }
    if (pattern.type === "StructPattern") {
      if (value.kind !== "struct_instance") return null;
      if (pattern.structType && value.def.id.name !== pattern.structType.name) return null;
      let env = new Environment(baseEnv);
      if (pattern.isPositional) {
        if (!Array.isArray(value.values)) return null;
        if (pattern.fields.length !== value.values.length) return null;
        for (let i = 0; i < pattern.fields.length; i++) {
          const field = pattern.fields[i];
          const val = value.values[i];
          if (!field || val === undefined) return null;
          const sub = this.tryMatchPattern(field.pattern, val as V10Value, env);
          if (!sub) return null;
          env = sub;
        }
        return env;
      }
      if (!(value.values instanceof Map)) return null;
      for (const f of pattern.fields) {
        if (!f.fieldName) return null;
        const name = f.fieldName.name;
        if (!value.values.has(name)) return null;
        const sub = this.tryMatchPattern(f.pattern, value.values.get(name) as V10Value, env);
        if (!sub) return null;
        env = sub;
      }
      return env;
    }
    if (pattern.type === "ArrayPattern") {
      if (value.kind !== "array") return null;
      const arr = value.elements;
      const minLen = pattern.elements.length;
      const hasRest = !!pattern.restPattern;
      if (!hasRest && arr.length !== minLen) return null;
      if (arr.length < minLen) return null;
      let env = new Environment(baseEnv);
      for (let i = 0; i < minLen; i++) {
        const pe = pattern.elements[i];
        const av = arr[i];
        if (!pe || av === undefined) return null;
        const sub = this.tryMatchPattern(pe, av, env);
        if (!sub) return null;
        env = sub;
      }
      if (hasRest && pattern.restPattern && pattern.restPattern.type === "Identifier") {
        env.define(pattern.restPattern.name, this.makeArrayValue(arr.slice(minLen)));
      }
      return env;
    }
    if ((pattern as any).type === "TypedPattern") {
      const tp = pattern as AST.TypedPattern;
      if (!this.matchesType(tp.typeAnnotation, value)) return null;
      const coerced = this.coerceValueToType(tp.typeAnnotation, value);
      return this.tryMatchPattern(tp.pattern, coerced, baseEnv);
    }
    return null;
  };

  cls.prototype.assignByPattern = function assignByPattern(
    this: InterpreterV10,
    pattern: AST.Pattern,
    value: V10Value,
    env: Environment,
    isDeclaration: boolean,
    options?: PatternAssignmentOptions,
  ): void {
    const declarationNames = options?.declarationNames;
    const fallbackToDeclaration = !!options?.fallbackToDeclaration;
    if (pattern.type === "Identifier") {
      if (isDeclaration) {
        const shouldDeclare = !declarationNames || declarationNames.has(pattern.name);
        if (shouldDeclare) {
          env.define(pattern.name, value);
        } else {
          env.assign(pattern.name, value);
        }
      } else {
        if (!env.assignExisting(pattern.name, value)) {
          if (fallbackToDeclaration) {
            env.define(pattern.name, value);
          } else {
            env.assign(pattern.name, value);
          }
        }
      }
      return;
    }
    if (pattern.type === "WildcardPattern") return;
    if (pattern.type === "LiteralPattern") {
      const lit = this.evaluate(pattern.literal, env);
      if (!valuesEqual(lit, value)) throw new Error("Pattern literal mismatch in assignment");
      return;
    }
    if (pattern.type === "StructPattern") {
      if (value.kind !== "struct_instance") throw new Error("Cannot destructure non-struct value");
      if (pattern.structType && value.def.id.name !== pattern.structType.name) throw new Error("Struct type mismatch in destructuring");
      if (pattern.isPositional) {
        if (!Array.isArray(value.values)) throw new Error("Expected positional struct");
        if (pattern.fields.length !== value.values.length) throw new Error("Struct field count mismatch");
        for (let i = 0; i < pattern.fields.length; i++) {
          const fieldPat = pattern.fields[i];
          const fieldVal = value.values[i];
          if (!fieldPat || fieldVal === undefined) throw new Error("Invalid positional field during destructuring");
          this.assignByPattern(fieldPat.pattern, fieldVal, env, isDeclaration, options);
        }
        return;
      }
      if (!(value.values instanceof Map)) throw new Error("Expected named struct");
      for (const f of pattern.fields) {
        if (!f.fieldName) throw new Error("Named struct pattern missing field name");
        const name = f.fieldName.name;
        if (!value.values.has(name)) throw new Error(`Missing field '${name}' during destructuring`);
        const fieldVal = value.values.get(name)!;
        this.assignByPattern(f.pattern, fieldVal, env, isDeclaration, options);
        if (f.binding?.name) {
          this.assignByPattern(f.binding as AST.Pattern, fieldVal, env, isDeclaration, options);
        }
      }
      return;
    }
    if (pattern.type === "ArrayPattern") {
      if (value.kind !== "array") throw new Error("Cannot destructure non-array value");
      const arr = value.elements;
      const minLen = pattern.elements.length;
      const hasRest = !!pattern.restPattern;
      if (!hasRest && arr.length !== minLen) throw new Error("Array length mismatch in destructuring");
      if (arr.length < minLen) throw new Error("Array too short for destructuring");
      for (let i = 0; i < minLen; i++) {
        const pe = pattern.elements[i];
        const av = arr[i];
        if (!pe || av === undefined) throw new Error("Invalid array element during destructuring");
        this.assignByPattern(pe, av, env, isDeclaration, options);
      }
      if (hasRest && pattern.restPattern && pattern.restPattern.type === "Identifier") {
        const rest = this.makeArrayValue(arr.slice(minLen)) as V10Value;
        if (isDeclaration) {
          const shouldDeclare = !declarationNames || declarationNames.has(pattern.restPattern.name);
          if (shouldDeclare) env.define(pattern.restPattern.name, rest);
          else env.assign(pattern.restPattern.name, rest);
        } else if (!env.assignExisting(pattern.restPattern.name, rest)) {
          if (fallbackToDeclaration) env.define(pattern.restPattern.name, rest);
          else env.assign(pattern.restPattern.name, rest);
        }
      }
      return;
    }
    if ((pattern as any).type === "TypedPattern") {
      const tp = pattern as AST.TypedPattern;
      if (!this.matchesType(tp.typeAnnotation, value)) {
        const expected = this.typeExpressionToString(tp.typeAnnotation);
        const actual = this.getTypeNameForValue(value) ?? value.kind;
        throw new Error(`Typed pattern mismatch in assignment: expected ${expected}, got ${actual}`);
      }
      const coerced = this.coerceValueToType(tp.typeAnnotation, value);
      this.assignByPattern(tp.pattern, coerced, env, isDeclaration, options);
      return;
    }
    throw new Error(`Unsupported pattern in assignment: ${(pattern as any).type}`);
  };
}
