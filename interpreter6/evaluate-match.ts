import * as AST from "./ast";
import { Environment } from "./environment";
import type { Interpreter } from "./interpreter";
import type { AbleArray, AbleValue } from "./runtime";
import { isAbleArray, isAblePrimitive, isAbleStructInstance } from "./runtime";

export function matchPattern(this: Interpreter, pattern: AST.Pattern, value: AbleValue, environment: Environment): Environment | null {
  const self = this as any;
  switch (pattern.type) {
    case "Identifier": {
      const matchEnv = new Environment(environment);
      matchEnv.define(pattern.name, value);
      return matchEnv;
    }
    case "WildcardPattern":
      return new Environment(environment);
    case "LiteralPattern": {
      const patternVal = self.evaluate(pattern.literal, environment);
      if (isAblePrimitive(value) && isAblePrimitive(patternVal)) {
        if (value.kind === patternVal.kind && value.value === patternVal.value) {
          return new Environment(environment);
        }
      }
      return null;
    }
    case "StructPattern":
      if (!value || !isAbleStructInstance(value)) return null;
      if (pattern.structType && value.definition.name !== pattern.structType.name) return null;

      let structMatchEnv = new Environment(environment);
      if (pattern.isPositional) {
        if (!Array.isArray(value.values)) return null;
        if (pattern.fields.length !== value.values.length) return null;

        for (let i = 0; i < pattern.fields.length; i++) {
          const fieldPatternNode = pattern.fields[i];
          const fieldValue = value.values[i];
          if (!fieldPatternNode?.pattern || fieldValue === undefined) {
            console.error("Internal Error: Invalid field pattern or value during struct pattern matching.", { fieldPatternNode, fieldValue });
            return null;
          }

        const subMatchEnv = self.matchPattern(fieldPatternNode.pattern, fieldValue, structMatchEnv);
          if (!subMatchEnv) return null;
          structMatchEnv = subMatchEnv;
        }
      } else {
        if (!(value.values instanceof Map)) return null;
        const matchedFields = new Set<string>();

        for (const fieldPatternNode of pattern.fields) {
          if (!fieldPatternNode.fieldName) return null;
          const fieldName = fieldPatternNode.fieldName.name;
          if (!value.values.has(fieldName)) return null;

          const fieldValue = value.values.get(fieldName)!;
          const subMatchEnv = self.matchPattern(fieldPatternNode.pattern, fieldValue, structMatchEnv);
          if (!subMatchEnv) return null;
          structMatchEnv = subMatchEnv;
          matchedFields.add(fieldName);
        }
      }
      return structMatchEnv;
    case "ArrayPattern":
      if (!value || !isAbleArray(value)) return null;

      const minLen = pattern.elements.length;
      const hasRest = !!pattern.restPattern;

      if (value.elements.length < minLen) return null;
      if (!hasRest && value.elements.length !== minLen) return null;

      let arrayMatchEnv = new Environment(environment);
      for (let i = 0; i < minLen; i++) {
        const elemPattern = pattern.elements[i];
        const elemValue = value.elements[i];
        if (!elemPattern || elemValue === undefined) {
          console.error("Internal Error: Invalid element pattern or value during array pattern matching.", { elemPattern, elemValue });
          return null;
        }

        const subMatchEnv = self.matchPattern(elemPattern, elemValue, arrayMatchEnv);
        if (!subMatchEnv) return null;
        arrayMatchEnv = subMatchEnv;
      }

      if (hasRest && pattern.restPattern) {
        const restValue: AbleArray = { kind: "array", elements: value.elements.slice(minLen) };
        if (pattern.restPattern.type === "Identifier" || pattern.restPattern.type === "WildcardPattern") {
          const subMatchEnv = self.matchPattern(pattern.restPattern, restValue, arrayMatchEnv);
          if (!subMatchEnv) return null;
          arrayMatchEnv = subMatchEnv;
        } else {
          return null;
        }
      }
      return arrayMatchEnv;
    default: {
      const _exhaustiveCheck: never = pattern;
      throw new Error(`Interpreter Error: Unsupported pattern type in matchPattern: ${(_exhaustiveCheck as any).type}`);
    }
  }
}

export function evaluateMatchExpression(this: Interpreter, node: AST.MatchExpression, environment: Environment): AbleValue {
  const self = this as any;
  const subjectValue = self.evaluate(node.subject, environment);

  for (const clause of node.clauses) {
    const matchEnv = self.matchPattern(clause.pattern, subjectValue, environment);

    if (matchEnv) {
      let guardResult = true;
      if (clause.guard) {
        const guardValue = self.evaluate(clause.guard, matchEnv);
        guardResult = self.isTruthy(guardValue);
      }

      if (guardResult) {
        return self.evaluate(clause.body, matchEnv);
      }
    }
  }

  throw new Error("Interpreter Error: Non-exhaustive match expression");
}
