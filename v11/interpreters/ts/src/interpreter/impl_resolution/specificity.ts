import type * as AST from "../../ast";
import type { Interpreter } from "../index";
import type { ConstraintSpec, ImplMethodEntry, RuntimeValue } from "../values";

export function applyImplSpecificityAugmentations(cls: typeof Interpreter): void {
  cls.prototype.compareMethodMatches = function compareMethodMatches(
    this: Interpreter,
    a: {
      method?: Extract<RuntimeValue, { kind: "function" | "function_overload" }>;
      score: number;
      entry: ImplMethodEntry;
      constraints: ConstraintSpec[];
      isConcreteTarget: boolean;
    },
    b: {
      method?: Extract<RuntimeValue, { kind: "function" | "function_overload" }>;
      score: number;
      entry: ImplMethodEntry;
      constraints: ConstraintSpec[];
      isConcreteTarget: boolean;
    },
  ): number {
    if (a.isConcreteTarget && !b.isConcreteTarget) return 1;
    if (b.isConcreteTarget && !a.isConcreteTarget) return -1;
    const aConstraints = this.buildConstraintKeySet(a.constraints);
    const bConstraints = this.buildConstraintKeySet(b.constraints);
    if (this.isConstraintSuperset(aConstraints, bConstraints)) return 1;
    if (this.isConstraintSuperset(bConstraints, aConstraints)) return -1;
    const aUnion = a.entry.unionVariantSignatures;
    const bUnion = b.entry.unionVariantSignatures;
    const aUnionSize = aUnion?.length ?? 0;
    const bUnionSize = bUnion?.length ?? 0;
    if (aUnionSize !== bUnionSize) {
      if (aUnionSize === 0) return 1;
      if (bUnionSize === 0) return -1;
    }
    if (aUnion && bUnion) {
      if (this.isProperSubset(aUnion, bUnion)) return 1;
      if (this.isProperSubset(bUnion, aUnion)) return -1;
      if (aUnion.length !== bUnion.length) {
        return aUnion.length < bUnion.length ? 1 : -1;
      }
    }
    if (a.score > b.score) return 1;
    if (a.score < b.score) return -1;
    const aPriority = typeof (a.method as any)?.methodResolutionPriority === "number"
      ? (a.method as any).methodResolutionPriority
      : 0;
    const bPriority = typeof (b.method as any)?.methodResolutionPriority === "number"
      ? (b.method as any).methodResolutionPriority
      : 0;
    if (aPriority > bPriority) return 1;
    if (aPriority < bPriority) return -1;
    return 0;
  };

  cls.prototype.buildConstraintKeySet = function buildConstraintKeySet(
    this: Interpreter,
    constraints: ConstraintSpec[],
  ): Set<string> {
    const set = new Set<string>();
    for (const c of constraints) {
      const expanded = this.collectInterfaceConstraintExpressions(c.ifaceType);
      for (const expr of expanded) {
        set.add(`${this.typeExpressionToString(c.subjectExpr)}->${this.typeExpressionToString(expr)}`);
      }
    }
    return set;
  };

  cls.prototype.isConstraintSuperset = function isConstraintSuperset(
    this: Interpreter,
    a: Set<string>,
    b: Set<string>,
  ): boolean {
    if (a.size <= b.size) return false;
    for (const key of b) {
      if (!a.has(key)) return false;
    }
    return true;
  };

  cls.prototype.isProperSubset = function isProperSubset(this: Interpreter, a: string[], b: string[]): boolean {
    const aSet = new Set(a);
    const bSet = new Set(b);
    if (aSet.size >= bSet.size) return false;
    for (const val of aSet) {
      if (!bSet.has(val)) return false;
    }
    return true;
  };

  cls.prototype.measureTemplateSpecificity = function measureTemplateSpecificity(
    this: Interpreter,
    t: AST.TypeExpression,
    genericNames: Set<string>,
  ): number {
    switch (t.type) {
      case "SimpleTypeExpression":
        return genericNames.has(t.name.name) ? 0 : 1;
      case "GenericTypeExpression": {
        let score = this.measureTemplateSpecificity(t.base, genericNames);
        for (const arg of t.arguments ?? []) {
          score += this.measureTemplateSpecificity(arg, genericNames);
        }
        return score;
      }
      case "NullableTypeExpression":
      case "ResultTypeExpression":
        return this.measureTemplateSpecificity(t.innerType, genericNames);
      case "UnionTypeExpression":
        return t.members.reduce((acc, member) => acc + this.measureTemplateSpecificity(member, genericNames), 0);
      default:
        return 0;
    }
  };
}
